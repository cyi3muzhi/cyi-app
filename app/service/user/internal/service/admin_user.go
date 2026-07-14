package service

import (
	"context"
	"fmt"
	"time"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AdminUserService 管理员用户服务
type AdminUserService struct {
	v1.UnimplementedAdminUserServer
	uc  *biz.AdminUserUseCase
	log *log.Helper
}

// NewAdminUserService 创建管理员用户服务
func NewAdminUserService(uc *biz.AdminUserUseCase, logger log.Logger) *AdminUserService {
	return &AdminUserService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// Login 管理员登录验证，返回 token
func (s *AdminUserService) Login(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginReply, error) {
	admin, token, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil || admin == nil {
		return nil, err
	}
	return &v1.AdminLoginReply{
		AdminId:  admin.ID,
		Username: admin.Username,
		Name:     admin.Name,
		Role:     admin.Role,
		Valid:    true,
		Token:    token,
	}, nil
}

// Logout 管理员登出，移除 Redis 中的 token
func (s *AdminUserService) Logout(ctx context.Context, req *v1.AdminLogoutRequest) (*v1.AdminLogoutReply, error) {
	if err := s.uc.Logout(ctx, req.Token); err != nil {
		s.log.Errorf("登出失败: %v", err)
		return nil, err
	}
	s.log.Infof("管理员登出成功")
	return &v1.AdminLogoutReply{Success: true}, nil
}

// ValidateToken 验证 token 是否有效（未被新登录覆盖）
func (s *AdminUserService) ValidateToken(ctx context.Context, req *v1.AdminValidateTokenRequest) (*v1.AdminValidateTokenReply, error) {
	valid, adminID, username, err := s.uc.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &v1.AdminValidateTokenReply{
		Valid:    valid,
		AdminId:  adminID,
		Username: username,
	}, nil
}

// CreateAdminUser 创建管理员
func (s *AdminUserService) CreateAdminUser(ctx context.Context, req *v1.CreateAdminUserRequest) (*v1.CreateAdminUserReply, error) {
	id, err := s.uc.Create(ctx, &biz.AdminUser{
		Username: req.Username,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
		Email:    req.Email,
		Status:   1,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAdminUserReply{Id: id}, nil
}

// UpdateAdminUser 更新管理员
func (s *AdminUserService) UpdateAdminUser(ctx context.Context, req *v1.UpdateAdminUserRequest) (*v1.UpdateAdminUserReply, error) {
	err := s.uc.Update(ctx, &biz.AdminUser{
		ID:    req.Id,
		Name:  req.Name,
		Role:  req.Role,
		Email: req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAdminUserReply{Success: true}, nil
}

// GetAdminUser 获取管理员
func (s *AdminUserService) GetAdminUser(ctx context.Context, req *v1.GetAdminUserRequest) (*v1.GetAdminUserReply, error) {
	admin, err := s.uc.Get(ctx, req.Id)
	if err != nil || admin == nil {
		return &v1.GetAdminUserReply{}, nil
	}
	return &v1.GetAdminUserReply{Admin: adminToProto(admin)}, nil
}

// ListAdminUser 管理员列表
func (s *AdminUserService) ListAdminUser(ctx context.Context, req *v1.ListAdminUserRequest) (*v1.ListAdminUserReply, error) {
	items, total, err := s.uc.List(ctx, req.Page, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		return nil, err
	}
	var list []*v1.AdminUserInfo
	for _, item := range items {
		list = append(list, adminToProto(item))
	}
	return &v1.ListAdminUserReply{Items: list, Total: int32(total)}, nil
}

// ToggleAdminUserStatus 切换状态
func (s *AdminUserService) ToggleAdminUserStatus(ctx context.Context, req *v1.ToggleAdminUserStatusRequest) (*v1.ToggleAdminUserStatusReply, error) {
	newStatus, err := s.uc.ToggleStatus(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.ToggleAdminUserStatusReply{Success: true, NewStatus: newStatus}, nil
}

// UpdateAdminAvatar 更新管理员头像
func (s *AdminUserService) UpdateAdminAvatar(ctx context.Context, req *v1.UpdateAdminAvatarRequest) (*v1.UpdateAdminAvatarReply, error) {
	if err := s.uc.UpdateAvatar(ctx, req.AdminId, req.AvatarUrl); err != nil {
		s.log.Errorf("更新头像失败: %v", err)
		return nil, err
	}
	return &v1.UpdateAdminAvatarReply{Success: true}, nil
}

// adminToProto biz -> proto 转换
func adminToProto(admin *biz.AdminUser) *v1.AdminUserInfo {
	if admin == nil {
		return nil
	}
	createdAt := admin.CreatedAt.Format("2006-01-02 15:04:05")
	lastLogin := ""
	if admin.LastLogin > 0 {
		lastLogin = time.Unix(admin.LastLogin, 0).Format("2006-01-02 15:04:05")
	}
	return &v1.AdminUserInfo{
		Id:        admin.ID,
		Username:  admin.Username,
		Name:      admin.Name,
		Role:      admin.Role,
		Email:     admin.Email,
		Avatar:    admin.Avatar,
		Status:    admin.Status,
		LastLogin: lastLogin,
		CreatedAt: createdAt,
	}
}

// 确保 fmt 被使用
var _ = fmt.Sprintf
