package data

import (
	"context"
	"fmt"
	"time"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/models"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// adminUserRepo 管理员数据仓库
type adminUserRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminUserRepo 创建管理员数据仓库
func NewAdminUserRepo(data *Data, logger log.Logger) biz.AdminUserRepo {
	return &adminUserRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// FindByUsername 根据用户名查找管理员
func (r *adminUserRepo) FindByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var admin models.AdminUser
	result := r.data.db.Where("username = ?", username).First(&admin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return modelToBiz(&admin), nil
}

// FindByID 根据ID查找管理员
func (r *adminUserRepo) FindByID(ctx context.Context, id int64) (*biz.AdminUser, error) {
	var admin models.AdminUser
	result := r.data.db.First(&admin, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return modelToBiz(&admin), nil
}

// Create 创建管理员
func (r *adminUserRepo) Create(ctx context.Context, admin *biz.AdminUser) (int64, error) {
	m := bizToModel(admin)
	result := r.data.db.Create(m)
	if result.Error != nil {
		return 0, result.Error
	}
	return m.ID, nil
}

// Update 更新管理员
func (r *adminUserRepo) Update(ctx context.Context, admin *biz.AdminUser) error {
	updates := map[string]interface{}{}
	if admin.Name != "" {
		updates["name"] = admin.Name
	}
	if admin.Role != "" {
		updates["role"] = admin.Role
	}
	if admin.Email != "" {
		updates["email"] = admin.Email
	}
	if admin.Status > 0 {
		updates["status"] = admin.Status
	}
	if admin.Password != "" {
		updates["password"] = admin.Password
	}
	if len(updates) == 0 {
		return nil
	}
	return r.data.db.Model(&models.AdminUser{}).Where("id = ?", admin.ID).Updates(updates).Error
}

// List 管理员列表
func (r *adminUserRepo) List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*biz.AdminUser, int64, error) {
	var total int64
	query := r.data.db.Model(&models.AdminUser{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	var admins []models.AdminUser
	offset := int((page - 1) * pageSize)
	if err := query.Order("id DESC").Offset(offset).Limit(int(pageSize)).Find(&admins).Error; err != nil {
		return nil, 0, err
	}
	var result []*biz.AdminUser
	for i := range admins {
		result = append(result, modelToBiz(&admins[i]))
	}
	return result, total, nil
}

// ToggleStatus 切换启用/冻结状态
func (r *adminUserRepo) ToggleStatus(ctx context.Context, id int64) (int32, error) {
	var admin models.AdminUser
	if err := r.data.db.First(&admin, id).Error; err != nil {
		return 0, err
	}
	newStatus := int32(2)
	if admin.Status == 2 {
		newStatus = 1
	}
	if err := r.data.db.Model(&models.AdminUser{}).Where("id = ?", id).Update("status", newStatus).Error; err != nil {
		return 0, err
	}
	return newStatus, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *adminUserRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.data.db.Model(&models.AdminUser{}).Where("id = ?", id).Update("last_login", time.Now().Unix()).Error
}

// VerifyPassword 验证管理员密码
func (r *adminUserRepo) VerifyPassword(ctx context.Context, id int64, password string) error {
	var admin models.AdminUser
	if err := r.data.db.First(&admin, id).Error; err != nil {
		return err
	}
	if !encryption.CheckPassword(password, admin.Password) {
		r.log.Warnf("管理员密码验证失败: id=%d", id)
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SaveToken 将 token 存入 Redis，24h 后自动过期
// key 格式: admin_token:{adminID}，新登录覆盖旧的，实现单点登录
func (r *adminUserRepo) SaveToken(ctx context.Context, token string, adminID int64) error {
	key := fmt.Sprintf("admin_token:%d", adminID)
	return r.data.cache.Set(ctx, key, token, 24*time.Hour).Err()
}

// RemoveToken 从 Redis 中移除 token（登出）
func (r *adminUserRepo) RemoveToken(ctx context.Context, token string) error {
	adminID, _, err := parseJWTClaims(token)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("admin_token:%d", adminID)
	return r.data.cache.Del(ctx, key).Err()
}

// ValidateToken 验证 token 是否有效（未被新登录覆盖）
func (r *adminUserRepo) ValidateToken(ctx context.Context, token string) (bool, int64, string, error) {
	adminID, _, err := parseJWTClaims(token)
	if err != nil {
		return false, 0, "", err
	}
	key := fmt.Sprintf("admin_token:%d", adminID)
	stored, err := r.data.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, adminID, "", nil // 不存在即无效
	}
	if err != nil {
		return false, 0, "", err
	}
	if stored != token {
		return false, adminID, "", nil // 已被新 token 覆盖
	}
	return true, adminID, "", nil
}

// UpdateAvatar 更新管理员头像
func (r *adminUserRepo) UpdateAvatar(ctx context.Context, adminID int64, avatarURL string) error {
	return r.data.db.Model(&models.AdminUser{}).Where("id = ?", adminID).Update("avatar", avatarURL).Error
}

// parseJWTClaims 从 JWT token 中提取 adminID 和 jti
func parseJWTClaims(tokenString string) (int64, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("无法解析JWT claims")
	}
	// admin_id 在 UserClaims 中字段名为 id
	idFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, "", fmt.Errorf("JWT中缺少id字段")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return 0, "", fmt.Errorf("JWT中缺少jti字段")
	}
	return int64(idFloat), jti, nil
}

// extractJTI 仅提取 jti（兼容旧逻辑）
func extractJTI(tokenString string) (string, error) {
	_, jti, err := parseJWTClaims(tokenString)
	return jti, err
}

// modelToBiz models -> biz 转换
func modelToBiz(m *models.AdminUser) *biz.AdminUser {
	if m == nil {
		return nil
	}
	return &biz.AdminUser{
		ID:        m.ID,
		Username:  m.Username,
		Password:  m.Password,
		Name:      m.Name,
		Role:      m.Role,
		Email:     m.Email,
		Avatar:    m.Avatar,
		Status:    m.Status,
		LastLogin: m.LastLogin,
		CreatedAt: time.Unix(m.CreateTime, 0),
	}
}

// bizToModel biz -> models 转换
func bizToModel(b *biz.AdminUser) *models.AdminUser {
	return &models.AdminUser{
		Username: b.Username,
		Password: b.Password,
		Name:     b.Name,
		Role:     b.Role,
		Email:    b.Email,
		Status:   b.Status,
	}
}
