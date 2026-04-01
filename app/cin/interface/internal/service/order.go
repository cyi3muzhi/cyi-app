package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderReply, error) {
	// 获取上下文的用户id
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	// 创建订单
	in := &biz.Order{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.UserID = userID
	// 查找商品是否存在
	// 库存是否充足
	// 扣减库存
	// 创建订单
	err = s.oc.CreateOrder(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.CreateOrderReply{
		OrderId: in.ID,
	}, nil
}

func (s *CinInterface) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	// 获取上下文的用户id
	// userID, err := service.GetUserID(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// 查找订单是否存在
	// order, err := s.oc.GetOrder(ctx, req.OrderID)
	// if err != nil {
	// 	return nil, err
	// }
	// 支付订单
	err := s.oc.PayOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{
		Success: true,
	}, nil
}
