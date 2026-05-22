package service

import (
	"context"

	"cyi-app/app/service/order/internal/biz"
	"cyi-app/pkg/util/convert"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderService struct {
	v1.UnimplementedOrderServer
	uc  *biz.OrderUseCase
	log log.Logger
}

func NewOrderService(uc *biz.OrderUseCase, logger log.Logger) *OrderService {
	return &OrderService{
		uc:  uc,
		log: logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
	in := &biz.Order{}
	convert.Convert(req, in)
	err := s.uc.CreateOrder(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateOrderReply{
		OrderId: in.ID,
	}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *v1.UpdateOrderRequest) (*v1.UpdateOrderReply, error) {
	in := &biz.Order{}
	convert.Convert(req, in)
	err := s.uc.UpdateOrder(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateOrderReply{}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *v1.DeleteOrderRequest) (*v1.DeleteOrderReply, error) {
	return &v1.DeleteOrderReply{}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.GetOrderReply, error) {
	return &v1.GetOrderReply{}, nil
}

func (s *OrderService) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	return &v1.ListOrderReply{}, nil
}

func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	in := &biz.Order{
		ID: req.OrderId,
	}
	// 校验订单
	err := s.uc.VerifyOrder(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	err = s.uc.Pay(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{}, nil
}
