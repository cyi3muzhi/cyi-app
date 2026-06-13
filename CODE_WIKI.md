# cyi-app Code Wiki

> 基于 Go + Kratos 的微服务架构项目。本 Wiki 全面介绍项目的架构设计、模块职责、关键代码、依赖关系和运行方式。

---

## 📑 目录

1. [项目概览](#1-项目概览)
2. [系统架构](#2-系统架构)
3. [技术栈](#3-技术栈)
4. [目录结构](#4-目录结构)
5. [API 协议定义](#5-api-协议定义)
6. [服务模块详解](#6-服务模块详解)
7. [公共工具包 pkg](#7-公共工具包-pkg)
8. [配置与启动流程](#8-配置与启动流程)
9. [部署与基础设施](#9-部署与基础设施)
10. [运行方式](#10-运行方式)
11. [Makefile 与构建工具](#11-makefile-与构建工具)
12. [CI/CD 流水线](#12-cicd-流水线)
13. [架构设计亮点与注意事项](#13-架构设计亮点与注意事项)

---

## 1. 项目概览

**cyi-app** 是一个采用**微服务架构**的后端项目，同时包含一个简单的商城系统（用户+订单）和一个 AI 聊天服务（Blades）。项目基于 Kratos 框架，采用 gRPC 进行服务间通信，并提供 HTTP RESTful 接口对外服务。异步任务通过 Asynq (Redis) 实现。

### 1.1 项目功能模块

| 模块 | 功能描述 |
|-----|---------|
| 🔐 用户认证 | 注册/登录/JWT Token 鉴权 |
| 👤 用户管理 | 用户信息查询/更新/统计 |
| 📦 商品管理 | 商品 CRUD（脚手架阶段） |
| 🛒 订单管理 | 创建订单/支付订单/状态流转 |
| 🤖 Blades AI 聊天 | 接入 DeepSeek 大模型 |
| ⏰ 定时任务 | 基于 Asynq 的定时 Job |
| 📋 异步任务 | 基于 Asynq 的异步 Task |

### 1.2 工程根目录文件

| 文件 | 说明 |
|-----|------|
| `go.mod` | Go 模块依赖（Go 1.25） |
| `Makefile` | 项目级构建脚本（Proto 生成、Docker 构建等） |
| `openapi.yaml` | OpenAPI 3.0 规范（自动生成） |
| `LICENSE` | 许可证 |
| `.github/workflows/` | GitHub Actions CI/CD 配置 |

---

## 2. 系统架构

### 2.1 架构图

```
                     ┌─────────────────────────────┐
                     │   External Clients (Web/App)│
                     └──────────────┬──────────────┘
                                    │ HTTPS
                                    ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                                 Gateway Layer                                  │
│                                                                                │
│   ┌──────────────────────────┐    ┌───────────────────────────┐              │
│   │  cin/interface (BFF)     │    │  blades (AI Chat)         │              │
│   │  HTTP: user/order 接口   │    │  HTTP: chat 接口          │              │
│   │  gRPC: 调用内部服务      │    │  gRPC: 可选                │              │
│   └──────────┬───────────────┘    └────────────┬──────────────┘              │
└──────────────┼─────────────────────────────────┼─────────────────────────────┘
               │ gRPC (服务发现: Consul)         │
               ▼                                  ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                          Internal Microservices                                │
│                                                                                │
│   ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌─────────┐  ┌────────┐   │
│   │  user      │  │  order     │  │  product   │  │  task   │  │  job   │   │
│   │  用户服务   │  │  订单服务   │  │  商品服务   │  │ 异步任务│  │定时任务│   │
│   └─────┬──────┘  └─────┬──────┘  └────────────┘  └────┬────┘  └────────┘   │
└─────────┼──────────────┼──────────────────────────────┼──────────────────────┘
          │              │                              │
          ▼              ▼                              ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                            Infrastructure Layer                                │
│                                                                                │
│   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐     │
│   │  MySQL   │  │  Redis   │  │  Consul  │  │  Jaeger  │  │  Elastic  │     │
│   │ (数据库)   │  │ (缓存/队列)│  │ (服务发现) │  │ (链路追踪) │  │ (日志)    │     │
│   └──────────┘  └──────────┘  └──────────┘  └──────────┘  └───────────┘     │
└────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 架构分层

- **网关层 (BFF / Gateway)**：`cin/interface`、`blades` — 对外暴露 HTTP 接口，对内通过 gRPC 调用微服务
- **业务服务层**：`user`、`order`、`product` — 核心业务逻辑服务
- **任务调度层**：`task`（异步任务处理器）、`job`（定时任务调度器）— 基于 Asynq + Redis
- **基础设施层**：MySQL、Redis、Consul、Jaeger、ELK、Prometheus/Grafana

### 2.3 核心设计模式

| 模式 | 应用场景 |
|-----|---------|
| **Repository 模式** | 数据访问层通过接口抽象，biz 层不直接依赖 ORM |
| **State 状态模式** | 订单状态流转 `OrderStatus` 接口 + 各状态实现类 |
| **Dependency Injection** | Google Wire 生成依赖注入代码 |
| **Clean Architecture** | Kratos 标准分层：`data → biz → service → server` |
| **Circuit Breaker / Singleflight** | 网关层使用 singleflight 防缓存击穿 + Redis 分布式锁 |

---

## 3. 技术栈

### 3.1 核心框架与库

| 分类 | 技术 | 版本 | 用途 |
|-----|------|------|------|
| 语言 | Go | 1.25 | 开发语言 |
| 框架 | Kratos | v2 | 微服务框架（HTTP/gRPC） |
| RPC | gRPC | v1.77 | 服务间通信 |
| API | Protocol Buffers | v3 | 接口定义 |
| ORM | GORM | v1.31 | 数据库操作 |
| ORM 代码生成 | GORM Gen | v0.3.6 | 类型安全的查询 |
| 缓存 | go-redis | v9.14 | Redis 客户端 |
| 消息队列 | streadway/amqp | v1.1 | RabbitMQ 客户端 |
| 任务调度 | Asynq | v0.26 | 基于 Redis 的异步任务 |
| 鉴权 | JWT (golang-jwt) | v5 | JSON Web Token |
| 密码加密 | x/crypto/bcrypt | - | bcrypt 哈希 |
| 服务发现 | Consul | v1.32 | 服务注册与发现 |
| 链路追踪 | OpenTelemetry | v1.40 | 分布式追踪 |
| 日志 | zerolog | v1.34 | 高性能日志库 |
| 配置 | Kratos Config + Consul | - | 分布式配置中心 |
| DI | Google Wire | v0.7 | 依赖注入代码生成 |
| AI 代理 | go-kratos/blades | v0.4 | AI 聊天代理框架 |
| 对象存储 | rustfs-go | v1.0.6 | Rust 实现的对象存储 |

### 3.2 基础设施

| 组件 | 版本/类型 | 用途 |
|-----|----------|------|
| **MySQL 8.0+** | 关系型数据库 | 用户/订单/商品数据 |
| **Redis 7.0+** | 内存数据库 | 缓存、分布式锁、Asynq 任务队列 |
| **Consul** | 服务网格 | 服务注册发现、配置中心 |
| **Jaeger** | v1 | 分布式链路追踪 |
| **Elasticsearch + Logstash + Kibana** | ELK 栈 | 日志收集与分析 |
| **Prometheus + Grafana** | v2.x | 指标监控与可视化 |
| **Docker** | 20+ | 容器化 |
| **Kubernetes** | v1.28+ | 生产级容器编排 |

---

## 4. 目录结构

```
cyi-app/
├── api/                            # Proto API 定义
│   ├── blades/v1/                  # Blades AI 聊天服务
│   ├── cin/                        # 网关层
│   │   ├── admin/user/             # 管理员用户接口（预留）
│   │   └── interface/v1/           # 对外接口（用户/订单）
│   ├── job/v1/                     # 定时任务枚举
│   ├── service/                    # 内部业务服务
│   │   ├── order/v1/               # 订单服务
│   │   ├── product/v1/             # 商品服务
│   │   └── user/v1/                # 用户服务
│   └── task/v1/                    # 异步任务定义
├── app/                            # 服务实现
│   ├── blades/                     # Blades AI 服务
│   ├── cin/interface/              # BFF 网关接口
│   ├── job/                        # 定时任务服务（Asynq Scheduler）
│   ├── rmq_consumers/              # RabbitMQ 消费者（预留）
│   ├── service/                    # 业务微服务
│   │   ├── order/                  # 订单服务
│   │   ├── product/                # 商品服务
│   │   └── user/                   # 用户服务
│   └── task/                       # 异步任务服务（Asynq Server）
├── deploy/                         # 部署配置
│   ├── build/                      # Dockerfile
│   ├── docker-compose/             # Docker Compose
│   └── k8s/                        # Kubernetes 配置
├── pkg/                            # 公共工具包
│   ├── blades/                     # Blades AI 代理封装
│   ├── kratos/                     # Kratos 框架扩展
│   │   ├── client/                 # gRPC 客户端
│   │   ├── config/                 # Consul 配置
│   │   ├── log/                    # zerolog 日志适配
│   │   ├── server/                 # 服务端中间件
│   │   ├── service/                # 上下文工具（获取用户 ID）
│   │   └── tracing/                # Jaeger 链路追踪
│   ├── mq/                         # RabbitMQ 封装
│   ├── rmq/                        # RabbitMQ 工厂模式封装
│   └── util/                       # 通用工具
│       ├── asynq_helper/           # Asynq 辅助
│       ├── convert/                # 结构体转换
│       ├── gorm/                   # GORM 通用模型
│       ├── poolconfig/             # 池配置
│       ├── redis/                  # Redis 辅助（锁/缓存）
│       ├── rustfs/                 # 对象存储封装
│       └── toon/                   # toon 工具
├── third_party/                    # 第三方 Proto 文件
│   ├── errors/
│   ├── google/api/
│   ├── openapi/v3/
│   └── validate/
├── test/                           # 测试与证书目录
├── .github/workflows/              # GitHub Actions
├── Makefile                        # 根 Makefile
├── go.mod                          # Go 模块依赖
└── openapi.yaml                    # OpenAPI 规范（自动生成）
```

### 4.1 每个服务的标准目录结构

每个 Kratos 服务（如 `app/service/user/`）遵循以下标准结构：

```
<service-name>/
├── cmd/
│   └── <service-name>/
│       ├── main.go         # 程序入口
│       ├── wire.go         # Wire 依赖注入声明
│       └── wire_gen.go     # Wire 自动生成代码
├── configs/
│   ├── auth.yaml           # 认证配置
│   ├── config.yaml         # 主配置
│   ├── register.yaml       # 服务注册配置
│   └── <service-specific>.yaml  # 服务特有配置
├── internal/
│   ├── biz/                # 业务逻辑层（Use Case）
│   │   ├── biz.go          # ProviderSet 聚合
│   │   └── <entity>.go     # 具体业务
│   ├── conf/               # 配置结构体（Proto 定义 + Go 代码）
│   │   ├── conf.proto
│   │   └── conf.pb.go
│   ├── data/               # 数据访问层（Repository 实现）
│   │   ├── data.go         # ProviderSet + 数据源初始化
│   │   ├── <entity>.go     # 具体 Repository
│   │   ├── curd/           # GORM Gen 生成代码
│   │   └── models/         # GORM 数据模型
│   ├── pkg/                # 服务内私有工具包（可选）
│   │   └── encryption/     # 如密码加密
│   ├── server/             # 服务端（HTTP/gRPC）
│   │   ├── grpc.go
│   │   ├── http.go
│   │   └── server.go       # ProviderSet
│   └── service/            # 服务实现（Protobuf Service 实现）
│       ├── service.go      # ProviderSet
│       └── <entity>.go     # 具体 Service
├── Makefile                # 服务级 Makefile
└── openapi.yaml            # 服务 OpenAPI（自动生成）
```

---

## 5. API 协议定义

所有服务接口通过 Protocol Buffers (Protobuf v3) 定义，支持 `google.api.http` 注解映射为 RESTful HTTP 接口，使用 `validate.proto` 做参数校验，使用 `errors.proto` 定义错误码。

### 5.1 对外 HTTP 接口总览

| Method | Path | Service | RPC | 描述 |
|--------|------|---------|-----|------|
| POST | `/user/v1/login` | CinInterface | `Login` | 用户登录 |
| POST | `/user/v1/register` | CinInterface | `Register` | 用户注册 |
| GET | `/user/v1/{id}` | CinInterface | `GetUserInfo` | 获取指定用户信息 |
| GET | `/user/v1/my_info` | CinInterface | `GetMyInfo` | 获取当前登录用户信息 |
| PUT | `/user/v1/update` | CinInterface | `UpdateUserInfo` | 更新当前用户信息 |
| POST | `/order/v1/create` | CinInterface | `CreateOrder` | 创建订单 |
| POST | `/order/v1/pay` | CinInterface | `PayOrder` | 支付订单 |
| POST | `/blades/v1/chat` | Blades | `Chat` | AI 聊天 |

### 5.2 内部 gRPC 服务

| 服务 | 主要 RPC | 说明 |
|-----|----------|------|
| **User** | `CreateUser / UpdateUser / DeleteUser / GetUser / ListUser / CheckUserExist / GetUserByUsername / VerifyPassword / UserCount / Register / Login` | 用户管理服务 |
| **Order** | `CreateOrder / UpdateOrder / DeleteOrder / GetOrder / ListOrder / PayOrder` | 订单服务（含状态流转） |
| **Product** | `CreateProduct / UpdateProduct / DeleteProduct / GetProduct / ListProduct` | 商品服务（脚手架） |

### 5.3 订单状态流转

在 `api/service/order/v1/order.proto` 中定义的枚举 `OrderStatus`：

```
0: Unspecified （未指定）
1: WaitPay     （待支付）
2: Pay         （已支付）
3: WaitShip    （待发货）
4: Ship        （已发货）
5: WaitReceive （待收货）
6: Complete    （已完成）
7: Cancel      （已取消）
8: Refunding   （退款中）
9: Refunded    （已退款）
```

### 5.4 任务类型枚举

- **`api/job/v1/job.proto` → `JobName`**：`user_count = 0`（定时统计用户数）
- **`api/task/v1/task.proto` → `TaskName`**：`create_user = 0`（异步创建用户），含 `CreateUserTask{username, password}` 消息

---

## 6. 服务模块详解

### 6.1 用户服务 (`app/service/user/`)

**职责**：管理用户注册、登录、信息更新、密码验证、用户统计。

#### 核心文件

| 文件 | 作用 |
|-----|------|
| [biz/user.go](file:///e:/me/cyi-app/app/service/user/internal/biz/user.go) | 用户业务逻辑、`UserRepo` 接口定义、`UserUseCase` |
| [biz/auth.go](file:///e:/me/cyi-app/app/service/user/internal/biz/auth.go) | 注册/登录逻辑、JWT Token 生成 |
| [data/user.go](file:///e:/me/cyi-app/app/service/user/internal/data/user.go) | `UserRepo` 实现，含 Redis 缓存策略 |
| [data/models/user.go](file:///e:/me/cyi-app/app/service/user/internal/data/models/user.go) | GORM 数据模型 `User{Name, Phone, Password, Username}` |
| [service/user.go](file:///e:/me/cyi-app/app/service/user/internal/service/user.go) | User gRPC Service 实现 |
| [pkg/encryption/encryption.go](file:///e:/me/cyi-app/app/service/user/internal/pkg/encryption/encryption.go) | bcrypt 密码加密与验证 |

#### 核心结构体

```go
// biz/user.go - 业务层用户对象
type User struct {
    ID       int64
    Name     string
    Phone    string
    Password string
    Username string
}

// biz/user.go - 仓库接口（数据访问抽象）
type UserRepo interface {
    CreateUser(ctx context.Context, u *User) (*User, error)
    UpdateUser(ctx context.Context, u *User) error
    GetUser(ctx context.Context, u *User) (*User, error)
    FindByUsername(ctx context.Context, username string) (*User, error)
    VerifyPassword(ctx context.Context, u *User) error
    UserCount(ctx context.Context) (int64, error)
}

// biz/user.go - 业务用例
type UserUseCase struct {
    repo UserRepo
    log  *log.Helper
}
```

#### 核心业务流程

**① 用户注册流程**：

```
Register(username, password)
    ├── 检查 username 是否已存在 → FindByUsername
    │   └── 已存在 → 返回 ErrorUserAlreadyRegistered
    ├── 密码 bcrypt 加密 → EncryptPassword(cost=12)
    └── 创建用户 → CreateUser
```

**② 用户登录流程**：

```
Login(username, password, apiKey)
    ├── 通过 username 查询用户 → FindByUsername
    ├── 验证密码正确性 → VerifyPassword（bcrypt CompareHashAndPassword）
    └── 生成 JWT Token:
          ├── Claims: UserClaims{ID, Username, RegisteredClaims}
          ├── Method: HS256
          ├── Expires: 24小时
          ├── Issuer:  "cin"
          └── Token ID: uuid
```

**③ 用户缓存策略**（`data/user.go`）：

```
查找用户:
    ├─ 首次查询: MySQL → 写入 Redis (TTL=24h)
    └─ 重复查询: 优先命中 Redis，未命中再查数据库
缓存键:
    "user_cache_key_" + username   （按用户名）
    "user_cache_key_" + id         （按ID）
```

#### JWT Claim 结构

```go
// pkg/kratos/server/middleware/auth.go
type UserClaims struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    jwt2.RegisteredClaims  // 标准字段: ExpiresAt/Issuer/...
}
```

---

### 6.2 订单服务 (`app/service/order/`)

**职责**：管理订单创建、支付、状态流转（核心：**状态模式**实现订单生命周期）。

#### 核心文件

| 文件 | 作用 |
|-----|------|
| [biz/order.go](file:///e:/me/cyi-app/app/service/order/internal/biz/order.go) | 订单用例 `OrderUseCase`、`OrderRepo` 接口 |
| [biz/order_status.go](file:///e:/me/cyi-app/app/service/order/internal/biz/order_status.go) | **状态模式核心**：`OrderStatus` 接口 + `NewOrderStatusHandler` 工厂 |
| [data/order.go](file:///e:/me/cyi-app/app/service/order/internal/data/order.go) | 订单数据访问实现 |
| [data/models/order.go](file:///e:/me/cyi-app/app/service/order/internal/data/models/order.go) | GORM 模型 `Order{UserID, ProductID, Count, Status}` |
| [service/order.go](file:///e:/me/cyi-app/app/service/order/internal/service/order.go) | Order gRPC Service 实现 |

#### 核心架构：状态模式

```go
// biz/order_status.go - 状态接口
type OrderStatus interface {
    Pay(ctx context.Context, in *Order) error
    Ship(ctx context.Context) error
    Receive(ctx context.Context) error
    Complete(ctx context.Context) error
    Cancel(ctx context.Context) error
    Refunding(ctx context.Context) error
    Refunded(ctx context.Context) error
}

// 订单对象持有当前状态处理器
type Order struct {
    ID        int64
    UserID    int64
    ProductID int64
    Count     int32
    Status    v1.OrderStatus
    statusHandler OrderStatus  // ← 当前状态处理器
}

// 状态切换由工厂统一管理
func NewOrderStatusHandler(order OrderInterface, repo OrderRepo, logger log.Logger) OrderStatus {
    switch order.GetStatus() {
    case v1.OrderStatus_WaitPay:   return NewWaitPayStatus(...)
    case v1.OrderStatus_Pay:       return NewPayStatus(...)
    case v1.OrderStatus_WaitShip:  return NewPayStatus(...)    // 复用
    case v1.OrderStatus_Ship:      return NewShipStatus(...)
    case v1.OrderStatus_WaitReceive: return NewShipStatus(...) // 复用
    case v1.OrderStatus_Complete:  return NewCompleteStatus(...)
    case v1.OrderStatus_Cancel:    return NewCancelStatus(...)
    case v1.OrderStatus_Refunding: return NewRefundingStatus(...)
    case v1.OrderStatus_Refunded:  return NewRefundedStatus(...)
    default: return NewWaitPayStatus(...)
    }
}
```

#### 支付流程

```
PayOrder(orderID):
    ├── 从 context 提取当前登录 userID → GetUserID(ctx)
    ├── 查询订单 → GetOrder(orderID)
    ├── 校验权限: 订单.UserID == userID
    ├── 初始化状态处理器 → SetStatus(order.Status)
    └── 调用状态处理器.Pay(ctx, order)
            └── [WaitPay 状态下允许支付, 其他状态拒绝]
```

---

### 6.3 网关接口服务 (`app/cin/interface/`)

**职责**：BFF（Backend for Frontend）层，对外暴露 HTTP 接口，对内通过 gRPC 调用 `user` 和 `order` 服务。

#### 核心文件

| 文件 | 作用 |
|-----|------|
| [biz/user.go](file:///e:/me/cyi-app/app/cin/interface/internal/biz/user.go) | 网关层用户用例 |
| [biz/order.go](file:///e:/me/cyi-app/app/cin/interface/internal/biz/order.go) | 网关层订单用例 |
| [data/user.go](file:///e:/me/cyi-app/app/cin/interface/internal/data/user.go) | 调用 user gRPC 服务（含 singleflight 防击穿 + Redis 分布式锁） |
| [data/order.go](file:///e:/me/cyi-app/app/cin/interface/internal/data/order.go) | 调用 order gRPC 服务（同样的防击穿策略） |
| [service/user.go](file:///e:/me/cyi-app/app/cin/interface/internal/service/user.go) | CinInterface User 相关 HTTP RPC 实现 |
| [service/order.go](file:///e:/me/cyi-app/app/cin/interface/internal/service/order.go) | CinInterface Order 相关 HTTP RPC 实现 |

#### 防缓存击穿设计

```go
// data/user.go - 使用 singleflight + Redis 分布式锁
func (r *userRepo) FindByUsername(ctx, username) (*biz.User, error) {
    // 1. 本地 singleflight: 合并同一 key 的并发请求
    result, err, _ := r.sg.Do("find_user_by_name_"+username, func() {
        // 2. 分布式锁: 10s 过期，防止多实例并发请求数据库
        lockKey := "lock:user:username:" + username
        if !cr.AcquireDistributedLock(ctx, r.data.rc, lockKey, 10*time.Second) {
            return nil, fmt.Errorf("lock failed")
        }
        defer cr.ReleaseDistributedLock(ctx, r.data.rc, lockKey)

        // 3. 真正调用内部 gRPC 服务
        return r.data.uc.GetUserByUsername(ctx, &usV1.GetUserByUsernameReq{...})
    })
    return result, err
}
```

#### 通过 context 传递用户身份

```go
// pkg/kratos/service/ctx.go
func GetUserID(ctx context.Context) (int64, error) {
    claims, ok := jwt.FromContext(ctx)   // 从 JWT 中间件注入的 claims
    userClaims, ok := claims.(*server.UserClaims)
    return userClaims.ID, nil
}
```

---

### 6.4 Blades AI 聊天服务 (`app/blades/`)

**职责**：接入 DeepSeek 大模型 API，提供 AI 聊天接口。

#### 核心流程

```
Chat(message):
    ├── 创建 Blades Agent（使用配置中的 api_key）
    ├── 记录请求（可选持久化）
    └── 调用 RunAgent(ctx, agent, message) → 返回 AI 响应
```

#### 配置字段

```yaml
blades:
  api_key:      ${DEEPSEEK_API_KEY}  # 大模型 API Key
  model:        deepseek-chat         # 模型名称
  max_tokens:   2048                  # 最大 token 数
  temperature:  0.7                   # 生成温度
```

---

### 6.5 Task 异步任务服务 (`app/task/`)

**职责**：基于 Asynq Server 处理异步任务（如 `create_user` 任务）。启动 Asynq Server 监听 Redis 任务队列。

### 6.6 Job 定时任务服务 (`app/job/`)

**职责**：基于 Asynq Scheduler 注册定时任务（如 `user_count` 每小时统计用户数）。通过 Cron 表达式配置调度时机。

---

## 7. 公共工具包 `pkg/`

`pkg/` 是整个项目的共享代码库，被所有服务引用。

### 7.1 `pkg/kratos/` - Kratos 框架扩展

| 子包 | 文件 | 功能 |
|-----|------|------|
| **client/** | [client.go](file:///e:/me/cyi-app/pkg/kratos/client/client.go) | 创建 gRPC 客户端连接 `NewClientConn` |
| **config/** | [consul.go](file:///e:/me/cyi-app/pkg/kratos/config/consul.go) | `NewConsulConfig` (从 Consul 读取多路径配置) + `NewConsulRegistry` (服务注册) |
| **log/** | `zerolog.go` | zerolog 日志适配（控制台/文件输出、颜色、格式） |
| **server/middleware/** | `auth.go` | JWT 服务端/客户端中间件 + 白名单匹配器 |
| | `transaction.go` | 【已弃用】分布式事务中间件（本地内存 map 存事务 ID） |
| | `common.go` | 通用中间件 |
| | `middleware.go` | ProviderSet |
| **server/transporter/** | `transporter.go` | 异步任务传输封装 |
| **service/** | [ctx.go](file:///e:/me/cyi-app/pkg/kratos/service/ctx.go) | `GetUserID(ctx)` — 从 JWT claims 提取用户 ID |
| | `metadata.go` | 元数据工具 |
| **tracing/** | `jeager.go` | Jaeger TracerProvider 初始化 |

#### 关键代码片段 - JWT 鉴权中间件

```go
// pkg/kratos/server/middleware/auth.go
func AuthServiceMiddleware(auth Auth) middleware.Middleware {
    return jwt.Server(
        func(token *jwt2.Token) (interface{}, error) {
            return []byte(auth.ApiKey), nil   // 对称密钥
        },
        jwt.WithSigningMethod(jwt2.SigningMethodHS256),
        jwt.WithClaims(func() jwt2.Claims { return &UserClaims{} }),
    )
}
```

### 7.2 `pkg/util/` - 通用工具

| 子包 | 功能 |
|-----|------|
| **redis/** | [cin_redis.go](file:///e:/me/cyi-app/pkg/util/redis/cin_redis.go) | `SetRedis/GetRedis/DelRedis` + 哈希/列表操作 + **分布式锁** (`AcquireDistributedLock/ReleaseDistributedLock` via SETNX) |
| **convert/** | [convert.go](file:///e:/me/cyi-app/pkg/util/convert/convert.go) | `Convert(src, dst)` 使用 mapstructure 进行结构体深拷贝；`ToMap(s)` 结构体转 map |
| **gorm/** | `model.go` | 通用 GORM 模型（含 `ID/CreatedAt/UpdatedAt/DeletedAt` — 软删除） |
| **asynq_helper/** | `asynq.go` | Asynq 客户端封装 |
| **poolconfig/** | `poolconfig.go` | 连接池配置（示例代码） |
| **rustfs/** | `rustfs.go` | Rust 对象存储服务 SDK 封装 |
| **toon/** | `toon.go` | toon 工具库 |

#### Redis 辅助工具核心 API

```go
SetRedis(ctx, client, key, value, expiration) error  // 支持 string/JSON
GetRedis(ctx, client, key, &value) error              // 自动 JSON 反序列化
DelRedis(ctx, client, key) error

// 分布式锁（基于 SETNX）
AcquireDistributedLock(ctx, client, key, expiration) bool
ReleaseDistributedLock(ctx, client, key) error

// 常量: OneDay=24h, OneWeek=7d, OneMonth=30d, OneYear=365d
```

### 7.3 `pkg/mq/` 与 `pkg/rmq/` - RabbitMQ 封装

| 包 | 文件 | 功能 |
|----|------|------|
| **mq/** | `rabbitmq.go` | `RabbitMQ` 结构体：Connect/Reconnect/Close/DeclareQueue/DeclareExchange/BindQueue/Publish/Consume/StartHeartbeat |
| | `consumer.go` | 消费者基类 |
| | `service.go` | 服务级封装 |
| | `handler.go` | 消息处理器接口 |
| | `options.go` | 配置选项 |
| **rmq/** | `factory.go` | 工厂模式封装 |
| | `producer.go` / `consumer.go` | 生产者/消费者 |
| | `connection.go` | 连接管理 |
| | `exchange.go` | 交换机声明 |
| | `constants.go` | 常量定义 |
| | `queue.go` | 队列声明 |
| | `options.go` | 配置选项 |

两套封装并存，`rmq` 为更成熟的工厂模式版本，`mq` 为简化版。

### 7.4 `pkg/blades/` - Blades AI 代理

| 文件 | 功能 |
|-----|------|
| `blades.go` | `CreateAgent(apiKey)` 创建 AI 代理 |
| `demo.go` | 使用示例 |

---

## 8. 配置与启动流程

### 8.1 统一配置结构 (`Bootstrap`)

所有服务共享相同的根配置结构体，定义在各自服务的 `internal/conf/conf.proto` 中：

```protobuf
message Bootstrap {
    Server server = 1;      // HTTP/gRPC 监听与超时
    Data   data   = 2;      // 数据库/Redis + service-specific 字段
    Auth   auth   = 3;      // API Key / 服务密钥 / 白名单
    Trace  trace  = 4;      // Jaeger 端点
    Log    log    = 5;      // 日志级别与输出目标
    // 服务特有字段:
    //  user/order/product: Data.oss 对象存储
    //  blades: Blades 配置（大模型参数）
    //  task: repeated Task tasks（任务列表）
}

message Server {
    message HTTP { string network = 1; string addr = 2; int32 timeout = 3; }
    message GRPC { string network = 1; string addr = 2; int32 timeout = 3; }
    HTTP http = 1;
    GRPC grpc = 2;
}

message Data {
    message Database { string driver = 1; string source = 2; }
    message Redis    { string network = 1; string addr = 2; ... }
    message OSS      { string endpoint = 1; string access_key = 2; ... }
    Database database = 1;
    Redis    redis    = 2;
    OSS      oss      = 3;   // 仅 user/product 服务
}
```

### 8.2 Consul 配置中心集成

服务**不使用本地 yaml 作为主配置**，而是通过环境变量 `CONSUL_ADDR` 和 `CONSUL_PATHS` 从 Consul 动态拉取：

```go
// pkg/kratos/config/consul.go
func NewConsulConfig(opts *ConsulConfigOptions) (config.Config, error) {
    // paths = 多路径读取，从 env CONSUL_PATHS 按逗号分隔
    for _, path := range paths {
        source, _ := consulConfig.New(client, consulConfig.WithPath(path))
        sources = append(sources, source)
    }
    return config.New(config.WithSource(sources...)), nil
}

// 使用: c.Load() → c.Scan(&bc)  // 读取到 Bootstrap 结构体
```

### 8.3 服务启动流程

所有服务 `main.go` 遵循以下标准流程：

```
Step 1: 初始化配置
        └── NewConsulConfig → config.Load → config.Scan(&bc)
              Bootstrap{Server, Data, Auth, Trace, Log, ...}

Step 2: 初始化日志系统
        └── zerolog.NewZerologAdapterWithConfig(bc.Log)
              → 控制台 / 文件 / 颜色 / 级别配置

Step 3: 初始化链路追踪
        └── tracing.NewTracerProvider(serviceName, bc.Trace.Endpoint)
              → OpenTelemetry + Jaeger

Step 4: 依赖注入（Google Wire 自动生成）
        └── wireApp(bc.Server, bc.Data, bc.Auth, ..., logger, tp)
              → 返回 *kratos.App, func(), error
              ProviderSet 链: server → data → biz → service

Step 5: 启动并阻塞
        └── app.Run()
              → 同时启动 HTTP Server 和 gRPC Server
              → 注册到 Consul（服务发现）
              → 监听中断信号优雅退出
```

---

## 9. 部署与基础设施

### 9.1 Docker Compose 部署

**目录**: `deploy/docker-compose/`

| 文件 | 用途 |
|-----|------|
| `docker-compose.yml` | 完整版：基础设施 + 应用服务 |
| `docker-compose.infra.yml` | 仅基础设施（MySQL/Redis/Consul/Jaeger/Rustfs/ELK/Prometheus/Grafana） |
| `docker-compose.app.yml` | 仅应用服务（interface/task/job/user/blades） |
| `.env` | 环境变量 |
| `config.tidb.example.yaml` | TiDB 配置示例（兼容 MySQL 的分布式数据库） |
| `TIDB_*.md` | TiDB 部署指南 |

**关键服务端口**：

| Service | Port |
|---------|------|
| Consul | 8500 (HTTP) |
| MySQL | 3306 |
| Redis | 6379 |
| Jaeger | 16686 (UI) / 4318 (OTLP) |
| Rustfs | 9000 / 9001 (UI) |
| Elasticsearch | 9200 |
| Kibana | 5601 |
| Prometheus | 9090 |
| Grafana | 3000 |
| App-cin-interface | 8000 (HTTP) / 9000 (gRPC) |
| App-service-user | 9000 (gRPC) |

### 9.2 Kubernetes 部署

**目录**: `deploy/k8s/`

```
deploy/k8s/
├── config/                  # ConfigMap & Secret
│   ├── configmap.yaml
│   └── secret.yaml
├── infrastructure/          # 底层服务
│   ├── consul/
│   ├── elk/
│   ├── jaeger/
│   ├── mysql/
│   └── redis/
├── microservices/           # 应用服务
│   ├── blades/              # ✅ 已配置
│   ├── interface/           # ✅ 已配置
│   ├── job/                 # ✅ 已配置
│   ├── task/                # ✅ 已配置
│   └── user/                # ✅ 已配置
├── monitoring/              # 监控
│   ├── prometheus/
│   └── grafana/
├── deploy-all.ps1           # 一键部署脚本 (PowerShell)
└── README.md                # 详细部署说明
```

**部署顺序**：`config → infrastructure → microservices → monitoring`

### 9.3 Dockerfile 多阶段构建

```dockerfile
# deploy/build/Dockerfile
FROM golang:1.25 AS builder
WORKDIR /src
COPY . .
# make build → 产出 bin/<service-name>

FROM debian:stable-slim
RUN apt-get update && apt-get install -y ca-certificates netbase
COPY --from=builder /src/.../bin /app
CMD ./${APP_NAME}
```

---

## 10. 运行方式

### 10.1 本地开发运行

**前置条件**：Go 1.25+、已安装 protoc 相关插件、本地或 Docker 启动 MySQL/Redis/Consul。

```bash
# ① 安装工具（一次性）
make init
# 等价于:
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
#   go install github.com/go-kratos/kratos/cmd/kratos/v2@v2.8.3
#   go install github.com/google/wire/cmd/wire@latest
#   ...

# ② 生成 Proto 代码
make api        # 生成 api/**/*.pb.go / *_grpc.pb.go / *_http.pb.go
make config     # 生成 internal/conf/conf.pb.go
make validate   # 生成 validate 校验代码

# ③ 生成 Wire 依赖注入代码（所有服务）
make wire
# 或单独某个服务:
cd app/service/user && wire

# ④ 启动单个服务（以 user 为例）
cd app/service/user
go run ./cmd/user
# 启动时需要设置 CONSUL_ADDR 和 CONSUL_PATHS 环境变量
CONSUL_ADDR=127.0.0.1:8500 \
CONSUL_PATHS="/develop/common,/develop/cyi-app/user" \
go run ./cmd/user

# ⑤ 启动网关
cd app/cin/interface && go run ./cmd/interface
```

### 10.2 Docker Compose 一键启动

```bash
# 启动所有服务（基础设施 + 应用）
cd deploy/docker-compose
docker compose up -d

# 仅启动基础设施
docker compose -f docker-compose.infra.yml up -d

# 仅启动应用服务
docker compose -f docker-compose.app.yml up -d

# 停止
docker compose down
```

### 10.3 Kubernetes 部署

```bash
cd deploy/k8s

# Windows PowerShell 一键部署
./deploy-all.ps1

# 或手动执行
kubectl apply -f config/
kubectl apply -f infrastructure/consul/
kubectl apply -f infrastructure/mysql/
kubectl apply -f infrastructure/redis/
kubectl apply -f infrastructure/jaeger/
kubectl apply -f microservices/
kubectl apply -f monitoring/
```

### 10.4 Docker 镜像手动构建

```bash
# 构建单个服务
cd app/service/user
make docker

# 或使用根 Makefile（构建所有服务）
make docker-build
```

---

## 11. Makefile 与构建工具

### 11.1 根目录 Makefile 目标

| Target | 功能 |
|--------|------|
| `make init` | 安装 protoc-gen-go/kratos/wire 等工具 |
| `make api` | 生成 api 目录下的 Proto 代码（Go/gRPC/HTTP/OpenAPI） |
| `make config` | 生成 internal/conf 配置 Proto 代码 |
| `make validate` | 生成 validate 校验代码 |
| `make errors` | 生成错误码代码 |
| `make build` | 编译 Go 二进制到 `bin/` |
| `make generate` | `go generate ./...` + `go mod tidy` |
| `make wire` | 为所有服务生成 Wire DI 代码 |
| `make docker-build` | 为所有服务构建 Docker 镜像 |
| `make docker-up` / `docker-down` | Docker Compose 启停全部 |
| `make docker-up-infra` / `docker-up-app` | 仅启动基础设施/应用 |
| `make k8s-up` / `k8s-down` | K8s 部署启停 |
| `make service <name>` | 通过 kratos 脚手架生成新服务模板 |
| `make env` | 显示环境变量（Go 版本等） |
| `make help` | 显示所有可用目标 |

### 11.2 app_makefile.mk（所有服务共用模板）

被各服务的 `Makefile` `include`，提供通用 `make build`、`make docker`、Proto 生成等逻辑。核心：

- **`APP_RELATIVE_PATH`**：自动计算服务相对根目录路径（如 `service/user`）
- **`DOCKER_IMAGE`**：生成 `cyi3muz/service-user:0.1.0` 镜像名

---

## 12. CI/CD 流水线

### 12.1 `.github/workflows/ci.yml` - 构建与部署

| 步骤 | Trigger | 说明 |
|-----|---------|------|
| **detect-changes** | push / PR | `git diff` 检测哪些服务变更，输出矩阵 `{interface, user, job, task}` |
| **build** | 无条件执行 | `go mod tidy` → `go test` → `make build` |
| **docker-build** | push 到 main/develop | 对变更服务执行 `make docker` |
| **deploy** | push 到 main/develop | `kubectl apply -f microservices/` + `monitoring/` |

### 12.2 `.github/workflows/deploy-infrastructure.yml` - 基础设施部署

- **Trigger**：`deploy/k8s/infrastructure/**`、`deploy/k8s/config/**` 变更，或手动 `workflow_dispatch`
- **步骤**：配置 kubectl → `kubectl apply` 基础设施目录

### 12.3 GitHub Secrets 依赖

```
DOCKER_USERNAME / DOCKER_PASSWORD   # Docker Hub 登录
KUBE_CONFIG                         # Base64 编码的 kubeconfig
```

---

## 13. 架构设计亮点与注意事项

### 13.1 ✅ 设计亮点

| 方面 | 说明 |
|-----|------|
| **统一入口模式** | 所有 6 个服务的 `main.go` 遵循完全相同的 5 步流程，维护成本低 |
| **Repository 抽象** | biz 层仅依赖接口，data 层负责具体实现（MySQL+Redis），便于测试与替换 |
| **订单状态模式** | 每个状态独立实现 `OrderStatus` 接口，新增状态只需加一个类，无需修改已有逻辑 |
| **网关防击穿策略** | `singleflight`（单机并发合并）+ `Redis SETNX 分布式锁`（多实例并发）双重保护 |
| **JWT 无状态鉴权** | HS256 对称签名，通过 context 传递用户 ID，无需额外会话存储 |
| **可观测性齐全** | Jaeger 链路追踪 + ELK 日志聚合 + Prometheus/Grafana 指标监控 |
| **密码安全** | 使用 bcrypt (cost=12) 而非明文或简单哈希 |
| **多环境部署支持** | 支持本地开发、Docker Compose、Kubernetes 三种运行模式 |
| **配置中心化** | 通过 Consul 管理所有服务配置，运行时动态热更新 |

### 13.2 ⚠️ 注意事项与改进建议

| 问题 | 说明 | 建议 |
|-----|------|------|
| **Transaction 中间件已弃用** | `pkg/kratos/server/middleware/transaction.go` 使用 `map[string]*gorm.DB` 本地内存存事务 ID，不适合分布式场景 | 建议使用分布式事务方案（如 Saga、TCC）或数据库本地事务替代 |
| **CI/CD 镜像命名不一致** | CI `docker-build` 产生 `cyi3muz/...` 镜像名，`deploy` 步骤 `sed` 替换为 `DOCKER_USERNAME/cyi-...:GITHUB_SHA` | 统一命名规范 |
| **K8s Secret 明文** | `MYSQL_ROOT_PASSWORD` 等仅 Base64 编码，不是加密 | 使用 Vault 或 Sealed Secrets 管理密钥 |
| **order/product 服务在 K8s 缺失** | docker-compose 中存在 order/product 容器镜像配置，但 K8s `microservices/` 目录下未对应 | 补齐 K8s Deployment/Service |
| **Product 服务处于脚手架** | Product 服务 CRUD 方法 Protobuf 定义但无具体实现 | 按 user 服务模式完善 biz/data/service 层 |
| **Redis 分布式锁无自动续期** | `AcquireDistributedLock` 只有 SETNX 无 Watchdog 续租，长任务可能过期 | 增加 Lock 自动续期机制（如 Redlock） |
| **密码通过 API Key 签名** | 登录接口接受 `apiKey` 字段，密钥管理需谨慎 | 使用环境变量或 KMS 管理密钥 |

### 13.3 关键依赖注入图解

以 `cin-interface` 网关服务为例：

```
main()
  │
  ├── bc (Bootstrap 配置)   ← 从 Consul 加载
  ├── logger (zerolog)
  ├── tp (TracerProvider)   ← Jaeger
  └── wireApp(bc.Server, bc.Data, bc.Auth, logger, tp, ...) → kratos.App
                              │
  ┌───────────────────────────┼───────────────────────────┐
  │                           ▼                           │
  │              ┌───────────────────────┐                │
  │              │   server.ProviderSet  │                │
  │              │   HTTP + gRPC Server  │                │
  │              └──────────┬────────────┘                │
  │                         │                             │
  │                         ▼                             │
  │              ┌───────────────────────┐                │
  │              │  service.ProviderSet  │                │
  │              │  CinInterface Service │                │
  │              └──────────┬────────────┘                │
  │                         │                             │
  │                         ▼                             │
  │              ┌───────────────────────┐                │
  │              │     biz.ProviderSet   │                │
  │              │  UserUseCase +        │                │
  │              │  OrderUseCase         │                │
  │              └──────────┬────────────┘                │
  │                         │                             │
  │                         ▼                             │
  │              ┌───────────────────────┐                │
  │              │    data.ProviderSet   │                │
  │              │  UserRepo(gRPC) +     │                │
  │              │  OrderRepo(gRPC) +    │                │
  │              │  Redis Client +       │                │
  │              │  Asynq Client          │                │
  │              └───────────────────────┘                │
  │                                                        │
  └────────────────────────────────────────────────────────┘
```

---

> 📌 **文档维护提示**：本 Wiki 基于 `cyi-app` 当前代码状态生成。后续新增服务或修改 Proto 定义时，请同步更新对应章节。建议在 CI 流程中加入文档自动生成钩子，以保持 Wiki 与代码同步。
