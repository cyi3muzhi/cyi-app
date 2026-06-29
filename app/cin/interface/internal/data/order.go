package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"
	cr "cyi-app/pkg/util/redis"
	"fmt"
	"time"

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
	res, err := r.data.oc.CreateOrder(ctx, &orderV1.CreateOrderRequest{
		UserId:    order.UserID,
		ProductId: order.ProductID,
		Count:     order.Count,
	})
	if err != nil {
		return err
	}
	order.ID = res.OrderId
	return nil
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

func (r *orderRepo) GetOrder(ctx context.Context, orderID int64) (*biz.Order, error) {
	result, err, _ := r.sg.Do(fmt.Sprintf("get_order_%d", orderID), func() (interface{}, error) {
		lockKey := fmt.Sprintf("lock:order:id:%d", orderID)
		lock := cr.AcquireDistributedLock(ctx, r.data.rc, lockKey, 10*time.Second)
		if !lock {
			r.log.Errorf("Failed to acquire distributed lock for order %d", orderID)
			return nil, fmt.Errorf("failed to acquire distributed lock for order %d", orderID)
		}
		defer func() {
			if err := cr.ReleaseDistributedLock(ctx, r.data.rc, lockKey); err != nil {
				r.log.Errorf("Failed to release distributed lock for order %d: %v", orderID, err)
			}
		}()

		order, err := r.data.oc.GetOrder(ctx, &orderV1.GetOrderRequest{
			OrderId: orderID,
		})
		if err != nil {
			return nil, err
		}

		return &biz.Order{
			ID:        order.OrderId,
			UserID:    order.UserId,
			ProductID: order.ProductId,
			Count:     order.Count,
			Status:    order.Status,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*biz.Order), nil
}
