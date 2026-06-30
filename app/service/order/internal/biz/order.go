package biz

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// OrderItem 订单行项目（聚合内的实体）
type OrderItem struct {
	ProductID   int64
	ProductName string
	Price       int64
	Quantity    int32
}

// SubTotal 小计
func (i *OrderItem) SubTotal() int64 {
	return i.Price * int64(i.Quantity)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
	GetOrder(ctx context.Context, id int64) (*Order, error)
}

// Order 订单聚合根
type Order struct {
	ID         int64
	UserID     int64
	OrderNo    string
	TotalPrice int64
	PayAmount  int64
	Status     v1.OrderStatus

	Items         []*OrderItem // 聚合内的实体
	statusHandler OrderStatus  // 状态模式
}

// AddItem 添加商品（保持一致性）
func (o *Order) AddItem(item *OrderItem) {
	o.Items = append(o.Items, item)
	o.TotalPrice += item.SubTotal()
}

// VerifyOwner 校验订单归属（领域行为）
func (o *Order) VerifyOwner(userID int64) error {
	if o.UserID != userID {
		return v1.ErrorUserNotOrder("user not order")
	}
	return nil
}

// GenerateOrderNo 生成订单号
func (o *Order) GenerateOrderNo() {
	now := time.Now()
	b := make([]byte, 4)
	rand.Read(b)
	o.OrderNo = fmt.Sprintf("%s%d%08x", now.Format("20060102150405"), o.UserID, b)
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
