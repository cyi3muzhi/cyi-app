package biz

import "context"

// Banner 公开 Banner 领域对象（前台展示用）
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

// BannerRepo Banner 数据仓库接口
type BannerRepo interface {
	ListBanners(ctx context.Context) ([]*Banner, error)
}

// BannerUseCase Banner 业务用例
type BannerUseCase struct {
	repo BannerRepo
}

// NewBannerUseCase 创建 BannerUseCase
func NewBannerUseCase(repo BannerRepo) *BannerUseCase {
	return &BannerUseCase{repo: repo}
}

// ListBanners 获取所有 Banner
func (uc *BannerUseCase) ListBanners(ctx context.Context) ([]*Banner, error) {
	return uc.repo.ListBanners(ctx)
}
