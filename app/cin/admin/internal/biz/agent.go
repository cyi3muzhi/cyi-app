package biz

import (
	"context"
	"io"

	bladesRpc "cyi-app/api/blades/v1"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
)

// AgentUseCase 管理员智能体用例（调用 blades gRPC 服务）
type AgentUseCase struct {
	client   bladesRpc.BladesClient
	enforcer mid.EnforcerProvider
	log      *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(client bladesRpc.BladesClient, enforcer mid.EnforcerProvider, logger log.Logger) *AgentUseCase {
	return &AgentUseCase{
		client:   client,
		enforcer: enforcer,
		log:      log.NewHelper(logger),
	}
}

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, adminID int64, role string, message string) (string, error) {
	req := &bladesRpc.ChatRequest{
		AdminId: adminID,
		Role:    role,
		Message: message,
	}
	resp, err := uc.client.Chat(ctx, req)
	if err != nil {
		uc.log.Error("Agent Chat failed: %v", err)
		return "", err
	}
	return resp.Message, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, adminID int64, role string, message string, chunkCb func(chunk string)) (string, error) {
	req := &bladesRpc.ChatStreamRequest{
		AdminId: adminID,
		Role:    role,
		Message: message,
	}
	stream, err := uc.client.ChatStream(ctx, req)
	if err != nil {
		uc.log.Error("Agent ChatStream failed: %v", err)
		return "", err
	}

	var fullResponse string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			uc.log.Error("ChatStream Recv failed: %v", err)
			return fullResponse, err
		}
		if resp != nil {
			fullResponse += resp.Chunk
			if resp.Done {
				break
			}
			if chunkCb != nil {
				chunkCb(resp.Chunk)
			}
		}
	}
	return fullResponse, nil
}

// ClearHistory 清除聊天历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, adminID int64) error {
	req := &bladesRpc.ClearChatHistoryRequest{
		AdminId: adminID,
	}
	_, err := uc.client.ClearChatHistory(ctx, req)
	if err != nil {
		uc.log.Error("ClearHistory failed: %v", err)
		return err
	}
	return nil
}

// HistoryMessage 历史消息
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// LoadChatHistory 加载聊天历史
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, adminID int64) ([]*HistoryMessage, error) {
	req := &bladesRpc.GetChatHistoryRequest{
		AdminId: adminID,
	}
	resp, err := uc.client.GetChatHistory(ctx, req)
	if err != nil {
		uc.log.Error("LoadChatHistory failed: %v", err)
		return nil, err
	}

	messages := make([]*HistoryMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		messages = append(messages, &HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	return messages, nil
}
