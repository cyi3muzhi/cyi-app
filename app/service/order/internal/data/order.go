package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	mod "cyi-app/app/service/order/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type orderRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

func (r *orderRepo) CreateOrder(ctx context.Context, in *biz.Order) error {
	Order := &mod.Order{}
	convert.Convert(in, Order)
	return nil
}

func (r *orderRepo) UpdateOrder(ctx context.Context, in *biz.Order) error {
	Order := &mod.Order{}
	convert.Convert(in, Order)
	return nil
}
