package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// ListOrder 获取订单列表
func (s *AdminService) ListOrder(ctx context.Context, req *adminv1.ListOrderRequest) (*adminv1.ListOrderReply, error) {
	items, total, err := s.oc.List(ctx, req.Page, req.PageSize, req.Status, req.Keyword, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.OrderItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizToAdminProtoOrder(item))
	}
	return &adminv1.ListOrderReply{
		Items: replyItems,
		Total: total,
	}, nil
}

// GetOrder 获取订单详情
func (s *AdminService) GetOrder(ctx context.Context, req *adminv1.GetOrderRequest) (*adminv1.GetOrderReply, error) {
	order, err := s.oc.Get(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetOrderReply{
		Order: bizToAdminProtoOrder(order),
	}, nil
}

// ShipOrder 发货
func (s *AdminService) ShipOrder(ctx context.Context, req *adminv1.ShipOrderRequest) (*adminv1.ShipOrderReply, error) {
	err := s.oc.Ship(ctx, req.OrderId, req.LogisticsCompany, req.LogisticsNo)
	if err != nil {
		return nil, err
	}
	return &adminv1.ShipOrderReply{
		Success: true,
	}, nil
}

// CancelOrder 取消订单
func (s *AdminService) CancelOrder(ctx context.Context, req *adminv1.CancelOrderRequest) (*adminv1.CancelOrderReply, error) {
	err := s.oc.Cancel(ctx, req.OrderId, req.Reason)
	if err != nil {
		return nil, err
	}
	return &adminv1.CancelOrderReply{
		Success: true,
	}, nil
}

// OrderStats 获取订单统计数据（顶部统计卡片）
func (s *AdminService) OrderStats(ctx context.Context, req *adminv1.OrderStatsRequest) (*adminv1.OrderStatsReply, error) {
	total, pending, today, refund, err := s.oc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &adminv1.OrderStatsReply{
		TotalCount:   total,
		PendingCount: pending,
		TodayCount:   today,
		RefundCount:  refund,
	}, nil
}

// bizToAdminProtoOrder 将 biz Order 转换为 admin proto OrderItem
func bizToAdminProtoOrder(order *biz.Order) *adminv1.OrderItem {
	products := make([]*adminv1.OrderProduct, 0, len(order.Products))
	for _, p := range order.Products {
		products = append(products, &adminv1.OrderProduct{
			Name:  p.Name,
			Spec:  p.Spec,
			Qty:   p.Qty,
			Price: p.Price,
			Image: p.Image,
		})
	}
	return &adminv1.OrderItem{
		Id:         order.ID,
		OrderNo:    order.OrderNo,
		CreatedAt:  order.CreatedAt,
		UserName:   order.UserName,
		Amount:     order.Amount,
		Status:     order.Status,
		StatusText: order.StatusText,
		Products:   products,
	}
}
