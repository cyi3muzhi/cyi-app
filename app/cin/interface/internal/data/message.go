package data

import (
	"context"

	"cyi-app/app/cin/interface/internal/biz"

	usV1 "cyi-app/api/service/user/v1"

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

// ListMyMessage 我的消息列表
func (r *messageRepo) ListMyMessage(ctx context.Context, userID int64, page, pageSize, msgType int32) ([]*biz.Message, int32, error) {
	reply, err := r.data.mc.ListMyMessage(ctx, &usV1.ListMyMessageRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
		MsgType:  msgType,
	})
	if err != nil {
		r.log.Errorf("ListMyMessage failed, userID=%d: %v", userID, err)
		return nil, 0, err
	}
	items := make([]*biz.Message, 0, len(reply.Items))
	for _, m := range reply.Items {
		items = append(items, &biz.Message{
			ID:        m.Id,
			Title:     m.Title,
			Content:   m.Content,
			MsgType:   m.MsgType,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt,
			ReadAt:    m.ReadAt,
		})
	}
	return items, reply.Total, nil
}

// UnreadMessageCount 未读消息数
func (r *messageRepo) UnreadMessageCount(ctx context.Context, userID int64) (int64, error) {
	reply, err := r.data.mc.UnreadMessageCount(ctx, &usV1.UnreadMessageCountRequest{UserId: userID})
	if err != nil {
		r.log.Errorf("UnreadMessageCount failed, userID=%d: %v", userID, err)
		return 0, err
	}
	return reply.Count, nil
}

// MarkMessageRead 标记单条消息已读
func (r *messageRepo) MarkMessageRead(ctx context.Context, userID, messageID int64) error {
	if _, err := r.data.mc.MarkMessageRead(ctx, &usV1.MarkMessageReadRequest{
		UserId:    userID,
		MessageId: messageID,
	}); err != nil {
		r.log.Errorf("MarkMessageRead failed, userID=%d, messageID=%d: %v", userID, messageID, err)
		return err
	}
	return nil
}

// MarkAllMessageRead 全部标记已读
func (r *messageRepo) MarkAllMessageRead(ctx context.Context, userID int64) error {
	if _, err := r.data.mc.MarkAllMessageRead(ctx, &usV1.MarkAllMessageReadRequest{UserId: userID}); err != nil {
		r.log.Errorf("MarkAllMessageRead failed, userID=%d: %v", userID, err)
		return err
	}
	return nil
}
