package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// OrderUseCase 订单业务逻辑
type OrderUseCase struct {
	repo OrderRepo
	log  *log.Helper
}

// OrderRepo 订单数据仓储接口
type OrderRepo interface {
	CreateOrder(ctx context.Context) error
}

// NewOrderUseCase 创建订单业务逻辑
func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateOrder 创建订单
func (u *OrderUseCase) CreateOrder(ctx context.Context) error {
	u.log.WithContext(ctx).Info("biz CreateOrder")
	return u.repo.CreateOrder(ctx)
}
