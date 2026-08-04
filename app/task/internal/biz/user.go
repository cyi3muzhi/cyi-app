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

func (uc *UserUseCase) CreateUser(ctx context.Context, user *User) error {
	uc.log.Infof("CreateUser: %v", user)
	return uc.repo.CreateUser(ctx, user)
}

// UserCount 统计用户数量（cron 定时任务调用）
func (uc *UserUseCase) UserCount(ctx context.Context) (int64, error) {
	return uc.repo.UserCount(ctx)
}
