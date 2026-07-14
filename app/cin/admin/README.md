# Admin BFF 服务

管理后台 BFF（Backend for Frontend）服务，为管理后台前端提供统一的 HTTP API 接口。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.21+ |
| 数据库 | Redis |
| 对象存储 | RustFS |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 目录结构

```
app/cin/admin/
├── cmd/admin/
│   ├── main.go          # 服务入口
│   ├── wire.go          # 依赖注入配置
│   └── wire_gen.go      # Wire 生成文件
├── configs/
│   ├── auth.yaml        # 认证配置（JWT 密钥）
│   ├── config.yaml      # 主配置
│   └── register.yaml    # 注册配置
├── internal/
│   ├── biz/
│   │   ├── auth.go      # 认证业务逻辑（登录、登出、头像）
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── dashboard.go # 仪表盘业务逻辑
│   │   └── product.go   # 商品业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── auth.go      # 认证数据访问（gRPC 调用、OSS 上传）
│   │   ├── data.go      # 数据层初始化
│   │   ├── dashboard.go # 仪表盘数据访问
│   │   └── product.go   # 商品数据访问
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── auth.go      # 认证服务（HTTP/gRPC Handler）
│       ├── dashboard.go # 仪表盘服务
│       ├── product.go   # 商品服务
│       └── service.go   # 服务层依赖注入
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.api_key` | JWT 签名密钥 | - |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.oss.endpoint` | RustFS 地址 | - |
| `data.oss.access_key` | RustFS 访问密钥 | - |
| `data.oss.secret_key` | RustFS 密钥 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `AuthUseCase` | 管理员认证（登录、登出、头像管理） |
| `DashboardUseCase` | 仪表盘数据统计 |
| `ProductUseCase` | 商品管理 |

### 认证流程

```
前端请求登录 → admin BFF → 调用 user service gRPC 验证账号密码
                          ↓
                    user service 返回管理员信息
                          ↓
                    admin BFF 生成独立 JWT token（用自己的密钥签名）
                          ↓
                    admin BFF 将 token 保存到 Redis（单点登录）
                          ↓
                    返回管理员信息和 token 给前端
```

### Token 管理

- **JWT 签名**：使用 `auth.api_key` 签名，独立于 user service
- **单点登录**：Redis Key 格式 `admin:token:{adminID}`，新登录覆盖旧 token
- **登出**：从 Redis 移除 token，下次请求验证失败返回 401
- **过期**：Token 默认 24 小时过期，Redis 自动过期

### 头像上传

```
前端选择图片 → 转为 Base64 → POST /admin/v1/auth/avatar
                              ↓
                        admin BFF 解码 Base64 → 上传到 RustFS OSS
                              ↓
                        生成永久 URL → 保存到数据库 → 返回给前端
```

### 服务接口

#### 认证接口（AdminAuth）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `Login` | POST | `/admin/v1/auth/login` | 管理员登录 |
| `Logout` | POST | `/admin/v1/auth/logout` | 管理员登出 |
| `GetAdminInfo` | GET | `/admin/v1/auth/info` | 获取当前管理员信息 |
| `ChangeAvatar` | POST | `/admin/v1/auth/avatar` | 更换头像（Base64） |

#### 仪表盘接口（Dashboard）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `GetDashboardStats` | GET | `/admin/v1/dashboard/stats` | 获取统计数据 |

#### 商品接口（Product）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListProducts` | GET | `/admin/v1/product/list` | 商品列表 |
| `GetProduct` | GET | `/admin/v1/product/{id}` | 获取商品详情 |
| `CreateProduct` | POST | `/admin/v1/product/create` | 创建商品 |
| `UpdateProduct` | PUT | `/admin/v1/product/update` | 更新商品 |
| `DeleteProduct` | DELETE | `/admin/v1/product/delete` | 删除商品 |

## 运行方式

### 本地开发

```bash
cd app/cin/admin
go run cmd/admin/main.go -conf configs/
```

### 构建部署

```bash
cd app/cin/admin
make build
./bin/admin -conf configs/
```

### Docker 部署

服务通过 Docker Compose 部署，服务名：`cin-admin`

## 服务发现

服务注册到 Consul，服务名：`cin-admin`

## 服务间调用

| 调用方 | 被调用方 | 协议 |
|--------|----------|------|
| admin BFF | User Service | gRPC |
| admin BFF | Product Service | gRPC |

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`
