package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ---------- Role Service ----------

// ListRoles 获取角色列表
func (s *AdminService) ListRoles(ctx context.Context, req *adminv1.ListRolesRequest) (*adminv1.ListRolesReply, error) {
	roles, err := s.rc.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, &adminv1.RoleItem{
			Name:      r.Name,
			Slug:      r.Slug,
			Desc:      r.Desc,
			Users:     int32(r.Users),
			Enabled:   r.Enabled,
			IsSystem:  r.IsSystem,
			CreatedAt: r.CreatedAt,
		})
	}
	return &adminv1.ListRolesReply{Items: items}, nil
}

// CreateRole 创建角色
func (s *AdminService) CreateRole(ctx context.Context, req *adminv1.CreateRoleRequest) (*adminv1.CreateRoleReply, error) {
	if err := s.rc.CreateRole(ctx, req.Name, req.Slug, req.Desc); err != nil {
		return nil, err
	}
	return &adminv1.CreateRoleReply{Success: true}, nil
}

// UpdateRole 更新角色
func (s *AdminService) UpdateRole(ctx context.Context, req *adminv1.UpdateRoleRequest) (*adminv1.UpdateRoleReply, error) {
	if err := s.rc.UpdateRole(ctx, req.Slug, req.Name, req.Desc); err != nil {
		return nil, err
	}
	return &adminv1.UpdateRoleReply{Success: true}, nil
}

// DeleteRole 删除角色
func (s *AdminService) DeleteRole(ctx context.Context, req *adminv1.DeleteRoleRequest) (*adminv1.DeleteRoleReply, error) {
	if err := s.rc.DeleteRole(ctx, req.Slug); err != nil {
		return nil, err
	}
	return &adminv1.DeleteRoleReply{Success: true}, nil
}

// GetRolePermissions 获取角色权限
func (s *AdminService) GetRolePermissions(ctx context.Context, req *adminv1.GetRolePermissionsRequest) (*adminv1.GetRolePermissionsReply, error) {
	perms, err := s.rc.GetRolePermissions(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetRolePermissionsReply{Permissions: convertBizPermsToProto(perms)}, nil
}

// UpdateRolePermissions 更新角色权限
func (s *AdminService) UpdateRolePermissions(ctx context.Context, req *adminv1.UpdateRolePermissionsRequest) (*adminv1.UpdateRolePermissionsReply, error) {
	perms := convertProtoPermsToBiz(req.Permissions)
	if err := s.rc.UpdateRolePermissions(ctx, req.Role, perms); err != nil {
		return nil, err
	}
	return &adminv1.UpdateRolePermissionsReply{Success: true}, nil
}

// ListAllPermissions 获取所有权限模块定义
func (s *AdminService) ListAllPermissions(ctx context.Context, req *adminv1.ListAllPermissionsRequest) (*adminv1.ListAllPermissionsReply, error) {
	modules := s.rc.ListAllPermissions(ctx)
	items := make([]*adminv1.PermModule, 0, len(modules))
	for _, m := range modules {
		items = append(items, &adminv1.PermModule{
			Key:   m.Key,
			Title: m.Title,
			Items: convertPermItemsToProto(m.Items),
		})
	}
	return &adminv1.ListAllPermissionsReply{Modules: items}, nil
}

// ---------- Helper ----------

// convertBizPermsToProto 将业务层的权限映射转换为 proto 格式
func convertBizPermsToProto(perms map[string]map[string]bool) map[string]*adminv1.PermissionValue {
	result := make(map[string]*adminv1.PermissionValue)
	for modKey, items := range perms {
		vals := make(map[string]bool)
		for itemKey, v := range items {
			vals[itemKey] = v
		}
		result[modKey] = &adminv1.PermissionValue{Values: vals}
	}
	return result
}

// convertProtoPermsToBiz 将 proto 格式的权限映射转换为业务层格式
func convertProtoPermsToBiz(perms map[string]*adminv1.PermissionValue) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for modKey, pv := range perms {
		vals := make(map[string]bool)
		for itemKey, v := range pv.Values {
			vals[itemKey] = v
		}
		result[modKey] = vals
	}
	return result
}

// convertPermItemsToProto 将业务层的权限项列表转换为 proto 格式
func convertPermItemsToProto(items []biz.PermItem) []*adminv1.PermItem {
	result := make([]*adminv1.PermItem, 0, len(items))
	for _, item := range items {
		result = append(result, &adminv1.PermItem{
			Key:   item.Key,
			Label: item.Label,
		})
	}
	return result
}
