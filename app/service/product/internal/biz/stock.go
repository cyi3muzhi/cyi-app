package biz

// Stock 库存业务实体
// SkuID 为0时表示 product 级库存（无SKU场景）
type Stock struct {
	ID        int64
	ProductID int64
	SkuID     int64
	Total     int32
	Available int32
	Reserved  int32
}
