package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// ListUser 用户列表（mock 数据）
func (s *AdminService) ListUser(ctx context.Context, req *adminv1.ListUserRequest) (*adminv1.ListUserReply, error) {
	return &adminv1.ListUserReply{
		Items: []*adminv1.UserItem{
			{Id: 1, Name: "张三", Phone: "138****1234", Level: "普通会员", RegTime: "2025-01-15", Amount: 899900, Orders: 12, Status: "正常"},
			{Id: 2, Name: "李四", Phone: "139****5678", Level: "黄金会员", RegTime: "2024-06-20", Amount: 1899900, Orders: 28, Status: "正常"},
			{Id: 3, Name: "王五", Phone: "137****9012", Level: "普通会员", RegTime: "2025-03-08", Amount: 12900, Orders: 3, Status: "冻结"},
			{Id: 4, Name: "赵六", Phone: "136****3456", Level: "钻石会员", RegTime: "2023-11-01", Amount: 5999900, Orders: 56, Status: "正常"},
			{Id: 5, Name: "孙七", Phone: "135****7890", Level: "黄金会员", RegTime: "2024-09-12", Amount: 899900, Orders: 15, Status: "正常"},
		},
		Total: 5,
	}, nil
}

// GetUser 用户详情（mock 数据）
func (s *AdminService) GetUser(ctx context.Context, req *adminv1.GetUserRequest) (*adminv1.GetUserReply, error) {
	return &adminv1.GetUserReply{
		User: &adminv1.UserItem{
			Id:     req.Id,
			Name:   "张三",
			Phone:  "138****1234",
			Level:  "普通会员",
			RegTime: "2025-01-15",
			Amount: 899900,
			Orders: 12,
			Status: "正常",
		},
	}, nil
}

// CreateUser 创建用户（mock 实现）
func (s *AdminService) CreateUser(ctx context.Context, req *adminv1.CreateUserRequest) (*adminv1.CreateUserReply, error) {
	return &adminv1.CreateUserReply{
		Id: 100,
	}, nil
}

// UpdateUser 更新用户（mock 实现）
func (s *AdminService) UpdateUser(ctx context.Context, req *adminv1.UpdateUserRequest) (*adminv1.UpdateUserReply, error) {
	return &adminv1.UpdateUserReply{
		Success: true,
	}, nil
}

// DeleteUser 删除用户（mock 实现）
func (s *AdminService) DeleteUser(ctx context.Context, req *adminv1.DeleteUserRequest) (*adminv1.DeleteUserReply, error) {
	return &adminv1.DeleteUserReply{
		Success: true,
	}, nil
}

// ToggleUserStatus 冻结/解冻用户（mock 实现）
func (s *AdminService) ToggleUserStatus(ctx context.Context, req *adminv1.ToggleUserStatusRequest) (*adminv1.ToggleUserStatusReply, error) {
	return &adminv1.ToggleUserStatusReply{
		Success:   true,
		NewStatus: "冻结",
	}, nil
}

// GetUserStats 获取用户统计数据（mock 数据）
func (s *AdminService) GetUserStats(ctx context.Context, req *adminv1.GetUserStatsRequest) (*adminv1.GetUserStatsReply, error) {
	return &adminv1.GetUserStatsReply{
		TotalUsers:     23568,
		TodayNew:       128,
		TodayNewGrowth: 5.2,
		ActiveUsers:    15680,
		MemberUsers:    8230,
		MemberRatio:    34.9,
	}, nil
}
