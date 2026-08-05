package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// KnowledgeUseCase 知识库管理用例
type KnowledgeUseCase struct {
	rag *RAGEngine
	rdb redis.UniversalClient
	log *log.Helper
}

// NewKnowledgeUseCase 创建知识库管理用例
func NewKnowledgeUseCase(rag *RAGEngine, rdb redis.UniversalClient, logger log.Logger) *KnowledgeUseCase {
	return &KnowledgeUseCase{
		rag: rag,
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

// UploadDocument 上传并索引文档
func (uc *KnowledgeUseCase) UploadDocument(ctx context.Context, docID, source, title, contentType string, content []byte, tags map[string]string) (chunkCount, tokenCount int, err error) {
	if uc.rag == nil {
		return 0, 0, fmt.Errorf("RAG engine not initialized")
	}

	// 创建索引（如果不存在）
	if uc.rag.vectorDB != nil {
		if err := uc.rag.vectorDB.CreateIndex(ctx); err != nil {
			uc.log.Warn("create vector index: %v", err)
		}
	}

	// 索引文档
	count, err := uc.rag.IndexDocument(ctx, docID, source, title, string(content), tags)
	if err != nil {
		return 0, 0, fmt.Errorf("index document failed: %w", err)
	}

	return count, len(content), nil
}

// SearchKnowledge 搜索知识库（直接返回结构化结果，避免重复检索）
func (uc *KnowledgeUseCase) SearchKnowledge(ctx context.Context, source, query string, topK int) ([]*SearchResult, error) {
	if uc.rag == nil {
		return nil, nil
	}
	return uc.rag.Retrieve(ctx, source, query, topK)
}

// DeleteDocument 删除文档
func (uc *KnowledgeUseCase) DeleteDocument(ctx context.Context, docID, source string) error {
	if uc.rag == nil || uc.rag.vectorDB == nil {
		return nil
	}
	return uc.rag.vectorDB.DeleteDoc(ctx, docID, source)
}

// ListDocuments 列出知识库文档
func (uc *KnowledgeUseCase) ListDocuments(ctx context.Context, source string) ([]*DocInfo, error) {
	if uc.rag == nil || uc.rag.vectorDB == nil {
		return nil, nil
	}
	return uc.rag.vectorDB.ListDocuments(ctx, source)
}
