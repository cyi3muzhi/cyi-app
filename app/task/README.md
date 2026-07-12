# Task 服务

异步任务服务，基于 Asynq 实现分布式任务队列处理。

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
│   │   ├── task.go      # 任务业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── client_conn.go # 客户端连接
│   │   ├── data.go      # 数据层初始化
│   │   ├── task.go      # 任务数据访问
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── http.go      # HTTP 服务注册
│   │   ├── server.go    # 服务器配置
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
| `auth.service_key` | 服务密钥映射 | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `data.redis.addr` | Redis 地址 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `TaskUseCase` | 异步任务调度、执行 |
| `UserUseCase` | 用户相关异步任务 |

### 任务处理

服务启动时启动 Asynq Server 处理队列中的任务：

```go
go func() {
    if err := server.Run(mux); err != nil {
        // 处理错误
    }
}()
```

### 任务队列优先级

- `critical`: 关键任务，最高优先级
- `default`: 默认队列，中等优先级
- `low`: 低优先级任务

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