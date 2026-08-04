# Order 服务

订单管理服务，提供订单创建、查询、支付功能，采用**状态模式**实现完整的订单状态机管理，支持正向流程和退款流程。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL（GORM + GORM Gen）、Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger (OpenTelemetry) |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/service/order/
├── cmd/
│   └── order/
│       ├── main.go              # gRPC 服务入口
│       ├── wire.go              # 依赖注入配置
│       └── wire_gen.go          # Wire 生成文件
├── internal/
│   ├── biz/
│   │   ├── biz.go                       # 业务层依赖注入
│   │   ├── order.go                     # 订单聚合根
│   │   ├── order_handle.go              # OrderUseCase 业务逻辑（含下单后投递超时取消任务）
│   │   ├── order_handle_test.go         # OrderUseCase 单元测试
│   │   ├── order_async.go               # OrderTimeoutTaskRepo 接口（投递超时取消延迟任务）
│   │   ├── order_item.go                # 订单行项目实体
│   │   ├── order_value.go               # Money 值对象
│   │   ├── order_status.go              # 订单状态接口、基类、工厂、Order 委托方法
│   │   ├── wait_pay_status.go           # 待支付状态处理器
│   │   ├── pay_status.go                # 已支付/待发货状态处理器
│   │   ├── ship_status.go               # 已发货/待收货状态处理器
│   │   ├── complete_status.go           # 已完成状态处理器
│   │   ├── terminal_status.go           # 终端状态处理器（已取消/已退款）
│   │   ├── refunding_status.go          # 退款中状态处理器
│   │   ├── service.go                   # ProductRepo、StockRepo 接口定义
│   │   ├── order_test.go                # 订单单元测试
│   │   ├── order_status_test.go         # 状态机测试
│   │   └── order_status_stock_test.go   # 状态机库存测试
│   ├── conf/
│   │   ├── conf.proto       # 配置定义
│   │   └── conf.pb.go       # 配置结构体
│   ├── data/
│   │   ├── data.go          # 数据层初始化（MySQL、Redis、NATS、GORM Gen）
│   │   ├── asynq.go         # asynq 客户端 + 订单超时任务投递（延迟队列）
│   │   ├── client_conn.go   # 商品/库存服务 gRPC 客户端连接
│   │   ├── order.go         # 订单数据访问
│   │   ├── product.go       # 商品仓储实现（远程调用商品服务）
│   │   ├── stock_nats.go    # 库存仓储实现（远程调用商品服务，NATS 事件广播库存变更）
│   │   ├── curd/
│   │   │   ├── gen.go           # GORM Gen 代码生成工具
│   │   │   ├── order.gen.go     # 订单 CURD
│   │   │   ├── order_item.gen.go # 订单项 CURD
│   │   │   ├── order_logistics.gen.go # 物流 CURD
│   │   │   └── payment_log.gen.go    # 支付日志 CURD
│   │   └── models/
│   │       ├── order.go         # 订单模型
│   │       ├── order_item.go    # 订单项模型
│   │       ├── logistics.go     # 物流模型
│   │       └── payment_log.go   # 支付日志模型
│   ├── server/
│   │   ├── server.go       # 服务器配置（Consul 注册/发现、gRPC/HTTP 服务器）
│   │   └── grpc.go         # gRPC 服务注册
│   └── service/
│       ├── service.go      # 服务层依赖注入
│       └── order.go        # Order 服务（gRPC）
└── Makefile                # 构建脚本
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| **Server** | | |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.http.timeout` | HTTP 超时 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `server.grpc.timeout` | gRPC 超时 | - |
| **Data** | | |
| `data.database.driver` | 数据库驱动 | mysql |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.network` | Redis 网络类型 | tcp |
| `data.redis.addr` | Redis 地址 | - |
| `data.redis.password` | Redis 密码 | - |
| `data.redis.read_timeout` | Redis 读取超时 | - |
| `data.redis.write_timeout` | Redis 写入超时 | - |
| **Auth** | | |
| `auth.api_key` | API 密钥 | - |
| `auth.service_key` | 服务间调用密钥 | - |
| **Registry** | | |
| `registry.consul.address` | Consul 地址 | - |
| `registry.consul.scheme` | Consul 协议（http/https） | http |
| `registry.consul.token` | Consul ACL 令牌 | - |
| `registry.consul.username` | Consul 基本认证用户名 | - |
| `registry.consul.password` | Consul 基本认证密码 | - |
| `registry.consul.enable_health_check` | 是否启用健康检查 | false |
| `registry.consul.check_interval` | 健康检查间隔（秒） | - |
| **Trace** | | |
| `trace.endpoint` | Jaeger 采集端点 | - |
| **Log** | | |
| `log.level` | 日志级别 | info |
| `log.enable_console` | 是否启用控制台输出 | true |
| `log.enable_file` | 是否启用文件输出 | false |
| `log.log_dir` | 日志目录 | - |
| `log.time_format` | 时间格式 | - |
| `log.enable_color` | 是否启用颜色 | false |

## 核心业务逻辑

### UseCase

| UseCase | 职责 |
|---------|------|
| `OrderUseCase` | 订单管理（创建、查询、更新、支付、发货），含商品校验、库存校验、预留库存、扣减库存、物流信息写入；下单成功后投递 30 分钟超时自动取消延迟任务（Asynq）；支付幂等 + 失败补偿库存 + 写支付流水；取消自动返还优惠券 |
| `OrderTimeoutTaskRepo` | 订单超时任务投递仓库（data 层基于 asynq 延迟队列实现：`ProcessIn(30min)` + 固定 TaskID 去重） |

> 待付款订单超时自动取消链路：**订单服务**下单成功 → 投递 `delay:order_timeout_cancel` 延迟任务（30 分钟，固定 TaskID `order_timeout:{orderId}`）→ **task 服务**消费时通过 `GetOrder` 校验状态，仍为待付款才调 `CancelOrder`（userID=0 系统操作）自动取消，已支付/已取消等状态幂等跳过。共享类型定义于 `pkg/ordertask`。

> **并发安全**：状态转换（支付/取消/收货）通过 `UpdateStatusIf` 条件更新落库（`UPDATE ... WHERE id=? AND status=?`），行数=0 表示状态已被并发变更（如支付时被超时取消），支付侧自动补偿已扣库存，避免库存泄漏。支付成功后写入 `payment_log` 支付流水；取消订单时若核销过优惠券则自动调用 marketing `ReleaseCoupon` 返还。

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateOrder` | gRPC | 创建订单（支持立即购买和购物车结算两种模式；结算可携带收货地址快照写入物流表、运费） |
| `UpdateOrder` | gRPC | 更新订单 |
| `DeleteOrder` | gRPC | 删除订单 |
| `GetOrder` | gRPC | 查询订单详情（含商品明细、物流信息） |
| `ListOrder` | gRPC | 订单列表 |
| `PayOrder` | gRPC | 支付订单（含扣减库存） |
| `OrderStats` | gRPC | 订单统计（按状态分组计数：总数/待处理/今日/退款） |
| `BatchGetOrder` | gRPC | 批量获取订单详情（含商品明细，用于导出场景，1 次调用替代 N 次 GetOrder） |
| `GetUserOrderStats` | gRPC | 用户订单统计（个人中心：各状态订单数 + 累计消费） |
| `ShipOrder` | gRPC | 发货（管理端调用：待发货→已发货，写入物流公司和快递单号） |
| `ReceiveOrder` | gRPC | 确认收货（用户端调用：已发货/待收货→已完成，校验订单归属） |
| `CancelOrder` | gRPC | 取消订单（用户端/管理端调用：待支付/已支付→已取消，含释放预留库存；userID>0 校验归属；若核销过优惠券自动返还） |
| `ListOverdueWaitPayOrders` | gRPC | 查询超时未支付的待付款订单（task 服务 cron 兜底扫描用，status=1 且创建时间早于 now-timeoutSeconds，秒级下推过滤） |

### 订单状态机

#### 状态定义

| 枚举值 | 名称 | 说明 |
|--------|------|------|
| 0 | `OrderStatusUnspecified` | 未指定（零值，无效状态，所有操作被拒绝） |
| 1 | `WaitPay` | 待支付（初始状态） |
| 2 | `Pay` | 已支付 |
| 3 | `WaitShip` | 待发货（与 Pay 共享行为） |
| 4 | `Ship` | 已发货 |
| 5 | `WaitReceive` | 待收货（与 Ship 共享行为） |
| 6 | `Complete` | 已完成 |
| 7 | `Cancel` | 已取消（终态） |
| 8 | `Refunding` | 退款中 |
| 9 | `Refunded` | 已退款（终态） |

#### 状态转换矩阵

```
              Pay  Ship  Receive  Complete  Cancel  Refunding  Refunded
Unspecified   ✗    ✗      ✗        ✗        ✗       ✗          ✗
WaitPay       ✓    ✗      ✗        ✗        ✓       ✗          ✗
Pay/WaitShip  ✗    ✓      ✗        ✗        ✓       ✓          ✗
Ship/WaitRec  ✗    ✗      ✓        ✗        ✗       ✓          ✗
Complete      ✗    ✗      ✗        ✗        ✗       ✓          ✗
Cancel        ✗    ✗      ✗        ✗        ✗       ✗          ✗
Refunding     ✗    ✗      ✗        ✗        ✗       ✗          ✓
Refunded      ✗    ✗      ✗        ✗        ✗       ✗          ✗
```

#### 架构说明

采用**状态模式（State Pattern）** 实现，核心结构：

```
Order (聚合根)
 ├── Status: OrderStatus        — 当前状态枚举
 ├── statusHandler: OrderStatus — 当前状态处理器
 │
 ├── Pay()      → 委托给 statusHandler.Pay()
 ├── Ship()     → 委托给 statusHandler.Ship()
 ├── ...
 │
 └── SetStatus() → 根据新状态创建对应的处理器

OrderStatus (接口)
 ├── statusBase (基类，提供默认实现 + transitionTo)
 │    ├── WaitPayStatus      — 待支付行为
 │    ├── PayStatus           — 已支付/待发货行为
 │    ├── ShipStatus          — 已发货/待收货行为
 │    ├── CompleteStatus      — 已完成行为
 │    ├── FinalStatus         — 终端状态（已取消/已退款，全部拒绝）
 │    └── RefundingStatus     — 退款中行为
```

- **默认拒绝**：`statusBase` 为所有操作提供默认实现，返回"无效操作"错误。各具体处理器只需重写**允许**的方法。
- **统一转换**：`transitionTo()` 方法封装状态转换逻辑，先写数据库（失败则回滚内存），成功后切换处理器。
- **库存联动**：状态转换时自动处理库存操作（取消释放预留库存、退款恢复库存等）。

#### 正向流程

```
用户下单 → WaitPay → Pay → Ship/WaitReceive → Complete
         ↘         ↗                    ↗
           Cancel  ←  ─ ─ ─ ─ ─ ─ ─ ─ ┘
```

**退款流程**：

```
Pay/Ship/Complete → Refunding → Refunded
```

#### 关键文件说明

| 文件 | 职责 |
|------|------|
| `order_status.go` | `OrderStatus` 接口、`statusBase` 基类、`NewOrderStatusHandler` 工厂、`Order` 聚合根的委托方法 |
| `wait_pay_status.go` | 待支付状态：允许 Pay、Cancel |
| `pay_status.go` | 已支付/待发货状态：允许 Ship、Cancel、Refunding |
| `ship_status.go` | 已发货/待收货状态：允许 Receive、Refunding |
| `complete_status.go` | 已完成状态：允许 Refunding |
| `terminal_status.go` | 终端状态（已取消/已退款）：全部拒绝 |
| `refunding_status.go` | 退款中状态：允许 Refunded |

#### 如何新增状态

1. 在 `api/service/order/v1/order.proto` 的 `OrderStatus` 枚举中新增值
2. 新建 `internal/biz/xxx_status.go`，实现 `OrderStatus` 接口（只需重写允许的操作）
3. 在 `NewOrderStatusHandler` 工厂中添加 case 分支
4. 在 `getStatusName` 中添加状态的中文描述
5. 在 `biz.go` 的 ProviderSet 中注册新状态处理器
6. 更新本文档的状态转换矩阵
7. 在 `order_status_test.go` 中添加对应的测试用例

## 运行方式

### 本地开发

```bash
cd app/service/order
go run cmd/order/main.go
```

### 构建部署

```bash
cd app/service/order
make build
./bin/order
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-order`

## 服务发现

服务注册到 Consul，服务名：`{SERVICE_ENV}.service.order`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`
