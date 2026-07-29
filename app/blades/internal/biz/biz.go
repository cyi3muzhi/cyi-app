package biz

import (
	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewAgentUseCase,
	NewAgentFactory,
	NewRAGEngine,
	ProviderEmbeddingClient,
	ProviderRAGConfig,
	NewKnowledgeUseCase,
	NewTextSplitterConfig,
	InitToolRegistry,
)

// NewTextSplitterConfig 创建默认分块配置
func NewTextSplitterConfig() TextSplitterConfig {
	return TextSplitterConfig{ChunkSize: 800, ChunkOverlap: 0.15}
}

// InitToolRegistry 创建工具注册中心并注册所有工具
func InitToolRegistry(
	pc productRpc.ProductClient,
	sc productRpc.StockClient,
	uc userRpc.UserClient,
	oc orderRpc.OrderClient,
	rag *RAGEngine,
	rdb redis.UniversalClient,
	roleClient userRpc.AdminRoleClient,
) *ToolRegistry {
	r := NewToolRegistry()
	RegisterOrderTools(r, oc)
	RegisterProductTools(r, pc)
	RegisterStockTools(r, sc, pc)
	RegisterUserTools(r, uc)
	RegisterKnowledgeTools(r, rag)
	RegisterPreferenceTool(r, rdb)
	RegisterPermissionTool(r, roleClient)
	return r
}
