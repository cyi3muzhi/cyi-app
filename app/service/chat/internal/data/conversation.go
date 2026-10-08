package data

import (
	"context"
	"errors"

	"cyi-app/app/service/chat/internal/biz"
	"cyi-app/app/service/chat/internal/data/curd"
	mod "cyi-app/app/service/chat/internal/data/models"

	v1 "cyi-app/api/service/chat/v1"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type conversationRepo struct {
	data *Data
	log  *log.Helper
}

// NewConversationRepo 创建会话仓储
func NewConversationRepo(data *Data, logger log.Logger) biz.ConversationRepo {
	return &conversationRepo{data: data, log: log.NewHelper(logger)}
}

// CreateConversation 创建会话并回填 ID
func (r *conversationRepo) CreateConversation(ctx context.Context, in *biz.Conversation) error {
	conv := &mod.Conversation{
		UserID:      in.UserID,
		AgentID:     in.AgentID,
		Status:      in.Status,
		ServiceType: in.ServiceType,
		LastMsgType: in.LastMsgType,
	}
	if err := curd.Q.Conversation.WithContext(ctx).Create(conv); err != nil {
		return err
	}
	in.ID = conv.ID
	in.CreateTime = conv.CreateTime
	return nil
}

// GetConversation 按 ID 查询会话
func (r *conversationRepo) GetConversation(ctx context.Context, id int64) (*biz.Conversation, error) {
	conv, err := curd.Q.Conversation.WithContext(ctx).Where(curd.Conversation.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorConversationNotFound("conversation not found")
		}
		return nil, err
	}
	return modelToBiz(conv), nil
}

// GetActiveByUser 查询用户未结束的会话（status=1 或 2）
func (r *conversationRepo) GetActiveByUser(ctx context.Context, userID int64) (*biz.Conversation, error) {
	conv, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.UserID.Eq(userID)).
		Where(curd.Conversation.Status.In(biz.ConversationStatusPending, biz.ConversationStatusActive)).
		Order(curd.Conversation.ID.Desc()).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelToBiz(conv), nil
}

// ListUserConversations 用户会话列表
func (r *conversationRepo) ListUserConversations(ctx context.Context, userID int64, status int32, page, pageSize int32) ([]*biz.Conversation, int32, error) {
	page, pageSize = normalizePage(page, pageSize)
	do := curd.Q.Conversation.WithContext(ctx).Where(curd.Conversation.UserID.Eq(userID))
	if status > 0 {
		do = do.Where(curd.Conversation.Status.Eq(status))
	}

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	convos, err := do.Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order(curd.Conversation.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}
	return listModelToBiz(convos), int32(total), nil
}

// ListPendingConversations 待接入会话列表（按创建时间正序，先来先接）
func (r *conversationRepo) ListPendingConversations(ctx context.Context, page, pageSize int32) ([]*biz.Conversation, int32, error) {
	page, pageSize = normalizePage(page, pageSize)
	do := curd.Q.Conversation.WithContext(ctx).Where(curd.Conversation.Status.Eq(biz.ConversationStatusPending))

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	convos, err := do.Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order(curd.Conversation.ID).Find()
	if err != nil {
		return nil, 0, err
	}
	return listModelToBiz(convos), int32(total), nil
}

// ListAgentConversations 客服名下会话列表
func (r *conversationRepo) ListAgentConversations(ctx context.Context, agentID int64, status, page, pageSize int32) ([]*biz.Conversation, int32, error) {
	page, pageSize = normalizePage(page, pageSize)
	do := curd.Q.Conversation.WithContext(ctx).Where(curd.Conversation.AgentID.Eq(agentID))
	if status > 0 {
		do = do.Where(curd.Conversation.Status.Eq(status))
	}

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	convos, err := do.Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order(curd.Conversation.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}
	return listModelToBiz(convos), int32(total), nil
}

// AcceptConversation 客服接入（仅待接入状态可接入，条件更新防并发抢单）
func (r *conversationRepo) AcceptConversation(ctx context.Context, id, agentID int64) error {
	result, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.ID.Eq(id)).
		Where(curd.Conversation.Status.Eq(biz.ConversationStatusPending)).
		UpdateSimple(
			curd.Conversation.Status.Value(biz.ConversationStatusActive),
			curd.Conversation.AgentID.Value(agentID),
			curd.Conversation.ServiceType.Value(biz.ServiceTypeHuman), // 客服接入后切换人工接待，停止 AI 兜底
		)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorConversationAccepted("conversation already accepted or not pending")
	}
	return nil
}

// TransferConversation 转接会话
func (r *conversationRepo) TransferConversation(ctx context.Context, id, targetAgentID int64) error {
	result, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.ID.Eq(id)).
		Update(curd.Conversation.AgentID, targetAgentID)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorConversationNotFound("conversation not found")
	}
	return nil
}

// CloseConversation 结束会话
func (r *conversationRepo) CloseConversation(ctx context.Context, id, now int64) error {
	result, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.ID.Eq(id)).
		UpdateSimple(
			curd.Conversation.Status.Value(biz.ConversationStatusClosed),
			curd.Conversation.EndedAt.Value(now),
		)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorConversationNotFound("conversation not found")
	}
	return nil
}

// UpdateLastMessage 更新会话最后消息摘要
func (r *conversationRepo) UpdateLastMessage(ctx context.Context, id int64, lastMsgType int32, lastMessage string, at int64) error {
	_, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.ID.Eq(id)).
		UpdateSimple(
			curd.Conversation.LastMsgType.Value(lastMsgType),
			curd.Conversation.LastMessage.Value(lastMessage),
			curd.Conversation.LastMessageAt.Value(at),
		)
	return err
}

// CountTodayByAgent 统计客服今日创建的会话数
func (r *conversationRepo) CountTodayByAgent(ctx context.Context, agentID, startAt int64) (int64, error) {
	count, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.AgentID.Eq(agentID)).
		Where(curd.Conversation.CreateTime.Gte(startAt)).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountProcessingByAgent 统计客服进行中会话数
func (r *conversationRepo) CountProcessingByAgent(ctx context.Context, agentID int64) (int64, error) {
	count, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.AgentID.Eq(agentID)).
		Where(curd.Conversation.Status.Eq(biz.ConversationStatusActive)).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountResolvedTodayByAgent 统计客服今日已解决会话数
func (r *conversationRepo) CountResolvedTodayByAgent(ctx context.Context, agentID, startAt int64) (int64, error) {
	count, err := curd.Q.Conversation.WithContext(ctx).
		Where(curd.Conversation.AgentID.Eq(agentID)).
		Where(curd.Conversation.Status.Eq(biz.ConversationStatusClosed)).
		Where(curd.Conversation.EndedAt.Gte(startAt)).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// modelToBiz 转换会话模型为业务对象
func modelToBiz(conv *mod.Conversation) *biz.Conversation {
	return &biz.Conversation{
		ID:            conv.ID,
		UserID:        conv.UserID,
		AgentID:       conv.AgentID,
		Status:        conv.Status,
		ServiceType:   conv.ServiceType,
		LastMsgType:   conv.LastMsgType,
		LastMessage:   conv.LastMessage,
		LastMessageAt: conv.LastMessageAt,
		EndedAt:       conv.EndedAt,
		CreateTime:    conv.CreateTime,
		UpdateTime:    conv.UpdateTime,
	}
}

// listModelToBiz 批量转换会话模型
func listModelToBiz(convos []*mod.Conversation) []*biz.Conversation {
	items := make([]*biz.Conversation, 0, len(convos))
	for _, c := range convos {
		items = append(items, modelToBiz(c))
	}
	return items
}
