package middleware

import (
	"context"
	"errors"
	"testing"

	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/stretchr/testify/assert"
)

// mockEnforcerWithError 支持返回 GetFilteredPolicy 错误的 mock
type mockEnforcerWithError struct {
	enforceFn            func(rvals ...interface{}) (bool, error)
	filteredPolicies     [][]string
	getFilteredPolicyErr error
}

func (m *mockEnforcerWithError) Enforce(rvals ...interface{}) (bool, error) {
	return m.enforceFn(rvals...)
}

func (m *mockEnforcerWithError) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return m.filteredPolicies, m.getFilteredPolicyErr
}

// TestPermissionMiddleware_NoPolicyDefinition 测试当操作没有 Casbin 策略定义时，所有已认证用户应被放行
// 这是"无策略 = 允许所有"的关键安全逻辑路径
func TestPermissionMiddleware_NoPolicyDefinition(t *testing.T) {
	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			// 不应该被调用，因为没有策略定义时应直接放行
			t.Fatal("不应调用 Enforce：无策略定义时应直接放行")
			return false, nil
		},
		filteredPolicies: [][]string{}, // 空策略 = 无权限管控
	}

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(newContextWithRole("普通用户"), nil)

	assert.NoError(t, err, "无策略定义时已认证用户应被放行")
}

// TestPermissionMiddleware_GetFilteredPolicyError 测试当 GetFilteredPolicy 返回错误时，中间件应返回错误而非静默放行
func TestPermissionMiddleware_GetFilteredPolicyError(t *testing.T) {
	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return true, nil
		},
		getFilteredPolicyErr: errors.New("database connection lost"),
	}

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(newContextWithRole("超级管理员"), nil)

	assert.Error(t, err, "GetFilteredPolicy 出错时应返回错误")
	assert.Contains(t, err.Error(), "权限校验异常", "错误信息应包含权限校验异常")
}

// TestPermissionMiddleware_InvalidClaimsType 测试当 JWT claims 存在但类型不是 *UserClaims 时，应返回错误
func TestPermissionMiddleware_InvalidClaimsType(t *testing.T) {
	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return true, nil
		},
	}

	// 使用非 *UserClaims 类型的 claims（jwt2.RegisteredClaims 实现了 jwt.Claims 接口）
	ctx := jwt.NewContext(context.Background(), &jwt2.RegisteredClaims{})
	ctx = transport.NewServerContext(ctx, &mockTransport{
		operation: "/api/test",
		method:    "GET",
	})

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(ctx, nil)

	assert.Error(t, err, "claims 类型无效时应返回错误")
	assert.Contains(t, err.Error(), "无效的 claims 类型", "错误信息应提示 claims 类型无效")
}

// TestPermissionMiddleware_NilClaims 测试当 claims 为 *UserClaims 但值为 nil 时，应返回错误
func TestPermissionMiddleware_NilClaims(t *testing.T) {
	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return true, nil
		},
	}

	// 将 nil 作为 *UserClaims 传入（类型断言通过但值为 nil）
	var nilClaims *UserClaims = nil
	ctx := jwt.NewContext(context.Background(), nilClaims)
	ctx = transport.NewServerContext(ctx, &mockTransport{
		operation: "/api/test",
		method:    "GET",
	})

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(ctx, nil)

	assert.Error(t, err, "nil claims 应返回错误")
	assert.Contains(t, err.Error(), "无效的 claims 类型", "错误信息应提示 claims 类型无效")
}

// TestPermissionMiddleware_NoTransportContext 测试当 transport 上下文缺失时，应返回错误
func TestPermissionMiddleware_NoTransportContext(t *testing.T) {
	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			return true, nil
		},
	}

	// 只有 JWT claims，没有 transport 上下文
	ctx := jwt.NewContext(context.Background(), &UserClaims{
		ID:       1,
		Username: "test_user",
		Role:     "超级管理员",
	})

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(ctx, nil)

	assert.Error(t, err, "缺少 transport 上下文时应返回错误")
	assert.Contains(t, err.Error(), "无法获取请求上下文", "错误信息应提示无法获取请求上下文")
}

// TestPermissionMiddleware_XMethodHeader 测试当 X-Method 请求头存在时，应使用其值作为 Enforce 的 method 参数
func TestPermissionMiddleware_XMethodHeader(t *testing.T) {
	var capturedMethod interface{}

	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			if len(rvals) >= 3 {
				capturedMethod = rvals[2]
			}
			return true, nil
		},
		filteredPolicies: [][]string{{"超级管理员", "/api/test", "CUSTOM_METHOD"}},
	}

	// X-Method 设置为 "CUSTOM_METHOD"，operation 为 "/api/test"
	ctx := newContextWithTransport("超级管理员", "/api/test", "CUSTOM_METHOD")

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(ctx, nil)

	assert.NoError(t, err, "有权限时应放行")
	assert.Equal(t, "CUSTOM_METHOD", capturedMethod, "Enforce 的 method 参数应使用 X-Method 头的值")
}

// TestPermissionMiddleware_XMethodEmpty 测试当 X-Method 请求头为空时，应使用 operation 作为 method 参数
func TestPermissionMiddleware_XMethodEmpty(t *testing.T) {
	var capturedMethod interface{}

	enforcer := &mockEnforcerWithError{
		enforceFn: func(rvals ...interface{}) (bool, error) {
			if len(rvals) >= 3 {
				capturedMethod = rvals[2]
			}
			return true, nil
		},
		filteredPolicies: [][]string{{"超级管理员", "/api/test", "/api/test"}},
	}

	// X-Method 为空，operation 为 "/api/test"
	ctx := newContextWithTransport("超级管理员", "/api/test", "")

	mw := PermissionMiddleware(enforcer)
	_, err := mw(mockHandler)(ctx, nil)

	assert.NoError(t, err, "有权限时应放行")
	assert.Equal(t, "/api/test", capturedMethod, "X-Method 为空时 Enforce 的 method 参数应使用 operation")
}
