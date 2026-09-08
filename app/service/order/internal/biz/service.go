package biz

import (
	"context"
	"fmt"

	productRpc "cyi-app/api/service/product/v1"
)

// ProductRepo 商品仓储接口（远程调用商品服务）
type ProductRepo interface {
	GetProduct(ctx context.Context, productID int64) (*productRpc.GetProductReply, error)
}

// StockOpKey 库存操作幂等键（DTM 屏障三元组）
// 经 NATS 透传给 product 服务，product 侧以屏障事务执行库存变更：
// 同一 (Gid, BranchID, Op) 在重试/补偿/重复请求下只生效一次，并自动处理空补偿/防悬挂。
// Gid 按业务事务确定性构造（与订单号绑定），保证同一订单的补偿/重试路由到同一屏障记录。
type StockOpKey struct {
	Gid      string // 全局事务ID
	BranchID string // 分支ID（按 SKU 维度：stock-{skuID}）
	Op       string // 屏障操作类型：try / confirm / cancel
}

// tccStockKey 构造下单 TCC 库存预留/释放的屏障键
// Gid=tcc-{orderNo}：Try=预留库存，Cancel=释放预留（DTM 回滚与用户未支付取消共用，补偿天然幂等）
func tccStockKey(orderNo string, skuID int64, op string) StockOpKey {
	return StockOpKey{
		Gid:      fmt.Sprintf("tcc-%s", orderNo),
		BranchID: fmt.Sprintf("stock-%d", skuID),
		Op:       op,
	}
}

// payStockKey 构造支付扣减/退还的屏障键
// Gid=pay-{orderNo}：Try=扣减库存，Cancel=逆扣减（取消已支付/退款/支付失败对账共用，屏障保证只退还一次）
func payStockKey(orderNo string, skuID int64, op string) StockOpKey {
	return StockOpKey{
		Gid:      fmt.Sprintf("pay-%s", orderNo),
		BranchID: fmt.Sprintf("stock-%d", skuID),
		Op:       op,
	}
}

// StockRepo 库存仓储接口（按 SKU 维度操作库存，经 NATS 调 product 服务）
type StockRepo interface {
	GetStock(ctx context.Context, skuID int64) (*productRpc.GetStockReply, error)
	// ReserveStock 预留库存（下单 TCC Try），key 为屏障幂等键
	ReserveStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// DeductStock 扣减库存（支付），reserved-，total-
	DeductStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// ReleaseStock 释放预留（取消未支付/TCC Cancel），reserved-，available+；product 侧条件不命中为幂等空操作
	ReleaseStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
	// AddStock 逆扣减退还库存（取消已支付/退款），total+，available+；屏障保证仅扣减真实执行过才生效
	AddStock(ctx context.Context, key StockOpKey, skuID int64, count int32) error
}
