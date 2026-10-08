package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Message 消息业务对象（用户端视角）
type Message struct {
	ID             int64
	Title          string // 标题
	Content        string // 内容（支持换行文本）
	MsgType        int32  // 消息类型 1=订单 2=系统 3=营销
	IsRead         int32  // 是否已读 0=未读 1=已读
	CreatedAt      int64  // 创建时间（Unix 秒）
	ReadAt         int64  // 已读时间（Unix 秒，未读为 0）
	ConversationID int64  // 关联会话ID（IM 客服离线提醒，0=普通通知）
}

// MessageRepo 消息数据仓库接口
type MessageRepo interface {
	// ListMyMessage 我的消息列表（分页 + 类型筛选）
	ListMyMessage(ctx context.Context, userID int64, page, pageSize, msgType int32) ([]*Message, int32, error)
	// UnreadMessageCount 未读消息数
	UnreadMessageCount(ctx context.Context, userID int64) (int64, error)
	// MarkMessageRead 标记单条消息已读
	MarkMessageRead(ctx context.Context, userID, messageID int64) error
	// MarkAllMessageRead 全部标记已读
	MarkAllMessageRead(ctx context.Context, userID int64) error
}

// MessageUseCase 消息业务逻辑
type MessageUseCase struct {
	log  *log.Helper
	repo MessageRepo
}

// NewMessageUseCase 创建 MessageUseCase
func NewMessageUseCase(repo MessageRepo, logger log.Logger) *MessageUseCase {
	return &MessageUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// ListMyMessage 我的消息列表
func (u *MessageUseCase) ListMyMessage(ctx context.Context, userID int64, page, pageSize, msgType int32) ([]*Message, int32, error) {
	items, total, err := u.repo.ListMyMessage(ctx, userID, page, pageSize, msgType)
	if err != nil {
		u.log.Errorf("ListMyMessage err: %v", err)
		return nil, 0, err
	}
	return items, total, nil
}

// UnreadMessageCount 未读消息数
func (u *MessageUseCase) UnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	count, err := u.repo.UnreadMessageCount(ctx, userID)
	if err != nil {
		u.log.Errorf("UnreadMessageCount err: %v", err)
		return 0, err
	}
	return count, nil
}

// MarkMessageRead 标记单条消息已读
func (u *MessageUseCase) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	if err := u.repo.MarkMessageRead(ctx, userID, messageID); err != nil {
		u.log.Errorf("MarkMessageRead err: %v", err)
		return err
	}
	return nil
}

// MarkAllMessageRead 全部标记已读
func (u *MessageUseCase) MarkAllMessageRead(ctx context.Context, userID int64) error {
	if err := u.repo.MarkAllMessageRead(ctx, userID); err != nil {
		u.log.Errorf("MarkAllMessageRead err: %v", err)
		return err
	}
	return nil
}
