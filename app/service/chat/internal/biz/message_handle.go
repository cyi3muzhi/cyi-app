package biz

import (
	"context"
	"fmt"
	"time"

	v1 "cyi-app/api/service/chat/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// 离线提醒去重 key：chat:remind:{conversationID}:{userID}
// 同一会话对同一用户只发一条站内信提醒，用户标记已读/打开会话后清除
const (
	remindKeyPrefix = "chat:remind:"
	remindKeyTTL    = 7 * 24 * time.Hour
)

// MessageRepo 消息数据仓储接口
type MessageRepo interface {
	// CreateMessage 创建消息，回填 ID
	CreateMessage(ctx context.Context, in *ChatMessage) error
	// ListMessages 游标分页查询会话消息
	// direction: 0=向下（比 cursor 更新） 1=向上（比 cursor 更早）；cursor=0 时取最新一页
	ListMessages(ctx context.Context, conversationID, cursorID int64, limit, direction int32) ([]*ChatMessage, bool, error)
	// MarkReadBySender 将会话中对端（senderType）发送的消息全部标记已读
	MarkReadBySender(ctx context.Context, conversationID int64, senderType int32) error
	// CountUnreadBySender 统计会话中对端（senderType）发送的未读消息数
	CountUnreadBySender(ctx context.Context, conversationID int64, senderType int32) (int32, error)
}

// ChatMessagePublisher 聊天消息事件发布器（data 层实现：落库后发布到 NATS chat.message）
type ChatMessagePublisher interface {
	// PublishMessage 发布一条消息事件
	PublishMessage(ctx context.Context, evt *event.ChatMessageEvent)
}

// MessageUseCase 消息业务用例（收发消息、历史、已读）
type MessageUseCase struct {
	convRepo  ConversationRepo
	msgRepo   MessageRepo
	publisher ChatMessagePublisher
	aiUC      *AIChatUseCase
	msgC      userRpc.MessageClient // 用户服务消息中心（离线提醒发站内信用）
	rdb       redis.UniversalClient
	log       *log.Helper
}

// NewMessageUseCase 创建消息用例
func NewMessageUseCase(convRepo ConversationRepo, msgRepo MessageRepo, publisher ChatMessagePublisher, aiUC *AIChatUseCase, msgC userRpc.MessageClient, rdb redis.UniversalClient, logger log.Logger) *MessageUseCase {
	return &MessageUseCase{convRepo: convRepo, msgRepo: msgRepo, publisher: publisher, aiUC: aiUC, msgC: msgC, rdb: rdb, log: log.NewHelper(logger)}
}

// SendMessage 用户发送消息
func (uc *MessageUseCase) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageReply, error) {
	if err := uc.checkMember(ctx, req.ConversationId, req.UserId, true); err != nil {
		return nil, err
	}
	reply, err := uc.send(ctx, req.ConversationId, SenderTypeUser, req.UserId, req.MsgType, req.Content, req.Extra)
	if err != nil {
		return nil, err
	}
	// AI 兜底：会话处于 AI 接待中时异步生成回复
	if uc.aiUC != nil && req.MsgType == MsgTypeText {
		uc.aiUC.HandleAIReply(ctx, req.ConversationId, req.UserId, req.Content)
	}
	return reply, nil
}

// SendAgentMessage 客服发送消息
func (uc *MessageUseCase) SendAgentMessage(ctx context.Context, req *v1.SendAgentMessageRequest) (*v1.SendAgentMessageReply, error) {
	if err := uc.checkMember(ctx, req.ConversationId, req.AgentId, false); err != nil {
		return nil, err
	}
	reply, err := uc.send(ctx, req.ConversationId, SenderTypeAgent, req.AgentId, req.MsgType, req.Content, req.Extra)
	if err != nil {
		return nil, err
	}
	// 离线提醒：用户不在线时通过消息中心发站内信（去重）
	uc.remindOffline(ctx, req.ConversationId)
	return &v1.SendAgentMessageReply{MessageId: reply.MessageId, CreatedAt: reply.CreatedAt}, nil
}

// send 落库消息、更新会话摘要并发布推送事件
func (uc *MessageUseCase) send(ctx context.Context, conversationID int64, senderType int32, senderID int64, msgType int32, content, extra string) (*v1.SendMessageReply, error) {
	now := time.Now().Unix()
	msg := &ChatMessage{
		ConversationID: conversationID,
		SenderType:     senderType,
		SenderID:       senderID,
		MsgType:        msgType,
		Content:        content,
		Extra:          extra,
		IsRead:         0,
	}
	if err := uc.msgRepo.CreateMessage(ctx, msg); err != nil {
		return nil, v1.ErrorConversationNotFound("create message failed")
	}
	// 更新会话最后消息摘要（超长截断）
	summary := content
	if len([]rune(summary)) > 100 {
		summary = string([]rune(summary)[:100]) + "..."
	}
	if err := uc.convRepo.UpdateLastMessage(ctx, conversationID, msgType, summary, now); err != nil {
		// 摘要更新失败不影响消息返回，记录日志
		uc.log.Warnf("update last message failed: %v", err)
	}
	// 发布推送事件（解析对端接收者）
	uc.publishEvent(ctx, conversationID, msg.ID, senderType, senderID, msgType, content, extra, now)
	return &v1.SendMessageReply{MessageId: msg.ID, CreatedAt: now}, nil
}

// publishEvent 组装并发布消息推送事件（用户/客服互发时才推送，AI/系统消息由后续里程碑处理）
func (uc *MessageUseCase) publishEvent(ctx context.Context, conversationID, msgID int64, senderType int32, senderID int64, msgType int32, content, extra string, createdAt int64) {
	if uc.publisher == nil {
		return
	}
	conv, err := uc.convRepo.GetConversation(ctx, conversationID)
	if err != nil {
		uc.log.Warnf("publish event: get conversation failed: %v", err)
		return
	}
	var receiverType int32
	var receiverID int64
	switch senderType {
	case SenderTypeUser:
		receiverType = event.ChatReceiverAgent
		receiverID = conv.AgentID
	case SenderTypeAgent:
		receiverType = event.ChatReceiverUser
		receiverID = conv.UserID
	default:
		return
	}
	// 对端未读 +1（Redis）
	IncrUnread(ctx, uc.rdb, conversationID, receiverType)
	uc.publisher.PublishMessage(ctx, &event.ChatMessageEvent{
		ConversationID: conversationID,
		MessageID:      msgID,
		SenderType:     senderType,
		SenderID:       senderID,
		ReceiverType:   receiverType,
		ReceiverID:     receiverID,
		MsgType:        msgType,
		Content:        content,
		Extra:          extra,
		CreatedAt:      createdAt,
	})
}

// ListMessage 历史消息（游标分页，校验成员）
func (uc *MessageUseCase) ListMessage(ctx context.Context, req *v1.ListMessageRequest) (*v1.ListMessageReply, error) {
	if err := uc.checkMember(ctx, req.ConversationId, req.UserId, true); err != nil {
		return nil, err
	}
	return uc.list(ctx, req.ConversationId, req.CursorId, req.Limit, req.Direction)
}

// ListAgentMessage 历史消息（客服端，校验客服成员）
func (uc *MessageUseCase) ListAgentMessage(ctx context.Context, req *v1.ListAgentMessageRequest) (*v1.ListMessageReply, error) {
	if err := uc.checkAgentCanView(ctx, req.ConversationId, req.AgentId); err != nil {
		return nil, err
	}
	return uc.list(ctx, req.ConversationId, req.CursorId, req.Limit, req.Direction)
}

// list 游标分页查询消息并转换为 proto 信息
func (uc *MessageUseCase) list(ctx context.Context, conversationID, cursorID int64, limit, direction int32) (*v1.ListMessageReply, error) {
	items, hasMore, err := uc.msgRepo.ListMessages(ctx, conversationID, cursorID, limit, direction)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("list message failed")
	}
	infos := make([]*v1.ChatMessageInfo, 0, len(items))
	for _, m := range items {
		infos = append(infos, &v1.ChatMessageInfo{
			Id:             m.ID,
			ConversationId: m.ConversationID,
			SenderType:     m.SenderType,
			SenderId:       m.SenderID,
			MsgType:        m.MsgType,
			Content:        m.Content,
			Extra:          m.Extra,
			IsRead:         m.IsRead,
			CreatedAt:      m.CreateTime,
		})
	}
	return &v1.ListMessageReply{Items: infos, HasMore: hasMore}, nil
}

// MarkConversationRead 用户标记会话已读（把客服发送的消息标已读，并清除离线提醒标记）
func (uc *MessageUseCase) MarkConversationRead(ctx context.Context, req *v1.MarkConversationReadRequest) (*v1.MarkConversationReadReply, error) {
	if err := uc.checkMember(ctx, req.ConversationId, req.UserId, true); err != nil {
		return nil, err
	}
	if err := uc.msgRepo.MarkReadBySender(ctx, req.ConversationId, SenderTypeAgent); err != nil {
		return nil, v1.ErrorConversationNotFound("mark read failed")
	}
	// 未读清零（Redis）+ 清除离线提醒标记
	ClearUnread(ctx, uc.rdb, req.ConversationId, event.ChatReceiverUser)
	if uc.rdb != nil {
		uc.rdb.Del(ctx, fmt.Sprintf("%s%d:%d", remindKeyPrefix, req.ConversationId, req.UserId))
	}
	return &v1.MarkConversationReadReply{Success: true}, nil
}

// MarkAgentRead 客服标记会话已读（把用户发送的消息标已读）
func (uc *MessageUseCase) MarkAgentRead(ctx context.Context, req *v1.MarkAgentReadRequest) (*v1.MarkAgentReadReply, error) {
	// 与历史查看一致：待接入会话（agent_id=0）任何客服可查看并标记已读；进行中仅接待客服
	if err := uc.checkAgentCanView(ctx, req.ConversationId, req.AgentId); err != nil {
		return nil, err
	}
	if err := uc.msgRepo.MarkReadBySender(ctx, req.ConversationId, SenderTypeUser); err != nil {
		return nil, v1.ErrorConversationNotFound("mark read failed")
	}
	// 未读清零（Redis）
	ClearUnread(ctx, uc.rdb, req.ConversationId, event.ChatReceiverAgent)
	return &v1.MarkAgentReadReply{Success: true}, nil
}

// checkMember 校验请求方是否为会话成员且会话未结束
func (uc *MessageUseCase) checkMember(ctx context.Context, conversationID, memberID int64, isUser bool) error {
	conv, err := uc.convRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return v1.ErrorConversationNotFound("conversation not found")
	}
	if conv.Status == ConversationStatusClosed {
		return v1.ErrorConversationClosed("conversation closed")
	}
	if isUser && conv.UserID != memberID {
		return v1.ErrorNotConversationMember("not conversation member")
	}
	if !isUser && conv.AgentID != memberID {
		return v1.ErrorNotConversationMember("not conversation member")
	}
	return nil
}

// checkAgentCanView 客服端历史查看校验：
// 待接入会话（agent_id=0，工作台预览/接待前查看）任何客服可查看；
// 进行中与已结束会话仅接待客服可查看（已结束供客服回看聊天记录）。
func (uc *MessageUseCase) checkAgentCanView(ctx context.Context, conversationID, agentID int64) error {
	conv, err := uc.convRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return v1.ErrorConversationNotFound("conversation not found")
	}
	// 待接入（未分配客服）任何客服可查看；进行中 / 已结束仅接待客服
	if conv.AgentID != 0 && conv.AgentID != agentID {
		return v1.ErrorNotConversationMember("not conversation member")
	}
	return nil
}

// remindOffline 客服发送消息后，若用户不在线则通过消息中心发一条系统站内信提醒。
// 同一会话对同一用户只发一次（Redis key 去重），用户标记已读后清除、可再次提醒。
// 失败不影响客服消息发送主流程，仅记录日志。
func (uc *MessageUseCase) remindOffline(ctx context.Context, conversationID int64) {
	if uc.rdb == nil || uc.msgC == nil {
		return
	}
	// 会话信息（取用户ID）
	conv, err := uc.convRepo.GetConversation(ctx, conversationID)
	if err != nil || conv == nil || conv.UserID <= 0 {
		return
	}
	// 用户在线则不提醒（在线状态由 interface WS 维护 chat:online:user:{uid}）
	if uc.isUserOnline(ctx, conv.UserID) {
		return
	}
	// 同会话去重
	key := fmt.Sprintf("%s%d:%d", remindKeyPrefix, conversationID, conv.UserID)
	ok, err := uc.rdb.SetNX(ctx, key, "1", remindKeyTTL).Result()
	if err != nil || !ok {
		return
	}
	// 调用户服务消息中心发送系统提醒（带会话ID，供前端跳转会话）
	if _, err := uc.msgC.CreateMessage(ctx, &userRpc.CreateMessageRequest{
		Title:          "您有一条新的客服回复",
		Content:        "客服已回复您的咨询，请打开会话查看。",
		MsgType:        2, // 系统
		TargetType:     1, // 指定用户
		TargetUserIds:  []int64{conv.UserID},
		CreatedBy:      0,
		ConversationId: conversationID,
	}); err != nil {
		uc.log.Warnf("offline remind create message failed: %v", err)
	}
}

// isUserOnline 查询用户是否在线（Redis chat:online:user:{uid}）
func (uc *MessageUseCase) isUserOnline(ctx context.Context, userID int64) bool {
	if uc.rdb == nil || userID <= 0 {
		return false
	}
	n, err := uc.rdb.Exists(ctx, onlineUserKeyPrefix+fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return false
	}
	return n > 0
}
