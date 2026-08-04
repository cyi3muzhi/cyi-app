package data

import (
	"context"

	"cyi-app/app/cin/interface/internal/biz"
	marketingV1 "cyi-app/api/service/marketing/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// couponRepo 优惠券数据层实现
type couponRepo struct {
	log  *log.Helper
	data *Data
}

// NewCouponRepo 创建优惠券仓储
func NewCouponRepo(data *Data, logger log.Logger) biz.CouponRepo {
	return &couponRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// ClaimCoupon 用户领取优惠券
func (r *couponRepo) ClaimCoupon(ctx context.Context, couponID int64, userID int64, source string) (int64, error) {
	req := &marketingV1.ClaimCouponRequest{
		CouponId: couponID,
		UserId:   userID,
		Source:   source,
	}
	res, err := r.data.mcc.ClaimCoupon(ctx, req)
	if err != nil {
		return 0, err
	}
	return res.RecordId, nil
}

// ListUserCoupons 查询用户优惠券列表
func (r *couponRepo) ListUserCoupons(ctx context.Context, userID int64, status int32, page int32, pageSize int32) ([]*biz.CouponItem, int32, error) {
	req := &marketingV1.ListUserCouponsRequest{
		UserId:   userID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}
	res, err := r.data.mcc.ListUserCoupons(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.CouponItem, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, &biz.CouponItem{
			RecordID:     item.RecordId,
			CouponID:     item.CouponId,
			Name:         item.Name,
			Type:         item.Type,
			Value:        item.Value,
			MinAmount:    item.MinAmount,
			ValidStart:   item.ValidStart,
			ValidEnd:     item.ValidEnd,
			RecordStatus: item.RecordStatus,
			UsedTime:     item.UsedTime,
			OrderID:      item.OrderId,
			Source:       item.Source,
			CreateTime:   item.CreateTime,
		})
	}
	return items, res.Total, nil
}

// ListAvailableCoupons 查询可领取优惠券列表（领券中心，经 marketing gRPC）
func (r *couponRepo) ListAvailableCoupons(ctx context.Context, userID int64, page int32, pageSize int32) ([]*biz.AvailableCoupon, int32, error) {
	req := &marketingV1.ListAvailableCouponsRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
	}
	res, err := r.data.mcc.ListAvailableCoupons(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.AvailableCoupon, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, &biz.AvailableCoupon{
			CouponID:     item.CouponId,
			Name:         item.Name,
			Type:         item.Type,
			Value:        item.Value,
			MinAmount:    item.MinAmount,
			ValidType:    item.ValidType,
			ValidDays:    item.ValidDays,
			ValidStart:   item.ValidStart,
			ValidEnd:     item.ValidEnd,
			RemainCount:  item.RemainCount,
			Claimed:      item.Claimed,
			LimitPerUser: item.LimitPerUser,
		})
	}
	return items, res.Total, nil
}
