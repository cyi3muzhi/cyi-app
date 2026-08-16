package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	marketingRpc "cyi-app/api/service/marketing/v1"

	"github.com/go-kratos/blades/tools"
)

// RegisterMarketingTools 注册优惠券工具（营销服务）
func RegisterMarketingTools(r *ToolRegistry, client marketingRpc.CouponServiceClient) {
	// query_coupons 优惠券列表
	t1, _ := tools.NewFunc("query_coupons", "查询优惠券列表，支持按状态、关键词筛选。当用户问'有哪些优惠券'、'优惠券活动'时调用",
		func(ctx context.Context, input struct {
			Page     int32  `json:"page"`
			PageSize int32  `json:"page_size"`
			Status   string `json:"status"`
			Keyword  string `json:"keyword"`
		}) (string, error) {
			page := input.Page
			if page <= 0 {
				page = 1
			}
			pageSize := input.PageSize
			if pageSize <= 0 || pageSize > 50 {
				pageSize = 20
			}
			reply, err := client.ListCoupon(ctx, &marketingRpc.ListCouponRequest{
				Page:     page,
				PageSize: pageSize,
				Status:   input.Status,
				Keyword:  input.Keyword,
			})
			if err != nil {
				return errText(err), nil
			}
			if reply.Total == 0 {
				return "没有找到匹配的优惠券", nil
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("共 %d 张优惠券，明细如下：\n\n", reply.Total))
			sb.WriteString("| 券名称 | 类型 | 面额 | 已发 | 已用 | 状态 |\n")
			sb.WriteString("|---|---|---|---|---|---|\n")
			for _, c := range reply.Items {
				state := "启用中"
				if !c.Enabled {
					state = "已停用"
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | ￥%.2f | %d | %d | %s |\n",
					c.Name, couponTypeText(c.Type), float64(c.Value)/100, c.IssuedCount, c.UsedCount, state))
			}
			return sb.String(), nil
		})
	r.Add(t1)

	// query_coupon_stats 优惠券统计
	t2, _ := tools.NewFunc("query_coupon_stats", "查询优惠券整体统计：总数、进行中、已发放、已过期。当用户问'优惠券数据'、'优惠券统计'时调用",
		func(ctx context.Context, _ struct{}) (string, error) {
			reply, err := client.CouponStats(ctx, &marketingRpc.CouponStatsRequest{})
			if err != nil {
				return errText(err), nil
			}
			return fmt.Sprintf("优惠券统计：\n\n| 指标 | 数量 |\n|---|---|\n| 优惠券总数 | %d |\n| 进行中 | %d |\n| 已发放 | %d 张 |\n| 已过期 | %d |\n",
				reply.TotalCount, reply.ActiveCount, reply.IssuedCount, reply.ExpiredCount), nil
		})
	r.Add(t2)

	// query_coupon_records 发放记录
	t3, _ := tools.NewFunc("query_coupon_records", "查询某张优惠券的发放记录：哪些用户领取、领取时间、使用状态。当用户问'优惠券发放记录'、'谁领了这张券'时调用",
		func(ctx context.Context, input struct {
			CouponId int64 `json:"coupon_id"`
			Page     int32 `json:"page"`
			PageSize int32 `json:"page_size"`
		}) (string, error) {
			if input.CouponId <= 0 {
				return "", fmt.Errorf("请提供有效的 coupon_id")
			}
			page := input.Page
			if page <= 0 {
				page = 1
			}
			pageSize := input.PageSize
			if pageSize <= 0 || pageSize > 50 {
				pageSize = 20
			}
			reply, err := client.ListDistRecords(ctx, &marketingRpc.ListDistRecordsRequest{
				CouponId: input.CouponId,
				Page:     page,
				PageSize: pageSize,
			})
			if err != nil {
				return errText(err), nil
			}
			if reply.Total == 0 {
				return "该优惠券暂无发放记录", nil
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("优惠券 #%d 共 %d 条发放记录：\n\n", input.CouponId, reply.Total))
			sb.WriteString("| 用户ID | 状态 | 来源 | 领取时间 |\n")
			sb.WriteString("|---|---|---|---|\n")
			for _, rec := range reply.Items {
				sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
					rec.UserId, distStatusText(rec.Status), rec.Source, formatTs(rec.CreateTime)))
			}
			return sb.String(), nil
		})
	r.Add(t3)
}

// couponTypeText 优惠券类型中文描述
func couponTypeText(t int32) string {
	switch t {
	case 1:
		return "立减"
	case 2:
		return "折扣"
	case 3:
		return "满减"
	default:
		return fmt.Sprintf("类型%d", t)
	}
}

// distStatusText 发放记录状态中文描述
func distStatusText(status int32) string {
	switch status {
	case 0:
		return "未使用"
	case 1:
		return "已使用"
	case 2:
		return "已过期"
	default:
		return fmt.Sprintf("状态%d", status)
	}
}

// formatTs Unix 秒时间戳格式化为日期时间
func formatTs(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04")
}
