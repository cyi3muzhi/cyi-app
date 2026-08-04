package biz

import "context"

// CouponRepo 优惠券数据仓库接口
type CouponRepo interface {
	// BatchExpireCoupons 批量过期优惠券发放记录
	// batchSize 每次处理的记录数
	// 返回已过期的记录数
	BatchExpireCoupons(ctx context.Context, batchSize int64) (int32, error)
}
