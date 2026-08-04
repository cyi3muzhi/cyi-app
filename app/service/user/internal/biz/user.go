package biz

import (
	"context"
	"errors"
	"time"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"github.com/go-kratos/kratos/v2/log"
)

// User 用户业务对象（字段对齐新表结构）
// CreatedAt/UpdatedAt 用 time.Time，repo 层手动与 model 的 int64 互转
type User struct {
	ID             int64
	Username       string
	Phone          string
	Email          string
	Password       string
	Nickname       string
	Avatar         string
	Gender         int32
	Birthday       string
	Status         int32 // 1=正常 2=冻结
	LastLoginAt    int64
	LastLoginIP    string
	RegisterSource string
	Level          string
	GrowthValue    int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserListRequest 列表查询参数
type UserListRequest struct {
	Page     int32
	PageSize int32
	Keyword  string
	Level    string
	Status   int32 // 0=全部 1=正常 2=冻结
}

// UserListReply 列表响应
type UserListReply struct {
	Items []*User
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

// UserRepo 用户数据仓库接口
type UserRepo interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id int64) error
	GetUser(ctx context.Context, u *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	VerifyPassword(ctx context.Context, u *User) error
	UserCount(ctx context.Context) (int64, error)
	// 新增接口
	ListUser(ctx context.Context, req *UserListRequest) (*UserListReply, error)
	ToggleUserStatus(ctx context.Context, id int64) (int32, error)
	GetUserStats(ctx context.Context) (*UserStats, error)
	UpdateUserAvatar(ctx context.Context, userID int64, avatarURL string) error
	// UpdateLastLogin 更新用户最近登录时间和 IP
	UpdateLastLogin(ctx context.Context, id int64, at int64, ip string) error
	// ChangePassword 修改密码（校验旧密码后更新为新密码）
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
}

// UserUseCase 用户业务逻辑
type UserUseCase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUseCase 创建 UserUseCase
func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{repo: repo, log: log.NewHelper(logger)}
}

// CreateUser 创建用户（密码加密 + 默认值兜底）
func (uc *UserUseCase) CreateUser(ctx context.Context, u *User) (*User, error) {
	encryptedPassword, err := encryption.EncryptPassword(u.Password)
	if err != nil {
		return nil, v1.ErrorPasswordEncryptionFailed("%s", err.Error())
	}
	u.Password = encryptedPassword
	// 默认值兜底
	if u.Level == "" {
		u.Level = "regular"
	}
	if u.Status == 0 {
		u.Status = 1
	}
	if u.RegisterSource == "" {
		u.RegisterSource = "web"
	}
	user, err := uc.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, v1.ErrorCreateUserFailed("%s", err.Error())
	}
	return user, nil
}

// GetUser 获取用户
func (uc *UserUseCase) GetUser(ctx context.Context, u *User) (*User, error) {
	return uc.repo.GetUser(ctx, u)
}

// GetUserByUsername 根据用户名获取用户
func (uc *UserUseCase) GetUserByUsername(ctx context.Context, in *v1.GetUserByUsernameReq) (*User, error) {
	user, err := uc.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		return nil, v1.ErrorUserNotFound("%s", err.Error())
	}
	return user, nil
}

// VerifyPassword 校验密码
func (uc *UserUseCase) VerifyPassword(ctx context.Context, u *User) error {
	return uc.repo.VerifyPassword(ctx, u)
}

// UpdateUser 更新用户信息
func (uc *UserUseCase) UpdateUser(ctx context.Context, u *User) error {
	if err := uc.repo.UpdateUser(ctx, u); err != nil {
		return v1.ErrorUpdateUserFailed("%s", err.Error())
	}
	return nil
}

// DeleteUser 删除用户（软删除）
func (uc *UserUseCase) DeleteUser(ctx context.Context, id int64) error {
	if err := uc.repo.DeleteUser(ctx, id); err != nil {
		return v1.ErrorDeleteUserFailed("%s", err.Error())
	}
	return nil
}

// UserCount 统计用户数量
func (uc *UserUseCase) UserCount(ctx context.Context) (int64, error) {
	return uc.repo.UserCount(ctx)
}

// ListUser 用户列表
func (uc *UserUseCase) ListUser(ctx context.Context, req *UserListRequest) (*UserListReply, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	return uc.repo.ListUser(ctx, req)
}

// ToggleUserStatus 冻结/解冻用户
func (uc *UserUseCase) ToggleUserStatus(ctx context.Context, id int64) (int32, error) {
	return uc.repo.ToggleUserStatus(ctx, id)
}

// GetUserStats 用户统计
func (uc *UserUseCase) GetUserStats(ctx context.Context) (*UserStats, error) {
	return uc.repo.GetUserStats(ctx)
}

// UpdateUserAvatar 更新头像
func (uc *UserUseCase) UpdateUserAvatar(ctx context.Context, userID int64, avatarURL string) error {
	return uc.repo.UpdateUserAvatar(ctx, userID, avatarURL)
}

// ChangePassword 修改密码
func (uc *UserUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if err := uc.repo.ChangePassword(ctx, userID, oldPassword, newPassword); err != nil {
		uc.log.Errorf("ChangePassword err: %v", err)
		if errors.Is(err, ErrUserNotFound) {
			return v1.ErrorUserNotFound("%s", err.Error())
		}
		if errors.Is(err, ErrPasswordError) {
			return v1.ErrorPasswordError("%s", err.Error())
		}
		return v1.ErrorUpdateUserFailed("%s", err.Error())
	}
	return nil
}
