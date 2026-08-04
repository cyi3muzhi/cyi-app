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
	in.UserID = req.UserId
	in.CouponRecordID = req.CouponRecordId                       // 优惠券发放记录ID，0=未使用优惠券
	in.Freight = biz.Money{Amount: req.Freight, Currency: "CNY"} // 运费（单位：分）
	in.AddressID = req.AddressId                                 // 收货地址快照
	in.Receiver = req.Receiver
	in.Phone = req.Phone
	in.Address = req.Address
	in.Items = nil // 清空 convert.Convert 错误填充的 Items（类型不匹配导致重复）

	// 支持两种模式：立即购买（单商品）和购物车结算（多商品）
	if len(req.Items) > 0 {
		for _, item := range req.Items {
			in.AddItem(&biz.OrderItem{
				ProductID: item.ProductId,
				SkuID:     item.SkuId,
				Quantity:  item.Quantity,
				Currency:  "CNY",
			})
		}
	} else {
		in.AddItem(&biz.OrderItem{
			ProductID: req.ProductId,
			SkuID:     req.SkuId,
			Price:     req.Price,
			Quantity:  req.Count,
			Currency:  "CNY",
		})
	}
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
	return orderToReply(order), nil
}

// BatchGetOrder 批量获取订单详情（含商品明细）
// task 服务导出订单时调用，1 次请求替代 N 次 GetOrder，消除 N+1 查询
func (s *OrderService) BatchGetOrder(ctx context.Context, req *v1.BatchGetOrderRequest) (*v1.BatchGetOrderReply, error) {
	orders, err := s.uc.BatchGet(ctx, req.OrderIds)
	if err != nil {
		return nil, err
	}
	reply := &v1.BatchGetOrderReply{Orders: make([]*v1.GetOrderReply, 0, len(orders))}
	for _, order := range orders {
		reply.Orders = append(reply.Orders, orderToReply(order))
	}
	return reply, nil
}

// orderToReply 将 biz.Order 转换为 proto GetOrderReply（GetOrder/BatchGetOrder 共用）
func orderToReply(order *biz.Order) *v1.GetOrderReply {
	reply := &v1.GetOrderReply{
		OrderId:          order.ID,
		UserId:           order.UserID,
		OrderNo:          order.OrderNo,
		Status:           int32(order.Status),
		TotalPrice:       order.TotalPrice.Amount,
		PayAmount:        order.PayAmount.Amount,
		Freight:          order.Freight.Amount,
		DiscountAmount:   order.Discount.Amount,
		CreatedAt:        order.CreateTime,
		LogisticsCompany: order.LogisticsCompany,
		LogisticsNo:      order.LogisticsNo,
		Receiver:         order.Receiver,
		Phone:            order.Phone,
		Address:          order.Address,
	}
	for _, item := range order.Items {
		reply.Items = append(reply.Items, &v1.OrderItemProto{
			ProductId:   item.ProductID,
			SkuId:       item.SkuID,
			SpecDesc:    item.SpecDesc,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return reply
}

func (s *OrderService) ListOrder(ctx context.Context, req *v1.ListOrderRequest) (*v1.ListOrderReply, error) {
	items, total, err := s.uc.List(ctx, req.Page, req.PageSize, req.Status, req.Keyword, req.DateFrom, req.DateTo, req.UserId)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.ListOrderItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &v1.ListOrderItem{
			OrderId:        item.ID,
			OrderNo:        item.OrderNo,
			UserId:         item.UserID,
			Status:         int32(item.Status),
			TotalPrice:     item.TotalPrice.Amount,
			PayAmount:      item.PayAmount.Amount,
			Freight:        item.Freight.Amount,
			DiscountAmount: item.Discount.Amount,
			CreatedAt:      item.CreateTime,
		})
	}
	return &v1.ListOrderReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// OrderStats 获取订单统计数据
func (s *OrderService) OrderStats(ctx context.Context, req *v1.OrderStatsRequest) (*v1.OrderStatsReply, error) {
	total, pending, today, refund, err := s.uc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	// 营收统计（今日/昨日已完成订单实付金额）
	todayRevenue, yesterdayRevenue, err := s.uc.RevenueStats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.OrderStatsReply{
		TotalCount:       total,
		PendingCount:     pending,
		TodayCount:       today,
		RefundCount:      refund,
		TodayRevenue:     todayRevenue,
		YesterdayRevenue: yesterdayRevenue,
	}, nil
}

// GetOrderTrend 获取近 N 日订单趋势（管理端仪表盘）
func (s *OrderService) GetOrderTrend(ctx context.Context, req *v1.GetOrderTrendRequest) (*v1.GetOrderTrendReply, error) {
	items, err := s.uc.OrderTrend(ctx, req.GetDays())
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.TrendItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &v1.TrendItem{
			Date:       item.Date,
			OrderCount: item.OrderCount,
		})
	}
	return &v1.GetOrderTrendReply{
		Items: replyItems,
	}, nil
}

// GetUserOrderStats 获取单个用户的订单统计（个人中心使用）
func (s *OrderService) GetUserOrderStats(ctx context.Context, req *v1.GetUserOrderStatsRequest) (*v1.GetUserOrderStatsReply, error) {
	stats, err := s.uc.UserStats(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.GetUserOrderStatsReply{
		TotalOrders:    stats.TotalOrders,
		PendingPay:     stats.PendingPay,
		PendingShip:    stats.PendingShip,
		PendingReceive: stats.PendingReceive,
		Complete:       stats.Complete,
		RefundCount:    stats.RefundCount,
		TotalSpent:     stats.TotalSpent,
	}, nil
}

// payAmount / freight / discount 在 service 层从 biz.Money 获取 Amount

func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	if err := s.uc.Pay(ctx, req.OrderId, req.UserId); err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{}, nil
}

// ShipOrder 发货（管理端调用）：待发货 -> 已发货，并写入物流信息
func (s *OrderService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderReply, error) {
	if err := s.uc.Ship(ctx, req.OrderId, req.LogisticsCompany, req.LogisticsNo); err != nil {
		return nil, err
	}
	return &v1.ShipOrderReply{}, nil
}

// ReceiveOrder 确认收货（用户端调用）：已发货/待收货 -> 已完成
func (s *OrderService) ReceiveOrder(ctx context.Context, req *v1.ReceiveOrderRequest) (*v1.ReceiveOrderReply, error) {
	if err := s.uc.Receive(ctx, req.OrderId, req.UserId); err != nil {
		return nil, err
	}
	return &v1.ReceiveOrderReply{}, nil
}

// CancelOrder 取消订单（用户端/管理端调用）：待支付/已支付 -> 已取消
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.CancelOrderReply, error) {
	if err := s.uc.Cancel(ctx, req.OrderId, req.UserId, req.Reason); err != nil {
		return nil, err
	}
	return &v1.CancelOrderReply{}, nil
}

// ListOverdueWaitPayOrders 查询超时未支付的待付款订单（task 服务 cron 兜底扫描用）
func (s *OrderService) ListOverdueWaitPayOrders(ctx context.Context, req *v1.ListOverdueWaitPayOrdersRequest) (*v1.ListOverdueWaitPayOrdersReply, error) {
	orders, total, err := s.uc.OverdueWaitPayOrders(ctx, req.Page, req.PageSize, req.TimeoutSeconds)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*v1.ListOrderItem, 0, len(orders))
	for _, item := range orders {
		replyItems = append(replyItems, &v1.ListOrderItem{
			OrderId:   item.ID,
			OrderNo:   item.OrderNo,
			UserId:    item.UserID,
			Status:    int32(item.Status),
			CreatedAt: item.CreateTime,
		})
	}
	return &v1.ListOverdueWaitPayOrdersReply{
		Items: replyItems,
		Total: total,
	}, nil
}
