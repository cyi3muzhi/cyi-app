package biz

// Product 商品业务实体
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      ProductStatus
	Sales       int32 // 销量（支付成功后累加，退款时回退）
	CategoryID  int64
	Image       string
	Images      []string `mapstructure:"-"` // 轮播图列表（手动序列化）
	SpecGroups  []*SpecGroup
	Skus        []*Sku
}

// IsValid 校验商品状态是否有效
func (p *Product) IsValid() bool {
	return p.Status.IsValid()
}
