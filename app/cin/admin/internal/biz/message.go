package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Message 消息业务对象（管理后台视角）
type Message struct {
	ID            int64
	Title         string  // 标题
	Content       string  // 内容（支持换行文本）
	MsgType       int32   // 消息类型 1=订单 2=系统 3=营销
	CreatedAt     int64   // 创建时间（Unix 秒）
	CreatedBy     int64   // 创建人（管理员ID）
	TargetType    int32   // 发送范围 0=全体用户 1=指定用户
	TargetUserIDs []int64 // 指定用户ID列表（仅 target_type=1 时有值）
	Status        int32   // 消息状态 1=启用 2=下线
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	Page     int32
	PageSize int32
	MsgType  int32 // 0=全部
	Status   int32 // 0=全部 1=启用 2=下线
	Keyword  string
}

// MessageListReply 消息列表响应
type MessageListReply struct {
	Items []*Message
	Total int32
}

// MessageRepo 消息数据仓库接口
type MessageRepo interface {
	// SendMessage 发送消息（全体/指定用户）
	SendMessage(ctx context.Context, m *Message) (int64, error)
	// ListMessage 消息列表（分页筛选）
	ListMessage(ctx context.Context, req *MessageListRequest) (*MessageListReply, error)
	// DeleteMessage 下线/删除消息
	DeleteMessage(ctx context.Context, id int64) error
}

// MessageUseCase 消息业务逻辑
type MessageUseCase struct {
	repo MessageRepo
	log  *log.Helper
}

// NewMessageUseCase 创建 MessageUseCase
func NewMessageUseCase(repo MessageRepo, logger log.Logger) *MessageUseCase {
	return &MessageUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// SendMessage 发送消息
func (uc *MessageUseCase) SendMessage(ctx context.Context, m *Message) (int64, error) {
	return uc.repo.SendMessage(ctx, m)
}

// ListMessage 消息列表
func (uc *MessageUseCase) ListMessage(ctx context.Context, req *MessageListRequest) (*MessageListReply, error) {
	return uc.repo.ListMessage(ctx, req)
}

// DeleteMessage 下线/删除消息
func (uc *MessageUseCase) DeleteMessage(ctx context.Context, id int64) error {
	return uc.repo.DeleteMessage(ctx, id)
}
