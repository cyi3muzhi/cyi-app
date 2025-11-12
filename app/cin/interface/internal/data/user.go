package data

import (
	"context"
	"time"

	v1 "github.com/cin/cin-app/api/service/user/v1"
	"github.com/cin/cin-app/app/cin/interface/internal/biz"
	cr "github.com/cin/cin-app/pkg/util/redis"
	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	log  *log.Helper
	data *Data
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// GetUser 调用用户服务获取用户信息
func (u *userRepo) GetUser(ctx context.Context, id int64) (*biz.User, error) {

	userReply := &biz.User{}
	// 从redis中获取用户信息
	err := cr.GetRedis(ctx, u.data.rc, id, userReply)
	if err != nil {
		u.log.Errorf("Failed to get user from redis: %v", err)
		return nil, err
	}
	if userReply.ID == 0 {
		// 从用户服务获取用户信息
		req := &v1.GetUserRequest{
			Id: id,
		}
		// 调用 gRPC 服务方法
		reply, err := u.data.uc.GetUser(ctx, req)
		if err != nil {
			u.log.Errorf("Failed to get user: %v", err)
			return nil, err
		}
		userReply = &biz.User{
			ID: reply.Id, // 其他字段映射
		}
	}

	// 将 gRPC 响应转换为业务模型
	return userReply, nil
}
func (u *userRepo) Save(ctx context.Context, user *biz.User) (*biz.User, error) {
	// 创建 gRPC 请求
	req := &v1.CreateUserRequest{
		Phone:    user.Phone,
		Password: user.Password,
		Username: user.Username,
	}
	// 调用 gRPC 服务方法
	reply, err := u.data.uc.CreateUser(ctx, req)
	if err != nil {
		u.log.Errorf("Failed to save user: %v", err)

		return nil, err
	}
	result := &biz.User{
		ID:       reply.Id,
		Name:     reply.Name,
		Phone:    reply.Phone,
		Username: reply.Username,
	}
	// 存储在redis中
	err = cr.SetRedis(ctx, u.data.rc, result.ID, result, 5*time.Minute)
	if err != nil {
		u.log.Errorf("Failed to save user to redis: %v", err)
		return nil, err
	}
	// 将 gRPC 响应转换为业务模型
	return result, nil
}
