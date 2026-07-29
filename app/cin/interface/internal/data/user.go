package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"
	"cyi-app/app/cin/interface/internal/conf"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	v1 "cyi-app/api/cin/interface/v1"
	usV1 "cyi-app/api/service/user/v1"

	cr "cyi-app/pkg/util/redis"

	ts "cyi-app/pkg/kratos/server/transporter"

	taskV1 "cyi-app/api/task/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	log  *log.Helper
	data *Data
	sg   *singleflight.Group
	key  string
}

func NewUserRepo(data *Data, authKey *conf.Auth, logger log.Logger) biz.UserRepo {
	return &userRepo{
		log:  log.NewHelper(logger),
		data: data,
		sg:   &singleflight.Group{},
		key:  authKey.ApiKey,
	}
}

// FindByUsername 根据用户名查找用户，使用本地singleflight + 分布式Redis锁防止缓存击穿
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	// 使用本地singleflight防止单机内的并发请求
	result, err, _ := r.sg.Do(fmt.Sprintf("find_user_by_name_%s", username), func() (interface{}, error) {
		// 生成分布式锁键
		lockKey := fmt.Sprintf("lock:user:username:%s", username)
		// 尝试获取分布式锁，防止多实例并发请求，设置10秒过期时间
		lock := cr.AcquireDistributedLock(ctx, r.data.rc, lockKey, 10*time.Second)
		if !lock {
			// 锁获取失败，返回错误或重试
			r.log.Errorf("Failed to acquire distributed lock for user %s", username)
			return nil, fmt.Errorf("failed to acquire distributed lock for user %s", username)
		}
		// 确保锁释放
		defer func() {
			if err := cr.ReleaseDistributedLock(ctx, r.data.rc, lockKey); err != nil {
				r.log.Errorf("Failed to release distributed lock for user %s: %v", username, err)
			}
		}()

		// 调用用户服务获取用户信息
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

func (r *userRepo) AddUser(ctx context.Context, u *biz.User) error {
	// _, err := r.data.uc.CreateUser(ctx, &usV1.CreateUserRequest{
	// 	Username: u.Username,
	// 	Password: u.Password,
	// })
	task := ts.NewAsyncTask(taskV1.TaskName_create_user.String(), &taskV1.CreateUserTask{
		Username: u.Username,
		Password: u.Password,
	}, r.log)
	info, err := r.data.mq.Enqueue(task)
	if err != nil {
		return err
	}
	r.log.Infof("CreateUserTask: %s", info)
	return nil
}

// RemoveUser 移除用户
// TODO: 待业务需求明确后实现
func (r *userRepo) RemoveUser(ctx context.Context, id int64) error {
	// _, err := r.data.uc.DeleteUser(ctx, &usV1.DeleteUserRequest{
	// 	Id: id,
	// })
	return nil
}

// ModifyUser 修改用户
func (r *userRepo) ModifyUser(ctx context.Context, u *biz.User) error {
	_, err := r.data.uc.UpdateUser(ctx, &usV1.UpdateUserRequest{
		Id:     u.ID,
		Name:   u.Name,
		Avatar: u.Avatar,
	})
	if err != nil {
		return err
	}
	return nil
}

// FindUserList 查询用户列表
// TODO: 待业务需求明确后实现
func (r *userRepo) FindUserList(ctx context.Context) ([]*biz.User, error) {
	return nil, nil
}

// VerifyPassword 校验密码是否正确
func (r *userRepo) VerifyPassword(ctx context.Context, u *biz.User, password string) error {
	_, err := r.data.uc.VerifyPassword(ctx, &usV1.VerifyPasswordRequest{
		Password: password,
		Username: u.Username,
	})
	return err
}

// GetUser 获取用户信息
func (r *userRepo) GetUser(ctx context.Context, id int64) (*biz.User, error) {
	result, err, _ := r.sg.Do(fmt.Sprintf("get_user_%d", id), func() (interface{}, error) {
		// 生成分布式锁键
		lockKey := fmt.Sprintf("lock:user:id:%d", id)
		// 尝试获取分布式锁，防止多实例并发请求，设置10秒过期时间
		lock := cr.AcquireDistributedLock(ctx, r.data.rc, lockKey, 10*time.Second)
		if !lock {
			// 锁获取失败，返回错误或重试
			r.log.Errorf("Failed to acquire distributed lock for user %d", id)
			return nil, fmt.Errorf("failed to acquire distributed lock for user %d", id)
		}
		// 确保锁释放
		defer func() {
			if err := cr.ReleaseDistributedLock(ctx, r.data.rc, lockKey); err != nil {
				r.log.Errorf("Failed to release distributed lock for user %d: %v", id, err)
			}
		}()
		user, err := r.data.uc.GetUser(ctx, &usV1.GetUserRequest{
			Id: id,
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

func (r *userRepo) Register(ctx context.Context, req *v1.RegisterReq) error {
	_, err := r.data.uc.Register(ctx, &usV1.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepo) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	reply, err := r.data.uc.Login(ctx, &usV1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		ApiKey:   r.key,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: reply.Token,
	}, nil
}
