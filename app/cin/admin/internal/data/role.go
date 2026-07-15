package data

import (
	"context"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// roleRepo 角色数据仓库（通过 gRPC 代理到 user 服务）
type roleRepo struct {
	client userV1.AdminRoleClient
	log    *log.Helper
}

// NewRoleRepo 创建 RoleRepo
func NewRoleRepo(client userV1.AdminRoleClient, logger log.Logger) biz.RoleRepo {
	return &roleRepo{client: client, log: log.NewHelper(logger)}
}

// ListRoles 获取所有角色
func (r *roleRepo) ListRoles(ctx context.Context) ([]*biz.RoleItem, error) {
	resp, err := r.client.ListRoles(ctx, &userV1.ListRolesRequest{})
	if err != nil {
		return nil, err
	}
	items := make([]*biz.RoleItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, &biz.RoleItem{
			Name:      item.Name,
			Slug:      item.Slug,
			Desc:      item.Desc,
			Users:     int(item.Users),
			Enabled:   item.Enabled,
			IsSystem:  item.IsSystem,
			CreatedAt: item.CreatedAt,
		})
	}
	return items, nil
}

// CreateRole 创建角色
func (r *roleRepo) CreateRole(ctx context.Context, name, slug, desc string) error {
	_, err := r.client.CreateRole(ctx, &userV1.CreateRoleRequest{Name: name, Slug: slug, Desc: desc})
	return err
}

// UpdateRole 更新角色
func (r *roleRepo) UpdateRole(ctx context.Context, slug, name, desc string) error {
	_, err := r.client.UpdateRole(ctx, &userV1.UpdateRoleRequest{Slug: slug, Name: name, Desc: desc})
	return err
}

// DeleteRole 删除角色
func (r *roleRepo) DeleteRole(ctx context.Context, slug string) error {
	_, err := r.client.DeleteRole(ctx, &userV1.DeleteRoleRequest{Slug: slug})
	return err
}

// GetRolePermissions 获取角色权限
func (r *roleRepo) GetRolePermissions(ctx context.Context, role string) (map[string]map[string]bool, error) {
	resp, err := r.client.GetRolePermissions(ctx, &userV1.GetRolePermissionsRequest{Role: role})
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]bool)
	for modKey, pv := range resp.Permissions {
		vals := make(map[string]bool)
		for itemKey, v := range pv.Values {
			vals[itemKey] = v
		}
		result[modKey] = vals
	}
	return result, nil
}

// UpdateRolePermissions 更新角色权限
func (r *roleRepo) UpdateRolePermissions(ctx context.Context, role string, permissions map[string]map[string]bool) error {
	perms := make(map[string]*userV1.PermissionValue)
	for modKey, items := range permissions {
		vals := make(map[string]bool)
		for itemKey, v := range items {
			vals[itemKey] = v
		}
		perms[modKey] = &userV1.PermissionValue{Values: vals}
	}
	_, err := r.client.UpdateRolePermissions(ctx, &userV1.UpdateRolePermissionsRequest{Role: role, Permissions: perms})
	return err
}

// ListAllPermissions 获取所有权限模块定义
func (r *roleRepo) ListAllPermissions(ctx context.Context) []biz.PermModule {
	resp, err := r.client.ListAllPermissions(ctx, &userV1.ListAllPermissionsRequest{})
	if err != nil {
		r.log.Warnf("获取权限模块定义失败: %v", err)
		return nil
	}
	modules := make([]biz.PermModule, 0, len(resp.Modules))
	for _, m := range resp.Modules {
		items := make([]biz.PermItem, 0, len(m.Items))
		for _, item := range m.Items {
			items = append(items, biz.PermItem{Key: item.Key, Label: item.Label})
		}
		modules = append(modules, biz.PermModule{Key: m.Key, Title: m.Title, Items: items})
	}
	return modules
}

// CheckPermission 校验权限（供 RemoteEnforcer 使用）
func (r *roleRepo) CheckPermission(ctx context.Context, role, object, action string) (bool, error) {
	resp, err := r.client.CheckPermission(ctx, &userV1.CheckPermissionRequest{Role: role, Object: object, Action: action})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// HasPolicyForOperation 检查是否有任何策略引用了指定操作路径
func (r *roleRepo) HasPolicyForOperation(ctx context.Context, operation string) (bool, error) {
	resp, err := r.client.HasPolicyForOperation(ctx, &userV1.HasPolicyForOperationRequest{Operation: operation})
	if err != nil {
		return false, err
	}
	return resp.Has, nil
}
