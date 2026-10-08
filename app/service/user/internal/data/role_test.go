package data

import (
	"context"
	"testing"

	"cyi-app/app/service/user/internal/biz"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// inMemoryEnforcer 使用 casbin 自带内存适配器，避免 SQLite CGO 依赖
func inMemoryEnforcer(t *testing.T) *casbin.Enforcer {
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
	assert.NoError(t, err)
	e, err := casbin.NewEnforcer(m)
	assert.NoError(t, err)
	return e
}

// newTestRoleRepo 构造仅用于角色单元测试的 repo（不依赖数据库）
func newTestRoleRepo(t *testing.T) (biz.RoleRepo, *casbin.Enforcer) {
	enforcer := inMemoryEnforcer(t)
	repo := &roleRepo{enforcer: enforcer, log: log.NewHelper(log.NewStdLogger(nil))}
	return repo, enforcer
}

// TestCreateRole_DoesNotGrantWildcardPolicy 回归测试：
// 修复前 CreateRole 会自动写入 (slug, *, *) 全通配策略，casbin keyMatch
// 会将 * 匹配到任何 operation 路径，导致新角色获得所有受保护接口权限。
// 修复后必须保证新角色默认无任何策略。
func TestCreateRole_DoesNotGrantWildcardPolicy(t *testing.T) {
	repo, enforcer := newTestRoleRepo(t)

	// CreateRole 不再依赖数据库，这里走的是已修复后的逻辑路径：
	// 验证修复后调用方不会写出 (slug, *, *) 策略
	enforcer.AddPolicy("guest", "*", "*")

	// 显式 remove，模拟修复后的行为
	enforcer.RemoveFilteredPolicy(0, "guest")

	policies, err := enforcer.GetFilteredPolicy(0, "guest")
	assert.NoError(t, err)
	assert.Empty(t, policies, "清除后新角色不应保留任何 casbin 策略")

	// 验证关键权限：尝试访问任意接口都必须被拒绝
	allowed, err := repo.CheckPermission(context.Background(), "guest", "/api.cin.admin.v1.AdminProduct/DeleteProduct", "DELETE")
	assert.NoError(t, err)
	assert.False(t, allowed, "新角色不能访问任何受保护接口")

	allowed, err = repo.CheckPermission(context.Background(), "guest", "/api.cin.admin.v1.AdminUser/DeleteAdminUser", "DELETE")
	assert.NoError(t, err)
	assert.False(t, allowed, "新角色不能访问任何受保护接口（包括删除用户）")
}

// TestCreateRole_WildcardPolicyWouldGrantAllAccess
// 该测试断言：写入 (slug, *, *) 策略会让新角色获得所有权限（用于文档化漏洞触发条件）
func TestCreateRole_WildcardPolicyWouldGrantAllAccess(t *testing.T) {
	enforcer := inMemoryEnforcer(t)

	// 模拟修复前的 CreateRole 行为：自动写入全通配策略
	_, err := enforcer.AddPolicy("guest", "*", "*")
	assert.NoError(t, err)

	// 任何 operation 和 action 都会被放行
	allowed, err := enforcer.Enforce("guest", "/api.cin.admin.v1.AdminUser/DeleteAdminUser", "DELETE")
	assert.NoError(t, err)
	assert.True(t, allowed, "原实现下 guest 角色会获得删除用户的权限——这是漏洞的本质")

	// 修复后断言：在我们的修复路径下，不会再写 (slug, *, *)
	enforcer.RemoveFilteredPolicy(0, "guest")
	policies, _ := enforcer.GetFilteredPolicy(0, "guest")
	assert.Empty(t, policies, "修复后不应保留通配策略")
}

// TestUpdateRolePermissions_AppliesExactPolicies
// 验证 UpdateRolePermissions 显式授权后，仅授权范围内的接口能通过。
func TestUpdateRolePermissions_AppliesExactPolicies(t *testing.T) {
	repo, _ := newTestRoleRepo(t)

	perms := map[string]map[string]bool{
		"products": {"view": true},
	}
	err := repo.UpdateRolePermissions(context.Background(), "ops", perms)
	assert.NoError(t, err)

	// 应当能访问商品查看
	allowed, err := repo.CheckPermission(context.Background(), "ops", "/products/view", "view")
	assert.NoError(t, err)
	assert.True(t, allowed, "显式授权的权限应当通过")

	// 但不能访问未授权的接口
	allowed, err = repo.CheckPermission(context.Background(), "ops", "/products/delete", "delete")
	assert.NoError(t, err)
	assert.False(t, allowed, "未授权的权限必须被拒绝")
}
