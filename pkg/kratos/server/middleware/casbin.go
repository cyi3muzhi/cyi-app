package middleware

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
)

// EnforcerProvider 提供 Casbin Enforcer 的接口（方便测试时 mock）
type EnforcerProvider interface {
	// Enforce 判断用户是否有权限执行操作
	Enforce(rvals ...interface{}) (bool, error)
	// GetFilteredPolicy 获取过滤后的策略（用于检查操作是否有策略定义）
	GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error)
}

// PermissionMiddleware 创建 Casbin 权限校验中间件
// 从 JWT claims 中提取用户角色，然后使用 Casbin Enforcer 校验权限
// 无 JWT claims 时放行（如登录）
// 规则：如果某个 operation 没有任何 Casbin 策略定义，则允许所有登录用户访问；
// 如果有策略定义，则按 Enforce 结果判断
func PermissionMiddleware(enforcer EnforcerProvider) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			claims, ok := jwt.FromContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			userClaims, ok := claims.(*UserClaims)
			if !ok || userClaims == nil {
				return nil, fmt.Errorf("权限校验失败: 无效的 claims 类型")
			}

			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, fmt.Errorf("权限校验失败: 无法获取请求上下文")
			}

			operation := tr.Operation()
			method := tr.RequestHeader().Get("X-Method")
			if method == "" {
				method = operation
			}

			// 检查该 operation 是否有任何 Casbin 策略定义
			// 如果没有策略定义，说明该接口无需权限管控，直接放行
			policies, err := enforcer.GetFilteredPolicy(1, operation)
			if err != nil {
				return nil, fmt.Errorf("权限校验异常: %w", err)
			}
			if len(policies) == 0 {
				return handler(ctx, req)
			}

			// 有策略定义 → 执行 Casbin 权限校验
			ok, err = enforcer.Enforce(userClaims.Role, operation, method)
			if err != nil {
				return nil, fmt.Errorf("权限校验异常: %w", err)
			}
			if !ok {
				return nil, fmt.Errorf("无权限执行此操作: role=%s, operation=%s", userClaims.Role, operation)
			}

			return handler(ctx, req)
		}
	}
}

// GetRoleFromClaims 从 context 中提取用户角色（供业务代码使用）
func GetRoleFromClaims(ctx context.Context) (string, error) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("无法获取 JWT claims")
	}
	userClaims, ok := claims.(*UserClaims)
	if !ok || userClaims == nil {
		return "", fmt.Errorf("无效的 claims 类型")
	}
	return userClaims.Role, nil
}

// Ensure casbin.Enforcer implements EnforcerProvider
var _ EnforcerProvider = (*casbin.Enforcer)(nil)
