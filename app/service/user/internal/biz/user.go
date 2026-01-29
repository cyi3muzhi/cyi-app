package biz

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/pkg/encryption"
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
	GetUser(ctx context.Context, u *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	VerifyPassword(ctx context.Context, u *User) error
}

type UserUseCase struct {
	repo UserRepo
}

func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
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
func (uc *UserUseCase)VerifyPassword(ctx context.Context, u *User)error{
	return uc.repo.VerifyPassword(ctx,u)
}