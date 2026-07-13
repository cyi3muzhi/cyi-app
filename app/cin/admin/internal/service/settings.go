package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// GetSettings 获取系统设置（mock 数据）
func (s *AdminService) GetSettings(ctx context.Context, req *adminv1.GetSettingsRequest) (*adminv1.GetSettingsReply, error) {
	return &adminv1.GetSettingsReply{
		ShopName:             "我的商城",
		ShopDescription:      "优质商品，尽在掌握",
		ContactPhone:         "400-888-8888",
		ContactEmail:         "admin@example.com",
		IcpRecord:            "京ICP备2025XXXXXX号",
		OrderNotification:    true,
		RegisterNotification: true,
		SystemNotification:   false,
	}, nil
}

// UpdateSettings 更新系统设置（mock 实现）
func (s *AdminService) UpdateSettings(ctx context.Context, req *adminv1.UpdateSettingsRequest) (*adminv1.UpdateSettingsReply, error) {
	return &adminv1.UpdateSettingsReply{
		Success: true,
	}, nil
}

// ChangePassword 修改密码（mock 实现）
func (s *AdminService) ChangePassword(ctx context.Context, req *adminv1.ChangePasswordRequest) (*adminv1.ChangePasswordReply, error) {
	return &adminv1.ChangePasswordReply{
		Success: true,
	}, nil
}
