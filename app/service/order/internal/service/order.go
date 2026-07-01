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
	in.AddItem(&biz.OrderItem{
		ProductID:   req.ProductId,
		Price:       req.Price,
		Quantity:    req.Count,
		Currency:    "CNY",
	})
	in.GenerateOrderNo()
	if err := s.uc.Create(ctx, in); err != nil {
		return nil, err
	}
	return &v1.CreateOrderReply{OrderId: in.ID}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *v1.UpdateOrderRequest) (*v1.UpdateOrderReply, error) {
	in := &biz.Order{}
	convert.Convert(req, in)
	if err := s.uc.Update(ctx, in); err != nil {
		return nil, err
	}
	return &v1.UpdateOrderReply{}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *v1.DeleteOrderRequest) (*v1.DeleteOrderReply, error) {
	return &v1.DeleteOrderReply{}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.GetOrderReply, error) {
	order, err := s.uc.Get(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetOrderReply{
		OrderId:    order.ID,
		UserId:     order.UserID,
		OrderNo:    order.OrderNo,
		Status:     int32(order.Status),
		TotalPrice: order.TotalPrice.Amount,
		PayAmount:  order.PayAmount.Amount,
	}
	for _, item := range order.Items {
		reply.Items = append(reply.Items, &v1.OrderItemProto{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return reply, nil
}

func (s *OrderService) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	return &v1.ListOrderReply{}, nil
}

func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	if err := s.uc.Pay(ctx, req.OrderId, req.UserId); err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{}, nil
}
