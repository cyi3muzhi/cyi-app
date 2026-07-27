package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Banner Banner 业务模型
type Banner struct {
	ID        int64
	Title     string
	Subtitle  string
	Image     string
	Link      string
	SortOrder int32
	Enabled   bool
	BgColor   string
	BtnText   string
	Template  string // 模板类型：classic | full-image | centered | dark | minimal
	TextColor string // 文字颜色
}

// BannerRepo Banner 仓库接口
type BannerRepo interface {
	GetBanners(ctx context.Context) ([]*Banner, error)
	SaveBanners(ctx context.Context, banners []*Banner) error
}

// BannerUseCase Banner 用例
type BannerUseCase struct {
	repo BannerRepo
	log  *log.Helper
}

// NewBannerUseCase 创建 BannerUseCase
func NewBannerUseCase(repo BannerRepo, logger log.Logger) *BannerUseCase {
	return &BannerUseCase{repo: repo, log: log.NewHelper(logger)}
}

// GetBanners 获取所有 Banner
func (uc *BannerUseCase) GetBanners(ctx context.Context) ([]*Banner, error) {
	return uc.repo.GetBanners(ctx)
}

// SaveBanners 全量保存 Banner
func (uc *BannerUseCase) SaveBanners(ctx context.Context, banners []*Banner) error {
	return uc.repo.SaveBanners(ctx, banners)
}
