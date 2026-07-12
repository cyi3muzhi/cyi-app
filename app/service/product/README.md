# Product 服务

商品管理服务，提供商品信息管理和库存管理功能。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 数据库 | MySQL、Redis、OSS |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/service/product/
├── cmd/product/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── product.go   # 商品业务逻辑
│   │   ├── product_handle.go # 商品处理
│   │   ├── product_value.go  # 商品值对象
│   │   ├── stock.go     # 库存业务逻辑
│   │   ├── stock_handle.go   # 库存处理
│   │   └── stock_value.go    # 库存值对象
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── curd/
│   │   │   ├── gen.go   # 代码生成工具
│   │   │   ├── product.gen.go # 商品 CURD
│   │   │   └── stock.gen.go   # 库存 CURD
│   │   ├── data.go      # 数据层初始化
│   │   ├── models/
│   │   │   ├── product.go # 商品模型
│   │   │   └── stock.go   # 库存模型
│   │   ├── product.go   # 商品数据访问
│   │   └── stock.go     # 库存数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── product.go   # 商品服务
│       ├── service.go   # 服务层依赖注入
│       └── stock.go     # 库存服务
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.api_key` | API 密钥 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.oss.endpoint` | OSS 地址 | - |
| `data.oss.access_key` | OSS 访问密钥 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `ProductUseCase` | 商品管理（创建、查询、更新、删除） |
| `StockUseCase` | 库存管理（查询、扣减、增加） |

### 商品值对象

- `Product`：商品聚合根
- `Stock`：库存值对象

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateProduct` | gRPC | 创建商品 |
| `GetProduct` | gRPC | 查询商品 |
| `ListProducts` | gRPC | 商品列表 |
| `UpdateProduct` | gRPC | 更新商品 |
| `DeleteProduct` | gRPC | 删除商品 |
| `GetStock` | gRPC | 查询库存 |
| `DeductStock` | gRPC | 扣减库存 |
| `AddStock` | gRPC | 增加库存 |

## 运行方式

### 本地开发

```bash
cd app/service/product
go run cmd/product/main.go -conf configs/
```

### 构建部署

```bash
cd app/service/product
make build
./bin/product -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-product`

## 服务发现

服务注册到 Consul，服务名：`product`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`