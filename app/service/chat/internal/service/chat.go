package service

import (
	"context"

	v1 "cyi-app/api/service/chat/v1"
	"cyi-app/app/service/chat/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// ChatService IM 客服 gRPC 服务实现
type ChatService struct {
	v1.UnimplementedChatServer

	convUC *biz.ConversationUseCase
	msgUC  *biz.MessageUseCase
	log    *log.Helper
}

// NewChatService 创建聊天服务
func NewChatService(convUC *biz.ConversationUseCase, msgUC *biz.MessageUseCase, logger log.Logger) *ChatService {
	return &ChatService{convUC: convUC, msgUC: msgUC, log: log.NewHelper(logger)}
}

// ===== 用户端 =====

// CreateConversation 创建会话
func (s *ChatService) CreateConversation(ctx context.Context, req *v1.CreateConversationRequest) (*v1.CreateConversationReply, error) {
	return s.convUC.CreateConversation(ctx, req)
}

// ListMyConversation 我的会话列表
func (s *ChatService) ListMyConversation(ctx context.Context, req *v1.ListMyConversationRequest) (*v1.ListMyConversationReply, error) {
	return s.convUC.ListMyConversation(ctx, req)
}

// ListMessage 会话历史消息
func (s *ChatService) ListMessage(ctx context.Context, req *v1.ListMessageRequest) (*v1.ListMessageReply, error) {
	return s.msgUC.ListMessage(ctx, req)
}

// ListAgentMessage 会话历史消息（客服端）
func (s *ChatService) ListAgentMessage(ctx context.Context, req *v1.ListAgentMessageRequest) (*v1.ListMessageReply, error) {
	return s.msgUC.ListAgentMessage(ctx, req)
}

// SendMessage 用户发送消息
func (s *ChatService) SendMessage(ctx context.Context, req *v1.SendMessageRequest) (*v1.SendMessageReply, error) {
	return s.msgUC.SendMessage(ctx, req)
}

// MarkConversationRead 用户标记会话已读
func (s *ChatService) MarkConversationRead(ctx context.Context, req *v1.MarkConversationReadRequest) (*v1.MarkConversationReadReply, error) {
	return s.msgUC.MarkConversationRead(ctx, req)
}

// CloseConversation 用户结束会话
func (s *ChatService) CloseConversation(ctx context.Context, req *v1.CloseConversationRequest) (*v1.CloseConversationReply, error) {
	return s.convUC.CloseConversation(ctx, req)
}

// ===== 客服端 =====

// ListPendingConversation 待接入会话列表
func (s *ChatService) ListPendingConversation(ctx context.Context, req *v1.ListPendingConversationRequest) (*v1.ListPendingConversationReply, error) {
	return s.convUC.ListPendingConversation(ctx, req)
}

// AcceptConversation 客服接入会话
func (s *ChatService) AcceptConversation(ctx context.Context, req *v1.AcceptConversationRequest) (*v1.AcceptConversationReply, error) {
	return s.convUC.AcceptConversation(ctx, req)
}

// TransferConversation 客服转接会话
func (s *ChatService) TransferConversation(ctx context.Context, req *v1.TransferConversationRequest) (*v1.TransferConversationReply, error) {
	return s.convUC.TransferConversation(ctx, req)
}

// ListAgentConversation 客服名下会话列表
func (s *ChatService) ListAgentConversation(ctx context.Context, req *v1.ListAgentConversationRequest) (*v1.ListAgentConversationReply, error) {
	return s.convUC.ListAgentConversation(ctx, req)
}

// SendAgentMessage 客服发送消息
func (s *ChatService) SendAgentMessage(ctx context.Context, req *v1.SendAgentMessageRequest) (*v1.SendAgentMessageReply, error) {
	return s.msgUC.SendAgentMessage(ctx, req)
}

// MarkAgentRead 客服标记会话已读
func (s *ChatService) MarkAgentRead(ctx context.Context, req *v1.MarkAgentReadRequest) (*v1.MarkAgentReadReply, error) {
	return s.msgUC.MarkAgentRead(ctx, req)
}

// CloseAgentConversation 客服结束会话
func (s *ChatService) CloseAgentConversation(ctx context.Context, req *v1.CloseAgentConversationRequest) (*v1.CloseAgentConversationReply, error) {
	return s.convUC.CloseAgentConversation(ctx, req)
}

// AgentStats 客服接待统计
func (s *ChatService) AgentStats(ctx context.Context, req *v1.AgentStatsRequest) (*v1.AgentStatsReply, error) {
	return s.convUC.AgentStats(ctx, req)
}
