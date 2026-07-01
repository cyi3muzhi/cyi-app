package biz

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      ProductStatus
	Stock       *Stock
}

func (p *Product) IsValid() bool {
	return p.Status.IsValid()
}
// 初始化库存
func (p *Product) InitStock(total int32) {
	p.Stock = &Stock{
		ProductID: p.ID,
		Total:     total,
		Available: total,
	}
}
