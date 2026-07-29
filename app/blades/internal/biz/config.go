package biz

// BladesConfig Blades 业务配置（biz 层定义，避免直接依赖 conf 包）
type BladesConfig struct {
	ApiKey    string
	Model     string
	Embedding *EmbeddingConfigBiz
	Rag       *RAGConfigBiz
}

// EmbeddingConfigBiz Embedding 业务配置
type EmbeddingConfigBiz struct {
	Model     string
	Dimension int
}

// RAGConfigBiz RAG 业务配置
type RAGConfigBiz struct {
	TopK           int
	ChunkSize      int
	ChunkOverlap   float64
	AlwaysRetrieve bool
}
