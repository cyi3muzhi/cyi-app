package data

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"cyi-app/app/service/user/internal/biz"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// roleRepo 角色数据仓库（封装 Casbin 策略操作）
type roleRepo struct {
	enforcer *casbin.Enforcer
	db       *gorm.DB
	log      *log.Helper
}

// NewRoleRepo 创建 RoleRepo
func NewRoleRepo(enforcer *casbin.Enforcer, db *gorm.DB, logger log.Logger) biz.RoleRepo {
	return &roleRepo{
		enforcer: enforcer,
		db:       db,
		log:      log.NewHelper(logger),
	}
}

// permModules 预定义权限模块（与前端一致）
var permModules = []biz.PermModule{
	{Key: "products", Title: "商品管理", Items: []biz.PermItem{
		{Key: "view", Label: "查看商品列表"}, {Key: "add", Label: "添加商品"},
		{Key: "edit", Label: "编辑商品"}, {Key: "delete", Label: "删除商品"},
		{Key: "publish", Label: "上架/下架"},
	}},
	{Key: "orders", Title: "订单管理", Items: []biz.PermItem{
		{Key: "view", Label: "查看订单"}, {Key: "edit", Label: "编辑订单"},
		{Key: "refund", Label: "退款处理"}, {Key: "ship", Label: "发货操作"},
		{Key: "cancel", Label: "取消订单"},
	}},
	{Key: "users", Title: "用户管理", Items: []biz.PermItem{
		{Key: "view", Label: "查看用户"}, {Key: "edit", Label: "编辑用户"},
		{Key: "freeze", Label: "冻结/解冻"}, {Key: "export", Label: "导出数据"},
	}},
	{Key: "settings", Title: "系统设置", Items: []biz.PermItem{
		{Key: "basic", Label: "基本设置"}, {Key: "security", Label: "安全设置"},
		{Key: "notification", Label: "通知设置"}, {Key: "logs", Label: "操作日志查看"},
	}},
	{Key: "roles", Title: "权限管理", Items: []biz.PermItem{
		{Key: "view", Label: "查看角色"}, {Key: "add", Label: "添加角色"},
		{Key: "edit", Label: "编辑角色"}, {Key: "delete", Label: "删除角色"},
	}},
	{Key: "coupons", Title: "优惠券管理", Items: []biz.PermItem{
		{Key: "view", Label: "查看优惠券"}, {Key: "create", Label: "创建优惠券"},
		{Key: "edit", Label: "编辑优惠券"}, {Key: "void", Label: "作废优惠券"},
	}},
	{Key: "data", Title: "数据查看", Items: []biz.PermItem{
		{Key: "dashboard", Label: "查看仪表盘"}, {Key: "report", Label: "查看报表"},
		{Key: "export", Label: "导出报表"},
	}},
}

// ListRoles 获取所有角色
func (r *roleRepo) ListRoles(ctx context.Context) ([]*biz.RoleItem, error) {
	roleNames := make(map[string]bool)
	policies, _ := r.enforcer.GetPolicy()
	for _, rule := range policies {
		if len(rule) >= 1 {
			roleNames[rule[0]] = true
		}
	}
	groupPolicies, _ := r.enforcer.GetGroupingPolicy()
	for _, rule := range groupPolicies {
		if len(rule) >= 2 {
			roleNames[rule[1]] = true
		}
	}

	// 统计每个角色的用户数（从 g 策略中统计分配给各角色的用户数）
	userCount := make(map[string]int)
	for _, rule := range groupPolicies {
		if len(rule) >= 2 {
			// V0 是数字（用户ID）才算用户分配
			if _, err := strconv.Atoi(rule[0]); err == nil {
				userCount[rule[1]]++
			}
		}
	}

	// 加载角色元数据（按 ID 排序）
	var metas []mod.RoleMeta
	if err := r.db.Order("id asc").Find(&metas).Error; err != nil {
		r.log.Warnf("加载角色元数据失败: %v", err)
	}
	metaOrder := make(map[string]int) // slug → 排序序号
	metaMap := make(map[string]mod.RoleMeta)
	for i, m := range metas {
		metaOrder[m.Slug] = i
		metaMap[m.Slug] = m
		// role_meta 中存在的角色，即使还没有 Casbin 策略也加入列表
		roleNames[m.Slug] = true
	}

	var roles []*biz.RoleItem
	for name := range roleNames {
		isSystem := name == "super_admin"
		meta, hasMeta := metaMap[name]
		displayName := name
		displayDesc := name + "角色"
		var createdAt string
		if hasMeta {
			displayName = meta.Name
			displayDesc = meta.Desc
			createdAt = meta.CreatedAt.Format("2006-01-02")
		}
		roles = append(roles, &biz.RoleItem{
			Name:      displayName,
			Slug:      name,
			Desc:      displayDesc,
			Users:     userCount[name],
			Enabled:   true,
			IsSystem:  isSystem,
			CreatedAt: createdAt,
		})
	}
	if roles == nil {
		roles = []*biz.RoleItem{}
	}

	// 按 role_meta ID 排序（有元数据的按 ID 先后，无元数据的排最后）
	sort.SliceStable(roles, func(i, j int) bool {
		oi, hasI := metaOrder[roles[i].Slug]
		oj, hasJ := metaOrder[roles[j].Slug]
		if hasI && hasJ {
			return oi < oj
		}
		if hasI {
			return true // i 有元数据，i 排在 j 前面
		}
		return false // i 无元数据，排在 j 后面
	})
	return roles, nil
}

// CreateRole 创建角色
// 注意：新角色默认不分配任何策略，必须通过 UpdateRolePermissions 显式授予权限。
// 之前的实现会在创建时自动添加 (slug, *, *) 全通配策略，casbin keyMatch
// 会将 * 匹配到任何 operation 路径，导致新角色瞬间获得所有受保护接口权限。
func (r *roleRepo) CreateRole(ctx context.Context, name, slug, desc string) error {
	if slug == "super_admin" {
		return fmt.Errorf("禁止创建超级管理员角色")
	}
	// 只保存角色元数据，权限由 UpdateRolePermissions 显式分配
	return r.db.Create(&mod.RoleMeta{Slug: slug, Name: name, Desc: desc}).Error
}

// UpdateRole 更新角色
func (r *roleRepo) UpdateRole(ctx context.Context, slug, name, desc string) error {
	if slug == "super_admin" {
		return fmt.Errorf("禁止修改超级管理员角色")
	}
	return r.db.Where("slug = ?", slug).Assign(mod.RoleMeta{Name: name, Desc: desc}).
		FirstOrCreate(&mod.RoleMeta{Slug: slug}).Error
}

// DeleteRole 删除角色（同时删除相关策略和元数据）
func (r *roleRepo) DeleteRole(ctx context.Context, slug string) error {
	if slug == "super_admin" {
		return fmt.Errorf("禁止删除超级管理员角色")
	}
	r.enforcer.RemoveFilteredPolicy(0, slug)
	r.enforcer.RemoveFilteredGroupingPolicy(0, slug)
	r.enforcer.RemoveFilteredGroupingPolicy(1, slug)
	r.db.Where("slug = ?", slug).Delete(&mod.RoleMeta{})
	return nil
}

// UpdateRolePermissions 更新角色权限
func (r *roleRepo) UpdateRolePermissions(ctx context.Context, role string, permissions map[string]map[string]bool) error {
	if role == "super_admin" {
		return fmt.Errorf("禁止修改超级管理员权限")
	}
	r.enforcer.RemoveFilteredPolicy(0, role)
	for _, mod := range permModules {
		for _, item := range mod.Items {
			if permissions[mod.Key] != nil && permissions[mod.Key][item.Key] {
				obj := "/" + mod.Key + "/" + item.Key
				r.enforcer.AddPolicy(role, obj, item.Key)
			}
		}
	}
	return nil
}

// GetRolePermissions 获取角色权限
func (r *roleRepo) GetRolePermissions(ctx context.Context, role string) (map[string]map[string]bool, error) {
	policies, _ := r.enforcer.GetFilteredPolicy(0, role)

	result := make(map[string]map[string]bool)
	for _, mod := range permModules {
		result[mod.Key] = make(map[string]bool)
		for _, item := range mod.Items {
			result[mod.Key][item.Key] = false
		}
	}

	for _, p := range policies {
		if len(p) >= 3 {
			obj := p[1]
			act := p[2]
			for _, mod := range permModules {
				for _, item := range mod.Items {
					key := "/" + mod.Key + "/" + item.Key
					if obj == "*" || obj == key || obj == "/"+mod.Key+"/*" {
						if act == "*" || act == item.Key {
							result[mod.Key][item.Key] = true
						}
					}
				}
			}
		}
	}

	if role == "super_admin" {
		for _, mod := range permModules {
			for _, item := range mod.Items {
				result[mod.Key][item.Key] = true
			}
		}
	}
	return result, nil
}

// GetUserRole 获取用户的角色
func (r *roleRepo) GetUserRole(ctx context.Context, userID int64) string {
	rules, _ := r.enforcer.GetFilteredGroupingPolicy(0, fmt.Sprintf("%d", userID))
	if len(rules) > 0 && len(rules[0]) >= 2 {
		return rules[0][1]
	}
	return ""
}

// AssignUserRole 为用户分配角色
func (r *roleRepo) AssignUserRole(ctx context.Context, userID int64, role string) error {
	r.enforcer.RemoveFilteredGroupingPolicy(0, fmt.Sprintf("%d", userID))
	_, err := r.enforcer.AddGroupingPolicy(fmt.Sprintf("%d", userID), role)
	return err
}

// CheckPermission 校验用户角色是否有权限执行操作
func (r *roleRepo) CheckPermission(ctx context.Context, role, object, action string) (bool, error) {
	return r.enforcer.Enforce(role, object, action)
}

// ListAllPermissions 获取所有权限模块定义
func (r *roleRepo) ListAllPermissions(ctx context.Context) []biz.PermModule {
	return permModules
}
