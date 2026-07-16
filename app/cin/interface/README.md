# CIN Interface 服务

接口网关服务，作为前端与后端服务的统一入口，提供用户、订单、商品、库存等业务接口。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 数据库 | MySQL、Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |
| AI 接口 | DeepSeek API |

## 目录结构

```
app/cin/interface/
├── cmd/interface/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── configs/
│   ├── auth.yaml        # 认证配置
│   ├── config.yaml      # 主配置
│   ├── register.yaml    # 注册配置
│   ├── ss.yaml          # 服务配置
│   └── transaction.yaml # 事务配置
├── internal/
│   ├── biz/
│   │   ├── auth.go      # 认证业务逻辑
│   │   ├── blades.go    # AI 助手业务
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── order.go     # 订单业务逻辑
│   │   ├── product.go   # 商品业务逻辑
│   │   ├── stock.go     # 库存业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── client_conn.go # 客户端连接
│   │   ├── data.go      # 数据层初始化
│   │   ├── order.go     # 订单数据访问
│   │   ├── product.go   # 商品数据访问
│   │   ├── stock.go     # 库存数据访问
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   ├── server.go    # 服务器配置
│   │   └── stream_route.go # 流式路由（WebSocket/SSE）
│   └── service/
│       ├── blades.go    # AI 助手服务
│       ├── order.go     # 订单服务
│       ├── product.go   # 商品服务
│       ├── service.go   # 服务层依赖注入
│       ├── stock.go     # 库存服务
│       └── user.go      # 用户服务
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

配置通过 Consul KV 加载，从 `CONSUL_ADDR` 和 `CONSUL_PATHS` 环境变量指定 Consul 地址和路径。

主要配置项（来自 `conf.proto`）：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.http.network` | HTTP 网络类型 | `tcp` |
| `server.http.addr` | HTTP 监听地址 | `:8000` |
| `server.http.timeout` | HTTP 超时时间 | - |
| `server.grpc.network` | gRPC 网络类型 | `tcp` |
| `server.grpc.addr` | gRPC 监听地址 | `:9000` |
| `server.grpc.timeout` | gRPC 超时时间 | - |
| `data.database.driver` | 数据库驱动 | `mysql` |
| `data.database.source` | 数据库连接字符串 | - |
| `data.redis.addr` | Redis 地址 | `localhost:6379` |
| `data.redis.password` | Redis 密码 | - |
| `data.redis.read_timeout` | Redis 读取超时 | - |
| `data.redis.write_timeout` | Redis 写入超时 | - |
| `auth.service_key` | 服务密钥映射 | - |
| `auth.api_key` | JWT 签名密钥 | - |
| `auth.white_list` | 认证白名单（无需认证的路径） | - |
| `trace.endpoint` | Jaeger 链路追踪端点 | - |
| `log.level` | 日志级别（debug/info/warn/error） | `info` |
| `log.enable_console` | 是否启用控制台日志输出 | `true` |
| `log.enable_file` | 是否启用文件日志输出 | `false` |
| `log.log_dir` | 日志文件目录 | - |
| `blades.api_key` | DeepSeek API 密钥 | - |
| `blades.model` | AI 模型名称 | `deepseek-chat` |
| `blades.api_endpoint` | API 地址 | `https://api.deepseek.com/v1` |
| `blades.system_prompt` | 系统提示词 | - |
| `blades.max_tokens` | 最大令牌数 | `4096` |
| `blades.temperature` | 温度参数 | `0.7` |
| `registry.consul.address` | Consul 注册中心地址 | - |
| `registry.consul.scheme` | Consul 协议 | `http` |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `UserUseCase` | 用户认证、注册、信息管理 |
| `OrderUseCase` | 订单创建、查询、状态管理 |
| `ProductUseCase` | 商品查询、管理 |
| `StockUseCase` | 库存查询、扣减 |
| `BladesUseCase` | AI 聊天助手 |

### 服务接口

#### gRPC/HTTP 接口（通过 Proto 注册）

| 服务 | 接口 | 方法 | 说明 |
|------|------|------|------|
| **Auth** | `Login` | gRPC/HTTP | 用户登录 |
| | `Register` | gRPC/HTTP | 用户注册 |
| **User** | `GetMyInfo` | gRPC/HTTP | 获取个人信息 |
| | `GetUserInfo` | gRPC/HTTP | 获取指定用户信息 |
| | `UpdateUserInfo` | gRPC/HTTP | 更新用户信息 |
| **Product** | `CreateProduct` | gRPC/HTTP | 创建商品 |
| | `UpdateProduct` | gRPC/HTTP | 更新商品 |
| | `ListProduct` | gRPC/HTTP | 商品列表 |
| | `GetProduct` | gRPC/HTTP | 商品详情 |
| **Order** | `PlaceOrder` | gRPC/HTTP | 下单 |
| | `PayOrder` | gRPC/HTTP | 支付订单 |
| | `CartCheckout` | gRPC/HTTP | 购物车结算 |
| | `ListOrder` | gRPC/HTTP | 订单列表 |
| | `GetOrder` | gRPC/HTTP | 订单详情 |
| **Cart** | `AddCartItem` | gRPC/HTTP | 加入购物车 |
| | `GetCart` | gRPC/HTTP | 获取购物车 |
| **Stock** | `UpdateStock` | gRPC/HTTP | 更新库存 |
| **Blades** | `AiChat` | gRPC/HTTP | AI 聊天（非流式） |

#### 流式路由（通过 HandleFunc 注册）

| 路由 | 协议 | 说明 |
|------|------|------|
| `POST /ai/v1/chat/stream` | SSE | AI 流式聊天（Server-Sent Events） |
| `WS /ws/chat` | WebSocket | AI 流式聊天（WebSocket） |

## 运行方式

> 配置从 Consul KV 加载，需通过环境变量指定 Consul 连接信息。

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONSUL_ADDR` | Consul 服务地址 | `127.0.0.1:8500` |
| `CONSUL_PATHS` | Consul KV 配置路径（逗号分隔） | - |
| `SERVICE_ENV` | 环境标识（用于服务名前缀） | - |

### 本地开发

```bash
cd app/cin/interface
$env:CONSUL_ADDR="127.0.0.1:8500"
$env:CONSUL_PATHS="/develop/cyi-app/cin/interface"
go run cmd/interface/main.go
```

### 构建部署

```bash
cd app/cin/interface
make build
./bin/interface
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-cin-interface`

## 服务发现

服务注册到 Consul，服务名：`cin-interface`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`