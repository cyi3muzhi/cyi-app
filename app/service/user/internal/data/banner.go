package data

import (
	"context"
	"encoding/json"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// bannerRepo Banner 数据仓库
type bannerRepo struct {
	data *Data
	log  *log.Helper
}

// NewBannerRepo 创建 BannerRepo
func NewBannerRepo(data *Data, logger log.Logger) biz.BannerRepo {
	return &bannerRepo{data: data, log: log.NewHelper(logger)}
}

// GetBanners 获取所有 Banner（从 JSON 字段反序列化）
func (r *bannerRepo) GetBanners(ctx context.Context) ([]*biz.Banner, error) {
	var row models.BannersRow
	err := r.data.db.WithContext(ctx).First(&row, 1).Error
	if err == gorm.ErrRecordNotFound {
		return []*biz.Banner{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []*biz.Banner
	if err := json.Unmarshal(row.Banners, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// SaveBanners 全量保存 Banner（替换 JSON 字段全部内容）
func (r *bannerRepo) SaveBanners(ctx context.Context, banners []*biz.Banner) error {
	jsonData, err := json.Marshal(banners)
	if err != nil {
		return err
	}
	var row models.BannersRow
	err = r.data.db.WithContext(ctx).First(&row, 1).Error
	if err == gorm.ErrRecordNotFound {
		// 首次保存，创建记录（ID=1）
		row = models.BannersRow{
			Banners:   jsonData,
			UpdatedBy: "admin",
		}
		row.ID = 1
		return r.data.db.WithContext(ctx).Create(&row).Error
	}
	if err != nil {
		return err
	}
	// 已存在，全量更新 JSON 字段
	return r.data.db.WithContext(ctx).Model(&models.BannersRow{}).
		Where("id = ?", 1).
		Update("banners", string(jsonData)).Error
}
