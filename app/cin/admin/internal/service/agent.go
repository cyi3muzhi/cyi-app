package service

import (
	"context"
	"encoding/json"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"
	"cyi-app/pkg/kratos/service"
)

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatReply 聊天响应
type ChatReply struct {
	Response string `json:"response"`
}

// Chat 发送聊天消息（非流式）
func (s *AdminService) Chat(ctx context.Context, req *ChatRequest) (*ChatReply, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}

	role, _ := mid.GetRoleFromClaims(ctx)
	response, err := s.agentUC.Chat(ctx, adminID, role, req.Message)
	if err != nil {
		return nil, err
	}
	return &ChatReply{Response: response}, nil
}

// ChatSSEHandler 处理 SSE 流式聊天请求，直接由 server/http.go 的 HandleFunc 注册
// message: 用户输入的消息
// adminID: 管理员 ID（已从 JWT 解析）
// flushFunc: 向客户端写入 SSE 数据并 flush
func (s *AdminService) ChatSSEHandler(ctx context.Context, message string, adminID int64, flushFunc func([]byte)) (string, error) {
	role, _ := mid.GetRoleFromClaims(ctx)
	return s.agentUC.ChatStream(ctx, adminID, role, message, func(chunk string) {
		data, _ := json.Marshal(map[string]string{"chunk": chunk})
		flushFunc(data)
	})
}

// ClearChatHistory 清除聊天历史
func (s *AdminService) ClearChatHistory(ctx context.Context) error {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return adminv1.ErrorUnauthorized("未登录")
	}
	return s.agentUC.ClearHistory(ctx, adminID)
}

// GetChatHistory 获取聊天历史
func (s *AdminService) GetChatHistory(ctx context.Context) ([]*biz.HistoryMessage, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}
	history, err := s.agentUC.LoadChatHistory(ctx, adminID)
	if err != nil {
		return nil, err
	}
	// 没有历史时返回空数组
	if history == nil {
		return []*biz.HistoryMessage{}, nil
	}
	return history, nil
}
