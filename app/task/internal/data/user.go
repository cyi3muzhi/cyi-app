package data

import (
	"context"
	"cyi-app/api/service/user/v1"
	"cyi-app/app/task/internal/biz"
	"cyi-app/pkg/util/convert"

	"google.golang.org/protobuf/types/known/emptypb"
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

// UserCount 统计用户数量（cron 定时任务调用 user 服务）
func (r *userRepo) UserCount(ctx context.Context) (int64, error) {
	res, err := r.data.userC.UserCount(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return res.Count, nil
}
