package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	TodayOrders    int64
	TodayRevenue   string
	TotalUsers     int64
	TotalProducts  int64
	OrdersGrowth   float64
	RevenueGrowth  float64
	UsersGrowth    float64
	ProductsGrowth float64
}

// TrendItem 趋势项
type TrendItem struct {
	Date  string
	Value int32
}

// RevenueCategory 营收分类占比
type RevenueCategory struct {
	Name       string
	Percentage float64
	Color      string
}

// DashboardRepo 仪表盘数据仓库接口
type DashboardRepo interface {
	// GetDashboardStats 获取仪表盘统计数据
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	// GetOrderTrend 获取近14日订单趋势
	GetOrderTrend(ctx context.Context) ([]*TrendItem, error)
	// GetRevenueCategories 获取营收分类占比
	GetRevenueCategories(ctx context.Context) ([]*RevenueCategory, error)
}

// DashboardUseCase 仪表盘业务用例
type DashboardUseCase struct {
	log  *log.Helper
	repo DashboardRepo
}

// NewDashboardUseCase 创建仪表盘业务用例
func NewDashboardUseCase(repo DashboardRepo, logger log.Logger) *DashboardUseCase {
	return &DashboardUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (uc *DashboardUseCase) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	return uc.repo.GetDashboardStats(ctx)
}

// GetOrderTrend 获取近14日订单趋势
func (uc *DashboardUseCase) GetOrderTrend(ctx context.Context) ([]*TrendItem, error) {
	return uc.repo.GetOrderTrend(ctx)
}

// GetRevenueCategories 获取营收分类占比
func (uc *DashboardUseCase) GetRevenueCategories(ctx context.Context) ([]*RevenueCategory, error) {
	return uc.repo.GetRevenueCategories(ctx)
}
