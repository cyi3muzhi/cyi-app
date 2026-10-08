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

// 客服在线状态 Redis key（M4：由 admin WS 连接维护）
const (
	agentOnlineKeyPrefix = "chat:online:agent:"
	agentOnlineKeyTTL    = 90 * time.Second
)

// ChatUseCase 客服端 IM 用例
// 持有 chat 服务 gRPC 客户端、本地 Hub、NATS 连接与 Redis：
//   - 提供客服端消息收发/已读/结束的 gRPC 转发
//   - 订阅 chat.message 主题，将消息实时推送至客服工作台连接
//   - 维护客服在线状态（WS 连接注册/心跳续期/注销清理）
type ChatUseCase struct {
	cc  chatRpc.ChatClient
	hub *ws.Hub
	nc  *nats.Conn
	rdb redis.UniversalClient
	log *log.Helper
}

// NewChatUseCase 创建客服端 IM 用例（构造时自动启动 NATS 订阅）
func NewChatUseCase(cc chatRpc.ChatClient, nc *nats.Conn, rdb redis.UniversalClient, logger log.Logger) *ChatUseCase {
	uc := &ChatUseCase{cc: cc, hub: ws.NewHub(), nc: nc, rdb: rdb, log: log.NewHelper(logger)}
	uc.startSubscriber()
	return uc
}

// OnlineKey 返回客服在线状态 Redis key
func (uc *ChatUseCase) OnlineKey(agentID int64) string {
	return fmt.Sprintf("%s%d", agentOnlineKeyPrefix, agentID)
}

// MarkAgentOnline 写入客服在线状态（TTL 保活，由心跳续期）
func (uc *ChatUseCase) MarkAgentOnline(key string) {
	if uc.rdb != nil {
		uc.rdb.Set(context.Background(), key, "1", agentOnlineKeyTTL)
	}
}

// MarkAgentOffline 清除客服在线状态（连接注销）
func (uc *ChatUseCase) MarkAgentOffline(key string) {
	if uc.rdb != nil {
		uc.rdb.Del(context.Background(), key)
	}
}

// Hub 返回连接分发器（供 WebSocket handler 注册/注销连接）
func (uc *ChatUseCase) Hub() *ws.Hub { return uc.hub }

// SendAgentMessage 客服发送消息
func (uc *ChatUseCase) SendAgentMessage(ctx context.Context, agentID, conversationID int64, msgType int32, content, extra string) (*chatRpc.SendAgentMessageReply, error) {
	return uc.cc.SendAgentMessage(ctx, &chatRpc.SendAgentMessageRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
		MsgType:        msgType,
		Content:        content,
		Extra:          extra,
	})
}

// MarkAgentRead 客服标记会话已读
func (uc *ChatUseCase) MarkAgentRead(ctx context.Context, agentID, conversationID int64) error {
	_, err := uc.cc.MarkAgentRead(ctx, &chatRpc.MarkAgentReadRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
	})
	return err
}

// CloseAgentConversation 客服结束会话
func (uc *ChatUseCase) CloseAgentConversation(ctx context.Context, agentID, conversationID int64) error {
	_, err := uc.cc.CloseAgentConversation(ctx, &chatRpc.CloseAgentConversationRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
	})
	return err
}

// ListPendingConversation 待接入会话列表（客服工作台）
func (uc *ChatUseCase) ListPendingConversation(ctx context.Context, agentID int64, page, pageSize int32) (*chatRpc.ListPendingConversationReply, error) {
	return uc.cc.ListPendingConversation(ctx, &chatRpc.ListPendingConversationRequest{
		AgentId:  agentID,
		Page:     page,
		PageSize: pageSize,
	})
}

// AcceptConversation 客服接入会话
func (uc *ChatUseCase) AcceptConversation(ctx context.Context, agentID, conversationID int64) error {
	_, err := uc.cc.AcceptConversation(ctx, &chatRpc.AcceptConversationRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
	})
	return err
}

// TransferConversation 客服转接会话
func (uc *ChatUseCase) TransferConversation(ctx context.Context, agentID, conversationID, targetAgentID int64) error {
	_, err := uc.cc.TransferConversation(ctx, &chatRpc.TransferConversationRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
		TargetAgentId:  targetAgentID,
	})
	return err
}

// ListAgentConversation 客服名下会话列表
func (uc *ChatUseCase) ListAgentConversation(ctx context.Context, agentID int64, page, pageSize, status int32) (*chatRpc.ListAgentConversationReply, error) {
	return uc.cc.ListAgentConversation(ctx, &chatRpc.ListAgentConversationRequest{
		AgentId:  agentID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
}

// ListAgentMessage 客服端历史消息（校验客服成员）
func (uc *ChatUseCase) ListAgentMessage(ctx context.Context, agentID, conversationID, cursorID int64, limit, direction int32) (*chatRpc.ListMessageReply, error) {
	return uc.cc.ListAgentMessage(ctx, &chatRpc.ListAgentMessageRequest{
		AgentId:        agentID,
		ConversationId: conversationID,
		CursorId:       cursorID,
		Limit:          limit,
		Direction:      direction,
	})
}

// AgentStats 客服接待统计
func (uc *ChatUseCase) AgentStats(ctx context.Context, agentID int64) (*chatRpc.AgentStatsReply, error) {
	return uc.cc.AgentStats(ctx, &chatRpc.AgentStatsRequest{AgentId: agentID})
}

// startSubscriber 订阅 NATS chat.message 与 chat.conversation：
//   - chat.message：将消息实时推送给在线客服连接
//   - chat.conversation：广播会话状态变更，客服工作台据此实时刷新待接入/会话列表
func (uc *ChatUseCase) startSubscriber() {
	if uc.nc == nil {
		return
	}
	// 消息推送
	if _, err := uc.nc.Subscribe(event.SubjectChatMessage, func(m *nats.Msg) {
		var evt event.ChatMessageEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			uc.log.Warnf("解析聊天消息事件失败: %v", err)
			return
		}
		// 只处理本网关关心的接收方（客服）
		if evt.ReceiverType != event.ChatReceiverAgent {
			return
		}
		frame := ws.NewServerFrame("message", messageEventToItem(&evt))
		if frame != nil {
			uc.hub.PushToAgent(evt.ReceiverID, frame)
		}
	}); err != nil {
		uc.log.Errorf("订阅聊天消息主题失败: %v", err)
	}
	// 会话状态广播（待接入列表/会话列表实时刷新）
	if _, err := uc.nc.Subscribe(event.SubjectChatConversation, func(m *nats.Msg) {
		var evt event.ChatConversationEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			uc.log.Warnf("解析会话状态事件失败: %v", err)
			return
		}
		frame := ws.NewServerFrame("conversation_status", evt)
		if frame != nil {
			uc.hub.BroadcastAgents(frame)
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
