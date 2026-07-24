package biz

import (
	"context"
	"time"

	v1 "cyi-app/api/service/marketing/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type CouponRepo interface {
	CreateCoupon(ctx context.Context, coupon *Coupon) error
	UpdateCoupon(ctx context.Context, coupon *Coupon) error
	DeleteCoupon(ctx context.Context, id int64) error
	GetCoupon(ctx context.Context, id int64) (*Coupon, error)
	ListCoupon(ctx context.Context, page, pageSize int32, status, keyword, dateFrom, dateTo string) ([]*Coupon, int32, error)
	UpdateStatus(ctx context.Context, id int64, status int32) error
	CouponStats(ctx context.Context) (*CouponStats, error)
}

type CouponUseCase struct {
	repo CouponRepo
	log  *log.Helper
}

func NewCouponUseCase(repo CouponRepo, logger log.Logger) *CouponUseCase {
	return &CouponUseCase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *CouponUseCase) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponReply, error) {
	coupon := &Coupon{
		Name:        req.Name,
		Type:        req.Type,
		Value:       req.Value,
		MinAmount:   req.MinAmount,
		Quantity:    req.Quantity,
		IssuedCount: 0,
		UsedCount:   0,
		ValidStart:  req.ValidStart,
		ValidEnd:    req.ValidEnd,
		Enabled:     true,
	}

	now := time.Now().Unix()
	coupon.Status = coupon.ComputeStatus(now)

	if err := coupon.Validate(); err != nil {
		return nil, v1.ErrorCouponInvalidValue("invalid coupon params")
	}

	if err := uc.repo.CreateCoupon(ctx, coupon); err != nil {
		return nil, v1.ErrorCouponCreateFailed("create coupon failed")
	}

	return &v1.CreateCouponReply{CouponId: coupon.ID}, nil
}

func (uc *CouponUseCase) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponReply, error) {
	coupon, err := uc.repo.GetCoupon(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if coupon.Status == CouponStatusVoided {
		return nil, v1.ErrorCouponVoided("coupon has been voided")
	}

	coupon.Name = req.Name
	coupon.Type = req.Type
	coupon.Value = req.Value
	coupon.MinAmount = req.MinAmount
	coupon.Quantity = req.Quantity
	coupon.ValidStart = req.ValidStart
	coupon.ValidEnd = req.ValidEnd
	coupon.Enabled = req.Enabled

	if err := uc.repo.UpdateCoupon(ctx, coupon); err != nil {
		return nil, v1.ErrorCouponUpdateFailed("update coupon failed")
	}

	return &v1.UpdateCouponReply{Success: true}, nil
}

func (uc *CouponUseCase) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponReply, error) {
	if err := uc.repo.DeleteCoupon(ctx, req.Id); err != nil {
		return nil, v1.ErrorCouponDeleteFailed("delete coupon failed")
	}

	return &v1.DeleteCouponReply{Success: true}, nil
}

func (uc *CouponUseCase) GetCoupon(ctx context.Context, req *v1.GetCouponRequest) (*v1.GetCouponReply, error) {
	coupon, err := uc.repo.GetCoupon(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetCouponReply{Coupon: couponToProto(coupon)}, nil
}

func (uc *CouponUseCase) ListCoupon(ctx context.Context, req *v1.ListCouponRequest) (*v1.ListCouponReply, error) {
	coupons, total, err := uc.repo.ListCoupon(ctx, req.Page, req.PageSize, req.Status, req.Keyword, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.Coupon, 0, len(coupons))
	for _, coupon := range coupons {
		items = append(items, couponToProto(coupon))
	}

	return &v1.ListCouponReply{Items: items, Total: total}, nil
}

func (uc *CouponUseCase) ToggleCouponStatus(ctx context.Context, req *v1.ToggleCouponStatusRequest) (*v1.ToggleCouponStatusReply, error) {
	if req.Action != 1 && req.Action != 2 {
		return nil, v1.ErrorCouponStatusInvalid("invalid action")
	}

	if req.Action == 1 {
		if err := uc.repo.UpdateStatus(ctx, req.Id, CouponStatusVoided); err != nil {
			return nil, err
		}
	} else {
		coupon, err := uc.repo.GetCoupon(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		now := time.Now().Unix()
		newStatus := coupon.ComputeStatus(now)
		if err := uc.repo.UpdateStatus(ctx, req.Id, newStatus); err != nil {
			return nil, err
		}
	}

	return &v1.ToggleCouponStatusReply{Success: true}, nil
}

func (uc *CouponUseCase) CouponStats(ctx context.Context, req *v1.CouponStatsRequest) (*v1.CouponStatsReply, error) {
	stats, err := uc.repo.CouponStats(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.CouponStatsReply{
		TotalCount:   stats.TotalCount,
		ActiveCount:  stats.ActiveCount,
		IssuedCount:  stats.IssuedCount,
		ExpiredCount: stats.ExpiredCount,
	}, nil
}

func couponToProto(coupon *Coupon) *v1.Coupon {
	return &v1.Coupon{
		Id:           coupon.ID,
		Name:         coupon.Name,
		Type:         coupon.Type,
		Value:        coupon.Value,
		MinAmount:    coupon.MinAmount,
		Quantity:     coupon.Quantity,
		IssuedCount:  coupon.IssuedCount,
		UsedCount:    coupon.UsedCount,
		ValidStart:   coupon.ValidStart,
		ValidEnd:     coupon.ValidEnd,
		Status:       coupon.Status,
		Enabled:      coupon.Enabled,
		CreateTime:   coupon.CreateTime,
		UpdateTime:   coupon.UpdateTime,
	}
}