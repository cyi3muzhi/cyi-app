package data

import (
	"context"

	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type dashboardRepo struct {
	log  *log.Helper
	data *Data
}

// NewDashboardRepo 创建仪表盘数据仓库，返回 mock 数据
func NewDashboardRepo(data *Data, logger log.Logger) biz.DashboardRepo {
	return &dashboardRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// GetDashboardStats 获取仪表盘统计数据（mock 数据）
func (r *dashboardRepo) GetDashboardStats(ctx context.Context) (*biz.DashboardStats, error) {
	return &biz.DashboardStats{
		TodayOrders:    128,
		TodayRevenue:   "45680",
		TotalUsers:     23568,
		TotalProducts:  1286,
		OrdersGrowth:   12.5,
		RevenueGrowth:  8.3,
		UsersGrowth:    5.2,
		ProductsGrowth: 3.1,
	}, nil
}

// GetOrderTrend 获取近14日订单趋势（mock 数据）
func (r *dashboardRepo) GetOrderTrend(ctx context.Context) ([]*biz.TrendItem, error) {
	return []*biz.TrendItem{
		{Date: "06-01", Value: 86},
		{Date: "06-02", Value: 92},
		{Date: "06-03", Value: 78},
		{Date: "06-04", Value: 105},
		{Date: "06-05", Value: 98},
		{Date: "06-06", Value: 112},
		{Date: "06-07", Value: 88},
		{Date: "06-08", Value: 95},
		{Date: "06-09", Value: 120},
		{Date: "06-10", Value: 108},
		{Date: "06-11", Value: 96},
		{Date: "06-12", Value: 115},
		{Date: "06-13", Value: 102},
		{Date: "06-14", Value: 128},
	}, nil
}

// GetRevenueCategories 获取营收分类占比（mock 数据）
func (r *dashboardRepo) GetRevenueCategories(ctx context.Context) ([]*biz.RevenueCategory, error) {
	return []*biz.RevenueCategory{
		{Name: "电子产品", Percentage: 35.5, Color: "#5470c6"},
		{Name: "服装鞋帽", Percentage: 22.3, Color: "#91cc75"},
		{Name: "食品饮料", Percentage: 18.2, Color: "#fac858"},
		{Name: "家居用品", Percentage: 14.8, Color: "#ee6666"},
		{Name: "其他", Percentage: 9.2, Color: "#73c0de"},
	}, nil
}
