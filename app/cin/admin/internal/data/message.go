package data

import (
	"context"
	"encoding/json"
	"fmt"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// messageRepo 消息数据仓库（通过 gRPC 调用 user service 的消息中心）
type messageRepo struct {
	data *Data
	log  *log.Helper
}

// NewMessageRepo 创建消息数据仓库
func NewMessageRepo(data *Data, logger log.Logger) biz.MessageRepo {
	return &messageRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// SendMessage 发送消息（全体/指定用户）
func (r *messageRepo) SendMessage(ctx context.Context, m *biz.Message) (int64, error) {
	resp, err := r.data.msgC.CreateMessage(ctx, &userRpc.CreateMessageRequest{
		Title:         m.Title,
		Content:       m.Content,
		MsgType:       m.MsgType,
		TargetType:    m.TargetType,
		TargetUserIds: m.TargetUserIDs,
		CreatedBy:     m.CreatedBy,
	})
	if err != nil {
		r.log.Errorf("CreateMessage failed: %v", err)
		return 0, fmt.Errorf("发送消息失败: %w", err)
	}
	return resp.Id, nil
}

// ListMessage 消息列表（分页筛选）
func (r *messageRepo) ListMessage(ctx context.Context, req *biz.MessageListRequest) (*biz.MessageListReply, error) {
	resp, err := r.data.msgC.AdminListMessage(ctx, &userRpc.AdminListMessageRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		MsgType:  req.MsgType,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		r.log.Errorf("AdminListMessage failed: %v", err)
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}

	items := make([]*biz.Message, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, r.protoToBiz(item))
	}
	return &biz.MessageListReply{
		Items: items,
		Total: resp.Total,
	}, nil
}

// DeleteMessage 下线/删除消息
func (r *messageRepo) DeleteMessage(ctx context.Context, id int64) error {
	if _, err := r.data.msgC.DeleteMessage(ctx, &userRpc.DeleteMessageRequest{Id: id}); err != nil {
		r.log.Errorf("DeleteMessage failed, id=%d: %v", id, err)
		return fmt.Errorf("下线消息失败: %w", err)
	}
	return nil
}

// protoToBiz userRpc.MessageInfo -> biz.Message 转换
func (r *messageRepo) protoToBiz(m *userRpc.MessageInfo) *biz.Message {
	targetIDs := []int64{}
	if m.TargetUserIds != "" {
		_ = json.Unmarshal([]byte(m.TargetUserIds), &targetIDs)
	}
	return &biz.Message{
		ID:            m.Id,
		Title:         m.Title,
		Content:       m.Content,
		MsgType:       m.MsgType,
		CreatedAt:     m.CreatedAt,
		CreatedBy:     m.CreatedBy,
		TargetType:    m.TargetType,
		TargetUserIDs: targetIDs,
		Status:        m.Status,
	}
}
