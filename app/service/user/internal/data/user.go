package data

import (
	"context"
	"errors"
	"strconv"
	"time"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/curd"
	mod "cyi-app/app/service/user/internal/data/models"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"cyi-app/pkg/util/convert"
	cr "cyi-app/pkg/util/redis"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

var userCacheKeyByUsername = func(username string) string {
	return "user_cache_key_" + username
}

var userCacheKeyByID = func(id int64) string {
	return "user_cache_key_" + strconv.FormatInt(id, 10)
}

type userRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {

	user := &mod.User{}
	convert.Convert(u, user)
	if err := curd.User.WithContext(ctx).Create(user); err != nil {
		return nil, err
	}
	// cache
	cacheKey := userCacheKeyByUsername(user.Username)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", user.Username, err)
	}
	// 2. 缓存命中且无错误
	cacheKey = userCacheKeyByID(user.ID)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %d: %v", user.ID, err)
	}
	convert.Convert(user, u)
	return u, nil
}
func (r *userRepo) GetUser(ctx context.Context, u *biz.User) (*biz.User, error) {

	// 确定缓存键
	var cacheKey string
	if u.ID > 0 {
		cacheKey = userCacheKeyByID(u.ID)
	} else if u.Username != "" {
		cacheKey = userCacheKeyByUsername(u.Username)
	} else {
		// 没有有效的查询条件
		return nil, errors.New("invalid user query conditions")
	}
	query := &mod.User{}
	convert.Convert(u, query)
	query.ID = u.ID
	// 构建查询条件
	// queryStr, _ := sonic.MarshalString(query)
	// url, err := r.data.oss.UploadFileAndGetURL(ctx, rustfs.UserServiceBucket, "user_query.json", strings.NewReader(queryStr))
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(url)

	// 调用通用函数
	return r.getUser(ctx, cacheKey, query)
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {

	if username == "" {
		return nil, errors.New("username is required")
	}

	// 生成缓存键
	cacheKey := userCacheKeyByUsername(username)

	// 构建查询条件
	query := &mod.User{
		Username: username,
	}

	// 调用通用函数
	return r.getUser(ctx, cacheKey, query)
}

func (r *userRepo) VerifyPassword(ctx context.Context, u *biz.User) error {

	// 查询用户密码
	target, err := curd.User.WithContext(ctx).QueryWith(mod.User{
		Username: u.Username,
	})
	if err != nil {
		return err
	}
	ok := encryption.CheckPassword(u.Password, target.Password)
	if ok {
		r.log.Warnf("Password verification failed for user %s", u.Username)
		return nil
	}
	return errors.New("密码不匹配")
}

// getUser 通用获取用户函数
func (r *userRepo) getUser(ctx context.Context, cacheKey string, query *mod.User) (*biz.User, error) {
	// 1. 尝试从缓存获取
	var target *mod.User
	err := cr.GetRedis(ctx, r.cache, cacheKey, &target)
	if err == nil && target != nil {
		return &biz.User{
			ID:       target.ID,
			Name:     target.Name,
			Phone:    target.Phone,
			Username: target.Username,
		}, nil
	}
	s := curd.User

	// 2. 缓存未命中，从数据库查询
	user, err := s.WithContext(ctx).QueryWith(*query)
	if err != nil {
		return nil, err
	}
	if user == nil || user.ID == 0 {
		return nil, errors.New("user not found")
	}

	// 3. 更新缓存
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, 24*time.Hour); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", cacheKey, err)
	}

	// 4. 返回转换后的业务对象
	return &biz.User{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		Username: user.Username,
	}, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) error {
	user := &mod.User{}
	convert.Convert(u, user)
	info, err := curd.User.WithContext(ctx).Updates(user); 
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return errors.New("user not found")
	}

	// cache
	cacheKey := userCacheKeyByUsername(user.Username)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", user.Username, err)
	}
	// 2. 缓存命中且无错误
	cacheKey = userCacheKeyByID(user.ID)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %d: %v", user.ID, err)
	}
	return nil
}

// 统计用户数量
func (r *userRepo) UserCount(ctx context.Context) (int64, error) {
	return curd.User.WithContext(ctx).Count()
}
