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

	// 1. 查找商品是否存在且上架
	product, err := s.pc.GetProduct(ctx, in.ProductID)
	if err != nil {
		s.log.Errorf("product not found: %v", err)
		return nil, v1.ErrorProductNotAvailable("product not found")
	}
	if product.Status != 1 {
		return nil, v1.ErrorProductNotAvailable("product is off shelf")
	}

	// 2. 检查库存是否充足
	stock, err := s.sc.GetStock(ctx, in.ProductID)
	if err != nil {
		s.log.Errorf("stock not found: %v", err)
		return nil, v1.ErrorProductNotAvailable("stock not found")
	}
	if stock.Available < in.Count {
		return nil, v1.ErrorInsufficientStock("insufficient stock")
	}

	// 3. 预留库存（乐观锁）
	_, err = s.sc.ReserveStock(ctx, in.ProductID, in.Count, stock.Version)
	if err != nil {
		s.log.Errorf("reserve stock failed: %v", err)
		return nil, v1.ErrorStockOperationFailed("reserve stock failed, please retry")
	}

	// 4. 创建订单
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

	// 1. 查询订单
	order, err := s.oc.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.log.Errorf("order not found: %v", err)
		return nil, v1.ErrorOrderIdNotExist("order not found")
	}

	// 2. 支付订单（状态流转）—— 先做，失败则库存不动
	err = s.oc.PayOrder(ctx, req.OrderId, userID)
	if err != nil {
		s.log.Errorf("pay order failed: %v", err)
		return nil, v1.ErrorOrderAlreadyPaid("pay order failed, may already be paid")
	}

	// 3. 查询库存
	stock, err := s.sc.GetStock(ctx, order.ProductID)
	if err != nil {
		s.log.Errorf("stock not found: %v", err)
		return nil, v1.ErrorProductNotAvailable("stock not found")
	}

	// 4. 实际扣减库存（Total-、Reserved-、乐观锁）
	_, err = s.sc.DeductStock(ctx, order.ProductID, order.Count, stock.Version)
	if err != nil {
		s.log.Errorf("deduct stock failed, order already paid, need manual fix: %v", err)
		return nil, v1.ErrorStockOperationFailed("deduct stock failed, already paid")
	}

	return &v1.PayOrderReply{
		Success: true,
	}, nil
}
