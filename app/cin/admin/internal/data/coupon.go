package data

import (
	"context"

	marketingRpc "cyi-app/api/service/marketing/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type couponRepo struct {
	log  *log.Helper
	data *Data
}

func NewCouponRepo(data *Data, logger log.Logger) biz.CouponRepo {
	return &couponRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

func (r *couponRepo) ListCoupon(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*biz.Coupon, int32, error) {
	reply, err := r.data.mc.ListCoupon(ctx, &marketingRpc.ListCouponRequest{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  keyword,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Coupon, 0, len(reply.Items))
	for _, item := range reply.Items {
		items = append(items, rpcToBizCoupon(item))
	}

	return items, reply.Total, nil
}

func (r *couponRepo) GetCoupon(ctx context.Context, couponID int64) (*biz.Coupon, error) {
	reply, err := r.data.mc.GetCoupon(ctx, &marketingRpc.GetCouponRequest{
		Id: couponID,
	})
	if err != nil {
		return nil, err
	}

	return rpcToBizCoupon(reply.Coupon), nil
}

func (r *couponRepo) CreateCoupon(ctx context.Context, coupon *biz.Coupon) (*biz.Coupon, error) {
	reply, err := r.data.mc.CreateCoupon(ctx, &marketingRpc.CreateCouponRequest{
		Name:        coupon.Name,
		Type:        coupon.Type,
		Value:       coupon.Value,
		MinAmount:   coupon.MinAmount,
		Quantity:    coupon.Quantity,
		ValidStart:  coupon.ValidStart,
		ValidEnd:    coupon.ValidEnd,
	})
	if err != nil {
		return nil, err
	}

	return r.GetCoupon(ctx, reply.CouponId)
}

func (r *couponRepo) UpdateCoupon(ctx context.Context, couponID int64, coupon *biz.Coupon) (*biz.Coupon, error) {
	_, err := r.data.mc.UpdateCoupon(ctx, &marketingRpc.UpdateCouponRequest{
		Id:          couponID,
		Name:        coupon.Name,
		Type:        coupon.Type,
		Value:       coupon.Value,
		MinAmount:   coupon.MinAmount,
		Quantity:    coupon.Quantity,
		ValidStart:  coupon.ValidStart,
		ValidEnd:    coupon.ValidEnd,
		Enabled:     coupon.Enabled,
	})
	if err != nil {
		return nil, err
	}

	return r.GetCoupon(ctx, couponID)
}

func (r *couponRepo) DeleteCoupon(ctx context.Context, couponID int64) error {
	_, err := r.data.mc.DeleteCoupon(ctx, &marketingRpc.DeleteCouponRequest{
		Id: couponID,
	})
	return err
}

func (r *couponRepo) ToggleCouponStatus(ctx context.Context, couponID int64) error {
	_, err := r.data.mc.ToggleCouponStatus(ctx, &marketingRpc.ToggleCouponStatusRequest{
		Id: couponID,
	})
	return err
}

func (r *couponRepo) CouponStats(ctx context.Context) (*biz.CouponStats, error) {
	reply, err := r.data.mc.CouponStats(ctx, &marketingRpc.CouponStatsRequest{})
	if err != nil {
		return nil, err
	}

	return &biz.CouponStats{
		Total:   reply.TotalCount,
		Active:  reply.ActiveCount,
		Issued:  reply.IssuedCount,
		Ended:   reply.ExpiredCount,
	}, nil
}

func rpcToBizCoupon(c *marketingRpc.Coupon) *biz.Coupon {
	return &biz.Coupon{
		ID:           c.Id,
		Name:         c.Name,
		Type:         c.Type,
		Value:        c.Value,
		MinAmount:    c.MinAmount,
		Quantity:     c.Quantity,
		IssuedCount:  c.IssuedCount,
		UsedCount:    c.UsedCount,
		ValidStart:   c.ValidStart,
		ValidEnd:     c.ValidEnd,
		Status:       c.Status,
		Enabled:      c.Enabled,
		CreateTime:   c.CreateTime,
		UpdateTime:   c.UpdateTime,
	}
}