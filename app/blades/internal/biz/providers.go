package biz

// ProviderEmbeddingClient 从配置创建 Embedding 客户端
// 优先使用 embedding 专用的 api_key 和 api_endpoint；为空则回退到 blades.api_key
func ProviderEmbeddingClient(cfg *BladesConfig) *EmbeddingClient {
	// 默认使用 blades.api_key 作为 embedding 的 key
	apiKey := cfg.ApiKey
	model := "BAAI/bge-large-zh-v1.5"
	dimension := 1024
	apiEndpoint := ""
	if cfg.Embedding != nil {
		if cfg.Embedding.Model != "" {
			model = cfg.Embedding.Model
		}
		if cfg.Embedding.Dimension > 0 {
			dimension = cfg.Embedding.Dimension
		}
		// embedding 专用 key 优先
		if cfg.Embedding.ApiKey != "" {
			apiKey = cfg.Embedding.ApiKey
		}
		// embedding 专用 endpoint
		apiEndpoint = cfg.Embedding.ApiEndpoint
	}
	return NewEmbeddingClient(apiKey, apiEndpoint, model, dimension)
}

// ProviderRAGConfig 从 BladesConfig 提取 RAG 配置
func ProviderRAGConfig(cfg *BladesConfig) *RAGConfigBiz {
	if cfg == nil || cfg.Rag == nil {
		return nil
	}
	return cfg.Rag
}
