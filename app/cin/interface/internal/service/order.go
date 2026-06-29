package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"
)

func (s *CinInterface) PlaceOrder(ctx context.Context, req *v1.PlaceOrderRequest) (*v1.PlaceOrderReply, error) {
	// 获取上下文的用户id
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.Order{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.UserID = userID

	// 查找商品是否存在
	product, err := s.pc.GetProduct(ctx, in.ProductID)
	if err != nil {
		s.log.Errorf("product not found: %v", err)
		return nil, v1.ErrorOrderIdNotExist("product not found")
	}
	if product.Status != 1 {
		return nil, v1.ErrorOrderIdNotExist("product is off shelf")
	}

	// 检查库存是否充足
	stock, err := s.sc.GetStock(ctx, in.ProductID)
	if err != nil {
		s.log.Errorf("stock not found: %v", err)
		return nil, v1.ErrorOrderIdNotExist("stock not found")
	}
	if stock.Available < in.Count {
		return nil, v1.ErrorOrderIdNotExist("insufficient stock")
	}

	// 预留库存（乐观锁）
	_, err = s.sc.ReserveStock(ctx, in.ProductID, in.Count, stock.Version)
	if err != nil {
		s.log.Errorf("reserve stock failed: %v", err)
		return nil, v1.ErrorOrderIdNotExist("reserve stock failed, please retry")
	}

	// 创建订单
	err = s.oc.CreateOrder(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.PlaceOrderReply{
		OrderId: in.ID,
	}, nil
}

func (s *CinInterface) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	// 支付订单
	err := s.oc.PayOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{
		Success: true,
	}, nil
}
