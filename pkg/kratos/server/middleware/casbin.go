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
	// 使用 variadic interface{} 以匹配 casbin.Enforcer 的签名
	Enforce(rvals ...interface{}) (bool, error)
}

// PermissionMiddleware 创建 Casbin 权限校验中间件
// 从 JWT claims 中提取用户角色，然后使用 Casbin Enforcer 校验权限
func PermissionMiddleware(enforcer EnforcerProvider) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从 JWT claims 中提取用户角色
			claims, ok := jwt.FromContext(ctx)
			if !ok {
				return nil, fmt.Errorf("权限校验失败: 无法获取 JWT claims")
			}
			userClaims, ok := claims.(*UserClaims)
			if !ok || userClaims == nil {
				return nil, fmt.Errorf("权限校验失败: 无效的 claims 类型")
			}

			// 从 transport context 中获取操作信息
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, fmt.Errorf("权限校验失败: 无法获取请求上下文")
			}

			// Operation 格式: /api.cin.admin.v1.AdminProduct/DeleteProduct
			operation := tr.Operation()
			// HTTP method（仅 HTTP 协议有，gRPC 用 operation 自身）
			method := tr.RequestHeader().Get("X-Method")
			if method == "" {
				method = operation
			}

			// 执行 Casbin 权限校验: Enforce(subject, object, action)
			ok, err := enforcer.Enforce(userClaims.Role, operation, method)
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
