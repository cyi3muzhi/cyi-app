package data

import (
	"context"
	"fmt"

	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// NewCasbinEnforcer 创建 Casbin Enforcer（简化版，仅加载模型，无策略适配器）
// 策略通过 gRPC 从 user 服务获取，此 enforcer 仅作为本地模型参考
func NewCasbinEnforcer() (*casbin.Enforcer, error) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("加载 Casbin 模型失败: %w", err)
	}
	// 创建空 Enforcer（无策略适配器，策略从 user 服务获取）
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
	}
	return e, nil
}

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

// ProvideEnforcerProvider 将 *RemoteEnforcer 转换为 EnforcerProvider 接口（用于 wire 注入）
func ProvideEnforcerProvider(e *RemoteEnforcer) mid.EnforcerProvider {
	return e
}
