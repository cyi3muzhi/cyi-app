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
		OrderId:        order.ID,
		UserId:         order.UserID,
		OrderNo:        order.OrderNo,
		Status:         int32(order.Status),
		TotalPrice:     order.TotalPrice.Amount,
		PayAmount:      order.PayAmount.Amount,
		Freight:        order.Freight.Amount,
		DiscountAmount: order.Discount.Amount,
		CreatedAt:      order.CreateTime,
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
	items, total, err := s.uc.List(ctx, req.Page, req.PageSize, req.Status, req.Keyword, req.DateFrom, req.DateTo)
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
	return &v1.OrderStatsReply{
		TotalCount:   total,
		PendingCount: pending,
		TodayCount:   today,
		RefundCount:  refund,
	}, nil
}

// payAmount / freight / discount 在 service 层从 biz.Money 获取 Amount

func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderReply, error) {
	if err := s.uc.Pay(ctx, req.OrderId, req.UserId); err != nil {
		return nil, err
	}
	return &v1.PayOrderReply{}, nil
}
