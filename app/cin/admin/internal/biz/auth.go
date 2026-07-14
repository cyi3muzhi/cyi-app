package biz

import (
	"context"
	"time"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AdminAuth 管理员认证信息
type AdminAuth struct {
	ID        int64
	Name      string
	Username  string
	Role      string
	Email     string
	Avatar    string
	CreatedAt time.Time
}

// AuthRepo 认证数据仓库接口
type AuthRepo interface {
	// VerifyLogin 验证管理员登录，返回管理员信息
	VerifyLogin(ctx context.Context, username, password string) (*AdminAuth, error)
	// SaveToken 保存 token 到 Redis
	SaveToken(ctx context.Context, userID int64, token string) error
	// RemoveToken 移除 token（登出）
	RemoveToken(ctx context.Context, token string) error
	// GetAdminInfo 根据管理员 ID 获取管理员信息
	GetAdminInfo(ctx context.Context, adminID int64) (*AdminAuth, error)
	// UpdateAvatar 更新管理员头像（接收 Base64 图片）
	UpdateAvatar(ctx context.Context, adminID int64, avatarBase64 string) (string, error)
}

// AuthUseCase 认证业务逻辑
type AuthUseCase struct {
	repo   AuthRepo
	log    *log.Helper
	apiKey string
}

// NewAuthUseCase 创建 AuthUseCase
func NewAuthUseCase(repo AuthRepo, logger log.Logger, apiKey string) *AuthUseCase {
	return &AuthUseCase{
		repo:   repo,
		log:    log.NewHelper(logger),
		apiKey: apiKey,
	}
}

// Login 验证管理员登录，用 admin BFF 自己的密钥生成独立的 JWT token
func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (*AdminAuth, string, error) {
	admin, err := uc.repo.VerifyLogin(ctx, username, password)
	if err != nil {
		return nil, "", err
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, server.UserClaims{
		ID:       admin.ID,
		Username: admin.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cin.admin",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	})

	token, err := claims.SignedString([]byte(uc.apiKey))
	if err != nil {
		return nil, "", err
	}

	if err := uc.repo.SaveToken(ctx, admin.ID, token); err != nil {
		return nil, "", err
	}

	return admin, token, nil
}

// Logout 登出，移除 token
func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	return uc.repo.RemoveToken(ctx, token)
}

// GetAdminInfo 根据管理员 ID 获取管理员信息
func (uc *AuthUseCase) GetAdminInfo(ctx context.Context, adminID int64) (*AdminAuth, error) {
	return uc.repo.GetAdminInfo(ctx, adminID)
}

// UpdateAvatar 更新管理员头像（接收 Base64 图片）
func (uc *AuthUseCase) UpdateAvatar(ctx context.Context, adminID int64, avatarBase64 string) (string, error) {
	return uc.repo.UpdateAvatar(ctx, adminID, avatarBase64)
}
