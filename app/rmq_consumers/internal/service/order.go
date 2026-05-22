package service

import (
	"context"
	"cyi-app/app/rmq_consumers/internal/biz"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderService struct {
	uc  *biz.OrderUseCase
	log *log.Helper
}

func NewOrderService(uc *biz.OrderUseCase, logger log.Logger) *OrderService {
	return &OrderService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, msg []byte) error {
	fmt.Println("service CreateOrder, msg:", string(msg))
	return s.uc.CreateOrder(ctx)
}

func (s *OrderService) CreateOrderLog(ctx context.Context, msg []byte) error {
	fmt.Println("service CreateOrderLog, msg:", string(msg))
	return nil
}
