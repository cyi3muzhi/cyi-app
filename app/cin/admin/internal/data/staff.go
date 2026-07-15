package data

import (
	"context"
	"strings"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// staffRepo 员工数据仓库（通过 gRPC 代理到 user 服务）
type staffRepo struct {
	data *Data
	log  *log.Helper
}

// NewStaffRepo 创建 StaffRepo
func NewStaffRepo(data *Data, logger log.Logger) biz.StaffRepo {
	return &staffRepo{data: data, log: log.NewHelper(logger)}
}

// List 获取员工列表
func (r *staffRepo) List(ctx context.Context, page, pageSize int32, keyword, status, role string) ([]*biz.Staff, int32, error) {
	req := &userV1.ListAdminUserRequest{Page: page, PageSize: pageSize, Keyword: keyword}
	if status != "" {
		req.Status = status
	}
	if role != "" {
		req.Role = role
	}
	resp, err := r.data.auc.ListAdminUser(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Staff, 0, len(resp.Items))
	for _, item := range resp.Items {
		avatar := item.Avatar
		if avatar != "" && !strings.HasPrefix(avatar, "http") {
			// OSS 路径转为预签名 URL
			parts := strings.SplitN(avatar, "/", 2)
			if len(parts) == 2 {
				if presigned, err := r.data.oss.GetPresignedURL(ctx, parts[0], parts[1], 3600*24*30); err == nil {
					avatar = presigned
				}
			}
		}
		items = append(items, &biz.Staff{
			ID:         item.Id,
			Username:   item.Username,
			Name:       item.Name,
			Role:       item.Role,
			Email:      item.Email,
			Phone:      item.Phone,
			Department: item.Department,
			Position:   item.Position,
			Bio:        item.Bio,
			Avatar:     avatar,
			Status:     int(item.Status),
			LastLogin:  item.LastLogin,
			CreatedAt:  item.CreatedAt,
		})
	}
	return items, resp.Total, nil
}

// Create 创建员工
func (r *staffRepo) Create(ctx context.Context, s *biz.Staff, password string) (int64, error) {
	resp, err := r.data.auc.CreateAdminUser(ctx, &userV1.CreateAdminUserRequest{
		Username: s.Username,
		Password: password,
		Name:     s.Name,
		Role:     s.Role,
		Email:    s.Email,
	})
	if err != nil {
		return 0, err
	}
	return resp.Id, nil
}

// Update 更新员工信息
func (r *staffRepo) Update(ctx context.Context, s *biz.Staff) error {
	_, err := r.data.auc.UpdateAdminUser(ctx, &userV1.UpdateAdminUserRequest{
		Id:         s.ID,
		Name:       s.Name,
		Role:       s.Role,
		Email:      s.Email,
		Phone:      s.Phone,
		Department: s.Department,
		Position:   s.Position,
		Bio:        s.Bio,
	})
	return err
}

// ToggleStatus 切换员工状态（启用/禁用）
func (r *staffRepo) ToggleStatus(ctx context.Context, id int64) (int, error) {
	resp, err := r.data.auc.ToggleAdminUserStatus(ctx, &userV1.ToggleAdminUserStatusRequest{Id: id})
	if err != nil {
		return 0, err
	}
	return int(resp.NewStatus), nil
}
