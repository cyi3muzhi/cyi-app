package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"
	"cyi-app/pkg/kratos/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ---------- Helper: biz -> proto 转换 ----------

func adminToProto(admin *biz.AdminAuth) *adminv1.AdminInfo {
	if admin == nil {
		return nil
	}
	return &adminv1.AdminInfo{
		Id:         admin.ID,
		Name:       admin.Name,
		Username:   admin.Username,
		Role:       admin.Role,
		Email:      admin.Email,
		Phone:      admin.Phone,
		Department: admin.Department,
		Position:   admin.Position,
		Bio:        admin.Bio,
		Avatar:     admin.Avatar,
		CreatedAt:  admin.CreatedAt,
		LastLogin:  admin.LastLogin,
	}
}

// 确保 AdminService 实现了 AuthServer 接口

// Login 管理员登录
func (s *AdminService) Login(ctx context.Context, req *adminv1.LoginRequest) (*adminv1.LoginReply, error) {
	s.log.Infof("管理员登录: %s", req.Username)

	// token 由 user 微服务生成并返回，BFF 透传
	admin, token, err := s.ac.Login(ctx, req.Username, req.Password)
	if err != nil || admin == nil {
		return nil, err
	}
	return &adminv1.LoginReply{
		Token: token,
		Admin: adminToProto(admin),
	}, nil
}

// Logout 管理员登出
func (s *AdminService) Logout(ctx context.Context, req *adminv1.LogoutRequest) (*adminv1.LogoutReply, error) {
	// 优先从 JWT 上下文取 token，其次从请求体取
	token, ok := mid.GetTokenFromContext(ctx)
	if !ok || token == "" {
		token = req.Token
	}
	if token == "" {
		return nil, adminv1.ErrorUnauthorized("无效的token")
	}
	if err := s.ac.Logout(ctx, token); err != nil {
		s.log.Errorf("登出失败: %v", err)
		return nil, err
	}
	return &adminv1.LogoutReply{Success: true}, nil
}

// GetAdminInfo 获取当前管理员信息
func (s *AdminService) GetAdminInfo(ctx context.Context, req *adminv1.GetAdminInfoRequest) (*adminv1.GetAdminInfoReply, error) {
	_ = req

	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}

	admin, err := s.ac.GetAdminInfo(ctx, adminID)
	if err != nil {
		s.log.Errorf("获取管理员信息失败: %v", err)
		return nil, err
	}

	return &adminv1.GetAdminInfoReply{
		Admin: adminToProto(admin),
	}, nil
}

// ChangeAvatar 更换头像（接收 Base64 图片）
func (s *AdminService) ChangeAvatar(ctx context.Context, req *adminv1.ChangeAvatarRequest) (*adminv1.ChangeAvatarReply, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}
	s.log.Infof("更换头像: adminID=%d", adminID)

	avatarURL, err := s.ac.UpdateAvatar(ctx, adminID, req.AvatarBase64)
	if err != nil {
		s.log.Errorf("更换头像失败: %v", err)
		return nil, err
	}

	return &adminv1.ChangeAvatarReply{
		Success:   true,
		AvatarUrl: avatarURL,
	}, nil
}

// UpdateAdminInfo 更新基本信息
func (s *AdminService) UpdateAdminInfo(ctx context.Context, req *adminv1.UpdateAdminInfoRequest) (*adminv1.UpdateAdminInfoReply, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}

	s.log.Infof("更新基本信息: adminID=%d", adminID)

	err = s.ac.UpdateAdminInfo(ctx, adminID, &biz.AdminAuth{
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
		Bio:        req.Bio,
	})
	if err != nil {
		s.log.Errorf("更新基本信息失败: %v", err)
		return nil, err
	}

	return &adminv1.UpdateAdminInfoReply{Success: true}, nil
}

// 确保 AdminService 实现了 AuthServer 接口
var _ adminv1.AdminAuthServer = (*AdminService)(nil)

// 占位：google.protobuf.Empty 需要导入
var _ = &emptypb.Empty{}
