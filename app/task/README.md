# Task 服务

异步任务服务，基于 Asynq 实现分布式任务队列处理。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 任务队列 | Asynq |
| 数据库 | Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/task/
├── cmd/task/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── configs/
│   ├── auth.yaml        # 认证配置
│   ├── config.yaml      # 主配置
│   └── register.yaml    # 注册配置
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── coupon.go    # 优惠券业务接口（cron 批量过期）
│   │   ├── export.go    # 订单导出业务逻辑（消费端）
│   │   ├── order_timeout.go # 订单超时自动取消业务逻辑（消费端）
│   │   ├── task.go      # 任务业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── client_conn.go # 客户端连接（user/order/marketing gRPC）
│   │   ├── coupon.go    # 优惠券数据访问（营销服务 gRPC）
│   │   ├── data.go      # 数据层初始化
│   │   ├── excel.go     # Excel 生成与上传 OSS
│   │   ├── export.go    # 导出任务状态仓库（Redis）
│   │   ├── order.go     # 订单查询（导出用）+ 订单超时取消（order gRPC）
│   │   ├── task.go      # 任务数据访问
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── http.go      # HTTP 服务注册
│   │   ├── server.go    # 服务器配置（asynq Server + Scheduler）
│   │   └── task.go      # Asynq 任务处理器
│   └── service/
│       ├── service.go   # 服务层依赖注入
│       └── task.go      # 任务服务
└── Makefile             # 构建脚本
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.service_key` | 服务密钥映射（user/marketing） | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.oss.endpoint` | OSS（RustFS）内部端点 | - |
| `data.oss.public_endpoint` | OSS 对外端点（生成下载链接） | - |
| `data.oss.access_key` | OSS AccessKey | - |
| `data.oss.secret_key` | OSS SecretKey | - |
| `tasks[].cron` | 定时任务 cron 表达式 | - |
| `tasks[].name` | 定时任务名称（如 `user_count`、`coupon_expire`） | - |
| `tasks[].options.queue` | 定时任务投递队列（critical） | critical |
| `tasks[].options.max_retry` | 最大重试次数 | - |
| `tasks[].options.timeout` | 任务超时时间 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `TaskUseCase` | 异步任务调度、执行 |
| `UserUseCase` | 用户相关异步任务（含 `UserCount` 定时统计） |
| `ExportUseCase` | 订单导出任务消费（查订单→生成Excel→上传OSS→更新状态） |
| `CouponRepo` | 优惠券批量过期（cron 定时任务，经 marketing gRPC 调用） |
| `OrderTimeoutUseCase` | 订单超时自动取消（延迟任务消费 + cron 兜底扫描：校验订单状态后调 order gRPC 取消） |

### 任务处理

服务启动时启动 Asynq Server 消费任务队列，同时启动 Scheduler 调度 cron 定时任务：

```go
go func() {
    if err := server.Run(mux); err != nil {
        // 处理错误
    }
}()
```

#### 订单导出任务（由 admin 服务投递）

admin 服务接收到 `/admin/v1/orders/export` 请求后，通过 `asynq.Client` 将任务投递到 `export` 队列，立即返回 `task_id`；task 服务消费该任务：

1. 标记任务为 `processing`
2. 通过 order gRPC + user gRPC 分页拉取订单
3. 生成 Excel 并上传到 OSS（RustFS）
4. 标记任务为 `completed`（记录 `object_key`）

admin 通过 `task_id` 从 Redis 读取任务状态，`completed` 时根据 `object_key` 动态生成 presigned 下载链接返回前端。

> 任务状态结构与 key 前缀由共享包 `pkg/exporttask` 定义，admin 与 task 共同引用以保证一致。

#### 订单超时自动取消任务（由订单服务投递延迟任务）

用户下单后（待付款状态），订单服务通过 asynq 延迟队列投递 `delay:order_timeout_cancel` 任务（`ProcessIn(30min)` + 固定 TaskID `order_timeout:{orderId}` 去重）；task 服务消费时：

1. 解析 payload 获取 `order_id`
2. 通过 order gRPC `GetOrder` 查询订单状态
3. 仅当订单仍为**待付款**（status=1）时调 `CancelOrder`（userID=0 系统操作）自动取消并释放预留库存
4. 已支付/已取消/已完成等状态幂等跳过

> 任务类型与 payload 结构由共享包 `pkg/ordertask` 定义，订单服务与 task 共同引用以保证一致。

此外，还注册了 cron 定时任务 `order_timeout_check`（每分钟）作为**兜底扫描**：调用订单服务 `ListOverdueWaitPayOrders`（status=1 且创建时间早于 now-30min，秒级过滤由数据库下推），把超时未支付的订单自动取消。该机制既覆盖功能上线前的存量待付款订单，也兜底下单时延迟任务丢失的场景，保证超时订单最终一定被取消。

#### cron 定时任务（Scheduler 调度）

服务启动时通过 `NewScheduler` + `RegisterScheduler` 注册配置中的定时任务（`config.yaml` 的 `tasks` 段），按 cron 表达式周期投递到 `critical` 队列并自行消费：

| 任务名 | cron | 说明 |
|--------|------|------|
| `user_count` | `*/1 * * * *` | 每分钟统计用户数量（调用 user 服务） |
| `coupon_expire` | `*/5 * * * *` | 每 5 分钟批量过期优惠券发放记录（调用 marketing 服务） |
| `order_timeout_check` | `*/1 * * * *` | 每分钟扫描并取消超时未支付的待付款订单（调用 order 服务，兜底保障） |

> 定时任务原属于独立 job 服务，已合并进 task 服务统一管理。

### 任务队列（task 服务消费）

task 服务消费以下队列（在 `server/server.go` 的 `NewTaskServer` 中声明）：

- `critical`: cron 定时任务专用队列（`user_count`、`coupon_expire`），权重 5
- `default`: 默认队列（`email:send`、`create_user` 等异步任务 + `delay:order_timeout_cancel` 订单超时延迟任务），权重 3
- `export`: 导出队列（订单导出任务），权重 2

## 运行方式

### 本地开发

```bash
cd app/task
go run cmd/task/main.go -conf configs/
```

### 构建部署

```bash
cd app/task
make build
./bin/task -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-task`

## 服务发现

服务注册到 Consul，服务名：`task`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`