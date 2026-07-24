package biz

const (
	CouponTypeCash      = int32(1)
	CouponTypeDiscount  = int32(2)
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
	CreateTime   int64
	UpdateTime   int64
}

type CouponStats struct {
	TotalCount   int32
	ActiveCount  int32
	IssuedCount  int64
	ExpiredCount int32
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
		return nil
	}
	if c.Type != CouponTypeCash && c.Type != CouponTypeDiscount && c.Type != CouponTypeFullReduce {
		return nil
	}
	if c.Value <= 0 {
		return nil
	}
	if c.MinAmount < 0 {
		return nil
	}
	if c.Quantity <= 0 {
		return nil
	}
	if c.ValidStart > c.ValidEnd {
		return nil
	}
	return nil
}