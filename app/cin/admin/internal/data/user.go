package data

import (
	"context"
	"fmt"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// userRepo 用户数据仓库（通过 gRPC 调用 user service 查询用户端的用户）
// 注意：这里查询的是用户端的用户表，不是管理员表
type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo 创建用户数据仓库
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ListUser 获取用户列表（通过 UserClient 调用 user service）
// 支持分页/关键字/会员等级/状态筛选
func (r *userRepo) ListUser(ctx context.Context, req *biz.UserListRequest) (*biz.UserListReply, error) {
	resp, err := r.data.uc.ListUser(ctx, &userRpc.ListUserRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Level:    req.Level,
		Status:   req.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	items := make([]*biz.UserInfo, 0, len(resp.Items))
	for _, u := range resp.Items {
		items = append(items, r.protoUserInfoToBiz(ctx, u))
	}

	return &biz.UserListReply{
		Items: items,
		Total: resp.Total,
	}, nil
}

// GetUser 获取用户详情
func (r *userRepo) GetUser(ctx context.Context, id int64) (*biz.UserInfo, error) {
	resp, err := r.data.uc.GetUser(ctx, &userRpc.GetUserRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("获取用户详情失败: %w", err)
	}
	if resp.User == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return r.protoUserInfoToBiz(ctx, resp.User), nil
}

// CreateUser 创建用户（通过 user service 创建用户端账号）
func (r *userRepo) CreateUser(ctx context.Context, req *biz.CreateUserRequest) (int64, error) {
	resp, err := r.data.uc.CreateUser(ctx, &userRpc.CreateUserRequest{
		Username:       req.Username,
		Phone:          req.Phone,
		Email:          req.Email,
		Password:       req.Password,
		Nickname:       req.Nickname,
		RegisterSource:  req.RegisterSource,
	})
	if err != nil {
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}
	return resp.Id, nil
}

// UpdateUser 更新用户资料
func (r *userRepo) UpdateUser(ctx context.Context, req *biz.UpdateUserRequest) error {
	_, err := r.data.uc.UpdateUser(ctx, &userRpc.UpdateUserRequest{
		Id:       req.ID,
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Email:    req.Email,
		Gender:   req.Gender,
		Birthday: req.Birthday,
		Level:    req.Level,
	})
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

// DeleteUser 删除用户（软删除，由 user service 处理）
func (r *userRepo) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.data.uc.DeleteUser(ctx, &userRpc.DeleteUserRequest{Id: id})
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// ToggleUserStatus 冻结/解冻用户，返回新的状态（1=正常 2=冻结）
func (r *userRepo) ToggleUserStatus(ctx context.Context, id int64) (int32, error) {
	resp, err := r.data.uc.ToggleUserStatus(ctx, &userRpc.ToggleUserStatusRequest{Id: id})
	if err != nil {
		return 0, fmt.Errorf("切换用户状态失败: %w", err)
	}
	return resp.NewStatus, nil
}

// GetUserStats 获取用户统计数据（由 user service 聚合真实数据）
func (r *userRepo) GetUserStats(ctx context.Context) (*biz.UserStats, error) {
	resp, err := r.data.uc.GetUserStats(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("获取用户统计数据失败: %w", err)
	}
	return &biz.UserStats{
		TotalUsers:  resp.TotalUsers,
		TodayNew:    resp.TodayNew,
		TodayGrowth: resp.TodayNewGrowth,
		ActiveUsers: resp.ActiveUsers,
		MemberUsers: resp.MemberUsers,
		MemberRatio: resp.MemberRatio,
	}, nil
}

// protoUserInfoToBiz 将 user service 的 proto.UserInfo 转换为 admin biz.UserInfo
// 字段一一映射，丢弃 admin 不需要的字段（如 growth_value/birthday 等）；
// 头像为 OSS 路径时解析为可访问的公开 URL
func (r *userRepo) protoUserInfoToBiz(ctx context.Context, u *userRpc.UserInfo) *biz.UserInfo {
	if u == nil {
		return nil
	}
	return &biz.UserInfo{
		ID:             u.Id,
		Username:       u.Username,
		Nickname:       u.Nickname,
		Phone:          u.Phone,
		Email:          u.Email,
		Avatar:         r.data.resolvePresignedURL(ctx, u.Avatar, 3600*24*30),
		Gender:         u.Gender,
		Level:          u.Level,
		LastLoginAt:    u.LastLoginAt,
		RegisterSource: u.RegisterSource,
		Status:         u.Status,
		CreatedAt:      u.CreatedAt,
	}
}
