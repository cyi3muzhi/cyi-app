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

type Order struct {
	ID int64
	// 用户ID
	UserID int64
	// 商品ID
	ProductID int64
	// 数量
	Count int32
	// 状态
	Status v1.OrderStatus
	// 状态处理器
	statusHandler OrderStatus
}

type OrderUseCase struct {
	repo OrderRepo
	log  *log.Helper
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, in *Order) error {
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo)
	return uc.repo.CreateOrder(ctx, in)
}

func (uc *OrderUseCase) UpdateOrder(ctx context.Context, in *Order) error {
	// 确保订单有状态处理器
	if in.statusHandler == nil {
		in.SetStatus(in.Status, uc.repo)
	}
	return uc.repo.UpdateOrder(ctx, in)
}

// Pay 支付订单
func (uc *OrderUseCase) Pay(ctx context.Context, in *Order) error {
	// 查询订单
	order, err := uc.repo.GetOrder(ctx, in.ID)
	if err != nil {
		return err
	}
	in.Status = order.Status
	in.SetStatus(in.Status, uc.repo)
	return in.Pay(ctx, in)
}
