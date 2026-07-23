package biz

// OrderItem 订单行项目（聚合内的实体）
type OrderItem struct {
	ProductID   int64
	SkuID       int64  // SKU ID，关联具体规格
	SpecDesc    string // 规格描述快照（下单时记录，如"红色/大号"）
	ProductName string
	Image       string // 商品图片快照（优先SKU图片，无则SPU主图）
	Price       int64
	Quantity    int32
	Currency    string
}

// SubTotal 小计
func (i *OrderItem) SubTotal() Money {
	return Money{i.Price * int64(i.Quantity), i.Currency}
}
