package biz

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CartItem 购物车商品项
type CartItem struct {
	ProductID int64
	Quantity  int32
}

type Order struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Count     int32  `json:"count"`
	Price     int64  `json:"price"`
	Status    int32  `json:"status"`
	Items     []*CartItem // 购物车结算：多个商品
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order) error
	PayOrder(ctx context.Context, orderID int64, userID int64) error
	GetOrder(ctx context.Context, orderID int64) (*Order, error)
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

func (u *OrderUseCase) PayOrder(ctx context.Context, orderID int64, userID int64) error {
	if orderID == 0 {
		return v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.PayOrder(ctx, orderID, userID)
}

func (u *OrderUseCase) GetOrder(ctx context.Context, orderID int64) (order *Order, err error) {
	return u.repo.GetOrder(ctx, orderID)
}
