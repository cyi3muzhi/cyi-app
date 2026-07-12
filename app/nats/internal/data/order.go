package data

import (
	"context"
	"cyi-app/app/nats/internal/biz"
	"fmt"
)

// orderRepo 订单数据仓储
type orderRepo struct {
	data *Data
}

// NewOrderRepo 创建订单数据仓储
func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{
		data: data,
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context) error {
	fmt.Println("data CreateOrder")
	return nil
}
