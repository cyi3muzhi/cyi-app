# Blades 服务

AI 智能助手服务，基于 DeepSeek API 提供聊天对话能力。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL、Redis、OSS |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |
| AI 接口 | DeepSeek API |

## 目录结构

```
app/blades/
├── cmd/blades/
│   ├── main.go              # 服务入口
│   ├── confadapter.go       # Consul 配置适配
│   ├── wire.go              # 依赖注入配置
│   └── wire_gen.go          # Wire 生成文件
├── configs/
│   ├── auth.yaml            # 认证配置
│   ├── config.yaml          # 主配置
│   └── register.yaml        # 注册配置
├── internal/
│   ├── biz/
│   │   ├── agent.go             # AgentUseCase 会话管理 + 偏好存储
│   │   ├── agent_factory.go     # AgentFactory 多端 Agent 工厂
│   │   ├── agent_middleware.go  # Agent 中间件（日志/恢复）
│   │   ├── agent_test.go        # Agent 测试
│   │   ├── biz.go               # 业务层依赖注入
│   │   ├── config.go            # 配置加载
│   │   ├── context.go           # 上下文工具（userID/source/role）
│   │   ├── embedding.go         # Embedding 客户端
│   │   ├── knowledge.go         # 知识库管理用例
│   │   ├── providers.go         # Embedding/RAG 配置 provider
│   │   ├── rag.go               # RAG 引擎（检索增强生成）
│   │   ├── text_splitter.go     # 文档分块器
│   │   ├── tools.go             # 工具注册中心
│   │   ├── tools_knowledge.go   # 知识库搜索工具
│   │   ├── tools_order.go       # 订单查询工具
│   │   ├── tools_permission.go  # 权限查询工具
│   │   ├── tools_preference.go  # 偏好保存工具
│   │   ├── tools_product.go     # 商品查询工具
│   │   ├── tools_stock.go       # 库存查询工具
│   │   └── tools_user.go        # 用户查询工具
│   ├── conf/
│   │   ├── conf.pb.go       # 配置结构体
│   │   └── conf.proto       # 配置定义
│   ├── data/
│   │   ├── client_conn.go   # gRPC 客户端连接
│   │   ├── data.go          # 数据层初始化
│   │   └── vector_store.go  # 向量存储实现
│   ├── server/
│   │   ├── grpc.go          # gRPC 服务注册
│   │   ├── http.go          # HTTP 服务注册
│   │   └── server.go        # 服务器配置
│   └── service/
│       ├── blades.go        # gRPC 服务实现
│       └── service.go       # 服务层依赖注入
├── Makefile                 # 构建脚本
└── openapi.yaml             # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `blades.api_key` | DeepSeek API 密钥 | - |
| `blades.model` | 模型名称 | deepseek-chat |
| `blades.max_tokens` | 最大令牌数 | - |
| `blades.temperature` | 温度参数 | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |

## 核心业务逻辑

### AgentUseCase 和 AgentFactory

多端 Agent 管理，每端（admin/user/other）有独立的系统提示词和工具集。

- `AgentFactory.GetOrCreate(ctx, source, role)`: 按端和角色获取或创建 Agent 实例（3 分钟缓存，动态按权限过滤工具）
- `AgentInstance.Chat/ChatStream`: 发送消息（支持 RAG + 会话历史 + 工具调用 + 偏好注入）
- `AgentLogging/AgentRecovery`: Agent 中间件（日志记录 + panic 恢复）
- **输出格式规范**：所有场景提示词统一拼接 `formatRules`，要求 AI 使用 Markdown 结构化输出（查询结果用 `|` 管道表格、要点用列表、结论置前、禁止客套寒暄），保证回复在前端（marked 渲染）展示整齐

### KnowledgeUseCase

知识库管理，支持文档上传索引、向量检索。

- `UploadDocument`: 上传文档并分块索引到向量库
- `SearchKnowledge`: 通过向量相似度搜索知识
- `DeleteDocument`: 删除已索引的文档

### 工具集（按角色动态启用）

| 工具 | 说明 | 所需权限 |
|------|------|----------|
| `search_products` | 搜索查询商品列表 | products.view |
| `get_product_detail` | 查询商品详细信息 | products.view |
| `query_stock` | 查询 SKU 实时库存 | products.view |
| `query_orders` | 查询订单列表 | orders.view |
| `query_order_detail` | 查询订单详情 | orders.view |
| `query_users` | 查询用户列表 | users.view |
| `search_knowledge` | 搜索知识库 | 通用（不限制） |
| `save_user_preference` | 保存用户偏好设置 | 通用（不限制） |
| `query_user_permissions` | 查询当前用户角色和权限 | 通用（不限制） |
| `check_user_permission` | 检查具体操作的权限 | 通用（不限制） |

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `Chat` | gRPC/HTTP | 非流式聊天，支持 RAG + 工具调用 |
| `ChatStream` | gRPC/HTTP 流式 | 流式聊天，支持 RAG + 工具调用 |
| `GetChatHistory` | gRPC/HTTP | 获取聊天历史 |
| `ClearChatHistory` | gRPC/HTTP | 清除聊天历史 |
| `UploadKnowledge` | gRPC 流式 | 上传文档到知识库 |
| `SearchKnowledge` | gRPC/HTTP | 搜索知识库 |
| `DeleteKnowledge` | gRPC/HTTP | 删除知识文档 |

## 运行方式

### 本地开发

```bash
cd app/blades
go run cmd/blades/main.go -conf configs/
```

### 构建部署

```bash
cd app/blades
make build
./bin/blades -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-blades`

## 服务发现

服务注册到 Consul，服务名：`blades`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`