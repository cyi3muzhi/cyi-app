package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// ListOrder 订单列表（mock 数据）
func (s *AdminService) ListOrder(ctx context.Context, req *adminv1.ListOrderRequest) (*adminv1.ListOrderReply, error) {
	return &adminv1.ListOrderReply{
		Items: []*adminv1.OrderItem{
			{
				Id:         1,
				OrderNo:    "ORD20260614001",
				CreatedAt:  "2026-06-14 10:30:00",
				UserName:   "张三",
				Amount:     899900,
				Payment:    "微信支付",
				Status:     "paid",
				StatusText: "已支付",
				Products: []*adminv1.OrderProduct{
					{Name: "iPhone 15 Pro", Spec: "256GB 黑色", Qty: 1, Price: 899900},
				},
			},
			{
				Id:         2,
				OrderNo:    "ORD20260614002",
				CreatedAt:  "2026-06-14 11:00:00",
				UserName:   "李四",
				Amount:     189900,
				Payment:    "支付宝",
				Status:     "pending",
				StatusText: "待支付",
				Products: []*adminv1.OrderProduct{
					{Name: "AirPods Pro 2", Spec: "白色", Qty: 1, Price: 189900},
				},
			},
			{
				Id:         3,
				OrderNo:    "ORD20260613001",
				CreatedAt:  "2026-06-13 15:20:00",
				UserName:   "王五",
				Amount:     12900,
				Payment:    "微信支付",
				Status:     "shipped",
				StatusText: "已发货",
				Products: []*adminv1.OrderProduct{
					{Name: "纯棉T恤", Spec: "白色 M", Qty: 1, Price: 12900},
				},
			},
		},
		Total: 3,
	}, nil
}

// GetOrder 订单详情（mock 数据）
func (s *AdminService) GetOrder(ctx context.Context, req *adminv1.GetOrderRequest) (*adminv1.GetOrderReply, error) {
	return &adminv1.GetOrderReply{
		Order: &adminv1.OrderItem{
			Id:         req.OrderId,
			OrderNo:    "ORD20260614001",
			CreatedAt:  "2026-06-14 10:30:00",
			UserName:   "张三",
			Amount:     899900,
			Payment:    "微信支付",
			Status:     "paid",
			StatusText: "已支付",
			Products: []*adminv1.OrderProduct{
				{Name: "iPhone 15 Pro", Spec: "256GB 黑色", Qty: 1, Price: 899900},
			},
		},
	}, nil
}

// ShipOrder 发货（mock 实现）
func (s *AdminService) ShipOrder(ctx context.Context, req *adminv1.ShipOrderRequest) (*adminv1.ShipOrderReply, error) {
	return &adminv1.ShipOrderReply{
		Success: true,
	}, nil
}

// CancelOrder 取消订单（mock 实现）
func (s *AdminService) CancelOrder(ctx context.Context, req *adminv1.CancelOrderRequest) (*adminv1.CancelOrderReply, error) {
	return &adminv1.CancelOrderReply{
		Success: true,
	}, nil
}

// ExportOrder 导出订单（mock 数据）
func (s *AdminService) ExportOrder(ctx context.Context, req *adminv1.ExportOrderRequest) (*adminv1.ExportOrderReply, error) {
	return &adminv1.ExportOrderReply{
		DownloadUrl: "https://example.com/export/orders_20260614.xlsx",
	}, nil
}
