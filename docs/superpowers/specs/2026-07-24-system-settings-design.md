# 系统设置模块重新设计方案

## 1. 概述

重新设计悦选商城管理后台的**系统设置**模块。当前实现仅有 Mock 数据，无后端持久化、无真实 API 对接。

### 目标

- 建立完整的后端链路（数据持久化、业务逻辑、API 接口）
- 提供 7 大设置分类，覆盖电商后台核心配置
- 前端界面精美，动效流畅
- 符合现有架构模式（BFF 透传 → User Service 持久化）

## 2. 架构

```
SettingsView.vue ──HTTP/JSON──→ Admin BFF (cin/admin) ──gRPC──→ User Service (service/user)
                                       │                             │
                                       ├─ service/settings.go        ├─ service/settings.go
                                       ├─ biz/SettingsUseCase         ├─ biz/settings.go
                                       └─ data/ (gRPC client)        └─ data/ (MySQL system_settings)
```

**设计原则：**
- Admin BFF 不直接访问数据库，委托 User Service 处理
- 与角色管理、员工管理等现有模块架构一致
- 设置以单行 JSON 存储，便于灵活扩展

## 3. 数据存储

### 表结构

```sql
CREATE TABLE `system_settings` (
    `id`         INT PRIMARY KEY DEFAULT 1 COMMENT '固定为 1，保证单行记录',
    `settings`   JSON NOT NULL COMMENT '所有配置项的 JSON 对象',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(64) DEFAULT '' COMMENT '最后修改人'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### JSON 数据结构

```json
{
  "shop_name": "悦选商城",
  "shop_logo": "https://oss.example.com/logo.png",
  "shop_description": "精选好物，品质生活",
  "contact_phone": "400-888-8888",
  "contact_email": "service@yuexuan.com",
  "icp_record": "京ICP备2025XXXXXX号",

  "auto_confirm_days": 7,
  "auto_close_minutes": 30,
  "refund_timeout_days": 3,

  "stock_alert_threshold": 10,
  "new_product_online": true,
  "product_approval": false,

  "new_order_notify": true,
  "refund_notify": true,
  "user_register_notify": false,
  "stock_alert_notify": true,

  "login_attempt_limit": true,
  "login_captcha": true,
  "session_timeout_minutes": 120,

  "enabled_payments": ["wechat", "alipay"],
  "default_payment": "wechat",

  "free_shipping_threshold": 99.00,
  "default_shipping_company": "顺丰速运"
}
```

## 4. Proto API 设计

### 4.1 User Service Proto（新增）

文件：`api/service/user/v1/settings.proto`

```protobuf
syntax = "proto3";
package api.service.user.v1;
option go_package = "cyi-app/api/service/user/v1;v1";

service AdminSettings {
  rpc GetSettings (GetSettingsRequest) returns (GetSettingsReply);
  rpc UpdateSettings (UpdateSettingsRequest) returns (UpdateSettingsReply);
}

message GetSettingsRequest {}

message GetSettingsReply {
  string shop_name = 1;
  string shop_logo = 2;
  string shop_description = 3;
  string contact_phone = 4;
  string contact_email = 5;
  string icp_record = 6;

  int32 auto_confirm_days = 7;
  int32 auto_close_minutes = 8;
  int32 refund_timeout_days = 9;

  int32 stock_alert_threshold = 10;
  bool new_product_online = 11;
  bool product_approval = 12;

  bool new_order_notify = 13;
  bool refund_notify = 14;
  bool user_register_notify = 15;
  bool stock_alert_notify = 16;

  bool login_attempt_limit = 17;
  bool login_captcha = 18;
  int32 session_timeout_minutes = 19;

  repeated string enabled_payments = 20;
  string default_payment = 21;

  double free_shipping_threshold = 22;
  string default_shipping_company = 23;
}

message UpdateSettingsRequest {
  string shop_name = 1;
  string shop_logo = 2;
  string shop_description = 3;
  string contact_phone = 4;
  string contact_email = 5;
  string icp_record = 6;

  int32 auto_confirm_days = 7;
  int32 auto_close_minutes = 8;
  int32 refund_timeout_days = 9;

  int32 stock_alert_threshold = 10;
  bool new_product_online = 11;
  bool product_approval = 12;

  bool new_order_notify = 13;
  bool refund_notify = 14;
  bool user_register_notify = 15;
  bool stock_alert_notify = 16;

  bool login_attempt_limit = 17;
  bool login_captcha = 18;
  int32 session_timeout_minutes = 19;

  repeated string enabled_payments = 20;
  string default_payment = 21;

  double free_shipping_threshold = 22;
  string default_shipping_company = 23;
}

message UpdateSettingsReply {
  bool success = 1;
}
```

### 4.2 Admin BFF Proto（重写）

文件：`api/cin/admin/v1/settings.proto`

与 User Service Proto 保持一致，HTTP 路由不变：
- `GET /admin/v1/settings` → GetSettings
- `PUT /admin/v1/settings` → UpdateSettings

移除 `ChangePassword` RPC（移至 Profile 模块）。

## 5. 后端实现

### 5.1 User Service

| 层 | 文件 | 说明 |
|----|------|------|
| service | `app/service/user/internal/service/settings.go` | gRPC 服务实现，调用 biz |
| biz | `app/service/user/internal/biz/settings.go` | SettingsUseCase，含 GetSettings / UpdateSettings |
| data | `app/service/user/internal/data/settings.go` | 操作 `system_settings` 表，JSON 序列化/反序列化 |

#### Biz 核心逻辑

```go
type SettingsRepo interface {
    Get(ctx context.Context) (*SystemSettings, error)
    Update(ctx context.Context, s *SystemSettings) error
}

type SettingsUseCase struct {
    repo SettingsRepo
    log  *log.Helper
}
```

#### Data 层

- 首次查询时自动创建默认配置行（id=1）
- 更新时只修改传入的字段，不覆盖未传的字段（全量更新模式）
- 使用 GORM 操作 JSON 类型字段

### 5.2 Admin BFF

| 层 | 文件 | 说明 |
|----|------|------|
| service | `app/cin/admin/internal/service/settings.go` | 重写，调用 biz，透传数据 |
| biz | `app/cin/admin/internal/biz/settings.go` | 新增 SettingsUseCase |
| data | `app/cin/admin/internal/data/settings.go` | 新增，gRPC 代理到 User Service |

## 6. 前端实现

### 6.1 页面结构

文件：`web/admin/src/views/settings/SettingsView.vue`

**Tab 导航 + 6 个标签页：**

| 标签页 | 设置项 | UI 控件 |
|--------|--------|---------|
| 🏪 店铺设置 | 商城名称、Logo、简介、客服电话/邮箱、ICP | Input, Upload, Textarea |
| 📦 订单设置 | 自动确认天数、自动关闭分钟、退款超时天数 | InputNumber |
| 🛒 商品设置 | 库存预警阈值、默认上架状态、审核开关 | InputNumber, Switch |
| 🔔 通知设置 | 4 个通知开关 | Switch |
| 🔒 安全策略 | 登录尝试限制、验证码开关、会话超时 | Switch, InputNumber |
| 💳 支付设置 | 支付方式多选、默认支付渠道 | CheckboxGroup, Select |
| 🚚 物流设置 | 包邮门槛、默认物流公司 | InputNumber, Input |

### 6.2 API 集成

新增 `web/admin/src/api/settings.ts`：

```typescript
export function getSettings(): Promise<SystemSettings> {
  return request.get('/admin/v1/settings')
}

export function updateSettings(data: Partial<SystemSettings>): Promise<{ success: boolean }> {
  return request.put('/admin/v1/settings', data)
}
```

### 6.3 UI 动效设计

| 动效 | 实现方式 | 效果 |
|------|----------|------|
| 页面入场 | CSS @keyframes + 错落延迟 | fadeSlideUp |
| Tab 切换 | Transition name="tab-slide" | 水平滑动 |
| 保存反馈 | 按钮 loading + 成功动画 | 按钮 → ✓ → 恢复 |
| 卡片入场 | 逐张淡入上滑 | 错落 0.05s 步进 |
| 开关切换 | Element Plus 原生弹性动画 | 流畅过渡 |

### 6.4 布局

- 使用 `el-tabs type="border-card"` 风格
- 每个 Tab 内部分组卡片（`el-card`）
- 底部固定的「保存设置」和「重置」按钮
- 响应式布局：宽屏多列，窄屏单列

## 7. 文件清单

### 新增文件

```
api/service/user/v1/settings.proto           # User Service Proto
app/service/user/internal/biz/settings.go     # User Service UseCase
app/service/user/internal/data/settings.go    # User Service Data 层
app/service/user/internal/service/settings.go # User Service gRPC 服务
web/admin/src/api/settings.ts                 # 前端 API
```

### 修改文件

```
api/cin/admin/v1/settings.proto              # 重写 Proto（移除 ChangePassword）
app/cin/admin/internal/service/settings.go    # 重写，调真实 UseCase
app/cin/admin/internal/biz/settings.go        # 新增 SettingsUseCase
app/cin/admin/internal/data/settings.go       # 新增 gRPC 代理
app/cin/admin/internal/biz/biz.go             # 注册 ProviderSet
app/cin/admin/internal/data/data.go           # 注册 gRPC Client
app/cin/admin/internal/service/service.go     # 注入 SettingsUseCase
web/admin/src/views/settings/SettingsView.vue # 重写完整页面
app/service/user/internal/biz/biz.go          # 注册 ProviderSet
app/service/user/internal/data/data.go        # 注册 ProviderSet
app/service/user/internal/service/service.go  # 注册 gRPC 服务
```

### 删除

```
app/cin/admin/internal/service/settings.go 中的 Mock 实现
```

## 8. 验证方式

1. `make api` - 生成 Proto 代码
2. `make build` - 编译 Admin BFF 和 User Service
3. 重新启动 User Service 和 Admin BFF
4. 前端访问 `/settings` 验证：
   - 各 Tab 能正常加载配置
   - 修改配置后保存成功
   - 刷新页面后配置持久化
