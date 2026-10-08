package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"
	chatRpc "cyi-app/api/service/chat/v1"

	kratosService "cyi-app/pkg/kratos/service"
)

// ===== 用户端 IM REST 接口（转发到 chat 服务，用户ID 从 JWT 提取） =====

// CreateConversation 创建会话
func (s *CinInterface) CreateConversation(ctx context.Context, req *v1.CreateConversationRequest) (*v1.CreateConversationReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	conv, err := s.chatUC.CreateConversation(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &v1.CreateConversationReply{Conversation: convInfoToInterface(conv)}, nil
}

// ListMyConversation 我的会话列表
func (s *CinInterface) ListMyConversation(ctx context.Context, req *v1.ListMyConversationRequest) (*v1.ListMyConversationReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.ListMyConversation(ctx, userID, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ConversationItem, 0, len(reply.Items))
	for _, c := range reply.Items {
		items = append(items, convInfoToInterface(c))
	}
	return &v1.ListMyConversationReply{Items: items, Total: reply.Total}, nil
}

// ListMessage 会话历史消息
func (s *CinInterface) ListMessage(ctx context.Context, req *v1.ListMessageRequest) (*v1.ListMessageReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.ListMessage(ctx, userID, req.ConversationId, req.CursorId, req.Limit, req.Direction)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ChatMessageItem, 0, len(reply.Items))
	for _, m := range reply.Items {
		items = append(items, chatMsgToInterface(m))
	}
	return &v1.ListMessageReply{Items: items, HasMore: reply.HasMore}, nil
}

// SendMessage 用户发送消息
func (s *CinInterface) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.SendMessage(ctx, userID, req.ConversationId, req.MsgType, req.Content, req.Extra)
	if err != nil {
		return nil, err
	}
	return &v1.SendMessageReply{MessageId: reply.MessageId, CreatedAt: reply.CreatedAt}, nil
}

// MarkConversationRead 用户标记会话已读
func (s *CinInterface) MarkConversationRead(ctx context.Context, req *v1.MarkConversationReadRequest) (*v1.MarkConversationReadReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.MarkConversationRead(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}
	return &v1.MarkConversationReadReply{Success: true}, nil
}

// CloseConversation 用户结束会话
func (s *CinInterface) CloseConversation(ctx context.Context, req *v1.CloseConversationRequest) (*v1.CloseConversationReply, error) {
	userID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.CloseConversation(ctx, userID, req.ConversationId); err != nil {
		return nil, err
	}
	return &v1.CloseConversationReply{Success: true}, nil
}

// convInfoToInterface chat 服务会话信息 → 用户端会话项
func convInfoToInterface(c *chatRpc.ConversationInfo) *v1.ConversationItem {
	return &v1.ConversationItem{
		Id:            c.Id,
		AgentId:       c.AgentId,
		Status:        c.Status,
		ServiceType:   c.ServiceType,
		LastMessage:   c.LastMessage,
		LastMsgType:   c.LastMsgType,
		LastMessageAt: c.LastMessageAt,
		UnreadCount:   c.UnreadCount,
		AgentName:     c.AgentName,
		CreatedAt:     c.CreatedAt,
	}
}

// chatMsgToInterface chat 服务消息信息 → 用户端消息项
func chatMsgToInterface(m *chatRpc.ChatMessageInfo) *v1.ChatMessageItem {
	return &v1.ChatMessageItem{
		Id:             m.Id,
		ConversationId: m.ConversationId,
		SenderType:     m.SenderType,
		SenderId:       m.SenderId,
		MsgType:        m.MsgType,
		Content:        m.Content,
		Extra:          m.Extra,
		IsRead:         m.IsRead,
		CreatedAt:      m.CreatedAt,
	}
}
