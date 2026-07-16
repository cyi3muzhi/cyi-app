package data

import (
	"context"

	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"
)

// RemoteEnforcer 远程权限校验器，通过 gRPC 调用 user 服务的 CheckPermission
// 实现 mid.EnforcerProvider 接口，供中间件使用
type RemoteEnforcer struct {
	checker biz.RoleRepo
}

// NewRemoteEnforcer 创建 RemoteEnforcer
func NewRemoteEnforcer(checker biz.RoleRepo) *RemoteEnforcer {
	return &RemoteEnforcer{checker: checker}
}

// Enforce 实现 EnforcerProvider 接口，通过 gRPC 调用 user 服务进行权限校验
func (e *RemoteEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	if len(rvals) < 3 {
		return false, nil
	}
	role, _ := rvals[0].(string)
	obj, _ := rvals[1].(string)
	act, _ := rvals[2].(string)
	// 使用 background context，中间件的 context 无法传递到这里
	return e.checker.CheckPermission(context.Background(), role, obj, act)
}

// GetFilteredPolicy 实现 EnforcerProvider 接口，检查 operation 是否有策略定义
func (e *RemoteEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	if len(fieldValues) == 0 {
		return nil, nil
	}
	has, err := e.checker.HasPolicyForOperation(context.Background(), fieldValues[0])
	if err != nil {
		return nil, err
	}
	if has {
		return [][]string{{}}, nil // 非空表示有策略
	}
	return nil, nil
}

// ProvideEnforcerProvider 将 *RemoteEnforcer 转换为 EnforcerProvider 接口（用于 wire 注入）
func ProvideEnforcerProvider(e *RemoteEnforcer) mid.EnforcerProvider {
	return e
}
