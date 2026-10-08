package data

import (
	"context"
	"sort"
	"time"

	"cyi-app/app/service/chat/internal/biz"
	"cyi-app/app/service/chat/internal/data/curd"
	mod "cyi-app/app/service/chat/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
)

type messageRepo struct {
	data *Data
	log  *log.Helper
}

// NewMessageRepo 创建消息仓储
func NewMessageRepo(data *Data, logger log.Logger) biz.MessageRepo {
	return &messageRepo{data: data, log: log.NewHelper(logger)}
}

// CreateMessage 创建消息并回填 ID 与时间
func (r *messageRepo) CreateMessage(ctx context.Context, in *biz.ChatMessage) error {
	msg := &mod.ChatMessage{
		ConversationID: in.ConversationID,
		SenderType:     in.SenderType,
		SenderID:       in.SenderID,
		MsgType:        in.MsgType,
		Content:        in.Content,
		Extra:          in.Extra,
		IsRead:         in.IsRead,
	}
	if err := curd.Q.ChatMessage.WithContext(ctx).Create(msg); err != nil {
		return err
	}
	in.ID = msg.ID
	in.CreateTime = msg.CreateTime
	return nil
}

// ListMessages 游标分页查询会话消息
// direction: 1=向上（更早） 0=向下（更新）
func (r *messageRepo) ListMessages(ctx context.Context, conversationID, cursorID int64, limit, direction int32) ([]*biz.ChatMessage, bool, error) {
	do := curd.Q.ChatMessage.WithContext(ctx).Where(curd.ChatMessage.ConversationID.Eq(conversationID))
	var (
		items   []*mod.ChatMessage
		err     error
		hasMore bool
	)

	if direction == 1 {
		// 向上翻页：取比游标更早的消息（cursor=0 取最新一页）
		if cursorID > 0 {
			do = do.Where(curd.ChatMessage.ID.Lt(cursorID))
		}
		items, err = do.Order(curd.ChatMessage.ID.Desc()).Limit(int(limit) + 1).Find()
		if err != nil {
			return nil, false, err
		}
		if len(items) > int(limit) {
			hasMore = true
			items = items[:limit]
		}
		// 反转成时间升序返回，便于客户端顺序渲染
		sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	} else {
		// 向下翻页：取比游标更新的消息（需提供游标）
		if cursorID <= 0 {
			return nil, false, nil
		}
		do = do.Where(curd.ChatMessage.ID.Gt(cursorID))
		items, err = do.Order(curd.ChatMessage.ID).Limit(int(limit) + 1).Find()
		if err != nil {
			return nil, false, err
		}
		if len(items) > int(limit) {
			hasMore = true
			items = items[:limit]
		}
	}
	return listMsgToBiz(items), hasMore, nil
}

// MarkReadBySender 将会话中对端（senderType）发送的未读消息全部标记已读
func (r *messageRepo) MarkReadBySender(ctx context.Context, conversationID int64, senderType int32) error {
	_, err := curd.Q.ChatMessage.WithContext(ctx).
		Where(curd.ChatMessage.ConversationID.Eq(conversationID)).
		Where(curd.ChatMessage.SenderType.Eq(senderType)).
		Where(curd.ChatMessage.IsRead.Eq(0)).
		UpdateSimple(
			curd.ChatMessage.IsRead.Value(1),
			curd.ChatMessage.ReadAt.Value(time.Now().Unix()),
		)
	return err
}

// CountUnreadBySender 统计会话中对端（senderType）发送的未读消息数
func (r *messageRepo) CountUnreadBySender(ctx context.Context, conversationID int64, senderType int32) (int32, error) {
	count, err := curd.Q.ChatMessage.WithContext(ctx).
		Where(curd.ChatMessage.ConversationID.Eq(conversationID)).
		Where(curd.ChatMessage.SenderType.Eq(senderType)).
		Where(curd.ChatMessage.IsRead.Eq(0)).
		Count()
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

// listMsgToBiz 批量转换消息模型
func listMsgToBiz(msgs []*mod.ChatMessage) []*biz.ChatMessage {
	items := make([]*biz.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, &biz.ChatMessage{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			SenderType:     m.SenderType,
			SenderID:       m.SenderID,
			MsgType:        m.MsgType,
			Content:        m.Content,
			Extra:          m.Extra,
			IsRead:         m.IsRead,
			ReadAt:         m.ReadAt,
			CreateTime:     m.CreateTime,
		})
	}
	return items
}
