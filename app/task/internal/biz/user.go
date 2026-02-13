package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type User struct {
	Username string
	Password string
}

type UserRepo interface {
	CreateUser(ctx context.Context, user *User) error
}

type UserUseCase struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, user *User) error {
	uc.log.Infof("CreateUser: %v", user)
	return uc.repo.CreateUser(ctx, user)
}
