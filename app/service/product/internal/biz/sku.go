package biz

import "context"

// Sku SKU 业务实体
// 每个SKU对应一个具体的规格组合，如"颜色:红色,尺码:XL"
type Sku struct {
	ID        int64
	ProductID int64
	SkuCode   string
	Price     int64
	Stock     int32 // 库存数量（从 inventory_stock 表加载，不持久化在 sku 表）
	Sales     int32 // 销量（支付成功后累加，退款时回退）
	Spec      map[string]string // 规格描述，如 {"颜色":"红色","尺码":"XL"}
	Image     string // SKU图片
}

// SkuRepo SKU 仓储接口
type SkuRepo interface {
	// CreateSkus 批量创建SKU
	CreateSkus(ctx context.Context, skus []*Sku) error
	// ReplaceSkus 替换SKU（先删后插）
	ReplaceSkus(ctx context.Context, productID int64, skus []*Sku) error
	// GetSkusByProductID 获取商品的所有SKU
	GetSkusByProductID(ctx context.Context, productID int64) ([]*Sku, error)
	// GetSkusByProductIDs 批量获取多个商品的SKU，返回 map[productID][]*Sku
	GetSkusByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*Sku, error)
	// DeleteSkusByProductID 删除商品的所有SKU
	DeleteSkusByProductID(ctx context.Context, productID int64) error
}

// SkuUseCase SKU 用例
type SkuUseCase struct {
	repo SkuRepo
}

// NewSkuUseCase 创建SKU用例
func NewSkuUseCase(repo SkuRepo) *SkuUseCase {
	return &SkuUseCase{repo: repo}
}

// CreateSkus 批量创建SKU
func (uc *SkuUseCase) CreateSkus(ctx context.Context, skus []*Sku) error {
	return uc.repo.CreateSkus(ctx, skus)
}

// ReplaceSkus 替换SKU
func (uc *SkuUseCase) ReplaceSkus(ctx context.Context, productID int64, skus []*Sku) error {
	return uc.repo.ReplaceSkus(ctx, productID, skus)
}

// GetSkusByProductID 获取商品的SKU列表
func (uc *SkuUseCase) GetSkusByProductID(ctx context.Context, productID int64) ([]*Sku, error) {
	return uc.repo.GetSkusByProductID(ctx, productID)
}

// DeleteSkusByProductID 删除商品的SKU
func (uc *SkuUseCase) DeleteSkusByProductID(ctx context.Context, productID int64) error {
	return uc.repo.DeleteSkusByProductID(ctx, productID)
}
