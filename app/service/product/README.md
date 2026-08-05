# Product 服务

商品管理服务，提供商品信息管理、库存管理和商品分类管理功能。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
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
│   │   ├── errors.go    # 业务错误定义
│   │   ├── product.go   # 商品业务实体
│   │   ├── product_handle.go # 商品处理（ProductUseCase）
│   │   ├── product_handle_test.go # 商品处理测试
│   │   ├── product_value.go  # 商品值对象
│   │   ├── product_value_test.go # 商品值对象测试
│   │   ├── sku.go       # SKU 业务逻辑
│   │   ├── sku_test.go  # SKU 测试
│   │   ├── spec_group.go # 规格组业务逻辑
│   │   ├── spec_group_test.go # 规格组测试
│   │   ├── stock.go     # 库存业务逻辑
│   │   ├── stock_handle.go   # 库存处理
│   │   ├── stock_handle_test.go # 库存处理测试
│   │   ├── stock_value.go    # 库存值对象
│   │   ├── category.go       # 分类业务实体
│   │   ├── category_handle.go # 分类业务逻辑
│   │   └── category_handle_test.go # 分类业务测试
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── curd/
│   │   │   ├── gen.go   # 代码生成工具
│   │   │   ├── product.gen.go # 商品 CURD
│   │   │   ├── sku.gen.go     # SKU CURD
│   │   │   ├── spec_group.gen.go # 规格组 CURD
│   │   │   ├── spec_value.gen.go # 规格值 CURD
│   │   │   ├── stock.gen.go   # 库存 CURD
│   │   │   └── category.gen.go # 分类 CURD
│   │   ├── data.go      # 数据层初始化
│   │   ├── models/
│   │   │   ├── product.go # 商品模型
│   │   │   ├── sku.go     # SKU 模型
│   │   │   ├── spec_group.go # 规格组模型
│   │   │   ├── spec_value.go # 规格值模型
│   │   │   ├── stock.go   # 库存模型
│   │   │   └── category.go # 分类模型
│   │   ├── category.go  # 分类数据访问
│   │   ├── category_test.go # 分类数据访问测试
│   │   ├── product.go   # 商品数据访问
│   │   ├── product_test.go # 商品数据访问测试
│   │   ├── sku.go       # SKU 数据访问
│   │   ├── spec_group.go # 规格组数据访问
│   │   ├── spec_group_test.go # 规格组数据访问测试
│   │   ├── stock.go     # 库存数据访问
│   │   └── test_helper_test.go # 测试辅助
│   ├── nats/
│   │   └── stock_handler.go # NATS 库存事件处理器（库存变更广播消费）
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── category.go  # 分类服务
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
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.oss.endpoint` | OSS 地址 | - |
| `data.oss.access_key` | OSS 访问密钥 | - |
| `data.oss.secret_key` | OSS 访问密钥 Secret | - |
| `auth.api_key` | API 密钥 | - |
| `trace.endpoint` | Jaeger 链路追踪端点 | - |
| `log.level` | 日志级别（debug/info/warn/error） | - |
| `log.enable_console` | 是否启用控制台输出 | - |
| `log.enable_file` | 是否启用文件输出 | - |
| `log.log_dir` | 日志目录 | - |
| `registry.consul.address` | Consul 注册中心地址 | - |
| `registry.consul.scheme` | Consul 协议（http/https） | - |
| `registry.consul.enable_health_check` | 是否启用健康检查 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `ProductUseCase` | 商品管理（创建、查询、更新、删除） |
| `StockUseCase` | 库存管理（查询、扣减、增加） |
| `CategoryUseCase` | 商品分类管理（分类树、创建、更新、删除、排序、状态管理） |
| `SkuUseCase` | SKU 管理（列表、创建、更新、删除） |
| `SpecGroupUseCase` | 规格组管理（列表、创建、更新、删除，含规格值） |

### 商品值对象

- `Product`：商品聚合根
- `Stock`：库存值对象
- `Category`：商品分类（树形结构，支持无限层级父子关系）

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
| `ListCategoryTree` | gRPC | 获取分类树 |
| `GetCategory` | gRPC | 获取单个分类 |
| `CreateCategory` | gRPC | 创建分类 |
| `UpdateCategory` | gRPC | 更新分类 |
| `DeleteCategory` | gRPC | 删除分类 |
| `UpdateCategorySort` | gRPC | 更新分类排序 |
| `UpdateCategoryStatus` | gRPC | 更新分类状态 |
| `ToggleCategoryShowcase` | gRPC | 切换首页展示 |

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