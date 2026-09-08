package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// statusBase 状态处理器基础结构，提供统一的状态转换方法
type statusBase struct {
	repo    OrderRepo
	stockRe StockRepo
	log     *log.Helper
}

// transitionTo 统一的状态转换方法：先条件更新数据库（乐观并发控制），成功后才切换状态处理器
// 使用 UpdateStatusIf 保证「仅当当前状态仍为期望状态时才落库」，避免支付/取消并发时互相覆盖导致库存泄漏
func (b *statusBase) transitionTo(ctx context.Context, order *Order, newStatus v1.OrderStatus) error {
	oldStatus := order.Status
	if err := b.repo.UpdateStatusIf(ctx, order.ID, int32(oldStatus), int32(newStatus)); err != nil {
		b.log.WithContext(ctx).Errorf("订单状态更新失败, orderID: %d, oldStatus: %v, newStatus: %v, error: %v", order.ID, oldStatus, newStatus, err)
		return err
	}
	order.SetStatus(newStatus, b.repo, b.stockRe, b.log.Logger()) // DB 成功，切换状态处理器
	return nil
}

// reconcileCancelStock 取消/退款终态库存对账（幂等，可安全重复/并发执行）
// 对订单每个 SKU 执行两个方向的逆操作，由 product 侧屏障分别裁决是否真正生效：
//  1. ReleaseStock(tcc-{orderNo}, cancel)：逆「下单预留」——未支付预留归还可售；
//     预留已被支付扣减消耗时 SQL 条件（reserved>=N）不命中，为幂等空操作
//  2. AddStock(pay-{orderNo}, cancel)：逆「支付扣减」——已支付扣减退还可售；
//     扣减从未真实执行时屏障空补偿自动跳过，不会多加库存
//
// 因此调用方无需判断订单当前处于「预留」还是「已扣减」状态：
// 未支付取消、已支付取消、退款完成、支付失败对账均可统一调用，重复调用结果一致
func reconcileCancelStock(ctx context.Context, stockRe StockRepo, logger *log.Helper, order *Order) error {
	for _, item := range order.Items {
		releaseKey := tccStockKey(order.OrderNo, item.SkuID, "cancel")
		if err := callStockOpWithRetry(ctx, logger, "逆预留", releaseKey, item.SkuID, item.Quantity, stockRe.ReleaseStock); err != nil {
			logger.WithContext(ctx).Errorf("取消对账释放预留库存失败, orderNo: %s, skuID: %d, error: %v", order.OrderNo, item.SkuID, err)
			return err
		}
		addKey := payStockKey(order.OrderNo, item.SkuID, "cancel")
		if err := callStockOpWithRetry(ctx, logger, "逆扣减", addKey, item.SkuID, item.Quantity, stockRe.AddStock); err != nil {
			logger.WithContext(ctx).Errorf("取消对账退还扣减库存失败, orderNo: %s, skuID: %d, error: %v", order.OrderNo, item.SkuID, err)
			return err
		}
	}
	return nil
}

// stockOpFunc 库存操作函数签名（与 StockRepo 的四个库存操作一致）
type stockOpFunc func(ctx context.Context, key StockOpKey, skuID int64, count int32) error

// callStockOpWithRetry 带 3 次重试的库存操作调用
// 屏障幂等保证重试安全：同一 (gid,branch,op) 重复请求不会产生重复库存变更
func callStockOpWithRetry(ctx context.Context, logger *log.Helper, opName string, key StockOpKey, skuID int64, count int32, fn stockOpFunc) error {
	var err error
	for retry := 0; retry < 3; retry++ {
		if err = fn(ctx, key, skuID, count); err == nil {
			return nil
		}
		logger.WithContext(ctx).Errorf("%s库存操作失败, gid: %s, skuID: %d, 第 %d 次重试, error: %v", opName, key.Gid, skuID, retry+1, err)
	}
	return err
}

// 以下是 OrderStatus 接口的默认实现，具体状态处理器可重写允许的操作

// Pay 默认支付处理
func (b *statusBase) Pay(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "支付")
}

// Ship 默认发货处理
func (b *statusBase) Ship(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "发货")
}

// Receive 默认收货处理
func (b *statusBase) Receive(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "收货")
}

// Complete 默认完成处理
func (b *statusBase) Complete(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "完成")
}

// Cancel 默认取消处理
func (b *statusBase) Cancel(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "取消")
}

// Refunding 默认退款中处理
func (b *statusBase) Refunding(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "退款")
}

// Refunded 默认已退款处理
func (b *statusBase) Refunded(ctx context.Context, order *Order) error {
	return b.invalidOperationError(order, "退款")
}

// invalidOperationError 生成统一的无效操作错误消息
func (b *statusBase) invalidOperationError(order *Order, operation string) error {
	currentStatus := getStatusName(order.Status)
	return v1.ErrorOrderStatusInvalid("订单%s，无法执行%s操作, orderID=%d", currentStatus, operation, order.ID)
}

// getStatusName 获取状态的中文描述
func getStatusName(status v1.OrderStatus) string {
	switch status {
	case v1.OrderStatus_WaitPay:
		return "未支付"
	case v1.OrderStatus_Pay, v1.OrderStatus_WaitShip:
		return "已支付"
	case v1.OrderStatus_Ship, v1.OrderStatus_WaitReceive:
		return "已发货"
	case v1.OrderStatus_Complete:
		return "已完成"
	case v1.OrderStatus_Cancel:
		return "已取消"
	case v1.OrderStatus_Refunding:
		return "退款中"
	case v1.OrderStatus_Refunded:
		return "已退款"
	default:
		return "状态异常"
	}
}

// 状态转换矩阵（✓ = 允许转换，✗ = 拒绝操作）：
//
//              Pay  Ship  Receive  Complete  Cancel  Refunding  Refunded
// Unspecified   ✗    ✗      ✗        ✗        ✗       ✗          ✗
// WaitPay       ✓    ✗      ✗        ✗        ✓       ✗          ✗
// Pay/WaitShip  ✗    ✓      ✗        ✗        ✓       ✓          ✗
// Ship/WaitRec  ✗    ✗      ✓        ✗        ✗       ✓          ✗
// Complete      ✗    ✗      ✗        ✗        ✗       ✓          ✗
// Cancel        ✗    ✗      ✗        ✗        ✗       ✗          ✗
// Refunding     ✗    ✗      ✗        ✗        ✗       ✗          ✓
// Refunded      ✗    ✗      ✗        ✗        ✗       ✗          ✗

// OrderStatus 订单状态接口，所有方法统一接收 order 参数，消除类型断言依赖
type OrderStatus interface {
	// 支付
	Pay(ctx context.Context, order *Order) error
	// 发货
	Ship(ctx context.Context, order *Order) error
	// 收货
	Receive(ctx context.Context, order *Order) error
	// 完成
	Complete(ctx context.Context, order *Order) error
	// 取消
	Cancel(ctx context.Context, order *Order) error
	// 退款中
	Refunding(ctx context.Context, order *Order) error
	// 已退款
	Refunded(ctx context.Context, order *Order) error
}

// NewOrderStatusHandler 根据状态创建对应的订单状态处理器
func NewOrderStatusHandler(status v1.OrderStatus, repo OrderRepo, stockRe StockRepo, logger log.Logger) OrderStatus {
	switch status {
	case v1.OrderStatus_WaitPay:
		return NewWaitPayStatus(repo, stockRe, logger)
	case v1.OrderStatus_Pay:
		return NewPayStatus(repo, stockRe, logger)
	case v1.OrderStatus_WaitShip:
		return NewPayStatus(repo, stockRe, logger) // 待发货状态使用已支付状态的处理逻辑
	case v1.OrderStatus_Ship:
		return NewShipStatus(repo, stockRe, logger)
	case v1.OrderStatus_WaitReceive:
		return NewShipStatus(repo, stockRe, logger) // 待收货状态使用已发货状态的处理逻辑
	case v1.OrderStatus_Complete:
		return NewCompleteStatus(repo, stockRe, logger)
	case v1.OrderStatus_Cancel:
		return NewFinalStatus(repo, stockRe, logger)
	case v1.OrderStatus_Refunding:
		return NewRefundingStatus(repo, stockRe, logger)
	case v1.OrderStatus_Refunded:
		return NewFinalStatus(repo, stockRe, logger)
	default:
		return nil
	}
}

// GetStatus 获取订单状态
func (o *Order) GetStatus() v1.OrderStatus {
	return o.Status
}

// SetStatus 设置状态并更新状态处理器
func (o *Order) SetStatus(newStatus v1.OrderStatus, repo OrderRepo, stockRe StockRepo, logger log.Logger) {
	o.Status = newStatus
	o.statusHandler = NewOrderStatusHandler(newStatus, repo, stockRe, logger)
}

// Pay 支付订单
func (o *Order) Pay(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Pay(ctx, o)
}

// Ship 发货
func (o *Order) Ship(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Ship(ctx, o)
}

// Receive 收货
func (o *Order) Receive(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Receive(ctx, o)
}

// Complete 完成
func (o *Order) Complete(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Complete(ctx, o)
}

// Cancel 取消
func (o *Order) Cancel(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Cancel(ctx, o)
}

// Refunding 退款中
func (o *Order) Refunding(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Refunding(ctx, o)
}

// Refunded 已退款
func (o *Order) Refunded(ctx context.Context) error {
	if o.statusHandler == nil {
		return v1.ErrorOrderStatusInvalid("订单状态无效，无法执行该操作, orderID=%d", o.ID)
	}
	return o.statusHandler.Refunded(ctx, o)
}
