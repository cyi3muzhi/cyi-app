package biz

import "context"

// StockOpKey 库存操作幂等键（DTM 屏障三元组）
// 由订单服务按业务事务构造并经 NATS 透传，库存服务据此以屏障事务执行库存变更：
// 同一 (Gid, BranchID, Op) 的操作在 DTM 重试/重复请求下只生效一次，并自动处理空补偿/悬挂。
// Gid 为空表示无幂等键（管理端 gRPC 直调等非事务场景），按普通条件更新执行。
type StockOpKey struct {
	Gid      string // 全局事务ID（TCC 下单为 tcc-{orderNo}，支付扣减为 pay-{orderNo}）
	BranchID string // 分支ID（按 SKU 维度：stock-{skuID}）
	Op       string // 屏障操作类型：try / confirm / cancel
}

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
	// ReserveStock 按 SKU ID 预留库存（原子条件更新：available >= count）
	// key 为 DTM 屏障幂等键（Gid 为空时按普通更新执行）
	ReserveStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// DeductStock 按 SKU ID 确认扣减库存（原子条件更新：reserved >= count）
	DeductStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// ReleaseStock 按 SKU ID 释放预留库存（原子条件更新：reserved >= count；预留不足时视为幂等空操作）
	ReleaseStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// AddStock 按 SKU ID 增加库存（原子更新，用于取消已支付/退款时逆扣减）
	AddStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
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

// ReserveStock 按 SKU ID 预留库存（原子条件更新）
func (uc *StockUseCase) ReserveStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
	return uc.repo.ReserveStock(ctx, key, skuID, count)
}

// DeductStock 按 SKU ID 确认扣减库存（原子条件更新）
func (uc *StockUseCase) DeductStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
	return uc.repo.DeductStock(ctx, key, skuID, count)
}

// ReleaseStock 按 SKU ID 释放预留库存（原子条件更新）
func (uc *StockUseCase) ReleaseStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
	return uc.repo.ReleaseStock(ctx, key, skuID, count)
}

// AddStock 按 SKU ID 增加库存（原子更新）
func (uc *StockUseCase) AddStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error {
	return uc.repo.AddStock(ctx, key, skuID, count)
}

// GetStock 按 SKU ID 获取库存
func (uc *StockUseCase) GetStock(ctx context.Context, skuID int64) (*Stock, error) {
	return uc.repo.GetStock(ctx, skuID)
}
