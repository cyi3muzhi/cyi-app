package biz

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"github.com/go-kratos/kratos/v2/log"
)

type User struct {
	ID       int64
	Name     string
	Phone    string
	Password string
	Username string
}

type UserRepo interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, u *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	VerifyPassword(ctx context.Context, u *User) error
	UserCount(ctx context.Context) (int64, error)
}

type UserUseCase struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, u *User) (*User, error) {
	// TODO: 密码加密
	encryptedPassword, err := encryption.EncryptPassword(u.Password)
	if err != nil {
		return nil, v1.ErrorPasswordEncryptionFailed("%s", err.Error())
	}
	u.Password = encryptedPassword
	user, err := uc.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, v1.ErrorCreateUserFailed("%s", err.Error())
	}
	return user, nil
}

func (uc *UserUseCase) GetUser(ctx context.Context, u *User) (*User, error) {
	return uc.repo.GetUser(ctx, u)
}

func (uc *UserUseCase) GetUserByUsername(ctx context.Context, in *v1.GetUserByUsernameReq) (*User, error) {
	user, err := uc.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		return nil, v1.ErrorUserNotFound("%s", err.Error())
	}
	return user, nil
}

// 校验密码
func (uc *UserUseCase) VerifyPassword(ctx context.Context, u *User) error {
	return uc.repo.VerifyPassword(ctx, u)
}

// 更新用户信息
func (uc *UserUseCase) UpdateUser(ctx context.Context, u *User) error {
	return uc.repo.UpdateUser(ctx, u)
}

// 统计用户数量
func (uc *UserUseCase) UserCount(ctx context.Context) (int64, error) {
	return uc.repo.UserCount(ctx)
}
