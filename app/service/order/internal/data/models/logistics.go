package models

import (
	"cyi-app/pkg/util/gorm"
)

// OrderLogistics 订单物流信息表
type OrderLogistics struct {
	gorm.Model
	// 订单ID（一个订单只能有一条物流记录）
	OrderID int64 `json:"order_id" gorm:"uniqueIndex;not null;comment:订单ID"`
	// 收货人
	Receiver string `json:"receiver" gorm:"type:varchar(64);not null;comment:收货人"`
	// 手机号
	Phone string `json:"phone" gorm:"type:varchar(20);not null;comment:手机号"`
	// 收货地址
	Address string `json:"address" gorm:"type:varchar(256);not null;comment:收货地址"`
	// 物流公司
	Company string `json:"company" gorm:"type:varchar(64);default:'';comment:物流公司"`
	// 快递单号
	ExpressNo string `json:"express_no" gorm:"type:varchar(64);default:'';comment:快递单号"`
	// 发货时间（Unix 秒，0=未发货）
	ShipTime int64 `json:"ship_time" gorm:"default:0;comment:发货时间（Unix秒）"`
}

type IOrderLogistics interface {
}
