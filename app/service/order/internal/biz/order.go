package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"fmt"
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

// GetStatus 获取订单状态
func (o *Order) GetStatus() v1.OrderStatus {
	return o.Status
}

// SetStatus 设置状态并更新状态处理器
func (o *Order) SetStatus(newStatus v1.OrderStatus, repo OrderRepo) {
	o.Status = newStatus
	o.statusHandler = NewOrderStatusHandler(o, repo)
}

// 支付
func (o *Order) Pay(ctx context.Context, in *Order) error {
	return o.statusHandler.Pay(ctx, in)
}

// 发货
func (o *Order) Ship(ctx context.Context) error {
	return o.statusHandler.Ship(ctx)
}

// 收货
func (o *Order) Receive(ctx context.Context) error {
	return o.statusHandler.Receive(ctx)
}

// 完成
func (o *Order) Complete(ctx context.Context) error {
	return o.statusHandler.Complete(ctx)
}

// 取消
func (o *Order) Cancel(ctx context.Context) error {
	return o.statusHandler.Cancel(ctx)
}

// 退款中
func (o *Order) Refunding(ctx context.Context) error {
	return o.statusHandler.Refunding(ctx)
}

// 已退款
func (o *Order) Refunded(ctx context.Context) error {
	return o.statusHandler.Refunded(ctx)
}

type OrderUseCase struct {
	repo OrderRepo
}

func NewOrderUseCase(repo OrderRepo) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, in *Order) error {
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo)
	fmt.Println(in)
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
