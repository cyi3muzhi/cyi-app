package data

import (
	"context"

	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type dashboardRepo struct {
	log  *log.Helper
	data *Data
}

// NewDashboardRepo 创建仪表盘数据仓库
func NewDashboardRepo(data *Data, logger log.Logger) biz.DashboardRepo {
	return &dashboardRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// GetDashboardStats 获取仪表盘统计数据（实时数据）
func (r *dashboardRepo) GetDashboardStats(ctx context.Context) (*biz.DashboardStats, error) {
	// 并行查询各服务
	type statsResult struct {
		orderStats *orderRpc.OrderStatsReply
		userCount  int64
		prodCount  int32
	}

	result := &statsResult{}
	errCh := make(chan error, 3)

	// 查询订单统计
	go func() {
		stats, err := r.data.oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
		if err != nil {
			errCh <- err
			return
		}
		result.orderStats = stats
		errCh <- nil
	}()

	// 查询用户总数
	go func() {
		rep, err := r.data.uc.UserCount(ctx, &emptypb.Empty{})
		if err != nil {
			errCh <- err
			return
		}
		result.userCount = rep.Count
		errCh <- nil
	}()

	// 查询商品总数（取第一页拿到 total）
	go func() {
		rep, err := r.data.pc.ListProduct(ctx, &productRpc.ListProductRequest{Page: 1, PageSize: 1})
		if err != nil {
			errCh <- err
			return
		}
		result.prodCount = rep.Total
		errCh <- nil
	}()

	// 等待所有查询完成
	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil {
			r.log.Warnf("获取仪表盘数据部分失败: %v", err)
		}
	}

	stats := &biz.DashboardStats{}
	if result.orderStats != nil {
		stats.TodayOrders = int64(result.orderStats.TodayCount)
		stats.TotalProducts = int64(result.prodCount)
	}
	stats.TotalUsers = result.userCount
	if result.prodCount > 0 {
		stats.TotalProducts = int64(result.prodCount)
	}

	return stats, nil
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
