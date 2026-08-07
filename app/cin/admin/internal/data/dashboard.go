package data

import (
	"context"
	"math"
	"sort"
	"strconv"
	"time"

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

// GetDashboardStats 获取仪表盘统计数据
// 真实数据来源：今日订单数/今日营收(OrderStats)、总用户数(UserCount)、总商品数(ListProduct)
// 增长率基于今日 vs 昨日对比（昨日订单数来自趋势统计，昨日营收来自 OrderStats）
func (r *dashboardRepo) GetDashboardStats(ctx context.Context) (*biz.DashboardStats, error) {
	// 并行查询各服务
	type statsResult struct {
		orderStats *orderRpc.OrderStatsReply
		trend      *orderRpc.GetOrderTrendReply
		userCount  int64
		prodCount  int32
	}

	result := &statsResult{}
	errCh := make(chan error, 4)

	// 查询订单统计（含今日/昨日营收）
	go func() {
		stats, err := r.data.oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
		if err != nil {
			errCh <- err
			return
		}
		result.orderStats = stats
		errCh <- nil
	}()

	// 查询近 2 天订单趋势（用于计算昨日订单数）
	go func() {
		trend, err := r.data.oc.GetOrderTrend(ctx, &orderRpc.GetOrderTrendRequest{Days: 2})
		if err != nil {
			errCh <- err
			return
		}
		result.trend = trend
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

	// 等待所有查询完成，部分失败则用零值/默认值
	for i := 0; i < 4; i++ {
		if err := <-errCh; err != nil {
			r.log.Warnf("获取仪表盘数据部分失败: %v", err)
		}
	}

	stats := &biz.DashboardStats{
		TodayOrders:   int64(result.orderStats.GetTodayCount()),
		TotalUsers:    result.userCount,
		TotalProducts: int64(result.prodCount),
	}
	// 今日营收（分 → 元，整数不带小数）
	stats.TodayRevenue = formatRevenue(float64(result.orderStats.GetTodayRevenue()) / 100)

	// 增长率：今日 vs 昨日
	stats.OrdersGrowth = growthRate(int64(result.orderStats.GetTodayCount()), yesterdayTrendCount(result.trend))
	stats.RevenueGrowth = growthRate(result.orderStats.GetTodayRevenue(), result.orderStats.GetYesterdayRevenue())

	return stats, nil
}

// GetOrderTrend 获取近14日订单趋势（真实数据：按日统计订单数）
func (r *dashboardRepo) GetOrderTrend(ctx context.Context) ([]*biz.TrendItem, error) {
	reply, err := r.data.oc.GetOrderTrend(ctx, &orderRpc.GetOrderTrendRequest{Days: 14})
	if err != nil {
		return nil, err
	}
	items := make([]*biz.TrendItem, 0, len(reply.Items))
	for _, it := range reply.Items {
		items = append(items, &biz.TrendItem{Date: it.Date, Value: it.OrderCount})
	}
	return items, nil
}

// GetRevenueCategories 获取营收分类占比（近30天已完成订单按商品分类聚合）
// 编排流程：已完成订单明细(order) → 商品分类映射(product) → 按分类聚合金额
func (r *dashboardRepo) GetRevenueCategories(ctx context.Context) ([]*biz.RevenueCategory, error) {
	// 1. 拉取近 30 天已完成订单 ID
	loc, _ := time.LoadLocation("Asia/Shanghai")
	dateFrom := time.Now().In(loc).AddDate(0, 0, -30).Format("2006-01-02")
	orderIDs, err := r.listCompletedOrderIDs(ctx, dateFrom)
	if err != nil {
		return nil, err
	}

	// 2. 批量获取订单明细，按商品聚合营收
	amountByProduct := make(map[int64]int64)
	if len(orderIDs) > 0 {
		batchReply, err := r.data.oc.BatchGetOrder(ctx, &orderRpc.BatchGetOrderRequest{OrderIds: orderIDs})
		if err != nil {
			return nil, err
		}
		for _, o := range batchReply.Orders {
			for _, item := range o.Items {
				amountByProduct[item.ProductId] += item.Price * int64(item.Quantity)
			}
		}
	}
	if len(amountByProduct) == 0 {
		return nil, nil
	}

	// 3. 商品 → 分类映射（全量分页拉取）
	catIDByProduct := make(map[int64]int64)
	for page := int32(1); ; page++ {
		reply, err := r.data.pc.ListProduct(ctx, &productRpc.ListProductRequest{Page: page, PageSize: 200})
		if err != nil {
			return nil, err
		}
		for _, p := range reply.Items {
			catIDByProduct[p.ProductId] = p.CategoryId
		}
		if int32(len(catIDByProduct)) >= reply.Total || len(reply.Items) == 0 {
			break
		}
	}

	// 4. 分类名称映射（扁平化分类树）
	catReply, err := r.data.cc.ListCategoryTree(ctx, &productRpc.ListCategoryTreeRequest{})
	if err != nil {
		return nil, err
	}
	catName := make(map[int64]string)
	var walk func(nodes []*productRpc.CategoryNode)
	walk = func(nodes []*productRpc.CategoryNode) {
		for _, n := range nodes {
			catName[n.Id] = n.Name
			walk(n.Children)
		}
	}
	walk(catReply.Items)

	// 5. 按分类聚合营收
	amountByCat := make(map[int64]int64)
	for pid, amount := range amountByProduct {
		amountByCat[catIDByProduct[pid]] += amount
	}

	// 6. 排序取 Top5，计算百分比并分配颜色
	type catAmount struct {
		name   string
		amount int64
	}
	list := make([]catAmount, 0, len(amountByCat))
	for cid, amt := range amountByCat {
		name := catName[cid]
		if name == "" {
			name = "其他"
		}
		list = append(list, catAmount{name: name, amount: amt})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].amount > list[j].amount })
	if len(list) > 5 {
		list = list[:5]
	}
	var total int64
	for _, ca := range list {
		total += ca.amount
	}
	colors := []string{"#5470c6", "#91cc75", "#fac858", "#ee6666", "#73c0de"}
	cats := make([]*biz.RevenueCategory, 0, len(list))
	for i, ca := range list {
		percent := 0.0
		if total > 0 {
			percent = math.Round(float64(ca.amount)/float64(total)*1000) / 10
		}
		cats = append(cats, &biz.RevenueCategory{
			Name:       ca.name,
			Percentage: percent,
			Color:      colors[i%len(colors)],
		})
	}
	return cats, nil
}

// listCompletedOrderIDs 分页拉取指定日期范围内已完成订单（status=6）的 ID 列表
func (r *dashboardRepo) listCompletedOrderIDs(ctx context.Context, dateFrom string) ([]int64, error) {
	const (
		pageSize = 100
		maxPages = 10 // 最多统计 1000 单
	)
	orderIDs := make([]int64, 0, pageSize)
	for page := int32(1); page <= maxPages; page++ {
		reply, err := r.data.oc.ListOrder(ctx, &orderRpc.ListOrderRequest{
			Page:     page,
			PageSize: pageSize,
			// 已完成状态（status=6），使用 proto 常量避免魔法数字
			Status:   strconv.Itoa(int(orderRpc.OrderStatus_Complete)),
			DateFrom: dateFrom,
		})
		if err != nil {
			return nil, err
		}
		for _, o := range reply.Items {
			orderIDs = append(orderIDs, o.OrderId)
		}
		if int32(len(orderIDs)) >= reply.Total || len(reply.Items) == 0 {
			break
		}
	}
	return orderIDs, nil
}

// yesterdayTrendCount 从近 2 天趋势中取昨日订单数（趋势第 0 项为昨天）
func yesterdayTrendCount(trend *orderRpc.GetOrderTrendReply) int64 {
	if trend == nil || len(trend.Items) == 0 {
		return 0
	}
	return int64(trend.Items[0].OrderCount)
}

// growthRate 计算今日相对昨日的增长率（百分比，昨日为 0 时返回 0）
func growthRate(today, yesterday int64) float64 {
	if yesterday <= 0 {
		return 0
	}
	return math.Round((float64(today)-float64(yesterday))/float64(yesterday)*1000) / 10
}

// formatRevenue 金额格式化：去掉无意义的小数（如 45680.0 → "45680"）
func formatRevenue(yuan float64) string {
	return strconv.FormatFloat(yuan, 'f', -1, 64)
}
