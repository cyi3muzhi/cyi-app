package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"
	userV1 "cyi-app/api/service/user/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ListBanners 获取公开 Banner 列表
func (s *CinInterface) ListBanners(ctx context.Context, _ *emptypb.Empty) (*v1.ListBannersReply, error) {
	banners, err := s.bannerUC.ListBanners(ctx)
	if err != nil {
		return nil, err
	}
	// 将 biz.Banner 转换为 proto BannerItem
	items := make([]*userV1.BannerItem, 0, len(banners))
	for _, b := range banners {
		items = append(items, &userV1.BannerItem{
			Id:        b.ID,
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
	return &v1.ListBannersReply{Banners: items}, nil
}
