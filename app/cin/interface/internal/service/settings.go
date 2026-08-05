package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

// GetPublicSettings 获取公开系统设置（商城名称、联系方式等）
// 从 user 服务的 AdminSettings gRPC 接口读取，仅返回前台展示所需的字段
func (s *CinInterface) GetPublicSettings(ctx context.Context, _ *emptypb.Empty) (*v1.PublicSettingsReply, error) {
	settings, err := s.settUC.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}
	hotKeywords := settings.HotKeywords
	if hotKeywords == nil {
		hotKeywords = []string{}
	}
	return &v1.PublicSettingsReply{
		ShopName:        settings.ShopName,
		ShopDescription: settings.ShopDescription,
		ContactPhone:    settings.ContactPhone,
		ContactEmail:    settings.ContactEmail,
		IcpRecord:       settings.IcpRecord,
		HotKeywords:     hotKeywords,
	}, nil
}
