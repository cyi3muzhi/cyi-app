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

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.service_key` | 服务密钥映射 | - |
| `auth.api_key` | API 密钥 | - |
| `auth.white_list` | 认证白名单 | - |
| `blades.api_key` | DeepSeek API 密钥 | - |
| `blades.model` | 模型名称 | deepseek-chat |
| `blades.api_endpoint` | API 地址 | https://api.deepseek.com/v1 |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |

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

| 接口 | 方法 | 说明 |
|------|------|------|
| `Login` | gRPC/HTTP | 用户登录 |
| `Register` | gRPC/HTTP | 用户注册 |
| `CreateOrder` | gRPC/HTTP | 创建订单 |
| `GetOrder` | gRPC/HTTP | 查询订单 |
| `ListOrders` | gRPC/HTTP | 订单列表 |
| `GetProduct` | gRPC/HTTP | 查询商品 |
| `ListProducts` | gRPC/HTTP | 商品列表 |
| `GetStock` | gRPC/HTTP | 查询库存 |
| `Chat` | gRPC/HTTP | AI 聊天 |

## 运行方式

### 本地开发

```bash
cd app/cin/interface
go run cmd/interface/main.go -conf configs/
```

### 构建部署

```bash
cd app/cin/interface
make build
./bin/interface -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-cin-interface`

## 服务发现

服务注册到 Consul，服务名：`cin-interface`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`