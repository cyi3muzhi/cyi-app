package biz

import "fmt"

const (
	CouponTypeCash       = int32(1)
	CouponTypeDiscount   = int32(2)
	CouponTypeFullReduce = int32(3)

	CouponStatusActive  = int32(1)
	CouponStatusPending = int32(2)
	CouponStatusEnded   = int32(3)
	CouponStatusVoided  = int32(4)
)

type Coupon struct {
	ID           int64
	Name         string
	Type         int32
	Value        int64
	MinAmount    int64
	Quantity     int32
	IssuedCount  int32
	UsedCount    int32
	ValidStart   int64
	ValidEnd     int64
	Status       int32
	Enabled      bool
	LimitPerUser int32
	CreateTime   int64
	UpdateTime   int64
}

type CouponStats struct {
	TotalCount   int32
	ActiveCount  int32
	IssuedCount  int64
	ExpiredCount int32
}

// DistRecord 发放记录
type DistRecord struct {
	ID         int64
	UserID     int64
	Source     string
	Status     int32
	UsedTime   int64
	OrderID    int64
	CreateTime int64
}

// UserCouponItem 用户领取的优惠券记录（含优惠券基本信息）
type UserCouponItem struct {
	RecordID     int64
	CouponID     int64
	Name         string
	Type         int32
	Value        int64
	MinAmount    int64
	ValidStart   int64
	ValidEnd     int64
	RecordStatus int32
	UsedTime     int64
	OrderID      int64
	Source       string
	CreateTime   int64
}

func (c *Coupon) ComputeStatus(now int64) int32 {
	if c.Status == CouponStatusVoided {
		return CouponStatusVoided
	}
	if now < c.ValidStart {
		return CouponStatusPending
	}
	if now > c.ValidEnd {
		return CouponStatusEnded
	}
	return CouponStatusActive
}

func (c *Coupon) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("coupon name is required")
	}
	if c.Type != CouponTypeCash && c.Type != CouponTypeDiscount && c.Type != CouponTypeFullReduce {
		return fmt.Errorf("invalid coupon type: %d", c.Type)
	}
	if c.Value <= 0 {
		return fmt.Errorf("coupon value must be positive")
	}
	if c.MinAmount < 0 {
		return fmt.Errorf("min_amount must not be negative")
	}
	if c.Quantity <= 0 {
		return fmt.Errorf("coupon quantity must be positive")
	}
	if c.ValidStart > c.ValidEnd {
		return fmt.Errorf("valid_start must be before valid_end")
	}
	return nil
}
