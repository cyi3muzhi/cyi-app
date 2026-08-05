package service

import (
	"context"
	"time"

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

func (s *CinInterface) CartCheckout(ctx context.Context, req *v1.CartCheckoutRequest) (*v1.CartCheckoutReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.Order{
		UserID:         userID,
		AddressID:      req.AddressId,
		Receiver:       req.Receiver,
		Phone:          req.Phone,
		Address:        req.Address,
		Freight:        req.Freight,
		CouponRecordID: req.CouponId, // 优惠券发放记录ID（0=不使用）
	}
	for _, item := range req.Items {
		in.Items = append(in.Items, &biz.CartItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			SkuID:     item.SkuId,
		})
	}
	// 订单服务内部完成：商品校验、库存校验、预留库存、创建订单（含地址快照/运费）
	err = s.oc.CreateOrder(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CartCheckoutReply{
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

// ReceiveOrder 确认收货（当前登录用户确认自己的订单收货）
func (s *CinInterface) ReceiveOrder(ctx context.Context, req *v1.ReceiveOrderRequest) (*v1.ReceiveOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.oc.ReceiveOrder(ctx, req.OrderId, userID)
	if err != nil {
		return nil, err
	}

	return &v1.ReceiveOrderReply{
		Success: true,
	}, nil
}

// CancelOrder 取消订单（当前登录用户取消自己的订单）
func (s *CinInterface) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.CancelOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = s.oc.CancelOrder(ctx, req.OrderId, userID, req.Reason)
	if err != nil {
		return nil, err
	}

	return &v1.CancelOrderReply{
		Success: true,
	}, nil
}

// 订单列表（当前登录用户的订单，含商品明细）
func (s *CinInterface) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.oc.ListOrder(ctx, userID, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.OrderItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, orderBizToProto(item))
	}
	return &v1.ListOrderReply{Items: replyItems, Total: total}, nil
}

// 订单详情（当前登录用户的订单）
func (s *CinInterface) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.GetOrderReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := s.oc.GetOrderDetail(ctx, req.OrderId, userID)
	if err != nil {
		return nil, err
	}
	return &v1.GetOrderReply{Order: orderBizToProto(order)}, nil
}

// ApplyRefund 申请退款（当前登录用户对自己的订单发起，仅已支付/待发货可申请）
func (s *CinInterface) ApplyRefund(ctx context.Context, req *v1.ApplyRefundRequest) (*v1.ApplyRefundReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	refundID, err := s.oc.ApplyRefund(ctx, req.OrderId, userID, req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.ApplyRefundReply{RefundId: refundID}, nil
}

// ListRefund 我的退款记录（当前登录用户）
func (s *CinInterface) ListRefund(ctx context.Context, req *v1.ListRefundRequest) (*v1.ListRefundReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.oc.ListRefund(ctx, userID, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.RefundItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &v1.RefundItem{
			RefundId:     item.ID,
			OrderId:      item.OrderID,
			OrderNo:      item.OrderNo,
			RefundAmount: item.RefundAmount,
			Reason:       item.Reason,
			Status:       item.Status,
			StatusText:   item.StatusText,
			RejectReason: item.RejectReason,
			CreatedAt:    time.Unix(item.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}
	return &v1.ListRefundReply{Items: replyItems, Total: total}, nil
}

// orderBizToProto biz.Order -> v1.OrderItem 转换（含状态字符串映射、时间格式化）
func orderBizToProto(o *biz.Order) *v1.OrderItem {
	// 支付截止时间：待付款订单 = 下单时间 + 30 分钟，其余状态为 0
	payDeadline := int64(0)
	if o.Status == 1 {
		payDeadline = o.CreatedAt + 30*60
	}
	item := &v1.OrderItem{
		Id:          o.ID,
		OrderNo:     o.OrderNo,
		CreatedAt:   time.Unix(o.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		Amount:      o.PayAmount,
		Status:      orderStatusStr(o.Status),
		StatusText:  orderStatusText(o.Status),
		PayDeadline: payDeadline,
		Products:    []*v1.OrderProduct{},
		// 物流信息（发货后展示）
		LogisticsCompany: o.LogisticsCompany,
		LogisticsNo:      o.LogisticsNo,
		Receiver:         o.Receiver,
		Phone:            o.Phone,
		Address:          o.Address,
	}
	for _, p := range o.Products {
		item.Products = append(item.Products, &v1.OrderProduct{
			Name:      p.Name,
			Spec:      p.Spec,
			Qty:       p.Quantity,
			Price:     p.Price,
			Image:     p.Image,
			ProductId: p.ProductID, // 再次购买用
			SkuId:     p.SkuID,
		})
	}
	return item
}

// orderStatusStr 订单状态数字 -> 前端状态字符串（与前端 OrderStatus 类型对齐）
// 状态 2（已支付）/3（待发货）归入"待发货"；状态 4（已发货）/5（待收货）归入"待收货"
// 状态 8（退款中）/9（已退款）分别映射，便于列表区分展示
func orderStatusStr(status int32) string {
	switch status {
	case 1:
		return "pending_payment"
	case 2, 3:
		return "pending_ship"
	case 4, 5:
		return "pending_receive"
	case 6:
		return "completed"
	case 7:
		return "cancelled"
	case 8:
		return "refunding"
	case 9:
		return "refunded"
	default:
		return "unknown"
	}
}

// orderStatusText 订单状态数字 -> 中文状态文本
func orderStatusText(status int32) string {
	switch status {
	case 1:
		return "待付款"
	case 2:
		return "已支付"
	case 3:
		return "待发货"
	case 4:
		return "已发货"
	case 5:
		return "待收货"
	case 6:
		return "已完成"
	case 7:
		return "已取消"
	case 8:
		return "退款中"
	case 9:
		return "已退款"
	default:
		return "未知状态"
	}
}
