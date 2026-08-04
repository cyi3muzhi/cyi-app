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
		Name:         in.Name,
		Type:         in.Type,
		Value:        in.Value,
		MinAmount:    in.MinAmount,
		Quantity:     in.Quantity,
		IssuedCount:  in.IssuedCount,
		UsedCount:    in.UsedCount,
		ValidType:    in.ValidType,
		ValidDays:    in.ValidDays,
		ValidStart:   in.ValidStart,
		ValidEnd:     in.ValidEnd,
		Status:       in.Status,
		Enabled:      in.Enabled,
		LimitPerUser: in.LimitPerUser,
	}
	if err := curd.Q.Coupon.WithContext(ctx).Create(coupon); err != nil {
		return err
	}
	in.ID = coupon.ID
	return nil
}

func (r *couponRepo) UpdateCoupon(ctx context.Context, in *biz.Coupon) error {
	// 使用 map 进行更新，避免 struct Updates 跳过零值（如 Enabled=false）
	result, err := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.ID.Eq(in.ID)).
		UpdateSimple(
			curd.Coupon.Name.Value(in.Name),
			curd.Coupon.Type.Value(in.Type),
			curd.Coupon.Value.Value(in.Value),
			curd.Coupon.MinAmount.Value(in.MinAmount),
			curd.Coupon.Quantity.Value(in.Quantity),
			curd.Coupon.ValidType.Value(in.ValidType),
			curd.Coupon.ValidDays.Value(in.ValidDays),
			curd.Coupon.ValidStart.Value(in.ValidStart),
			curd.Coupon.ValidEnd.Value(in.ValidEnd),
			curd.Coupon.Enabled.Value(in.Enabled),
		)
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

	var issuedCount *int64
	err = curd.Q.Coupon.WithContext(ctx).
		Select(curd.Coupon.IssuedCount.Sum()).Scan(&issuedCount)
	if err != nil {
		return nil, err
	}
	issuedVal := int64(0)
	if issuedCount != nil {
		issuedVal = *issuedCount
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
		IssuedCount:  issuedVal,
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
		ValidType:    coupon.ValidType,
		ValidDays:    coupon.ValidDays,
		ValidStart:   coupon.ValidStart,
		ValidEnd:     coupon.ValidEnd,
		Status:       coupon.Status,
		Enabled:      coupon.Enabled,
		LimitPerUser: coupon.LimitPerUser,
		CreateTime:   coupon.CreateTime,
		UpdateTime:   coupon.UpdateTime,
	}
}

// modelToUserCouponItem 将发放记录模型转为业务对象
func modelToUserCouponItem(record *mod.CouponDistRecord, coupon *mod.Coupon) *biz.UserCouponItem {
	return &biz.UserCouponItem{
		RecordID:     record.ID,
		CouponID:     record.CouponID,
		Name:         coupon.Name,
		Type:         coupon.Type,
		Value:        coupon.Value,
		MinAmount:    coupon.MinAmount,
		ValidStart:   record.ValidStart,
		ValidEnd:     record.ValidEnd,
		RecordStatus: record.Status,
		UsedTime:     record.UsedTime,
		OrderID:      record.OrderID,
		Source:       record.Source,
		CreateTime:   record.CreateTime,
	}
}

// GetCouponByID 根据ID获取优惠券
func (r *couponRepo) GetCouponByID(ctx context.Context, id int64) (*biz.Coupon, error) {
	coupon, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(id)).First()
	if err != nil {
		return nil, v1.ErrorCouponNotFound("coupon not found")
	}
	return modelToBiz(coupon), nil
}

// CountUserCoupons 统计用户领取某张优惠券的数量
func (r *couponRepo) CountUserCoupons(ctx context.Context, couponID, userID int64) (int32, error) {
	count, err := curd.Q.CouponDistRecord.WithContext(ctx).
		Where(curd.CouponDistRecord.CouponID.Eq(couponID)).
		Where(curd.CouponDistRecord.UserID.Eq(userID)).
		Count()
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

// UpdateIssuedCount 原子更新发放数量（带库存条件）
// 执行 UPDATE coupon SET issued_count = issued_count + 1 WHERE id = ? AND issued_count < quantity
func (r *couponRepo) UpdateIssuedCount(ctx context.Context, id int64) error {
	result, err := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.ID.Eq(id)).
		Where(curd.Coupon.IssuedCount.LtCol(curd.Coupon.Quantity)).
		UpdateColumn(curd.Coupon.IssuedCount, curd.Coupon.IssuedCount.Add(1))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return v1.ErrorCouponInvalidValue("coupon is sold out")
	}
	return nil
}

// ClaimCoupon 用户领券（事务中原子更新库存并插入发放记录）
func (r *couponRepo) ClaimCoupon(ctx context.Context, couponID, userID int64, source string) (int64, error) {
	var recordID int64

	err := curd.Q.Transaction(func(tx *curd.Query) error {
		// 1. 查询优惠券信息（获取有效期配置）
		coupon, err := tx.Coupon.WithContext(ctx).
			Where(tx.Coupon.ID.Eq(couponID)).
			First()
		if err != nil {
			return v1.ErrorCouponNotFound("coupon not found")
		}

		// 2. 原子更新发放数量（带库存条件）
		result, err := tx.Coupon.WithContext(ctx).
			Where(tx.Coupon.ID.Eq(couponID)).
			Where(tx.Coupon.IssuedCount.LtCol(tx.Coupon.Quantity)).
			UpdateColumn(tx.Coupon.IssuedCount, tx.Coupon.IssuedCount.Add(1))
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return v1.ErrorCouponInvalidValue("coupon is sold out")
		}

		// 3. 根据有效期类型计算实际有效期
		now := time.Now().Unix()
		var validStart, validEnd int64
		switch coupon.ValidType {
		case biz.CouponValidTypeFixed:
			validStart = coupon.ValidStart
			validEnd = coupon.ValidEnd
		case biz.CouponValidTypeDays:
			validStart = now
			validEnd = now + int64(coupon.ValidDays)*24*60*60
		case biz.CouponValidTypeForever:
			validStart = now
			validEnd = 0 // 0表示永久有效
		}

		// 4. 创建发放记录
		record := &mod.CouponDistRecord{
			CouponID:   couponID,
			UserID:     userID,
			Source:     source,
			Status:     0, // 未使用
			ValidStart: validStart,
			ValidEnd:   validEnd,
		}
		if err := tx.CouponDistRecord.WithContext(ctx).Create(record); err != nil {
			return err
		}
		recordID = record.ID
		return nil
	})

	if err != nil {
		return 0, err
	}
	return recordID, nil
}

// ListUserCoupons 查询用户优惠券列表（含优惠券基本信息）
func (r *couponRepo) ListUserCoupons(ctx context.Context, userID int64, status int32, page, pageSize int32) ([]*biz.UserCouponItem, int32, error) {
	do := curd.Q.CouponDistRecord.WithContext(ctx).Where(curd.CouponDistRecord.UserID.Eq(userID))

	// 状态筛选：外部语义 1=未使用 2=已使用 3=已过期，与发放记录状态（0=未使用 1=已使用 2=已过期）差 1
	if status > 0 {
		recordStatus := status - 1
		do = do.Where(curd.CouponDistRecord.Status.Eq(recordStatus))
	}

	// 查询总数
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询记录
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	records, err := do.Offset(offset).Limit(limit).Order(curd.CouponDistRecord.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	// 批量加载优惠券信息
	items := make([]*biz.UserCouponItem, 0, len(records))
	for _, record := range records {
		coupon, err := curd.Q.Coupon.WithContext(ctx).
			Where(curd.Coupon.ID.Eq(record.CouponID)).
			First()
		if err != nil {
			// 如果优惠券已被删除，使用空名称
			coupon = &mod.Coupon{}
		}
		items = append(items, modelToUserCouponItem(record, coupon))
	}

	return items, int32(total), nil
}

// ListAvailableCoupons 查询可领取优惠券列表（上架中且启用）
// 是否已领/剩余量由 biz 层结合用户已领数计算，这里只返回券基础信息
func (r *couponRepo) ListAvailableCoupons(ctx context.Context, userID int64, page, pageSize int32) ([]*biz.Coupon, int32, error) {
	do := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.Status.Eq(biz.CouponStatusActive)). // 上架中
		Where(curd.Coupon.Enabled.Is(true))                    // 启用

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	offset := int((page - 1) * pageSize)
	coupons, err := do.Offset(offset).Limit(int(pageSize)).Order(curd.Coupon.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Coupon, 0, len(coupons))
	for _, c := range coupons {
		items = append(items, &biz.Coupon{
			ID:           c.ID,
			Name:         c.Name,
			Type:         c.Type,
			Value:        c.Value,
			MinAmount:    c.MinAmount,
			Quantity:     c.Quantity,
			IssuedCount:  c.IssuedCount,
			ValidType:    c.ValidType,
			ValidDays:    c.ValidDays,
			ValidStart:   c.ValidStart,
			ValidEnd:     c.ValidEnd,
			Status:       c.Status,
			Enabled:      c.Enabled,
			LimitPerUser: c.LimitPerUser,
			CreateTime:   c.CreateTime,
			UpdateTime:   c.UpdateTime,
		})
	}
	return items, int32(total), nil
}

// GetUserCouponRecord 根据记录ID和用户ID查询单条发放记录
func (r *couponRepo) GetUserCouponRecord(ctx context.Context, recordID, userID int64) (*biz.UserCouponItem, error) {
	record, err := curd.Q.CouponDistRecord.WithContext(ctx).
		Where(curd.CouponDistRecord.ID.Eq(recordID)).
		Where(curd.CouponDistRecord.UserID.Eq(userID)).
		First()
	if err != nil {
		return nil, v1.ErrorCouponNotFound("coupon record not found")
	}

	coupon, err := curd.Q.Coupon.WithContext(ctx).
		Where(curd.Coupon.ID.Eq(record.CouponID)).
		First()
	if err != nil {
		return nil, v1.ErrorCouponNotFound("coupon not found")
	}

	return modelToUserCouponItem(record, coupon), nil
}

// UseCoupon 核销优惠券
func (r *couponRepo) UseCoupon(ctx context.Context, recordID, userID, orderID int64) error {
	err := curd.Q.Transaction(func(tx *curd.Query) error {
		// 1. 先查询发放记录，获取 coupon_id
		record, err := tx.CouponDistRecord.WithContext(ctx).
			Where(tx.CouponDistRecord.ID.Eq(recordID)).
			Where(tx.CouponDistRecord.UserID.Eq(userID)).
			Where(tx.CouponDistRecord.Status.Eq(0)). // 仅更新未使用的
			First()
		if err != nil {
			return v1.ErrorCouponInvalidValue("coupon record not found or already used")
		}

		// 2. 更新发放记录状态为已使用
		now := time.Now().Unix()
		record.Status = 1 // 已使用
		record.UsedTime = now
		record.OrderID = orderID
		if _, err := tx.CouponDistRecord.WithContext(ctx).
			Where(tx.CouponDistRecord.ID.Eq(recordID)).
			Updates(record); err != nil {
			return err
		}

		// 3. 递增 coupon 表的 used_count
		_, err = tx.Coupon.WithContext(ctx).
			Where(tx.Coupon.ID.Eq(record.CouponID)).
			UpdateColumn(tx.Coupon.UsedCount, tx.Coupon.UsedCount.Add(1))
		return err
	})
	return err
}

// ReleaseCoupon 返还优惠券（订单取消/退款时调用）：解除核销，恢复为未使用
// 校验：记录归属该用户、状态为已使用（1）、核销订单与传入订单一致；成功则递减 used_count
func (r *couponRepo) ReleaseCoupon(ctx context.Context, recordID, userID, orderID int64) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		// 1. 查询已核销的发放记录（状态=1 已使用，且绑定该订单）
		record, err := tx.CouponDistRecord.WithContext(ctx).
			Where(tx.CouponDistRecord.ID.Eq(recordID)).
			Where(tx.CouponDistRecord.UserID.Eq(userID)).
			Where(tx.CouponDistRecord.Status.Eq(1)). // 已使用
			Where(tx.CouponDistRecord.OrderID.Eq(orderID)).
			First()
		if err != nil {
			return v1.ErrorCouponInvalidValue("coupon record not found or not bound to order")
		}

		// 2. 恢复为未使用，清空核销信息
		_, err = tx.CouponDistRecord.WithContext(ctx).
			Where(tx.CouponDistRecord.ID.Eq(recordID)).
			Updates(map[string]interface{}{
				"status":   0, // 未使用
				"used_time": 0,
				"order_id": 0,
			})
		if err != nil {
			return err
		}

		// 3. 递减 coupon 表的 used_count
		_, err = tx.Coupon.WithContext(ctx).
			Where(tx.Coupon.ID.Eq(record.CouponID)).
			UpdateColumn(tx.Coupon.UsedCount, tx.Coupon.UsedCount.Sub(1))
		return err
	})
}

// BatchExpireCoupons 批量过期已过期的优惠券发放记录
// 找出所有已过期（record.valid_end < now 且 valid_end > 0）但状态还是"未使用"的发放记录，批量更新状态为"已过期"
func (r *couponRepo) BatchExpireCoupons(ctx context.Context, batchSize int64) (int32, error) {
	now := time.Now().Unix()

	// 批量更新状态为"已过期"（status=2），限制处理数量
	// valid_end=0 表示永久有效，不过期
	result, err := curd.Q.CouponDistRecord.WithContext(ctx).
		Where(curd.CouponDistRecord.ValidEnd.Gt(0)).       // 排除永久有效
		Where(curd.CouponDistRecord.ValidEnd.Lt(now)).     // 已过期
		Where(curd.CouponDistRecord.Status.Eq(0)).         // 仅更新未使用的
		Where(curd.CouponDistRecord.UsedTime.Eq(0)).
		Limit(int(batchSize)).
		UpdateColumn(curd.CouponDistRecord.Status, 2) // 已过期

	if err != nil {
		return 0, err
	}

	return int32(result.RowsAffected), nil
}

// ListDistRecords 按优惠券ID查询发放记录列表
func (r *couponRepo) ListDistRecords(ctx context.Context, couponID int64, page, pageSize int32) ([]*biz.DistRecord, int32, error) {
	do := curd.Q.CouponDistRecord.WithContext(ctx).Where(curd.CouponDistRecord.CouponID.Eq(couponID))

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	records, err := do.Offset(offset).Limit(limit).Order(curd.CouponDistRecord.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	coupon, err := curd.Q.Coupon.WithContext(ctx).Where(curd.Coupon.ID.Eq(couponID)).First()
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DistRecord, 0, len(records))
	for _, rec := range records {
		validStart := rec.ValidStart
		validEnd := rec.ValidEnd

		if validStart == 0 {
			switch coupon.ValidType {
			case biz.CouponValidTypeFixed:
				validStart = coupon.ValidStart
				validEnd = coupon.ValidEnd
			case biz.CouponValidTypeDays:
				validStart = rec.CreateTime
				if coupon.ValidDays > 0 {
					validEnd = rec.CreateTime + int64(coupon.ValidDays)*24*60*60
				}
			case biz.CouponValidTypeForever:
				validStart = rec.CreateTime
				validEnd = 0
			default:
				validStart = rec.CreateTime
				validEnd = coupon.ValidEnd
			}
		}

		items = append(items, &biz.DistRecord{
			ID:         rec.ID,
			UserID:     rec.UserID,
			Source:     rec.Source,
			Status:     rec.Status,
			UsedTime:   rec.UsedTime,
			OrderID:    rec.OrderID,
			CreateTime: rec.CreateTime,
			ValidStart: validStart,
			ValidEnd:   validEnd,
		})
	}
	return items, int32(total), nil
}
