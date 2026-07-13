package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// GetDashboardStats 获取仪表盘统计数据
func (s *AdminService) GetDashboardStats(ctx context.Context, req *adminv1.GetDashboardStatsRequest) (*adminv1.GetDashboardStatsReply, error) {
	stats, err := s.dc.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetDashboardStatsReply{
		TodayOrders:    stats.TodayOrders,
		TodayRevenue:   stats.TodayRevenue,
		TotalUsers:     stats.TotalUsers,
		TotalProducts:  stats.TotalProducts,
		OrdersGrowth:   stats.OrdersGrowth,
		RevenueGrowth:  stats.RevenueGrowth,
		UsersGrowth:    stats.UsersGrowth,
		ProductsGrowth: stats.ProductsGrowth,
	}, nil
}

// GetOrderTrend 获取近14日订单趋势
func (s *AdminService) GetOrderTrend(ctx context.Context, req *adminv1.GetOrderTrendRequest) (*adminv1.GetOrderTrendReply, error) {
	items, err := s.dc.GetOrderTrend(ctx)
	if err != nil {
		return nil, err
	}
	replyItems := make([]*adminv1.TrendItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, &adminv1.TrendItem{
			Date:  item.Date,
			Value: item.Value,
		})
	}
	return &adminv1.GetOrderTrendReply{
		Items: replyItems,
	}, nil
}

// GetRevenueCategories 获取营收分类占比
func (s *AdminService) GetRevenueCategories(ctx context.Context, req *adminv1.GetRevenueCategoriesRequest) (*adminv1.GetRevenueCategoriesReply, error) {
	categories, err := s.dc.GetRevenueCategories(ctx)
	if err != nil {
		return nil, err
	}
	replyCategories := make([]*adminv1.RevenueCategory, 0, len(categories))
	for _, c := range categories {
		replyCategories = append(replyCategories, &adminv1.RevenueCategory{
			Name:       c.Name,
			Percentage: c.Percentage,
			Color:      c.Color,
		})
	}
	return &adminv1.GetRevenueCategoriesReply{
		Categories: replyCategories,
	}, nil
}
