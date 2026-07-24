package service

import (
	"context"
	"time"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

var couponTypeMap = map[int32]string{
	1: "立减",
	2: "折扣",
	3: "满减",
}

var couponStatusMap = map[int32]string{
	1: "进行中",
	2: "未开始",
	3: "已结束",
	4: "已作废",
}

func couponTypeText(t int32) string {
	if text, ok := couponTypeMap[t]; ok {
		return text
	}
	return "未知"
}

func couponStatusText(s int32) string {
	if text, ok := couponStatusMap[s]; ok {
		return text
	}
	return "未知"
}

func parseCouponType(t string) int32 {
	switch t {
	case "立减":
		return 1
	case "折扣":
		return 2
	case "满减":
		return 3
	default:
		return 1
	}
}

func parseTime(s string) int64 {
	if s == "" {
		return 0
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func (s *AdminService) ListCoupon(ctx context.Context, req *adminv1.CouponListRequest) (*adminv1.CouponListReply, error) {
	items, total, err := s.couponUC.List(ctx, req.Page, req.PageSize, req.Status, req.Keyword, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, err
	}

	replyItems := make([]*adminv1.CouponItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizToAdminProtoCoupon(item))
	}

	return &adminv1.CouponListReply{
		Items: replyItems,
		Total: total,
	}, nil
}

func (s *AdminService) CouponStats(ctx context.Context, req *adminv1.CouponStatsRequest) (*adminv1.CouponStatsReply, error) {
	stats, err := s.couponUC.Stats(ctx)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponStatsReply{
		TotalCount:   stats.Total,
		ActiveCount:  stats.Active,
		IssuedCount:  stats.Issued,
		ExpiredCount: stats.Ended,
	}, nil
}

func (s *AdminService) GetCoupon(ctx context.Context, req *adminv1.CouponGetRequest) (*adminv1.CouponGetReply, error) {
	coupon, err := s.couponUC.Get(ctx, req.CouponId)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponGetReply{
		Coupon: bizToAdminProtoCoupon(coupon),
	}, nil
}

func (s *AdminService) CreateCoupon(ctx context.Context, req *adminv1.CouponCreateRequest) (*adminv1.CouponCreateReply, error) {
	coupon := &biz.Coupon{
		Name:        req.Name,
		Type:        parseCouponType(req.Type),
		Value:       req.Value,
		MinAmount:   req.MinAmount,
		Quantity:    req.Quantity,
		ValidStart:  parseTime(req.ValidStart),
		ValidEnd:    parseTime(req.ValidEnd),
		Enabled:     true,
	}

	result, err := s.couponUC.Create(ctx, coupon)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponCreateReply{
		CouponId: result.ID,
	}, nil
}

func (s *AdminService) UpdateCoupon(ctx context.Context, req *adminv1.CouponUpdateRequest) (*adminv1.CouponUpdateReply, error) {
	coupon := &biz.Coupon{
		Name:        req.Name,
		Type:        parseCouponType(req.Type),
		Value:       req.Value,
		MinAmount:   req.MinAmount,
		Quantity:    req.Quantity,
		ValidStart:  parseTime(req.ValidStart),
		ValidEnd:    parseTime(req.ValidEnd),
		Enabled:     req.Enabled,
	}

	_, err := s.couponUC.Update(ctx, req.CouponId, coupon)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponUpdateReply{
		Success: true,
	}, nil
}

func (s *AdminService) DeleteCoupon(ctx context.Context, req *adminv1.CouponDeleteRequest) (*adminv1.CouponDeleteReply, error) {
	err := s.couponUC.Delete(ctx, req.CouponId)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponDeleteReply{
		Success: true,
	}, nil
}

func (s *AdminService) ToggleCouponStatus(ctx context.Context, req *adminv1.CouponToggleStatusRequest) (*adminv1.CouponToggleStatusReply, error) {
	err := s.couponUC.ToggleStatus(ctx, req.CouponId)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponToggleStatusReply{
		Success: true,
	}, nil
}

func bizToAdminProtoCoupon(c *biz.Coupon) *adminv1.CouponItem {
	return &adminv1.CouponItem{
		Id:          c.ID,
		Name:        c.Name,
		Type:        couponTypeText(c.Type),
		Value:       c.Value,
		MinAmount:   c.MinAmount,
		Quantity:    c.Quantity,
		IssuedCount: c.IssuedCount,
		UsedCount:   c.UsedCount,
		ValidStart:  time.Unix(c.ValidStart, 0).Format("2006-01-02"),
		ValidEnd:    time.Unix(c.ValidEnd, 0).Format("2006-01-02"),
		Status:      couponStatusText(c.Status),
		Enabled:     c.Enabled,
		CreateTime:  c.CreateTime,
		UpdateTime:  c.UpdateTime,
	}
}