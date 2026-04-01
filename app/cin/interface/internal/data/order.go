package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"

	"golang.org/x/sync/singleflight"

	orderV1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type orderRepo struct {
	log  *log.Helper
	data *Data
	sg   *singleflight.Group
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		log:  log.NewHelper(logger),
		data: data,
		sg:   &singleflight.Group{},
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.Order) error {
	_, err := r.data.oc.CreateOrder(ctx, &orderV1.CreateOrderRequest{
		UserId:    order.UserID,
		ProductId: order.ProductID,
		Count:     order.Count,
	})
	if err != nil {
		return err
	}
	return err
}

func (r *orderRepo) PayOrder(ctx context.Context, orderID int64) error {
	_, err := r.data.oc.PayOrder(ctx, &orderV1.PayOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return err
	}
	return nil
}
