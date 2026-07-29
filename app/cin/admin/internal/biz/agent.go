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

// KnowledgeDocInfo 知识库文档信息
type KnowledgeDocInfo struct {
	DocID      string            `json:"doc_id"`
	Source     string            `json:"source"`
	Title      string            `json:"title"`
	ChunkCount int               `json:"chunk_count"`
	Tags       map[string]string `json:"tags"`
}

// ListKnowledge 列出知识库文档
func (uc *AgentUseCase) ListKnowledge(ctx context.Context, source string) ([]*KnowledgeDocInfo, error) {
	req := &bladesRpc.ListKnowledgeRequest{
		Source: source,
	}
	resp, err := uc.client.ListKnowledge(ctx, req)
	if err != nil {
		uc.log.Error("ListKnowledge failed: %v", err)
		return nil, err
	}

	docs := make([]*KnowledgeDocInfo, 0, len(resp.Docs))
	for _, d := range resp.Docs {
		docs = append(docs, &KnowledgeDocInfo{
			DocID:      d.DocId,
			Source:     d.Source,
			Title:      d.Title,
			ChunkCount: int(d.ChunkCount),
			Tags:       d.Tags,
		})
	}
	return docs, nil
}

// DeleteKnowledge 删除知识库文档
func (uc *AgentUseCase) DeleteKnowledge(ctx context.Context, docID, source string) error {
	req := &bladesRpc.DeleteKnowledgeRequest{
		DocId:  docID,
		Source: source,
	}
	_, err := uc.client.DeleteKnowledge(ctx, req)
	if err != nil {
		uc.log.Error("DeleteKnowledge failed: %v", err)
		return err
	}
	return nil
}

// KnowledgeSearchResult 知识库搜索结果
type KnowledgeSearchResult struct {
	DocID   string  `json:"doc_id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Score   float32 `json:"score"`
}

// SearchKnowledge 搜索知识库
func (uc *AgentUseCase) SearchKnowledge(ctx context.Context, query, source string, topK int) ([]*KnowledgeSearchResult, error) {
	req := &bladesRpc.SearchKnowledgeRequest{
		Source: source,
		Query:  query,
		TopK:   int32(topK),
	}
	resp, err := uc.client.SearchKnowledge(ctx, req)
	if err != nil {
		uc.log.Error("SearchKnowledge failed: %v", err)
		return nil, err
	}

	results := make([]*KnowledgeSearchResult, 0, len(resp.Chunks))
	for _, c := range resp.Chunks {
		results = append(results, &KnowledgeSearchResult{
			DocID:   c.DocId,
			Title:   c.Title,
			Content: c.Content,
			Score:   c.Score,
		})
	}
	return results, nil
}

// UploadKnowledge 上传知识库文档（通过 gRPC 流式调用 blades 服务）
func (uc *AgentUseCase) UploadKnowledge(ctx context.Context, docID, source, title, contentType string, content []byte, tags map[string]string) (chunkCount int, err error) {
	stream, err := uc.client.UploadKnowledge(ctx)
	if err != nil {
		uc.log.Error("UploadKnowledge stream failed: %v", err)
		return 0, err
	}

	// 发送元信息
	if err := stream.Send(&bladesRpc.UploadKnowledgeRequest{
		Data: &bladesRpc.UploadKnowledgeRequest_Meta{
			Meta: &bladesRpc.KnowledgeMeta{
				DocId:        docID,
				Source:       source,
				Title:        title,
				ContentType:  contentType,
				Tags:         tags,
			},
		},
	}); err != nil {
		uc.log.Error("UploadKnowledge send meta failed: %v", err)
		return 0, err
	}

	// 发送内容分片（每片 64KB）
	chunkSize := 64 * 1024
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		if err := stream.Send(&bladesRpc.UploadKnowledgeRequest{
			Data: &bladesRpc.UploadKnowledgeRequest_ContentChunk{
				ContentChunk: content[i:end],
			},
		}); err != nil {
			uc.log.Error("UploadKnowledge send chunk failed: %v", err)
			return 0, err
		}
	}

	// 关闭发送并接收响应
	resp, err := stream.CloseAndRecv()
	if err != nil {
		uc.log.Error("UploadKnowledge close and recv failed: %v", err)
		return 0, err
	}
	return int(resp.ChunkCount), nil
}
