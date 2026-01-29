package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	usV1 "cyi-app/api/service/user/v1"

	cr "cyi-app/pkg/util/redis"

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

func (rp *userRepo) Save(ctx context.Context, u *biz.User) error {
	_, err := rp.data.uc.CreateUser(ctx, &usV1.CreateUserRequest{
		Username: u.Username,
		Password: u.Password,
	})
	return err
}

// VerifyPassword 校验密码是否正确
func (rp *userRepo) VerifyPassword(ctx context.Context, u *biz.User, password string) error {
	_, err := rp.data.uc.VerifyPassword(ctx, &usV1.VerifyPasswordRequest{
		Password: password,
		Username: u.Username,
	})
	return err
}
