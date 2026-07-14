# User 服务

用户管理服务，提供用户认证、注册和信息管理功能。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 数据库 | MySQL、Redis、OSS |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/service/user/
├── cmd/user/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── configs/
│   ├── auth.yaml        # 认证配置
│   ├── config.yaml      # 主配置
│   └── register.yaml    # 注册配置
├── internal/
│   ├── biz/
│   │   ├── admin_user.go # 管理员业务逻辑
│   │   ├── auth.go      # 认证业务逻辑
│   │   ├── biz.go       # 业务层依赖注入
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── admin_user.go # 管理员数据访问
│   │   ├── curd/
│   │   │   ├── address.gen.go # 地址 CURD
│   │   │   ├── gen.go   # 代码生成工具
│   │   │   └── user.gen.go   # 用户 CURD
│   │   ├── data.go      # 数据层初始化
│   │   ├── models/
│   │   │   ├── admin_user.go # 管理员模型
│   │   │   └── user.go  # 用户模型
│   │   ├── pkg/
│   │   │   └── encryption/
│   │   │       ├── encryption.go      # 加密工具
│   │   │       └── encryption_test.go # 加密测试
│   │   └── user.go      # 用户数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── admin_user.go # 管理员服务（gRPC）
│       ├── service.go   # 服务层依赖注入
│       └── user.go      # 用户服务
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.api_key` | API 密钥 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.oss.endpoint` | OSS 地址 | - |
| `data.oss.access_key` | OSS 访问密钥 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `UserUseCase` | 用户管理（注册、登录、信息管理） |
| `AdminUserUseCase` | 管理员管理（登录验证、CRUD、头像管理） |

### 认证流程

**普通用户认证**：
```
用户请求 → JWT Token 验证 → 用户信息查询 → 权限校验 → 业务处理
```

**管理员认证**：
```
admin BFF 请求 → 调用 Login gRPC 验证账号密码 → 返回管理员信息
（管理员 token 由 admin BFF 生成和管理）
```

### 加密工具

提供密码加密和验证功能，位于 `pkg/encryption/encryption.go`

### 服务接口

#### 用户服务（User）

| 接口 | 方法 | 说明 |
|------|------|------|
| `Register` | gRPC | 用户注册 |
| `Login` | gRPC | 用户登录 |
| `GetUser` | gRPC | 查询用户信息 |
| `UpdateUser` | gRPC | 更新用户信息 |
| `DeleteUser` | gRPC | 删除用户 |
| `GetAddresses` | gRPC | 获取地址列表 |
| `AddAddress` | gRPC | 添加地址 |
| `UpdateAddress` | gRPC | 更新地址 |
| `DeleteAddress` | gRPC | 删除地址 |

#### 管理员服务（AdminUser）

| 接口 | 方法 | 说明 |
|------|------|------|
| `Login` | gRPC | 管理员登录验证（返回管理员信息，token 由 admin BFF 生成） |
| `CreateAdminUser` | gRPC | 创建管理员 |
| `UpdateAdminUser` | gRPC | 更新管理员 |
| `GetAdminUser` | gRPC | 获取管理员信息 |
| `ListAdminUser` | gRPC | 管理员列表 |
| `ToggleAdminUserStatus` | gRPC | 冻结/解冻管理员 |
| `UpdateAdminAvatar` | gRPC | 更新管理员头像 |

## 运行方式

### 本地开发

```bash
cd app/service/user
go run cmd/user/main.go -conf configs/
```

### 构建部署

```bash
cd app/service/user
make build
./bin/user -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`service-user`

## 服务发现

服务注册到 Consul，服务名：`user`

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`