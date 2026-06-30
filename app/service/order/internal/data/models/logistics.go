package models

import (
	"cyi-app/pkg/util/gorm"
)

// OrderLogistics 订单物流信息表
type OrderLogistics struct {
	gorm.Model
	// 订单ID
	OrderID int64 `json:"order_id" gorm:"index;not null;comment:订单ID"`
	// 收货人
	Receiver string `json:"receiver" gorm:"type:varchar(64);not null;comment:收货人"`
	// 手机号
	Phone string `json:"phone" gorm:"type:varchar(20);not null;comment:手机号"`
	// 收货地址
	Address string `json:"address" gorm:"type:varchar(256);not null;comment:收货地址"`
	// 快递单号
	ExpressNo string `json:"express_no" gorm:"type:varchar(64);default:'';comment:快递单号"`
}

type IOrderLogistics interface {
}
