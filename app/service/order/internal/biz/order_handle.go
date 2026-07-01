package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
	GetOrder(ctx context.Context, id int64) (*Order, error)
}

type OrderUseCase struct {
	repo   OrderRepo
	log    *log.Helper
	logger log.Logger
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, log: log.NewHelper(logger), logger: logger}
}

// Create 创建订单
func (uc *OrderUseCase) Create(ctx context.Context, in *Order) error {
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo, uc.logger)
	return uc.repo.CreateOrder(ctx, in)
}

// Update 更新订单
func (uc *OrderUseCase) Update(ctx context.Context, in *Order) error {
	if in.statusHandler == nil {
		in.SetStatus(in.Status, uc.repo, uc.logger)
	}
	return uc.repo.UpdateOrder(ctx, in)
}

// Get 查询订单
func (uc *OrderUseCase) Get(ctx context.Context, orderID int64) (*Order, error) {
	order, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.SetStatus(order.Status, uc.repo, uc.logger)
	return order, nil
}

// Pay 支付订单
func (uc *OrderUseCase) Pay(ctx context.Context, orderID int64, userID int64) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.VerifyOwner(userID); err != nil {
		return err
	}
	return order.Pay(ctx)
}
