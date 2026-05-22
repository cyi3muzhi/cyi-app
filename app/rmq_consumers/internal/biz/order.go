package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderUseCase struct {
	repo OrderRepo
	log  *log.Helper
}

type OrderRepo interface {
	CreateOrder(ctx context.Context) error
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (u *OrderUseCase) CreateOrder(ctx context.Context) error {
	fmt.Println("biz CreateOrder")
	return u.repo.CreateOrder(ctx)
}
