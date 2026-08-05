package data

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	v1 "cyi-app/api/service/order/v1"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
)

// refundRepo 退款单数据仓库实现
type refundRepo struct {
	data *Data
	log  *log.Helper
}

// NewRefundRepo 创建退款单数据仓库
func NewRefundRepo(data *Data, logger log.Logger) biz.RefundRepo {
	return &refundRepo{data: data, log: log.NewHelper(logger)}
}

// CreateRefund 创建退款单（申请退款时写入）
func (r *refundRepo) CreateRefund(ctx context.Context, refund *biz.Refund) error {
	m := &mod.OrderRefund{
		OrderID:      refund.OrderID,
		UserID:       refund.UserID,
		OrderNo:      refund.OrderNo,
		RefundAmount: refund.RefundAmount.Amount,
		Reason:       refund.Reason,
		Status:       refund.Status,
		OriginStatus: refund.OriginStatus,
	}
	if err := curd.OrderRefund.WithContext(ctx).Create(m); err != nil {
		r.log.Errorf("create refund failed, orderID=%d: %v", refund.OrderID, err)
		return errors.New("create refund failed")
	}
	refund.ID = m.ID
	refund.CreateTime = m.CreateTime
	return nil
}

// GetRefund 查询退款单，不存在时返回 REFUND_NOT_FOUND 错误
func (r *refundRepo) GetRefund(ctx context.Context, refundID int64) (*biz.Refund, error) {
	m, err := curd.OrderRefund.WithContext(ctx).Where(curd.OrderRefund.ID.Eq(refundID)).First()
	if err != nil {
		r.log.Errorf("get refund failed, refundID=%d: %v", refundID, err)
		return nil, v1.ErrorRefundNotFound("refund not found, refundID=%d", refundID)
	}
	if m == nil || m.ID == 0 {
		return nil, v1.ErrorRefundNotFound("refund not found, refundID=%d", refundID)
	}
	return refundToBiz(m), nil
}

// UpdateRefundAudit 审核后更新退款单（状态 + 审核信息 + 审核时间）
func (r *refundRepo) UpdateRefundAudit(ctx context.Context, refundID int64, status int32, rejectReason, auditor string) error {
	if _, err := curd.OrderRefund.WithContext(ctx).
		Where(curd.OrderRefund.ID.Eq(refundID)).
		Updates(&mod.OrderRefund{
			Status:       status,
			RejectReason: rejectReason,
			Auditor:      auditor,
			AuditTime:    time.Now().Unix(),
		}); err != nil {
		r.log.Errorf("update refund audit failed, refundID=%d: %v", refundID, err)
		return errors.New("update refund audit failed")
	}
	return nil
}

// ListRefund 分页查询退款单，支持按状态筛选（逗号分隔）、订单号模糊搜索、按用户过滤
func (r *refundRepo) ListRefund(ctx context.Context, page, pageSize int32, status string, keyword string, userID int64) ([]*biz.Refund, int32, error) {
	do := curd.OrderRefund.WithContext(ctx)

	// 按用户过滤（>0 时生效，用户端个人中心使用）
	if userID > 0 {
		do = do.Where(curd.OrderRefund.UserID.Eq(userID))
	}

	// 按状态筛选（支持逗号分隔多状态，如 "1,2"）
	if status != "" {
		statuses := strings.Split(status, ",")
		var statusInts []int32
		for _, s := range statuses {
			if v, err := strconv.ParseInt(s, 10, 32); err == nil {
				statusInts = append(statusInts, int32(v))
			}
		}
		if len(statusInts) > 0 {
			if len(statusInts) == 1 {
				do = do.Where(curd.OrderRefund.Status.Eq(statusInts[0]))
			} else {
				do = do.Where(curd.OrderRefund.Status.In(statusInts...))
			}
		}
	}

	// 按订单号模糊搜索
	if keyword != "" {
		do = do.Where(curd.OrderRefund.OrderNo.Like("%" + keyword + "%"))
	}

	// 查询总数
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询，按ID降序
	offset := int((page - 1) * pageSize)
	items, err := do.Offset(offset).Limit(int(pageSize)).Order(curd.OrderRefund.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.Refund, 0, len(items))
	for _, m := range items {
		result = append(result, refundToBiz(m))
	}
	return result, int32(total), nil
}

// refundToBiz 数据库模型 -> 领域模型转换
func refundToBiz(m *mod.OrderRefund) *biz.Refund {
	return &biz.Refund{
		ID:           m.ID,
		OrderID:      m.OrderID,
		UserID:       m.UserID,
		OrderNo:      m.OrderNo,
		RefundAmount: biz.Money{Amount: m.RefundAmount, Currency: "CNY"},
		Reason:       m.Reason,
		Status:       m.Status,
		OriginStatus: m.OriginStatus,
		RejectReason: m.RejectReason,
		Auditor:      m.Auditor,
		AuditTime:    m.AuditTime,
		CreateTime:   m.CreateTime,
	}
}
