package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Message 消息业务对象
type Message struct {
	ID             int64
	Title          string  // 标题
	Content        string  // 内容（支持换行文本）
	MsgType        int32   // 消息类型 1=订单 2=系统 3=营销
	IsRead         int32   // 是否已读 0=未读 1=已读（用户端视角）
	CreatedAt      int64   // 创建时间（Unix 秒）
	CreatedBy      int64   // 创建人（管理员ID）
	TargetType     int32   // 发送范围 0=全体用户 1=指定用户
	TargetUserIDs  []int64 // 指定用户ID列表（仅 target_type=1 时有值）
	Status         int32   // 消息状态 1=启用 2=下线
	ReadAt         int64   // 已读时间（Unix 秒，未读为 0）
	ConversationID int64   // 关联会话ID（IM 客服离线提醒，0=普通通知）
}

// MessageRepo 消息数据仓库接口
type MessageRepo interface {
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, m *Message) (int64, error)
	// AdminListMessage 管理后台消息列表（分页）
	AdminListMessage(ctx context.Context, page, pageSize int32, msgType, status int32, keyword string) ([]*Message, int32, error)
	// DeleteMessage 下线/删除消息（软删除）
	DeleteMessage(ctx context.Context, id int64) error
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
	repo MessageRepo
	log  *log.Helper
}

// NewMessageUseCase 创建 MessageUseCase
func NewMessageUseCase(repo MessageRepo, logger log.Logger) *MessageUseCase {
	return &MessageUseCase{repo: repo, log: log.NewHelper(logger)}
}

// CreateMessage 创建消息
func (uc *MessageUseCase) CreateMessage(ctx context.Context, m *Message) (int64, error) {
	id, err := uc.repo.CreateMessage(ctx, m)
	if err != nil {
		uc.log.Errorf("CreateMessage err: %v", err)
		return 0, ErrMessageCreateFailed
	}
	return id, nil
}

// AdminListMessage 管理后台消息列表
func (uc *MessageUseCase) AdminListMessage(ctx context.Context, page, pageSize int32, msgType, status int32, keyword string) ([]*Message, int32, error) {
	return uc.repo.AdminListMessage(ctx, page, pageSize, msgType, status, keyword)
}

// DeleteMessage 下线/删除消息
func (uc *MessageUseCase) DeleteMessage(ctx context.Context, id int64) error {
	if err := uc.repo.DeleteMessage(ctx, id); err != nil {
		uc.log.Errorf("DeleteMessage err: %v", err)
		if err == ErrMessageNotFound {
			return ErrMessageNotFound
		}
		return err
	}
	return nil
}

// ListMyMessage 我的消息列表
func (uc *MessageUseCase) ListMyMessage(ctx context.Context, userID int64, page, pageSize, msgType int32) ([]*Message, int32, error) {
	return uc.repo.ListMyMessage(ctx, userID, page, pageSize, msgType)
}

// UnreadMessageCount 未读消息数
func (uc *MessageUseCase) UnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	return uc.repo.UnreadMessageCount(ctx, userID)
}

// MarkMessageRead 标记单条消息已读
func (uc *MessageUseCase) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	if err := uc.repo.MarkMessageRead(ctx, userID, messageID); err != nil {
		uc.log.Errorf("MarkMessageRead err: %v", err)
		return err
	}
	return nil
}

// MarkAllMessageRead 全部标记已读
func (uc *MessageUseCase) MarkAllMessageRead(ctx context.Context, userID int64) error {
	if err := uc.repo.MarkAllMessageRead(ctx, userID); err != nil {
		uc.log.Errorf("MarkAllMessageRead err: %v", err)
		return err
	}
	return nil
}
