package data

import (
	"context"
	"fmt"

	"cyi-app/app/service/user/internal/biz"
	gormImp "cyi-app/app/service/user/internal/gorm_imp"
	mod "cyi-app/app/service/user/internal/model"

	"gorm.io/gorm"
)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) userDB() gorm.Interface[mod.User] {
	return gorm.G[mod.User](r.data.db)
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &mod.User{
		Name:     u.Name,
		Phone:    u.Phone,
		Password: u.Password,
		Username: u.Username,
	}
	if err := r.userDB().Create(ctx, user); err != nil {
		return nil, err
	}
	return &biz.User{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		Username: user.Username,
	}, nil
}

func (r *userRepo) GetUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	var user mod.User
	fmt.Println("go 1")
	user, err := gormImp.Query[mod.User](r.data.db).QueryWith(ctx, mod.User{
		Phone:    u.Phone,
		Username: u.Username,
	})
	if err != nil {
		return nil, err
	}
	return &biz.User{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		Username: user.Username,
	}, nil
}
