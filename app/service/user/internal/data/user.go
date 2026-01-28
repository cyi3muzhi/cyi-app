package data

import (
	"context"

	"github.com/pkg/errors"

	"cyi-app/app/service/user/internal/biz"
	mod "cyi-app/app/service/user/internal/data/models"
	userImp "cyi-app/app/service/user/internal/data/models/user"

	cr "cyi-app/pkg/util/redis"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var userCacheKey = func(username string) string {
	return "user_cache_key_" + username
}

type userRepo struct {
	data  *Data
	db    func(ctx context.Context) *gorm.DB
	cache *redis.Client
	log   *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{data: data, db: data.db, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &mod.User{
		Name:     u.Name,
		Phone:    u.Phone,
		Password: u.Password,
		Username: u.Username,
	}
	// 使用userDB方法简化调用
	if err := userImp.UserRepo[mod.User](r.db(ctx)).Create(ctx, user); err != nil {
		return nil, err
	}
	// cache
	cacheKey := userCacheKey(user.Username)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, redis.KeepTTL); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", user.Username, err)
	}
	return &biz.User{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		Username: user.Username,
	}, nil
}

func (r *userRepo) GetUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user, err := userImp.UserRepo[mod.User](r.db(ctx)).QueryWith(ctx, mod.User{
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

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var target *mod.User // 修复变量名拼写
	cacheKey := userCacheKey(username)

	// 1. 尝试从缓存获取
	err := cr.GetRedis(ctx, r.cache, cacheKey, &target)

	// 2. 缓存命中且无错误
	if err == nil && target != nil {
		return &biz.User{
			ID:       target.ID,
			Name:     target.Name,
			Phone:    target.Phone,
			Username: target.Username,
		}, nil
	}

	// 3. 缓存未命中或错误，查询数据库
	target, err = userImp.UserRepo[*mod.User](r.db(ctx)).QueryWith(ctx, mod.User{
		Username: username,
	})

	if err != nil {
		return nil, err
	}

	// 4. 处理用户不存在的情况
	if target == nil && target.ID == 0 {
		return nil, errors.New("user not found") // 或返回自定义错误
	}

	// 5. 设置缓存，使用明确的过期时间 后期可能根据token持续时间设置
	if err := cr.SetRedis(ctx, r.cache, cacheKey, target, redis.KeepTTL); err != nil {
		// 缓存设置失败，记录日志但不影响主流程
		r.log.Warnf("Failed to set user cache for %s: %v", username, err)
	}

	// 6. 返回转换后的业务对象
	return &biz.User{
		ID:       target.ID,
		Name:     target.Name,
		Phone:    target.Phone,
		Username: target.Username,
	}, nil
}
