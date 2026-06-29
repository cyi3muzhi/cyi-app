package data

import (
	"context"
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
