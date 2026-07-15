package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ---------- Staff Service ----------

// ListStaff 获取员工列表
func (s *AdminService) ListStaff(ctx context.Context, req *adminv1.ListStaffRequest) (*adminv1.ListStaffReply, error) {
	items, total, err := s.sc.List(ctx, req.Page, req.PageSize, req.Keyword, req.Status, req.Role)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.StaffInfo, 0, len(items))
	for _, item := range items {
		list = append(list, &adminv1.StaffInfo{
			Id:         item.ID,
			Name:       item.Name,
			Username:   item.Username,
			Phone:      item.Phone,
			Email:      item.Email,
			Role:       item.Role,
			Department: item.Department,
			Position:   item.Position,
			Bio:        item.Bio,
			Avatar:     item.Avatar,
			Status:     int32(item.Status),
			LastLogin:  item.LastLogin,
			CreatedAt:  item.CreatedAt,
		})
	}
	return &adminv1.ListStaffReply{Items: list, Total: total}, nil
}

// CreateStaff 创建员工
func (s *AdminService) CreateStaff(ctx context.Context, req *adminv1.CreateStaffRequest) (*adminv1.CreateStaffReply, error) {
	id, err := s.sc.Create(ctx, &biz.Staff{
		Username:   req.Username,
		Name:       req.Name,
		Role:       req.Role,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
	}, req.Password)
	if err != nil {
		return nil, err
	}
	return &adminv1.CreateStaffReply{Id: id}, nil
}

// UpdateStaff 更新员工信息
func (s *AdminService) UpdateStaff(ctx context.Context, req *adminv1.UpdateStaffRequest) (*adminv1.UpdateStaffReply, error) {
	err := s.sc.Update(ctx, &biz.Staff{
		ID:         req.Id,
		Name:       req.Name,
		Role:       req.Role,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
		Bio:        req.Bio,
	})
	if err != nil {
		return nil, err
	}
	return &adminv1.UpdateStaffReply{Success: true}, nil
}

// ToggleStaffStatus 切换员工状态
func (s *AdminService) ToggleStaffStatus(ctx context.Context, req *adminv1.ToggleStaffStatusRequest) (*adminv1.ToggleStaffStatusReply, error) {
	newStatus, err := s.sc.ToggleStatus(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &adminv1.ToggleStaffStatusReply{Success: true, NewStatus: int32(newStatus)}, nil
}

// GetRoleNameMap 获取角色名称映射（无需角色管理权限）
func (s *AdminService) GetRoleNameMap(ctx context.Context, req *adminv1.GetRoleNameMapRequest) (*adminv1.GetRoleNameMapReply, error) {
	items, err := s.rc.ListRoles(ctx)
	if err != nil {
		s.log.Errorf("获取角色名称映射失败: %v", err)
		return &adminv1.GetRoleNameMapReply{Items: map[string]string{}}, nil
	}
	nameMap := make(map[string]string)
	for _, item := range items {
		nameMap[item.Slug] = item.Name
	}
	return &adminv1.GetRoleNameMapReply{Items: nameMap}, nil
}
