package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	chatRpc "cyi-app/api/service/chat/v1"

	kratosService "cyi-app/pkg/kratos/service"
)

// ===== 客服工作台 REST 接口（转发到 chat 服务，客服ID 从 JWT 提取） =====

// ListPendingConversation 待接入会话列表
func (s *AdminService) ListPendingConversation(ctx context.Context, req *adminv1.ListPendingConversationRequest) (*adminv1.ListPendingConversationReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.ListPendingConversation(ctx, agentID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.ConversationItem, 0, len(reply.Items))
	for _, c := range reply.Items {
		items = append(items, convInfoToAdmin(c))
	}
	return &adminv1.ListPendingConversationReply{Items: items, Total: reply.Total}, nil
}

// AcceptConversation 客服接入会话
func (s *AdminService) AcceptConversation(ctx context.Context, req *adminv1.AcceptConversationRequest) (*adminv1.AcceptConversationReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.AcceptConversation(ctx, agentID, req.ConversationId); err != nil {
		return nil, err
	}
	return &adminv1.AcceptConversationReply{Success: true}, nil
}

// TransferConversation 客服转接会话
func (s *AdminService) TransferConversation(ctx context.Context, req *adminv1.TransferConversationRequest) (*adminv1.TransferConversationReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.TransferConversation(ctx, agentID, req.ConversationId, req.TargetAgentId); err != nil {
		return nil, err
	}
	return &adminv1.TransferConversationReply{Success: true}, nil
}

// ListAgentConversation 客服名下会话列表
func (s *AdminService) ListAgentConversation(ctx context.Context, req *adminv1.ListAgentConversationRequest) (*adminv1.ListAgentConversationReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.ListAgentConversation(ctx, agentID, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.ConversationItem, 0, len(reply.Items))
	for _, c := range reply.Items {
		items = append(items, convInfoToAdmin(c))
	}
	return &adminv1.ListAgentConversationReply{Items: items, Total: reply.Total}, nil
}

// ChatListMessage 客服端会话历史消息
func (s *AdminService) ChatListMessage(ctx context.Context, req *adminv1.ChatListMessageRequest) (*adminv1.ChatListMessageReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.ListAgentMessage(ctx, agentID, req.ConversationId, req.CursorId, req.Limit, req.Direction)
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.ChatMessageItem, 0, len(reply.Items))
	for _, m := range reply.Items {
		items = append(items, chatMsgToAdmin(m))
	}
	return &adminv1.ChatListMessageReply{Items: items, HasMore: reply.HasMore}, nil
}

// SendAgentMessage 客服发送消息
func (s *AdminService) SendAgentMessage(ctx context.Context, req *adminv1.SendAgentMessageRequest) (*adminv1.SendAgentMessageReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.SendAgentMessage(ctx, agentID, req.ConversationId, req.MsgType, req.Content, req.Extra)
	if err != nil {
		return nil, err
	}
	return &adminv1.SendAgentMessageReply{MessageId: reply.MessageId, CreatedAt: reply.CreatedAt}, nil
}

// MarkAgentRead 客服标记会话已读
func (s *AdminService) MarkAgentRead(ctx context.Context, req *adminv1.MarkAgentReadRequest) (*adminv1.MarkAgentReadReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.MarkAgentRead(ctx, agentID, req.ConversationId); err != nil {
		return nil, err
	}
	return &adminv1.MarkAgentReadReply{Success: true}, nil
}

// CloseAgentConversation 客服结束会话
func (s *AdminService) CloseAgentConversation(ctx context.Context, req *adminv1.CloseAgentConversationRequest) (*adminv1.CloseAgentConversationReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.chatUC.CloseAgentConversation(ctx, agentID, req.ConversationId); err != nil {
		return nil, err
	}
	return &adminv1.CloseAgentConversationReply{Success: true}, nil
}

// AgentStats 客服接待统计
func (s *AdminService) AgentStats(ctx context.Context, req *adminv1.AgentStatsRequest) (*adminv1.AgentStatsReply, error) {
	agentID, err := kratosService.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	reply, err := s.chatUC.AgentStats(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return &adminv1.AgentStatsReply{
		TotalToday:    reply.TotalToday,
		Processing:    reply.Processing,
		ResolvedToday: reply.ResolvedToday,
	}, nil
}

// convInfoToAdmin chat 服务会话信息 → 客服端会话项
func convInfoToAdmin(c *chatRpc.ConversationInfo) *adminv1.ConversationItem {
	return &adminv1.ConversationItem{
		Id:            c.Id,
		UserId:        c.UserId,
		AgentId:       c.AgentId,
		Status:        c.Status,
		ServiceType:   c.ServiceType,
		LastMessage:   c.LastMessage,
		LastMsgType:   c.LastMsgType,
		LastMessageAt: c.LastMessageAt,
		UnreadCount:   c.UnreadCount,
		UserName:      c.UserName,
		UserOnline:    c.UserOnline,
		CreatedAt:     c.CreatedAt,
	}
}

// chatMsgToAdmin chat 服务消息信息 → 客服端消息项
func chatMsgToAdmin(m *chatRpc.ChatMessageInfo) *adminv1.ChatMessageItem {
	return &adminv1.ChatMessageItem{
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
