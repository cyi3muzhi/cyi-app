package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// ListBanners 获取 Banner 列表（代理到 User Service）
func (s *AdminService) ListBanners(ctx context.Context, _ *adminv1.ListBannersRequest) (*adminv1.ListBannersReply, error) {
	reply, err := s.br.List(ctx)
	if err != nil {
		return nil, err
	}
	return &adminv1.ListBannersReply{Banners: reply.Banners}, nil
}

// SaveBanners 保存 Banner（全量替换，代理到 User Service）
func (s *AdminService) SaveBanners(ctx context.Context, req *adminv1.SaveBannersRequest) (*adminv1.SaveBannersReply, error) {
	if err := s.br.Save(ctx, req.Banners); err != nil {
		return nil, err
	}
	return &adminv1.SaveBannersReply{Success: true}, nil
}
