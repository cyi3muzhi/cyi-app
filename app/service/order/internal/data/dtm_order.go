package data

import (
	"context"
	"database/sql"
	"errors"

	marketingRpc "cyi-app/api/service/marketing/v1"
	v1 "cyi-app/api/service/order/v1"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	pkgdtm "cyi-app/pkg/dtm"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// DtmOrderStore DTM TCC 订单分支仓储实现：在 DTM 屏障事务（*sql.Tx）内操作订单表
type DtmOrderStore struct {
	mc  marketingRpc.CouponServiceClient
	log *log.Helper
}

// NewDtmOrderStore 创建 DTM TCC 订单分支仓储
func NewDtmOrderStore(mc marketingRpc.CouponServiceClient, logger log.Logger) biz.DtmTCCRepo {
	return &DtmOrderStore{mc: mc, log: log.NewHelper(logger)}
}

// CreateInTx 在 DTM 屏障事务内创建订单（复用 createOrderTx 的建单逻辑）
// 建单失败且已核销优惠券时，尝试对营销侧做补偿释放（与 CreateOrder 语义一致）
func (s *DtmOrderStore) CreateInTx(ctx context.Context, tx *sql.Tx, in *biz.Order) error {
	gdb, err := pkgdtm.GormFromTx(tx)
	if err != nil {
		return err // 事务包装失败：返回错误让 DTM 分支回调映射为 500 重试，避免 panic 崩溃
	}
	couponUsed, err := createOrderTx(ctx, curd.Use(gdb), in, s.mc, s.log)
	if err != nil && couponUsed && in.CouponRecordID > 0 {
		if _, rerr := s.mc.ReleaseCoupon(ctx, &marketingRpc.ReleaseCouponRequest{
			RecordId: in.CouponRecordID,
			UserId:   in.UserID,
			OrderId:  in.ID,
		}); rerr != nil {
			s.log.Errorf("dtm create order release coupon compensation failed, recordID=%d, orderID=%d: %v", in.CouponRecordID, in.ID, rerr)
		} else {
			s.log.Infof("dtm create order release coupon compensation success, recordID=%d, orderID=%d", in.CouponRecordID, in.ID)
		}
	}
	return err
}

// CancelInTx 在 DTM 屏障事务内取消订单（TCC 补偿）：
// 仅待支付订单可取消；已支付/已取消等状态视为已处理（幂等），订单不存在同样视为已处理
func (s *DtmOrderStore) CancelInTx(ctx context.Context, tx *sql.Tx, orderNo string) error {
	gdb, err := pkgdtm.GormFromTx(tx)
	if err != nil {
		return err // 事务包装失败：返回错误让 DTM 分支回调映射为 500 重试，避免 panic 崩溃
	}
	q := curd.Use(gdb)
	order, err := q.Order.WithContext(ctx).Where(q.Order.OrderNo.Eq(orderNo)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 订单不存在：Try 阶段未建单，视为已处理
		}
		return err
	}
	// 非待支付状态（如已支付、已取消）不执行取消，保证幂等
	if order.Status != int32(v1.OrderStatus_WaitPay) {
		return nil
	}
	_, err = q.Order.WithContext(ctx).
		Where(q.Order.ID.Eq(order.ID)).
		Where(q.Order.Status.Eq(int32(v1.OrderStatus_WaitPay))).
		UpdateColumn(q.Order.Status, int32(v1.OrderStatus_Cancel))
	return err
}
