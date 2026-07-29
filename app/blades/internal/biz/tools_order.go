package biz

import (
	"context"
	"fmt"

	orderRpc "cyi-app/api/service/order/v1"
	"github.com/go-kratos/blades/tools"
)

// RegisterOrderTools 注册订单工具
func RegisterOrderTools(r *ToolRegistry, client orderRpc.OrderClient) {
	// query_orders
	t1, _ := tools.NewFunc("query_orders", "按日期范围、状态、关键词查询订单列表",
		func(ctx context.Context, input struct {
			Page     int32  `json:"page"`
			PageSize int32  `json:"page_size"`
			Status   string `json:"status"`
			Keyword  string `json:"keyword"`
			DateFrom string `json:"date_from"`
			DateTo   string `json:"date_to"`
		}) (string, error) {
			req := &orderRpc.ListOrderRequest{
				Page:     input.Page,
				PageSize: input.PageSize,
				Status:   input.Status,
				Keyword:  input.Keyword,
				DateFrom: input.DateFrom,
				DateTo:   input.DateTo,
			}
			reply, err := client.ListOrder(ctx, req)
			if err != nil {
				return "", err
			}
			if reply.Total == 0 {
				return "没有找到匹配的订单", nil
			}
			result := fmt.Sprintf("共 %d 条订单：\n", reply.Total)
			for _, item := range reply.Items {
				result += fmt.Sprintf("- 订单#%d | %s | ￥%.2f | %s\n",
					item.OrderId, item.OrderNo, float64(item.PayAmount)/100, orderStatusText(int32(item.Status)))
			}
			return result, nil
		})
	r.Add(t1)

	// query_order_detail
	t2, _ := tools.NewFunc("query_order_detail", "查询单个订单的详细信息",
		func(ctx context.Context, input struct {
			OrderId int64 `json:"order_id"`
		}) (string, error) {
			reply, err := client.GetOrder(ctx, &orderRpc.GetOrderRequest{OrderId: input.OrderId})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("订单 #%d\n单号: %s\n状态: %s\n支付金额: ￥%.2f",
				reply.OrderId, reply.OrderNo, orderStatusText(int32(reply.Status)), float64(reply.PayAmount)/100), nil
		})
	r.Add(t2)
}

func orderStatusText(status int32) string {
	switch status {
	case 1:
		return "待支付"
	case 2:
		return "已支付"
	case 3:
		return "待发货"
	case 4:
		return "已发货"
	case 5:
		return "待收货"
	case 6:
		return "已完成"
	case 7:
		return "已取消"
	case 8:
		return "退款中"
	case 9:
		return "已退款"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}
