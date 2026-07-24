# Blades 服务

AI 智能助手服务，基于 DeepSeek API 提供聊天对话能力。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL、Redis、OSS |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |
| AI 接口 | DeepSeek API |

## 目录结构

```
app/blades/
├── cmd/blades/
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
│   │   └── blades.go    # Blades 业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── blades.go    # 数据访问层
│   │   └── data.go      # 数据层初始化
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── blades.go    # gRPC 服务实现
│       └── service.go   # 服务层依赖注入
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `blades.api_key` | DeepSeek API 密钥 | - |
| `blades.model` | 模型名称 | deepseek-chat |
| `blades.max_tokens` | 最大令牌数 | - |
| `blades.temperature` | 温度参数 | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |

## 核心业务逻辑

### BladesUseCase

处理 AI 聊天请求：

- `Chat(ctx, message)`: 接收用户消息，调用 DeepSeek API 返回响应

### 服务接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `Chat` | gRPC/HTTP | 发送消息并获取 AI 响应 |

## 运行方式

### 本地开发

```bash
cd app/blades
go run cmd/blades/main.go -conf configs/
```

### 构建部署

```bash
cd app/blades
make build
./bin/blades -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-blades`

## 服务发现

服务注册到 Consul，服务名：`blades`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`