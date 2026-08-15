// Package client 提供 blades 服务的 gRPC 客户端封装，供 admin / BFF 等各服务复用。
// 注意：这里是「调用方」的薄封装，AI 业务逻辑本体（LLM 调用、RAG、聊天历史、知识库）仍在 blades 服务内。
package client

import (
	"context"
	"io"

	bladesRpc "cyi-app/api/blades/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Client blades 服务 gRPC 客户端封装
type Client struct {
	client bladesRpc.BladesClient
	log    *log.Helper
}

// NewClient 创建 blades 服务客户端封装
// client: 底层 gRPC 连接（由各服务 data 层创建，含服务发现/追踪/监控）
func NewClient(client bladesRpc.BladesClient, logger log.Logger) *Client {
	return &Client{
		client: client,
		log:    log.NewHelper(logger),
	}
}

// Chat 非流式聊天
// source: 来源标识（如 "admin" / "user"），用于隔离不同端的历史与知识库
func (c *Client) Chat(ctx context.Context, source string, userID int64, role, message string, enableRag bool) (string, error) {
	req := &bladesRpc.ChatRequest{
		UserId:    userID,
		Source:    source,
		Role:      role,
		Message:   message,
		EnableRag: enableRag,
	}
	resp, err := c.client.Chat(ctx, req)
	if err != nil {
		c.log.Error("Chat failed: %v", err)
		return "", err
	}
	return resp.Message, nil
}

// GenerateText 生成文本（不保存聊天历史），直接调 blades GenerateText 接口
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	req := &bladesRpc.GenerateTextRequest{
		Prompt: prompt,
	}
	resp, err := c.client.GenerateText(ctx, req)
	if err != nil {
		c.log.Error("GenerateText failed: %v", err)
		return "", err
	}
	return resp.Text, nil
}

// ChatStream 流式聊天
// source: 来源标识；chunkCb: 逐块回调（可空，为空时仅聚合完整响应）
func (c *Client) ChatStream(ctx context.Context, source string, userID int64, role, message string, enableRag bool, chunkCb func(chunk string)) (string, error) {
	req := &bladesRpc.ChatStreamRequest{
		UserId:    userID,
		Source:    source,
		Role:      role,
		Message:   message,
		EnableRag: enableRag,
	}
	stream, err := c.client.ChatStream(ctx, req)
	if err != nil {
		c.log.Error("ChatStream failed: %v", err)
		return "", err
	}

	var fullResponse string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Error("ChatStream Recv failed: %v", err)
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
func (c *Client) ClearHistory(ctx context.Context, source string, userID int64) error {
	req := &bladesRpc.ClearChatHistoryRequest{
		UserId: userID,
		Source: source,
	}
	_, err := c.client.ClearChatHistory(ctx, req)
	if err != nil {
		c.log.Error("ClearHistory failed: %v", err)
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
func (c *Client) LoadChatHistory(ctx context.Context, source string, userID int64) ([]*HistoryMessage, error) {
	req := &bladesRpc.GetChatHistoryRequest{
		UserId: userID,
		Source: source,
	}
	resp, err := c.client.GetChatHistory(ctx, req)
	if err != nil {
		c.log.Error("LoadChatHistory failed: %v", err)
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
func (c *Client) ListKnowledge(ctx context.Context, source string) ([]*KnowledgeDocInfo, error) {
	req := &bladesRpc.ListKnowledgeRequest{
		Source: source,
	}
	resp, err := c.client.ListKnowledge(ctx, req)
	if err != nil {
		c.log.Error("ListKnowledge failed: %v", err)
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
func (c *Client) DeleteKnowledge(ctx context.Context, docID, source string) error {
	req := &bladesRpc.DeleteKnowledgeRequest{
		DocId:  docID,
		Source: source,
	}
	_, err := c.client.DeleteKnowledge(ctx, req)
	if err != nil {
		c.log.Error("DeleteKnowledge failed: %v", err)
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
func (c *Client) SearchKnowledge(ctx context.Context, query, source string, topK int) ([]*KnowledgeSearchResult, error) {
	req := &bladesRpc.SearchKnowledgeRequest{
		Source: source,
		Query:  query,
		TopK:   int32(topK),
	}
	resp, err := c.client.SearchKnowledge(ctx, req)
	if err != nil {
		c.log.Error("SearchKnowledge failed: %v", err)
		return nil, err
	}

	results := make([]*KnowledgeSearchResult, 0, len(resp.Chunks))
	for _, chunk := range resp.Chunks {
		results = append(results, &KnowledgeSearchResult{
			DocID:   chunk.DocId,
			Title:   chunk.Title,
			Content: chunk.Content,
			Score:   chunk.Score,
		})
	}
	return results, nil
}

// UploadKnowledge 上传知识库文档（gRPC 流式调用，内容按 64KB 分片发送）
// 返回分块数量
func (c *Client) UploadKnowledge(ctx context.Context, docID, source, title, contentType string, content []byte, tags map[string]string) (int, error) {
	stream, err := c.client.UploadKnowledge(ctx)
	if err != nil {
		c.log.Error("UploadKnowledge stream failed: %v", err)
		return 0, err
	}

	// 发送元信息
	if err := stream.Send(&bladesRpc.UploadKnowledgeRequest{
		Data: &bladesRpc.UploadKnowledgeRequest_Meta{
			Meta: &bladesRpc.KnowledgeMeta{
				DocId:       docID,
				Source:      source,
				Title:       title,
				ContentType: contentType,
				Tags:        tags,
			},
		},
	}); err != nil {
		c.log.Error("UploadKnowledge send meta failed: %v", err)
		return 0, err
	}

	// 发送内容分片（每片 64KB）
	const chunkSize = 64 * 1024
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
			c.log.Error("UploadKnowledge send chunk failed: %v", err)
			return 0, err
		}
	}

	// 关闭发送并接收响应
	resp, err := stream.CloseAndRecv()
	if err != nil {
		c.log.Error("UploadKnowledge close and recv failed: %v", err)
		return 0, err
	}
	return int(resp.ChunkCount), nil
}
