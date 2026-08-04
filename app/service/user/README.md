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
│   │   ├── biz.go       # 业务层依赖注入
│   │   ├── role.go      # 角色管理业务逻辑
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
│   │   │   ├── casbin_rule.go # Casbin 策略规则模型
│   │   │   ├── role_meta.go # 角色元数据模型
│   │   │   └── user.go  # 用户模型
│   │   ├── role.go      # 角色数据访问
│   │   ├── role_test.go # 角色数据访问测试
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
│       ├── role_admin.go # 角色管理服务（gRPC）
│       ├── service.go   # 服务层依赖注入
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
