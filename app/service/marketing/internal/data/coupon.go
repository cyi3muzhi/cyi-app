package data

import (
	"context"
	"cyi-app/app/service/marketing/internal/biz"
	"cyi-app/app/service/marketing/internal/data/curd"
	mod "cyi-app/app/service/marketing/internal/data/models"
	"strconv"
	"strings"
	"time"

	v1 "cyi-app/api/service/marketing/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type couponRepo struct {
	data *Data
	log  *log.Helper
}

func NewCouponRepo(data *Data, logger log.Logger) biz.CouponRepo {
	return &couponRepo{data: data, log: log.NewHelper(logger)}
}

func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func (r *couponRepo) CreateCoupon(ctx context.Context, in *biz.Coupon) error {
	coupon := &mod.Coupon{
		Name:        in.Name,
		Type:        in.Type,
		Value:       in.Value,
		MinAmount:   in.MinAmount,
		Quantity:    in.Quantity,
		IssuedCount: in.IssuedCount,
		UsedCount:   in.UsedCount,
		ValidStart:  in.ValidStart,
		ValidEnd:    in.ValidEnd,
		Status:      in.Status,
		Enabled:     in.Enabled,
	}
	if err := curd.Q.Coupon.WithContext(ctx).Create(coupon); err != nil {
		return err
	}
	in.ID = coupon.ID
	return nil
}

func (r *couponRepo) UpdateCoupon(ctx context.Context, in *biz.Coupon) error {
	coupon := &mod.Coupon{
		Name:        in.Name,
		Type:        in.Type,
		Value:       in.Value,
		MinAmount:   in.MinAmount,
		Quantity:    in.Quantity,
		ValidStart:  in.ValidStart,
		ValidEnd:    in.ValidEnd,
		Enabled:     in.Enabled,
	}
	coupon.ID = in.ID
	result, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(in.ID)).Updates(coupon)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorCouponNotFound("coupon not found")
	}
	return nil
}

func (r *couponRepo) DeleteCoupon(ctx context.Context, id int64) error {
	result, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorCouponNotFound("coupon not found")
	}
	return nil
}

func (r *couponRepo) GetCoupon(ctx context.Context, id int64) (*biz.Coupon, error) {
	coupon, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(id)).First()
	if err != nil {
		return nil, v1.ErrorCouponNotFound("coupon not found")
	}
	return modelToBiz(coupon), nil
}

func (r *couponRepo) ListCoupon(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*biz.Coupon, int32, error) {
	do := curd.Q.Coupon.WithContext(ctx)

	if status != "" {
		statuses := strings.Split(status, ",")
		var statusInts []int32
		for _, s := range statuses {
			if sInt, err := parseInt32(s); err == nil {
				statusInts = append(statusInts, sInt)
			}
		}
		if len(statusInts) > 0 {
			if len(statusInts) == 1 {
				do = do.Where(curd.Coupon.Status.Eq(statusInts[0]))
			} else {
				do = do.Where(curd.Coupon.Status.In(statusInts...))
			}
		}
	}

	if keyword != "" {
		do = do.Where(curd.Coupon.Name.Like("%" + keyword + "%"))
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	if dateFrom != "" {
		fromTime, err := time.ParseInLocation("2006-01-02", dateFrom, loc)
		if err == nil {
			do = do.Where(curd.Coupon.ValidStart.Gte(fromTime.Unix()))
		}
	}
	if dateTo != "" {
		toTime, err := time.ParseInLocation("2006-01-02", dateTo, loc)
		if err == nil {
			toTime = toTime.Add(24*time.Hour - time.Second)
			do = do.Where(curd.Coupon.ValidEnd.Lte(toTime.Unix()))
		}
	}

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	coupons, err := do.Offset(offset).Limit(limit).Order(curd.Coupon.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Coupon, 0, len(coupons))
	for _, coupon := range coupons {
		items = append(items, modelToBiz(coupon))
	}

	return items, int32(total), nil
}

func (r *couponRepo) UpdateStatus(ctx context.Context, id int64, status int32) error {
	result, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(id)).Update(curd.Coupon.Status, status)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorCouponNotFound("coupon not found")
	}
	return nil
}

func (r *couponRepo) CouponStats(ctx context.Context) (*biz.CouponStats, error) {
	total, err := curd.Q.Coupon.WithContext(ctx).Count()
	if err != nil {
		return nil, err
	}

	activeCount, err := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.Status.Eq(biz.CouponStatusActive)).
		Count()
	if err != nil {
		return nil, err
	}

	var issuedCount int64
	err = curd.Q.Coupon.WithContext(ctx).
		Select(curd.Coupon.IssuedCount.Sum()).Scan(&issuedCount)
	if err != nil {
		return nil, err
	}

	expiredCount, err := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.Status.In(biz.CouponStatusEnded, biz.CouponStatusVoided)).
		Count()
	if err != nil {
		return nil, err
	}

	return &biz.CouponStats{
		TotalCount:   int32(total),
		ActiveCount:  int32(activeCount),
		IssuedCount:  issuedCount,
		ExpiredCount: int32(expiredCount),
	}, nil
}

func modelToBiz(coupon *mod.Coupon) *biz.Coupon {
	return &biz.Coupon{
		ID:           coupon.ID,
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