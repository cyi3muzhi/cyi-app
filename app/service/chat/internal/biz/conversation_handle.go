package biz

import (
	"context"
	"strconv"
	"time"

	v1 "cyi-app/api/service/chat/v1"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// ConversationRepo 会话数据仓储接口
type ConversationRepo interface {
	// CreateConversation 创建会话，回填 ID
	CreateConversation(ctx context.Context, in *Conversation) error
	// GetConversation 按 ID 查询会话
	GetConversation(ctx context.Context, id int64) (*Conversation, error)
	// GetActiveByUser 查询用户未结束的会话（用于防重复建会话）
	GetActiveByUser(ctx context.Context, userID int64) (*Conversation, error)
	// ListUserConversations 用户会话列表
	ListUserConversations(ctx context.Context, userID int64, status int32, page, pageSize int32) ([]*Conversation, int32, error)
	// ListPendingConversations 待接入会话列表
	ListPendingConversations(ctx context.Context, page, pageSize int32) ([]*Conversation, int32, error)
	// ListAgentConversations 客服名下会话列表
	ListAgentConversations(ctx context.Context, agentID int64, status, page, pageSize int32) ([]*Conversation, int32, error)
	// AcceptConversation 客服接入（仅待接入状态可接入，防止并发抢单）
	AcceptConversation(ctx context.Context, id, agentID int64) error
	// TransferConversation 转接会话
	TransferConversation(ctx context.Context, id, targetAgentID int64) error
	// CloseConversation 结束会话
	CloseConversation(ctx context.Context, id int64, now int64) error
	// UpdateLastMessage 更新会话最后一条消息摘要
	UpdateLastMessage(ctx context.Context, id int64, lastMsgType int32, lastMessage string, at int64) error
	// CountTodayByAgent 统计客服今日（created_at >= startAt）会话数
	CountTodayByAgent(ctx context.Context, agentID, startAt int64) (int64, error)
	// CountProcessingByAgent 统计客服进行中会话数
	CountProcessingByAgent(ctx context.Context, agentID int64) (int64, error)
	// CountResolvedTodayByAgent 统计客服今日已解决（ended_at >= startAt）会话数
	CountResolvedTodayByAgent(ctx context.Context, agentID, startAt int64) (int64, error)
}

// ConversationEventPublisher 会话状态事件发布器（data 层实现：发布到 NATS chat.conversation）
type ConversationEventPublisher interface {
	// PublishConversation 发布一条会话状态事件
	PublishConversation(ctx context.Context, evt *event.ChatConversationEvent)
}

// 在线状态 Redis key 前缀（网关 WS 连接时写入并心跳续期）
const (
	onlineUserKeyPrefix  = "chat:online:user:"
	onlineAgentKeyPrefix = "chat:online:agent:"
)

// ConversationUseCase 会话业务用例（用户端会话 + 客服端接待操作）
type ConversationUseCase struct {
	repo    ConversationRepo
	msgRepo MessageRepo
	convPub ConversationEventPublisher
	rdb     redis.UniversalClient
	log     *log.Helper
}

// NewConversationUseCase 创建会话用例
func NewConversationUseCase(repo ConversationRepo, msgRepo MessageRepo, convPub ConversationEventPublisher, rdb redis.UniversalClient, logger log.Logger) *ConversationUseCase {
	return &ConversationUseCase{repo: repo, msgRepo: msgRepo, convPub: convPub, rdb: rdb, log: log.NewHelper(logger)}
}

// CreateConversation 创建会话：用户存在未结束会话则直接复用，否则新建待接入会话。
// 注：自动分配在线客服与 AI 兜底在 M4 实现，M3 统一进入待接入。
func (uc *ConversationUseCase) CreateConversation(ctx context.Context, req *v1.CreateConversationRequest) (*v1.CreateConversationReply, error) {
	// 防重复：复用用户未结束的会话
	existing, err := uc.repo.GetActiveByUser(ctx, req.UserId)
	if err == nil && existing != nil {
		unread := uc.unreadCount(ctx, existing.ID, event.ChatReceiverUser, SenderTypeAgent)
		return &v1.CreateConversationReply{Conversation: uc.toInfo(existing, unread, uc.isUserOnline(ctx, existing.UserID))}, nil
	}

	conv := &Conversation{
		UserID:      req.UserId,
		AgentID:     0,
		Status:      ConversationStatusPending,
		ServiceType: ServiceTypeAI, // M4：新会话默认 AI 兜底接待，客服接入后切换人工
		LastMsgType: MsgTypeText,
	}
	if err := uc.repo.CreateConversation(ctx, conv); err != nil {
		return nil, v1.ErrorConversationNotFound("create conversation failed")
	}
	// 发布会话新建事件（客服工作台待接入列表实时刷新）
	uc.publishConvEvent(ctx, conv.ID, event.ChatEventCreated, conv.UserID, 0)
	return &v1.CreateConversationReply{Conversation: uc.toInfo(conv, 0, uc.isUserOnline(ctx, conv.UserID))}, nil
}

// ListMyConversation 我的会话列表（用户端）
func (uc *ConversationUseCase) ListMyConversation(ctx context.Context, req *v1.ListMyConversationRequest) (*v1.ListMyConversationReply, error) {
	items, total, err := uc.repo.ListUserConversations(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("list conversation failed")
	}
	infos := make([]*v1.ConversationInfo, 0, len(items))
	for _, c := range items {
		unread := uc.unreadCount(ctx, c.ID, event.ChatReceiverUser, SenderTypeAgent)
		infos = append(infos, uc.toInfo(c, unread, uc.isUserOnline(ctx, c.UserID)))
	}
	return &v1.ListMyConversationReply{Items: infos, Total: total}, nil
}

// ListPendingConversation 待接入会话列表（客服工作台）
func (uc *ConversationUseCase) ListPendingConversation(ctx context.Context, req *v1.ListPendingConversationRequest) (*v1.ListPendingConversationReply, error) {
	items, total, err := uc.repo.ListPendingConversations(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("list pending conversation failed")
	}
	infos := make([]*v1.ConversationInfo, 0, len(items))
	for _, c := range items {
		unread := uc.unreadCount(ctx, c.ID, event.ChatReceiverAgent, SenderTypeUser)
		infos = append(infos, uc.toInfo(c, unread, uc.isUserOnline(ctx, c.UserID)))
	}
	return &v1.ListPendingConversationReply{Items: infos, Total: total}, nil
}

// AcceptConversation 客服接入会话
func (uc *ConversationUseCase) AcceptConversation(ctx context.Context, req *v1.AcceptConversationRequest) (*v1.AcceptConversationReply, error) {
	if err := uc.repo.AcceptConversation(ctx, req.ConversationId, req.AgentId); err != nil {
		return nil, v1.ErrorConversationAccepted("conversation already accepted or not pending")
	}
	// 发布接入事件（用户端收到客服接入通知，客服工作台移除待接入项）
	uc.publishConvEvent(ctx, req.ConversationId, event.ChatEventAccepted, 0, req.AgentId)
	return &v1.AcceptConversationReply{Success: true}, nil
}

// TransferConversation 客服转接会话
func (uc *ConversationUseCase) TransferConversation(ctx context.Context, req *v1.TransferConversationRequest) (*v1.TransferConversationReply, error) {
	conv, err := uc.repo.GetConversation(ctx, req.ConversationId)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("conversation not found")
	}
	// 仅当前接待客服可转接
	if conv.AgentID != req.AgentId || conv.Status != ConversationStatusActive {
		return nil, v1.ErrorNotConversationMember("only the serving agent can transfer")
	}
	if err := uc.repo.TransferConversation(ctx, req.ConversationId, req.TargetAgentId); err != nil {
		return nil, v1.ErrorConversationNotFound("transfer conversation failed")
	}
	// 发布转接事件
	uc.publishConvEvent(ctx, req.ConversationId, event.ChatEventTransferred, conv.UserID, req.TargetAgentId)
	return &v1.TransferConversationReply{Success: true}, nil
}

// ListAgentConversation 客服名下会话列表
func (uc *ConversationUseCase) ListAgentConversation(ctx context.Context, req *v1.ListAgentConversationRequest) (*v1.ListAgentConversationReply, error) {
	items, total, err := uc.repo.ListAgentConversations(ctx, req.AgentId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("list agent conversation failed")
	}
	infos := make([]*v1.ConversationInfo, 0, len(items))
	for _, c := range items {
		unread := uc.unreadCount(ctx, c.ID, event.ChatReceiverAgent, SenderTypeUser)
		infos = append(infos, uc.toInfo(c, unread, uc.isUserOnline(ctx, c.UserID)))
	}
	return &v1.ListAgentConversationReply{Items: infos, Total: total}, nil
}

// CloseConversation 结束会话（用户端）
func (uc *ConversationUseCase) CloseConversation(ctx context.Context, req *v1.CloseConversationRequest) (*v1.CloseConversationReply, error) {
	if err := uc.close(ctx, req.UserId, req.ConversationId, true); err != nil {
		return nil, err
	}
	return &v1.CloseConversationReply{Success: true}, nil
}

// CloseAgentConversation 结束会话（客服端）
func (uc *ConversationUseCase) CloseAgentConversation(ctx context.Context, req *v1.CloseAgentConversationRequest) (*v1.CloseAgentConversationReply, error) {
	if err := uc.close(ctx, req.AgentId, req.ConversationId, false); err != nil {
		return nil, err
	}
	return &v1.CloseAgentConversationReply{Success: true}, nil
}

// close 关闭会话：isUser 为 true 时校验用户成员，否则校验客服成员
func (uc *ConversationUseCase) close(ctx context.Context, memberID, conversationID int64, isUser bool) error {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return v1.ErrorConversationNotFound("conversation not found")
	}
	if isUser && conv.UserID != memberID {
		return v1.ErrorNotConversationMember("not conversation member")
	}
	if !isUser && conv.AgentID != memberID {
		return v1.ErrorNotConversationMember("not conversation member")
	}
	if conv.Status == ConversationStatusClosed {
		return nil
	}
	if err := uc.repo.CloseConversation(ctx, conversationID, time.Now().Unix()); err != nil {
		return err
	}
	// 清理未读数（Redis）与发布结束事件
	DelConvUnread(ctx, uc.rdb, conversationID)
	// 将用户发送的消息标记已读，避免 unreadCount 的 DB 兜底把已结束会话重新算成未读
	_ = uc.msgRepo.MarkReadBySender(ctx, conversationID, SenderTypeUser)
	uc.publishConvEvent(ctx, conversationID, event.ChatEventClosed, conv.UserID, conv.AgentID)
	return nil
}

// AgentStats 客服接待统计
func (uc *ConversationUseCase) AgentStats(ctx context.Context, req *v1.AgentStatsRequest) (*v1.AgentStatsReply, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	totalToday, err := uc.repo.CountTodayByAgent(ctx, req.AgentId, startOfDay)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("stat conversation failed")
	}
	processing, err := uc.repo.CountProcessingByAgent(ctx, req.AgentId)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("stat conversation failed")
	}
	resolvedToday, err := uc.repo.CountResolvedTodayByAgent(ctx, req.AgentId, startOfDay)
	if err != nil {
		return nil, v1.ErrorConversationNotFound("stat conversation failed")
	}
	return &v1.AgentStatsReply{
		TotalToday:    totalToday,
		Processing:    processing,
		ResolvedToday: resolvedToday,
	}, nil
}

// toInfo 转换会话业务对象为 proto 信息
func (uc *ConversationUseCase) toInfo(c *Conversation, unread, userOnline int32) *v1.ConversationInfo {
	return &v1.ConversationInfo{
		Id:            c.ID,
		UserId:        c.UserID,
		AgentId:       c.AgentID,
		Status:        c.Status,
		ServiceType:   c.ServiceType,
		LastMessage:   c.LastMessage,
		LastMsgType:   c.LastMsgType,
		LastMessageAt: c.LastMessageAt,
		UnreadCount:   unread,
		UserOnline:    userOnline,
		CreatedAt:     c.CreateTime,
	}
}

// publishConvEvent 发布会话状态事件（供客服工作台实时刷新与用户端状态通知）
func (uc *ConversationUseCase) publishConvEvent(ctx context.Context, conversationID int64, eventType int32, userID, agentID int64) {
	if uc.convPub == nil {
		return
	}
	uc.convPub.PublishConversation(ctx, &event.ChatConversationEvent{
		ConversationID: conversationID,
		EventType:      eventType,
		UserID:         userID,
		AgentID:        agentID,
	})
}

// isUserOnline 查询用户是否在线（Redis chat:online:user:{uid}，由网关 WS 连接维护）
func (uc *ConversationUseCase) isUserOnline(ctx context.Context, userID int64) int32 {
	if uc.rdb == nil || userID <= 0 {
		return 0
	}
	n, err := uc.rdb.Exists(ctx, onlineUserKeyPrefix+strconv.FormatInt(userID, 10)).Result()
	if err != nil {
		return 0
	}
	if n > 0 {
		return 1
	}
	return 0
}

// unreadCount 读取会话未读数：Redis 优先，Redis 无记录（如服务重启）回退 DB 统计。
// receiverType 指定读取哪一侧（user=客服发给用户的消息、agent=用户发给客服的消息）；
// fallbackSenderType 为 DB 兜底统计的发送者类型。
func (uc *ConversationUseCase) unreadCount(ctx context.Context, conversationID int64, receiverType, fallbackSenderType int32) int32 {
	if v, ok := GetUnread(ctx, uc.rdb, conversationID, receiverType); ok {
		return v
	}
	if n, err := uc.msgRepo.CountUnreadBySender(ctx, conversationID, fallbackSenderType); err == nil {
		return n
	}
	return 0
}
