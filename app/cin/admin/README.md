# Admin BFF 服务

管理后台 BFF（Backend for Frontend）服务，为管理后台前端提供统一的 HTTP API 接口。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
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
│   ├── tracer.go        # Jaeger 追踪初始化
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
│   │   ├── agent.go       # AI 智能助手业务逻辑（聊天、生成文案、知识库）
│   │   ├── auth.go       # 认证业务逻辑（登录、登出、头像、在线管理）
│   │   ├── biz.go        # 业务层依赖注入
│   │   ├── category.go   # 商品分类管理业务逻辑
│   │   ├── coupon.go     # 优惠券管理业务逻辑
│   │   ├── dashboard.go  # 仪表盘业务逻辑
│   │   ├── export.go     # 订单导出业务逻辑（投递异步任务、查询状态、生成下载链接）
│   │   ├── operation_log.go # 操作日志业务逻辑（查询、统计、操作人列表）
│   │   ├── order.go      # 订单管理业务逻辑（列表、详情、发货、取消、统计）
│   │   ├── product.go    # 商品业务逻辑
│   │   ├── role.go       # 角色与权限管理业务逻辑
│   │   ├── settings.go   # 系统设置业务逻辑
│   │   ├── staff.go      # 员工管理业务逻辑
│   │   ├── staff_test.go # 员工管理单元测试
│   │   └── user.go       # 用户管理业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── asynq.go       # Asynq 客户端初始化（投递导出任务）
│   │   ├── auth.go        # 认证数据访问（gRPC 调用 user service、OSS 上传）
│   │   ├── banner.go      # Banner 数据访问（gRPC 调用 user service）
│   │   ├── casbin.go      # Casbin Enforcer 和远程权限校验器
│   │   ├── category.go    # 分类数据访问（gRPC 调用 product service）
│   │   ├── client_conn.go # 下游微服务 gRPC 客户端连接
│   │   ├── coupon.go      # 优惠券数据访问（gRPC 调用 marketing service）
│   │   ├── dashboard.go   # 仪表盘数据访问
│   │   ├── data.go        # 数据层初始化与依赖注入
│   │   ├── export.go      # 导出数据访问（Asynq 投递、Redis 状态读写、OSS 下载链接缓存）
│   │   ├── operation_log.go # 操作日志数据访问（gRPC 查询 user service、NATS 发布器）
│   │   ├── order.go       # 订单数据访问（gRPC 调用 order service）
│   │   ├── oss_helper.go  # OSS 公开 URL 解析辅助
│   │   ├── product.go     # 商品数据访问（gRPC 调用 product service）
│   │   ├── role.go        # 角色数据访问（gRPC 调用 user service）
│   │   ├── settings.go    # 系统设置数据访问（gRPC 调用 user service）
│   │   ├── staff.go       # 员工数据访问（gRPC 调用 user service）
│   │   ├── upload.go      # 文件上传数据访问（OSS）
│   │   ├── upload_test.go # 文件上传单元测试
│   │   └── user.go        # 用户管理数据访问（gRPC 调用 user service）
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   ├── http.go      # HTTP 服务注册
│   │   └── server.go    # 服务器配置（HTTP/gRPC/服务发现）
│   └── service/
│       ├── agent.go      # AI 智能助手服务（聊天、生成文案、快捷提问、知识库）
│       ├── auth.go      # 认证服务（登录、登出、个人信息、权限查询）
│       ├── banner.go    # Banner 管理服务
│       ├── category.go  # 分类服务
│       ├── coupon.go    # 优惠券管理服务
│       ├── dashboard.go # 仪表盘服务
│       ├── export.go    # 订单导出服务（投递任务、查询状态）
│       ├── logs.go      # 操作日志服务（列表、统计、操作人列表）
│       ├── order.go     # 订单服务（列表、详情、发货、取消、统计）
│       ├── product.go   # 商品服务
│       ├── role.go      # 角色与权限服务
│       ├── service.go   # 服务层依赖注入
│       ├── settings.go  # 系统设置服务
│       ├── staff.go     # 员工管理服务
│       ├── upload.go    # 文件上传服务
│       └── user.go      # 用户管理服务
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
| `NATS_URL`（环境变量） | NATS 地址，操作日志发布用（Docker 环境配 `nats://nats:4222`） | `nats://localhost:4222` |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `AuthUseCase` | 管理员认证（登录、登出、头像管理、在线管理、信息更新） |
| `AgentUseCase` | AI 智能助手（聊天、生成商品描述、Banner 文案、分类 SEO、运营摘要、快捷提问、知识库管理） |
| `DashboardUseCase` | 仪表盘数据统计、订单趋势、营收分类占比 |
| `OrderUseCase` | 订单管理（列表、详情、发货、取消、统计） |
| `ExportUseCase` | 订单导出（投递异步任务到 task 服务、查询任务状态、生成下载链接） |
| `ProductUseCase` | 商品管理（CRUD、批量操作） |
| `CategoryUseCase` | 商品分类管理（分类树、CRUD、排序、状态、批量操作） |
| `RoleUseCase` | 角色管理（CRUD、权限管理、权限校验） |
| `StaffUseCase` | 员工管理（列表、创建、更新、状态切换） |
| `CouponUseCase` | 优惠券管理（CRUD、上下架、统计、发放记录） |
| `UserUseCase` | 用户管理（列表、详情、创建、更新、冻结/解冻、统计） |
| `BannerUseCase` | Banner 管理（列表、创建、更新、删除、上下架） |
| `SettingsUseCase` | 系统设置（获取、更新、修改密码） |
| `OperationLogUseCase` | 操作日志（分页查询、统计、操作人列表） |

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
| `ListOrder` | GET | `/admin/v1/orders` | 订单列表 |
| `OrderStats` | GET | `/admin/v1/orders/stats` | 订单统计（顶部统计卡片） |
| `ExportOrder` | GET | `/admin/v1/orders/export` | 导出订单（投递异步任务到 task 服务消费，返回 task_id） |
| `ExportTaskStatus` | GET | `/admin/v1/orders/export/{task_id}` | 查询导出任务状态（completed 时返回下载链接） |
| `GetOrder` | GET | `/admin/v1/orders/{order_id}` | 订单详情 |
| `ShipOrder` | POST | `/admin/v1/orders/{order_id}/ship` | 发货 |
| `CancelOrder` | POST | `/admin/v1/orders/{order_id}/cancel` | 取消订单 |

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

#### 优惠券接口（AdminCoupon）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListCoupon` | GET | `/admin/v1/coupons` | 优惠券列表（分页/状态/关键词/日期筛选） |
| `GetCouponStats` | GET | `/admin/v1/coupons/stats` | 优惠券统计（总数/进行中/已发放/已过期） |
| `GetCoupon` | GET | `/admin/v1/coupons/{coupon_id}` | 优惠券详情（含发放记录） |
| `CreateCoupon` | POST | `/admin/v1/coupons` | 创建优惠券 |
| `UpdateCoupon` | PUT | `/admin/v1/coupons/{coupon_id}` | 更新优惠券 |
| `DeleteCoupon` | DELETE | `/admin/v1/coupons/{coupon_id}` | 删除优惠券 |
| `ToggleCouponStatus` | POST | `/admin/v1/coupons/{coupon_id}/toggle` | 作废/恢复优惠券 |
| `ListCouponRecords` | GET | `/admin/v1/coupons/{coupon_id}/records` | 发放记录（含用户昵称/头像、关联订单） |

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

#### 操作日志接口（AdminLogs）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListOperationLogs` | GET | `/admin/v1/logs` | 操作日志分页查询（支持日期/操作人/模块/关键词筛选） |
| `CountOperationLogs` | GET | `/admin/v1/logs/stats` | 操作日志统计（今日/本周/本月/累计） |
| `ListOperationLogOperators` | GET | `/admin/v1/logs/operators` | 操作人列表（筛选下拉） |

#### 智能助手接口（Agent）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `Chat` | POST | `/admin/v1/agent/chat` | 非流式聊天（支持工具查询/RAG） |
| `GenerateDescription` | POST | `/admin/v1/agent/generate-description` | AI 生成商品描述 |
| `ExpandDescription` | POST | `/admin/v1/agent/expand-description` | AI 扩写商品描述 |
| `GenerateBannerCopy` | POST | `/admin/v1/agent/generate-banner` | AI 生成 Banner 文案 |
| `GenerateCategorySEO` | POST | `/admin/v1/agent/generate-seo` | AI 生成分类 SEO |
| `GenerateSummary` | POST | `/admin/v1/agent/generate-summary` | AI 生成运营摘要 |
| `SuggestQuestions` | POST | `/admin/v1/agent/suggest-questions` | AI 推荐快捷提问 |
| `GetChatHistory` | GET | `/admin/v1/agent/chat/history` | 获取聊天历史 |
| `ClearChatHistory` | POST | `/admin/v1/agent/chat/clear` | 清除聊天历史 |
| `ListKnowledge` | GET | `/admin/v1/agent/knowledge/list` | 列出知识库文档 |
| `DeleteKnowledge` | POST | `/admin/v1/agent/knowledge/delete` | 删除知识库文档 |
| `SearchKnowledge` | POST | `/admin/v1/agent/knowledge/search` | 搜索知识库 |

#### Banner 接口（AdminBanner）

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| `ListBanners` | GET | `/admin/v1/banners` | Banner 列表 |
| `CreateBanner` | POST | `/admin/v1/banners` | 创建 Banner |
| `UpdateBanner` | PUT | `/admin/v1/banners/{id}` | 更新 Banner |
| `DeleteBanner` | DELETE | `/admin/v1/banners/{id}` | 删除 Banner |
| `ToggleBannerStatus` | POST | `/admin/v1/banners/{id}/toggle` | 上下架 Banner |

### 操作日志链路

管理员的写操作（非 GET）由 HTTP 中间件自动记录，通过 NATS 异步落库：

```
管理员写操作请求
      ↓
admin HTTP 操作日志中间件（认证之后，从 JWT 取操作人）
      ↓ 发布事件到 NATS admin.operation 主题
user 服务订阅 → 写入 admin_operation_log 表
      ↓
管理端查询 GET /admin/v1/logs（admin → user 服务 gRPC 读库返回）
```

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
| admin BFF | Order Service | gRPC |
| admin BFF | Redis（Asynq） | 投递导出任务到 `export` 队列，由 task 服务消费 |
| admin BFF | NATS | 发布操作日志事件（admin.operation 主题） |

### 订单导出流程（异步）

admin 服务仅作为**投递端**，不处理导出任务本身（任务处理由 task 服务消费完成）：

```
前端 GET /admin/v1/orders/export
      ↓
admin ExportUseCase.SubmitExport
  1. 生成 task_id（UUID）
  2. 写 Redis 任务状态 = pending（key: admin:export:{task_id}，TTL 24h）
  3. 通过 asynq.Client 投递任务到 export 队列
  4. 立即返回 task_id 给前端
      ↓
前端轮询 GET /admin/v1/orders/export/{task_id}
      ↓
admin ExportUseCase.GetTaskStatus
  - 从 Redis 读取任务状态
  - 若 completed：根据 object_key 生成 presigned 下载 URL（带 6h Redis 缓存）
  - 返回状态 + 下载链接
```

> 任务状态结构与 Redis key 前缀由共享包 `pkg/exporttask` 定义，admin（投递端）与 task（消费端）共同引用以保证一致。

## 监控指标

服务暴露 Prometheus 指标：
- 请求计数：`kratos_server_requests_total`
- 请求耗时：`kratos_server_request_duration_seconds`
