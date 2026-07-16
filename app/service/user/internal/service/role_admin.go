package service

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AdminRoleService 实现 proto 定义的 AdminRole 服务
type AdminRoleService struct {
	v1.UnimplementedAdminRoleServer
	uc  *biz.RoleUseCase
	log *log.Helper
}

// NewAdminRoleService 创建 AdminRoleService
func NewAdminRoleService(uc *biz.RoleUseCase, logger log.Logger) *AdminRoleService {
	return &AdminRoleService{uc: uc, log: log.NewHelper(logger)}
}

// ListRoles 获取角色列表
func (s *AdminRoleService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesReply, error) {
	roles, err := s.uc.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, &v1.RoleItem{
			Name:      r.Name,
			Slug:      r.Slug,
			Desc:      r.Desc,
			Color:     r.Color,
			Users:     int32(r.Users),
			Enabled:   r.Enabled,
			IsSystem:  r.IsSystem,
			CreatedAt: r.CreatedAt,
		})
	}
	return &v1.ListRolesReply{Items: items}, nil
}

// CreateRole 创建角色
func (s *AdminRoleService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleReply, error) {
	if err := s.uc.CreateRole(ctx, req.Name, req.Slug, req.Desc, req.Color); err != nil {
		return nil, err
	}
	return &v1.CreateRoleReply{Success: true}, nil
}

// UpdateRole 更新角色
func (s *AdminRoleService) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleReply, error) {
	if err := s.uc.UpdateRole(ctx, req.Slug, req.Name, req.Desc, req.Color); err != nil {
		return nil, err
	}
	return &v1.UpdateRoleReply{Success: true}, nil
}

// DeleteRole 删除角色
func (s *AdminRoleService) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleReply, error) {
	if err := s.uc.DeleteRole(ctx, req.Slug); err != nil {
		return nil, err
	}
	return &v1.DeleteRoleReply{Success: true}, nil
}

// GetRolePermissions 获取角色权限
func (s *AdminRoleService) GetRolePermissions(ctx context.Context, req *v1.GetRolePermissionsRequest) (*v1.GetRolePermissionsReply, error) {
	perms, err := s.uc.GetRolePermissions(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*v1.PermissionValue)
	for modKey, items := range perms {
		vals := make(map[string]bool)
		for itemKey, v := range items {
			vals[itemKey] = v
		}
		result[modKey] = &v1.PermissionValue{Values: vals}
	}
	return &v1.GetRolePermissionsReply{Permissions: result}, nil
}

// UpdateRolePermissions 更新角色权限
func (s *AdminRoleService) UpdateRolePermissions(ctx context.Context, req *v1.UpdateRolePermissionsRequest) (*v1.UpdateRolePermissionsReply, error) {
	perms := make(map[string]map[string]bool)
	for modKey, pv := range req.Permissions {
		vals := make(map[string]bool)
		for itemKey, v := range pv.Values {
			vals[itemKey] = v
		}
		perms[modKey] = vals
	}
	if err := s.uc.UpdateRolePermissions(ctx, req.Role, perms); err != nil {
		return nil, err
	}
	return &v1.UpdateRolePermissionsReply{Success: true}, nil
}

// ListAllPermissions 获取所有权限模块定义
func (s *AdminRoleService) ListAllPermissions(ctx context.Context, req *v1.ListAllPermissionsRequest) (*v1.ListAllPermissionsReply, error) {
	modules := s.uc.ListAllPermissions(ctx)
	items := make([]*v1.PermModule, 0, len(modules))
	for _, m := range modules {
		permItems := make([]*v1.PermItem, 0, len(m.Items))
		for _, item := range m.Items {
			permItems = append(permItems, &v1.PermItem{Key: item.Key, Label: item.Label})
		}
		items = append(items, &v1.PermModule{Key: m.Key, Title: m.Title, Items: permItems})
	}
	return &v1.ListAllPermissionsReply{Modules: items}, nil
}

// CheckPermission 校验权限（供 admin BFF 中间件使用）
func (s *AdminRoleService) CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionReply, error) {
	allowed, err := s.uc.CheckPermission(ctx, req.Role, req.Object, req.Action)
	if err != nil {
		return nil, err
	}
	return &v1.CheckPermissionReply{Allowed: allowed}, nil
}

// HasPolicyForOperation 检查是否有任何策略引用了指定操作路径
func (s *AdminRoleService) HasPolicyForOperation(ctx context.Context, req *v1.HasPolicyForOperationRequest) (*v1.HasPolicyForOperationReply, error) {
	has, err := s.uc.HasPolicyForOperation(ctx, req.Operation)
	if err != nil {
		return nil, err
	}
	return &v1.HasPolicyForOperationReply{Has: has}, nil
}
