package biz

import (
	"crypto/rand"
	"fmt"
	"time"

	v1 "cyi-app/api/service/order/v1"
)



// Order 订单聚合根
type Order struct {
	ID         int64
	UserID     int64
	OrderNo    string
	TotalPrice Money
	PayAmount  Money
	Freight    Money // 运费（单位：分）
	Discount   Money // 优惠金额（单位：分）
	Status     v1.OrderStatus
	CreateTime int64 // 创建时间（Unix 时间戳秒）

	Items         []*OrderItem // 聚合内的实体
	statusHandler OrderStatus  // 状态模式
}

// AddItem 添加商品（保持一致性）
func (o *Order) AddItem(item *OrderItem) {
	o.Items = append(o.Items, item)
	o.TotalPrice = o.TotalPrice.Add(item.SubTotal())
}

// VerifyOwner 校验订单归属（领域行为）
func (o *Order) VerifyOwner(userID int64) error {
	if o.UserID != userID {
		return v1.ErrorUserNotOrder("user not order")
	}
	return nil
}

// GenerateOrderNo 生成订单号（CIN开头，后接年月日时分秒+6位随机hex）
// 格式: CIN240723163206a3f2b1 → CIN + YYMMDDHHmmss + 6位hex
func (o *Order) GenerateOrderNo() {
	now := time.Now()
	b := make([]byte, 3)
	rand.Read(b)
	o.OrderNo = fmt.Sprintf("CIN%s%06x", now.Format("060102150405"), b)
}
