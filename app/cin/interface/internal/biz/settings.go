package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// PublicSettings 公开系统设置（前台展示用）
type PublicSettings struct {
	ShopName        string
	ShopDescription string
	ContactPhone    string
	ContactEmail    string
	IcpRecord       string
	HotKeywords     []string  // 热门搜索词（用户端搜索框联想展示）
}

// SettingsRepo 设置仓库接口
type SettingsRepo interface {
	GetPublicSettings(ctx context.Context) (*PublicSettings, error)
}

// SettingsUseCase 设置用例
type SettingsUseCase struct {
	repo SettingsRepo
	log  *log.Helper
}

// NewSettingsUseCase 创建 SettingsUseCase
func NewSettingsUseCase(repo SettingsRepo, logger log.Logger) *SettingsUseCase {
	return &SettingsUseCase{repo: repo, log: log.NewHelper(logger)}
}

// GetPublicSettings 获取公开设置
func (uc *SettingsUseCase) GetPublicSettings(ctx context.Context) (*PublicSettings, error) {
	return uc.repo.GetPublicSettings(ctx)
}
