package biz

import (
	"cyi-app/app/blades/internal/conf"
)

// ProviderEmbeddingClient 从配置创建 Embedding 客户端
func ProviderEmbeddingClient(cfg *conf.Blades) *EmbeddingClient {
	apiKey := cfg.ApiKey
	model := "deepseek-embedding"
	dimension := 768
	if cfg.Embedding != nil {
		if cfg.Embedding.Model != "" {
			model = cfg.Embedding.Model
		}
		if cfg.Embedding.Dimension > 0 {
			dimension = int(cfg.Embedding.Dimension)
		}
	}
	return NewEmbeddingClient(apiKey, "", model, dimension)
}

// ProviderRAGConfig 从配置提取 RAG 配置
func ProviderRAGConfig(cfg *conf.Blades) *conf.RAGConfig {
	if cfg != nil {
		return cfg.Rag
	}
	return nil
}
