package data

import (
	"context"
	"database/sql"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/go-kratos/kratos/v2/log"
)

type stockRepo struct {
	data *Data
	log  *log.Helper
	// barrierDB 业务库连接（product 服务 MySQL；DTM 屏障表 dtm_barrier.barrier 与其同一实例，跨库访问）
	barrierDB *sql.DB
}

func NewStockRepo(data *Data, logger log.Logger, barrierDB *sql.DB) biz.StockRepo {
	return &stockRepo{data: data, log: log.NewHelper(logger), barrierDB: barrierDB}
}

// runWithBarrier 在数据库事务内执行库存变更：
// 携带 DTM 屏障幂等键时，屏障记录与库存 SQL 在同一事务提交，保证同一 (gid,branch,op)
// 在 DTM 重试/补偿/重复请求下只生效一次（幂等、空补偿、防悬挂）；
// 无幂等键时（管理端 gRPC 直调）退化为普通事务执行，语义与历史行为一致。
func (r *stockRepo) runWithBarrier(ctx context.Context, key biz.StockOpKey, fn func(ctx context.Context, tx *sql.Tx) error) error {
	if key.Gid == "" {
		tx, err := r.barrierDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := fn(ctx, tx); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}
	barrier := &dtmcli.BranchBarrier{TransType: "tcc", Gid: key.Gid, BranchID: key.BranchID, Op: key.Op}
	return barrier.CallWithDB(r.barrierDB, func(tx *sql.Tx) error {
		return fn(ctx, tx)
	})
}

// CreateStock 创建库存记录
func (r *stockRepo) CreateStock(ctx context.Context, in *biz.Stock) error {
	return createStockTx(ctx, curd.Q, in)
}

// createStockTx 在指定事务（或默认连接）内创建库存记录
// 供 CreateStock 及商品编辑事务（UpdateProductTx）复用
func createStockTx(ctx context.Context, q *curd.Query, in *biz.Stock) error {
	stock := &mod.Stock{}
	convert.Convert(in, stock)
	if err := q.Stock.WithContext(ctx).Create(stock); err != nil {
		return err
	}
	in.ID = stock.ID
	return nil
}

// UpdateStock 更新库存，支持两种场景：
// 1. in.ID > 0：按主键ID更新 total/available（商品编辑时逐条更新 SKU 库存）
// 2. in.ID == 0 && in.ProductID > 0：按商品ID设置总库存（管理端补货/调库存），均分到该商品所有库存记录
func (r *stockRepo) UpdateStock(ctx context.Context, in *biz.Stock) error {
	if in.ID > 0 {
		return updateStockTx(ctx, curd.Q, in)
	}
	return r.updateStockByProduct(ctx, in.ProductID, in.Total)
}

// updateStockTx 在指定事务（或默认连接）内按主键ID更新库存 total/available
// 仅用于已存在的库存记录，保留 reserved；供 UpdateStock 及商品编辑事务（UpdateProductTx）复用
func updateStockTx(ctx context.Context, q *curd.Query, in *biz.Stock) error {
	stock := &mod.Stock{}
	convert.Convert(in, stock)
	_, err := q.Stock.WithContext(ctx).
		Where(q.Stock.ID.Eq(in.ID)).
		UpdateSimple(
			q.Stock.Total.Value(in.Total),
			q.Stock.Available.Value(in.Available),
		)
	return err
}

// updateStockByProduct 按商品ID设置总库存（管理端补货/调库存）
// 将 total 均分到该商品的所有库存记录（余数计入第一条）；无库存记录时创建 product 级记录（sku_id=0）
func (r *stockRepo) updateStockByProduct(ctx context.Context, productID int64, total int32) error {
	stocks, err := curd.Stock.WithContext(ctx).Where(curd.Stock.ProductID.Eq(productID)).Find()
	if err != nil {
		return err
	}
	// 无库存记录：创建 product 级库存占位记录
	if len(stocks) == 0 {
		return curd.Stock.WithContext(ctx).Create(&mod.Stock{
			ProductID: productID,
			Total:     total,
			Available: total,
		})
	}
	// 均分：每份 = total/数量，余数加给第一条
	each := total / int32(len(stocks))
	remainder := total % int32(len(stocks))
	for i, s := range stocks {
		v := each
		if i == 0 {
			v += remainder
		}
		if _, err := curd.Stock.WithContext(ctx).Where(curd.Stock.ID.Eq(s.ID)).
			UpdateSimple(
				curd.Stock.Total.Value(v),
				curd.Stock.Available.Value(v),
			); err != nil {
			return err
		}
	}
	return nil
}

// UpdateStockBySkuID 按 SKU ID 设置库存（按 SKU 明细补货）
// 带 product_id 条件校验 SKU 归属，防止越权更新其他商品的 SKU
func (r *stockRepo) UpdateStockBySkuID(ctx context.Context, productID, skuID int64, total int32) error {
	_, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.ProductID.Eq(productID)).
		Where(curd.Stock.SkuID.Eq(skuID)).
		UpdateSimple(
			curd.Stock.Total.Value(total),
			curd.Stock.Available.Value(total),
		)
	return err
}

// ReserveStock 预留库存（下单时），原子条件更新避免超卖：available-，reserved+
// 条件: available >= count；携带幂等键时由 DTM 屏障保证重复请求不重复预留
func (r *stockRepo) ReserveStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	return r.runWithBarrier(ctx, key, func(ctx context.Context, tx *sql.Tx) error {
		return stockReserveTx(ctx, tx, skuID, count)
	})
}

// DeductStock 确认扣减库存（支付后）：total-，reserved-，同时累加 SKU 与商品销量
// 条件: reserved >= count；携带幂等键时由 DTM 屏障保证重试不重复扣减
func (r *stockRepo) DeductStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	return r.runWithBarrier(ctx, key, func(ctx context.Context, tx *sql.Tx) error {
		return stockDeductTx(ctx, tx, skuID, count)
	})
}

// ReleaseStock 释放预留库存（取消未支付订单时）：reserved-，available+
// 条件: reserved >= count；预留不足（库存已扣减/已释放）时 SQL 不命中，视为幂等空操作返回 nil，
// 避免补偿链路因状态已达目标而永久失败；是否真正执行由 DTM 屏障统一裁决
func (r *stockRepo) ReleaseStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	return r.runWithBarrier(ctx, key, func(ctx context.Context, tx *sql.Tx) error {
		return stockReleaseTx(ctx, tx, skuID, count)
	})
}

// AddStock 增加库存（退款或取消已支付订单时逆扣减）：total+，available+，同时回退 SKU 与商品销量（下限 0）
// 携带幂等键（op=cancel）时由 DTM 屏障保证：仅对应扣减（try）真实执行过才会生效
func (r *stockRepo) AddStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	return r.runWithBarrier(ctx, key, func(ctx context.Context, tx *sql.Tx) error {
		return stockAddTx(ctx, tx, skuID, count)
	})
}

// stockReserveTx 事务内预留库存（available-，reserved+），条件不命中返回库存不足错误
func stockReserveTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	result, err := tx.ExecContext(ctx,
		"UPDATE stock SET available = available - ?, reserved = reserved + ? WHERE sku_id = ? AND available >= ? AND deleted_at IS NULL",
		count, count, skuID, count)
	if err != nil {
		return err
	}
	if rowsAffected(result) == 0 {
		return v1.ErrorInsufficientStock("商品 SKU %d 库存不足", skuID)
	}
	return nil
}

// stockDeductTx 事务内扣减库存（total-，reserved-），条件不命中返回操作失败错误；成功后累加销量
func stockDeductTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	result, err := tx.ExecContext(ctx,
		"UPDATE stock SET total = total - ?, reserved = reserved - ? WHERE sku_id = ? AND reserved >= ? AND deleted_at IS NULL",
		count, count, skuID, count)
	if err != nil {
		return err
	}
	if rowsAffected(result) == 0 {
		return v1.ErrorStockOperationFailed("扣减库存失败 sku=%d: 预留库存不足", skuID)
	}
	return increaseSalesTx(ctx, tx, skuID, count)
}

// stockReleaseTx 事务内释放预留（reserved-，available+）；预留不足时不命中即为幂等空操作
func stockReleaseTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE stock SET reserved = reserved - ?, available = available + ? WHERE sku_id = ? AND reserved >= ? AND deleted_at IS NULL",
		count, count, skuID, count)
	return err
}

// stockAddTx 事务内增加库存（total+，available+），库存记录不存在时返回错误；成功后回退销量
func stockAddTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	result, err := tx.ExecContext(ctx,
		"UPDATE stock SET total = total + ?, available = available + ? WHERE sku_id = ? AND deleted_at IS NULL",
		count, count, skuID)
	if err != nil {
		return err
	}
	if rowsAffected(result) == 0 {
		return v1.ErrorStockOperationFailed("增加库存失败 sku=%d: 库存记录不存在", skuID)
	}
	return decreaseSalesTx(ctx, tx, skuID, count)
}

// rowsAffected 读取 SQL 执行影响行数（读取失败按 0 处理，由调用方按条件不命中兜底）
func rowsAffected(result sql.Result) int64 {
	n, err := result.RowsAffected()
	if err != nil {
		return 0
	}
	return n
}

// increaseSalesTx 事务内累加 SKU 与商品销量（支付扣减成功后调用，与库存变更同事务提交）
func increaseSalesTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	productID, err := querySkuProductID(ctx, tx, skuID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE sku SET sales = sales + ? WHERE id = ? AND deleted_at IS NULL", count, skuID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE product SET sales = sales + ? WHERE id = ? AND deleted_at IS NULL", count, productID); err != nil {
		return err
	}
	return nil
}

// decreaseSalesTx 事务内回退 SKU 与商品销量（退款/取消已支付时调用，下限为 0；CASE WHEN 兼容 MySQL/SQLite）
func decreaseSalesTx(ctx context.Context, tx *sql.Tx, skuID int64, count int32) error {
	productID, err := querySkuProductID(ctx, tx, skuID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE sku SET sales = CASE WHEN sales > ? THEN sales - ? ELSE 0 END WHERE id = ? AND deleted_at IS NULL",
		count, count, skuID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE product SET sales = CASE WHEN sales > ? THEN sales - ? ELSE 0 END WHERE id = ? AND deleted_at IS NULL",
		count, count, productID); err != nil {
		return err
	}
	return nil
}

// querySkuProductID 事务内查询 SKU 归属的商品ID
func querySkuProductID(ctx context.Context, tx *sql.Tx, skuID int64) (int64, error) {
	var productID int64
	if err := tx.QueryRowContext(ctx,
		"SELECT product_id FROM sku WHERE id = ? AND deleted_at IS NULL", skuID).Scan(&productID); err != nil {
		return 0, err
	}
	return productID, nil
}

// GetStock 按 SKU ID 获取库存
func (r *stockRepo) GetStock(ctx context.Context, skuID int64) (*biz.Stock, error) {
	stock, err := curd.Stock.WithContext(ctx).Where(curd.Stock.SkuID.Eq(skuID)).First()
	if err != nil {
		return nil, err
	}
	out := &biz.Stock{}
	convert.Convert(stock, out)
	out.ID = stock.ID
	return out, nil
}

// GetStocksBySkuIDs 批量获取多个SKU的库存，返回 map[skuID]*Stock
func (r *stockRepo) GetStocksBySkuIDs(ctx context.Context, skuIDs []int64) (map[int64]*biz.Stock, error) {
	return getStocksBySkuIDsTx(ctx, curd.Q, skuIDs)
}

// getStocksBySkuIDsTx 在指定事务（或默认连接）内批量获取多个SKU的库存，返回 map[skuID]*Stock
// 供 GetStocksBySkuIDs 及商品编辑事务（UpdateProductTx）复用
func getStocksBySkuIDsTx(ctx context.Context, q *curd.Query, skuIDs []int64) (map[int64]*biz.Stock, error) {
	if len(skuIDs) == 0 {
		return nil, nil
	}
	stocks, err := q.Stock.WithContext(ctx).Where(q.Stock.SkuID.In(skuIDs...)).Find()
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*biz.Stock, len(stocks))
	for _, s := range stocks {
		result[s.SkuID] = &biz.Stock{
			ID:        s.ID,
			ProductID: s.ProductID,
			SkuID:     s.SkuID,
			Total:     s.Total,
			Available: s.Available,
			Reserved:  s.Reserved,
		}
	}
	return result, nil
}

// DeleteStockByProductID 根据商品 ID 批量删除库存记录
func (r *stockRepo) DeleteStockByProductID(ctx context.Context, productID int64) error {
	_, err := curd.Stock.WithContext(ctx).Where(curd.Stock.ProductID.Eq(productID)).Delete()
	return err
}

// DeleteStockBySkuID 按 SKU ID 删除库存记录
func (r *stockRepo) DeleteStockBySkuID(ctx context.Context, skuID int64) error {
	return deleteStockBySkuIDTx(ctx, curd.Q, skuID)
}

// deleteStockBySkuIDTx 在指定事务（或默认连接）内按 SKU ID 删除库存记录
// 供 DeleteStockBySkuID 及商品编辑事务（UpdateProductTx）复用
func deleteStockBySkuIDTx(ctx context.Context, q *curd.Query, skuID int64) error {
	_, err := q.Stock.WithContext(ctx).Where(q.Stock.SkuID.Eq(skuID)).Delete()
	return err
}
