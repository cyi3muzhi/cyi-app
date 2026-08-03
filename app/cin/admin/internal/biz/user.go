package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// UserInfo 用户信息（管理后台视角）
// 不包含消费金额/订单数（这些数据由 order service 聚合，后续在用户详情里展示）
type UserInfo struct {
	ID             int64
	Username       string
	Nickname       string
	Phone          string
	Email          string
	Avatar         string
	Gender         int32
	Level          string
	LastLoginAt    int64
	RegisterSource string
	Status         int32 // 1=正常 2=冻结
	CreatedAt      int64 // Unix 秒
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int32
	PageSize int32
	Keyword  string
	Level    string
	Status   int32 // 0=全部 1=正常 2=冻结
}

// UserListReply 用户列表响应
type UserListReply struct {
	Items []*UserInfo
	Total int32
}

// UserStats 用户统计数据
type UserStats struct {
	TotalUsers  int64
	TodayNew    int64
	TodayGrowth float64
	ActiveUsers int64
	MemberUsers int64
	MemberRatio float64
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username       string
	Phone          string
	Email          string
	Password       string
	Nickname       string
	RegisterSource string
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID       int64
	Nickname string
	Phone    string
	Email    string
	Gender   int32
	Birthday string
	Level    string
}

// UserRepo 用户数据仓库接口
type UserRepo interface {
	// ListUser 获取用户列表
	ListUser(ctx context.Context, req *UserListRequest) (*UserListReply, error)
	// GetUser 获取用户详情
	GetUser(ctx context.Context, id int64) (*UserInfo, error)
	// CreateUser 创建用户
	CreateUser(ctx context.Context, req *CreateUserRequest) (int64, error)
	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, req *UpdateUserRequest) error
	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, id int64) error
	// ToggleUserStatus 冻结/解冻用户
	ToggleUserStatus(ctx context.Context, id int64) (int32, error)
	// GetUserStats 获取用户统计数据
	GetUserStats(ctx context.Context) (*UserStats, error)
}

// UserUseCase 用户业务逻辑
type UserUseCase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUseCase 创建 UserUseCase
func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// ListUser 获取用户列表
func (uc *UserUseCase) ListUser(ctx context.Context, req *UserListRequest) (*UserListReply, error) {
	return uc.repo.ListUser(ctx, req)
}

// GetUser 获取用户详情
func (uc *UserUseCase) GetUser(ctx context.Context, id int64) (*UserInfo, error) {
	return uc.repo.GetUser(ctx, id)
}

// CreateUser 创建用户
func (uc *UserUseCase) CreateUser(ctx context.Context, req *CreateUserRequest) (int64, error) {
	return uc.repo.CreateUser(ctx, req)
}

// UpdateUser 更新用户
func (uc *UserUseCase) UpdateUser(ctx context.Context, req *UpdateUserRequest) error {
	return uc.repo.UpdateUser(ctx, req)
}

// DeleteUser 删除用户
func (uc *UserUseCase) DeleteUser(ctx context.Context, id int64) error {
	return uc.repo.DeleteUser(ctx, id)
}

// ToggleUserStatus 冻结/解冻用户
func (uc *UserUseCase) ToggleUserStatus(ctx context.Context, id int64) (int32, error) {
	return uc.repo.ToggleUserStatus(ctx, id)
}

// GetUserStats 获取用户统计数据
func (uc *UserUseCase) GetUserStats(ctx context.Context) (*UserStats, error) {
	return uc.repo.GetUserStats(ctx)
}
