package data

import (
	"context"

	marketingV1 "cyi-app/api/service/marketing/v1"
	"cyi-app/app/task/internal/biz"
)

// couponRepo 优惠券数据仓库实现
type couponRepo struct {
	data *Data
}

// NewCouponRepo 创建优惠券数据仓库
func NewCouponRepo(data *Data) biz.CouponRepo {
	return &couponRepo{data: data}
}

// BatchExpireCoupons 调用 marketing 服务批量过期优惠券发放记录
func (r *couponRepo) BatchExpireCoupons(ctx context.Context, batchSize int64) (int32, error) {
	res, err := r.data.couponC.BatchExpireCoupons(ctx, &marketingV1.BatchExpireCouponsRequest{
		BatchSize: batchSize,
	})
	if err != nil {
		return 0, err
	}
	return res.AffectedCount, nil
}
