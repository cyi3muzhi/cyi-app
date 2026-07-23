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
| 权限管理 | Casbin（RBAC） |
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
│   ├── rbac_model.conf  # Casbin RBAC 模型定义
│   ├── rbac_policy.csv  # Casbin 策略配置
│   └── register.yaml    # 注册配置
├── internal/
│   ├── biz/
│   │   ├── auth.go      # 认证业务逻辑（登录、登出、头像、在线管理）
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── category.go  # 商品分类管理业务逻辑
│   │   ├── dashboard.go # 仪表盘业务逻辑
│   │   ├── product.go   # 商品业务逻辑
│   │   ├── role.go      # 角色与权限管理业务逻辑
│   │   └── staff.go     # 员工管理业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── auth.go      # 认证数据访问（gRPC 调用 user service、OSS 上传）
│   │   ├── casbin.go    # Casbin Enforcer 和远程权限校验器
│   │   ├── category.go  # 分类数据访问（gRPC 调用 product service）
│   │   ├── client_conn.go # 下游微服务 gRPC 客户端连接
│   │   ├── dashboard.go # 仪表盘数据访问（mock 数据）
│   │   ├── data.go      # 数据层初始化与依赖注入
│   │   ├── product.go   # 商品数据访问（mock 数据）
│   │   ├── role.go      # 角色数据访问（gRPC 调用 user service）
│   │   └── staff.go     # 员工数据访问（gRPC 调用 user service）
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   └── server.go    # 服务器配置（HTTP/gRPC/服务发现）
│   └── service/
│       ├── auth.go      # 认证服务（登录、登出、个人信息、权限查询）
│       ├── category.go  # 分类服务
│       ├── dashboard.go # 仪表盘服务
│       ├── order.go     # 订单服务（mock 数据）
│       ├── product.go   # 商品服务
│       ├── role.go      # 角色与权限服务
│       ├── service.go   # 服务层依赖注入
│       ├── settings.go  # 系统设置服务（mock 数据）
│       ├── staff.go     # 员工管理服务
│       └── user.go      # 用户管理服务（mock 数据）
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| `auth.api_key` | JWT 签名密钥 | - |
| `auth.service_key` | 下游服务调用密钥（map） | - |
| `auth.white_list` | 无需认证的白名单路径列表 | - |
| `data.redis.addr` | Redis 地址 | - |
| `data.redis.password` | Redis 密码 | - |
| `data.database.driver` | 数据库驱动 | - |
| `data.database.source` | 数据库连接字符串 | - |
| `data.oss.endpoint` | RustFS 地址 | - |
| `data.oss.public_endpoint` | RustFS 公网访问地址 | - |
| `data.oss.access_key` | RustFS 访问密钥 | - |
| `data.oss.secret_key` | RustFS 密钥 | - |
| `registry.consul.address` | Consul 注册中心地址 | - |
| `registry.consul.scheme` | Consul 协议 | - |
| `trace.endpoint` | Jaeger 链路追踪地址 | - |
| `log.level` | 日志级别 | - |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `AuthUseCase` | 管理员认证（登录、登出、头像管理、在线管理、信息更新） |
| `DashboardUseCase` | 仪表盘数据统计、订单趋势、营收分类占比 |
| `ProductUseCase` | 商品管理（CRUD、批量操作） |
| `CategoryUseCase` | 商品分类管理（分类树、CRUD、排序、状态、批量操作） |
| `RoleUseCase` | 角色管理（CRUD、权限管理、权限校验） |
| `StaffUseCase` | 员工管理（列表、创建、更新、状态切换） |

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
                        生成预签名 URL → 保存到数据库 → 返回给前端
```

### 权限控制

服务使用 Casbin RBAC 模型进行权限管理：

- **模型定义**：`configs/rbac_model.conf` 定义了请求、策略、角色、匹配器规则
- **策略配置**：`configs/rbac_policy.csv` 存储静态策略
- **远程校验**：`RemoteEnforcer` 通过 gRPC 调用 user 服务的 `CheckPermission` 接口进行实时权限校验
- **中间件集成**：HTTP 和 gRPC 中间件自动拦截请求，校验操作权限

### 服务接口

#### 认证接口（AdminAuth）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `Login` | POST | `/admin/v1/auth/login` | 管理员登录 |
| `Logout` | POST | `/admin/v1/auth/logout` | 管理员登出 |
| `GetAdminInfo` | GET | `/admin/v1/auth/info` | 获取当前管理员信息 |
| `ChangeAvatar` | POST | `/admin/v1/auth/avatar` | 更换头像（Base64） |
| `UpdateAdminInfo` | PUT | `/admin/v1/auth/info` | 更新基本信息 |
| `GetMyPermissions` | GET | `/admin/v1/auth/permissions` | 获取当前登录用户权限 |
| `GetOnlineAdminIds` | GET | `/admin/v1/auth/online` | 获取在线管理员 ID 列表 |

#### 仪表盘接口（Dashboard）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `GetDashboardStats` | GET | `/admin/v1/dashboard/stats` | 获取统计数据 |
| `GetOrderTrend` | GET | `/admin/v1/dashboard/order-trend` | 获取近14日订单趋势 |
| `GetRevenueCategories` | GET | `/admin/v1/dashboard/revenue-categories` | 获取营收分类占比 |

#### 商品接口（AdminProduct）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListProducts` | GET | `/admin/v1/product/list` | 商品列表 |
| `GetProduct` | GET | `/admin/v1/product/{id}` | 获取商品详情 |
| `CreateProduct` | POST | `/admin/v1/product/create` | 创建商品 |
| `UpdateProduct` | PUT | `/admin/v1/product/update` | 更新商品 |
| `DeleteProduct` | DELETE | `/admin/v1/product/delete` | 删除商品 |
| `BatchOperation` | POST | `/admin/v1/product/batch` | 批量操作商品 |

#### 分类接口（AdminCategory）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListCategoryTree` | GET | `/admin/v1/category/tree` | 获取分类树 |
| `ListChildCategories` | GET | `/admin/v1/category/children` | 获取子分类列表 |
| `GetCategory` | GET | `/admin/v1/category/{id}` | 获取单个分类 |
| `CreateCategory` | POST | `/admin/v1/category/create` | 创建分类 |
| `UpdateCategory` | PUT | `/admin/v1/category/update` | 更新分类 |
| `DeleteCategory` | DELETE | `/admin/v1/category/delete` | 删除分类 |
| `UpdateCategorySort` | PUT | `/admin/v1/category/sort` | 更新排序 |
| `UpdateCategoryStatus` | PUT | `/admin/v1/category/status` | 更新状态 |
| `ToggleCategoryShowcase` | PUT | `/admin/v1/category/showcase` | 切换首页展示 |
| `BatchCategoryOperation` | POST | `/admin/v1/category/batch` | 批量操作分类 |

#### 订单接口（AdminOrder）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListOrder` | GET | `/admin/v1/order/list` | 订单列表 |
| `GetOrder` | GET | `/admin/v1/order/{order_id}` | 订单详情 |
| `ShipOrder` | POST | `/admin/v1/order/ship` | 发货 |
| `CancelOrder` | POST | `/admin/v1/order/cancel` | 取消订单 |
| `ExportOrder` | GET | `/admin/v1/order/export` | 导出订单（投递异步任务到 task 服务消费） |

#### 用户接口（AdminUser）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListUser` | GET | `/admin/v1/user/list` | 用户列表 |
| `GetUser` | GET | `/admin/v1/user/{id}` | 用户详情 |
| `CreateUser` | POST | `/admin/v1/user/create` | 创建用户 |
| `UpdateUser` | PUT | `/admin/v1/user/update` | 更新用户 |
| `DeleteUser` | DELETE | `/admin/v1/user/delete` | 删除用户 |
| `ToggleUserStatus` | PUT | `/admin/v1/user/toggle-status` | 冻结/解冻用户 |
| `GetUserStats` | GET | `/admin/v1/user/stats` | 获取用户统计数据 |

#### 角色接口（AdminRole）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListRoles` | GET | `/admin/v1/role/list` | 获取角色列表 |
| `CreateRole` | POST | `/admin/v1/role/create` | 创建角色 |
| `UpdateRole` | PUT | `/admin/v1/role/update` | 更新角色 |
| `DeleteRole` | DELETE | `/admin/v1/role/delete` | 删除角色 |
| `GetRolePermissions` | GET | `/admin/v1/role/permissions` | 获取角色权限 |
| `UpdateRolePermissions` | PUT | `/admin/v1/role/permissions` | 更新角色权限 |
| `ListAllPermissions` | GET | `/admin/v1/role/permissions/all` | 获取所有权限模块定义 |

#### 员工接口（Staff）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListStaff` | GET | `/admin/v1/staff/list` | 获取员工列表 |
| `CreateStaff` | POST | `/admin/v1/staff/create` | 创建员工 |
| `UpdateStaff` | PUT | `/admin/v1/staff/update` | 更新员工信息 |
| `ToggleStaffStatus` | PUT | `/admin/v1/staff/toggle-status` | 切换员工状态 |
| `GetRoleNameMap` | GET | `/admin/v1/staff/role-map` | 获取角色名称映射 |

#### 设置接口（Settings）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `GetSettings` | GET | `/admin/v1/settings` | 获取系统设置 |
| `UpdateSettings` | PUT | `/admin/v1/settings` | 更新系统设置 |
| `ChangePassword` | PUT | `/admin/v1/settings/password` | 修改密码 |

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
