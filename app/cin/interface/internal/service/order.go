package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) PlaceOrder(ctx context.Context, req *v1.PlaceOrderRequest) (*v1.PlaceOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.Order{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.UserID = userID

	// 订单服务内部完成：商品校验、库存校验、预留库存、创建订单
	err = s.oc.CreateOrder(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.PlaceOrderReply{
		OrderId: in.ID,
	}, nil
}

func (s *CinInterface) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 订单服务内部完成：支付状态流转、扣减库存
	err = s.oc.PayOrder(ctx, req.OrderId, userID)
	if err != nil {
		return nil, err
	}

	return &v1.PayOrderReply{
		Success: true,
	}, nil
}
