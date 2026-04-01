package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type Order struct {
	ID        int64 `json:"id"`
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Count     int32 `json:"count"`
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order) error
	PayOrder(ctx context.Context, orderID int64) error
}

type OrderUseCase struct {
	log  *log.Helper
	repo OrderRepo
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (u *OrderUseCase) CreateOrder(ctx context.Context, order *Order) error {
	return u.repo.CreateOrder(ctx, order)
}

func (u *OrderUseCase) PayOrder(ctx context.Context, orderID int64) error {
	// 支付订单
	if err := u.repo.PayOrder(ctx, orderID); err != nil {
		return err
	}
	return nil
}