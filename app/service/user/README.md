# User 服务

用户管理服务，提供用户认证、注册、信息管理、管理员管理以及基于 Casbin 的角色权限控制功能。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Kratos v2 |
| 语言 | Go 1.25 |
| 数据库 | MySQL、Redis、OSS |
| 权限管理 | Casbin |
| 服务发现 | Consul |
| 分布式追踪 | Jaeger (OpenTelemetry) |
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
│   │   ├── address.go   # 收货地址业务逻辑
│   │   ├── admin_user.go # 管理员业务逻辑
│   │   ├── auth.go      # 认证业务逻辑（注册/登录）
│   │   ├── banner.go    # Banner 业务逻辑
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── errors.go    # 业务错误定义
│   │   ├── operation_log.go # 操作日志业务逻辑（写入、查询、统计、操作人列表）
│   │   ├── role.go      # 角色管理业务逻辑
│   │   ├── settings.go  # 店铺设置业务逻辑
│   │   └── user.go      # 用户业务逻辑
│   ├── conf/
│   │   ├── conf.pb.go   # 配置结构体
│   │   └── conf.proto   # 配置定义
│   ├── data/
│   │   ├── address.go   # 收货地址数据访问
│   │   ├── admin_user.go # 管理员数据访问
│   │   ├── admin_user_test.go # 管理员数据访问测试
│   │   ├── casbin.go    # Casbin 适配器（GORM 存储）
│   │   ├── curd/
│   │   │   ├── address.gen.go   # 地址 CURD
│   │   │   ├── admin_user.gen.go # 管理员 CURD
│   │   │   ├── casbin_rule.gen.go # Casbin 策略 CURD
│   │   │   ├── gen.go   # 代码生成工具
│   │   │   ├── role_meta.gen.go # 角色元数据 CURD
│   │   │   └── user.gen.go   # 用户 CURD
│   │   ├── data.go      # 数据层初始化
│   │   ├── models/
│   │   │   ├── admin_user.go # 管理员模型
│   │   │   ├── banner.go # Banner 模型
│   │   │   ├── casbin_rule.go # Casbin 策略规则模型
│   │   │   ├── operation_log.go # 操作日志模型
│   │   │   ├── role_meta.go # 角色元数据模型
│   │   │   ├── settings.go # 店铺设置模型
│   │   │   └── user.go  # 用户模型
│   │   ├── address.go   # 收货地址数据访问
│   │   ├── admin_user.go # 管理员数据访问
│   │   ├── admin_user_test.go # 管理员数据访问测试
│   │   ├── banner.go    # Banner 数据访问
│   │   ├── casbin.go    # Casbin 适配器（GORM 存储）
│   │   ├── operation_log.go # 操作日志数据访问（GORM 读写 + 头像关联查询）
│   │   ├── operation_log_nats.go # 操作日志 NATS 消费者（订阅落库）
│   │   ├── role.go      # 角色数据访问
│   │   ├── role_test.go # 角色数据访问测试
│   │   ├── settings.go  # 店铺设置数据访问
│   │   └── user.go      # 用户数据访问
│   ├── pkg/
│   │   └── encryption/
│   │       ├── encryption.go      # 加密工具
│   │       └── encryption_test.go # 加密测试
│   ├── server/
│   │   ├── grpc.go      # gRPC 服务注册
│   │   └── server.go    # 服务器配置
│   └── service/
│       ├── admin_user.go # 管理员服务（gRPC）
│       ├── banner.go    # Banner 服务（gRPC）
│       ├── operation_log.go # 操作日志服务（gRPC，供 admin 查询）
│       ├── role_admin.go # 角色管理服务（gRPC）
│       ├── service.go   # 服务层依赖注入
│       ├── settings.go  # 店铺设置服务（gRPC）
│       └── user.go      # 用户服务
├── Makefile             # 构建脚本
└── openapi.yaml         # OpenAPI 文档
```

## 配置说明

主要配置项：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| **Server** | | |
| `server.http.addr` | HTTP 监听地址 | - |
| `server.grpc.addr` | gRPC 监听地址 | - |
| **Data** | | |
| `data.database.driver` | 数据库驱动 | mysql |
| `data.database.source` | MySQL 连接字符串 | - |
| `data.redis.network` | Redis 网络类型 | tcp |
| `data.redis.addr` | Redis 地址 | - |
| `data.redis.password` | Redis 密码 | - |
| `data.redis.read_timeout` | Redis 读取超时 | - |
| `data.redis.write_timeout` | Redis 写入超时 | - |
| `data.oss.endpoint` | OSS 地址 | - |
| `data.oss.access_key` | OSS 访问密钥 | - |
| `data.oss.secret_key` | OSS 秘密密钥 | - |
| **NATS** | | |
| `NATS_URL`（环境变量） | NATS 地址，操作日志订阅用（Docker 环境配 `nats://nats:4222`） | `nats://localhost:4222` |
| **Auth** | | |
| `auth.api_key` | API 密钥（用于 JWT 签名） | - |
| **Registry** | | |
| `registry.consul.address` | Consul 地址 | - |
| `registry.consul.scheme` | Consul 协议（http/https） | http |
| `registry.consul.enable_health_check` | 是否启用健康检查 | false |
| `registry.consul.check_interval` | 健康检查间隔（秒） | - |
| **Trace** | | |
| `trace.endpoint` | Jaeger 采集端点 | - |
| **Log** | | |
| `log.level` | 日志级别 | info |
| `log.enable_console` | 是否启用控制台输出 | true |
| `log.enable_file` | 是否启用文件输出 | false |
| `log.log_dir` | 日志目录 | - |
| `log.time_format` | 时间格式 | - |
| `log.enable_color` | 是否启用颜色 | false |

## 核心业务逻辑

### UseCase 列表

| UseCase | 职责 |
|---------|------|
| `UserUseCase` | 用户管理（注册、登录、信息管理、密码验证、修改密码） |
| `AddressUseCase` | 收货地址管理（增删改查、设置默认地址） |
| `AdminUserUseCase` | 管理员管理（登录验证、CRUD、头像管理、单点登录） |
| `RoleUseCase` | 角色管理（角色 CRUD、权限分配、权限校验） |
| `BannerUseCase` | Banner 管理（列表、创建、更新、删除、上下架） |
| `SettingsUseCase` | 店铺设置（获取、更新） |
| `OperationLogUseCase` | 操作日志（NATS 消费写入、分页查询、统计、操作人列表） |
| `MessageUseCase` | 消息中心（发送通知、消息列表、已读管理、未读数统计） |

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

### 权限控制

基于 Casbin（RBAC）实现管理员角色权限控制：
- **角色**：通过 `role_meta` 表存储角色元数据（名称、描述、标签颜色）
- **策略**：通过 `casbin_rule` 表存储 Casbin 策略规则
- **权限模型**：支持 `sub, obj, act` 三元组，使用 `keyMatch` 匹配器
- **内置角色**：`super_admin`（超级管理员）拥有所有权限，不可删除/修改

### 加密工具

提供密码加密和验证功能，位于 `pkg/encryption/encryption.go`

### 操作日志消费

admin 服务将管理端写操作发布到 NATS `admin.operation` 主题，user 服务订阅并落库：

```
admin 操作日志中间件 → NATS 发布（admin.operation 主题）
      ↓
user 服务 OperationLogConsumer.Start() 订阅
      ↓ 反序列化事件 → OperationLogUseCase.Create
写入 admin_operation_log 表
```

- 查询接口通过 gRPC 暴露给 admin 服务（`ListOperationLog` / `CountOperationLog` / `ListOperationLogOperators`）
- 查询时按 `admin_user_id` 批量关联管理员头像
- NATS 连接失败时降级跳过订阅（`NewNATSConn` 不 panic，最多重试 5 次），不影响服务启动

### 服务接口

#### 用户服务（User）

| 接口 | 方法 | 说明 |
|------|------|------|
| `Register` | gRPC | 用户注册 |
| `Login` | gRPC | 用户登录（返回 JWT Token） |
| `CreateUser` | gRPC | 创建用户 |
| `GetUser` | gRPC | 查询用户信息 |
| `UpdateUser` | gRPC | 更新用户信息 |
| `DeleteUser` | gRPC | 删除用户 |
| `ListUser` | gRPC | 用户列表 |
| `CheckUserExist` | gRPC | 检查用户是否存在 |
| `GetUserByUsername` | gRPC | 根据用户名查询用户 |
| `VerifyPassword` | gRPC | 校验用户密码 |
| `UserCount` | gRPC | 统计用户数量 |
| `ChangePassword` | gRPC | 修改密码（校验旧密码后更新） |
| `ListAddress` | gRPC | 查询用户收货地址列表 |
| `CreateAddress` | gRPC | 新增收货地址 |
| `UpdateAddress` | gRPC | 更新收货地址 |
| `DeleteAddress` | gRPC | 删除收货地址 |
| `SetDefaultAddress` | gRPC | 设置默认收货地址 |

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

#### 角色管理服务（AdminRole）

| 接口 | 方法 | 说明 |
|------|------|------|
| `ListRoles` | gRPC | 获取角色列表 |
| `CreateRole` | gRPC | 创建角色 |
| `UpdateRole` | gRPC | 更新角色 |
| `DeleteRole` | gRPC | 删除角色 |
| `GetRolePermissions` | gRPC | 获取角色权限 |
| `UpdateRolePermissions` | gRPC | 更新角色权限 |
| `ListAllPermissions` | gRPC | 获取所有权限模块定义 |
| `CheckPermission` | gRPC | 校验角色权限 |
| `HasPolicyForOperation` | gRPC | 检查是否有策略引用指定操作 |

#### Banner 服务（Banner）

| 接口 | 方法 | 说明 |
|------|------|------|
| `ListBanners` | gRPC | Banner 列表 |
| `CreateBanner` | gRPC | 创建 Banner |
| `UpdateBanner` | gRPC | 更新 Banner |
| `DeleteBanner` | gRPC | 删除 Banner |
| `ToggleBannerStatus` | gRPC | 上下架 Banner |

#### 店铺设置服务（Settings）

| 接口 | 方法 | 说明 |
|------|------|------|
| `GetSettings` | gRPC | 获取店铺设置 |
| `UpdateSettings` | gRPC | 更新店铺设置 |

#### 操作日志服务（OperationLog）

| 接口 | 方法 | 说明 |
|------|------|------|
| `ListOperationLog` | gRPC | 操作日志分页查询（日期/操作人/模块/关键词筛选） |
| `CountOperationLog` | gRPC | 操作日志统计（今日/近7天/近30天/累计） |
| `ListOperationLogOperators` | gRPC | 操作人列表（按姓名去重） |

#### 消息中心服务（Message）

| 接口 | 方法 | 说明 |
|------|------|------|
| `CreateMessage` | gRPC | 发送消息（全体用户/指定用户，类型：订单/系统/营销） |
| `AdminListMessage` | gRPC | 管理后台消息列表（分页 + 类型/状态/关键字筛选） |
| `DeleteMessage` | gRPC | 下线/删除消息（软删除） |
| `ListMyMessage` | gRPC | 我的消息列表（分页 + 类型筛选，含已读状态） |
| `UnreadMessageCount` | gRPC | 我的未读消息数 |
| `MarkMessageRead` | gRPC | 标记单条消息已读 |
| `MarkAllMessageRead` | gRPC | 全部标记已读 |

- 数据表：`message`（消息内容）+ `user_message`（用户已读状态，唯一索引 user_id+message_id）
- 可见性：`target_type=0` 全体消息对所有人可见；`target_type=1` 仅对 `target_user_ids` 中指定用户可见
- 未读判定：用户消息表中无已读记录（或未读）即为未读，标记已读采用 UPSERT 幂等写入

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
