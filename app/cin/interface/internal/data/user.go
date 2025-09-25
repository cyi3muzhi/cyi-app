package data

import (
	"context"

	"github.com/cin/cin-app/app/cin/interface/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	log  *log.Helper
	data *Data
	user *biz.UserRepo
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (u *userRepo) GetUser(ctx context.Context, id int64) (*biz.User, error) {
	return nil, nil
}
