package biz

// OrderItem 订单行项目（聚合内的实体）
type OrderItem struct {
	ProductID   int64
	ProductName string
	Price       int64
	Quantity    int32
	Currency    string
}

// SubTotal 小计
func (i *OrderItem) SubTotal() Money {
	return Money{i.Price * int64(i.Quantity), i.Currency}
}
