package biz

import (
	"context"
	"fmt"
	"strings"

	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"

	"github.com/go-kratos/blades/tools"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterStatsTools 注册统计类工具（订单统计/趋势、今日概览、用户统计、用户详情）
func RegisterStatsTools(r *ToolRegistry, oc orderRpc.OrderClient, uc userRpc.UserClient, pc productRpc.ProductClient) {
	// query_order_stats 订单统计
	t1, _ := tools.NewFunc("query_order_stats", "查询订单统计信息：总订单数、待处理数（待支付+待发货）、今日订单数、退款数、今日营收。当用户问'今天订单多少'、'待处理订单'、'订单营收'时调用",
		func(ctx context.Context, _ struct{}) (string, error) {
			reply, err := oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
			if err != nil {
				return "", err
			}
			var sb strings.Builder
			sb.WriteString("订单统计：\n\n")
			sb.WriteString("| 指标 | 数值 |\n")
			sb.WriteString("|---|---|\n")
			sb.WriteString(fmt.Sprintf("| 总订单数 | %d |\n", reply.TotalCount))
			sb.WriteString(fmt.Sprintf("| 待处理（待支付/待发货） | %d |\n", reply.PendingCount))
			sb.WriteString(fmt.Sprintf("| 今日订单 | %d（昨日 %d） |\n", reply.TodayCount, reply.YesterdayCount))
			sb.WriteString(fmt.Sprintf("| 今日营收 | ￥%.2f（昨日 ￥%.2f） |\n", float64(reply.TodayRevenue)/100, float64(reply.YesterdayRevenue)/100))
			sb.WriteString(fmt.Sprintf("| 退款/售后 | %d |\n", reply.RefundCount))
			return sb.String(), nil
		})
	r.Add(t1)

	// query_order_trend 订单趋势
	t2, _ := tools.NewFunc("query_order_trend", "查询最近 N 天的每日订单数量趋势（默认 14 天）。当用户问'订单趋势'、'近几天订单走势'时调用",
		func(ctx context.Context, input struct {
			Days int32 `json:"days"`
		}) (string, error) {
			days := input.Days
			if days <= 0 {
				days = 14
			}
			reply, err := oc.GetOrderTrend(ctx, &orderRpc.GetOrderTrendRequest{Days: days})
			if err != nil {
				return "", err
			}
			if len(reply.Items) == 0 {
				return "暂无订单趋势数据", nil
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("近 %d 日订单趋势：\n\n", days))
			sb.WriteString("| 日期 | 订单数 |\n")
			sb.WriteString("|---|---|\n")
			for _, item := range reply.Items {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", item.Date, item.OrderCount))
			}
			return sb.String(), nil
		})
	r.Add(t2)

	// query_dashboard_stats 今日运营概览
	t3, _ := tools.NewFunc("query_dashboard_stats", "查询今日运营概览：今日订单数、今日营收、平台总用户数、今日新增用户、商品总数。当用户问'今天生意怎么样'、'店铺数据'、'运营情况'时调用",
		func(ctx context.Context, _ struct{}) (string, error) {
			orderReply, err := oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
			if err != nil {
				return "", fmt.Errorf("查询订单统计失败: %v", err)
			}
			userReply, err := uc.GetUserStats(ctx, &emptypb.Empty{})
			if err != nil {
				return "", fmt.Errorf("查询用户统计失败: %v", err)
			}
			productReply, err := pc.ListProduct(ctx, &productRpc.ListProductRequest{Page: 1, PageSize: 1})
			if err != nil {
				return "", fmt.Errorf("查询商品统计失败: %v", err)
			}
			var sb strings.Builder
			sb.WriteString("今日运营概览：\n\n")
			sb.WriteString("| 指标 | 数值 |\n")
			sb.WriteString("|---|---|\n")
			sb.WriteString(fmt.Sprintf("| 今日订单 | %d 单 |\n", orderReply.TodayCount))
			sb.WriteString(fmt.Sprintf("| 今日营收 | ￥%.2f |\n", float64(orderReply.TodayRevenue)/100))
			sb.WriteString(fmt.Sprintf("| 平台总用户 | %d 人（今日新增 %d） |\n", userReply.TotalUsers, userReply.TodayNew))
			sb.WriteString(fmt.Sprintf("| 商品总数 | %d 个 |\n", productReply.Total))
			return sb.String(), nil
		})
	r.Add(t3)

	// query_user_stats 用户统计
	t4, _ := tools.NewFunc("query_user_stats", "查询平台用户统计信息：总用户数、今日新增、活跃用户、会员用户数及占比。当用户问'多少用户'、'用户增长'、'注册了多少人'时调用",
		func(ctx context.Context, _ struct{}) (string, error) {
			reply, err := uc.GetUserStats(ctx, &emptypb.Empty{})
			if err != nil {
				return "", err
			}
			var sb strings.Builder
			sb.WriteString("用户统计：\n\n")
			sb.WriteString("| 指标 | 数值 |\n")
			sb.WriteString("|---|---|\n")
			sb.WriteString(fmt.Sprintf("| 总用户数 | %d |\n", reply.TotalUsers))
			sb.WriteString(fmt.Sprintf("| 今日新增 | %d |\n", reply.TodayNew))
			sb.WriteString(fmt.Sprintf("| 活跃用户 | %d |\n", reply.ActiveUsers))
			sb.WriteString(fmt.Sprintf("| 会员用户 | %d（占比 %.1f%%） |\n", reply.MemberUsers, reply.MemberRatio*100))
			return sb.String(), nil
		})
	r.Add(t4)

	// query_user_detail 用户详情
	t5, _ := tools.NewFunc("query_user_detail", "查询单个用户的详细信息：昵称、手机号、邮箱、会员等级、状态、注册时间等。当用户问'某个用户的情况'、'查一下用户'时调用",
		func(ctx context.Context, input struct {
			UserId int64 `json:"user_id"`
		}) (string, error) {
			if input.UserId <= 0 {
				return "", fmt.Errorf("请提供有效的 user_id")
			}
			reply, err := uc.GetUser(ctx, &userRpc.GetUserRequest{Id: input.UserId})
			if err != nil {
				return "", err
			}
			u := reply.User
			if u == nil {
				return "未找到该用户", nil
			}
			statusText := "正常"
			if u.Status == 2 {
				statusText = "已冻结"
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("用户 #%d 详情：\n\n", u.Id))
			sb.WriteString("| 字段 | 值 |\n")
			sb.WriteString("|---|---|\n")
			sb.WriteString(fmt.Sprintf("| 昵称 | %s |\n", u.Nickname))
			sb.WriteString(fmt.Sprintf("| 账号 | %s |\n", u.Username))
			sb.WriteString(fmt.Sprintf("| 手机号 | %s |\n", u.Phone))
			sb.WriteString(fmt.Sprintf("| 邮箱 | %s |\n", u.Email))
			sb.WriteString(fmt.Sprintf("| 会员等级 | %s |\n", u.Level))
			sb.WriteString(fmt.Sprintf("| 状态 | %s |\n", statusText))
			sb.WriteString(fmt.Sprintf("| 注册来源 | %s |\n", u.RegisterSource))
			sb.WriteString(fmt.Sprintf("| 注册时间 | %s |\n", formatTs(u.CreatedAt)))
			return sb.String(), nil
		})
	r.Add(t5)
}
