package data

import (
	"context"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// settingsRepo 公开设置数据仓库（gRPC 代理到 User Service 的 AdminSettings 接口）
type settingsRepo struct {
	data *Data
	log  *log.Helper
}

// NewSettingsRepo 创建 settingsRepo
func NewSettingsRepo(data *Data, logger log.Logger) biz.SettingsRepo {
	return &settingsRepo{data: data, log: log.NewHelper(logger)}
}

// GetPublicSettings 从 user 服务获取公开的系统设置（前台展示用）
func (r *settingsRepo) GetPublicSettings(ctx context.Context) (*biz.PublicSettings, error) {
	resp, err := r.data.asc.GetSettings(ctx, &userV1.GetSettingsRequest{})
	if err != nil {
		return nil, err
	}
	hotKeywords := resp.HotKeywords
	if hotKeywords == nil {
		hotKeywords = []string{}
	}
	return &biz.PublicSettings{
		ShopName:        resp.ShopName,
		ShopDescription: resp.ShopDescription,
		ContactPhone:    resp.ContactPhone,
		ContactEmail:    resp.ContactEmail,
		IcpRecord:       resp.IcpRecord,
		HotKeywords:     hotKeywords,
	}, nil
}
