# Marketing 服务

优惠券营销服务，提供优惠券的创建、管理、发放（领券）、核销、返还、批量过期及发放记录查询等功能，支撑管理后台优惠券运营与用户端领券中心、下单核销等场景。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL（GORM + gorm/gen 代码生成） |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/service/marketing/
├── cmd/marketing/
│   ├── main.go          # 服务入口（配置从 Consul 加载）
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── coupon.go    # 优惠券业务实体（类型/状态/有效期定义与校验）
│   │   └── coupon_handle.go # 优惠券处理（CouponUseCase）
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── curd/
│   │   │   ├── gen.go   # gorm/gen 代码生成工具
│   │   │   ├── coupon.gen.go # 优惠券 CURD
│   │   │   └── coupon_dist_record.gen.go # 发放记录 CURD
│   │   ├── data.go      # 数据层初始化
│   │   ├── models/
│   │   │   ├── coupon.go # 优惠券模型
│   │   │   └── coupon_dist_record.go # 发放记录模型
│   │   └── coupon.go    # 优惠券数据访问（领券事务、核销、批量过期等）
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── coupon.go    # 优惠券服务（CouponService）
│       └── service.go   # 服务层依赖注入
├── Makefile             # 构建脚本（复用根目录 app_makefile.mk）
```

## 配置说明

服务配置统一存储在 Consul KV 中，启动时自动加载。主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.grpc.addr` | gRPC 监听地址 | - |
| `server.grpc.timeout` | gRPC 超时时间 | - |
| `data.database.driver` | 数据库驱动 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.redis.password` | Redis 密码 | - |
| `auth.api_key` | 服务间调用 API 密钥 | - |
| `auth.service_key` | 各服务调用方密钥映射 | - |
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
| `CouponUseCase` | 优惠券全生命周期管理（创建、更新、删除、查询、列表、启停用、统计、领券、核销、返还、批量过期、发放记录查询） |

### 优惠券业务模型

- **类型**：1-立减、2-折扣、3-满减
- **有效期类型**：1-固定日期、2-领取后 N 天、3-永久有效
- **状态机**：启用（Active）、待生效（Pending）、已结束（Ended）、已作废（Voided），状态随当前时间动态计算

### 关键设计

- **领券防超发**：`ClaimCoupon` 在数据库事务中通过行锁原子更新发放数量并插入发放记录，防止并发超发；支持每人限领数量校验
- **核销与返还**：下单时调用 `UseCoupon` 核销并计算实际优惠金额；订单取消/退款时调用 `ReleaseCoupon` 解除核销（校验发放记录归属用户与核销订单一致）
- **批量过期**：`BatchExpireCoupons` 分批将过期发放记录置为已过期状态，供定时任务调用
- **领券中心**：`ListAvailableCoupons` 返回上架且有剩余库存的券，并标记当前用户是否已领取

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateCoupon` | gRPC | 创建优惠券 |
| `UpdateCoupon` | gRPC | 更新优惠券（已作废的券不可更新） |
| `DeleteCoupon` | gRPC | 删除优惠券 |
| `GetCoupon` | gRPC | 获取优惠券详情 |
| `ListCoupon` | gRPC | 优惠券列表（支持状态、关键词、创建时间筛选） |
| `ToggleCouponStatus` | gRPC | 启停用优惠券 |
| `CouponStats` | gRPC | 优惠券统计（总数、启用数、累计发放、已过期数） |
| `ClaimCoupon` | gRPC | 用户领券 |
| `ListUserCoupons` | gRPC | 用户优惠券列表（按状态筛选） |
| `ListAvailableCoupons` | gRPC | 可领取优惠券列表（领券中心用，含剩余量/是否已领） |
| `UseCoupon` | gRPC | 使用/核销优惠券 |
| `ReleaseCoupon` | gRPC | 返还优惠券（订单取消/退款时调用） |
| `BatchExpireCoupons` | gRPC | 批量过期优惠券发放记录（定时任务调用） |
| `ListDistRecords` | gRPC | 查询优惠券发放记录列表 |

## 运行方式

### 本地开发

```bash
cd app/service/marketing
go run cmd/marketing/main.go
```

> 配置从 Consul KV 自动加载，需先启动 Consul 并在 `/develop/cyi-app/service/marketing/` 路径下配置。

### 构建部署

```bash
cd app/service/marketing
make build
./bin/marketing
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-marketing`

## 服务发现

服务注册到 Consul，服务名：`marketing`（注册完整名称为 `{SERVICE_ENV}.marketing.service`）

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`
