package data

import (
	"context"
	"cyi-app/app/nats/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// orderRepo 订单数据仓储
type orderRepo struct {
	data *Data
	log  *log.Helper
}

// NewOrderRepo 创建订单数据仓储
func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context) error {
	r.log.WithContext(ctx).Info("data CreateOrder")
	return nil
}
