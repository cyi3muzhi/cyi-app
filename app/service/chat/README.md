# Chat 服务

IM 客服服务（方案 A：用户 ↔ 客服）。提供会话管理（创建、列表、接入、转接、结束）、消息收发（文本/图片/商品卡片/订单卡片）、历史消息游标分页、已读标记与客服接待统计，支撑用户端咨询入口与管理后台客服工作台。

> 当前进度：M1 基础服务 + M2 实时推送 + M3 客服工作台 + M4 AI 兜底与在线状态 + 离线提醒联动消息中心。IM 客服模块（方案 A）已全部落地，见 [IM 客服模块详细设计](../../../docs/im-customer-service-design.md)。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL（GORM + gorm/gen 代码生成） |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/service/chat/
├── cmd/chat/
│   ├── main.go          # 服务入口（配置从 Consul 加载）
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── conversation.go # 会话/消息业务实体与状态常量
│   │   ├── conversation_handle.go # 会话用例（ConversationUseCase）
│   │   └── message_handle.go # 消息用例（MessageUseCase）
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义（含 NATS/Blades/Agent 预留项）
│   ├── data/
│   │   ├── curd/
│   │   │   ├── gen.go   # gorm/gen 代码生成工具
│   │   │   ├── conversation.gen.go # 会话 CURD
│   │   │   └── chat_message.gen.go # 消息 CURD
│   │   ├── data.go      # 数据层初始化（非生产自动迁移表结构）
│   │   ├── util.go      # 分页等通用工具
│   │   ├── models/
│   │   │   ├── conversation.go # 会话模型
│   │   │   └── chat_message.go # 消息模型
│   │   ├── conversation.go # 会话数据访问
│   │   └── chat_message.go # 消息数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── chat.go      # 聊天服务（ChatService）
│       └── service.go   # 服务层依赖注入
├── Makefile             # 构建脚本（复用根目录 app_makefile.mk）
```

## 配置说明

服务配置统一存储在 Consul KV 中（`/develop/cyi-app/service/chat/`），启动时自动加载。主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.grpc.addr` | gRPC 监听地址 | - |
| `server.grpc.timeout` | gRPC 超时时间 | - |
| `data.database.driver` | 数据库驱动 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址（M3 未读/在线状态用，预留） | - |
| `auth.api_key` | 服务间调用 API 密钥 | - |
| `auth.service_key` | 各服务调用方密钥映射 | - |
| `trace.endpoint` | Jaeger 链路追踪端点 | - |
| `log.level` | 日志级别（debug/info/warn/error） | - |
| `log.enable_console` | 是否启用控制台输出 | - |
| `log.enable_file` | 是否启用文件输出 | - |
| `log.log_dir` | 日志目录 | - |
| `registry.consul.address` | Consul 注册中心地址 | - |
| `registry.consul.scheme` | Consul 协议（http/https） | - |
| `nats.addr` | NATS 消息总线（M2 消息推送用，预留） | - |
| `blades.endpoint` | Blades AI 服务 gRPC 地址（M4 AI 兜底用，预留） | - |
| `agent.max_sessions` | 客服最大同时接待会话数（M3 分配用，预留） | 10 |
| `agent.ai_fallback` | 无客服在线时是否启用 AI 兜底（M4，预留） | true |
| `agent.online_ttl` | 在线状态 TTL（秒，M3 用，预留） | 90 |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `ConversationUseCase` | 会话全生命周期（创建、用户/客服列表、待接入、接入、转接、结束）与接待统计 |
| `MessageUseCase` | 消息收发（用户/客服）、历史消息游标分页、已读标记 |

### 业务模型

- **会话状态**：1-待接入、2-进行中、3-已结束
- **接待类型**：1-人工、2-AI 接待中（M4 启用）
- **发送者类型**：1-用户、2-客服、3-系统、4-AI
- **消息类型**：1-文本、2-图片、3-商品卡片、4-订单卡片、5-系统提示

### 关键设计

- **防重复建会话**：`CreateConversation` 复用用户未结束的会话
- **防并发抢单**：`AcceptConversation` 条件更新（仅待接入状态可接入，RowsAffected=0 视为已被接入）
- **成员校验**：拉取历史/发消息/已读/结束均校验请求方是否为会话成员且会话未结束
- **游标分页**：`ListMessage` 支持向上（更早）/向下（更新）双向翻页，`cursor_id=0` 取最新一页
- **未读数**：迁移至 Redis（`chat:unread:conv:{convID}:{user|agent}`，发消息对端 +1、标记已读清零、会话结束清理）；列表读取 Redis 优先，Redis 无记录时回退 DB 统计兜底
- **实时推送**：消息落库并更新会话摘要后，由 `ChatMessagePublisher` 发布 NATS `chat.message` 事件（含接收方类型与 ID），interface/admin 网关订阅后推送给对应连接；NATS 不可用时仅记日志、不影响消息落库
- **会话状态事件**：会话新建/接入/转接/结束由 `ConversationEventPublisher` 发布 NATS `chat.conversation`，客服工作台据此实时刷新待接入/会话列表
- **客服端历史**：`ListAgentMessage` 独立于用户端 `ListMessage`，按客服成员校验拉取历史
- **客服会话筛选**：`ListAgentConversation` 支持按状态（0=全部 1=待接入 2=进行中 3=已结束）筛选
- **AI 兜底**：新会话默认 `service_type=2`（AI 接待中），用户发消息由 `AIChatUseCase` 异步调用 blades 生成回复（sender_type=4）落库并推送；客服 `AcceptConversation` 后切换人工接待、停止 AI
- **在线状态**：用户/客服 WS 连接注册时写入 Redis `chat:online:user:{uid}` / `chat:online:agent:{id}`（TTL 90s），心跳续期、注销清理
- **离线提醒**：客服发消息时用户不在线，则通过 user 服务消息中心发系统站内信提醒（携带会话ID，用户端消息中心点击直接跳转对应会话）；同一会话去重（Redis `chat:remind:` key，TTL 7 天），用户标记已读后清除、可再次提醒

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateConversation` | gRPC | 创建会话（用户端） |
| `ListMyConversation` | gRPC | 我的会话列表（用户端） |
| `ListMessage` | gRPC | 会话历史消息（游标分页，用户端） |
| `ListAgentMessage` | gRPC | 会话历史消息（客服端，校验客服成员） |
| `SendMessage` | gRPC | 用户发送消息 |
| `MarkConversationRead` | gRPC | 用户标记会话已读 |
| `CloseConversation` | gRPC | 用户结束会话 |
| `ListPendingConversation` | gRPC | 待接入会话列表（客服工作台） |
| `AcceptConversation` | gRPC | 客服接入会话 |
| `TransferConversation` | gRPC | 客服转接会话 |
| `ListAgentConversation` | gRPC | 客服名下会话列表 |
| `SendAgentMessage` | gRPC | 客服发送消息 |
| `MarkAgentRead` | gRPC | 客服标记会话已读 |
| `CloseAgentConversation` | gRPC | 客服结束会话 |
| `AgentStats` | gRPC | 客服接待统计（今日接入/进行中/今日解决） |

## 运行方式

### 本地开发

```bash
cd app/service/chat
go run cmd/chat/main.go
```

> 配置从 Consul KV 自动加载，需先启动 Consul 并在 `/develop/cyi-app/service/chat/` 路径下配置。

### 生成代码

```bash
# 在 app/service/chat 目录下
make api      # 生成 api/service/chat/v1 的 proto 代码
make errors   # 生成错误码
make config   # 生成 internal/conf 的配置代码
```

### 构建部署

```bash
cd app/service/chat
make build
./bin/chat
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-chat`

## 服务发现

服务注册到 Consul，服务名：`chat`（注册完整名称为 `{SERVICE_ENV}.chat.service`）

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`
