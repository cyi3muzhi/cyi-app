# CYI-APP 微服务系统

基于 Go + Kratos 框架的电商微服务系统，采用 DDD 领域驱动设计，集成 AI 智能助手。

## 架构图

```
                          ┌─────────────┐
                          │   前端 / 客户端  │
                          └──────┬──────┘
                                 │
                          ┌──────▼──────┐
                          │ CIN Interface │  ← 接口网关（HTTP/gRPC/WS/SSE）
                          └──┬───┬───┬───┘
                             │   │   │
              ┌──────────────┘   │   └──────────────┐
              │                  │                  │
       ┌──────▼──────┐   ┌──────▼──────┐   ┌──────▼──────┐
       │ User Service │   │Order Service│   │Product Service│
       │    用户服务    │   │   订单服务   │   │    商品服务    │
       └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
              │                  │                  │
              │            ┌─────┴─────┐            │
              │            │           │            │
       ┌──────▼──────┐ ┌───▼───┐ ┌─────▼──────┐    │
       │    MySQL     │ │ NATS  │ │ RabbitMQ  │    │
       │    Redis     │ │Consumer│ │ Consumer  │    │
       └─────────────┘ └───────┘ └────────────┘    │
                                                    │
                          ┌─────────────────────────┘
                          │
                   ┌──────▼──────┐
                   │Blades Service│  ← AI 智能助手（DeepSeek）
                   └─────────────┘

     ┌──────────────────────────────────────────────────┐
     │                   基础设施层                       │
     │  Consul │ Jaeger │ Prometheus │ Grafana │ ELK    │
     └──────────────────────────────────────────────────┘
```

## 技术栈

| 分类 | 技术 |
|------|------|
| 语言 | Go 1.25 |
| 框架 | Kratos v2 |
| 依赖注入 | Google Wire |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7.0 |
| 对象存储 | RustFS |
| 消息队列 | NATS、RabbitMQ |
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
| Order Service | `app/service/order` | 订单管理（状态机） | [README](app/service/order/README.md) |
| Product Service | `app/service/product` | 商品 & 库存管理 | [README](app/service/product/README.md) |
| User Service | `app/service/user` | 用户认证 & 管理 | [README](app/service/user/README.md) |
| Blades Service | `app/blades` | AI 智能助手 | [README](app/blades/README.md) |

### 任务 & 消息服务

| 服务 | 目录 | 说明 | 文档 |
|------|------|------|------|
| Job Service | `app/job` | 定时任务（Asynq Scheduler） | [README](app/job/README.md) |
| Task Service | `app/task` | 异步任务（Asynq Server） | [README](app/task/README.md) |
| NATS Consumer | `app/nats` | NATS 消息消费者 | [README](app/nats/README.md) |
| RabbitMQ Consumer | `app/rabbit` | RabbitMQ 消息消费者 | [README](app/rabbit/README.md) |

## 基础设施

通过 Docker Compose 一键部署，配置文件位于 `deploy/docker-compose/`：

| 基础设施 | 端口 | 说明 |
|----------|------|------|
| Consul | 8500 | 服务发现 & 配置管理 |
| MySQL | 3306 | 主数据库 |
| Redis | 6379 | 缓存 & 任务队列 |
| NATS | 4222 / 8222 | 消息队列 / 监控 |
| RabbitMQ | 5672 / 15672 | 消息队列 / 管理界面 |
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
│   ├── cin/interface/v1/   # 接口网关 API
│   ├── service/order/v1/   # 订单 API
│   ├── service/product/v1/ # 商品 & 库存 API
│   ├── service/user/v1/    # 用户 API
│   ├── job/v1/             # 任务 API
│   └── task/v1/            # 任务 API
├── app/                    # 服务实现
│   ├── blades/             # AI 助手服务
│   ├── cin/interface/      # 接口网关
│   ├── job/                # 定时任务服务
│   ├── nats/               # NATS 消费者
│   ├── rabbit/             # RabbitMQ 消费者
│   ├── service/
│   │   ├── order/          # 订单服务
│   │   ├── product/        # 商品服务
│   │   └── user/           # 用户服务
│   └── task/               # 异步任务服务
├── pkg/                    # 公共包
│   ├── blades/             # AI 助手封装（Agent、Session、SSE）
│   ├── kratos/             # Kratos 扩展
│   │   ├── client/         # gRPC 客户端 & 服务名常量
│   │   ├── config/         # Consul 配置加载
│   │   ├── log/            # Zerolog 日志适配器
│   │   ├── server/         # 中间件 & 传输器
│   │   ├── service/        # 上下文 & 元数据
│   │   └── tracing/        # Jaeger 追踪
│   ├── mq/                 # 消息队列抽象
│   ├── rmq/                # RabbitMQ 封装
│   └── util/               # 工具包
│       ├── asynq_helper/   # Asynq 工具
│       ├── convert/        # 类型转换
│       ├── gorm/           # GORM 模型基类
│       ├── poolconfig/     # 连接池配置
│       ├── redis/          # Redis 工具
│       ├── rustfs/         # RustFS 对象存储
│       └── toon/           # Toon 工具
├── deploy/                 # 部署配置
│   └── docker-compose/     # Docker Compose 编排
│       ├── docker-compose.yml          # 主配置（基础设施 + 应用）
│       ├── docker-compose.infra.yml    # 纯基础设施
│       ├── docker-compose.app.yml      # 纯应用服务
│       ├── prometheus/                 # Prometheus 配置
│       ├── logstash/                   # Logstash 配置
│       └── .env                        # 环境变量
└── .trae/
    └── rules/              # Trae AI 规则
        ├── git-commit-message.md  # Git 提交规范
        └── docs-sync.md          # 文档同步规则
```

## 快速开始

### 环境要求

- Go 1.25+
- Docker & Docker Compose
- Make（可选）

### 1. 启动基础设施

```bash
cd deploy/docker-compose
docker compose up -d
```

### 2. 生成代码

```bash
# 在项目根目录执行
make api    # 生成 Protobuf 代码
```

### 3. 启动单个服务（本地开发）

```bash
cd app/cin/interface
go run cmd/interface/main.go -conf configs/
```

### 4. Docker 部署全部服务

```bash
cd deploy/docker-compose
docker compose up -d
```

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
| NATS Consumer | NATS Server | NATS Protocol |
| RabbitMQ Consumer | RabbitMQ Server | AMQP |
| Job/Task Service | Redis | Asynq |

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