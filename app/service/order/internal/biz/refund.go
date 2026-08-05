package biz

import (
	"context"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Refund 退款单领域模型（仅退款售后）
type Refund struct {
	ID           int64
	OrderID      int64
	UserID       int64
	OrderNo      string
	RefundAmount Money // 退款金额（单位：分，即订单实付金额）
	Reason       string
	Status       int32 // 1-待审核 2-已通过 3-已拒绝（对应 v1.RefundStatus）
	OriginStatus int32 // 申请退款前的订单状态（审核拒绝后恢复用）
	RejectReason string
	Auditor      string // 审核人名称
	AuditTime    int64  // 审核时间（Unix 秒）
	CreateTime   int64  // 申请时间（Unix 秒）
}

// RefundRepo 退款单数据仓库接口
type RefundRepo interface {
	// CreateRefund 创建退款单（申请退款时）
	CreateRefund(ctx context.Context, refund *Refund) error
	// GetRefund 查询退款单
	GetRefund(ctx context.Context, refundID int64) (*Refund, error)
	// UpdateRefundAudit 审核后更新退款单（状态 + 审核信息）
	UpdateRefundAudit(ctx context.Context, refundID int64, status int32, rejectReason, auditor string) error
	// ListRefund 分页查询退款单（支持按状态筛选、订单号搜索、用户过滤）
	ListRefund(ctx context.Context, page, pageSize int32, status string, keyword string, userID int64) ([]*Refund, int32, error)
}

// RefundUseCase 退款业务用例（仅退款售后闭环：用户申请 -> 管理员审核）
type RefundUseCase struct {
	refundRepo RefundRepo
	orderUC    *OrderUseCase
	log        *log.Helper
}

// NewRefundUseCase 创建退款业务用例
func NewRefundUseCase(refundRepo RefundRepo, orderUC *OrderUseCase, logger log.Logger) *RefundUseCase {
	return &RefundUseCase{
		refundRepo: refundRepo,
		orderUC:    orderUC,
		log:        log.NewHelper(logger),
	}
}

// ApplyRefund 用户申请退款（仅已支付/待发货订单可申请）
// 流程：校验归属与状态 -> 订单进入退款中 -> 创建退款单（待审核）
func (uc *RefundUseCase) ApplyRefund(ctx context.Context, orderID, userID int64, reason string) (*Refund, error) {
	// 1. 查询订单并加载状态处理器
	order, err := uc.orderUC.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}
	// 2. 校验订单归属
	if err := order.VerifyOwner(userID); err != nil {
		return nil, err
	}
	// 3. 校验订单状态：仅已支付(2)/待发货(3)可申请退款
	if order.GetStatus() != v1.OrderStatus_Pay && order.GetStatus() != v1.OrderStatus_WaitShip {
		return nil, v1.ErrorOrderStatusInvalid("当前订单状态不可申请退款, orderID=%d", orderID)
	}
	originStatus := int32(order.GetStatus())

	// 4. 订单状态流转：已支付/待发货 -> 退款中（状态机校验并落库）
	if err := order.Refunding(ctx); err != nil {
		uc.log.Errorf("order enter refunding failed, orderID: %d: %v", orderID, err)
		return nil, err
	}

	// 5. 创建退款单（待审核）
	refund := &Refund{
		OrderID:      orderID,
		UserID:       userID,
		OrderNo:      order.OrderNo,
		RefundAmount: order.PayAmount,
		Reason:       reason,
		Status:       int32(v1.RefundStatus_RefundPending),
		OriginStatus: originStatus,
	}
	if err := uc.refundRepo.CreateRefund(ctx, refund); err != nil {
		// 创建退款单失败，回滚订单状态到申请前状态，避免订单卡在退款中
		uc.log.Errorf("create refund failed, orderID: %d: %v, rolling back order status", orderID, err)
		if rbErr := uc.orderUC.repo.UpdateStatusIf(ctx, orderID, int32(v1.OrderStatus_Refunding), originStatus); rbErr != nil {
			uc.log.Errorf("rollback order status failed, orderID: %d: %v", orderID, rbErr)
		}
		return nil, err
	}
	return refund, nil
}

// AuditRefund 管理端审核退款单
// 通过：退款中 -> 已退款（恢复库存）；拒绝：退款中 -> 恢复申请前状态
func (uc *RefundUseCase) AuditRefund(ctx context.Context, refundID int64, approve bool, rejectReason, auditor string) error {
	// 1. 查询退款单，校验待审核状态
	refund, err := uc.refundRepo.GetRefund(ctx, refundID)
	if err != nil {
		return err
	}
	if refund.Status != int32(v1.RefundStatus_RefundPending) {
		return v1.ErrorOrderStatusInvalid("退款单已审核, refundID=%d", refundID)
	}

	// 2. 查询订单并加载状态处理器
	order, err := uc.orderUC.Get(ctx, refund.OrderID)
	if err != nil {
		return err
	}
	// 3. 校验订单仍处于退款中（防止并发审核/订单状态被其他操作改变）
	if order.GetStatus() != v1.OrderStatus_Refunding {
		return v1.ErrorOrderStatusInvalid("订单状态已变更，无法审核退款, orderID=%d", refund.OrderID)
	}

	if approve {
		// 审核通过：退款完成（恢复库存 + 状态 -> 已退款）
		if err := order.Refunded(ctx); err != nil {
			uc.log.Errorf("order refund failed, orderID: %d: %v", refund.OrderID, err)
			return err
		}
	} else {
		// 审核拒绝：恢复订单到申请前状态（退款中不涉及库存变动，无需补偿库存）
		if err := uc.orderUC.repo.UpdateStatusIf(ctx, refund.OrderID, int32(v1.OrderStatus_Refunding), refund.OriginStatus); err != nil {
			uc.log.Errorf("restore order status failed, orderID: %d: %v", refund.OrderID, err)
			return err
		}
	}

	// 4. 更新退款单审核结果
	status := int32(v1.RefundStatus_RefundApproved)
	if !approve {
		status = int32(v1.RefundStatus_RefundRejected)
	}
	if err := uc.refundRepo.UpdateRefundAudit(ctx, refundID, status, rejectReason, auditor); err != nil {
		uc.log.Errorf("update refund audit failed, refundID: %d: %v", refundID, err)
		return err
	}
	return nil
}

// ListRefund 分页查询退款单（管理端全量 / 用户端按 userID 过滤）
func (uc *RefundUseCase) ListRefund(ctx context.Context, page, pageSize int32, status string, keyword string, userID int64) ([]*Refund, int32, error) {
	return uc.refundRepo.ListRefund(ctx, page, pageSize, status, keyword, userID)
}
