package biz

// ProviderEmbeddingClient 从配置创建 Embedding 客户端
func ProviderEmbeddingClient(cfg *BladesConfig) *EmbeddingClient {
	apiKey := cfg.ApiKey
	model := "deepseek-embedding"
	dimension := 768
	if cfg.Embedding != nil {
		if cfg.Embedding.Model != "" {
			model = cfg.Embedding.Model
		}
		if cfg.Embedding.Dimension > 0 {
			dimension = cfg.Embedding.Dimension
		}
	}
	return NewEmbeddingClient(apiKey, "", model, dimension)
}

// ProviderRAGConfig 从 BladesConfig 提取 RAG 配置
func ProviderRAGConfig(cfg *BladesConfig) *RAGConfigBiz {
	if cfg == nil || cfg.Rag == nil {
		return nil
	}
	return cfg.Rag
}
