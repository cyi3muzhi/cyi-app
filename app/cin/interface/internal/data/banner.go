package data

import (
	"context"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/interface/internal/biz"
)

// bannerRepo Banner 数据仓库（gRPC 代理到 User Service 的 AdminBanner 接口）
type bannerRepo struct {
	client userV1.AdminBannerClient
}

// NewBannerRepo 创建 bannerRepo
func NewBannerRepo(client userV1.AdminBannerClient) biz.BannerRepo {
	return &bannerRepo{client: client}
}

// ListBanners 从 user 服务获取 Banner 列表
func (r *bannerRepo) ListBanners(ctx context.Context) ([]*biz.Banner, error) {
	resp, err := r.client.GetBanners(ctx, &userV1.GetBannersRequest{})
	if err != nil {
		return nil, err
	}
	items := make([]*biz.Banner, 0, len(resp.Banners))
	for _, b := range resp.Banners {
		items = append(items, &biz.Banner{
			ID:        b.Id,
			Title:     b.Title,
			Subtitle:  b.Subtitle,
			Image:     b.Image,
			Link:      b.Link,
			SortOrder: b.SortOrder,
			Enabled:   b.Enabled,
			BgColor:   b.BgColor,
			BtnText:   b.BtnText,
			Template:  b.Template,
			TextColor: b.TextColor,
		})
	}
	return items, nil
}
