package data

import (
	"context"
	"cyi-app/app/cin/interface/internal/biz"
	cr "cyi-app/pkg/util/redis"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	v1 "cyi-app/api/cin/interface/v1"
	orderV1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type orderRepo struct {
	log  *log.Helper
	data *Data
	sg   *singleflight.Group
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		log:  log.NewHelper(logger),
		data: data,
		sg:   &singleflight.Group{},
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.Order) error {
	req := &orderV1.CreateOrderRequest{
		UserId:         order.UserID,
		AddressId:      order.AddressID,
		Receiver:       order.Receiver,
		Phone:          order.Phone,
		Address:        order.Address,
		Freight:        order.Freight, // 运费由订单服务端重算，此处传值仅占位
		CouponRecordId: order.CouponRecordID, // 优惠券发放记录ID（0=不使用优惠券）
		IdempotencyKey: order.IdempotencyKey, // 客户端幂等键（防重复下单）
	}
	// 购物车结算：多商品走 items
	if len(order.Items) > 0 {
		for _, item := range order.Items {
			req.Items = append(req.Items, &orderV1.OrderItemProto{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
				SkuId:     item.SkuID,
			})
		}
	} else {
		// 立即购买：单商品
		req.ProductId = order.ProductID
		req.Count = order.Count
		req.Price = order.Price
		req.SkuId = order.SkuID
	}
	res, err := r.data.oc.CreateOrder(ctx, req)
	if err != nil {
		return err
	}
	order.ID = res.OrderId
	return nil
}

func (r *orderRepo) PayOrder(ctx context.Context, orderID int64, userID int64) error {
	_, err := r.data.oc.PayOrder(ctx, &orderV1.PayOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		return err
	}
	return nil
}

// ReceiveOrder 确认收货（调用订单服务，校验订单归属）
func (r *orderRepo) ReceiveOrder(ctx context.Context, orderID int64, userID int64) error {
	_, err := r.data.oc.ReceiveOrder(ctx, &orderV1.ReceiveOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		return err
	}
	return nil
}

// CancelOrder 取消订单（调用订单服务，校验订单归属）
func (r *orderRepo) CancelOrder(ctx context.Context, orderID int64, userID int64, reason string) error {
	_, err := r.data.oc.CancelOrder(ctx, &orderV1.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
		Reason:  reason,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *orderRepo) GetOrder(ctx context.Context, orderID int64) (*biz.Order, error) {
	result, err, _ := r.sg.Do(fmt.Sprintf("get_order_%d", orderID), func() (interface{}, error) {
		lockKey := fmt.Sprintf("lock:order:id:%d", orderID)
		lock := cr.AcquireDistributedLock(ctx, r.data.rc, lockKey, 10*time.Second)
		if !lock {
			r.log.Errorf("Failed to acquire distributed lock for order %d", orderID)
			return nil, fmt.Errorf("failed to acquire distributed lock for order %d", orderID)
		}
		defer func() {
			if err := cr.ReleaseDistributedLock(ctx, r.data.rc, lockKey); err != nil {
				r.log.Errorf("Failed to release distributed lock for order %d: %v", orderID, err)
			}
		}()

		order, err := r.data.oc.GetOrder(ctx, &orderV1.GetOrderRequest{
			OrderId: orderID,
		})
		if err != nil {
			return nil, err
		}

		return &biz.Order{
			ID:        order.OrderId,
			UserID:    order.UserId,
			Status:    order.Status,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*biz.Order), nil
}

// ListOrder 查询当前用户订单列表（含商品明细）
// 两次调用：ListOrder 拿订单列表 + BatchGetOrder 批量拿商品明细，避免 N+1
func (r *orderRepo) ListOrder(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*biz.Order, int32, error) {
	reply, err := r.data.oc.ListOrder(ctx, &orderV1.ListOrderRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		UserId:   userID,
	})
	if err != nil {
		r.log.Errorf("list order failed, userID=%d: %v", userID, err)
		return nil, 0, err
	}

	// 批量获取商品明细
	ids := make([]int64, 0, len(reply.Items))
	for _, item := range reply.Items {
		ids = append(ids, item.OrderId)
	}
	detailMap := make(map[int64]*orderV1.GetOrderReply, len(ids))
	if len(ids) > 0 {
		batch, err := r.data.oc.BatchGetOrder(ctx, &orderV1.BatchGetOrderRequest{OrderIds: ids})
		if err != nil {
			r.log.Errorf("batch get order failed, ids=%v: %v", ids, err)
			return nil, 0, err
		}
		for _, o := range batch.Orders {
			detailMap[o.OrderId] = o
		}
	}

	items := make([]*biz.Order, 0, len(reply.Items))
	for _, item := range reply.Items {
		order := &biz.Order{
			ID:          item.OrderId,
			UserID:      item.UserId,
			OrderNo:     item.OrderNo,
			Status:      item.Status,
			TotalAmount: item.TotalPrice,
			PayAmount:   item.PayAmount,
			Freight:     item.Freight,
			Discount:    item.DiscountAmount,
			CreatedAt:   item.CreatedAt,
			Products:    []*biz.OrderProduct{},
		}
		if detail, ok := detailMap[item.OrderId]; ok {
			for _, p := range detail.Items {
				order.Products = append(order.Products, &biz.OrderProduct{
					ProductID: p.ProductId,
					SkuID:     p.SkuId,
					Name:      p.ProductName,
					Spec:      p.SpecDesc,
					Image:     p.Image,
					Price:     p.Price,
					Quantity:  p.Quantity,
				})
			}
			// 物流信息（发货后）
			order.LogisticsCompany = detail.LogisticsCompany
			order.LogisticsNo = detail.LogisticsNo
			order.Receiver = detail.Receiver
			order.Phone = detail.Phone
			order.Address = detail.Address
		}
		items = append(items, order)
	}
	return items, reply.Total, nil
}

// GetOrderDetail 查询当前用户订单详情（含商品明细）
func (r *orderRepo) GetOrderDetail(ctx context.Context, orderID int64, userID int64) (*biz.Order, error) {
	reply, err := r.data.oc.GetOrder(ctx, &orderV1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		r.log.Errorf("get order detail failed, orderID=%d: %v", orderID, err)
		return nil, err
	}
	// 校验订单归属（防止越权访问他人订单）
	if reply.UserId != userID {
		return nil, v1.ErrorOrderIdNotExist("order not found")
	}
	order := &biz.Order{
		ID:          reply.OrderId,
		UserID:      reply.UserId,
		OrderNo:     reply.OrderNo,
		Status:      reply.Status,
		TotalAmount: reply.TotalPrice,
		PayAmount:   reply.PayAmount,
		Freight:     reply.Freight,
		Discount:    reply.DiscountAmount,
		CreatedAt:   reply.CreatedAt,
		Products:    []*biz.OrderProduct{},
		// 物流信息（发货后）
		LogisticsCompany: reply.LogisticsCompany,
		LogisticsNo:      reply.LogisticsNo,
		Receiver:         reply.Receiver,
		Phone:            reply.Phone,
		Address:          reply.Address,
	}
	for _, p := range reply.Items {
		order.Products = append(order.Products, &biz.OrderProduct{
			ProductID: p.ProductId,
			SkuID:     p.SkuId,
			Name:      p.ProductName,
			Spec:      p.SpecDesc,
			Image:     p.Image,
			Price:     p.Price,
			Quantity:  p.Quantity,
		})
	}
	return order, nil
}

// ApplyRefund 申请退款（调用订单服务，校验订单归属）
func (r *orderRepo) ApplyRefund(ctx context.Context, orderID int64, userID int64, reason string) (int64, error) {
	reply, err := r.data.oc.ApplyRefund(ctx, &orderV1.ApplyRefundRequest{
		OrderId: orderID,
		UserId:  userID,
		Reason:  reason,
	})
	if err != nil {
		r.log.Errorf("apply refund failed, orderID=%d, userID=%d: %v", orderID, userID, err)
		return 0, err
	}
	return reply.RefundId, nil
}

// ListRefund 查询当前用户退款记录（支持按状态筛选）
func (r *orderRepo) ListRefund(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*biz.Refund, int32, error) {
	reply, err := r.data.oc.ListRefund(ctx, &orderV1.ListRefundRequest{
		Page:     page,
		PageSize: pageSize,
		UserId:   userID,
		Status:   status,
	})
	if err != nil {
		r.log.Errorf("list refund failed, userID=%d: %v", userID, err)
		return nil, 0, err
	}
	items := make([]*biz.Refund, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, &biz.Refund{
			ID:           item.RefundId,
			OrderID:      item.OrderId,
			OrderNo:      item.OrderNo,
			RefundAmount: item.RefundAmount,
			Reason:       item.Reason,
			Status:       item.Status,
			StatusText:   item.StatusText,
			RejectReason: item.RejectReason,
			CreatedAt:    item.CreatedAt,
		})
	}
	return items, reply.Total, nil
}

// GetLogistics 查询当前用户订单物流轨迹（先校验订单归属，再调用订单服务）
func (r *orderRepo) GetLogistics(ctx context.Context, orderID int64, userID int64) (*biz.Logistics, error) {
	// 校验订单归属（防止越权查看他人订单物流）
	order, err := r.data.oc.GetOrder(ctx, &orderV1.GetOrderRequest{OrderId: orderID})
	if err != nil {
		r.log.Errorf("get order for logistics failed, orderID=%d: %v", orderID, err)
		return nil, err
	}
	if order.UserId != userID {
		return nil, v1.ErrorOrderIdNotExist("order not found")
	}

	reply, err := r.data.oc.GetLogistics(ctx, &orderV1.GetLogisticsRequest{OrderId: orderID})
	if err != nil {
		r.log.Errorf("get logistics failed, orderID=%d, userID=%d: %v", orderID, userID, err)
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
