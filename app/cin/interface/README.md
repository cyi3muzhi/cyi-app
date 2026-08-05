# CIN Interface 服务

接口网关服务，作为前端与后端服务的统一入口，提供用户、订单、商品、库存等业务接口。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
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
│   │   ├── banner.go    # 轮播图业务
│   │   ├── blades.go    # AI 助手业务
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── cart.go      # 购物车业务逻辑
│   │   ├── category.go  # 分类业务
│   │   ├── coupon.go    # 优惠券业务逻辑
│   │   ├── order.go     # 订单业务逻辑
│   │   ├── product.go   # 商品业务逻辑
│   │   ├── settings.go  # 店铺设置业务
│   │   ├── stock.go     # 库存业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── banner.go    # 轮播图数据访问
│   │   ├── cart.go      # 购物车数据访问（Redis Hash）
│   │   ├── category.go  # 分类数据访问
│   │   ├── client_conn.go # 客户端连接
│   │   ├── coupon.go    # 优惠券数据访问（gRPC 调用 marketing service）
│   │   ├── data.go      # 数据层初始化
│   │   ├── order.go     # 订单数据访问
│   │   ├── product.go   # 商品数据访问
│   │   ├── settings.go  # 店铺设置数据访问
│   │   ├── stock.go     # 库存数据访问
│   │   ├── upload.go    # 文件上传数据访问（OSS 预签名）
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   ├── server.go    # 服务器配置
│   │   └── stream_route.go # 流式路由（WebSocket/SSE）
│   └── service/
│       ├── banner.go    # 轮播图服务
│       ├── blades.go    # AI 助手服务
│       ├── cart.go      # 购物车服务
│       ├── category.go  # 分类服务
│       ├── coupon.go    # 优惠券服务
│       ├── order.go     # 订单服务
│       ├── product.go   # 商品服务
│       ├── service.go   # 服务层依赖注入
│       ├── settings.go  # 店铺设置服务
│       ├── stock.go     # 库存服务
│       ├── upload.go    # 文件上传服务（获取上传凭证）
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
| `data.oss.endpoint` | OSS 服务地址 | - |
| `data.oss.public_endpoint` | OSS 公开访问地址（Nginx 代理） | - |
| `data.oss.access_key` | OSS 访问密钥 | - |
| `data.oss.secret_key` | OSS 密钥 | - |
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
| `UserUseCase` | 用户认证、注册、信息管理、修改密码、收货地址管理、个人中心概览 |
| `OrderUseCase` | 订单创建、查询、状态管理、物流轨迹查询 |
| `ProductUseCase` | 商品查询、管理、搜索联想（SuggestProduct） |
| `StockUseCase` | 库存查询、扣减 |
| `CartUseCase` | 购物车管理（Redis 存储 + 商品信息聚合，按商品+SKU 维度） |
| `CouponUseCase` | 优惠券领取、我的优惠券列表、领券中心（可领取券）、结算可用券筛选 |
| `BladesUseCase` | AI 聊天助手 |
| `MessageUseCase` | 消息中心（消息列表、未读数、标记已读/全部已读） |
| `BannerUseCase` | 轮播图查询 |
| `CategoryUseCase` | 商品分类查询 |
| `SettingsUseCase` | 店铺设置查询 |

### 服务接口

#### gRPC/HTTP 接口（通过 Proto 注册）

| 服务 | 接口 | 方法 | 说明 |
|------|------|------|------|
| **Auth** | `Login` | gRPC/HTTP | 用户登录 |
| | `Register` | gRPC/HTTP | 用户注册 |
| **User** | `GetMyInfo` | gRPC/HTTP | 获取个人信息 |
| | `GetUserInfo` | gRPC/HTTP | 获取指定用户信息 |
| | `UpdateUserInfo` | gRPC/HTTP | 更新用户信息（昵称/头像/邮箱/性别/生日） |
| | `ChangePassword` | gRPC/HTTP | 修改密码 |
| | `GetUserCenter` | gRPC/HTTP | 个人中心概览（用户信息 + 订单/优惠券统计） |
| | `ListAddress` | gRPC/HTTP | 收货地址列表 |
| | `CreateAddress` | gRPC/HTTP | 新增收货地址 |
| | `UpdateAddress` | gRPC/HTTP | 更新收货地址 |
| | `DeleteAddress` | gRPC/HTTP | 删除收货地址 |
| | `SetDefaultAddress` | gRPC/HTTP | 设置默认收货地址 |
| **Product** | `CreateProduct` | gRPC/HTTP | 创建商品 |
| | `UpdateProduct` | gRPC/HTTP | 更新商品 |
| | `ListProduct` | gRPC/HTTP | 商品列表（支持关键词 FULLTEXT 搜索） |
| | `GetProduct` | gRPC/HTTP | 商品详情 |
| | `SuggestProduct` | gRPC/HTTP | 搜索联想（输入关键词返回商品名建议，GET /product/v1/suggest，免鉴权） |
| **Order** | `PlaceOrder` | gRPC/HTTP | 下单 |
| | `PayOrder` | gRPC/HTTP | 支付订单 |
| | `ReceiveOrder` | gRPC/HTTP | 确认收货 |
| | `CancelOrder` | gRPC/HTTP | 取消订单 |
| | `CartCheckout` | gRPC/HTTP | 购物车结算下单（含收货地址快照与运费，支持优惠券 coupon_id） |
| | `ListOrder` | gRPC/HTTP | 订单列表（当前用户，按状态筛选，含商品明细、支付截止时间 pay_deadline） |
| | `GetOrder` | gRPC/HTTP | 订单详情（当前用户，含商品明细、物流信息、支付截止时间 pay_deadline） |
| | `ApplyRefund` | gRPC/HTTP | 申请退款（仅已支付/待发货订单，POST /order/v1/refund/apply） |
| | `ListRefund` | gRPC/HTTP | 我的退款记录（GET /order/v1/refunds，按状态筛选、分页） |
| | `GetLogistics` | gRPC/HTTP | 查询物流轨迹（当前用户订单，GET /order/v1/{order_id}/logistics，校验订单归属，发货后展示轨迹时间线） |
| **Cart** | `AddCartItem` | gRPC/HTTP | 加入购物车（同商品同规格数量累加，携带 sku_id） |
| | `GetCart` | gRPC/HTTP | 获取购物车（含商品信息聚合，按商品+SKU 维度返回价格/库存/规格） |
| | `UpdateCartItem` | gRPC/HTTP | 修改购物车商品数量（按商品+SKU） |
| | `ToggleCartItem` | gRPC/HTTP | 选中/取消选中购物车商品（按商品+SKU） |
| | `SelectAllCart` | gRPC/HTTP | 全选/取消全选 |
| | `DeleteCartItem` | gRPC/HTTP | 删除购物车商品（按商品+SKU，sku_id 缺省删除该商品全部规格） |
| | `ClearCart` | gRPC/HTTP | 清空购物车 |
| **Coupon** | `ClaimCoupon` | gRPC/HTTP | 领取优惠券（POST /api/v1/coupons/{coupon_id}/claim） |
| | `ListMyCoupons` | gRPC/HTTP | 我的优惠券列表（GET /api/v1/coupons/my，按状态筛选） |
| | `ListAvailableCoupons` | gRPC/HTTP | 领券中心可领取列表（GET /api/v1/coupons/available） |
| | `ListUsableCoupons` | gRPC/HTTP | 结算可用券（GET /api/v1/coupons/usable，按订单金额筛选） |
| **Banner** | `ListBanners` | gRPC/HTTP | 轮播图列表（启用中的） |
| **Category** | `ListCategoryTree` | gRPC/HTTP | 分类树（启用中的） |
| **Settings** | `GetSettings` | gRPC/HTTP | 获取店铺设置（名称、logo、公告等） |
| | `GetPublicSettings` | gRPC/HTTP | 获取公开设置（商城名称、联系方式、ICP、热门搜索词 hot_keywords，免鉴权） |
| **Stock** | `UpdateStock` | gRPC/HTTP | 更新库存 |
| **Upload** | `GetUploadToken` | gRPC/HTTP | 获取上传凭证（OSS 预签名直传，用于头像等） |
| **Blades** | `AiChat` | gRPC/HTTP | AI 聊天（非流式） |
| **Message** | `ListMyMessage` | gRPC/HTTP | 我的消息列表（GET /message/v1/list，分页 + 类型筛选，含未读数） |
| | `UnreadMessageCount` | gRPC/HTTP | 我的未读消息数（GET /message/v1/unread_count） |
| | `MarkMessageRead` | gRPC/HTTP | 标记单条消息已读（POST /message/v1/read/{id}） |
| | `MarkAllMessageRead` | gRPC/HTTP | 全部标记已读（POST /message/v1/read_all） |

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