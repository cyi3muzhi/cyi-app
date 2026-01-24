package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type UserRepo interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	Save(ctx context.Context, user *User) error
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
