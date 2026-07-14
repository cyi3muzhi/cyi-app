package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ---------- Helper: biz -> proto 转换 ----------

func adminToProto(admin *biz.AdminAuth) *adminv1.AdminInfo {
	if admin == nil {
		return nil
	}
	return &adminv1.AdminInfo{
		Id:       admin.ID,
		Name:     admin.Name,
		Username: admin.Username,
		Role:     admin.Role,
		Email:    admin.Email,
		Avatar:   admin.Avatar,
	}
}

// ---------- 实现 AdminAuth 接口 ----------

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
	if err := s.ac.Logout(ctx, req.Token); err != nil {
		s.log.Errorf("登出失败: %v", err)
		return nil, err
	}
	return &adminv1.LogoutReply{Success: true}, nil
}

// GetAdminInfo 获取当前管理员信息
func (s *AdminService) GetAdminInfo(ctx context.Context, req *adminv1.GetAdminInfoRequest) (*adminv1.GetAdminInfoReply, error) {
	_ = req
	return &adminv1.GetAdminInfoReply{
		Admin: &adminv1.AdminInfo{
			Id:       1,
			Name:     "管理员",
			Username: "admin",
			Role:     "超级管理员",
			Email:    "admin@yuexuan.com",
			Avatar:   "管",
		},
	}, nil
}

// 确保 AdminService 实现了 AuthServer 接口
var _ adminv1.AdminAuthServer = (*AdminService)(nil)

// 占位：google.protobuf.Empty 需要导入
var _ = &emptypb.Empty{}
