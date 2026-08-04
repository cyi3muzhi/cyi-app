package biz

import (
	"context"
	"time"

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

// AvailableCoupon 可领取优惠券（领券中心展示）
type AvailableCoupon struct {
	CouponID     int64
	Name         string
	Type         int32  // 1-立减 2-折扣 3-满减
	Value        int64  // 面值（分）或折扣百分比（1-99）
	MinAmount    int64  // 使用门槛（分）
	ValidType    int32  // 有效期类型：1-固定日期 2-领取后N天 3-永久
	ValidDays    int32  // 领取后有效天数
	ValidStart   int64  // 有效期开始（Unix 秒）
	ValidEnd     int64  // 有效期结束（Unix 秒），0=永久
	RemainCount  int32  // 剩余可领数量
	Claimed      bool   // 当前用户是否已领取
	LimitPerUser int32  // 每人限领数量
}

// UsableCoupon 结算可用优惠券（含不可用原因）
type UsableCoupon struct {
	RecordID  int64
	CouponID  int64
	Name      string
	Type      int32
	Value     int64
	MinAmount int64
	ValidEnd  int64
	Usable    bool  // 是否可用（满足门槛且在有效期内）
	Discount  int64 // 可用时预估优惠金额（分）
	GapAmount int64 // 不可用时还差多少金额可用（分）
}

// CouponRepo 优惠券仓储接口
type CouponRepo interface {
	// ClaimCoupon 用户领取优惠券
	ClaimCoupon(ctx context.Context, couponID int64, userID int64, source string) (int64, error)
	// ListUserCoupons 查询用户优惠券列表
	ListUserCoupons(ctx context.Context, userID int64, status int32, page int32, pageSize int32) ([]*CouponItem, int32, error)
	// ListAvailableCoupons 查询可领取优惠券列表（领券中心）
	ListAvailableCoupons(ctx context.Context, userID int64, page int32, pageSize int32) ([]*AvailableCoupon, int32, error)
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

// ListAvailableCoupons 领券中心：可领取优惠券列表
func (uc *CouponUseCase) ListAvailableCoupons(ctx context.Context, userID int64, page int32, pageSize int32) ([]*AvailableCoupon, int32, error) {
	return uc.repo.ListAvailableCoupons(ctx, userID, page, pageSize)
}

// ListUsableCoupons 结算可用优惠券：按订单金额筛选用户未使用且有效的券
// amount 为订单商品金额（分）；券未过期且满足门槛即可用，否则标记不可用并给出还差金额
func (uc *CouponUseCase) ListUsableCoupons(ctx context.Context, userID int64, amount int64) ([]*UsableCoupon, error) {
	items, _, err := uc.repo.ListUserCoupons(ctx, userID, 1, 1, 100) // status=1 未使用
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	result := make([]*UsableCoupon, 0, len(items))
	for _, it := range items {
		usable := true
		var gap int64
		// 有效期判断：valid_end=0 表示永久有效
		if it.ValidEnd > 0 && now > it.ValidEnd {
			usable = false
		}
		// 门槛判断：订单金额需达到使用门槛
		if usable && it.MinAmount > amount {
			usable = false
			gap = it.MinAmount - amount
		}
		var discount int64
		if usable {
			discount = calculateCouponDiscount(it.Type, it.Value, amount)
		}
		result = append(result, &UsableCoupon{
			RecordID:  it.RecordID,
			CouponID:  it.CouponID,
			Name:      it.Name,
			Type:      it.Type,
			Value:     it.Value,
			MinAmount: it.MinAmount,
			ValidEnd:  it.ValidEnd,
			Usable:    usable,
			Discount:  discount,
			GapAmount: gap,
		})
	}
	return result, nil
}

// calculateCouponDiscount 计算优惠金额（与营销服务 calculateDiscount 口径一致）
// type：1-立减（优惠=min(面值,金额)） 2-折扣（value 为折扣百分比 1-99） 3-满减（同立减）
func calculateCouponDiscount(couponType int32, value int64, orderAmount int64) int64 {
	switch couponType {
	case 1, 3: // 立减 / 满减
		if value > orderAmount {
			return orderAmount
		}
		return value
	case 2: // 折扣：优惠 = 金额 * (100 - value) / 100
		return orderAmount * (100 - value) / 100
	}
	return 0
}
