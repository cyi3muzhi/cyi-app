package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type UserRepo interface {
	GetUser(ctx context.Context, id int64) (*User, error)
	Save(ctx context.Context, user *User) (*User, error)
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

	return u.repo.GetUser(ctx, id)

}

func (u *UserUseCase) Register(ctx context.Context, user *User) (bool, error) {
	_, err := u.repo.Save(ctx, user)
	if err != nil {
		return false, err
	}
	return true, nil
}
