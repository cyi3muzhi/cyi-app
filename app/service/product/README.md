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
│   │   ├── stock.go     # 库存数据访问（四个写操作走 DTM 屏障事务 runWithBarrier，原生条件更新 SQL）
│   │   └── test_helper_test.go # 测试辅助
│   ├── nats/
│   │   └── stock_handler.go # NATS 库存请求处理器（订阅 reserve/deduct/release/add/get，透传屏障键给 biz 层）
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
| `ProductUseCase` | 商品管理（创建、查询、更新、删除、搜索联想 SuggestProduct） |
| `StockUseCase` | 库存管理（查询、预留、扣减、释放、增加、按商品设置总库存）；预留/扣减/释放/增加四个写操作支持 DTM 屏障幂等：NATS 请求携带 `gid/branch_id/op` 三元组时在 `dtm_barrier.barrier` 屏障事务内执行，重试/补偿/重复请求只生效一次（空补偿自动跳过、防悬挂）；无幂等键（管理端 gRPC 直调）退化为普通事务 |
| `CategoryUseCase` | 商品分类管理（分类树、创建、更新、删除、排序、状态管理） |
| `SkuUseCase` | SKU 管理（列表、创建、更新、删除） |
| `SpecGroupUseCase` | 规格组管理（列表、创建、更新、删除，含规格值） |

### 商品值对象

- `Product`：商品聚合根
- `Stock`：库存值对象
- `Category`：商品分类（树形结构，支持无限层级父子关系）

### 库存操作与 DTM 屏障幂等

订单服务经 NATS（subject：`stock.reserve`/`stock.deduct`/`stock.release`/`stock.add`/`stock.get`）调用库存操作，写操作请求体携带 DTM 屏障三元组：

- `gid`：全局事务ID（下单预留/释放为 `tcc-{orderNo}`，支付扣减/退还为 `pay-{orderNo}`）
- `branch_id`：分支ID（按 SKU 维度 `stock-{skuID}`）
- `op`：`try` / `cancel`

product 侧（`internal/data/stock.go` 的 `runWithBarrier`）以 `dtmcli.BranchBarrier{TransType:"tcc",...}.CallWithDB` 在屏障事务内执行原生条件更新 SQL，屏障记录与库存变更同事务提交：

| 操作 | SQL 语义 | 条件不命中 |
|------|----------|-----------|
| 预留 reserve | available-、reserved+（`WHERE available>=?`） | 返回库存不足错误（业务失败 409，触发 TCC 回滚） |
| 扣减 deduct | total-、reserved-（`WHERE reserved>=?`），成功后累加销量 | 返回操作失败错误 |
| 释放 release | reserved-、available+（`WHERE reserved>=?`） | **幂等空操作返回 nil**（预留已扣减/已释放不报错，由屏障裁决） |
| 增加 add | total+、available+，成功后回退销量 | 返回操作失败错误 |

屏障表 `dtm_barrier.barrier` 与业务库同 MySQL 实例跨库访问；屏障连接由 `pkg/dtm.BarrierDB` 提供。空补偿（cancel 先于/无 try）由屏障自动跳过，悬挂（try 晚于 cancel）被屏障拦截。

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateProduct` | gRPC | 创建商品 |
| `GetProduct` | gRPC | 查询商品 |
| `ListProduct` | gRPC | 商品列表（支持 MySQL FULLTEXT 关键词搜索，ngram 分词器支持中文；支持 `stock_eq` 库存精确筛选、`stock_lte` 库存上限筛选（库存预警）） |
| `SuggestProduct` | gRPC | 搜索联想（返回匹配关键词的商品名列表，用于搜索框下拉联想） |
| `UpdateProduct` | gRPC | 更新商品 |
| `DeleteProduct` | gRPC | 删除商品 |

### 全文搜索部署要求

FULLTEXT 搜索依赖以下 MySQL 配置（否则可能报错或搜不到）：

1. **组合索引**：`ListProduct` 使用 `MATCH(name, description)`，需组合 FULLTEXT 索引 `ft_product_search (name, description)`；`SuggestProduct` 使用 `MATCH(name)`，需单列索引 `ft_product_name (name)`。服务启动时 `ensureFullTextIndexes` 自动创建（幂等），详见实现方案文档。
2. **空停用词表**：ngram 默认英文停用词会过滤含 `i`/`on` 等 token 的英文关键词（如 "ip" 搜不到 "iPhone"）。需手动配置一次：
   - 创建 `cin_test.ft_stopwords` 表（含一行占位值）
   - `SET PERSIST innodb_ft_server_stopword_table = 'cin_test/ft_stopwords'`
   - 重建两个 FULLTEXT 索引
| `GetStock` | gRPC | 查询库存 |
| `UpdateStock` | gRPC | 更新库存（按商品ID设置总库存，均分到该商品所有 SKU，管理端补货用） |
| `UpdateSkuStocks` | gRPC | 按 SKU 批量设置库存（管理端按 SKU 明细精确补货，带商品归属校验） |
| `ReserveStock` | gRPC | 预留库存 |
| `DeductStock` | gRPC | 扣减库存 |
| `ReleaseStock` | gRPC | 释放预留库存 |
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