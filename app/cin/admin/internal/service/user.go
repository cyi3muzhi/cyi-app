package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// userItemToProto 将 biz.UserInfo 转换为 proto.UserItem
// 字段一一映射，admin 端不展示消费金额/订单数（后续在用户详情由 order service 聚合）
func userItemToProto(user *biz.UserInfo) *adminv1.UserItem {
	if user == nil {
		return nil
	}
	return &adminv1.UserItem{
		Id:             user.ID,
		Username:       user.Username,
		Nickname:       user.Nickname,
		Phone:          user.Phone,
		Email:          user.Email,
		Avatar:         user.Avatar,
		Gender:         user.Gender,
		Level:           user.Level,
		LastLoginAt:    user.LastLoginAt,
		RegisterSource: user.RegisterSource,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
	}
}

// ListUser 获取用户列表
func (s *AdminService) ListUser(ctx context.Context, req *adminv1.ListUserRequest) (*adminv1.ListUserReply, error) {
	s.log.Infof("获取用户列表: page=%d, pageSize=%d, keyword=%s", req.Page, req.PageSize, req.Keyword)

	listReq := &biz.UserListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Level:    req.Level,
		Status:   req.Status,
	}

	reply, err := s.userUC.ListUser(ctx, listReq)
	if err != nil {
		s.log.Errorf("获取用户列表失败: %v", err)
		return nil, err
	}

	items := make([]*adminv1.UserItem, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, userItemToProto(item))
	}

	return &adminv1.ListUserReply{
		Items: items,
		Total: reply.Total,
	}, nil
}

// GetUser 获取用户详情
func (s *AdminService) GetUser(ctx context.Context, req *adminv1.GetUserRequest) (*adminv1.GetUserReply, error) {
	s.log.Infof("获取用户详情: id=%d", req.Id)

	user, err := s.userUC.GetUser(ctx, req.Id)
	if err != nil {
		s.log.Errorf("获取用户详情失败: %v", err)
		return nil, err
	}

	return &adminv1.GetUserReply{
		User: userItemToProto(user),
	}, nil
}

// CreateUser 创建用户
func (s *AdminService) CreateUser(ctx context.Context, req *adminv1.CreateUserRequest) (*adminv1.CreateUserReply, error) {
	s.log.Infof("创建用户: username=%s, phone=%s", req.Username, req.Phone)

	id, err := s.userUC.CreateUser(ctx, &biz.CreateUserRequest{
		Username:       req.Username,
		Phone:          req.Phone,
		Email:          req.Email,
		Password:       req.Password,
		Nickname:       req.Nickname,
		RegisterSource:  req.RegisterSource,
	})
	if err != nil {
		s.log.Errorf("创建用户失败: %v", err)
		return nil, err
	}

	return &adminv1.CreateUserReply{
		Id: id,
	}, nil
}

// UpdateUser 更新用户资料
func (s *AdminService) UpdateUser(ctx context.Context, req *adminv1.UpdateUserRequest) (*adminv1.UpdateUserReply, error) {
	s.log.Infof("更新用户: id=%d", req.Id)

	err := s.userUC.UpdateUser(ctx, &biz.UpdateUserRequest{
		ID:       req.Id,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Email:    req.Email,
		Gender:   req.Gender,
		Birthday: req.Birthday,
		Level:    req.Level,
	})
	if err != nil {
		s.log.Errorf("更新用户失败: %v", err)
		return nil, err
	}

	return &adminv1.UpdateUserReply{
		Success: true,
	}, nil
}

// DeleteUser 删除用户（软删除）
func (s *AdminService) DeleteUser(ctx context.Context, req *adminv1.DeleteUserRequest) (*adminv1.DeleteUserReply, error) {
	s.log.Infof("删除用户: id=%d", req.Id)

	err := s.userUC.DeleteUser(ctx, req.Id)
	if err != nil {
		s.log.Errorf("删除用户失败: %v", err)
		return nil, err
	}

	return &adminv1.DeleteUserReply{
		Success: true,
	}, nil
}

// ToggleUserStatus 冻结/解冻用户
func (s *AdminService) ToggleUserStatus(ctx context.Context, req *adminv1.ToggleUserStatusRequest) (*adminv1.ToggleUserStatusReply, error) {
	s.log.Infof("切换用户状态: id=%d", req.Id)

	newStatus, err := s.userUC.ToggleUserStatus(ctx, req.Id)
	if err != nil {
		s.log.Errorf("切换用户状态失败: %v", err)
		return nil, err
	}

	return &adminv1.ToggleUserStatusReply{
		Success:   true,
		NewStatus: newStatus,
	}, nil
}

// GetUserStats 获取用户统计数据
func (s *AdminService) GetUserStats(ctx context.Context, req *adminv1.GetUserStatsRequest) (*adminv1.GetUserStatsReply, error) {
	s.log.Infof("获取用户统计数据")

	stats, err := s.userUC.GetUserStats(ctx)
	if err != nil {
		s.log.Errorf("获取用户统计数据失败: %v", err)
		return nil, err
	}

	return &adminv1.GetUserStatsReply{
		TotalUsers:     stats.TotalUsers,
		TodayNew:       stats.TodayNew,
		TodayNewGrowth: stats.TodayGrowth,
		ActiveUsers:    stats.ActiveUsers,
		MemberUsers:    stats.MemberUsers,
		MemberRatio:    stats.MemberRatio,
	}, nil
}

// 确保 AdminService 实现了 AdminUserServer 接口
var _ adminv1.AdminUserServer = (*AdminService)(nil)
