package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
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
	// VerifyLogin 验证管理员登录，返回 (admin, token, error)
	VerifyLogin(ctx context.Context, username, password string) (*AdminAuth, string, error)
	// SaveToken 保存 token
	SaveToken(ctx context.Context, userID int64, token string) error
	// GetByToken 根据 token 获取管理员信息
	GetByToken(ctx context.Context, token string) (*AdminAuth, error)
	// RemoveToken 移除 token（登出）
	RemoveToken(ctx context.Context, token string) error
}

// AuthUseCase 认证业务逻辑
type AuthUseCase struct {
	repo AuthRepo
	log  *log.Helper
}

// NewAuthUseCase 创建 AuthUseCase
func NewAuthUseCase(repo AuthRepo, logger log.Logger) *AuthUseCase {
	return &AuthUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Login 验证管理员登录，token 由 user 微服务生成，BFF 透传
func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (*AdminAuth, string, error) {
	return uc.repo.VerifyLogin(ctx, username, password)
}

// Logout 登出，移除 token
func (uc *AuthUseCase) Logout(ctx context.Context, token string) error {
	return uc.repo.RemoveToken(ctx, token)
}

// GetAdminInfo 根据 token 获取管理员信息
func (uc *AuthUseCase) GetAdminInfo(ctx context.Context, token string) (*AdminAuth, error) {
	return uc.repo.GetByToken(ctx, token)
}
