package biz

import (
	"context"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// WaitPayOrder 待付款订单信息（超时扫描用）
type WaitPayOrder struct {
	OrderID   int64 // 订单ID
	CreatedAt int64 // 下单时间（Unix 秒）
}

// OrderTimeoutRepo 订单超时取消仓库接口（由 data 层通过 order gRPC 实现）
type OrderTimeoutRepo interface {
	// GetOrderStatus 查询订单状态（数字状态码，如 1=待付款）
	GetOrderStatus(ctx context.Context, orderID int64) (int32, error)
	// CancelOrder 取消订单（userID=0 表示系统/管理端操作，不校验归属）
	CancelOrder(ctx context.Context, orderID int64, reason string) error
	// ListWaitPayOrders 分页查询待付款订单（status=1），用于超时兜底扫描
	ListWaitPayOrders(ctx context.Context, page, pageSize int32) ([]*WaitPayOrder, error)
}

// OrderTimeoutUseCase 订单超时取消业务用例（消费端）
// 由 asynq 延迟任务 handler 调用：订单仍为待付款则自动取消，否则幂等跳过。
type OrderTimeoutUseCase struct {
	repo OrderTimeoutRepo
	log  *log.Helper
}

// NewOrderTimeoutUseCase 创建订单超时取消业务用例
func NewOrderTimeoutUseCase(repo OrderTimeoutRepo, logger log.Logger) *OrderTimeoutUseCase {
	return &OrderTimeoutUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Process 处理订单超时取消
// 幂等设计：仅当订单仍处于待付款（WaitPay）状态时才自动取消；
// 已支付/已取消/已完成等状态一律跳过，避免误伤正常订单。
func (uc *OrderTimeoutUseCase) Process(ctx context.Context, orderID int64) error {
	status, err := uc.repo.GetOrderStatus(ctx, orderID)
	if err != nil {
		uc.log.Errorf("订单超时任务查询状态失败, orderID: %d: %v", orderID, err)
		return err
	}
	if status != int32(v1.OrderStatus_WaitPay) {
		uc.log.Infof("订单 %d 状态为 %d，非待付款，跳过超时自动取消", orderID, status)
		return nil
	}
	if err := uc.repo.CancelOrder(ctx, orderID, "超时未支付自动取消"); err != nil {
		uc.log.Errorf("订单超时自动取消失败, orderID: %d: %v", orderID, err)
		return err
	}
	uc.log.Infof("订单超时自动取消成功, orderID: %d", orderID)
	return nil
}

// ScanAndCancelOverdue 扫描并取消所有超时未支付的待付款订单（cron 兜底）
// 兜底机制：即使下单时投递的延迟任务丢失，cron 周期扫描也能保证超时订单最终被取消；
// 同时覆盖功能上线前的存量待付款订单。超时筛选已由订单服务数据库下推（秒级精确），
// 单个订单取消失败不中断，记录日志后继续。
// @returns 本次取消的订单数
func (uc *OrderTimeoutUseCase) ScanAndCancelOverdue(ctx context.Context) (int32, error) {
	const pageSize int32 = 100
	cancelled := int32(0)

	for page := int32(1); ; page++ {
		orders, err := uc.repo.ListWaitPayOrders(ctx, page, pageSize)
		if err != nil {
			return cancelled, err
		}
		if len(orders) == 0 {
			break
		}
		for _, o := range orders {
			if err := uc.repo.CancelOrder(ctx, o.OrderID, "超时未支付自动取消"); err != nil {
				uc.log.Errorf("订单超时扫描取消失败, orderID: %d: %v", o.OrderID, err)
				continue
			}
			cancelled++
			uc.log.Infof("订单超时扫描自动取消成功, orderID: %d", o.OrderID)
		}
		// 不足一页说明已拉完
		if int32(len(orders)) < pageSize {
			break
		}
	}
	return cancelled, nil
}
