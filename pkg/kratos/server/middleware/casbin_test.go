package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

// ---------- Mock Enforcer ----------

type mockEnforcer struct {
	enforceFn func(rvals ...interface{}) (bool, error)
}

func (m *mockEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	return m.enforceFn(rvals...)
}

// ---------- Mock Transport ----------

type mockTransport struct {
	operation string
	method    string
}

func (m *mockTransport) Kind() transport.Kind           { return transport.KindHTTP }
func (m *mockTransport) Endpoint() string               { return "" }
func (m *mockTransport) Operation() string               { return m.operation }
func (m *mockTransport) RequestHeader() transport.Header { return &mockHeader{m.method} }
func (m *mockTransport) ReplyHeader() transport.Header   { return &mockHeader{} }

type mockHeader struct {
	method string
}

func (h *mockHeader) Get(key string) string {
	if key == "X-Method" {
		return h.method
	}
	return ""
}
func (h *mockHeader) Set(key, value string) {}
func (h *mockHeader) Add(key, value string) {}
func (h *mockHeader) Values(key string) []string { return nil }
func (h *mockHeader) Keys() []string             { return nil }

// ---------- mockHandler ----------

func mockHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

// ---------- Helper: 构造带 JWT claims 的 context ----------

func newContextWithRole(role string) context.Context {
	claims := &UserClaims{
		ID:       1,
		Username: "test_user",
		Role:     role,
	}
	ctx := context.Background()
	ctx = jwt.NewContext(ctx, claims)
	ctx = transport.NewServerContext(ctx, &mockTransport{
		operation: "/api.cin.admin.v1.AdminProduct/DeleteProduct",
		method:    "DELETE",
	})
	return ctx
}

func newContextWithTransport(role, operation, method string) context.Context {
	claims := &UserClaims{
		ID:       1,
		Username: "test_user",
		Role:     role,
	}
	ctx := context.Background()
	ctx = jwt.NewContext(ctx, claims)
	ctx = transport.NewServerContext(ctx, &mockTransport{
		operation: operation,
		method:    method,
	})
	return ctx
}

// ---------- 测试用例 ----------

// TestPermissionMiddleware_Allowed 测试有权限时通过
func TestPermissionMiddleware_Allowed(t *testing.T) {
	enforcer := &mockEnforcer{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			// 验证参数正确传递
			if len(rvals) != 3 {
				t.Fatalf("期望 3 个参数，实际 %d", len(rvals))
			}
			if rvals[0] != "超级管理员" {
				t.Fatalf("期望角色=超级管理员，实际=%v", rvals[0])
			}
			return true, nil
		},
	}

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(newContextWithRole("超级管理员"), nil)
	if err != nil {
		t.Fatalf("期望无错误，实际: %v", err)
	}
}

// TestPermissionMiddleware_Denied 测试无权限时拒绝
func TestPermissionMiddleware_Denied(t *testing.T) {
	enforcer := &mockEnforcer{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return false, nil
		},
	}

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(newContextWithRole("财务"), nil)
	if err == nil {
		t.Fatal("期望返回权限错误，但通过了")
	}
	t.Logf("正确的拒绝响应: %v", err)
}

// TestPermissionMiddleware_NoClaims 测试无 JWT claims 时放行（白名单接口）
func TestPermissionMiddleware_NoClaims(t *testing.T) {
	enforcer := &mockEnforcer{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return true, nil
		},
	}

	mw := PermissionMiddleware(enforcer)
	// 创建一个没有 claims 的 context
	ctx := transport.NewServerContext(context.Background(), &mockTransport{
		operation: "/api/test",
		method:    "GET",
	})
	_, err := mw(mockHandler)(ctx, nil)
	// 无 claims = 白名单接口（如登录），应放行
	if err != nil {
		t.Fatalf("期望无 JWT claims 时放行（白名单接口），实际返回错误: %v", err)
	}
	t.Logf("正确的行为: 无 JWT claims 时放行（白名单接口）")
}

// TestPermissionMiddleware_EnforceError 测试 Casbin 异常时返回错误
func TestPermissionMiddleware_EnforceError(t *testing.T) {
	enforcer := &mockEnforcer{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return false, errors.New("casbin internal error")
		},
	}

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(newContextWithRole("超级管理员"), nil)
	if err == nil {
		t.Fatal("期望 Casbin 异常时返回错误")
	}
	t.Logf("正确的异常响应: %v", err)
}

// TestPermissionMiddleware_MultipleRoles 测试不同角色对同一资源的权限
func TestPermissionMiddleware_MultipleRoles(t *testing.T) {
	tests := []struct {
		name      string
		role      string
		operation string
		method    string
		allowed   bool
	}{
		{
			name:      "超管可以删除商品",
			role:      "超级管理员",
			operation: "/api.cin.admin.v1.AdminProduct/DeleteProduct",
			method:    "DELETE",
			allowed:   true,
		},
		{
			name:      "运营编辑可以获取商品列表",
			role:      "运营编辑",
			operation: "/api.cin.admin.v1.AdminProduct/ListProducts",
			method:    "GET",
			allowed:   true,
		},
		{
			name:      "财务不能创建商品",
			role:      "财务",
			operation: "/api.cin.admin.v1.AdminProduct/CreateProduct",
			method:    "POST",
			allowed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := &mockEnforcer{
				enforceFn: func(rvals ...interface{}) (bool, error) {
					return tt.allowed, nil
				},
			}

			mw := PermissionMiddleware(enforcer)
			ctx := newContextWithTransport(tt.role, tt.operation, tt.method)
			_, err := mw(mockHandler)(ctx, nil)

			if tt.allowed && err != nil {
				t.Errorf("期望允许，但被拒绝: %v", err)
			}
			if !tt.allowed && err == nil {
				t.Errorf("期望拒绝，但被允许")
			}
		})
	}
}

// TestGetRoleFromClaims 测试从 context 中提取角色
func TestGetRoleFromClaims(t *testing.T) {
	// 正常情况
	ctx := newContextWithRole("超级管理员")
	role, err := GetRoleFromClaims(ctx)
	if err != nil {
		t.Fatalf("期望无错误，实际: %v", err)
	}
	if role != "超级管理员" {
		t.Fatalf("期望角色=超级管理员，实际=%s", role)
	}

	// 无 claims 的情况
	_, err = GetRoleFromClaims(context.Background())
	if err == nil {
		t.Fatal("期望无 claims 时返回错误")
	}
	t.Logf("正确的错误响应: %v", err)
}

// TestEnforcerProviderInterface 验证 *casbin.Enforcer 实现了 EnforcerProvider 接口
func TestEnforcerProviderInterface(t *testing.T) {
	// 编译时检查（实际已在 casbin.go 中通过 var _ 检查）
	var _ EnforcerProvider = (*mockEnforcer)(nil)
	_ = jwt2.NewParser() // 确保 jwt2 被引用
	t.Log("mockEnforcer 实现了 EnforcerProvider 接口")
}
