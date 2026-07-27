package service

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
)

// AdminBannerService Banner 管理 gRPC 服务
type AdminBannerService struct {
	v1.UnimplementedAdminBannerServer
	uc *biz.BannerUseCase
}

// NewAdminBannerService 创建 AdminBannerService
func NewAdminBannerService(uc *biz.BannerUseCase) *AdminBannerService {
	return &AdminBannerService{uc: uc}
}

// GetBanners 获取 Banner 列表
func (s *AdminBannerService) GetBanners(ctx context.Context, _ *v1.GetBannersRequest) (*v1.GetBannersReply, error) {
	banners, err := s.uc.GetBanners(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.BannerItem, 0, len(banners))
	for _, b := range banners {
		items = append(items, &v1.BannerItem{
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
	return &v1.GetBannersReply{Banners: items}, nil
}

// SaveBanners 保存 Banner（全量替换）
func (s *AdminBannerService) SaveBanners(ctx context.Context, req *v1.SaveBannersRequest) (*v1.SaveBannersReply, error) {
	banners := make([]*biz.Banner, 0, len(req.Banners))
	for _, b := range req.Banners {
		banners = append(banners, &biz.Banner{
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
	if err := s.uc.SaveBanners(ctx, banners); err != nil {
		return nil, err
	}
	return &v1.SaveBannersReply{Success: true}, nil
}
