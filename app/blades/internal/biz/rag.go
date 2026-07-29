package biz

import (
	"context"
	"fmt"

	"cyi-app/app/blades/internal/conf"
)

// VectorStore 向量存储接口（由 data 层实现）
type VectorStore interface {
	CreateIndex(ctx context.Context) error
	Store(ctx context.Context, doc *KnowledgeDoc) error
	Search(ctx context.Context, source string, queryVec []float32, topK int) ([]*SearchResult, error)
	DeleteDoc(ctx context.Context, docID string, source string) error
}

// KnowledgeDoc 知识文档块
type KnowledgeDoc struct {
	DocID     string            `json:"doc_id"`
	Source    string            `json:"source"`
	Title     string            `json:"title"`
	ChunkIdx  int               `json:"chunk_idx"`
	Content   string            `json:"content"`
	Tags      map[string]string `json:"tags"`
	Embedding []float32         `json:"-"`
}

// SearchResult 搜索结果
type SearchResult struct {
	DocID   string            `json:"doc_id"`
	Title   string            `json:"title"`
	Content string            `json:"content"`
	Score   float64           `json:"score"`
	Tags    map[string]string `json:"tags"`
}

// RAGEngine 检索增强生成引擎
type RAGEngine struct {
	embedder *EmbeddingClient
	vectorDB VectorStore
	splitter *TextSplitter
	config   *conf.RAGConfig
}

// NewRAGEngine 创建 RAG 引擎
func NewRAGEngine(embedder *EmbeddingClient, vectorDB VectorStore, cfg *conf.RAGConfig) *RAGEngine {
	splitterCfg := TextSplitterConfig{
		ChunkSize:    800,
		ChunkOverlap: 0.15,
	}
	if cfg != nil {
		if cfg.ChunkSize > 0 {
			splitterCfg.ChunkSize = int(cfg.ChunkSize)
		}
		if cfg.ChunkOverlap > 0 {
			splitterCfg.ChunkOverlap = float64(cfg.ChunkOverlap)
		}
	}
	return &RAGEngine{
		embedder: embedder,
		vectorDB: vectorDB,
		splitter: NewTextSplitter(splitterCfg),
		config:   cfg,
	}
}

// IndexDocument 索引一篇知识文档
func (e *RAGEngine) IndexDocument(ctx context.Context, docID, source, title, content string, tags map[string]string) (int, error) {
	chunks := e.splitter.Split(title, content)

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	vectors, err := e.embedder.BatchEmbed(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("batch embed failed: %w", err)
	}

	for i, chunk := range chunks {
		doc := &KnowledgeDoc{
			DocID:     docID,
			Source:    source,
			Title:     chunk.Title,
			ChunkIdx:  i,
			Content:   chunk.Content,
			Tags:      tags,
			Embedding: vectors[i],
		}
		if err := e.vectorDB.Store(ctx, doc); err != nil {
			return 0, fmt.Errorf("store chunk %d failed: %w", i, err)
		}
	}

	return len(chunks), nil
}

// Retrieve 检索相关知识
func (e *RAGEngine) Retrieve(ctx context.Context, source, query string, topK int) (string, error) {
	if topK <= 0 {
		topK = 5
		if e.config != nil && e.config.TopK > 0 {
			topK = int(e.config.TopK)
		}
	}

	queryVec, err := e.embedder.Embed(ctx, query)
	if err != nil {
		return "", fmt.Errorf("embed query failed: %w", err)
	}

	results, err := e.vectorDB.Search(ctx, source, queryVec, topK)
	if err != nil {
		return "", fmt.Errorf("vector search failed: %w", err)
	}

	if len(results) == 0 {
		return "", nil
	}

	context := "以下是与问题相关的知识库内容：\n\n"
	for _, r := range results {
		if r.Score < 0.3 {
			continue
		}
		context += fmt.Sprintf("- %s（相关度: %.0f%%）\n%s\n\n", r.Title, r.Score*100, r.Content)
	}

	return context, nil
}

// ShouldRetrieve 判断是否需要 RAG 检索
func (e *RAGEngine) ShouldRetrieve(message string) bool {
	if e.config != nil && e.config.AlwaysRetrieve {
		return true
	}
	keywords := []string{"规则", "政策", "说明", "指南", "流程", "怎么", "如何",
		"什么", "啥", "什么意思", "规定", "退", "保修", "运费", "发货", "尺码", "售后",
		"优惠券", "会员", "积分", "等级"}
	for _, kw := range keywords {
		if stringsContains(message, kw) {
			return true
		}
	}
	return false
}

func stringsContains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
