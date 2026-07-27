package models

import (
	"encoding/json"

	"cyi-app/pkg/util/gorm"
)

// BannerItem 单个 Banner 数据
type BannerItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Image     string `json:"image"`
	Link      string `json:"link"`
	SortOrder int32  `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
	BgColor   string `json:"bg_color"`
	BtnText   string `json:"btn_text"`
}

// BannersRow 数据库行记录，嵌入 gorm.Model 统一管理时间
type BannersRow struct {
	gorm.Model
	Banners   json.RawMessage `gorm:"type:json;not null"`
	UpdatedBy string          `gorm:"size:64"`
}

// TableName 指定表名
func (BannersRow) TableName() string {
	return "banners"
}
