package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// CouponItem 用户优惠券记录
type CouponItem struct {
	RecordID     int64  // 发放记录 ID
	CouponID     int64  // 优惠券 ID
	Name         string // 优惠券名称
	Type         int32  // 类型
	Value        int64  // 面值
	MinAmount    int64  // 使用门槛
	ValidStart   int64  // 有效期开始（时间戳）
	ValidEnd     int64  // 有效期结束（时间戳）
	RecordStatus int32  // 发放记录状态：0=未使用 1=已使用 2=已过期
	UsedTime     int64  // 使用时间（时间戳）
	OrderID      int64  // 核销订单 ID
	Source       string // 领取来源
	CreateTime   int64  // 领取时间（时间戳）
}

// CouponRepo 优惠券仓储接口
type CouponRepo interface {
	// ClaimCoupon 用户领取优惠券
	ClaimCoupon(ctx context.Context, couponID int64, userID int64, source string) (int64, error)
	// ListUserCoupons 查询用户优惠券列表
	ListUserCoupons(ctx context.Context, userID int64, status int32, page int32, pageSize int32) ([]*CouponItem, int32, error)
}

// CouponUseCase 优惠券业务逻辑
type CouponUseCase struct {
	log  *log.Helper
	repo CouponRepo
}

// NewCouponUseCase 创建优惠券业务逻辑
func NewCouponUseCase(repo CouponRepo, logger log.Logger) *CouponUseCase {
	return &CouponUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// ClaimCoupon 用户领取优惠券
func (uc *CouponUseCase) ClaimCoupon(ctx context.Context, couponID int64, userID int64, source string) (int64, error) {
	return uc.repo.ClaimCoupon(ctx, couponID, userID, source)
}

// ListMyCoupons 查询用户优惠券列表
func (uc *CouponUseCase) ListMyCoupons(ctx context.Context, userID int64, status int32, page int32, pageSize int32) ([]*CouponItem, int32, error) {
	return uc.repo.ListUserCoupons(ctx, userID, status, page, pageSize)
}
