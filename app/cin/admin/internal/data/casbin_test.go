package data

import (
	"context"
	"errors"
	"testing"

	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
)

// mockRoleRepo 用于测试 RemoteEnforcer 的 mock 实现
type mockRoleRepo struct {
	checkPermissionFn func(ctx context.Context, role, obj, act string) (bool, error)
}

func (m *mockRoleRepo) CheckPermission(ctx context.Context, role, obj, act string) (bool, error) {
	if m.checkPermissionFn != nil {
		return m.checkPermissionFn(ctx, role, obj, act)
	}
	return false, nil
}

func (m *mockRoleRepo) ListRoles(_ context.Context) ([]*biz.RoleItem, error) {
	return nil, nil
}

func (m *mockRoleRepo) CreateRole(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRoleRepo) UpdateRole(_ context.Context, _, _, _ string) error {
	return nil
}

func (m *mockRoleRepo) DeleteRole(_ context.Context, _ string) error {
	return nil
}

func (m *mockRoleRepo) GetRolePermissions(_ context.Context, _ string) (map[string]map[string]bool, error) {
	return nil, nil
}

func (m *mockRoleRepo) UpdateRolePermissions(_ context.Context, _ string, _ map[string]map[string]bool) error {
	return nil
}

func (m *mockRoleRepo) ListAllPermissions(_ context.Context) []biz.PermModule {
	return nil
}

// TestNewCasbinEnforcer_Success 测试创建 Casbin Enforcer 成功
func TestNewCasbinEnforcer_Success(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()

	assert.NoError(t, err, "创建 Casbin Enforcer 不应返回错误")
	assert.NotNil(t, enforcer, "返回的 Enforcer 不应为 nil")

	// 验证模型已正确加载
	model := enforcer.GetModel()
	assert.NotNil(t, model, "Casbin 模型不应为 nil")
}

// TestNewCasbinEnforcer_ModelValidation 测试模型定义的正确性
func TestNewCasbinEnforcer_ModelValidation(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()
	assert.NoError(t, err)

	// 测试 RBAC 模型基本功能：角色继承
	// 添加策略：admin 角色可以访问所有资源
	enforcer.AddPolicy("admin", "*", "*")
	// 添加角色继承：user 继承 admin 的权限
	enforcer.AddGroupingPolicy("user", "admin")

	// 验证角色继承生效
	allowed, err := enforcer.Enforce("user", "/api/v1/resource", "GET")
	assert.NoError(t, err)
	assert.True(t, allowed, "user 继承 admin 权限后应能访问所有资源")
}

// TestRemoteEnforcer_Enforce_Success 测试远程权限校验成功
func TestRemoteEnforcer_Enforce_Success(t *testing.T) {
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			assert.Equal(t, "超级管理员", role, "角色参数应正确传递")
			assert.Equal(t, "/api.cin.admin.v1.AdminProduct/DeleteProduct", obj, "对象参数应正确传递")
			assert.Equal(t, "DELETE", act, "操作参数应正确传递")
			return true, nil
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	allowed, err := enforcer.Enforce("超级管理员", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE")

	assert.NoError(t, err, "权限校验不应返回错误")
	assert.True(t, allowed, "有权限时应返回 true")
}

// TestRemoteEnforcer_Enforce_Denied 测试远程权限校验被拒绝
func TestRemoteEnforcer_Enforce_Denied(t *testing.T) {
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			return false, nil
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	allowed, err := enforcer.Enforce("财务", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE")

	assert.NoError(t, err, "权限校验不应返回错误")
	assert.False(t, allowed, "无权限时应返回 false")
}

// TestRemoteEnforcer_Enforce_Error 测试远程权限校验异常
func TestRemoteEnforcer_Enforce_Error(t *testing.T) {
	expectedErr := errors.New("gRPC call failed")
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			return false, expectedErr
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	allowed, err := enforcer.Enforce("超级管理员", "/api/v1/resource", "GET")

	assert.Error(t, err, "gRPC 调用失败应返回错误")
	assert.Equal(t, expectedErr, err, "应返回原始错误")
	assert.False(t, allowed, "错误时应返回 false")
}

// TestRemoteEnforcer_Enforce_InsufficientArgs 测试参数不足的情况
func TestRemoteEnforcer_Enforce_InsufficientArgs(t *testing.T) {
	mockRepo := &mockRoleRepo{}
	enforcer := NewRemoteEnforcer(mockRepo)

	// 少于 3 个参数
	allowed, err := enforcer.Enforce("超级管理员", "/api/v1/resource")

	assert.NoError(t, err, "参数不足不应返回错误（应返回 false）")
	assert.False(t, allowed, "参数不足时应返回 false")
}

// TestRemoteEnforcer_Enforce_TypeAssertion 测试类型断言处理
func TestRemoteEnforcer_Enforce_TypeAssertion(t *testing.T) {
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			// 验证类型断言后的值
			assert.IsType(t, "string", role)
			assert.IsType(t, "string", obj)
			assert.IsType(t, "string", act)
			return true, nil
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	// 传递 interface{} 类型的参数
	allowed, err := enforcer.Enforce(
		interface{}("admin"),
		interface{}("/api/v1/resource"),
		interface{}("GET"),
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

// TestProvideEnforcerProvider 测试依赖注入转换函数
func TestProvideEnforcerProvider(t *testing.T) {
	mockRepo := &mockRoleRepo{}
	enforcer := NewRemoteEnforcer(mockRepo)

	provider := ProvideEnforcerProvider(enforcer)

	assert.NotNil(t, provider, "返回的 EnforcerProvider 不应为 nil")

	// 验证转换后的接口仍然可以正常调用
	_, err := provider.Enforce("test", "test", "test")
	assert.NoError(t, err)
}

// TestRemoteEnforcer_ContextHandling 测试 context 处理
func TestRemoteEnforcer_ContextHandling(t *testing.T) {
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			// 验证 context 不为 nil（即使是 background context）
			assert.NotNil(t, ctx, "传递的 context 不应为 nil")
			return true, nil
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	// RemoteEnforcer 使用 background context，验证不会 panic
	allowed, err := enforcer.Enforce("admin", "/api/v1/resource", "GET")

	assert.NoError(t, err)
	assert.True(t, allowed)
}

// TestRemoteEnforcer_ConcurrentCalls 测试并发调用安全性
func TestRemoteEnforcer_ConcurrentCalls(t *testing.T) {
	mockRepo := &mockRoleRepo{
		checkPermissionFn: func(ctx context.Context, role, obj, act string) (bool, error) {
			return true, nil
		},
	}

	enforcer := NewRemoteEnforcer(mockRepo)

	// 并发调用 100 次
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			allowed, err := enforcer.Enforce("admin", "/api/v1/resource", "GET")
			assert.NoError(t, err)
			assert.True(t, allowed)
			done <- true
		}()
	}

	// 等待所有并发调用完成
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestNewCasbinEnforcer_MatcherFunction 测试 Casbin 匹配器函数
func TestNewCasbinEnforcer_MatcherFunction(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()
	assert.NoError(t, err)

	// 添加策略：admin 可以访问所有 product 相关的接口
	enforcer.AddPolicy("admin", "/api.cin.admin.v1.AdminProduct/*", "*")

	// keyMatch 应该匹配通配符
	tests := []struct {
		name     string
		role     string
		obj      string
		act      string
		expected bool
	}{
		{
			name:     "通配符匹配 - ListProducts",
			role:     "admin",
			obj:      "/api.cin.admin.v1.AdminProduct/ListProducts",
			act:      "GET",
			expected: true,
		},
		{
			name:     "通配符匹配 - DeleteProduct",
			role:     "admin",
			obj:      "/api.cin.admin.v1.AdminProduct/DeleteProduct",
			act:      "DELETE",
			expected: true,
		},
		{
			name:     "不匹配其他资源",
			role:     "admin",
			obj:      "/api.cin.admin.v1.AdminUser/DeleteAdminUser",
			act:      "DELETE",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(tt.role, tt.obj, tt.act)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// TestRemoteEnforcer_InterfaceCompliance 验证 RemoteEnforcer 实现了 EnforcerProvider 接口
func TestRemoteEnforcer_InterfaceCompliance(t *testing.T) {
	// 编译时检查接口实现
	var _ mid.EnforcerProvider = (*RemoteEnforcer)(nil)
	t.Log("RemoteEnforcer 实现了 EnforcerProvider 接口")
}

// TestNewCasbinEnforcer_InterfaceCompliance 验证 casbin.Enforcer 实现了 EnforcerProvider 接口
func TestNewCasbinEnforcer_InterfaceCompliance(t *testing.T) {
	// 编译时检查接口实现（已在 casbin.go 中验证）
	var _ mid.EnforcerProvider = (*casbin.Enforcer)(nil)
	t.Log("*casbin.Enforcer 实现了 EnforcerProvider 接口")
}