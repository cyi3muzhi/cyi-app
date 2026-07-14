package biz

import (
	"context"
	"time"

	v1 "cyi-app/api/service/user/v1"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AdminUser 管理员用户
type AdminUser struct {
	ID        int64
	Username  string
	Password  string
	Name      string
	Role      string
	Email     string
	Status    int32 // 1=正常 2=冻结
	LastLogin int64
	CreatedAt time.Time
}

// AdminUserRepo 管理员数据仓库接口
type AdminUserRepo interface {
	FindByUsername(ctx context.Context, username string) (*AdminUser, error)
	FindByID(ctx context.Context, id int64) (*AdminUser, error)
	Create(ctx context.Context, admin *AdminUser) (int64, error)
	Update(ctx context.Context, admin *AdminUser) error
	List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*AdminUser, int64, error)
	ToggleStatus(ctx context.Context, id int64) (int32, error)
	UpdateLastLogin(ctx context.Context, id int64) error
	VerifyPassword(ctx context.Context, id int64, password string) error
	SaveToken(ctx context.Context, token string, adminID int64) error
	RemoveToken(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (bool, int64, string, error)
}

// AdminUserUseCase 管理员业务逻辑
type AdminUserUseCase struct {
	repo   AdminUserRepo
	apiKey string
	log    *log.Helper
}

// NewAdminUserUseCase 创建 AdminUserUseCase
func NewAdminUserUseCase(repo AdminUserRepo, apiKey string, logger log.Logger) *AdminUserUseCase {
	return &AdminUserUseCase{
		repo:   repo,
		apiKey: apiKey,
		log:    log.NewHelper(logger),
	}
}

// Login 管理员登录验证，验证通过返回 admin + JWT token，token 存入 Redis
func (uc *AdminUserUseCase) Login(ctx context.Context, username, password string) (*AdminUser, string, error) {
	admin, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", v1.ErrorUserNotFound("管理员账号不存在")
	}
	if admin == nil {
		return nil, "", v1.ErrorUserNotFound("管理员账号不存在")
	}
	if admin.Status == 2 {
		return nil, "", v1.ErrorUserNotFound("管理员账号已冻结")
	}
	if err := uc.repo.VerifyPassword(ctx, admin.ID, password); err != nil {
		return nil, "", v1.ErrorPasswordError("管理员密码错误")
	}
	// 更新最后登录时间
	_ = uc.repo.UpdateLastLogin(ctx, admin.ID)

	// JWT 签发 token
	token, err := uc.generateToken(admin)
	if err != nil {
		uc.log.Errorf("JWT签名失败: %v", err)
		return nil, "", v1.ErrorPasswordError("令牌生成失败")
	}
	// token 存入 Redis（后续登出时移除，实现黑名单）
	if err := uc.repo.SaveToken(ctx, token, admin.ID); err != nil {
		uc.log.Warnf("token存入Redis失败: %v", err)
		// 不影响登录，Redis 不可用时 token 仍可正常使用（仅到期失效）
	}
	return admin, token, nil
}

// Logout 管理员登出，移除 Redis 中的 token
func (uc *AdminUserUseCase) Logout(ctx context.Context, token string) error {
	return uc.repo.RemoveToken(ctx, token)
}

// ValidateToken 验证 token 是否有效（未被新登录覆盖）
func (uc *AdminUserUseCase) ValidateToken(ctx context.Context, token string) (bool, int64, string, error) {
	valid, adminID, username, err := uc.repo.ValidateToken(ctx, token)
	if err != nil {
		return false, 0, "", err
	}
	return valid, adminID, username, nil
}

// generateToken 使用 JWT 签发管理员令牌
func (uc *AdminUserUseCase) generateToken(admin *AdminUser) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, server.UserClaims{
		ID:       admin.ID,
		Username: admin.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cin",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	})
	signedString, err := claims.SignedString([]byte(uc.apiKey))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

// Create 创建管理员
func (uc *AdminUserUseCase) Create(ctx context.Context, admin *AdminUser) (int64, error) {
	return uc.repo.Create(ctx, admin)
}

// Update 更新管理员
func (uc *AdminUserUseCase) Update(ctx context.Context, admin *AdminUser) error {
	return uc.repo.Update(ctx, admin)
}

// Get 获取管理员
func (uc *AdminUserUseCase) Get(ctx context.Context, id int64) (*AdminUser, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 管理员列表
func (uc *AdminUserUseCase) List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*AdminUser, int64, error) {
	return uc.repo.List(ctx, page, pageSize, keyword, status)
}

// ToggleStatus 切换状态
func (uc *AdminUserUseCase) ToggleStatus(ctx context.Context, id int64) (int32, error) {
	return uc.repo.ToggleStatus(ctx, id)
}

// 确保 import 不被移除
var _ = &v1.AdminLoginReply{}
