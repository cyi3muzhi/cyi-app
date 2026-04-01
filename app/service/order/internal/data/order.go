package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"
	"cyi-app/pkg/util/convert"
	"errors"
	"fmt"

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
	order := &mod.Order{}
	fmt.Println("CreateOrder", in)
	convert.Convert(in, order)
	fmt.Println("CreateOrder Convert:", order)
	if err := curd.Order.WithContext(ctx).Create(order); err != nil {
		return err
	}
	in.ID = order.ID
	return nil
}

func (r *orderRepo) UpdateOrder(ctx context.Context, in *biz.Order) error {
	fmt.Println("UpdateOrder", in)
	order := &mod.Order{}
	convert.Convert(in, order)
	fmt.Println("UpdateOrder Convert:", order)
	if line, err := curd.Order.Debug().WithContext(ctx).Where(curd.Order.ID.Eq(in.ID)).Updates(order); err != nil && line.RowsAffected == 0 {
		if line.RowsAffected == 0 {
			return errors.New("order not found")
		}
		return err
	}
	return nil
}

func (r *orderRepo) GetOrder(ctx context.Context, id int64) (*biz.Order, error) {
	order, err := curd.Order.WithContext(ctx).Where(curd.Order.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	if order == nil || order.ID == 0 {
		return nil, errors.New("order not found")
	}
	out := &biz.Order{}
	convert.Convert(order, out)
	return out, nil
}
