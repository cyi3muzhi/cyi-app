package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"

	service "cyi-app/pkg/kratos/service"
)

// ListMyMessage 我的消息列表
func (s *CinInterface) ListMyMessage(ctx context.Context, req *v1.ListMyMessageRequest) (*v1.ListMyMessageReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.msgUC.ListMyMessage(ctx, id, req.Page, req.PageSize, req.MsgType)
	if err != nil {
		return nil, err
	}
	// 未读数随列表一并返回，避免前端多次请求
	unread, err := s.msgUC.UnreadMessageCount(ctx, id)
	if err != nil {
		unread = 0
	}
	messageItems := make([]*v1.MessageItem, 0, len(items))
	for _, m := range items {
		messageItems = append(messageItems, messageBizToProto(m))
	}
	return &v1.ListMyMessageReply{
		Items:       messageItems,
		Total:       total,
		UnreadCount: unread,
	}, nil
}

// UnreadMessageCount 我的未读消息数
func (s *CinInterface) UnreadMessageCount(ctx context.Context, _ *v1.UnreadMessageCountRequest) (*v1.UnreadMessageCountReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	count, err := s.msgUC.UnreadMessageCount(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.UnreadMessageCountReply{Count: count}, nil
}

// MarkMessageRead 标记单条消息已读
func (s *CinInterface) MarkMessageRead(ctx context.Context, req *v1.MarkMessageReadRequest) (*v1.MarkMessageReadReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.msgUC.MarkMessageRead(ctx, id, req.Id); err != nil {
		return &v1.MarkMessageReadReply{Success: false}, err
	}
	return &v1.MarkMessageReadReply{Success: true}, nil
}

// MarkAllMessageRead 全部标记已读
func (s *CinInterface) MarkAllMessageRead(ctx context.Context, _ *v1.MarkAllMessageReadRequest) (*v1.MarkAllMessageReadReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.msgUC.MarkAllMessageRead(ctx, id); err != nil {
		return &v1.MarkAllMessageReadReply{Success: false}, err
	}
	return &v1.MarkAllMessageReadReply{Success: true}, nil
}

// messageBizToProto biz.Message -> v1.MessageItem 转换
func messageBizToProto(m *biz.Message) *v1.MessageItem {
	if m == nil {
		return nil
	}
	return &v1.MessageItem{
		Id:        m.ID,
		Title:     m.Title,
		Content:   m.Content,
		MsgType:   m.MsgType,
		IsRead:    m.IsRead,
		CreatedAt: m.CreatedAt,
		ReadAt:    m.ReadAt,
	}
}
