# Blades AI Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use executing-plans.

**Goal:** 将 blades 升级为完整的 AI 智能体服务，支持多端隔离、RAG、工具调用

**Architecture:** Agent Factory（按 source 分发）→ RAG Engine（Redis Stack 向量搜索）→ Tool System（下游 gRPC 工具化），三层架构

**Tech Stack:** Go + Kratos + go-kratos/blades + Redis Stack + DeepSeek API (chat + embedding)

---

### Task 1: 升级基础设施配置

**Files:**
- Modify: `deploy/docker-compose/docker-compose.infra.yml`
- Modify: `app/blades/configs/config.yaml`
- Modify: `app/blades/internal/conf/conf.proto`

- [ ] **Step 1: 升级 Redis 镜像**

`deploy/docker-compose/docker-compose.infra.yml` 中 redis 服务改为：
```yaml
redis:
  image: redis/redis-stack-server:7.2.0-v10
  restart: unless-stopped
  ports:
    - "6379:6379"
  volumes:
    - redis-data:/data
  networks:
    - cyi-app
```

- [ ] **Step 2: 更新配置结构体 conf.proto**

`app/blades/internal/conf/conf.proto` 的 Blades message 新增字段：
```protobuf
message Blades {
    string api_key = 1;
    string model = 2;
    int32 max_tokens = 3;
    float temperature = 4;
    EmbeddingConfig embedding = 5;
    RAGConfig rag = 6;
    KnowledgeConfig knowledge = 7;
}
message EmbeddingConfig {
    string model = 1;
    int32 dimension = 2;
}
message RAGConfig {
    int32 top_k = 1;
    int32 chunk_size = 2;
    float chunk_overlap = 3;
    bool always_retrieve = 4;
}
message KnowledgeConfig {
    int32 max_docs_per_source = 1;
}
```

```bash
cd app/blades; protoc --proto_path=.\internal --proto_path=..\..\third_party --go_out=paths=source_relative:.\internal .\internal\conf\conf.proto
```

- [ ] **Step 3: 新增配置到 config.yaml**

```yaml
blades:
  api_key: sk-9563c2ca6c3e44cd9c51b8faf5dd82e3
  model: deepseek-chat
  max_tokens: 4096
  temperature: 0.7
  embedding:
    model: deepseek-embedding
    dimension: 768
  rag:
    top_k: 5
    chunk_size: 800
    chunk_overlap: 0.15
    always_retrieve: false
  knowledge:
    max_docs_per_source: 1000
```

---

### Task 2: 更新 Proto API 定义并生成代码

**Files:**
- Modify: `api/blades/v1/blades.proto`

- [ ] **Step 1: 重写 blades.proto**

核心变更：
- ChatRequest/ChatStreamRequest 加 `source` (admin/user/other) + `enable_rag` 字段
- 新增 Knowledge 管理相关 message
- 移除过时的 max_len=200 限制改为 2000

关键 message 结构：

```protobuf
message ChatRequest {
    int64 user_id = 1;
    string source = 2;     // admin / user / other
    string role = 3;
    string message = 4;
    bool enable_rag = 5;
}

// 知识库上传（流式）
message UploadKnowledgeRequest {
    oneof data {
        KnowledgeMeta meta = 1;
        bytes content_chunk = 2;
    }
}
message KnowledgeMeta {
    string doc_id = 1;
    string source = 2;      // admin / user / public
    string title = 3;
    string content_type = 4;
    map<string, string> tags = 5;
}

message SearchKnowledgeRequest {
    string source = 1;
    string query = 2;
    int32 top_k = 3;
}
message KnowledgeChunk {
    string doc_id = 1;
    string title = 2;
    string content = 3;
    float score = 4;
    map<string, string> tags = 5;
}
```

- [ ] **Step 2: 生成代码**

```bash
cd e:\me\cyi-app; make api
```

---

### Task 3: 创建 Embedding Client

**Files:**
- Create: `app/blades/internal/biz/embedding.go`

封装 DeepSeek Embedding API，提供 `Embed(text) []float32` 和 `BatchEmbed(texts) [][]float32`。

```go
type EmbeddingClient struct {
    apiKey, apiEndpoint, model string
    dimension int
    client *http.Client
}

func NewEmbeddingClient(apiKey, apiEndpoint, model string, dimension int) *EmbeddingClient
func (c *EmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error)
func (c *EmbeddingClient) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
```

实现要点：POST `/embeddings`，body 带 `{"model": "deepseek-embedding", "input": [...]}`，返回的 embedding 从 float64 转 float32。

---

### Task 4: 创建向量存储（Redis Stack VectorStore）

**Files:**
- Create: `app/blades/internal/data/vector_store.go`

实现 Redis Stack 的 VECTOR 类型操作：

```go
type VectorStore struct {
    client redis.UniversalClient
    indexName string  // idx:knowledge
    dimension int
}

func NewVectorStore(client redis.UniversalClient, dimension int) *VectorStore
func (vs *VectorStore) CreateIndex(ctx context.Context) error          // FT.CREATE
func (vs *VectorStore) Store(ctx context.Context, doc *KnowledgeDoc)    // HSET
func (vs *VectorStore) Search(ctx, source, queryVec, topK) []*SearchResult  // FT.SEARCH KNN
func (vs *VectorStore) DeleteDoc(ctx, docID, source) error             // SCAN + DEL
```

Search 核心：`FT.SEARCH idx:knowledge "@source:{admin|public}=>[KNN 5 @embedding $vec AS score]" RETURN 6 doc_id title content score tags source DIALECT 2`

---

### Task 5: 实现文档分块器（TextSplitter）

**Files:**
- Create: `app/blades/internal/biz/text_splitter.go`

分层分块策略：
- 一级：按 Markdown 标题（`##`, `###`）分块
- 二级：超过 800 tokens 的块递归按段落→句子→字符切割
- 重叠：相邻块末尾 10-20% 追加到下一块开头

```go
type TextSplitterConfig struct {
    ChunkSize    int     // 默认 800
    ChunkOverlap float64 // 默认 0.15
}
func NewTextSplitter(cfg TextSplitterConfig) *TextSplitter
func (ts *TextSplitter) Split(docTitle, content string) []*Chunk

type Chunk struct {
    Index   int
    Content string
    Title   string  // 来源标题
}
```

辅助函数 `estimateTokens(text string) int`：中英文混排按 2 chars/token 估算。

---

### Task 6: 实现 RAG 引擎

**Files:**
- Create: `app/blades/internal/biz/rag.go`

```go
type RAGEngine struct {
    embedder *EmbeddingClient
    vectorDB *data.VectorStore
    splitter *TextSplitter
    config   *conf.RAGConfig
}

func NewRAGEngine(embedder, vectorDB, cfg) *RAGEngine

// IndexDocument 索引文档：分块 → Embedding → 存向量库
func (e *RAGEngine) IndexDocument(ctx, docID, source, title, content string, tags map[string]string) (chunkCount int, err error)

// Retrieve 检索：问题 Embedding → KNN 搜索 → 组装上下文
func (e *RAGEngine) Retrieve(ctx, source, query string, topK int) (context string, err error)

// ShouldRetrieve 判断是否需要检索
func (e *RAGEngine) ShouldRetrieve(message string) bool
```

ShouldRetrieve 判断：如果 always_retrieve=true 始终检索；否则检测是否包含"规则/政策/怎么/如何"等关键词。

---

### Task 7: 实现工具系统

**Files:**
- Create: `app/blades/internal/biz/tools.go`
- Create: `app/blades/internal/biz/tools_order.go`
- Create: `app/blades/internal/biz/tools_product.go`
- Create: `app/blades/internal/biz/tools_user.go`
- Create: `app/blades/internal/biz/tools_marketing.go`

- [ ] **Step 1: ToolRegistry 工具注册中心**

```go
type ToolFactory func() (blades.Tool, error)

type ToolRegistry struct {
    factories map[string]ToolFactory
}
func NewToolRegistry() *ToolRegistry
func (r *ToolRegistry) Register(name string, factory ToolFactory)
func (r *ToolRegistry) GetToolsForSource(source string) ([]blades.Tool, error)
```

- [ ] **Step 2: 订单工具 `tools_order.go`**

```go
func RegisterOrderTools(r *ToolRegistry, client orderRpc.OrderClient) {
    r.Register("query_orders", func() (blades.Tool, error) {
        return tools.NewFunc(
            "query_orders",
            "按日期范围/状态查询订单列表和统计",
            func(ctx context.Context, input struct {
                Page int32, PageSize int32, Status string, DateFrom string, DateTo string
            }) (string, error) {
                reply, err := client.ListOrder(ctx, &orderRpc.ListOrderRequest{...})
                return formatOrderList(reply), nil
            },
        )
    })
    r.Register("query_order_detail", func() (blades.Tool, error) {
        return tools.NewFunc(...)   // GetOrder
    })
}
```

- [ ] **Step 3: 商品工具 `tools_product.go`**

```go
func RegisterProductTools(r *ToolRegistry, client productRpc.ProductClient) {
    r.Register("search_products", ...)  // ListProduct
    r.Register("get_product_detail", ...) // GetProduct
}
```

- [ ] **Step 4: 用户工具 `tools_user.go`**

```go
func RegisterUserTools(r *ToolRegistry, client userRpc.UserClient) {
    r.Register("query_users", ...) // UserCount + 按条件查
}
```

- [ ] **Step 5: 营销工具 `tools_marketing.go`**

```go
func RegisterMarketingTools(r *ToolRegistry, client marketingRpc.CouponServiceClient) {
    r.Register("query_coupons", ...) // 优惠券查询
}
```

---

### Task 8: 实现 Agent Factory

**Files:**
- Create: `app/blades/internal/biz/agent_factory.go`

- [ ] **Step 1: AgentFactory 实现**

```go
type AgentFactory struct {
    mu      sync.RWMutex
    agents  map[string]*AgentInstance  // key: source
    cfg     *conf.Blades
    tools   *ToolRegistry
    rag     *RAGEngine
}

type AgentInstance struct {
    source       string
    agent        blades.Agent
    prompt       string
    tools        []blades.Tool
    knowledgeTag string   // admin_kb / user_kb / public
}

func NewAgentFactory(cfg *conf.Blades, tools *ToolRegistry, rag *RAGEngine) *AgentFactory

func (f *AgentFactory) GetOrCreate(ctx context.Context, source string) (*AgentInstance, error)
```

不同端的配置差异：

| source | system_prompt | 工具集 | 知识库标签 |
|--------|---------------|--------|-----------|
| admin | "你是商城管理助手，可以查订单/商品/用户数据，进行运营操作" | order + product + user + marketing | admin_kb + public |
| user | "你是商城购物助手，帮助用户解答购物问题" | 仅查询类 + coupon | user_kb + public |
| other | "你是商城通用助手" | 仅搜索知识库 | public |

- [ ] **Step 2: AgentInstance.Chat 方法**

```go
func (inst *AgentInstance) Chat(ctx context.Context, userID int64, message string, enableRAG bool) (string, error) {
    // 1. RAG 检索
    if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
        contextStr, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
        if contextStr != "" {
            message = contextStr + "\n\n用户问题：" + message
        }
    }
    // 2. 加载历史
    // 3. 调用 Agent
    // 4. 保存历史
}
func (inst *AgentInstance) ChatStream(ctx, userID, message, enableRAG, chunkCb) error
```

---

### Task 9: 重建 AgentUseCase

**Files:**
- Rewrite: `app/blades/internal/biz/agent.go`

AgentUseCase 改为组合 AgentFactory + SessionManager：

```go
type AgentUseCase struct {
    factory *AgentFactory
    rdb     redis.UniversalClient
    log     *log.Helper
}

func NewAgentUseCase(cfg *conf.Blades, rdb redis.UniversalClient, factory *AgentFactory, logger log.Logger) (*AgentUseCase, error)

func (uc *AgentUseCase) Chat(ctx, userID int64, source, message string, enableRAG bool) (string, error)
func (uc *AgentUseCase) ChatStream(ctx, userID int64, source, message string, enableRAG bool, chunkCb func(string,bool)) error
func (uc *AgentUseCase) ClearHistory(ctx, userID int64, source string) error
func (uc *AgentUseCase) LoadChatHistory(ctx, userID int64, source string) ([]*HistoryMessage, error)
```

---

### Task 10: 实现知识库管理用例

**Files:**
- Create: `app/blades/internal/biz/knowledge.go`

```go
type KnowledgeUseCase struct {
    rag  *RAGEngine
    rdb  redis.UniversalClient
    log  *log.Helper
}

func NewKnowledgeUseCase(rag *RAGEngine, rdb redis.UniversalClient, logger log.Logger) *KnowledgeUseCase

// UploadDocument 上传并索引文档
func (uc *KnowledgeUseCase) UploadDocument(ctx, docID, source, title, contentType string, content []byte, tags map[string]string) (chunkCount, tokenCount int, err error)

// SearchKnowledge 搜索知识库
func (uc *KnowledgeUseCase) SearchKnowledge(ctx, source, query string, topK int) ([]*SearchResult, error)

// DeleteDocument 删除文档
func (uc *KnowledgeUseCase) DeleteDocument(ctx, docID, source string) error
```

---

### Task 11: 更新 Service 层

**Files:**
- Rewrite: `app/blades/internal/service/blades.go`

```go
type BladesService struct {
    pb.UnimplementedBladesServer
    agentUC    *biz.AgentUseCase
    knowledgeUC *biz.KnowledgeUseCase
    log        *log.Helper
}

// Chat / ChatStream / GetChatHistory / ClearChatHistory 委托给 agentUC
// UploadKnowledge / SearchKnowledge / DeleteKnowledge 委托给 knowledgeUC
```

ChatStream 实现要点（gRPC 服务端流式）：

```go
func (s *BladesService) ChatStream(req *pb.ChatStreamRequest, stream pb.Blades_ChatStreamServer) error {
    err := s.agentUC.ChatStream(stream.Context(), req.UserId, req.Source, req.Message, req.EnableRag,
        func(chunk string, done bool) {
            stream.Send(&pb.ChatChunkReply{Chunk: chunk, Done: done})
        })
    return err
}
```

---

### Task 12: 更新 Wire 注入

**Files:**
- Modify: `app/blades/internal/biz/biz.go`
- Modify: `app/blades/internal/data/data.go`
- Modify: `app/blades/internal/service/service.go`
- Regenerate: `app/blades/cmd/blades/wire_gen.go`

- [ ] **Step 1: 更新 biz/biz.go ProviderSet**

```go
var ProviderSet = wire.NewSet(
    NewBladesUseCase,
    NewAgentUseCase,
    NewAgentFactory,
    NewToolRegistry,
    NewRAGEngine,
    NewEmbeddingClient,
    NewKnowledgeUseCase,
    NewTextSplitter,
)
```

- [ ] **Step 2: 更新 service/service.go ProviderSet**

```go
var ProviderSet = wire.NewSet(NewBladesService)
```

- [ ] **Step 3: 更新 data/data.go ProviderSet**

```go
var ProviderSet = wire.NewSet(
    NewBladesRepo,
    NewData,
    NewRedisClient,
    ProvideUniversalClient,
    NewVectorStore,
    NewProductServiceClient,
    NewUserServiceClient,
    NewOrderServiceClient,
)
```

注意 NewData 需要额外注入 VectorStore 等相关依赖。

- [ ] **Step 4: 重新生成 wire**

```bash
cd app/blades/cmd/blades; wire
```

---

### Task 13: 编译验证

- [ ] **Step 1: 编译 blades**

```bash
cd app/blades; go build ./...
```

Expected: 编译成功，无错误。

- [ ] **Step 2: 编译 admin（确认 BFF 侧兼容）**

```bash
cd app/cin/admin; go build ./...
```

Expected: 编译成功，admin 调用 blades 的接口不变。
