package models

import (
	"cyi-app/pkg/util/gorm"
)

// Category 商品分类 GORM 模型
type Category struct {
	gorm.Model
	// 父分类ID（0表示顶级分类）
	ParentID int64 `json:"parent_id" gorm:"index;default:0;comment:父分类ID"`
	// 分类名称
	Name string `json:"name" gorm:"type:varchar(32);not null;comment:分类名称"`
	// 别名
	Slug string `json:"slug" gorm:"type:varchar(64);default:'';comment:别名"`
	// 图标Emoji
	Icon string `json:"icon" gorm:"type:varchar(16);default:'';comment:图标Emoji"`
	// 分类图片
	Image string `json:"image" gorm:"type:varchar(255);default:'';comment:分类图片"`
	// 排序值
	Sort int32 `json:"sort" gorm:"not null;default:0;comment:排序值"`
	// 状态 true-启用 false-禁用
	Status bool `json:"status" gorm:"not null;default:1;comment:状态"`
	// 首页展示
	Showcase bool `json:"showcase" gorm:"not null;default:0;comment:首页展示"`
	// 商品数量
	Count int32 `json:"count" gorm:"not null;default:0;comment:商品数量"`
	// 层级
	Level int32 `json:"level" gorm:"not null;default:0;comment:层级"`
	// SEO标题
	SeoTitle string `json:"seo_title" gorm:"type:varchar(128);default:'';comment:SEO标题"`
	// SEO关键词
	SeoKeywords string `json:"seo_keywords" gorm:"type:varchar(255);default:'';comment:SEO关键词"`
	// SEO描述
	SeoDesc string `json:"seo_desc" gorm:"type:varchar(512);default:'';comment:SEO描述"`
}

// ICategory GORM gen 接口标记
type ICategory interface{}
