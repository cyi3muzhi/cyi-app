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
		UserId:    adminID,
		Source:    "admin",
		Role:      role,
		Message:   message,
		EnableRag: true,
	}
	resp, err := uc.client.Chat(ctx, req)
	if err != nil {
		uc.log.Error("Agent Chat failed: %v", err)
		return "", err
	}
	return resp.Message, nil
}

// GenerateText 生成文本（不保存聊天历史），直接调 blades GenerateText 接口
func (uc *AgentUseCase) GenerateText(ctx context.Context, prompt string) (string, error) {
	req := &bladesRpc.GenerateTextRequest{
		Prompt: prompt,
	}
	resp, err := uc.client.GenerateText(ctx, req)
	if err != nil {
		uc.log.Error("Agent GenerateText failed: %v", err)
		return "", err
	}
	return resp.Text, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, adminID int64, role string, message string, chunkCb func(chunk string)) (string, error) {
	req := &bladesRpc.ChatStreamRequest{
		UserId:    adminID,
		Source:    "admin",
		Role:      role,
		Message:   message,
		EnableRag: true,
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
		UserId: adminID,
		Source: "admin",
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
		UserId: adminID,
		Source: "admin",
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
