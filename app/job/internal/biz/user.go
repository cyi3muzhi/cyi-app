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
	UserCount(ctx context.Context) (int64, error)
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

func (uc *UserUseCase) UserCount(ctx context.Context) (int64, error) {
	return uc.repo.UserCount(ctx)
}
