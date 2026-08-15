package biz

import (
	"context"

	bladesRpc "cyi-app/api/blades/v1"
	bladesclient "cyi-app/pkg/blades/client"

	"github.com/go-kratos/kratos/v2/log"
)

// agentSource 管理员端的来源标识，用于 blades 服务隔离不同端的历史与知识库
const agentSource = "admin"

// AgentUseCase 管理员智能体用例（底层复用 pkg/blades/client 的 blades 服务 gRPC 客户端封装）
type AgentUseCase struct {
	client *bladesclient.Client
	log    *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(client bladesRpc.BladesClient, logger log.Logger) *AgentUseCase {
	return &AgentUseCase{
		client: bladesclient.NewClient(client, logger),
		log:    log.NewHelper(logger),
	}
}

// Chat 非流式聊天（管理员来源，启用 RAG）
func (uc *AgentUseCase) Chat(ctx context.Context, adminID int64, role string, message string) (string, error) {
	return uc.client.Chat(ctx, agentSource, adminID, role, message, true)
}

// GenerateText 生成文本（不保存聊天历史），直接调 blades GenerateText 接口
func (uc *AgentUseCase) GenerateText(ctx context.Context, prompt string) (string, error) {
	return uc.client.GenerateText(ctx, prompt)
}

// ChatStream 流式聊天（管理员来源，启用 RAG）
func (uc *AgentUseCase) ChatStream(ctx context.Context, adminID int64, role string, message string, chunkCb func(chunk string)) (string, error) {
	return uc.client.ChatStream(ctx, agentSource, adminID, role, message, true, chunkCb)
}

// ClearHistory 清除聊天历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, adminID int64) error {
	return uc.client.ClearHistory(ctx, agentSource, adminID)
}

// HistoryMessage 历史消息（类型别名，复用 pkg/blades/client 的类型）
type HistoryMessage = bladesclient.HistoryMessage

// LoadChatHistory 加载聊天历史
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, adminID int64) ([]*HistoryMessage, error) {
	return uc.client.LoadChatHistory(ctx, agentSource, adminID)
}

// KnowledgeDocInfo 知识库文档信息（类型别名，复用 pkg/blades/client 的类型）
type KnowledgeDocInfo = bladesclient.KnowledgeDocInfo

// ListKnowledge 列出知识库文档
func (uc *AgentUseCase) ListKnowledge(ctx context.Context, source string) ([]*KnowledgeDocInfo, error) {
	return uc.client.ListKnowledge(ctx, source)
}

// DeleteKnowledge 删除知识库文档
func (uc *AgentUseCase) DeleteKnowledge(ctx context.Context, docID, source string) error {
	return uc.client.DeleteKnowledge(ctx, docID, source)
}

// KnowledgeSearchResult 知识库搜索结果（类型别名，复用 pkg/blades/client 的类型）
type KnowledgeSearchResult = bladesclient.KnowledgeSearchResult

// SearchKnowledge 搜索知识库
func (uc *AgentUseCase) SearchKnowledge(ctx context.Context, query, source string, topK int) ([]*KnowledgeSearchResult, error) {
	return uc.client.SearchKnowledge(ctx, query, source, topK)
}

// UploadKnowledge 上传知识库文档
func (uc *AgentUseCase) UploadKnowledge(ctx context.Context, docID, source, title, contentType string, content []byte, tags map[string]string) (int, error) {
	return uc.client.UploadKnowledge(ctx, docID, source, title, contentType, content, tags)
}
