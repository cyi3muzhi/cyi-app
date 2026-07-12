# NATS 服务

NATS 消息消费者服务，订阅 NATS 主题并处理消息。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 消息队列 | NATS |
| 数据库 | MySQL、Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |

## 目录结构

```
app/nats/
├── cmd/nats/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   └── order.go     # 订单消息处理
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── consumers/
│   │   ├── consumer.go  # NATS 消费者实现
│   │   ├── consumer_test.go # 消费者测试
│   │   └── init.go      # 消费者初始化
│   ├── data/
│   │   ├── data.go      # 数据层初始化
│   │   └── order.go     # 订单数据访问
│   ├── server/
│   │   └── server.go    # NATS 连接配置
│   └── service/
│       ├── order.go     # 订单服务
│       └── service.go   # 服务层依赖注入
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `nats.url` | NATS 服务器地址 | nats://localhost:4222 |
| `auth.api_key` | API 密钥 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |

## 核心业务逻辑

### Consumer

NATS 消费者实现：

- `Start()`: 启动消费者协程
- `Subscribe(subject, handler)`: 订阅主题并注册处理函数

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `OrderUseCase` | 处理订单相关消息 |

### 消息处理流程

```
NATS Publisher → NATS Server → NATS Consumer → OrderUseCase → Database
```

## 运行方式

### 本地开发

```bash
cd app/nats
go run cmd/nats/main.go -conf configs/
```

### 构建部署

```bash
cd app/nats
make build
./bin/nats -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-nats`

## 服务发现

服务注册到 Consul，服务名：`nats`

## 测试

消费者测试覆盖：

| 测试 | 说明 |
|------|------|
| `TestNATS_Connect` | 测试 NATS 连接 |
| `TestNATS_PublishSubscribe` | 测试发布订阅 |
| `TestNATS_QueueGroup` | 测试队列组 |
| `TestConsumer_FullCycle` | 测试完整消费链路 |

```bash
cd app/nats
go test ./internal/consumers/... -v
```