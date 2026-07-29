# Blades AI 智能体服务设计文档

> 版本：v1.0
> 日期：2026-07-28
> 状态：草稿

## 1. 概述

### 1.1 背景

Blades 服务是商城体系中的统一 AI 入口，所有 BFF（管理端、用户端、后续其他端）的 AI 请求统一经过 blades 处理。blades 需要提供 RAG（检索增强生成）、工具调用（Tool Calling）、多端隔离等能力，并与现有微服务体系（product、order、user、marketing）深度集成。

### 1.2 目标

- 统一的 AI 接入层，支持多端（admin/user/other）隔离
- RAG 能力，基于知识库检索增强 AI 回答质量
- 工具调用，将下游服务 gRPC 包装成 AI 可调用的工具
- 会话管理，分端分用户独立会话
- 可扩展架构，预留编排层（串联/并联/子Agent）接口

### 1.3 非目标

- 不涉及用户认证（由 BFF 层完成 JWT 校验）
- 不直接连接数据库（通过 gRPC 调下游服务）
- 第一阶段不实现编排层（串联/并联/循环），但预留扩展点

## 2. 架构设计

### 2.1 整体架构

```
┌──────────────────────────────────────────────────────────────┐
│                         BFF 层                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ admin BFF   │  │ user BFF    │  │ other BFF   │          │
│  │ (Gin/Kratos)│  │ (Kratos)    │  │             │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                │                │                  │
└─────────┼────────────────┼────────────────┼──────────────────┘
          │     gRPC       │     gRPC       │      gRPC
          │ +source=admin  │ +source=user   │  +source=other
          ▼                ▼                ▼
┌──────────────────────────────────────────────────────────────┐
│                    Blades Service                             │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │                 Agent Factory                         │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │    │
│  │  │ Admin    │  │ User     │  │ Other    │           │    │
│  │  │ Agent    │  │ Agent    │  │ Agent    │           │    │
│  │  └─────┬────┘  └─────┬────┘  └─────┬────┘           │    │
│  └────────┼──────────────┼──────────────┼───────────────┘    │
│           │              │              │                     │
│  ┌────────┴──────────────┴──────────────┴───────────────┐    │
│  │                    RAG Engine                          │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │    │
│  │  │ Embedding    │  │ Vector Store │  │ Retriever  │  │    │
│  │  │ (DeepSeek)   │  │ (RedisStack) │  │ (Context)  │  │    │
│  │  └──────────────┘  └──────────────┘  └────────────┘  │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │                 Tool System                            │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │    │
│  │  │ Order    │  │ Product  │  │ User     │           │    │
│  │  │ Tool     │  │ Tool     │  │ Tool     │           │    │
│  │  └──────────┘  └──────────┘  └──────────┘           │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │    │
│  │  │Marketing │  │Knowledge │  │ Calculator│          │    │
│  │  │ Tool     │  │ Tool     │  │ Tool     │           │    │
│  │  └──────────┘  └──────────┘  └──────────┘           │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐    │
│  │              Session Manager                           │    │
│  │              Redis (会话历史)                          │    │
│  └──────────────────────────────────────────────────────┘    │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

**管理员对话完整流程**：

```
1. 管理员发消息 "昨天订单多少？"
2. admin BFF 做 JWT 校验 → 提取 adminID + role
3. admin BFF 调 blades.ChatStream(source=admin, user_id=adminID, message=...)
4. blades AgentFactory 根据 source=admin 获取 Admin Agent 实例
5. Admin Agent 调用 RAG Engine：
   a. 选知识库 → admin_kb（管理员知识库）
   b. DeepSeek Embedding API 将问题转向量
   c. Redis Stack KNN 搜索，召回相关文档
   d. 组装上下文注入 system prompt
6. Admin Agent 调用 Tool System：
   a. 检测到是数据查询 → 调 Order Tool
   b. Order Tool → gRPC order.service.ListOrder → 返回数据
   c. 数据注入消息上下文
7. 增强后的消息发给 DeepSeek API
8. 流式返回 → admin BFF → SSE 推给前端
9. 会话保存到 Redis（按 adminID）
```

**用户端对话流程**（类似，但使用 User Agent）：

```
1. 用户发消息 "帮我查一下我的优惠券"
2. user BFF 做 JWT 校验 → 提取 userID
3. user BFF 调 blades.ChatStream(source=user, user_id=userID, message=...)
4. AgentFactory 根据 source=user 获取 User Agent
5. User Agent → RAG Engine（user_kb 知识库）
6. User Agent → Marketing Tool → gRPC marketing.service → 查用户优惠券
7. → DeepSeek → 返回
```

## 3. 基础设施变更

### 3.1 Redis 升级

| 当前 | 目标 |
|------|------|
| `redis:7.0.12` | `redis/redis-stack-server:7.2.0-v10` |

升级原因：Redis Stack 内置 RediSearch 模块，支持向量相似度搜索（VECTOR 类型 + KNN 查询）。

docker-compose 变更：

```yaml
redis:
  image: redis/redis-stack-server:7.2.0-v10
  ports:
    - "6379:6379"
```

Redis Stack 提供以下所需能力：

| 能力 | 用途 |
|------|------|
| FT.CREATE | 创建带 VECTOR 字段的索引 |
| FT.SEARCH KNN | 向量相似度搜索 |
| JSON.SET | 存储结构化文档元数据 |
| 现有 KV 操作 | 会话管理（不变） |

### 3.2 DeepSeek 模型依赖

| 用途 | 模型 |
|------|------|
| 对话生成 | deepseek-chat（已有） |
| 文本 Embedding | deepseek-embedding（新增） |

Embedding API 接口：`POST https://api.deepseek.com/v1/embeddings`

## 4. API 设计

### 4.1 Proto 定义

```protobuf
syntax = "proto3";
package api.blades.v1;

service Blades {
    // Chat 非流式聊天
    rpc Chat (ChatRequest) returns (ChatReply);
    // ChatStream 流式聊天（服务端流式）
    rpc ChatStream (ChatStreamRequest) returns (stream ChatChunkReply);
    // GetChatHistory 获取聊天历史
    rpc GetChatHistory (GetChatHistoryRequest) returns (GetChatHistoryReply);
    // ClearChatHistory 清除聊天历史
    rpc ClearChatHistory (ClearChatHistoryRequest) returns (ClearChatHistoryReply);
    
    // --- 知识库管理 ---
    // UploadKnowledge 上传知识文档
    rpc UploadKnowledge (stream UploadKnowledgeRequest) returns (UploadKnowledgeReply);
    // SearchKnowledge 搜索知识库（供管理后台使用）
    rpc SearchKnowledge (SearchKnowledgeRequest) returns (SearchKnowledgeReply);
    // DeleteKnowledge 删除知识文档
    rpc DeleteKnowledge (DeleteKnowledgeRequest) returns (DeleteKnowledgeReply);
}
```

**ChatRequest 改造**（核心变更）：

```protobuf
message ChatRequest {
    int64 user_id = 1;           // 用户ID（管理端=adminID，用户端=userID）
    string source = 2;           // 来源：admin / user / other
    string role = 3;             // 角色（管理端权限校验用）
    string message = 4;          // 用户消息
    bool enable_rag = 5;         // 是否启用 RAG（默认 true）
    repeated string extra_tools = 6; // 额外启用的工具列表（可选）
}
```

**知识库管理接口**：

```protobuf
message UploadKnowledgeRequest {
    string source = 1;           // admin / user / public
    string doc_id = 2;           // 文档ID（分片用）
    string title = 3;            // 文档标题
    bytes content = 4;           // 文档内容
    string content_type = 5;     // text/markdown/pdf
    map<string, string> tags = 6; // 标签
}

message UploadKnowledgeReply {
    string doc_id = 1;
    int32 chunk_count = 2;       // 分块数
}

message SearchKnowledgeRequest {
    string source = 1;           // admin / user / public
    string query = 2;            // 搜索关键词
    int32 top_k = 3;             // 返回条数（默认 5）
}

message SearchKnowledgeReply {
    repeated KnowledgeChunk chunks = 1;
}

message KnowledgeChunk {
    string doc_id = 1;
    string title = 2;
    string content = 3;
    float score = 4;             // 相关性分数
    map<string, string> tags = 5;
}

message DeleteKnowledgeRequest {
    string doc_id = 1;           // 文档ID
    string source = 2;           // admin / user / public
}

message DeleteKnowledgeReply {
    bool success = 1;
}
```

### 4.2 内部接口（预留编排层扩展）

```go
// WorkflowStep 编排步骤（第二阶段实现）
type WorkflowStep struct {
    Type      StepType           // chain / parallel / loop / sub_agent
    AgentName string             // 目标 Agent 名称
    Input     string             // 输入消息模板
    Tools     []string           // 可用工具列表
    Next      func(output string) bool  // 下一步判断条件
}

type StepType int
const (
    StepChain    StepType = iota  // 串联
    StepParallel                  // 并联
    StepLoop                      // 循环
    StepSubAgent                  // 子 Agent
)
```

## 5. 模块详细设计

### 5.1 Agent Factory

**职责**：根据 `source` 创建/缓存不同端的 Agent 实例。

```go
type AgentFactory struct {
    mu       sync.RWMutex
    agents   map[string]*AgentInstance   // key: source
    cfg      *conf.Blades               // 全局配置
    tools    *ToolRegistry              // 工具注册中心
    rag      *RAGEngine                 // RAG 引擎
    session  *SessionManager            // 会话管理
}

type AgentInstance struct {
    source       string                  // admin / user / other
    agent        blades.Agent           // go-kratos/blades Agent
    prompt       string                 // system prompt
    tools        []blades.Tool          // 绑定到该 Agent 的工具
    knowledgeTag string                 // 知识库标签（admin_kb / user_kb / public）
}

// GetOrCreate 获取或创建 Agent 实例
func (f *AgentFactory) GetOrCreate(ctx context.Context, source string) (*AgentInstance, error)
```

**不同端配置**：

| 端 | System Prompt 重点 | 工具集 | 知识库 |
|----|-------------------|--------|--------|
| admin | 数据查询、订单管理、商品管理、用户管理 | 全部工具（含写操作） | admin_kb + public |
| user | 购物咨询、优惠券、客服 | 查询类 + 优惠券 | user_kb + public |
| other | 通用 | 最小工具集 | public |

### 5.2 RAG Engine

**职责**：知识文档的向量化、存储、检索。

**核心结构**：

```go
type RAGEngine struct {
    redis       redis.UniversalClient
    embedClient *EmbeddingClient      // DeepSeek Embedding 封装
}

type EmbeddingClient struct {
    apiKey      string
    apiEndpoint string
    model       string                // deepseek-embedding
    batchSize   int                   // 批量 embedding 大小
}

// Embedding 结果
type Embedding struct {
    Vector    []float32               // 768 维向量
    Dimension int
}
```

**文档处理流程**：

```
上传文档 → 文档格式解析（markdown/txt）
  → 文档结构分块（一级）
    按 Markdown 标题（## / ###）分割
    每个标题为一个语义独立的文档块
  → 递归字符兜底（二级）
    若某块超过 max_chunk_size（800 tokens）
    递归按段落 → 句子 → 字符切割
  → 语义分块（可选增强）
    检测块之间的语义连贯性
    若相邻块主题相近则合并，避免断句
  → Chunk 重叠 10-20%（令牌级别）
    相邻块之间保留前一块末尾 10-20% 的内容
    确保上下文连续，不丢失跨块信息
  → 批量调用 DeepSeek Embedding API
  → 存 Redis Stack：
    - FT.CREATE idx:knowledge ON HASH PREFIX 1 kb:
    - SCHEMA content TEXT embedding VECTOR FLAT 6 TYPE FLOAT32 DIM 768
    - HSET kb:{doc_id}:{chunk_idx} content ... embedding ...
```

**分块策略总结**：

| 层级 | 方法 | 触发条件 |
|------|------|----------|
| 一级 | 文档结构分块 | 按 Markdown 标题分割（语义独立） |
| 二级 | 递归字符分块 | 某块超过 800 tokens 时 |
| 重叠 | 10-20% 令牌重叠 | 所有块之间 |
| 增强 | 语义分块 | 可选，检测相邻块主题连续性 |

**检索流程**：

```
用户问题 → EmbeddingClient.Embed(question)
  → FT.SEARCH idx:knowledge 
    "@type:{source} => KNN 5 $vec AS score"
  → 返回 Top-K 块
  → 组装为 context 字符串
```

### 5.3 Tool System

**职责**：将下游服务 gRPC 调用包装为 AI 可调用的工具。

**工具定义方式**：

```go
// NewOrderTool 创建订单查询工具
func NewOrderTool(client orderRpc.OrderClient) (blades.Tool, error) {
    return tools.NewFunc(
        "query_orders",
        "查询订单数据，支持按日期范围、状态筛选",
        func(ctx context.Context, input QueryOrderInput) (QueryOrderOutput, error) {
            // 内部调 order.service gRPC
            reply, err := client.ListOrder(ctx, &orderRpc.ListOrderRequest{
                Page:     input.Page,
                PageSize: input.PageSize,
                Status:   input.Status,
                DateFrom: input.DateFrom,
                DateTo:   input.DateTo,
            })
            // 格式化返回
            return QueryOrderOutput{...}, nil
        },
    )
}
```

**初期工具清单**：

| 工具名 | 服务 | 说明 |
|--------|------|------|
| query_orders | order.service | 查订单列表/统计 |
| query_order_detail | order.service | 查订单详情 |
| search_products | product.service | 商品搜索/列表 |
| get_product_detail | product.service | 商品详情 |
| query_users | user.service | 用户查询/统计 |
| query_coupons | marketing.service | 优惠券查询 |
| search_knowledge | 内置 | RAG 知识库搜索 |
| calculator | 内置 | 数学计算 |

**工具注册**：

```go
type ToolRegistry struct {
    mu     sync.RWMutex
    tools  map[string]ToolFactory   // 工具名 → 工厂函数
}

// Register 注册工具
func (r *ToolRegistry) Register(name string, factory ToolFactory)

// GetToolsForSource 获取某端可用的工具列表
func (r *ToolRegistry) GetToolsForSource(source string) []blades.Tool
```

**可扩展性**：
- 新服务加入时，只需在新服务目录下实现 `ToolFactory` 并注册
- 每个工具独立 gRPC 调用，互不影响
- 预留工具版本管理接口（后续可以做到热更新）

### 5.4 Session Manager

沿用现有设计，增加 `source` 维度：

```go
func sessionKey(source string, userID int64) string {
    return fmt.Sprintf("session:%s:%d", source, userID)
}
```

## 6. 目录结构

```
app/blades/
├── cmd/blades/
│   ├── main.go
│   ├── wire.go
│   └── wire_gen.go
├── internal/
│   ├── biz/
│   │   ├── agent.go              # AgentUseCase（核心对话逻辑）
│   │   ├── agent_factory.go      # Agent 工厂（多端管理）
│   │   ├── tools.go              # 工具注册中心
│   │   ├── tools_order.go        # 订单工具
│   │   ├── tools_product.go      # 商品工具
│   │   ├── tools_user.go         # 用户工具
│   │   ├── tools_marketing.go    # 营销工具
│   │   ├── tools_knowledge.go    # 知识库搜索工具
│   │   ├── rag.go                # RAG 引擎
│   │   ├── embedding.go          # Embedding 客户端
│   │   ├── knowledge.go          # 知识库管理用例
│   │   └── biz.go                # wire ProviderSet
│   ├── data/
│   │   ├── data.go               # Data 结构体（Redis + gRPC clients）
│   │   ├── client_conn.go        # gRPC 客户端连接
│   │   ├── vector_store.go       # Redis Stack 向量存储操作
│   │   └── knowledge_repo.go     # 知识库数据访问
│   ├── conf/                     # 配置（不变）
│   ├── server/                   # 服务注册（不变）
│   └── service/
│       ├── blades.go             # gRPC 接口实现
│       └── service.go            # wire ProviderSet
├── configs/
│   ├── config.yaml
│   ├── auth.yaml
│   └── register.yaml
└── api/blades/v1/
    └── blades.proto              # 更新后的 API 定义
```

## 7. 配置变更

### 7.1 config.yaml 新增

```yaml
blades:
  api_key: xxx
  model: deepseek-chat
  max_tokens: 4096
  temperature: 0.7
  embedding:
    api_key: xxx                    # 与 chat 共用，可选单独配置
    model: deepseek-embedding
    dimension: 768
  rag:
    top_k: 5                        # 默认召回条数
    chunk_size: 800                 # 分块大小（tokens）
    chunk_overlap: 50               # 分块重叠
    always_retrieve: false          # 是否始终检索
  knowledge:
    max_docs_per_source: 1000       # 每端知识库最大文档数
```

## 8. 可扩展性设计

### 8.1 编排层扩展点（第二阶段）

当前设计预留的扩展接口：

```go
// AgentFactory 预留：
type AgentInstance struct {
    // ... 现有字段
    // 第二阶段新增：
    Orchestrator *WorkflowOrchestrator  // nil 时做单 Agent
}

// Tool 扩展：
type ToolResult struct {
    Data    interface{}
    // 第二阶段新增：
    SubAgent string        // 是否委派给子 Agent
    NextStep *WorkflowStep // 下一步
}
```

### 8.2 知识库扩展

- 支持多种文档格式（txt / markdown / pdf 逐步支持）
- 支持知识库版本管理（后续实现）
- 支持定时重新索引（后续实现）

## 9. 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| DeepSeek Embedding API 限流 | 大量文档上传时可能失败 | 批量 + 重试 + 限速 |
| Redis Stack 内存占用 | 向量数据占用内存 | 设置 maxmemory-policy，控制知识库大小 |
| Token 消耗 | RAG + 工具调用增加 token 使用 | 系统提示词控制、按需检索、缓存 |
| Embedding 延迟 | 每次对话增加 Embedding API 调用 | 缓存常见问题的 embedding 结果 |

## 10. 后续规划

### v1.0（当前阶段）
- [x] 基础 gRPC 接口（Chat/ChatStream/History）
- [x] 会话管理（Redis）
- [x] 多端基础支持（source 参数）
- [x] 下游服务 gRPC 客户端
- [ ] Tool Calling 系统
- [ ] RAG 引擎
- [ ] Agent Factory（多端隔离）
- [ ] Redis Stack 升级
- [ ] 知识库管理接口

### v2.0（后续）
- [ ] 编排层（串联/并联/循环/子 Agent）
- [ ] 工具版本管理与热更新
- [ ] 知识库版本管理
- [ ] 对话质量监控与评估
- [ ] A/B 测试框架
