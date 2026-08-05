package data

import (
	"context"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type stockRepo struct {
	data *Data
	log  *log.Helper
}

func NewStockRepo(data *Data, logger log.Logger) biz.StockRepo {
	return &stockRepo{data: data, log: log.NewHelper(logger)}
}

// CreateStock 创建库存记录
func (r *stockRepo) CreateStock(ctx context.Context, in *biz.Stock) error {
	stock := &mod.Stock{}
	convert.Convert(in, stock)
	if err := curd.Stock.WithContext(ctx).Create(stock); err != nil {
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
		stock := &mod.Stock{}
		convert.Convert(in, stock)
		_, err := curd.Stock.WithContext(ctx).
			Where(curd.Stock.ID.Eq(in.ID)).
			UpdateSimple(
				curd.Stock.Total.Value(in.Total),
				curd.Stock.Available.Value(in.Available),
			)
		return err
	}
	return r.updateStockByProduct(ctx, in.ProductID, in.Total)
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

// ReserveStock 预留库存（下单时）
// available- , reserved+ , version++
// 条件: version 匹配 且 available >= count
func (r *stockRepo) ReserveStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.SkuID.Eq(skuID)).
		Where(curd.Stock.Version.Eq(version)).
		Where(curd.Stock.Available.Gte(count)).
		UpdateSimple(
			curd.Stock.Available.Add(0-count),
			curd.Stock.Reserved.Add(count),
			curd.Stock.Version.Add(1),
		)
	if err != nil {
		return 0, err
	}
	if result.RowsAffected == 0 {
		return 0, v1.ErrorInsufficientStock("insufficient stock or version conflict")
	}
	return version + 1, nil
}

// DeductStock 确认扣减库存（支付后）
// total- , reserved- , version++，同时累加 SKU 与商品销量
// 条件: version 匹配
func (r *stockRepo) DeductStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.SkuID.Eq(skuID)).
		Where(curd.Stock.Version.Eq(version)).
		UpdateSimple(
			curd.Stock.Total.Add(0-count),
			curd.Stock.Reserved.Add(0-count),
			curd.Stock.Version.Add(1),
		)
	if err != nil {
		return 0, err
	}
	if result.RowsAffected == 0 {
		return 0, v1.ErrorStockVersionConflict("deduct stock version conflict")
	}
	// 支付成功：SKU 与商品销量各累加 count
	if err := r.increaseSales(ctx, skuID, count); err != nil {
		return 0, err
	}
	return version + 1, nil
}

// ReleaseStock 释放预留库存（取消未支付订单时）
// reserved- , available+ , version++
// 条件: version 匹配 且 reserved >= count
func (r *stockRepo) ReleaseStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.SkuID.Eq(skuID)).
		Where(curd.Stock.Version.Eq(version)).
		Where(curd.Stock.Reserved.Gte(count)).
		UpdateSimple(
			curd.Stock.Reserved.Add(0-count),
			curd.Stock.Available.Add(count),
			curd.Stock.Version.Add(1),
		)
	if err != nil {
		return 0, err
	}
	if result.RowsAffected == 0 {
		return 0, v1.ErrorInsufficientStock("release stock failed: version conflict or insufficient reserved stock")
	}
	return version + 1, nil
}

// AddStock 增加库存（退款或取消已支付订单时）
// total+ , available+ , version++，同时扣减 SKU 与商品销量
// 条件: version 匹配
func (r *stockRepo) AddStock(ctx context.Context, skuID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.SkuID.Eq(skuID)).
		Where(curd.Stock.Version.Eq(version)).
		UpdateSimple(
			curd.Stock.Total.Add(count),
			curd.Stock.Available.Add(count),
			curd.Stock.Version.Add(1),
		)
	if err != nil {
		return 0, err
	}
	if result.RowsAffected == 0 {
		return 0, v1.ErrorStockVersionConflict("add stock version conflict")
	}
	// 退款/取消已支付订单：销量回退（下限为 0）
	if err := r.decreaseSales(ctx, skuID, count); err != nil {
		return 0, err
	}
	return version + 1, nil
}

// increaseSales 累加 SKU 与商品销量（支付成功后调用）
// 使用原生 SQL 累加，因为 gorm gen 生成的 Product 字段未包含 sales 列
func (r *stockRepo) increaseSales(ctx context.Context, skuID int64, count int32) error {
	// 查询 SKU 归属商品
	var sku mod.Sku
	if err := curd.Sku.WithContext(ctx).UnderlyingDB().First(&sku, skuID).Error; err != nil {
		return err
	}
	db := curd.Sku.WithContext(ctx).UnderlyingDB()
	if err := db.Model(&mod.Sku{}).Where("id = ?", skuID).
		UpdateColumn("sales", gorm.Expr("sales + ?", count)).Error; err != nil {
		return err
	}
	if err := db.Model(&mod.Product{}).Where("id = ?", sku.ProductID).
		UpdateColumn("sales", gorm.Expr("sales + ?", count)).Error; err != nil {
		return err
	}
	return nil
}

// decreaseSales 扣减 SKU 与商品销量（退款/取消已支付订单时调用），下限为 0
func (r *stockRepo) decreaseSales(ctx context.Context, skuID int64, count int32) error {
	// 查询 SKU 归属商品
	var sku mod.Sku
	if err := curd.Sku.WithContext(ctx).UnderlyingDB().First(&sku, skuID).Error; err != nil {
		return err
	}
	db := curd.Sku.WithContext(ctx).UnderlyingDB()
	if err := db.Model(&mod.Sku{}).Where("id = ?", skuID).
		UpdateColumn("sales", gorm.Expr("GREATEST(sales - ?, 0)", count)).Error; err != nil {
		return err
	}
	if err := db.Model(&mod.Product{}).Where("id = ?", sku.ProductID).
		UpdateColumn("sales", gorm.Expr("GREATEST(sales - ?, 0)", count)).Error; err != nil {
		return err
	}
	return nil
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
	if len(skuIDs) == 0 {
		return nil, nil
	}
	stocks, err := curd.Stock.WithContext(ctx).Where(curd.Stock.SkuID.In(skuIDs...)).Find()
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
			Version:   s.Version,
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
	_, err := curd.Stock.WithContext(ctx).Where(curd.Stock.SkuID.Eq(skuID)).Delete()
	return err
}
