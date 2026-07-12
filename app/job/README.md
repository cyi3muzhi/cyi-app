# Job 服务

定时任务服务，基于 Asynq 实现分布式定时任务调度和处理。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 任务队列 | Asynq |
| 数据库 | Redis |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/job/
├── cmd/job/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── configs/
│   ├── auth.yaml        # 认证配置
│   ├── config.yaml      # 主配置
│   ├── register.yaml    # 注册配置
│   └── task.yaml        # 定时任务配置
├── internal/
│   ├── biz/
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── job.go       # 任务业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── client_conn.go # 客户端连接
│   │   ├── data.go      # 数据层初始化
│   │   ├── job.go       # 任务数据访问
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── http.go      # HTTP 服务注册
│   │   ├── server.go    # 服务器配置
│   │   └── task.go      # Asynq 任务处理器
│   └── service/
│       ├── job.go       # 任务服务
│       └── service.go   # 服务层依赖注入
└── Makefile             # 构建脚本
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `tasks[].cron` | 定时任务表达式 | - |
| `tasks[].name` | 任务名称 | - |
| `tasks[].options.queue` | 任务队列（critical/default/low） | default |
| `tasks[].options.max_retry` | 最大重试次数 | - |
| `tasks[].options.timeout` | 超时时间 | - |
| `auth.service_key` | 服务密钥映射 | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `data.redis.addr` | Redis 地址 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `JobUseCase` | 定时任务调度、执行 |
| `UserUseCase` | 用户相关定时任务 |

### 任务处理

服务启动时并行运行两个后台协程：

1. **Asynq Server**: 处理队列中的任务
2. **Asynq Scheduler**: 按 cron 表达式调度定时任务

### 任务队列优先级

- `critical`: 关键任务，最高优先级
- `default`: 默认队列，中等优先级
- `low`: 低优先级任务

## 运行方式

### 本地开发

```bash
cd app/job
go run cmd/job/main.go -conf configs/
```

### 构建部署

```bash
cd app/job
make build
./bin/job -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-job`

## 服务发现

服务注册到 Consul，服务名：`job`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`