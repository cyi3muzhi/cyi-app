package data

import (
	"context"
	"cyi-app/api/service/user/v1"
	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/util/convert"
)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) CreateUser(ctx context.Context, user *biz.User) error {
	in := &v1.CreateUserRequest{}
	convert.Convert(user, in)
	_, err := r.data.userC.CreateUser(ctx, in)
	return err
}
