package models

import (
	"cyi-app/pkg/util/gorm"
)

type Order struct {
	*gorm.Model
	// 用户ID
	UserID int64 `json:"user_id"`
	// 商品ID
	ProductID int64 `json:"product_id"`
	// 数量
	Count int32 `json:"count"`
	// 状态
	Status int32 `json:"status"`
}

type IOrder interface {
}
