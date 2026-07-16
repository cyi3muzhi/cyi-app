package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// RoleRepo 角色仓库接口
type RoleRepo interface {
	ListRoles(ctx context.Context) ([]*RoleItem, error)
	CreateRole(ctx context.Context, name, slug, desc, color string) error
	UpdateRole(ctx context.Context, slug, name, desc, color string) error
	DeleteRole(ctx context.Context, slug string) error
	GetRolePermissions(ctx context.Context, role string) (map[string]map[string]bool, error)
	UpdateRolePermissions(ctx context.Context, role string, permissions map[string]map[string]bool) error
	ListAllPermissions(ctx context.Context) []PermModule
	CheckPermission(ctx context.Context, role, object, action string) (bool, error)
	HasPolicyForOperation(ctx context.Context, operation string) (bool, error)
}

// RoleUseCase 角色管理业务逻辑
type RoleUseCase struct {
	repo RoleRepo
	log  *log.Helper
}

// NewRoleUseCase 创建 RoleUseCase
func NewRoleUseCase(repo RoleRepo, logger log.Logger) *RoleUseCase {
	return &RoleUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// RoleItem 角色
type RoleItem struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Desc      string `json:"desc"`
	Color     string `json:"color"`
	Users     int    `json:"users"`
	Enabled   bool   `json:"enabled"`
	IsSystem  bool   `json:"isSystem"`
	CreatedAt string `json:"createdAt"`
}

// PermModule 权限模块
type PermModule struct {
	Key   string     `json:"key"`
	Title string     `json:"title"`
	Items []PermItem `json:"items"`
}

// PermItem 权限项
type PermItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ListRoles 获取角色列表
func (uc *RoleUseCase) ListRoles(ctx context.Context) ([]*RoleItem, error) {
	return uc.repo.ListRoles(ctx)
}

// CreateRole 创建角色
func (uc *RoleUseCase) CreateRole(ctx context.Context, name, slug, desc, color string) error {
	return uc.repo.CreateRole(ctx, name, slug, desc, color)
}

// UpdateRole 更新角色
func (uc *RoleUseCase) UpdateRole(ctx context.Context, slug, name, desc, color string) error {
	return uc.repo.UpdateRole(ctx, slug, name, desc, color)
}

// DeleteRole 删除角色
func (uc *RoleUseCase) DeleteRole(ctx context.Context, slug string) error {
	return uc.repo.DeleteRole(ctx, slug)
}

// GetRolePermissions 获取角色权限
func (uc *RoleUseCase) GetRolePermissions(ctx context.Context, role string) (map[string]map[string]bool, error) {
	return uc.repo.GetRolePermissions(ctx, role)
}

// UpdateRolePermissions 更新角色权限
func (uc *RoleUseCase) UpdateRolePermissions(ctx context.Context, role string, permissions map[string]map[string]bool) error {
	return uc.repo.UpdateRolePermissions(ctx, role, permissions)
}

// ListAllPermissions 获取所有权限模块定义
func (uc *RoleUseCase) ListAllPermissions(ctx context.Context) []PermModule {
	return uc.repo.ListAllPermissions(ctx)
}

// CheckPermission 校验用户角色是否有权限执行操作
func (uc *RoleUseCase) CheckPermission(ctx context.Context, role, object, action string) (bool, error) {
	return uc.repo.CheckPermission(ctx, role, object, action)
}

// HasPolicyForOperation 检查是否有任何策略引用了指定操作
func (uc *RoleUseCase) HasPolicyForOperation(ctx context.Context, operation string) (bool, error) {
	return uc.repo.HasPolicyForOperation(ctx, operation)
}
