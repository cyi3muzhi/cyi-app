package data

import (
	"context"

	chatRpc "cyi-app/api/service/chat/v1"
	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// chatRepo 聊天数据仓库（gRPC 代理到 Chat Service）
type chatRepo struct {
	data *Data
	log  *log.Helper
}

// NewChatRepo 创建 chatRepo
func NewChatRepo(data *Data, logger log.Logger) biz.ChatRepo {
	return &chatRepo{data: data, log: log.NewHelper(logger)}
}

// CreateConversation 创建会话（复用未结束会话）
func (r *chatRepo) CreateConversation(ctx context.Context, userID int64) (*chatRpc.ConversationInfo, error) {
	reply, err := r.data.chc.CreateConversation(ctx, &chatRpc.CreateConversationRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return reply.Conversation, nil
}

// ListMyConversation 我的会话列表
func (r *chatRepo) ListMyConversation(ctx context.Context, userID int64, page, pageSize, status int32) (*chatRpc.ListMyConversationReply, error) {
	return r.data.chc.ListMyConversation(ctx, &chatRpc.ListMyConversationRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
}

// ListMessage 会话历史消息（校验用户成员）
func (r *chatRepo) ListMessage(ctx context.Context, userID, conversationID, cursorID int64, limit, direction int32) (*chatRpc.ListMessageReply, error) {
	return r.data.chc.ListMessage(ctx, &chatRpc.ListMessageRequest{
		UserId:         userID,
		ConversationId: conversationID,
		CursorId:       cursorID,
		Limit:          limit,
		Direction:      direction,
	})
}

// SendMessage 用户发送消息
func (r *chatRepo) SendMessage(ctx context.Context, userID, conversationID int64, msgType int32, content, extra string) (*chatRpc.SendMessageReply, error) {
	return r.data.chc.SendMessage(ctx, &chatRpc.SendMessageRequest{
		UserId:         userID,
		ConversationId: conversationID,
		MsgType:        msgType,
		Content:        content,
		Extra:          extra,
	})
}

// MarkConversationRead 用户标记会话已读
func (r *chatRepo) MarkConversationRead(ctx context.Context, userID, conversationID int64) error {
	_, err := r.data.chc.MarkConversationRead(ctx, &chatRpc.MarkConversationReadRequest{
		UserId:         userID,
		ConversationId: conversationID,
	})
	return err
}

// CloseConversation 用户结束会话
func (r *chatRepo) CloseConversation(ctx context.Context, userID, conversationID int64) error {
	_, err := r.data.chc.CloseConversation(ctx, &chatRpc.CloseConversationRequest{
		UserId:         userID,
		ConversationId: conversationID,
	})
	return err
}