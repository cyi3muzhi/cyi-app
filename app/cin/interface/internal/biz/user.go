package biz

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type UserRepo interface {
	// 添加用户
	AddUser(ctx context.Context, user *User) error
	// 移除用户
	RemoveUser(ctx context.Context, id int64) error
	// 修改用户
	ModifyUser(ctx context.Context, user *User) error
	// 获取用户
	GetUser(ctx context.Context, id int64) (*User, error)
	// 查询用户列表
	FindUserList(ctx context.Context) ([]*User, error)

	FindByUsername(ctx context.Context, username string) (*User, error)
	// 校验密码是否正确
	VerifyPassword(ctx context.Context, user *User, password string) error
	// 获取用户信息

	// 注册用户
	Register(ctx context.Context, req *v1.RegisterReq) error
	// 登录用户
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error)
}

type UserUseCase struct {
	log  *log.Helper
	repo UserRepo
}

func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (u *UserUseCase) GetUser(ctx context.Context, id int64) (*User, error) {
	return nil, nil
}

// 获取用户信息
func (u *UserUseCase) UserInfo(ctx context.Context, id int64) (*User, error) {
	user, err := u.repo.GetUser(ctx, id)
	if err != nil {
		u.log.Errorf("UserInfo err: %v", err)
		return nil, v1.ErrorUserNotExist("user not exist")
	}
	return user, nil
}

// 更新用户信息
func (u *UserUseCase) UpdateUserInfo(ctx context.Context, user *User) error {
	err := u.repo.ModifyUser(ctx, user)
	if err != nil {
		u.log.Errorf("UpdateUserInfo err: %v", err)
		return v1.ErrorUserUpdateFailed("user update failed")
	}
	return nil
}

// 注册用户
func (u *UserUseCase) Register(ctx context.Context, req *v1.RegisterReq) error {
	return u.repo.Register(ctx, req)
}

func (u *UserUseCase) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return u.repo.Login(ctx, req)
}
