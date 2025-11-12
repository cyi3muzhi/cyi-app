package data

import (
	"context"

	"github.com/cin/cin-app/app/service/user/internal/biz"
	cgorm "github.com/cin/cin-app/pkg/util/gorm"
	"gorm.io/gorm"
)

type User struct {
	cgorm.Model
	Name     string `gorm:"column:name"`
	Phone    string `gorm:"column:phone"`
	Password string `gorm:"column:password"`
	Username string `gorm:"column:username"`
}

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}

func (r *userRepo) userDB() gorm.Interface[User] {
	return gorm.G[User](r.data.db)
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &User{
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
