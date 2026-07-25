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
	// 用户领券：事务中加锁更新库存并插入发放记录，返回记录ID
	ClaimCoupon(ctx context.Context, couponID, userID int64, source string) (int64, error)
	// ListUserCoupons 查询用户优惠券列表（含优惠券基本信息）
	ListUserCoupons(ctx context.Context, userID int64, status int32, page, pageSize int32) ([]*UserCouponItem, int32, error)
	// GetUserCouponRecord 根据记录ID和用户ID查询单条发放记录
	GetUserCouponRecord(ctx context.Context, recordID, userID int64) (*UserCouponItem, error)
	// UseCoupon 核销优惠券：更新发放记录状态和使用时间
	UseCoupon(ctx context.Context, recordID, userID, orderID int64) error
	// BatchExpireCoupons 批量过期已过期的优惠券发放记录
	BatchExpireCoupons(ctx context.Context, batchSize int64) (int32, error)
	// GetCouponByID 根据ID获取优惠券基本信息
	GetCouponByID(ctx context.Context, id int64) (*Coupon, error)
	// CountUserCoupons 统计用户领取某张优惠券的数量
	CountUserCoupons(ctx context.Context, couponID, userID int64) (int32, error)
	// UpdateIssuedCount 原子更新发放数量（带库存条件）
	UpdateIssuedCount(ctx context.Context, id int64) error
	// ListDistRecords 按优惠券ID查询发放记录列表
	ListDistRecords(ctx context.Context, couponID int64, page, pageSize int32) ([]*DistRecord, int32, error)
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
		Name:         req.Name,
		Type:         req.Type,
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		Quantity:     req.Quantity,
		IssuedCount:  0,
		UsedCount:    0,
		ValidStart:   req.ValidStart,
		ValidEnd:     req.ValidEnd,
		Enabled:      true,
		LimitPerUser: req.LimitPerUser,
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
	coupon.LimitPerUser = req.LimitPerUser

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
		LimitPerUser: coupon.LimitPerUser,
		CreateTime:   coupon.CreateTime,
		UpdateTime:   coupon.UpdateTime,
	}
}

// calculateDiscount 计算优惠金额
// couponType: 1-立减 2-折扣 3-满减
func calculateDiscount(couponType int32, value int64, orderAmount int64) int64 {
	switch couponType {
	case 1: // 立减
		if value > orderAmount {
			return orderAmount
		}
		return value
	case 2: // 折扣：value 为折扣百分比（1-99），优惠金额 = 订单金额 * (100 - value) / 100
		discount := orderAmount * (100 - value) / 100
		return discount
	case 3: // 满减
		if value > orderAmount {
			return orderAmount
		}
		return value
	}
	return 0
}

// ClaimCoupon 用户领券
func (uc *CouponUseCase) ClaimCoupon(ctx context.Context, req *v1.ClaimCouponRequest) (*v1.ClaimCouponReply, error) {
	// 1. 获取优惠券信息
	coupon, err := uc.repo.GetCouponByID(ctx, req.CouponId)
	if err != nil {
		return nil, err
	}

	// 2. 检查优惠券是否启用
	if !coupon.Enabled {
		return nil, v1.ErrorCouponInvalidValue("coupon is disabled")
	}

	// 3. 检查有效期
	now := time.Now().Unix()
	if now < coupon.ValidStart {
		return nil, v1.ErrorCouponInvalidValue("coupon not yet available")
	}
	if now > coupon.ValidEnd {
		return nil, v1.ErrorCouponInvalidValue("coupon has expired")
	}

	// 4. 检查库存
	if coupon.IssuedCount >= coupon.Quantity {
		return nil, v1.ErrorCouponInvalidValue("coupon is sold out")
	}

	// 5. 检查每人限领
	if coupon.LimitPerUser > 0 {
		count, err := uc.repo.CountUserCoupons(ctx, req.CouponId, req.UserId)
		if err != nil {
			return nil, err
		}
		if count >= coupon.LimitPerUser {
			return nil, v1.ErrorCouponInvalidValue("already claimed max times")
		}
	}

	// 6. 领券（repo 里使用事务加锁防止并发）
	recordID, err := uc.repo.ClaimCoupon(ctx, req.CouponId, req.UserId, req.Source)
	if err != nil {
		return nil, err
	}

	return &v1.ClaimCouponReply{RecordId: recordID}, nil
}

// ListUserCoupons 用户优惠券列表
func (uc *CouponUseCase) ListUserCoupons(ctx context.Context, req *v1.ListUserCouponsRequest) (*v1.ListUserCouponsReply, error) {
	items, total, err := uc.repo.ListUserCoupons(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	reply := &v1.ListUserCouponsReply{Total: total}
	for _, item := range items {
		reply.Items = append(reply.Items, &v1.UserCouponItem{
			RecordId:     item.RecordID,
			CouponId:     item.CouponID,
			Name:         item.Name,
			Type:         item.Type,
			Value:        item.Value,
			MinAmount:    item.MinAmount,
			ValidStart:   item.ValidStart,
			ValidEnd:     item.ValidEnd,
			RecordStatus: item.RecordStatus,
			UsedTime:     item.UsedTime,
			OrderId:      item.OrderID,
			Source:       item.Source,
			CreateTime:   item.CreateTime,
		})
	}
	return reply, nil
}

// UseCoupon 使用/核销优惠券
func (uc *CouponUseCase) UseCoupon(ctx context.Context, req *v1.UseCouponRequest) (*v1.UseCouponReply, error) {
	// 1. 获取发放记录
	record, err := uc.repo.GetUserCouponRecord(ctx, req.RecordId, req.UserId)
	if err != nil {
		return nil, err
	}

	// 2. 检查状态
	if record.RecordStatus != 0 {
		return nil, v1.ErrorCouponInvalidValue("coupon already used or expired")
	}

	// 3. 检查有效期
	coupon, err := uc.repo.GetCouponByID(ctx, record.CouponID)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if now < coupon.ValidStart || now > coupon.ValidEnd {
		return nil, v1.ErrorCouponInvalidValue("coupon is not valid")
	}

	// 4. 检查使用门槛
	if coupon.MinAmount > 0 && req.OrderAmount < coupon.MinAmount {
		return nil, v1.ErrorCouponInvalidValue("order amount does not meet minimum")
	}

	// 5. 计算优惠金额
	discount := calculateDiscount(coupon.Type, coupon.Value, req.OrderAmount)

	// 6. 核销
	if err := uc.repo.UseCoupon(ctx, req.RecordId, req.UserId, req.OrderId); err != nil {
		return nil, err
	}

	return &v1.UseCouponReply{Success: true, DiscountAmount: discount}, nil
}

// ListDistRecords 按优惠券ID查询发放记录
func (uc *CouponUseCase) ListDistRecords(ctx context.Context, req *v1.ListDistRecordsRequest) (*v1.ListDistRecordsReply, error) {
	records, total, err := uc.repo.ListDistRecords(ctx, req.CouponId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DistRecord, 0, len(records))
	for _, r := range records {
		items = append(items, &v1.DistRecord{
			Id:         r.ID,
			UserId:     r.UserID,
			Source:     r.Source,
			Status:     r.Status,
			UsedTime:   r.UsedTime,
			OrderId:    r.OrderID,
			CreateTime: r.CreateTime,
		})
	}
	return &v1.ListDistRecordsReply{Items: items, Total: total}, nil
}

// BatchExpireCoupons 批量过期优惠券
func (uc *CouponUseCase) BatchExpireCoupons(ctx context.Context, req *v1.BatchExpireCouponsRequest) (*v1.BatchExpireCouponsReply, error) {
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	count, err := uc.repo.BatchExpireCoupons(ctx, batchSize)
	if err != nil {
		return nil, err
	}
	return &v1.BatchExpireCouponsReply{AffectedCount: count}, nil
}