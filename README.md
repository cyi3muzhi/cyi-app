# CYI-APP 微服务系统

基于 Go + Kratos 框架的电商微服务系统，采用 DDD 领域驱动设计，集成 AI 智能助手。

## 架构图

![CYI-APP 微服务架构](docs/architecture.png)

> 💡 **高清交互版**：本地打开 [docs/architecture.html](docs/architecture.html) 查看完整 SVG 架构图（可缩放、可交互）

<details>
<summary>📊 GitHub 内嵌 Mermaid 图（点击展开）</summary>

```mermaid
flowchart TB
    subgraph 接入层
        FE1[用户端前端<br/>H5 / 小程序]
        FE2[管理后台前端<br/>web / admin]
    end

    subgraph 网关层
        GW1{CIN Interface<br/>接口网关}
        GW2{CIN Admin<br/>管理后台 BFF}
    end

    subgraph 业务服务层
        SVC1[User Service<br/>用户认证]
        SVC2[Blades Service<br/>AI 智能助手]
        SVC3[Chat Service<br/>IM 客服]
        SVC4[Product Service<br/>商品库存]
        SVC5[Marketing Service<br/>优惠券营销]
        SVC6[Order Service<br/>订单管理]
    end

    subgraph 消息任务层
        MSG1[NATS Consumer]
        TASK[Task Service<br/>异步 / 定时任务]
    end

    subgraph 基础设施层
        INFRA[(MySQL · Redis · NATS<br/>Consul · Jaeger · Prometheus<br/>Grafana · ELK · RustFS)]
    end

    FE1 -->|HTTP/WS/SSE| GW1
    FE2 -->|HTTP| GW2

    GW1 -->|gRPC| SVC1
    GW1 -->|AI Call| SVC2
    GW1 -->|gRPC| SVC3
    GW1 -->|gRPC| SVC4
    GW1 -->|gRPC| SVC6

    GW2 -->|gRPC| SVC1
    GW2 -->|gRPC| SVC4
    GW2 -->|gRPC| SVC5
    GW2 -->|gRPC| SVC6
    GW2 -->|Asynq 投递| TASK

    SVC1 & SVC2 & SVC3 & SVC4 & SVC5 & SVC6 -.->|数据存取| INFRA
    MSG1 -->|操作日志| SVC1
    TASK -.->|gRPC| SVC1 & SVC6

    classDef default fill:#ffffff,stroke:#2d3142,stroke-width:1px
    classDef gateway fill:rgba(191,69,32,0.08),stroke:#bf4520,stroke-width:1.5px
    classDef ai fill:rgba(191,69,32,0.04),stroke:rgba(191,69,32,0.60),stroke-width:1px
    classDef infra fill:rgba(45,49,66,0.05),stroke:#4f5d75,stroke-width:1px

    class GW1,GW2 gateway
    class SVC2 ai
    class INFRA infra
```

</details>

> **说明**：
> - **CIN Interface**：用户端统一入口，聚合调用各业务服务
> - **CIN Admin**：管理后台 BFF，独立生成 JWT、独立权限管控（Casbin RBAC）
> - **Task Service**：消费 Asynq 的 `default`（一般异步任务）、`export`（订单导出）、`critical`（cron 定时任务）队列
> - **Admin → Task**：管理后台导出订单时，通过 Asynq（Redis）投递任务，task 服务消费完成后写回 Redis，admin 轮询状态
> - **操作日志链路**：admin 服务 HTTP 中间件将管理员写操作发布 NATS（`admin.operation` 主题），user 服务订阅落库，admin 查询接口读库返回
> - **IM 客服链路**：Chat Service（`chat.service`）提供会话/消息/客服接待；消息落库后发布 NATS（`chat.message`），interface/admin 网关订阅并按接收方推送至 WebSocket 连接（用户端 `/ws/chat/v2`、客服端 `/admin/v1/chat/ws`）；会话状态经 `chat.conversation` 主题实时广播（客服工作台待接入刷新）；新会话默认 AI 兜底接待（调用 Blades），客服接入后切换人工；在线状态/未读数存 Redis，客服发消息时用户离线则通过消息中心发站内信提醒（可跳转会话）
> - 所有服务通过 **Consul** 做服务发现与配置管理

## 技术栈

| 分类 | 技术 |
|------|------|
| 语言 | Go 1.25 |
| 框架 | Kratos v2 |
| 依赖注入 | Google Wire |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7.0 |
| 对象存储 | RustFS |
| 消息队列 | NATS |
| 任务队列 | Asynq（基于 Redis） |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |
| 日志 | Zerolog + ELK（Elasticsearch + Logstash + Kibana + Filebeat） |
| AI | DeepSeek API |
| API 协议 | gRPC + HTTP + WebSocket + SSE |
| 配置管理 | Consul KV |

## 服务清单

### 业务服务

| 服务 | 目录 | 说明 | 文档 |
|------|------|------|------|
| CIN Interface | `app/cin/interface` | 接口网关，统一入口 | [README](app/cin/interface/README.md) |
| Admin BFF | `app/cin/admin` | 管理后台 BFF | [README](app/cin/admin/README.md) |
| Order Service | `app/service/order` | 订单管理（状态机） | [README](app/service/order/README.md) |
| Product Service | `app/service/product` | 商品 & 库存管理 | [README](app/service/product/README.md) |
| User Service | `app/service/user` | 用户认证 & 管理 | [README](app/service/user/README.md) |
| Marketing Service | `app/service/marketing` | 优惠券营销管理 | [README](app/service/marketing/README.md) |
| Chat Service | `app/service/chat` | IM 客服（会话/消息/AI 兜底/离线提醒，含用户端 H5 聊天页与客服工作台） | [README](app/service/chat/README.md) |
| Blades Service | `app/blades` | AI 智能助手 | [README](app/blades/README.md) |

### 任务 & 消息服务

| 服务 | 目录 | 说明 | 文档 |
|------|------|------|------|
| Task Service | `app/task` | 异步任务 + 定时任务（Asynq Server + Scheduler，消费 default/export/critical 队列，含订单导出与 cron 定时任务） | [README](app/task/README.md) |
| NATS Consumer | `app/nats` | NATS 消息消费者 | [README](app/nats/README.md) |

## 基础设施

通过 Docker Compose 一键部署，配置文件位于 `deploy/docker-compose/`：

| 基础设施 | 端口 | 说明 |
|----------|------|------|
| Consul | 8500 | 服务发现 & 配置管理 |
| MySQL | 3306 | 主数据库（含 DTM 事务库 `dtm`、屏障库 `dtm_barrier`） |
| Redis | 6379 | 缓存 & 任务队列 |
| NATS | 4222 / 8222 | 消息队列 / 监控 |
| DTM | 36789 | 分布式事务管理器（TCC/SAGA/二阶段消息），管理台 `/admin` |
| Jaeger | 16686 | 分布式追踪 UI |
| Prometheus | 9090 | 监控数据采集 |
| Grafana | 3000 | 监控可视化 |
| Elasticsearch | 9200 | 日志存储 |
| Kibana | 5601 | 日志可视化 |
| RustFS | 9000 / 9001 | 对象存储 |

## 目录结构

```
cyi-app/
├── api/                    # Protobuf API 定义
│   ├── blades/v1/          # AI 助手 API
│   ├── cin/
│   │   ├── admin/v1/       # 管理后台 API（含操作日志 logs、消息中心 message）
│   │   └── interface/v1/   # 接口网关 API（含消息中心 message）
│   ├── common/             # 公共错误码
│   ├── service/
│   │   ├── chat/v1/        # IM 客服 API
│   │   ├── marketing/v1/   # 营销（优惠券）API
│   │   ├── order/v1/       # 订单 API
│   │   ├── product/v1/     # 商品 & 库存 API
│   │   └── user/v1/        # 用户 API（含操作日志 operation_log、消息中心 message）
│   └── task/v1/            # 任务 API
├── app/                    # 服务实现
│   ├── blades/             # AI 助手服务
│   ├── cin/
│   │   ├── admin/           # 管理后台 BFF
│   │   └── interface/      # 接口网关
│   ├── nats/               # NATS 消费者
│   ├── service/
│   │   ├── chat/           # IM 客服服务（会话/消息）
│   │   ├── marketing/      # 优惠券营销服务
│   │   ├── order/          # 订单服务
│   │   ├── product/        # 商品服务
│   │   └── user/           # 用户服务
│   └── task/               # 异步任务服务（含定时任务）
├── pkg/                    # 公共包
│   ├── blades/             # AI 助手封装（Agent、Session、SSE）
│   │   └── client/          # blades 服务 gRPC 客户端封装（admin/BFF 复用）
│   ├── dtm/                # DTM 分布式事务封装（TCC 分支回调注册、屏障事务、配置）
│   ├── exporttask/         # 订单导出异步任务共享类型（admin 投递端 / task 消费端共用）
│   ├── kratos/             # Kratos 扩展
│   │   ├── client/         # gRPC 客户端 & 服务名常量
│   │   ├── config/         # Consul 配置加载
│   │   ├── event/          # NATS 事件定义（操作日志主题与结构）
│   │   ├── log/            # Zerolog 日志适配器
│   │   ├── server/         # 中间件 & 传输器（含操作日志中间件）
│   │   │   └── httproute/  # 手动路由注册工具（SSE 流式聊天、multipart 上传、JWT 中间件）
│   │   ├── service/        # 上下文 & 元数据
│   │   └── tracing/        # Jaeger 追踪
│   ├── upload/             # 文件上传安全校验（扩展名白名单、目录参数校验）
│   └── util/               # 工具包
│       ├── convert/        # 类型转换
│       ├── gorm/           # GORM 模型基类
│       ├── redis/          # Redis 工具
│       └── rustfs/         # RustFS 对象存储
├── deploy/                 # 部署配置
│   └── docker-compose/     # Docker Compose 编排
│       ├── docker-compose.yml          # 主配置（基础设施 + 应用）
│       ├── docker-compose.infra.yml    # 纯基础设施
│       ├── docker-compose.app.yml      # 纯应用服务
│       ├── prometheus/                 # Prometheus 配置
│       ├── logstash/                   # Logstash 配置
│       └── .env                        # 环境变量
├── scripts/                # 本地开发辅助脚本
│   ├── dev-setup.ps1       # 一键准备环境（Windows PowerShell）
│   └── dev-setup.sh        # 一键准备环境（macOS / Linux）
└── .trae/
    └── rules/              # Trae AI 规则
        ├── git-commit-message.md  # Git 提交规范
        └── docs-sync.md          # 文档同步规则
```

## 快速开始

### 环境要求

| 依赖 | 说明 |
|------|------|
| Docker Desktop / Docker Engine | 必须，用于运行 MySQL、Redis、Consul、NATS 等基础设施 |
| Go 1.25+ | 本地运行服务时需要 |
| Make（可选） | 有则可用 `make setup`，没有也能直接用脚本 |
| protoc（可选） | 仅在修改 proto 后需要重新生成代码 |

> Consul 由 Docker Compose 提供，**不需要在本机单独安装**。控制台地址：http://localhost:8500/ui

### 方式一：一键准备（推荐）

Windows PowerShell：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\dev-setup.ps1
```

macOS / Linux：

```bash
bash ./scripts/dev-setup.sh
```

装了 Make 也可以直接：

```bash
make setup
```

脚本会依次完成：

1. 检查 Docker，自动从 `.env.example` 生成 `.env`
2. 启动基础设施并等待 Consul 就绪
3. 把各服务 `configs/*.yaml` 转成本机可访问的地址和端口，写入 Consul KV
4. 输出下一步提示

重复执行是安全的，配置会覆盖更新。只想同步配置、不重启基础设施：

```bash
# Windows
powershell -ExecutionPolicy Bypass -File .\scripts\dev-setup.ps1 -SkipInfra

# macOS / Linux
bash ./scripts/dev-setup.sh --skip-infra

# 或使用 Make
make consul-sync
```

### 方式二：VS Code 启动服务

1. 先执行上面的准备脚本
2. 打开「运行和调试」面板
3. 顶部下拉框选择 `user_service`、`interface`、`admin` 等配置，按 F5
4. 需要全部启动时选择 `⚙️ 全部后端服务` 或 `🚀 全部服务（前端 + 后端）`

各服务通过 `CONSUL_ADDR` / `CONSUL_PATHS` 环境变量读取 Consul 配置，本地端口与 `deploy/docker-compose/.env` 保持一致：

| 启动配置 | Consul 配置路径 | 本地端口 |
|----------|-----------------|----------|
| interface | `/local/cyi-app/cin/interface` | HTTP 8010 / gRPC 9010 |
| admin | `/local/cyi-app/cin/admin` | HTTP 8080 / gRPC 9091 |
| user_service | `/local/cyi-app/service/user` | gRPC 9100 |
| order_service | `/local/cyi-app/service/order` | gRPC 9101 / HTTP 9104 |
| product_service | `/local/cyi-app/service/product` | gRPC 9102 |
| marketing_service | `/local/cyi-app/service/marketing` | gRPC 9103 |
| chat_service | `/local/cyi-app/service/chat` | gRPC 9105 |
| blades | `/local/cyi-app/blades` | HTTP 30800 / gRPC 30900 |
| task | `/local/cyi-app/task` | HTTP 10100 |
| nats | `/local/cyi-app/nats` | HTTP 38000 / gRPC 39000 |

### 方式三：命令行启动

```powershell
# 在项目根目录执行，环境变量按需调整
$env:CONSUL_ADDR="127.0.0.1:8500"
$env:CONSUL_PATHS="/local/common,/local/cyi-app/service/user"
$env:SERVICE_ENV="develop"
go run ./app/service/user/cmd/user
```

### 修改 Proto 后生成代码

```bash
make init    # 首次安装代码生成工具
make api     # 生成 API 代码
make errors  # 生成错误码（在服务目录下执行）
```

### 常见问题

- **服务启动报 Consul 连接失败**：先确认 `docker compose -f deploy/docker-compose/docker-compose.infra.yml ps` 中 `consul` 为 running，再执行准备脚本。
- **服务报配置为空或字段缺失**：重新执行 `make consul-sync`，或打开 http://localhost:8500/ui 检查 `local/cyi-app/<服务>/config.yaml` 是否存在。
- **端口被占用**：修改 `deploy/docker-compose/.env` 中对应端口后重新执行准备脚本。
- **生产部署**：不要使用该脚本，容器内通过服务名访问基础设施，配置由 CI 写入 `develop/cyi-app/*`。

## 开发规范

### 代码结构

每个服务遵循 Kratos 标准分层：

```
app/<service>/
├── cmd/<service>/     # 入口 + Wire 依赖注入
├── configs/           # 配置文件
└── internal/
    ├── biz/           # 业务逻辑层（UseCase）
    ├── conf/          # 配置定义（Protobuf）
    ├── data/          # 数据访问层（Repository）
    ├── server/        # 服务器配置（HTTP/gRPC）
    └── service/       # 服务实现层（gRPC/HTTP Handler）
```

### Protobuf 生成

```bash
# API 代码生成（项目根目录）
make api

# 错误码生成（服务目录下）
cd app/service/order
make errors
```

### Git 提交规范

- 提交信息必须为中文
- 必须包含提交类型：`feat`、`fix`、`docs`、`refactor` 等
- 示例：`feat: 添加订单退款功能`

### 文档同步

修改代码时需同步更新对应文档，详见 [文档同步规则](.trae/rules/docs-sync.md)。

## 服务间通信

| 调用方 | 被调用方 | 协议 |
|--------|----------|------|
| CIN Interface | User Service | gRPC |
| CIN Interface | Order Service | gRPC |
| CIN Interface | Product Service | gRPC |
| CIN Interface | Blades Service | gRPC |
| CIN Interface | Chat Service | gRPC（IM 客服会话/消息） |
| CIN Interface | NATS | 订阅 chat.message，推送消息给在线用户连接 |
| Admin BFF | Chat Service | gRPC（IM 客服客服端操作） |
| Admin BFF | NATS | 订阅 chat.message，推送消息给在线客服连接 |
| Admin BFF | User Service | gRPC |
| Admin BFF | Product Service | gRPC |
| Admin BFF | Order Service | gRPC |
| Admin BFF | Marketing Service | gRPC（优惠券管理） |
| Admin BFF | Redis | Asynq（投递导出任务到 export 队列，由 task 服务消费） |
| Admin BFF | NATS | 发布操作日志事件（admin.operation 主题） |
| User Service | NATS | 订阅操作日志事件并落库 |
| CIN Interface | Marketing Service | gRPC（优惠券领取/查询） |
| Chat Service | NATS | 发布聊天消息事件（chat.message 主题，消息落库后） |
| Chat Service | User Service | gRPC（离线提醒调用消息中心 CreateMessage） |
| NATS Consumer | NATS Server | NATS Protocol |
| Task Service | Redis | Asynq（cron 定时 + 消费 default/export/critical 队列） |
| Task Service | Order Service | gRPC（导出拉单、超时订单取消） |
| Task Service | User Service | gRPC（用户统计） |
| Task Service | Marketing Service | gRPC（优惠券批量过期） |

## 配置管理

所有服务配置统一存储在 Consul KV 中，按路径分隔：

```
/develop/common/              # 公共配置
/develop/cyi-app/cin/interface/  # 接口网关配置
/develop/cyi-app/service/user/   # 用户服务配置
/develop/cyi-app/blades/         # AI 助手配置
...
```

服务启动时从 Consul 加载配置，支持环境变量覆盖。