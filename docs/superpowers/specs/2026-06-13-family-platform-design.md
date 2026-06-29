# 家庭信息管理平台 - 设计文档

> 日期：2026-06-13
> 状态：待用户审核
> 原型地址：`web/family-app/prototype.html`

---

## 1. 项目概述

### 1.1 项目背景

构建一个面向小家庭的私有信息管理平台，支持家庭成员之间安全地存储和共享照片、文件、档案信息，并提供实时聊天和日程提醒功能。平台设计为可扩展架构，初期支持 3-5 人小家庭使用，后期可扩展至大家族场景。

### 1.2 核心目标

- **私密安全**：细粒度权限控制，每个资源可单独设置可见性
- **家庭协作**：成员间实时沟通、共享文件和照片
- **温暖易用**：温馨的 UI 设计，家庭成员（包括长辈）都能轻松使用
- **AI 增强**：集成 AI Agent 助手，提供智能辅助

### 1.3 技术选型

| 层面 | 技术方案 | 说明 |
|------|----------|------|
| 后端框架 | Go-Kratos v2 | 复用现有微服务框架 |
| 前端框架 | React 18 + TypeScript | 组件丰富，生态成熟 |
| UI 组件库 | Ant Design 5 | 适合管理类应用 |
| 状态管理 | Zustand | 轻量高效 |
| 实时通信 | WebSocket | 聊天功能 |
| 构建工具 | Vite | 快速开发体验 |
| 数据库 | MySQL | 复用现有基础设施 |
| 缓存 | Redis | 复用现有基础设施 |
| 对象存储 | RustFS/MinIO | 复用现有基础设施 |
| 消息队列 | RabbitMQ | 复用现有基础设施 |
| 服务发现 | Consul | 复用现有基础设施 |
| 链路追踪 | Jaeger + OpenTelemetry | 复用现有基础设施 |

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                    前端 (React SPA)                       │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ │
│  │首页  │ │成员  │ │相册  │ │文件  │ │聊天  │ │提醒  │ │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ │
│                    ↓ HTTP/gRPC + WebSocket                │
├─────────────────────────────────────────────────────────┤
│                  API 网关 (cin/interface)                 │
│              JWT 认证 · 路由转发 · 限流                     │
├─────────────────────────────────────────────────────────┤
│                    微服务层                               │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │
│  │ family  │ │storage  │ │  chat   │ │reminder │       │
│  │家庭管理  │ │文件存储  │ │实时聊天  │ │日程提醒  │       │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │
│  ┌─────────┐ ┌─────────┐                                │
│  │  user   │ │ blades  │                                │
│  │用户认证  │ │AI助手   │                                │
│  └─────────┘ └─────────┘                                │
├─────────────────────────────────────────────────────────┤
│                    基础设施层                             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐       │
│  │MySQL │ │Redis │ │RustFS│ │RMQ   │ │Consul│       │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘       │
└─────────────────────────────────────────────────────────┘
```

### 2.2 服务划分

#### 新增服务

| 服务 | 路径 | 职责 |
|------|------|------|
| family | `app/service/family/` | 家庭创建/管理、成员关系、邀请码、角色权限 |
| storage | `app/service/storage/` | 文件/照片上传下载、相册管理、可见性控制 |
| chat | `app/service/chat/` | WebSocket 实时聊天、消息持久化、已读状态 |
| reminder | `app/service/reminder/` | 日程/纪念日管理、重复规则、提醒推送 |

#### 复用服务（需改造）

| 服务 | 改造内容 |
|------|----------|
| user | 支持家庭邀请码注册、成员角色字段、JWT 增加 family_id |
| blades | 增加家庭场景的 AI 对话模板（照片分类建议、日程提醒等） |
| cin/interface | 新增家庭相关服务的路由转发、WebSocket 代理 |
| rmq_consumers | 新增提醒推送消费者、消息通知消费者 |

---

## 3. 权限模型

### 3.1 资源级 ACL 设计

每个可共享的资源（相册、文件、档案条目）都有一个统一的权限结构：

```go
type Visibility string

const (
    VisibilityPrivate  Visibility = "private"   // 仅创建者可见
    VisibilityFamily   Visibility = "family"    // 所有家庭成员可见
    VisibilitySpecific Visibility = "specific"  // 指定成员可见
)

type ResourcePermission struct {
    ResourceID     string      `json:"resource_id"`
    ResourceType   string      `json:"resource_type"`   // album, file, profile_field
    OwnerID        int64       `json:"owner_id"`
    Visibility     Visibility  `json:"visibility"`
    AllowedUserIDs []int64     `json:"allowed_user_ids"` // specific 模式下的白名单
}
```

### 3.2 家庭角色

| 角色 | 权限 |
|------|------|
| owner（创建者） | 邀请/移除成员、修改角色、管理家庭设置、删除家庭 |
| admin（管理员） | 邀请成员、管理内容 |
| member（普通成员） | 管理自己的内容、查看有权限的内容 |

### 3.3 权限校验流程

```
用户请求访问资源
    → 从 JWT 获取 user_id 和 family_id
    → 查询资源的 ResourcePermission
    → 判断：
        - owner == user_id → 允许
        - visibility == "family" 且用户是该家庭成员 → 允许
        - visibility == "specific" 且 user_id in allowed_user_ids → 允许
        - 其他 → 拒绝（403）
```

---

## 4. 各服务详细设计

### 4.1 家庭管理服务（family）

#### 数据模型

```protobuf
message Family {
    int64 id = 1;
    string name = 2;           // 家庭名称，如"温馨小家"
    int64 owner_id = 3;        // 创建者 ID
    string invite_code = 4;    // 邀请码（一次性）
    string avatar = 5;         // 家庭头像
    int64 created_at = 6;
    int64 updated_at = 7;
}

message FamilyMember {
    int64 id = 1;
    int64 family_id = 2;
    int64 user_id = 3;
    string role = 4;           // owner / admin / member
    string nickname = 5;       // 在家庭中的昵称
    string relation_tag = 6;  // 关系标签：爸爸/妈妈/儿子/女儿等
    int64 joined_at = 7;
}
```

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/family/create` | 创建家庭 |
| POST | `/family/join` | 通过邀请码加入 |
| POST | `/family/leave` | 退出家庭 |
| POST | `/family/members/remove` | 移除成员（owner/admin） |
| PUT | `/family/members/role` | 修改成员角色（owner） |
| GET | `/family/info` | 获取家庭信息 |
| GET | `/family/members` | 获取成员列表 |
| POST | `/family/invite/generate` | 生成邀请码 |
| PUT | `/family/update` | 更新家庭信息 |

### 4.2 文件存储服务（storage）

#### 数据模型

```protobuf
message Album {
    int64 id = 1;
    string name = 2;
    string description = 3;
    string cover_url = 4;      // 封面图
    int64 creator_id = 5;
    string visibility = 6;     // private / family / specific
    repeated int64 allowed_user_ids = 7;
    int64 created_at = 8;
}

message Photo {
    int64 id = 1;
    int64 album_id = 2;
    string original_url = 3;   // 原图路径（RustFS）
    string thumbnail_url = 4;   // 缩略图路径
    int64 uploader_id = 5;
    int64 taken_at = 6;        // 拍摄时间（EXIF）
    string description = 7;
    int64 created_at = 8;
}

message File {
    int64 id = 1;
    string name = 2;
    string path = 3;           // RustFS 路径
    int64 size = 4;
    string mime_type = 5;
    int64 uploader_id = 6;
    int64 folder_id = 7;
    string visibility = 8;
    repeated int64 allowed_user_ids = 9;
    int64 created_at = 10;
}

message Folder {
    int64 id = 1;
    string name = 2;
    int64 parent_id = 3;       // 0 表示根目录
    int64 creator_id = 4;
    int64 family_id = 5;
}
```

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/storage/album/create` | 创建相册 |
| GET | `/storage/album/list` | 列出可访问的相册 |
| GET | `/storage/album/{id}` | 获取相册详情 |
| PUT | `/storage/album/{id}` | 修改相册信息/可见性 |
| DELETE | `/storage/album/{id}` | 删除相册 |
| POST | `/storage/photo/upload` | 上传照片（支持批量） |
| GET | `/storage/photo/list` | 获取相册照片列表 |
| DELETE | `/storage/photo/{id}` | 删除照片 |
| POST | `/storage/file/upload` | 上传文件 |
| GET | `/storage/file/list` | 获取文件夹内文件列表 |
| GET | `/storage/file/{id}/download` | 下载文件（权限校验） |
| DELETE | `/storage/file/{id}` | 删除文件 |
| POST | `/storage/folder/create` | 创建文件夹 |

### 4.3 实时聊天服务（chat）

#### 技术方案

- 使用 WebSocket 长连接维持实时通信
- 消息持久化到 MySQL，保证消息不丢失
- 通过 RabbitMQ 广播消息给接收者所在的 chat 服务实例

#### 数据模型

```protobuf
message Conversation {
    int64 id = 1;
    string type = 2;           // "private" / "group"
    string name = 3;           // 群聊名称
    repeated int64 member_ids = 4;
    int64 last_message_at = 5;
    int64 created_at = 6;
}

message Message {
    int64 id = 1;
    int64 conversation_id = 2;
    int64 sender_id = 3;
    string content = 4;
    string type = 5;           // "text" / "image" / "file"
    string file_url = 6;       // 图片/文件的 URL
    repeated int64 read_by = 7; // 已读用户列表
    int64 created_at = 8;
}
```

#### 消息流转

```
发送者 → WebSocket → chat 服务
                    → MySQL（持久化）
                    → RabbitMQ（广播）
                        → 接收者所在 chat 实例
                            → WebSocket → 接收者前端
```

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| WS | `/chat/ws` | WebSocket 连接（JWT 认证） |
| GET | `/chat/conversations` | 获取会话列表 |
| POST | `/chat/conversation/create` | 创建群聊 |
| GET | `/chat/message/history` | 拉取历史消息（分页） |
| POST | `/chat/message/read` | 标记消息已读 |
| GET | `/chat/unread/count` | 获取未读消息总数 |

### 4.4 日程提醒服务（reminder）

#### 数据模型

```protobuf
message Reminder {
    int64 id = 1;
    string title = 2;
    string description = 3;
    int64 remind_at = 4;        // 提醒时间（时间戳）
    int64 event_date = 5;      // 关联日期（如生日当天）
    string repeat_rule = 6;    // "none" / "yearly" / "monthly" / "weekly"
    repeated int64 notify_user_ids = 7; // 通知哪些成员
    int64 creator_id = 8;
    int64 family_id = 9;
    int64 created_at = 10;
}
```

#### 提醒推送机制

```
reminder 服务 → RabbitMQ → rmq_consumers
                            → 查询目标用户的 WebSocket 连接
                            → 推送提醒通知到前端
```

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/reminder/create` | 创建提醒 |
| PUT | `/reminder/{id}` | 修改提醒 |
| DELETE | `/reminder/{id}` | 删除提醒 |
| GET | `/reminder/list` | 按时间范围查询 |
| GET | `/reminder/upcoming` | 获取即将到来的提醒 |

### 4.5 AI 助手集成（blades）

复用现有 blades 服务，在家庭场景中扩展以下能力：

- **家庭助手对话**：通过前端 AI 助手页面直接对话
- **智能照片分类**：上传照片时 AI 自动建议分类标签
- **日程智能提醒**：根据历史数据智能推荐提醒时间

---

## 5. 前端设计

### 5.1 页面结构

| 页面 | 路由 | 说明 |
|------|------|------|
| 登录/注册 | `/login` | 登录、创建家庭、加入家庭 |
| 仪表盘 | `/dashboard` | 家庭概览、统计数据、快捷操作 |
| 家庭成员 | `/members` | 成员列表、邀请管理 |
| 相册 | `/albums` | 相册列表、照片浏览 |
| 相册详情 | `/albums/:id` | 单个相册的照片网格 |
| 文件存储 | `/files` | 文件夹树 + 文件列表 |
| 聊天 | `/chat` | 会话列表 + 聊天窗口 |
| 日程提醒 | `/reminders` | 日历 + 提醒列表 |
| AI 助手 | `/ai` | AI 对话界面 |

### 5.2 UI 设计风格

- **色调**：暖色系，米白/奶油色底 + 珊瑚色/琥珀色强调
- **字体**：Noto Serif SC（标题） + Noto Sans SC（正文）
- **布局**：左侧固定导航栏 + 右侧内容区
- **交互**：页面切换动画、卡片悬浮效果、聊天气泡

### 5.3 原型文件

完整可交互原型位于 `web/family-app/prototype.html`，包含所有 8 个页面的模拟交互。

---

## 6. 数据库设计

### 6.1 新增数据表

```sql
-- 家庭表
CREATE TABLE families (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    owner_id BIGINT NOT NULL,
    invite_code VARCHAR(32) UNIQUE,
    avatar VARCHAR(255),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 家庭成员关系表
CREATE TABLE family_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    family_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    nickname VARCHAR(50),
    relation_tag VARCHAR(20),
    joined_at BIGINT NOT NULL,
    UNIQUE KEY (family_id, user_id)
);

-- 相册表
CREATE TABLE albums (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    cover_url VARCHAR(255),
    creator_id BIGINT NOT NULL,
    family_id BIGINT NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'family',
    allowed_user_ids JSON,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 照片表
CREATE TABLE photos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    album_id BIGINT NOT NULL,
    original_url VARCHAR(255) NOT NULL,
    thumbnail_url VARCHAR(255),
    uploader_id BIGINT NOT NULL,
    taken_at BIGINT,
    description TEXT,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 文件表
CREATE TABLE files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploader_id BIGINT NOT NULL,
    folder_id BIGINT DEFAULT 0,
    family_id BIGINT NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'private',
    allowed_user_ids JSON,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 文件夹表
CREATE TABLE folders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    parent_id BIGINT DEFAULT 0,
    creator_id BIGINT NOT NULL,
    family_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 会话表
CREATE TABLE conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    family_id BIGINT,
    last_message_at BIGINT,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);

-- 会话成员表
CREATE TABLE conversation_members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    last_read_at BIGINT DEFAULT 0,
    UNIQUE KEY (conversation_id, user_id)
);

-- 消息表
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    content TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    file_url VARCHAR(255),
    read_by JSON,
    created_at BIGINT NOT NULL,
    INDEX idx_conversation_created (conversation_id, created_at)
);

-- 提醒表
CREATE TABLE reminders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    remind_at BIGINT NOT NULL,
    event_date BIGINT,
    repeat_rule VARCHAR(20) NOT NULL DEFAULT 'none',
    notify_user_ids JSON,
    creator_id BIGINT NOT NULL,
    family_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT DEFAULT 0
);
```

### 6.2 user 表改造

```sql
-- 在现有 users 表中新增字段
ALTER TABLE users ADD COLUMN family_id BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN avatar VARCHAR(255);
ALTER TABLE users ADD COLUMN nickname VARCHAR(50);
```

---

## 7. 项目目录结构

```
cyi-app/
├── api/
│   ├── family/v1/              ← 新增：家庭服务 proto
│   ├── storage/v1/             ← 新增：存储服务 proto
│   ├── chat/v1/                ← 新增：聊天服务 proto
│   └── reminder/v1/            ← 新增：提醒服务 proto
├── app/
│   ├── service/
│   │   ├── family/             ← 新增：家庭管理服务
│   │   ├── storage/            ← 新增：文件存储服务
│   │   ├── chat/               ← 新增：实时聊天服务
│   │   ├── reminder/           ← 新增：日程提醒服务
│   │   ├── user/               ← 改造：增加家庭相关字段
│   │   └── order/              ← 保持不变
│   ├── blades/                 ← 改造：增加家庭 AI 模板
│   ├── cin/interface/          ← 改造：增加新服务路由 + WS 代理
│   └── rmq_consumers/          ← 扩展：增加提醒推送消费者
├── web/
│   └── family-app/             ← 新增：前端项目
│       ├── prototype.html       ← 原型文件
│       ├── src/
│       │   ├── pages/
│       │   ├── components/
│       │   ├── stores/
│       │   ├── services/
│       │   └── utils/
│       ├── package.json
│       └── vite.config.ts
├── deploy/
│   └── docker-compose/
│       └── docker-compose.app.yml  ← 更新：增加新服务
└── docs/
    └── superpowers/specs/
        └── 2026-06-13-family-platform-design.md  ← 本文档
```

---

## 8. 实施计划

### 阶段一：基础框架（1-2 周）

1. 使用 `tools/service-gen` 生成 family、storage、chat、reminder 四个服务骨架
2. 编写 Proto 文件并生成代码
3. 配置 Wire 依赖注入
4. 改造 user 服务（增加 family_id 等字段）
5. 更新 cin/interface 网关路由

### 阶段二：核心功能（2-3 周）

1. 实现家庭管理服务（创建/加入/退出/邀请）
2. 实现文件存储服务（上传/下载/相册管理）
3. 实现细粒度权限校验中间件
4. 前端项目初始化 + 登录/注册页面
5. 前端仪表盘 + 家庭成员页面

### 阶段三：实时通信（1-2 周）

1. 实现 WebSocket 服务（chat 服务）
2. 实现消息持久化和历史消息查询
3. RabbitMQ 消息广播
4. 前端聊天页面 + WebSocket 客户端

### 阶段四：辅助功能（1 周）

1. 实现日程提醒服务
2. 提醒推送（RabbitMQ → WebSocket）
3. 前端日程页面
4. 集成 blades AI 助手

### 阶段五：完善部署（1 周）

1. 更新 docker-compose 配置
2. 更新 K8s 部署配置
3. 更新 CI/CD 流水线
4. 测试和优化

---

## 9. 风险和注意事项

1. **WebSocket 连接管理**：需要处理断线重连、心跳检测、多设备登录
2. **文件上传大小限制**：需要配置合理的上传大小限制和超时时间
3. **照片缩略图生成**：上传时异步生成缩略图，避免阻塞主流程
4. **消息顺序保证**：RabbitMQ 广播时需要保证同一会话内消息的顺序性
5. **数据库隔离**：家庭相关表与电商业务表在同一数据库中，需注意表名前缀隔离
