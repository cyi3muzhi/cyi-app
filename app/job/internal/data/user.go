package data

import (
	"context"
	"cyi-app/app/job/internal/biz"

	"google.golang.org/protobuf/types/known/emptypb"
)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) UserCount(ctx context.Context) (int64, error) {
	res, err := r.data.userC.UserCount(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return res.Count, nil
}
