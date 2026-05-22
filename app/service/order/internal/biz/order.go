package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"cyi-app/pkg/kratos/service"

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
	repo   OrderRepo
	log    *log.Helper
	logger log.Logger
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, log: log.NewHelper(logger), logger: logger}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, in *Order) error {
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo, uc.logger)
	return uc.repo.CreateOrder(ctx, in)
}

func (uc *OrderUseCase) UpdateOrder(ctx context.Context, in *Order) error {
	if in.statusHandler == nil {
		in.SetStatus(in.Status, uc.repo, uc.logger)
	}
	return uc.repo.UpdateOrder(ctx, in)
}

// GetOrderWithStatus 查询订单并初始化状态处理器
func (uc *OrderUseCase) GetOrderWithStatus(ctx context.Context, orderID int64) (*Order, error) {
	order, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.SetStatus(order.Status, uc.repo, uc.logger)
	return order, nil
}

// GetOrderWithVerify 查询订单并校验权限
func (uc *OrderUseCase) VerifyOrder(ctx context.Context, orderID int64) error {
	order, err := uc.GetOrderWithStatus(ctx, orderID)
	if err != nil {
		return err
	}

	userID, err := service.GetUserID(ctx)
	if err != nil {
		return err
	}
	if order.UserID != userID {
		return v1.ErrorUserNotOrder("user not order")
	}

	return nil
}

// Pay 支付订单
func (uc *OrderUseCase) Pay(ctx context.Context, in *Order) error {
	order, err := uc.GetOrderWithStatus(ctx, in.ID)
	if err != nil {
		return err
	}
	return order.Pay(ctx, in)
}
