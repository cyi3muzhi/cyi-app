package biz

import "context"

// StockRepo 库存仓储接口（SKU 维度操作）
type StockRepo interface {
	// CreateStock 创建库存记录（Stock 结构体中含 ProductID 和 SkuID）
	CreateStock(ctx context.Context, in *Stock) error
	// UpdateStock 更新库存
	UpdateStock(ctx context.Context, in *Stock) error
	// UpdateStockBySkuID 按 SKU ID 设置库存（带商品归属校验，按 SKU 明细补货用）
	UpdateStockBySkuID(ctx context.Context, productID, skuID int64, total int32) error
	// DeleteStockByProductID 按商品ID批量删除库存记录
	DeleteStockByProductID(ctx context.Context, productID int64) error
	// DeleteStockBySkuID 按 SKU ID 删除库存记录
	DeleteStockBySkuID(ctx context.Context, skuID int64) error
	// ReserveStock 按 SKU ID 预留库存
	ReserveStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error)
	// DeductStock 按 SKU ID 确认扣减库存
	DeductStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error)
	// ReleaseStock 按 SKU ID 释放预留库存
	ReleaseStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error)
	// AddStock 按 SKU ID 增加库存
	AddStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error)
	// GetStock 按 SKU ID 获取库存
	GetStock(ctx context.Context, skuID int64) (*Stock, error)
	// GetStocksBySkuIDs 批量获取多个SKU的库存，返回 map[skuID]*Stock
	GetStocksBySkuIDs(ctx context.Context, skuIDs []int64) (map[int64]*Stock, error)
}

// StockUseCase 库存用例
type StockUseCase struct {
	repo StockRepo
}

// NewStockUseCase 创建库存用例
func NewStockUseCase(repo StockRepo) *StockUseCase {
	return &StockUseCase{repo: repo}
}

// CreateStock 创建库存
func (uc *StockUseCase) CreateStock(ctx context.Context, in *Stock) error {
	return uc.repo.CreateStock(ctx, in)
}

// UpdateStock 更新库存
func (uc *StockUseCase) UpdateStock(ctx context.Context, in *Stock) error {
	return uc.repo.UpdateStock(ctx, in)
}

// UpdateSkuStocks 按 SKU 批量设置库存（管理端按 SKU 明细补货）
// 逐个 SKU 设置 total/available，任一失败即返回错误（由调用方决定是否重试）
func (uc *StockUseCase) UpdateSkuStocks(ctx context.Context, productID int64, items map[int64]int32) error {
	for skuID, total := range items {
		if err := uc.repo.UpdateStockBySkuID(ctx, productID, skuID, total); err != nil {
			return err
		}
	}
	return nil
}

// ReserveStock 按 SKU ID 预留库存
func (uc *StockUseCase) ReserveStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	return uc.repo.ReserveStock(ctx, skuID, count, version)
}

// DeductStock 按 SKU ID 确认扣减库存
func (uc *StockUseCase) DeductStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	return uc.repo.DeductStock(ctx, skuID, count, version)
}

// ReleaseStock 按 SKU ID 释放预留库存
func (uc *StockUseCase) ReleaseStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	return uc.repo.ReleaseStock(ctx, skuID, count, version)
}

// AddStock 按 SKU ID 增加库存
func (uc *StockUseCase) AddStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	return uc.repo.AddStock(ctx, skuID, count, version)
}

// GetStock 按 SKU ID 获取库存
func (uc *StockUseCase) GetStock(ctx context.Context, skuID int64) (*Stock, error) {
	return uc.repo.GetStock(ctx, skuID)
}
