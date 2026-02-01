package biz

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type UserRepo interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	Save(ctx context.Context, user *User) error
	// 校验密码是否正确
	VerifyPassword(ctx context.Context, user *User, password string) error
	// 获取用户信息
	GetUser(ctx context.Context, id int64) (*User, error)
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
