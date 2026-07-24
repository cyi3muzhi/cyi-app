# 系统设置模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重新实现系统设置模块，包含 7 大设置分类，后端完整链路（User Service 持久化），前端对接真实 API。

**Architecture:** SettingsView.vue → Admin BFF (cin/admin) HTTP API → gRPC → User Service (service/user) → MySQL。Admin BFF 不直接访问数据库，委托 User Service 处理，与角色管理架构一致。

**Tech Stack:** Kratos, Go, GORM, MySQL JSON, Vue 3, Element Plus, Wire

---

### Task 1: User Service Proto 定义

**Files:**
- Create: `api/service/user/v1/settings.proto`

- [ ] **Step 1: 创建 User Service 的 settings.proto**

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

- [ ] **Step 2: 更新 Admin BFF 的 settings.proto**

修改 `api/cin/admin/v1/settings.proto`，移除 `ChangePassword` RPC，数据模型与 User Service 一致，保留 HTTP 路由：

```protobuf
syntax = "proto3";
package api.cin.admin.v1;
option go_package = "cyi-app/api/cin/admin/v1;adminv1";
option java_multiple_files = true;
option java_package = "api.cin.admin.v1";
import "google/api/annotations.proto";

service Settings {
  rpc GetSettings (GetSettingsRequest) returns (GetSettingsReply) {
    option (google.api.http) = { get: "/admin/v1/settings" };
  }
  rpc UpdateSettings (UpdateSettingsRequest) returns (UpdateSettingsReply) {
    option (google.api.http) = { put: "/admin/v1/settings" body: "*" };
  }
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

- [ ] **Step 3: 生成 Proto 代码**

Run: `make api`
Expected: 
- `api/service/user/v1/settings.pb.go`, `settings_grpc.pb.go`, `settings_vtproto.pb.go` 生成
- `api/cin/admin/v1/settings.pb.go`, `settings_grpc.pb.go`, `settings_http.pb.go`, `settings_vtproto.pb.go` 重新生成（不再有 ChangePassword）

- [ ] **Step 4: Commit**

```bash
git add api/service/user/v1/settings.proto api/cin/admin/v1/settings.proto
git add api/
git commit -m "feat(settings): 定义系统设置 Proto API，移除 ChangePassword"
```

---

### Task 2: User Service — Model & Data 层

**Files:**
- Create: `app/service/user/internal/data/models/settings.go`
- Create: `app/service/user/internal/data/settings.go`

- [ ] **Step 1: 创建 SystemSettings 模型**

`app/service/user/internal/data/models/settings.go`:

```go
package models

import "encoding/json"

// SystemSettings 系统设置，以 JSON 存储在 system_settings 表
type SystemSettings struct {
	ShopName              string    `json:"shop_name"`
	ShopLogo              string    `json:"shop_logo"`
	ShopDescription       string    `json:"shop_description"`
	ContactPhone          string    `json:"contact_phone"`
	ContactEmail          string    `json:"contact_email"`
	IcpRecord             string    `json:"icp_record"`
	AutoConfirmDays       int32     `json:"auto_confirm_days"`
	AutoCloseMinutes      int32     `json:"auto_close_minutes"`
	RefundTimeoutDays     int32     `json:"refund_timeout_days"`
	StockAlertThreshold   int32     `json:"stock_alert_threshold"`
	NewProductOnline      bool      `json:"new_product_online"`
	ProductApproval       bool      `json:"product_approval"`
	NewOrderNotify        bool      `json:"new_order_notify"`
	RefundNotify          bool      `json:"refund_notify"`
	UserRegisterNotify    bool      `json:"user_register_notify"`
	StockAlertNotify      bool      `json:"stock_alert_notify"`
	LoginAttemptLimit     bool      `json:"login_attempt_limit"`
	LoginCaptcha          bool      `json:"login_captcha"`
	SessionTimeoutMinutes int32     `json:"session_timeout_minutes"`
	EnabledPayments       []string  `json:"enabled_payments"`
	DefaultPayment        string    `json:"default_payment"`
	FreeShippingThreshold float64   `json:"free_shipping_threshold"`
	DefaultShippingCompany string   `json:"default_shipping_company"`
}

// SettingsRow 数据库行记录（system_settings 表）
type SettingsRow struct {
	ID        int32           `gorm:"primaryKey;default:1"`
	Settings  json.RawMessage `gorm:"type:json;not null"`
	UpdatedAt int64           `gorm:"autoUpdateTime:milli"`
	UpdatedBy string          `gorm:"size:64"`
}

// TableName 指定表名
func (SettingsRow) TableName() string {
	return "system_settings"
}
```

- [ ] **Step 2: 创建 User Service data/settings.go**

`app/service/user/internal/data/settings.go`:

```go
package data

import (
	"context"
	"encoding/json"
	"time"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type settingsRepo struct {
	data *Data
	log  *log.Helper
}

func NewSettingsRepo(data *Data, logger log.Logger) biz.SettingsRepo {
	return &settingsRepo{data: data, log: log.NewHelper(logger)}
}

// defaultSettings 返回默认系统设置
func defaultSettings() *biz.SystemSettings {
	return &biz.SystemSettings{
		ShopName:              "悦选商城",
		ShopDescription:       "精选好物，品质生活",
		ContactPhone:          "400-888-8888",
		ContactEmail:          "service@yuexuan.com",
		IcpRecord:             "京ICP备2025XXXXXX号",
		AutoConfirmDays:       7,
		AutoCloseMinutes:      30,
		RefundTimeoutDays:     3,
		StockAlertThreshold:   10,
		NewProductOnline:      true,
		ProductApproval:       false,
		NewOrderNotify:        true,
		RefundNotify:          true,
		UserRegisterNotify:    false,
		StockAlertNotify:      true,
		LoginAttemptLimit:     true,
		LoginCaptcha:          true,
		SessionTimeoutMinutes: 120,
		EnabledPayments:       []string{"wechat", "alipay"},
		DefaultPayment:        "wechat",
		FreeShippingThreshold: 99.00,
		DefaultShippingCompany: "顺丰速运",
	}
}

// settingsToJSON 将 biz 结构体序列化为 JSON
func settingsToJSON(s *biz.SystemSettings) (json.RawMessage, error) {
	return json.Marshal(s)
}

// settingsFromJSON 将 JSON 反序列化为 biz 结构体
func settingsFromJSON(data json.RawMessage) (*biz.SystemSettings, error) {
	var s biz.SystemSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Get 获取系统设置，首次查询时自动创建默认配置
func (r *settingsRepo) Get(ctx context.Context) (*biz.SystemSettings, error) {
	var row models.SettingsRow
	err := r.data.db.WithContext(ctx).First(&row, 1).Error
	if err == gorm.ErrRecordNotFound {
		// 首次使用，创建默认配置
		def := defaultSettings()
		jsonData, _ := settingsToJSON(def)
		row = models.SettingsRow{
			ID:        1,
			Settings:  jsonData,
			UpdatedAt: time.Now().UnixMilli(),
			UpdatedBy: "system",
		}
		if err := r.data.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return def, nil
	}
	if err != nil {
		return nil, err
	}
	return settingsFromJSON(row.Settings)
}

// Update 更新系统设置（全量覆盖）
func (r *settingsRepo) Update(ctx context.Context, s *biz.SystemSettings) error {
	jsonData, err := settingsToJSON(s)
	if err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&models.SettingsRow{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"settings":   string(jsonData),
			"updated_at": time.Now().UnixMilli(),
		}).Error
}
```

- [ ] **Step 3: 注册 Data ProviderSet**

修改 `app/service/user/internal/data/data.go`，在 `var ProviderSet` 中添加 `NewSettingsRepo`：

```go
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewAdminUserRepo,
	NewRoleRepo,
	NewSettingsRepo,  // ← 新增
	NewData,
	NewDB,
	NewRedisClient,
	NewOssClient,
	NewCasbinEnforcer,
)
```

同时，在 `data.go` 的 `NewDB` 方法中添加 AutoMigrate：

```go
// 在 env != "production" 的 AutoMigrate 中添加
if err := db.AutoMigrate(&mod.User{}, &mod.AdminUser{}, &mod.CasbinRule{}, &mod.RoleMeta{}, &mod.SettingsRow{}); err != nil {
	panic(err)
}
```

需要在顶部 import models 时使用别名，例如 `mod "cyi-app/app/service/user/internal/data/models"`，`SettingsRow` 也需要注册进去。检查现有代码发现 model 引用路径为 `mod "cyi-app/app/service/user/internal/data/models"`，所以需要将 `models/settings.go` 中的结构体用于 AutoMigrate：

```go
// 在 data.go 中修改 import 增加对 models 的引用，AutoMigrate 增加新的表
db.AutoMigrate(&mod.User{}, &mod.AdminUser{}, &mod.CasbinRule{}, &mod.RoleMeta{}, &mod.SettingsRow{})
```

- [ ] **Step 4: Commit**

```bash
git add app/service/user/internal/data/models/settings.go
git add app/service/user/internal/data/settings.go
git add app/service/user/internal/data/data.go
git commit -m "feat(settings): 添加 SystemSettings 模型和 Data 层"
```

---

### Task 3: User Service — Biz 层

**Files:**
- Create: `app/service/user/internal/biz/settings.go`

- [ ] **Step 1: 创建 SettingsUseCase**

`app/service/user/internal/biz/settings.go`:

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// SystemSettings 系统设置业务模型
type SystemSettings struct {
	ShopName              string
	ShopLogo              string
	ShopDescription       string
	ContactPhone          string
	ContactEmail          string
	IcpRecord             string
	AutoConfirmDays       int32
	AutoCloseMinutes      int32
	RefundTimeoutDays     int32
	StockAlertThreshold   int32
	NewProductOnline      bool
	ProductApproval       bool
	NewOrderNotify        bool
	RefundNotify          bool
	UserRegisterNotify    bool
	StockAlertNotify      bool
	LoginAttemptLimit     bool
	LoginCaptcha          bool
	SessionTimeoutMinutes int32
	EnabledPayments       []string
	DefaultPayment        string
	FreeShippingThreshold float64
	DefaultShippingCompany string
}

// SettingsRepo 系统设置仓库接口
type SettingsRepo interface {
	Get(ctx context.Context) (*SystemSettings, error)
	Update(ctx context.Context, s *SystemSettings) error
}

// SettingsUseCase 系统设置用例
type SettingsUseCase struct {
	repo SettingsRepo
	log  *log.Helper
}

// NewSettingsUseCase 创建 SettingsUseCase
func NewSettingsUseCase(repo SettingsRepo, logger log.Logger) *SettingsUseCase {
	return &SettingsUseCase{repo: repo, log: log.NewHelper(logger)}
}

// Get 获取系统设置
func (uc *SettingsUseCase) Get(ctx context.Context) (*SystemSettings, error) {
	return uc.repo.Get(ctx)
}

// Update 更新系统设置
func (uc *SettingsUseCase) Update(ctx context.Context, s *SystemSettings) error {
	return uc.repo.Update(ctx, s)
}
```

- [ ] **Step 2: 注册 Biz ProviderSet**

修改 `app/service/user/internal/biz/biz.go`：

```go
var ProviderSet = wire.NewSet(
	NewUserUseCase,
	NewAdminUserUseCase,
	NewRoleUseCase,
	NewSettingsUseCase,  // ← 新增
	ProvideApiKey,
)
```

- [ ] **Step 3: Commit**

```bash
git add app/service/user/internal/biz/settings.go
git add app/service/user/internal/biz/biz.go
git commit -m "feat(settings): 添加 SettingsUseCase"
```

---

### Task 4: User Service — Service 层 & gRPC 注册

**Files:**
- Create: `app/service/user/internal/service/settings.go`

- [ ] **Step 1: 创建 gRPC 服务层**

`app/service/user/internal/service/settings.go`:

```go
package service

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
)

// AdminSettingsService 系统设置 gRPC 服务
type AdminSettingsService struct {
	v1.UnimplementedAdminSettingsServer
	uc *biz.SettingsUseCase
}

// NewAdminSettingsService 创建 AdminSettingsService
func NewAdminSettingsService(uc *biz.SettingsUseCase) *AdminSettingsService {
	return &AdminSettingsService{uc: uc}
}

// bizToProto 将 biz 结构体转换为 proto reply
func bizToProto(s *biz.SystemSettings) *v1.GetSettingsReply {
	if s == nil {
		return nil
	}
	payments := s.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &v1.GetSettingsReply{
		ShopName:              s.ShopName,
		ShopLogo:              s.ShopLogo,
		ShopDescription:       s.ShopDescription,
		ContactPhone:          s.ContactPhone,
		ContactEmail:          s.ContactEmail,
		IcpRecord:             s.IcpRecord,
		AutoConfirmDays:       s.AutoConfirmDays,
		AutoCloseMinutes:      s.AutoCloseMinutes,
		RefundTimeoutDays:     s.RefundTimeoutDays,
		StockAlertThreshold:   s.StockAlertThreshold,
		NewProductOnline:      s.NewProductOnline,
		ProductApproval:       s.ProductApproval,
		NewOrderNotify:        s.NewOrderNotify,
		RefundNotify:          s.RefundNotify,
		UserRegisterNotify:    s.UserRegisterNotify,
		StockAlertNotify:      s.StockAlertNotify,
		LoginAttemptLimit:     s.LoginAttemptLimit,
		LoginCaptcha:          s.LoginCaptcha,
		SessionTimeoutMinutes: s.SessionTimeoutMinutes,
		EnabledPayments:       payments,
		DefaultPayment:        s.DefaultPayment,
		FreeShippingThreshold: s.FreeShippingThreshold,
		DefaultShippingCompany: s.DefaultShippingCompany,
	}
}

// protoToBiz 将 proto request 转换为 biz 结构体
func protoToBiz(req *v1.UpdateSettingsRequest) *biz.SystemSettings {
	return &biz.SystemSettings{
		ShopName:              req.ShopName,
		ShopLogo:              req.ShopLogo,
		ShopDescription:       req.ShopDescription,
		ContactPhone:          req.ContactPhone,
		ContactEmail:          req.ContactEmail,
		IcpRecord:             req.IcpRecord,
		AutoConfirmDays:       req.AutoConfirmDays,
		AutoCloseMinutes:      req.AutoCloseMinutes,
		RefundTimeoutDays:     req.RefundTimeoutDays,
		StockAlertThreshold:   req.StockAlertThreshold,
		NewProductOnline:      req.NewProductOnline,
		ProductApproval:       req.ProductApproval,
		NewOrderNotify:        req.NewOrderNotify,
		RefundNotify:          req.RefundNotify,
		UserRegisterNotify:    req.UserRegisterNotify,
		StockAlertNotify:      req.StockAlertNotify,
		LoginAttemptLimit:     req.LoginAttemptLimit,
		LoginCaptcha:          req.LoginCaptcha,
		SessionTimeoutMinutes: req.SessionTimeoutMinutes,
		EnabledPayments:       req.EnabledPayments,
		DefaultPayment:        req.DefaultPayment,
		FreeShippingThreshold: req.FreeShippingThreshold,
		DefaultShippingCompany: req.DefaultShippingCompany,
	}
}

// GetSettings 获取系统设置
func (s *AdminSettingsService) GetSettings(ctx context.Context, req *v1.GetSettingsRequest) (*v1.GetSettingsReply, error) {
	settings, err := s.uc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return bizToProto(settings), nil
}

// UpdateSettings 更新系统设置
func (s *AdminSettingsService) UpdateSettings(ctx context.Context, req *v1.UpdateSettingsRequest) (*v1.UpdateSettingsReply, error) {
	if err := s.uc.Update(ctx, protoToBiz(req)); err != nil {
		return nil, err
	}
	return &v1.UpdateSettingsReply{Success: true}, nil
}
```

- [ ] **Step 2: 注册 Service ProviderSet**

修改 `app/service/user/internal/service/service.go`：

```go
var ProviderSet = wire.NewSet(
	NewUserService,
	NewAdminUserService,
	NewAdminRoleService,
	NewAdminSettingsService,  // ← 新增
)
```

- [ ] **Step 3: 注册 gRPC Server**

修改 `app/service/user/internal/server/grpc.go`，注入 `*service.AdminSettingsService` 并注册：

```go
func NewServer(s *ts.GrpcServer, userService *service.UserService, adminUserService *service.AdminUserService, adminRoleService *service.AdminRoleService, adminSettingsService *service.AdminSettingsService) *grpc.Server {
	v1.RegisterUserServer(s.GrpcServer, userService)
	v1.RegisterAdminUserServer(s.GrpcServer, adminUserService)
	v1.RegisterAdminRoleServer(s.GrpcServer, adminRoleService)
	v1.RegisterAdminSettingsServer(s.GrpcServer, adminSettingsService)  // ← 新增
	return s.GrpcServer
}
```

- [ ] **Step 4: 更新 Wire 生成**

Run: `cd app/service/user && make wire`
Expected: `wire_gen.go` 重新生成，编译通过。

- [ ] **Step 5: 验证编译**

Run: `cd app/service/user && go build ./...`
Expected: 编译成功，无错误。

- [ ] **Step 6: Commit**

```bash
git add app/service/user/internal/service/settings.go
git add app/service/user/internal/service/service.go
git add app/service/user/internal/server/grpc.go
git commit -m "feat(settings): 添加 gRPC 服务层并注册到 User Service"
```

---

### Task 5: Admin BFF — Data & Biz 层

**Files:**
- Create: `app/cin/admin/internal/data/settings.go`
- Create: `app/cin/admin/internal/biz/settings.go`

- [ ] **Step 1: 创建 Admin BFF 的 Biz 层**

`app/cin/admin/internal/biz/settings.go`:

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// SystemSettings 系统设置（Admin BFF 业务模型）
type SystemSettings struct {
	ShopName              string
	ShopLogo              string
	ShopDescription       string
	ContactPhone          string
	ContactEmail          string
	IcpRecord             string
	AutoConfirmDays       int32
	AutoCloseMinutes      int32
	RefundTimeoutDays     int32
	StockAlertThreshold   int32
	NewProductOnline      bool
	ProductApproval       bool
	NewOrderNotify        bool
	RefundNotify          bool
	UserRegisterNotify    bool
	StockAlertNotify      bool
	LoginAttemptLimit     bool
	LoginCaptcha          bool
	SessionTimeoutMinutes int32
	EnabledPayments       []string
	DefaultPayment        string
	FreeShippingThreshold float64
	DefaultShippingCompany string
}

// SettingsRepo 系统设置仓库接口
type SettingsRepo interface {
	Get(ctx context.Context) (*SystemSettings, error)
	Update(ctx context.Context, s *SystemSettings) error
}

// SettingsUseCase 系统设置用例
type SettingsUseCase struct {
	repo SettingsRepo
	log  *log.Helper
}

// NewSettingsUseCase 创建 SettingsUseCase
func NewSettingsUseCase(repo SettingsRepo, logger log.Logger) *SettingsUseCase {
	return &SettingsUseCase{repo: repo, log: log.NewHelper(logger)}
}

// Get 获取系统设置
func (uc *SettingsUseCase) Get(ctx context.Context) (*SystemSettings, error) {
	return uc.repo.Get(ctx)
}

// Update 更新系统设置
func (uc *SettingsUseCase) Update(ctx context.Context, s *SystemSettings) error {
	return uc.repo.Update(ctx, s)
}
```

- [ ] **Step 2: 创建 Admin BFF 的 Data 层（gRPC 代理）**

`app/cin/admin/internal/data/settings.go`:

```go
package data

import (
	"context"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type settingsRepo struct {
	client userV1.AdminSettingsClient
	log    *log.Helper
}

func NewSettingsRepo(client userV1.AdminSettingsClient, logger log.Logger) biz.SettingsRepo {
	return &settingsRepo{client: client, log: log.NewHelper(logger)}
}

// bizToProto 将 biz 结构体转换为 proto request
func bizToProto(s *biz.SystemSettings) *userV1.UpdateSettingsRequest {
	payments := s.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &userV1.UpdateSettingsRequest{
		ShopName:              s.ShopName,
		ShopLogo:              s.ShopLogo,
		ShopDescription:       s.ShopDescription,
		ContactPhone:          s.ContactPhone,
		ContactEmail:          s.ContactEmail,
		IcpRecord:             s.IcpRecord,
		AutoConfirmDays:       s.AutoConfirmDays,
		AutoCloseMinutes:      s.AutoCloseMinutes,
		RefundTimeoutDays:     s.RefundTimeoutDays,
		StockAlertThreshold:   s.StockAlertThreshold,
		NewProductOnline:      s.NewProductOnline,
		ProductApproval:       s.ProductApproval,
		NewOrderNotify:        s.NewOrderNotify,
		RefundNotify:          s.RefundNotify,
		UserRegisterNotify:    s.UserRegisterNotify,
		StockAlertNotify:      s.StockAlertNotify,
		LoginAttemptLimit:     s.LoginAttemptLimit,
		LoginCaptcha:          s.LoginCaptcha,
		SessionTimeoutMinutes: s.SessionTimeoutMinutes,
		EnabledPayments:       payments,
		DefaultPayment:        s.DefaultPayment,
		FreeShippingThreshold: s.FreeShippingThreshold,
		DefaultShippingCompany: s.DefaultShippingCompany,
	}
}

// protoToBiz 将 proto reply 转换为 biz 结构体
func protoToBiz(resp *userV1.GetSettingsReply) *biz.SystemSettings {
	return &biz.SystemSettings{
		ShopName:              resp.ShopName,
		ShopLogo:              resp.ShopLogo,
		ShopDescription:       resp.ShopDescription,
		ContactPhone:          resp.ContactPhone,
		ContactEmail:          resp.ContactEmail,
		IcpRecord:             resp.IcpRecord,
		AutoConfirmDays:       resp.AutoConfirmDays,
		AutoCloseMinutes:      resp.AutoCloseMinutes,
		RefundTimeoutDays:     resp.RefundTimeoutDays,
		StockAlertThreshold:   resp.StockAlertThreshold,
		NewProductOnline:      resp.NewProductOnline,
		ProductApproval:       resp.ProductApproval,
		NewOrderNotify:        resp.NewOrderNotify,
		RefundNotify:          resp.RefundNotify,
		UserRegisterNotify:    resp.UserRegisterNotify,
		StockAlertNotify:      resp.StockAlertNotify,
		LoginAttemptLimit:     resp.LoginAttemptLimit,
		LoginCaptcha:          resp.LoginCaptcha,
		SessionTimeoutMinutes: resp.SessionTimeoutMinutes,
		EnabledPayments:       resp.EnabledPayments,
		DefaultPayment:        resp.DefaultPayment,
		FreeShippingThreshold: resp.FreeShippingThreshold,
		DefaultShippingCompany: resp.DefaultShippingCompany,
	}
}

func (r *settingsRepo) Get(ctx context.Context) (*biz.SystemSettings, error) {
	resp, err := r.client.GetSettings(ctx, &userV1.GetSettingsRequest{})
	if err != nil {
		return nil, err
	}
	return protoToBiz(resp), nil
}

func (r *settingsRepo) Update(ctx context.Context, s *biz.SystemSettings) error {
	_, err := r.client.UpdateSettings(ctx, bizToProto(s))
	return err
}
```

- [ ] **Step 3: 创建 gRPC 客户端连接**

在 `app/cin/admin/internal/data/client_conn.go` 中添加：

```go
// NewAdminSettingsClient 创建系统设置服务 gRPC 客户端
func NewAdminSettingsClient(r registry.Discovery, conf *conf.Server, auth *conf.Auth, tp trace.TracerProvider, metric *mid.Meter) userRpc.AdminSettingsClient {
	conn, err := grpc.NewClientConn(ts.NewGrpcClientOpt(
		&ts.TransporterOpt{
			Timeout:     conf.Grpc.Timeout,
			ServiceName: grpc.USER_SERVICE,
			Discovery:   r,
		},
		&ts.MiddlewareOpt{
			TracerProvider: tp,
			ServiceKey:     auth.ServiceKey[grpc.USER_SERVICE],
			Meter:          metric,
		},
	))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to %v.%v: %v", os.Getenv("SERVICE_ENV"), grpc.USER_SERVICE, err))
	}
	return userRpc.NewAdminSettingsClient(conn)
}
```

需要在 `client_conn.go` 中添加 `"fmt"` 和 `"os"` 的 import（如果还没有的话，检查现有代码发现已经引入了）。

- [ ] **Step 4: 注册 Data ProviderSet**

修改 `app/cin/admin/internal/data/data.go`，添加 `NewSettingsRepo` 和 `NewAdminSettingsClient`：

```go
var ProviderSet = wire.NewSet(
	NewData,
	NewDashboardRepo,
	NewProductRepo,
	NewCategoryRepo,
	NewAuthRepo,
	NewStaffRepo,
	NewUploadRepo,
	NewOrderRepo,
	NewCouponRepo,
	NewSettingsRepo,        // ← 新增
	NewProductServiceClient,
	NewCategoryServiceClient,
	NewAdminUserServiceClient,
	NewAdminRoleClient,
	NewAdminSettingsClient, // ← 新增
	NewOrderServiceClient,
	NewUserServiceClient,
	NewCouponServiceClient,
	NewTokenValidator,
	NewOssClient,
	NewRedisClient,
	ProvideUniversalClient,
	NewRoleRepo,
	NewRemoteEnforcer,
	ProvideEnforcerProvider,
	NewExportRepo,
	NewAsynqClient,
)
```

- [ ] **Step 5: 注册 Biz ProviderSet**

修改 `app/cin/admin/internal/biz/biz.go`：

```go
var ProviderSet = wire.NewSet(
	NewDashboardUseCase,
	NewProductUseCase,
	NewCategoryUseCase,
	NewAuthUseCase,
	NewRoleUseCase,
	NewStaffUseCase,
	NewOrderUseCase,
	NewExportUseCase,
	NewCouponUseCase,
	NewSettingsUseCase,  // ← 新增
	NewAgentUseCase,
	ProvideApiKey,
)
```

- [ ] **Step 6: Commit**

```bash
git add app/cin/admin/internal/biz/settings.go
git add app/cin/admin/internal/data/settings.go
git add app/cin/admin/internal/data/client_conn.go
git add app/cin/admin/internal/data/data.go
git add app/cin/admin/internal/biz/biz.go
git commit -m "feat(settings): 添加 Admin BFF 的 Data 和 Biz 层"
```

---

### Task 6: Admin BFF — Service 层 & Wire

**Files:**
- Modify: `app/cin/admin/internal/service/settings.go`

- [ ] **Step 1: 重写 Service 层**

`app/cin/admin/internal/service/settings.go`（完全替换原有内容）：

```go
package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// GetSettings 获取系统设置
func (s *AdminService) GetSettings(ctx context.Context, req *adminv1.GetSettingsRequest) (*adminv1.GetSettingsReply, error) {
	settings, err := s.sc.Get(ctx)
	if err != nil {
		return nil, err
	}
	payments := settings.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &adminv1.GetSettingsReply{
		ShopName:              settings.ShopName,
		ShopLogo:              settings.ShopLogo,
		ShopDescription:       settings.ShopDescription,
		ContactPhone:          settings.ContactPhone,
		ContactEmail:          settings.ContactEmail,
		IcpRecord:             settings.IcpRecord,
		AutoConfirmDays:       settings.AutoConfirmDays,
		AutoCloseMinutes:      settings.AutoCloseMinutes,
		RefundTimeoutDays:     settings.RefundTimeoutDays,
		StockAlertThreshold:   settings.StockAlertThreshold,
		NewProductOnline:      settings.NewProductOnline,
		ProductApproval:       settings.ProductApproval,
		NewOrderNotify:        settings.NewOrderNotify,
		RefundNotify:          settings.RefundNotify,
		UserRegisterNotify:    settings.UserRegisterNotify,
		StockAlertNotify:      settings.StockAlertNotify,
		LoginAttemptLimit:     settings.LoginAttemptLimit,
		LoginCaptcha:          settings.LoginCaptcha,
		SessionTimeoutMinutes: settings.SessionTimeoutMinutes,
		EnabledPayments:       payments,
		DefaultPayment:        settings.DefaultPayment,
		FreeShippingThreshold: settings.FreeShippingThreshold,
		DefaultShippingCompany: settings.DefaultShippingCompany,
	}, nil
}

// UpdateSettings 更新系统设置
func (s *AdminService) UpdateSettings(ctx context.Context, req *adminv1.UpdateSettingsRequest) (*adminv1.UpdateSettingsReply, error) {
	payments := req.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	settings := &biz.SystemSettings{
		ShopName:              req.ShopName,
		ShopLogo:              req.ShopLogo,
		ShopDescription:       req.ShopDescription,
		ContactPhone:          req.ContactPhone,
		ContactEmail:          req.ContactEmail,
		IcpRecord:             req.IcpRecord,
		AutoConfirmDays:       req.AutoConfirmDays,
		AutoCloseMinutes:      req.AutoCloseMinutes,
		RefundTimeoutDays:     req.RefundTimeoutDays,
		StockAlertThreshold:   req.StockAlertThreshold,
		NewProductOnline:      req.NewProductOnline,
		ProductApproval:       req.ProductApproval,
		NewOrderNotify:        req.NewOrderNotify,
		RefundNotify:          req.RefundNotify,
		UserRegisterNotify:    req.UserRegisterNotify,
		StockAlertNotify:      req.StockAlertNotify,
		LoginAttemptLimit:     req.LoginAttemptLimit,
		LoginCaptcha:          req.LoginCaptcha,
		SessionTimeoutMinutes: req.SessionTimeoutMinutes,
		EnabledPayments:       payments,
		DefaultPayment:        req.DefaultPayment,
		FreeShippingThreshold: req.FreeShippingThreshold,
		DefaultShippingCompany: req.DefaultShippingCompany,
	}
	if err := s.sc.Update(ctx, settings); err != nil {
		return nil, err
	}
	return &adminv1.UpdateSettingsReply{Success: true}, nil
}

- [ ] **Step 2: 注入 AdminService**

修改 `app/cin/admin/internal/service/service.go`，在 `AdminService` 结构体添加 `sc *biz.SettingsUseCase` 字段，并在 `NewAdminService` 构造函数中添加参数：

```go
type AdminService struct {
	// ... 其他字段不变
	sc        *biz.SettingsUseCase  // ← 新增
}

func NewAdminService(dc *biz.DashboardUseCase, pc *biz.ProductUseCase, ac *biz.AuthUseCase, rc *biz.RoleUseCase, sc *biz.StaffUseCase, cc *biz.CategoryUseCase, oc *biz.OrderUseCase, ec *biz.ExportUseCase, couponUC *biz.CouponUseCase, ur *data.UploadRepo, agentUC *biz.AgentUseCase, settingsUC *biz.SettingsUseCase, logger log.Logger) *AdminService {
	return &AdminService{
		// ... 其他字段不变
		sc:        settingsUC,
	}
}
```

注意：wire 自动装配中 `sc` 字段名会与 `*biz.SettingsUseCase` 类型冲突（`sc` 已被 `*biz.StaffUseCase` 占用），所以需要使用不同的字段名。在结构体中用 `settUC`：

```go
type AdminService struct {
	// ... 原有字段
	settUC    *biz.SettingsUseCase  // ← 新增
}

func NewAdminService(..., settingsUC *biz.SettingsUseCase, ...) *AdminService {
	return &AdminService{
		// ...
		settUC:   settingsUC,
	}
}
```

在 `settings.go` 的 `GetSettings` 和 `UpdateSettings` 方法中，`s.sc` 改为 `s.settUC`。

- [ ] **Step 3: 生成 Wire 代码**

Run: `cd app/cin/admin && make wire`
Expected: `wire_gen.go` 重新生成，编译通过。

- [ ] **Step 4: 验证编译**

Run: `cd app/cin/admin && go build ./...`
Expected: 编译成功，无错误。

- [ ] **Step 5: Commit**

```bash
git add app/cin/admin/internal/service/settings.go
git add app/cin/admin/internal/service/service.go
git commit -m "feat(settings): 重写 Admin BFF Service 层并配置 Wire"
```

---

### Task 7: 前端 API & SettingsView 重构

**Files:**
- Create: `web/admin/src/api/settings.ts`
- Rewrite: `web/admin/src/views/settings/SettingsView.vue`

- [ ] **Step 1: 创建前端 API**

`web/admin/src/api/settings.ts`:

```typescript
/**
 * 悦选商城管理后台 - 系统设置 API
 */
import request from './index'

/** 系统设置接口定义（与服务端 proto 对应） */
export interface SystemSettings {
  shop_name: string
  shop_logo: string
  shop_description: string
  contact_phone: string
  contact_email: string
  icp_record: string
  auto_confirm_days: number
  auto_close_minutes: number
  refund_timeout_days: number
  stock_alert_threshold: number
  new_product_online: boolean
  product_approval: boolean
  new_order_notify: boolean
  refund_notify: boolean
  user_register_notify: boolean
  stock_alert_notify: boolean
  login_attempt_limit: boolean
  login_captcha: boolean
  session_timeout_minutes: number
  enabled_payments: string[]
  default_payment: string
  free_shipping_threshold: number
  default_shipping_company: string
}

/** 获取系统设置 */
export function getSettings(): Promise<SystemSettings> {
  return request.get('/admin/v1/settings')
}

/** 更新系统设置 */
export function updateSettings(data: Partial<SystemSettings>): Promise<{ success: boolean }> {
  return request.put('/admin/v1/settings', data)
}
```

- [ ] **Step 2: 重写 SettingsView.vue**

完整重写 `web/admin/src/views/settings/SettingsView.vue`，包含 7 个 Tab 标签页：

**关键设计要求：**
- 使用 `el-tabs type="border-card"` 风格
- 每个 Tab 内部分组卡片（`el-card`），带错落入场动效
- Tab 切换使用 Transition 水平滑入
- 保存按钮带 loading 状态和成功反馈动效
- 响应式布局
- 调用真实的 API（`getSettings`, `updateSettings`）

具体代码较长，参考设计文档的 Tab 结构和 UI 控件映射表：
- 店铺设置：Input, Upload（Logo 上传调用 upload API）
- 订单设置：InputNumber
- 商品设置：InputNumber, Switch
- 通知设置：Switch
- 安全策略：Switch, InputNumber
- 支付设置：CheckboxGroup, Select
- 物流设置：InputNumber, Input

- [ ] **Step 3: 验证前端编译**

Run: `cd web/admin && npm run build`
Expected: 编译成功，无错误。

- [ ] **Step 4: Commit**

```bash
git add web/admin/src/api/settings.ts
git add web/admin/src/views/settings/SettingsView.vue
git commit -m "feat(settings): 重写前端设置页面并对接真实 API"
```

---

### Task 8: 集成验证与启动

**Files:** 无代码变更

- [ ] **Step 1: 创建 system_settings 表**

通过 User Service 自动迁移（非生产环境），或在数据库中手动执行：

```sql
CREATE TABLE IF NOT EXISTS `system_settings` (
    `id`         INT PRIMARY KEY DEFAULT 1 COMMENT '固定为 1，保证单行记录',
    `settings`   JSON NOT NULL COMMENT '所有配置项的 JSON 对象',
    `updated_at` BIGINT NOT NULL DEFAULT 0,
    `updated_by` VARCHAR(64) DEFAULT '' COMMENT '最后修改人'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: 启动 User Service**

Run: `cd app/service/user && go run cmd/user/ -conf configs/`
Expected: 启动成功，日志显示 AutoMigrate 添加 system_settings 表。

- [ ] **Step 3: 启动 Admin BFF**

Run: `cd app/cin/admin && go run cmd/admin/ -conf configs/`
Expected: 启动成功。

- [ ] **Step 4: 启动前端**

Run: `cd web/admin && npm run dev`
Expected: 开发服务器启动成功。

- [ ] **Step 5: 功能验证**

在浏览器中访问 `/settings`，验证：
1. 所有 7 个 Tab 都能正常加载默认配置数据
2. 修改某一设置项（如商城名称），点击保存，提示成功
3. 刷新页面，修改后数据保持持久化
4. 切换 Tab 动效流畅
5. 页面入场动效正常

- [ ] **Step 6: 最终 Commit**

```bash
git add .
git commit -m "feat(settings): 完成系统设置模块重新设计与验证"
```

---

## 执行计划

计划已保存到 `docs/superpowers/plans/2026-07-24-system-settings.md`。共 8 个 Task，按顺序执行：

**后端 (Task 1-6):** Proto → Model/Data → Biz → Service → Wire 注册 → Admin BFF
**前端 (Task 7):** API → SettingsView 重构
**验证 (Task 8):** 启动服务 → 功能测试