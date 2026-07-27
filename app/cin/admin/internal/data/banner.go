package data

import (
	"context"

	userV1 "cyi-app/api/service/user/v1"
)

// BannerRepo Banner 管理数据仓库（gRPC 代理到 User Service）
type BannerRepo struct {
	client userV1.AdminBannerClient
}

// NewBannerRepo 创建 Banner 数据仓库
func NewBannerRepo(client userV1.AdminBannerClient) *BannerRepo {
	return &BannerRepo{client: client}
}

// List 获取 Banner 列表
func (r *BannerRepo) List(ctx context.Context) (*userV1.GetBannersReply, error) {
	return r.client.GetBanners(ctx, &userV1.GetBannersRequest{})
}

// Save 保存 Banner（全量替换）
func (r *BannerRepo) Save(ctx context.Context, banners []*userV1.BannerItem) error {
	_, err := r.client.SaveBanners(ctx, &userV1.SaveBannersRequest{Banners: banners})
	return err
}
