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
	"gorm.io/gorm"
)

// 缓存 key 升级到 v2 前缀：新 User 结构变更后强制失效旧缓存
var userCacheKeyByUsername = func(username string) string {
	return "user_cache_v2_" + username
}

var userCacheKeyByID = func(id int64) string {
	return "user_cache_v2_" + strconv.FormatInt(id, 10)
}

type userRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

// NewUserRepo 创建用户数据仓库
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

// CreateUser 创建用户
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &mod.User{}
	convert.Convert(u, user)
	if err := curd.User.WithContext(ctx).Create(user); err != nil {
		return nil, err
	}
	// 缓存
	cacheKey := userCacheKeyByUsername(user.Username)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", user.Username, err)
	}
	cacheKey = userCacheKeyByID(user.ID)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, cr.OneDay); err != nil {
		r.log.Warnf("Failed to set user cache for %d: %v", user.ID, err)
	}
	return userModelToBiz(user), nil
}

// GetUser 获取用户
func (r *userRepo) GetUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	var cacheKey string
	if u.ID > 0 {
		cacheKey = userCacheKeyByID(u.ID)
	} else if u.Username != "" {
		cacheKey = userCacheKeyByUsername(u.Username)
	} else if u.Phone != "" {
		// phone 暂无独立缓存，使用统一查询路径
		return r.getUserFromDB(ctx, &mod.User{Phone: u.Phone})
	} else {
		return nil, errors.New("invalid user query conditions")
	}
	query := &mod.User{}
	convert.Convert(u, query)
	query.ID = u.ID
	return r.getUser(ctx, cacheKey, query)
}

// FindByUsername 根据用户名查询
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	cacheKey := userCacheKeyByUsername(username)
	query := &mod.User{Username: username}
	return r.getUser(ctx, cacheKey, query)
}

// VerifyPassword 校验密码
func (r *userRepo) VerifyPassword(ctx context.Context, u *biz.User) error {
	target, err := curd.User.WithContext(ctx).QueryWith(mod.User{Username: u.Username})
	if err != nil {
		return err
	}
	ok := encryption.CheckPassword(u.Password, target.Password)
	if !ok {
		// 不记录具体用户名，防止账号枚举
		r.log.Warnf("password verification failed")
		return errors.New("password not match")
	}
	r.log.Infof("password verified successfully")
	return nil
}

// getUser 通用获取用户函数（带缓存）
func (r *userRepo) getUser(ctx context.Context, cacheKey string, query *mod.User) (*biz.User, error) {
	// 1. 尝试从缓存获取
	var target *mod.User
	err := cr.GetRedis(ctx, r.cache, cacheKey, &target)
	if err == nil && target != nil {
		return userModelToBiz(target), nil
	}

	// 2. 缓存未命中，从数据库查询
	return r.getUserFromDB(ctx, query)
}

// getUserFromDB 从数据库查询用户并写缓存
func (r *userRepo) getUserFromDB(ctx context.Context, query *mod.User) (*biz.User, error) {
	user, err := curd.User.WithContext(ctx).QueryWith(*query)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	if user == nil || user.ID == 0 {
		return nil, biz.ErrUserNotFound
	}

	// 3. 更新缓存
	cacheKey := userCacheKeyByID(user.ID)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, 24*time.Hour); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", cacheKey, err)
	}
	cacheKey = userCacheKeyByUsername(user.Username)
	if err := cr.SetRedis(ctx, r.cache, cacheKey, user, 24*time.Hour); err != nil {
		r.log.Warnf("Failed to set user cache for %s: %v", cacheKey, err)
	}

	return userModelToBiz(user), nil
}

// UpdateUser 更新用户
func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) error {
	user := &mod.User{}
	convert.Convert(u, user)
	// convert.Convert（mapstructure）无法映射嵌入 gorm.Model 的主键字段，需显式补上，
	// 否则 GORM Updates 缺少主键条件，报 "WHERE conditions required"
	user.ID = u.ID
	info, err := curd.User.WithContext(ctx).Updates(user)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return biz.ErrUserNotFound
	}

	// 失效缓存
	r.cache.Del(ctx, userCacheKeyByUsername(user.Username))
	r.cache.Del(ctx, userCacheKeyByID(user.ID))
	return nil
}

// UpdateLastLogin 更新用户最近登录时间和 IP
// 登录成功后调用；失败只影响记录，不影响登录本身
func (r *userRepo) UpdateLastLogin(ctx context.Context, id int64, at int64, ip string) error {
	// 先查 username（用于失效缓存键）
	var target mod.User
	if err := r.data.db.WithContext(ctx).Select("username").
		Where("id = ?", id).First(&target).Error; err != nil {
		return err
	}
	// 更新数据库
	if err := r.data.db.WithContext(ctx).Model(&mod.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": at,
			"last_login_ip": ip,
		}).Error; err != nil {
		return err
	}
	// 失效缓存（按 id 和 username 双删，避免读到旧数据）
	r.cache.Del(ctx, userCacheKeyByID(id))
	r.cache.Del(ctx, userCacheKeyByUsername(target.Username))
	return nil
}

// UserCount 统计用户数量
func (r *userRepo) UserCount(ctx context.Context) (int64, error) {
	return curd.User.WithContext(ctx).Count()
}

// ListUser 分页查询用户列表
// 支持关键字模糊搜索（username/phone/nickname/email）+ 等级筛选 + 状态筛选
func (r *userRepo) ListUser(ctx context.Context, req *biz.UserListRequest) (*biz.UserListReply, error) {
	query := r.data.db.Model(&mod.User{})
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("username LIKE ? OR phone LIKE ? OR nickname LIKE ? OR email LIKE ?",
			kw, kw, kw, kw)
	}
	if req.Level != "" {
		query = query.Where("level = ?", req.Level)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []mod.User
	offset := int((req.Page - 1) * req.PageSize)
	if err := query.Order("id DESC").Offset(offset).Limit(int(req.PageSize)).Find(&users).Error; err != nil {
		return nil, err
	}

	items := make([]*biz.User, 0, len(users))
	for i := range users {
		items = append(items, userModelToBiz(&users[i]))
	}
	return &biz.UserListReply{Items: items, Total: int32(total)}, nil
}

// ToggleUserStatus 切换用户状态 1<->2
func (r *userRepo) ToggleUserStatus(ctx context.Context, id int64) (int32, error) {
	var user mod.User
	if err := r.data.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, biz.ErrUserNotFound
		}
		return 0, err
	}
	newStatus := int32(2)
	if user.Status == 2 {
		newStatus = 1
	}
	if err := r.data.db.Model(&mod.User{}).Where("id = ?", id).Update("status", newStatus).Error; err != nil {
		return 0, err
	}
	// 失效缓存
	r.cache.Del(ctx, userCacheKeyByID(id))
	r.cache.Del(ctx, userCacheKeyByUsername(user.Username))
	return newStatus, nil
}

// GetUserStats 用户统计
func (r *userRepo) GetUserStats(ctx context.Context) (*biz.UserStats, error) {
	var stats biz.UserStats
	// 总用户数
	r.data.db.Model(&mod.User{}).Count(&stats.TotalUsers)

	// 今日新增
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	r.data.db.Model(&mod.User{}).Where("create_time >= ?", todayStart).Count(&stats.TodayNew)

	// 昨日新增（用于计算增长率）
	yesterdayStart := todayStart - 86400
	var yesterdayNew int64
	r.data.db.Model(&mod.User{}).Where("create_time >= ? AND create_time < ?", yesterdayStart, todayStart).Count(&yesterdayNew)
	if yesterdayNew > 0 {
		stats.TodayGrowth = float64(stats.TodayNew-yesterdayNew) / float64(yesterdayNew)
	} else if stats.TodayNew > 0 {
		stats.TodayGrowth = 1.0
	}

	// 活跃用户数（30 天内登录过）
	activeStart := now.AddDate(0, 0, -30).Unix()
	r.data.db.Model(&mod.User{}).Where("last_login_at >= ?", activeStart).Count(&stats.ActiveUsers)

	// 会员用户数（level != regular 且 level != ''）
	r.data.db.Model(&mod.User{}).Where("level != ? AND level != ''", "regular").Count(&stats.MemberUsers)

	// 会员占比
	if stats.TotalUsers > 0 {
		stats.MemberRatio = float64(stats.MemberUsers) / float64(stats.TotalUsers)
	}
	return &stats, nil
}

// UpdateUserAvatar 单独更新头像
func (r *userRepo) UpdateUserAvatar(ctx context.Context, userID int64, avatarURL string) error {
	if err := r.data.db.Model(&mod.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		return err
	}
	// 失效缓存
	r.cache.Del(ctx, userCacheKeyByID(userID))
	return nil
}

// ChangePassword 修改密码：校验旧密码后更新为新密码，并失效缓存
func (r *userRepo) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// 1. 查询用户（含密码）
	var target mod.User
	if err := r.data.db.WithContext(ctx).First(&target, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return biz.ErrUserNotFound
		}
		return err
	}
	// 2. 校验旧密码
	if !encryption.CheckPassword(oldPassword, target.Password) {
		return biz.ErrPasswordError
	}
	// 3. 加密新密码
	encrypted, err := encryption.EncryptPassword(newPassword)
	if err != nil {
		r.log.Errorf("encrypt new password failed, userID=%d: %v", userID, err)
		return err
	}
	// 4. 更新密码
	if err := r.data.db.WithContext(ctx).Model(&mod.User{}).
		Where("id = ?", userID).
		Update("password", encrypted).Error; err != nil {
		r.log.Errorf("update password failed, userID=%d: %v", userID, err)
		return err
	}
	// 5. 失效缓存
	r.cache.Del(ctx, userCacheKeyByID(userID))
	r.cache.Del(ctx, userCacheKeyByUsername(target.Username))
	return nil
}

// DeleteUser 软删除用户
func (r *userRepo) DeleteUser(ctx context.Context, id int64) error {
	// 先查出 username（用于失效缓存）
	var user mod.User
	if err := r.data.db.Select("username").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return biz.ErrUserNotFound
		}
		return err
	}
	res := r.data.db.Delete(&mod.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return biz.ErrUserNotFound
	}
	// 失效缓存
	r.cache.Del(ctx, userCacheKeyByID(id))
	r.cache.Del(ctx, userCacheKeyByUsername(user.Username))
	return nil
}

// userModelToBiz models.User -> biz.User 转换
// 手动处理嵌入的 gorm.Model 字段（convert.Convert 无法自动映射嵌入字段）
func userModelToBiz(m *mod.User) *biz.User {
	if m == nil {
		return nil
	}
	return &biz.User{
		ID:             m.ID,
		Username:       m.Username,
		Phone:          m.Phone,
		Email:          m.Email,
		Password:       m.Password,
		Nickname:       m.Nickname,
		Avatar:         m.Avatar,
		Gender:         m.Gender,
		Birthday:       m.Birthday,
		Status:         m.Status,
		LastLoginAt:    m.LastLoginAt,
		LastLoginIP:    m.LastLoginIP,
		RegisterSource: m.RegisterSource,
		Level:          m.Level,
		GrowthValue:    m.GrowthValue,
		CreatedAt:      time.Unix(m.CreateTime, 0),
		UpdatedAt:      time.Unix(m.UpdateTime, 0),
	}
}
