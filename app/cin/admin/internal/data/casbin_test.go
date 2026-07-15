package data

import (
	"path/filepath"
	"testing"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// newTestEnforcer 创建测试用的 Casbin Enforcer（直接从文件加载）
func newTestEnforcer(t *testing.T) *casbin.Enforcer {
	t.Helper()

	configDir := GetConfigDir()
	modelPath := filepath.Join(configDir, "rbac_model.conf")
	policyPath := filepath.Join(configDir, "rbac_policy.csv")

	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		t.Fatalf("加载模型失败: %v", err)
	}

	a := fileadapter.NewAdapter(policyPath)
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		t.Fatalf("创建 Enforcer 失败: %v", err)
	}
	return e
}

// TestCasbinEnforcer_Load 测试 Enforcer 能正确加载模型和策略
func TestCasbinEnforcer_Load(t *testing.T) {
	e := newTestEnforcer(t)
	if e == nil {
		t.Fatal("Enforcer 不应为 nil")
	}
	t.Log("Casbin Enforcer 初始化成功")
}

// TestCasbinEnforcer_SuperAdmin 测试超级管理员拥有所有权限
func TestCasbinEnforcer_SuperAdmin(t *testing.T) {
	e := newTestEnforcer(t)

	tests := []struct {
		name      string
		operation string
		method    string
	}{
		{"删除商品", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE"},
		{"创建商品", "/api.cin.admin.v1.AdminProduct/CreateProduct", "POST"},
		{"查看订单", "/api.cin.admin.v1.AdminOrder/GetOrder", "GET"},
		{"其他任意操作", "/api.some.other/AnyAction", "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := e.Enforce("超级管理员", tt.operation, tt.method)
			if err != nil {
				t.Fatalf("Enforce 失败: %v", err)
			}
			if !ok {
				t.Errorf("超级管理员应该能执行 %s(%s)", tt.operation, tt.method)
			}
		})
	}
}

// TestCasbinEnforcer_Admin 测试管理员的权限
func TestCasbinEnforcer_Admin(t *testing.T) {
	e := newTestEnforcer(t)

	tests := []struct {
		name      string
		operation string
		method    string
		allowed   bool
	}{
		{"查看仪表盘 - 允许", "/api.cin.admin.v1.Dashboard/GetDashboardStats", "GET", true},
		{"查看商品列表 - 允许", "/api.cin.admin.v1.AdminProduct/ListProducts", "GET", true},
		{"创建商品 - 允许", "/api.cin.admin.v1.AdminProduct/CreateProduct", "POST", true},
		{"查看订单 - 允许", "/api.cin.admin.v1.AdminOrder/GetOrder", "GET", true},
		{"删除商品 - 允许（匹配 AdminProduct/* 通配符）", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := e.Enforce("管理员", tt.operation, tt.method)
			if err != nil {
				t.Fatalf("Enforce 失败: %v", err)
			}
			if ok != tt.allowed {
				t.Errorf("管理员对 %s(%s) 期望=%v, 实际=%v",
					tt.operation, tt.method, tt.allowed, ok)
			}
		})
	}
}

// TestCasbinEnforcer_Editor 测试运营编辑的权限
func TestCasbinEnforcer_Editor(t *testing.T) {
	e := newTestEnforcer(t)

	tests := []struct {
		name      string
		operation string
		method    string
		allowed   bool
	}{
		{"查看商品 - 允许", "/api.cin.admin.v1.AdminProduct/ListProducts", "GET", true},
		{"创建商品 - 允许", "/api.cin.admin.v1.AdminProduct/CreateProduct", "POST", true},
		{"编辑商品 - 允许", "/api.cin.admin.v1.AdminProduct/UpdateProduct", "PUT", true},
		{"删除商品 - 拒绝", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE", false},
		{"查看订单 - 允许", "/api.cin.admin.v1.AdminOrder/GetOrder", "GET", true},
		{"查看仪表盘 - 拒绝", "/api.cin.admin.v1.Dashboard/GetDashboardStats", "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := e.Enforce("运营编辑", tt.operation, tt.method)
			if err != nil {
				t.Fatalf("Enforce 失败: %v", err)
			}
			if ok != tt.allowed {
				t.Errorf("运营编辑对 %s(%s) 期望=%v, 实际=%v",
					tt.operation, tt.method, tt.allowed, ok)
			}
		})
	}
}

// TestCasbinEnforcer_Finance 测试财务的权限
func TestCasbinEnforcer_Finance(t *testing.T) {
	e := newTestEnforcer(t)

	tests := []struct {
		name      string
		operation string
		method    string
		allowed   bool
	}{
		{"查看订单 - 允许", "/api.cin.admin.v1.AdminOrder/GetOrder", "GET", true},
		{"查看商品 - 拒绝", "/api.cin.admin.v1.AdminProduct/ListProducts", "GET", false},
		{"创建商品 - 拒绝", "/api.cin.admin.v1.AdminProduct/CreateProduct", "POST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := e.Enforce("财务", tt.operation, tt.method)
			if err != nil {
				t.Fatalf("Enforce 失败: %v", err)
			}
			if ok != tt.allowed {
				t.Errorf("财务对 %s(%s) 期望=%v, 实际=%v",
					tt.operation, tt.method, tt.allowed, ok)
			}
		})
	}
}

// TestCasbinEnforcer_UnknownRole 测试未定义角色的权限（应被拒绝）
func TestCasbinEnforcer_UnknownRole(t *testing.T) {
	e := newTestEnforcer(t)

	ok, err := e.Enforce("未知角色", "/api.cin.admin.v1.AdminProduct/ListProducts", "GET")
	if err != nil {
		t.Fatalf("Enforce 失败: %v", err)
	}
	if ok {
		t.Error("未知角色不应有任何权限")
	}
}

// TestProvideEnforcerProvider 测试接口适配器
func TestProvideEnforcerProvider(t *testing.T) {
	e := newTestEnforcer(t)
	provider := ProvideEnforcerProvider(e)

	// 验证返回的确实是 EnforcerProvider 接口
	if provider == nil {
		t.Fatal("provider 不应为 nil")
	}

	// 验证接口能正常使用
	ok, err := provider.Enforce("超级管理员", "/api/test/any", "GET")
	if err != nil {
		t.Fatalf("Enforce 失败: %v", err)
	}
	if !ok {
		t.Error("超级管理员应该有任意操作的权限")
	}

	// 编译时检查
	var _ mid.EnforcerProvider = provider
	t.Log("ProvideEnforcerProvider 正确返回 EnforcerProvider 接口")
}

// TestGetConfigDir 测试配置文件目录定位
func TestGetConfigDir(t *testing.T) {
	configDir := GetConfigDir()
	if configDir == "" {
		t.Fatal("配置目录不应为空")
	}
	t.Logf("配置目录: %s", configDir)
}
