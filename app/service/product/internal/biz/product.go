package biz

// Product 商品业务实体
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      ProductStatus
	CategoryID  int64
	Image       string
	Sales       int32
	SpecGroups  []*SpecGroup
	Skus        []*Sku
}

// IsValid 校验商品状态是否有效
func (p *Product) IsValid() bool {
	return p.Status.IsValid()
}
