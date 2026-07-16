package data

import (
	"context"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type skuRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

// NewSkuRepo 创建SKU仓储
func NewSkuRepo(data *Data, logger log.Logger) biz.SkuRepo {
	return &skuRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

// CreateSkus 批量创建SKU
func (r *skuRepo) CreateSkus(ctx context.Context, skus []*biz.Sku) error {
	for _, sku := range skus {
		model := &mod.Sku{
			ProductID: sku.ProductID,
			SkuCode:   sku.SkuCode,
			Price:     sku.Price,
			Stock:     sku.Stock,
			Sales:     sku.Sales,
			SpecDesc:  marshalSpec(sku.Spec),
		}
		if err := curd.Sku.WithContext(ctx).Create(model); err != nil {
			return err
		}
		sku.ID = model.ID
	}
	return nil
}

// ReplaceSkus 替换SKU（先删后插）
func (r *skuRepo) ReplaceSkus(ctx context.Context, productID int64, skus []*biz.Sku) error {
	// 删除旧的SKU
	if _, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.Eq(productID)).Delete(); err != nil {
		return err
	}
	// 创建新的SKU
	for _, sku := range skus {
		sku.ProductID = productID
	}
	return r.CreateSkus(ctx, skus)
}

// GetSkusByProductID 获取商品的所有SKU
func (r *skuRepo) GetSkusByProductID(ctx context.Context, productID int64) ([]*biz.Sku, error) {
	models, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.Eq(productID)).Find()
	if err != nil {
		return nil, err
	}
	result := make([]*biz.Sku, 0, len(models))
	for _, m := range models {
		result = append(result, &biz.Sku{
			ID:        m.ID,
			ProductID: m.ProductID,
			SkuCode:   m.SkuCode,
			Price:     m.Price,
			Stock:     m.Stock,
			Sales:     m.Sales,
			Spec:      unmarshalSpec(m.SpecDesc),
		})
	}
	return result, nil
}

// DeleteSkusByProductID 删除商品的所有SKU
func (r *skuRepo) DeleteSkusByProductID(ctx context.Context, productID int64) error {
	_, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.Eq(productID)).Delete()
	return err
}
