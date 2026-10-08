package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chatRpc "cyi-app/api/service/chat/v1"
	"cyi-app/pkg/kratos/event"
	"cyi-app/pkg/kratos/server/ws"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// 在线状态 Redis key（与 chat 服务 biz 层约定一致）
const (
	userOnlineKeyPrefix = "chat:online:user:"
	onlineKeyTTL        = 90 * time.Second
)

// ChatRepo 聊天数据仓储接口（data 层实现，转发到 chat 服务）
type ChatRepo interface {
	// CreateConversation 创建会话（复用未结束会话）
	CreateConversation(ctx context.Context, userID int64) (*chatRpc.ConversationInfo, error)
	// ListMyConversation 我的会话列表
	ListMyConversation(ctx context.Context, userID int64, page, pageSize, status int32) (*chatRpc.ListMyConversationReply, error)
	// ListMessage 会话历史消息
	ListMessage(ctx context.Context, userID, conversationID, cursorID int64, limit, direction int32) (*chatRpc.ListMessageReply, error)
	// SendMessage 发送消息
	SendMessage(ctx context.Context, userID, conversationID int64, msgType int32, content, extra string) (*chatRpc.SendMessageReply, error)
	// MarkConversationRead 标记会话已读
	MarkConversationRead(ctx context.Context, userID, conversationID int64) error
	// CloseConversation 结束会话
	CloseConversation(ctx context.Context, userID, conversationID int64) error
}

// ChatUseCase 用户端 IM 用例
// 持有 chat 数据仓储、本地 Hub、NATS 连接与 Redis：
//   - 提供用户端会话/消息操作的转发（走 ChatRepo）
//   - 订阅 chat.message 主题，将消息实时推送至本网关的用户连接
//   - 维护用户在线状态（WS 连接注册/心跳续期/注销清理）
type ChatUseCase struct {
	chatRepo ChatRepo
	hub      *ws.Hub
	nc       *nats.Conn
	rdb      redis.UniversalClient
	log      *log.Helper
}

// NewChatUseCase 创建用户端 IM 用例（构造时自动启动 NATS 订阅）
func NewChatUseCase(chatRepo ChatRepo, nc *nats.Conn, rdb redis.UniversalClient, logger log.Logger) *ChatUseCase {
	uc := &ChatUseCase{chatRepo: chatRepo, hub: ws.NewHub(), nc: nc, rdb: rdb, log: log.NewHelper(logger)}
	uc.startSubscriber()
	return uc
}

// OnlineKey 返回用户在线状态 Redis key
func (uc *ChatUseCase) OnlineKey(userID int64) string {
	return fmt.Sprintf("%s%d", userOnlineKeyPrefix, userID)
}

// MarkUserOnline 写入用户在线状态（TTL 保活，由心跳续期）
func (uc *ChatUseCase) MarkUserOnline(key string) {
	if uc.rdb != nil {
		uc.rdb.Set(context.Background(), key, "1", onlineKeyTTL)
	}
}

// MarkUserOffline 清除用户在线状态（连接注销）
func (uc *ChatUseCase) MarkUserOffline(key string) {
	if uc.rdb != nil {
		uc.rdb.Del(context.Background(), key)
	}
}

// Hub 返回连接分发器（供 WebSocket handler 注册/注销连接）
func (uc *ChatUseCase) Hub() *ws.Hub { return uc.hub }

// CreateConversation 创建会话（复用未结束会话）
func (uc *ChatUseCase) CreateConversation(ctx context.Context, userID int64) (*chatRpc.ConversationInfo, error) {
	return uc.chatRepo.CreateConversation(ctx, userID)
}

// ListMyConversation 我的会话列表
func (uc *ChatUseCase) ListMyConversation(ctx context.Context, userID int64, page, pageSize, status int32) (*chatRpc.ListMyConversationReply, error) {
	return uc.chatRepo.ListMyConversation(ctx, userID, page, pageSize, status)
}

// ListMessage 会话历史消息（校验用户成员）
func (uc *ChatUseCase) ListMessage(ctx context.Context, userID, conversationID, cursorID int64, limit, direction int32) (*chatRpc.ListMessageReply, error) {
	return uc.chatRepo.ListMessage(ctx, userID, conversationID, cursorID, limit, direction)
}

// SendMessage 用户发送消息
func (uc *ChatUseCase) SendMessage(ctx context.Context, userID, conversationID int64, msgType int32, content, extra string) (*chatRpc.SendMessageReply, error) {
	return uc.chatRepo.SendMessage(ctx, userID, conversationID, msgType, content, extra)
}

// MarkConversationRead 用户标记会话已读
func (uc *ChatUseCase) MarkConversationRead(ctx context.Context, userID, conversationID int64) error {
	return uc.chatRepo.MarkConversationRead(ctx, userID, conversationID)
}

// CloseConversation 用户结束会话
func (uc *ChatUseCase) CloseConversation(ctx context.Context, userID, conversationID int64) error {
	return uc.chatRepo.CloseConversation(ctx, userID, conversationID)
}

// startSubscriber 订阅 NATS chat.message，将新消息实时推送给在线用户连接
func (uc *ChatUseCase) startSubscriber() {
	if uc.nc == nil {
		return
	}
	_, err := uc.nc.Subscribe(event.SubjectChatMessage, func(m *nats.Msg) {
		var evt event.ChatMessageEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			uc.log.Warnf("解析聊天消息事件失败: %v", err)
			return
		}
		// 只处理本网关关心的接收方（用户）
		if evt.ReceiverType != event.ChatReceiverUser {
			return
		}
		frame := ws.NewServerFrame("message", messageEventToItem(&evt))
		if frame != nil {
			uc.hub.PushToUser(evt.ReceiverID, frame)
		}
	})
	if err != nil {
		uc.log.Errorf("订阅聊天消息主题失败: %v", err)
	}
	// 会话状态事件（新建/接入/转接/结束）推送给对应用户连接，刷新用户端会话状态
	if _, err := uc.nc.Subscribe(event.SubjectChatConversation, func(m *nats.Msg) {
		var evt event.ChatConversationEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			uc.log.Warnf("解析会话状态事件失败: %v", err)
			return
		}
		frame := ws.NewServerFrame("conversation_status", evt)
		if frame != nil {
			uc.hub.PushToUser(evt.UserID, frame)
		}
	}); err != nil {
		uc.log.Errorf("订阅会话状态主题失败: %v", err)
	}
}

// messageEventToItem 将 NATS 事件转换为下发给客户端的消息项
func messageEventToItem(evt *event.ChatMessageEvent) *chatRpc.ChatMessageInfo {
	return &chatRpc.ChatMessageInfo{
		Id:             evt.MessageID,
		ConversationId: evt.ConversationID,
		SenderType:     evt.SenderType,
		SenderId:       evt.SenderID,
		MsgType:        evt.MsgType,
		Content:        evt.Content,
		Extra:          evt.Extra,
		CreatedAt:      evt.CreatedAt,
	}
}
