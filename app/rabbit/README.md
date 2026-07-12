# RabbitMQ 服务

RabbitMQ 消息消费者服务，订阅 RabbitMQ 队列并处理消息。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 消息队列 | RabbitMQ |
| 数据库 | MySQL、Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |

## 目录结构

```
app/rabbit/
├── cmd/rabbit/
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
│   │   ├── init.go      # 消费者初始化
│   │   ├── simple.go    # 简单队列消费者
│   │   └── topic.go     # 主题队列消费者
│   ├── data/
│   │   ├── data.go      # 数据层初始化
│   │   └── order.go     # 订单数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── order.go     # 订单服务
│       └── service.go   # 服务层依赖注入
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.api_key` | API 密钥 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |

## 核心业务逻辑

### 消费者类型

| 消费者 | 类型 | 说明 |
|--------|------|------|
| `SimpleConsumer` | 简单队列 | 点对点消息传递 |
| `TopicConsumer` | 主题队列 | 发布订阅模式 |

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `OrderUseCase` | 处理订单相关消息 |

### 消息处理流程

```
RabbitMQ Producer → Exchange → Queue → Consumer → OrderUseCase → Database
```

## 运行方式

### 本地开发

```bash
cd app/rabbit
go run cmd/rabbit/main.go -conf configs/
```

### 构建部署

```bash
cd app/rabbit
make build
./bin/rabbit -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-rabbit`

## 服务发现

服务注册到 Consul，服务名：`rabbit`

## RabbitMQ 配置

连接地址：`amqp://admin:admin@localhost:5672/`

管理界面：`http://localhost:15672`（用户名/密码：admin/admin）