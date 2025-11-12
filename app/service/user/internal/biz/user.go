package biz

import "context"

type UserRepo interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
}

type UserUseCase struct {
	repo UserRepo
}

func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, u *User) (*User, error) {
	// TODO: 密码加密
	return uc.repo.CreateUser(ctx, u)
}
