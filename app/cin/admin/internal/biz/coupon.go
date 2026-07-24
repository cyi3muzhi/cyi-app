package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
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
	Total   int32
	Active  int32
	Pending int32
	Ended   int32
	Issued  int64
	Used    int64
}

type CouponRepo interface {
	ListCoupon(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Coupon, int32, error)
	GetCoupon(ctx context.Context, couponID int64) (*Coupon, error)
	CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error)
	UpdateCoupon(ctx context.Context, couponID int64, coupon *Coupon) (*Coupon, error)
	DeleteCoupon(ctx context.Context, couponID int64) error
	ToggleCouponStatus(ctx context.Context, couponID int64) error
	CouponStats(ctx context.Context) (*CouponStats, error)
}

type CouponUseCase struct {
	log  *log.Helper
	repo CouponRepo
}

func NewCouponUseCase(repo CouponRepo, logger log.Logger) *CouponUseCase {
	return &CouponUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (uc *CouponUseCase) List(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Coupon, int32, error) {
	return uc.repo.ListCoupon(ctx, page, pageSize, status, keyword, dateFrom, dateTo)
}

func (uc *CouponUseCase) Get(ctx context.Context, couponID int64) (*Coupon, error) {
	return uc.repo.GetCoupon(ctx, couponID)
}

func (uc *CouponUseCase) Create(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	return uc.repo.CreateCoupon(ctx, coupon)
}

func (uc *CouponUseCase) Update(ctx context.Context, couponID int64, coupon *Coupon) (*Coupon, error) {
	return uc.repo.UpdateCoupon(ctx, couponID, coupon)
}

func (uc *CouponUseCase) Delete(ctx context.Context, couponID int64) error {
	return uc.repo.DeleteCoupon(ctx, couponID)
}

func (uc *CouponUseCase) ToggleStatus(ctx context.Context, couponID int64) error {
	return uc.repo.ToggleCouponStatus(ctx, couponID)
}

func (uc *CouponUseCase) Stats(ctx context.Context) (*CouponStats, error) {
	return uc.repo.CouponStats(ctx)
}