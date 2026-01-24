package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"
	"fmt"

	"golang.org/x/sync/singleflight"

	usV1 "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	log  *log.Helper
	data *Data
	sg   *singleflight.Group
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		log:  log.NewHelper(logger),
		data: data,
		sg:   &singleflight.Group{},
	}
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	result, err, _ := r.sg.Do(fmt.Sprintf("find_user_by_name_%s", username), func() (interface{}, error) {
		user, err := r.data.uc.GetUserByUsername(ctx, &usV1.GetUserByUsernameReq{
			Username: username,
		})
		if err != nil {
			return nil, err
		}
		return &biz.User{
			ID:       user.Id,
			Username: user.Username,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*biz.User), nil
}

func (rp *userRepo) Save(ctx context.Context, u *biz.User) error {
	_, err := rp.data.uc.CreateUser(ctx, &usV1.CreateUserRequest{
		Username: u.Username,
		Password: u.Password,
	})
	return err
}
