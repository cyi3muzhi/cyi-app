package biz

// START 商品状态
type ProductStatus int32

const (
	// ProductStatusOffSale 下架
	ProductStatusOffSale ProductStatus = 0
	// ProductStatusOnSale 上架
	ProductStatusOnSale ProductStatus = 1
)

func (s ProductStatus)IsOnSale() bool {
	return s == ProductStatusOnSale
}
func (s ProductStatus)IsOffSale() bool {
	return s == ProductStatusOffSale
}
func (s ProductStatus)IsValid() bool {
	return s == ProductStatusOnSale || s == ProductStatusOffSale
}
func (s ProductStatus)String() string {
	if s.IsOnSale() {
		return "上架"
	}
	if s.IsOffSale() {
		return "下架"
	}
	return "未知状态"
}
func (s ProductStatus)Value() int32 {
	return int32(s)
}
// END 商品状态