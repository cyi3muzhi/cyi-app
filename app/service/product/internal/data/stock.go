package data

import (
	"context"

	v1 "cyi-app/api/service/product/v1"
	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type stockRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewStockRepo(data *Data, logger log.Logger) biz.StockRepo {
	return &stockRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *stockRepo) CreateStock(ctx context.Context, in *biz.Stock) error {
	stock := &mod.Stock{}
	convert.Convert(in, stock)
	if err := curd.Stock.WithContext(ctx).Create(stock); err != nil {
		return err
	}
	in.ID = stock.ID
	return nil
}

func (r *stockRepo) UpdateStock(ctx context.Context, in *biz.Stock) error {
	stock := &mod.Stock{}
	convert.Convert(in, stock)
	_, err := curd.Stock.Debug().WithContext(ctx).
		Where(curd.Stock.ProductID.Eq(in.ProductID)).
		UpdateSimple(
			curd.Stock.Total.Value(in.Total),
			curd.Stock.Available.Value(in.Total),
		)
	return err
}

// ReserveStock 预留库存（下单时）
// available- , reserved+ , version++
// 条件: version 匹配 且 available >= count
func (r *stockRepo) ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.ProductID.Eq(productID)).
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
// total- , reserved- , version++
// 条件: version 匹配
func (r *stockRepo) DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	result, err := curd.Stock.WithContext(ctx).
		Where(curd.Stock.ProductID.Eq(productID)).
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
	return version + 1, nil
}

func (r *stockRepo) GetStock(ctx context.Context, productID int64) (*biz.Stock, error) {
	stock, err := curd.Stock.WithContext(ctx).Where(curd.Stock.ProductID.Eq(productID)).First()
	if err != nil {
		return nil, err
	}
	out := &biz.Stock{}
	convert.Convert(stock, out)
	out.ID = stock.ID
	return out, nil
}
