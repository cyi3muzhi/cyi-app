package service

import (
	"context"

	v1 "cyi-app/api/service/marketing/v1"
	"cyi-app/app/service/marketing/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type CouponService struct {
	v1.UnimplementedCouponServiceServer

	uc  *biz.CouponUseCase
	log *log.Helper
}

func NewCouponService(uc *biz.CouponUseCase, logger log.Logger) *CouponService {
	return &CouponService{uc: uc, log: log.NewHelper(logger)}
}

func (s *CouponService) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponReply, error) {
	return s.uc.CreateCoupon(ctx, req)
}

func (s *CouponService) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponReply, error) {
	return s.uc.UpdateCoupon(ctx, req)
}

func (s *CouponService) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponReply, error) {
	return s.uc.DeleteCoupon(ctx, req)
}

func (s *CouponService) GetCoupon(ctx context.Context, req *v1.GetCouponRequest) (*v1.GetCouponReply, error) {
	return s.uc.GetCoupon(ctx, req)
}

func (s *CouponService) ListCoupon(ctx context.Context, req *v1.ListCouponRequest) (*v1.ListCouponReply, error) {
	return s.uc.ListCoupon(ctx, req)
}

func (s *CouponService) ToggleCouponStatus(ctx context.Context, req *v1.ToggleCouponStatusRequest) (*v1.ToggleCouponStatusReply, error) {
	return s.uc.ToggleCouponStatus(ctx, req)
}

func (s *CouponService) CouponStats(ctx context.Context, req *v1.CouponStatsRequest) (*v1.CouponStatsReply, error) {
	return s.uc.CouponStats(ctx, req)
}

func (s *CouponService) ClaimCoupon(ctx context.Context, req *v1.ClaimCouponRequest) (*v1.ClaimCouponReply, error) {
	return s.uc.ClaimCoupon(ctx, req)
}

func (s *CouponService) ListUserCoupons(ctx context.Context, req *v1.ListUserCouponsRequest) (*v1.ListUserCouponsReply, error) {
	return s.uc.ListUserCoupons(ctx, req)
}

// ListAvailableCoupons 可领取优惠券列表（领券中心用）
func (s *CouponService) ListAvailableCoupons(ctx context.Context, req *v1.ListAvailableCouponsRequest) (*v1.ListAvailableCouponsReply, error) {
	return s.uc.ListAvailableCoupons(ctx, req)
}

func (s *CouponService) UseCoupon(ctx context.Context, req *v1.UseCouponRequest) (*v1.UseCouponReply, error) {
	return s.uc.UseCoupon(ctx, req)
}

// ReleaseCoupon 返还优惠券（订单取消/退款时调用，解除核销）
func (s *CouponService) ReleaseCoupon(ctx context.Context, req *v1.ReleaseCouponRequest) (*v1.ReleaseCouponReply, error) {
	return s.uc.ReleaseCoupon(ctx, req)
}

func (s *CouponService) BatchExpireCoupons(ctx context.Context, req *v1.BatchExpireCouponsRequest) (*v1.BatchExpireCouponsReply, error) {
	return s.uc.BatchExpireCoupons(ctx, req)
}

func (s *CouponService) ListDistRecords(ctx context.Context, req *v1.ListDistRecordsRequest) (*v1.ListDistRecordsReply, error) {
	return s.uc.ListDistRecords(ctx, req)
}