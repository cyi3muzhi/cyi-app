package data

import (
	"context"
	"fmt"
	"time"

	orderRpc "cyi-app/api/service/order/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// orderRepo 订单数据仓库，通过 gRPC 调用订单服务和用户服务
type orderRepo struct {
	log  *log.Helper
	data *Data
}

// NewOrderRepo 创建订单数据仓库
func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// orderStatusMap 订单状态数字到中文文本的映射
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

// ListOrder 通过 gRPC 调用订单服务获取订单列表
func (r *orderRepo) ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*biz.Order, int32, error) {
	// 将管理端状态编码转换为订单服务的数字状态
	orderStatus := parseAdminOrderStatus(status)

	// 调用订单服务获取列表
	reply, err := r.data.oc.ListOrder(ctx, &orderRpc.ListOrderRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   orderStatus,
		Keyword:  keyword,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Order, 0, len(reply.Items))
	for _, item := range reply.Items {
		// 时间戳转格式化字符串
		createdAt := time.Unix(item.CreatedAt, 0).Format("2006-01-02 15:04:05")
		// 状态文本
		statusText := orderStatusText(item.Status)

		order := &biz.Order{
			ID:          item.OrderId,
			OrderNo:     item.OrderNo,
			CreatedAt:   createdAt,
			Amount:      item.PayAmount,
			Status:      formatOrderStatus(item.Status),
			StatusText:  statusText,
			Freight:     item.Freight,
			DiscountAmt: item.DiscountAmount,
		}
		items = append(items, order)
	}

	// 批量获取用户名称（去重后查询）
	userIDs := make([]int64, 0, len(items))
	userIDSet := make(map[int64]struct{})
	for _, item := range items {
		// 从 reply.Items 中查找对应的 user_id
		for _, ri := range reply.Items {
			if ri.OrderId == item.ID {
				if _, ok := userIDSet[ri.UserId]; !ok {
					userIDSet[ri.UserId] = struct{}{}
					userIDs = append(userIDs, ri.UserId)
				}
				break
			}
		}
	}

	// 查询用户信息
	userNameMap := r.batchGetUserNames(ctx, userIDs)

	// 填充用户名称
	for _, item := range items {
		for _, ri := range reply.Items {
			if ri.OrderId == item.ID {
				if name, ok := userNameMap[ri.UserId]; ok {
					item.UserName = name
				}
				break
			}
		}
	}

	return items, reply.Total, nil
}

// OrderStats 通过 gRPC 调用订单服务获取订单统计数据
func (r *orderRepo) OrderStats(ctx context.Context) (total, pending, today, refund int32, err error) {
	reply, err := r.data.oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return reply.TotalCount, reply.PendingCount, reply.TodayCount, reply.RefundCount, nil
}

// userDisplay 从 GetUserReply 提取用户显示名（优先 Nickname，其次 Username）
func userDisplay(u *userRpc.GetUserReply) string {
	if u == nil || u.User == nil {
		return ""
	}
	if u.User.Nickname != "" {
		return u.User.Nickname
	}
	return u.User.Username
}

// batchGetUserNames 批量查询用户名称
func (r *orderRepo) batchGetUserNames(ctx context.Context, userIDs []int64) map[int64]string {
	result := make(map[int64]string, len(userIDs))
	for _, uid := range userIDs {
		user, err := r.data.uc.GetUser(ctx, &userRpc.GetUserRequest{
			Id: uid,
		})
		if err != nil {
			r.log.Warnf("failed to get user %d: %v", uid, err)
			result[uid] = fmt.Sprintf("用户% d", uid)
			continue
		}
		result[uid] = userDisplay(user)
	}
	return result
}

// formatOrderStatus 将订单状态数字映射为管理端使用的状态编码
// 订单服务 OrderStatus: 0=未指定, 1=待支付, 2=已支付, 3=待发货, 4=已发货, 5=待收货, 6=已完成, 7=已取消, 8=退款中, 9=已退款
// 管理端状态编码: pending=待支付, paid=已支付, shipped=已发货, completed=已完成, cancelled=已取消, refunding=退款中
func formatOrderStatus(status int32) string {
	switch status {
	case 1:
		return "pending"
	case 2:
		return "paid"
	case 3:
		return "paid"
	case 4:
		return "shipped"
	case 5:
		return "shipped"
	case 6:
		return "completed"
	case 7:
		return "cancelled"
	case 8:
		return "refunding"
	case 9:
		return "refunded"
	default:
		return "unknown"
	}
}

// parseAdminOrderStatus 将管理端状态编码转换为订单服务的数字状态（支持多状态）
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

// GetOrder 通过 gRPC 调用订单服务获取订单详情
func (r *orderRepo) GetOrder(ctx context.Context, orderID int64) (*biz.Order, error) {
	reply, err := r.data.oc.GetOrder(ctx, &orderRpc.GetOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return nil, err
	}

	statusText := orderStatusText(reply.Status)

	// 查询用户名称
	userName := ""
	user, err := r.data.uc.GetUser(ctx, &userRpc.GetUserRequest{
		Id: reply.UserId,
	})
	if err != nil {
		r.log.Warnf("failed to get user %d: %v", reply.UserId, err)
		userName = fmt.Sprintf("用户% d", reply.UserId)
	} else {
		userName = userDisplay(user)
	}

	order := &biz.Order{
		ID:          reply.OrderId,
		OrderNo:     reply.OrderNo,
		CreatedAt:   time.Unix(reply.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UserName:    userName,
		Amount:      reply.PayAmount,
		Status:      formatOrderStatus(reply.Status),
		StatusText:  statusText,
		Freight:     reply.Freight,
		DiscountAmt: reply.DiscountAmount,
		// 物流信息（发货后展示）
		LogisticsCompany: reply.LogisticsCompany,
		LogisticsNo:      reply.LogisticsNo,
		Receiver:         reply.Receiver,
		Phone:            reply.Phone,
		Address:          reply.Address,
	}

	for _, item := range reply.Items {
		order.Products = append(order.Products, &biz.OrderProduct{
			Name:  item.ProductName,
			Spec:  item.SpecDesc,
			Qty:   item.Quantity,
			Price: item.Price,
			Image: r.resolveImageURL(ctx, item.Image),
		})
	}

	return order, nil
}

// resolveImageURL 解析商品图片：OSS 路径转 presigned URL，已有 URL 原样返回
func (r *orderRepo) resolveImageURL(ctx context.Context, image string) string {
	return r.data.resolvePresignedURL(ctx, image, 3600*24*30)
}

// ShipOrder 通过 gRPC 调用订单服务发货（状态流转 + 写入物流信息）
func (r *orderRepo) ShipOrder(ctx context.Context, orderID int64, logisticsCompany string, logisticsNo string) error {
	_, err := r.data.oc.ShipOrder(ctx, &orderRpc.ShipOrderRequest{
		OrderId:          orderID,
		LogisticsCompany: logisticsCompany,
		LogisticsNo:      logisticsNo,
	})
	if err != nil {
		r.log.Errorf("ship order failed, orderID=%d: %v", orderID, err)
		return err
	}
	return nil
}

// CancelOrder 通过 gRPC 调用订单服务取消订单（userID=0 表示管理端操作，不校验归属）
func (r *orderRepo) CancelOrder(ctx context.Context, orderID int64, reason string) error {
	_, err := r.data.oc.CancelOrder(ctx, &orderRpc.CancelOrderRequest{
		OrderId: orderID,
		UserId:  0,
		Reason:  reason,
	})
	if err != nil {
		r.log.Errorf("cancel order failed, orderID=%d: %v", orderID, err)
		return err
	}
	return nil
}

// ListRefund 通过 gRPC 调用订单服务获取退款/售后单列表，并填充用户名
func (r *orderRepo) ListRefund(ctx context.Context, page, pageSize int32, status string, keyword string) ([]*biz.Refund, int32, error) {
	reply, err := r.data.oc.ListRefund(ctx, &orderRpc.ListRefundRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  keyword,
	})
	if err != nil {
		r.log.Errorf("list refund failed: %v", err)
		return nil, 0, err
	}

	items := make([]*biz.Refund, 0, len(reply.Items))
	userIDSet := make(map[int64]struct{})
	for _, item := range reply.Items {
		items = append(items, &biz.Refund{
			ID:           item.RefundId,
			OrderID:      item.OrderId,
			OrderNo:      item.OrderNo,
			UserID:       item.UserId,
			RefundAmount: item.RefundAmount,
			Reason:       item.Reason,
			Status:       item.Status,
			StatusText:   refundStatusText(item.Status),
			RejectReason: item.RejectReason,
			Auditor:      item.Auditor,
			CreatedAt:    time.Unix(item.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			OriginStatus: item.OriginStatus,
		})
		if _, ok := userIDSet[item.UserId]; !ok {
			userIDSet[item.UserId] = struct{}{}
		}
	}

	// 批量查询用户名并填充
	userIDs := make([]int64, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}
	userNameMap := r.batchGetUserNames(ctx, userIDs)
	for _, item := range items {
		if name, ok := userNameMap[item.UserID]; ok {
			item.UserName = name
		}
	}
	return items, reply.Total, nil
}

// AuditRefund 通过 gRPC 调用订单服务审核退款（通过/拒绝）
func (r *orderRepo) AuditRefund(ctx context.Context, refundID int64, approve bool, rejectReason, auditor string) error {
	_, err := r.data.oc.AuditRefund(ctx, &orderRpc.AuditRefundRequest{
		RefundId:     refundID,
		Approve:      approve,
		RejectReason: rejectReason,
		Auditor:      auditor,
	})
	if err != nil {
		r.log.Errorf("audit refund failed, refundID=%d, approve=%v: %v", refundID, approve, err)
		return err
	}
	return nil
}

// GetLogistics 通过 gRPC 调用订单服务查询物流轨迹
func (r *orderRepo) GetLogistics(ctx context.Context, orderID int64) (*biz.Logistics, error) {
	reply, err := r.data.oc.GetLogistics(ctx, &orderRpc.GetLogisticsRequest{
		OrderId: orderID,
	})
	if err != nil {
		r.log.Errorf("get logistics failed, orderID=%d: %v", orderID, err)
		return nil, err
	}
	logistics := &biz.Logistics{
		OrderID:          reply.OrderId,
		LogisticsCompany: reply.LogisticsCompany,
		LogisticsNo:      reply.LogisticsNo,
		ShipTime:         reply.ShipTime,
		Traces:           make([]*biz.LogisticsTrace, 0, len(reply.Traces)),
	}
	if reply.ShipTime > 0 {
		logistics.ShipTimeText = time.Unix(reply.ShipTime, 0).Format("2006-01-02 15:04:05")
	}
	for _, t := range reply.Traces {
		logistics.Traces = append(logistics.Traces, &biz.LogisticsTrace{
			Time:     t.Time,
			TimeText: time.Unix(t.Time, 0).Format("2006-01-02 15:04:05"),
			Desc:     t.Desc,
			Status:   t.Status,
		})
	}
	return logistics, nil
}

// refundStatusText 退款单状态 -> 中文文本
func refundStatusText(status int32) string {
	switch status {
	case 1:
		return "待审核"
	case 2:
		return "已通过"
	case 3:
		return "已拒绝"
	default:
		return "未知"
	}
}
