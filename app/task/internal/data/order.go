package data

import (
	"context"
	"fmt"
	"time"

	orderRpc "cyi-app/api/service/order/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/task/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// orderStatusMap 订单状态数字到中文文本的映射
// 订单服务 OrderStatus: 1=待支付, 2=已支付, 3=待发货, 4=已发货, 5=待收货, 6=已完成, 7=已取消, 8=退款中, 9=已退款
var orderStatusMap = map[int32]string{
	1: "待支付",
	2: "已支付",
	3: "待发货",
	4: "已发货",
	5: "待收货",
	6: "已完成",
	7: "已取消",
	8: "退款中",
	9: "已退款",
}

// orderStatusText 根据状态数字返回状态文本
func orderStatusText(status int32) string {
	if text, ok := orderStatusMap[status]; ok {
		return text
	}
	return "未知"
}

// parseAdminOrderStatus 将管理端状态编码转换为订单服务的数字状态（支持多状态）
// admin 投递任务时传的是管理端编码（paid/shipped 等），task 消费时需转回数字状态查询订单服务。
// 返回逗号分隔的数字字符串，如 "2,3"
func parseAdminOrderStatus(status string) string {
	switch status {
	case "pending":
		return "1"
	case "paid":
		return "2,3"
	case "shipped":
		return "4,5"
	case "completed":
		return "6"
	case "cancelled":
		return "7"
	case "refunding":
		return "8"
	case "refunded":
		return "9"
	default:
		return status
	}
}

// orderRepo 订单查询仓库，通过 order gRPC + user gRPC 查询订单用于导出
type orderRepo struct {
	log  *log.Helper
	data *Data
}

// NewOrderRepo 创建订单查询仓库
func NewOrderRepo(data *Data, logger log.Logger) biz.OrderQueryRepo {
	return &orderRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// ListOrdersForExport 分页查询订单用于导出
// 通过 order gRPC 拉订单列表，BatchGetOrder 批量补全商品明细，再用 user gRPC 批量补全用户名。
// 注意：order proto 未定义支付方式字段，Payment 列暂显示"-"，后续如需可在 proto 中补充。
func (r *orderRepo) ListOrdersForExport(ctx context.Context, page, pageSize int32, status, dateFrom, dateTo string) ([]*biz.ExportOrder, int32, error) {
	// 管理端状态编码转订单服务数字状态
	orderStatus := parseAdminOrderStatus(status)

	reply, err := r.data.oc.ListOrder(ctx, &orderRpc.ListOrderRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   orderStatus,
		Keyword:  "",
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.ExportOrder, 0, len(reply.Items))
	userIDSet := make(map[int64]struct{})
	userIDs := make([]int64, 0, len(reply.Items))
	for _, item := range reply.Items {
		createdAt := time.Unix(item.CreatedAt, 0).Format("2006-01-02 15:04:05")
		order := &biz.ExportOrder{
			OrderNo:    item.OrderNo,
			CreatedAt:  createdAt,
			Amount:     item.PayAmount,
			StatusText: orderStatusText(item.Status),
		}
		items = append(items, order)
		// 收集去重 user_id
		if _, ok := userIDSet[item.UserId]; !ok {
			userIDSet[item.UserId] = struct{}{}
			userIDs = append(userIDs, item.UserId)
		}
	}

	// 批量补全商品明细（1 次 BatchGetOrder 替代 N 次 GetOrder）
	r.batchFillProducts(ctx, items, reply.Items)

	// 批量查询用户名
	userNameMap := r.batchGetUserNames(ctx, userIDs)

	// 回填用户名
	for i, item := range reply.Items {
		if name, ok := userNameMap[item.UserId]; ok {
			items[i].UserName = name
		}
	}

	return items, reply.Total, nil
}

// batchFillProducts 批量补全订单商品明细
// 通过 BatchGetOrder 一次性查询所有订单的商品明细（1 次 gRPC 替代 N 次 GetOrder）。
// orders 与 rpcItems 按索引一一对应；查询失败时整体回退，商品信息留空显示"-"。
func (r *orderRepo) batchFillProducts(ctx context.Context, orders []*biz.ExportOrder, rpcItems []*orderRpc.ListOrderItem) {
	if len(rpcItems) == 0 {
		return
	}
	// 收集 order_id
	orderIDs := make([]int64, 0, len(rpcItems))
	for _, item := range rpcItems {
		orderIDs = append(orderIDs, item.OrderId)
	}
	// 1 次批量调用替代 N 次 GetOrder
	reply, err := r.data.oc.BatchGetOrder(ctx, &orderRpc.BatchGetOrderRequest{OrderIds: orderIDs})
	if err != nil {
		r.log.Warnf("export: batch get orders failed: %v", err)
		return
	}
	// 按 order_id 建立 products 映射
	productsMap := make(map[int64][]*biz.ExportOrderProduct, len(reply.Orders))
	for _, o := range reply.Orders {
		products := make([]*biz.ExportOrderProduct, 0, len(o.Items))
		for _, it := range o.Items {
			products = append(products, &biz.ExportOrderProduct{
				Name: it.ProductName,
				Spec: it.SpecDesc,
				Qty:  it.Quantity,
			})
		}
		productsMap[o.OrderId] = products
	}
	// 按 rpcItems 索引回填到 orders（两者按索引一一对应）
	for i, rpcItem := range rpcItems {
		if products, ok := productsMap[rpcItem.OrderId]; ok {
			orders[i].Products = products
		}
	}
}

// batchGetUserNames 批量查询用户名称
func (r *orderRepo) batchGetUserNames(ctx context.Context, userIDs []int64) map[int64]string {
	result := make(map[int64]string, len(userIDs))
	for _, uid := range userIDs {
		user, err := r.data.userC.GetUser(ctx, &userRpc.GetUserRequest{
			Id: uid,
		})
		if err != nil {
			r.log.Warnf("failed to get user %d: %v", uid, err)
			result[uid] = fmt.Sprintf("用户% d", uid)
			continue
		}
		result[uid] = user.GetName()
	}
	return result
}
