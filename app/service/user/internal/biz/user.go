package biz

import (
	"context"

	"cyi-app/app/service/user/internal/pkg/encryption"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
	GetUser(ctx context.Context, u *User) (*User, error)
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
		return nil, err
	}
	u.Password = encryptedPassword
	return uc.repo.CreateUser(ctx, u)
}

func (uc *UserUseCase) GetUser(ctx context.Context, u *User) (*User, error) {
	return uc.repo.GetUser(ctx, u)
}
