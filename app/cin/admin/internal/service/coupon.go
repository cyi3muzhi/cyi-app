package service

import (
	"context"
	"sync"
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

var distRecordStatusMap = map[int32]string{
	0: "未使用",
	1: "已使用",
	2: "已过期",
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

func distRecordStatusText(s int32) string {
	if text, ok := distRecordStatusMap[s]; ok {
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

	// 列表不查发放记录，展开行时按需加载
	replyItems := make([]*adminv1.CouponItem, 0, len(items))
	for _, item := range items {
		replyItems = append(replyItems, bizToAdminProtoCoupon(item))
	}

	return &adminv1.CouponListReply{
		Items: replyItems,
		Total: total,
	}, nil
}

func (s *AdminService) GetCouponStats(ctx context.Context, req *adminv1.CouponStatsRequest) (*adminv1.CouponStatsReply, error) {
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

	proto := bizToAdminProtoCoupon(coupon)
	if records, _, err := s.couponUC.ListDistRecords(ctx, req.CouponId, 1, 100); err == nil {
		proto.DistRecords = bizDistRecordsToProto(records)
	} else {
		s.log.Warnf("query dist records for coupon %d failed: %v", req.CouponId, err)
	}

	return &adminv1.CouponGetReply{
		Coupon: proto,
	}, nil
}

func (s *AdminService) CreateCoupon(ctx context.Context, req *adminv1.CouponCreateRequest) (*adminv1.CouponCreateReply, error) {
	coupon := &biz.Coupon{
		Name:         req.Name,
		Type:         parseCouponType(req.Type),
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		Quantity:     req.Quantity,
		LimitPerUser: req.LimitPerUser,
		ValidType:    req.ValidType,
		ValidDays:    req.ValidDays,
		ValidStart:   parseTime(req.ValidStart),
		ValidEnd:     parseTime(req.ValidEnd),
		Enabled:      true,
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
		Name:         req.Name,
		Type:         parseCouponType(req.Type),
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		Quantity:     req.Quantity,
		LimitPerUser: req.LimitPerUser,
		ValidType:    req.ValidType,
		ValidDays:    req.ValidDays,
		ValidStart:   parseTime(req.ValidStart),
		ValidEnd:     parseTime(req.ValidEnd),
		Enabled:      req.Enabled,
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

func (s *AdminService) ListCouponRecords(ctx context.Context, req *adminv1.CouponRecordsRequest) (*adminv1.CouponRecordsReply, error) {
	records, total, err := s.couponUC.ListDistRecords(ctx, req.CouponId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := bizDistRecordsToProto(records)
	s.fillRecordUsers(ctx, items)
	return &adminv1.CouponRecordsReply{
		Items: items,
		Total: total,
	}, nil
}

// fillRecordUsers 批量填充发放记录的用户昵称与头像
// 去重后并发查询用户服务，避免逐条 N+1
func (s *AdminService) fillRecordUsers(ctx context.Context, items []*adminv1.DistRecord) {
	idSet := make(map[int64]struct{}, len(items))
	for _, it := range items {
		if it.UserId > 0 {
			idSet[it.UserId] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return
	}

	userMap := make(map[int64]*biz.UserInfo, len(idSet))
	var wg sync.WaitGroup
	var mu sync.Mutex
	// 限制并发查询数量，避免用户数过多时一次性开大量 goroutine
	sem := make(chan struct{}, 10)
	for uid := range idSet {
		wg.Add(1)
		sem <- struct{}{}
		go func(id int64) {
			defer wg.Done()
			defer func() { <-sem }()
			u, err := s.userUC.GetUser(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			userMap[id] = u
			mu.Unlock()
		}(uid)
	}
	wg.Wait()

	for _, it := range items {
		u, ok := userMap[it.UserId]
		if !ok || u == nil {
			continue
		}
		// 优先昵称，缺失时用登录账号兜底
		it.User = u.Nickname
		if it.User == "" {
			it.User = u.Username
		}
		it.Avatar = u.Avatar
	}
}

func (s *AdminService) ToggleCouponStatus(ctx context.Context, req *adminv1.CouponToggleStatusRequest) (*adminv1.CouponToggleStatusReply, error) {
	err := s.couponUC.ToggleStatus(ctx, req.CouponId, req.Action)
	if err != nil {
		return nil, err
	}

	return &adminv1.CouponToggleStatusReply{
		Success: true,
	}, nil
}

func bizToAdminProtoCoupon(c *biz.Coupon) *adminv1.CouponItem {
	var validStartStr, validEndStr string
	switch c.ValidType {
	case 2:
		validStartStr = ""
		validEndStr = ""
	case 3:
		validStartStr = ""
		validEndStr = ""
	default:
		if c.ValidStart > 0 {
			validStartStr = time.Unix(c.ValidStart, 0).Format("2006-01-02")
		}
		if c.ValidEnd > 0 {
			validEndStr = time.Unix(c.ValidEnd, 0).Format("2006-01-02")
		}
	}

	return &adminv1.CouponItem{
		Id:           c.ID,
		Name:         c.Name,
		Type:         couponTypeText(c.Type),
		Value:        c.Value,
		MinAmount:    c.MinAmount,
		Quantity:     c.Quantity,
		IssuedCount:  c.IssuedCount,
		UsedCount:    c.UsedCount,
		LimitPerUser: c.LimitPerUser,
		ValidStart:   validStartStr,
		ValidEnd:     validEndStr,
		Status:       couponStatusText(c.Status),
		Enabled:      c.Enabled,
		CreateTime:   c.CreateTime,
		UpdateTime:   c.UpdateTime,
		ValidType:    c.ValidType,
		ValidDays:    c.ValidDays,
	}
}

func bizDistRecordsToProto(records []*biz.DistRecord) []*adminv1.DistRecord {
	items := make([]*adminv1.DistRecord, 0, len(records))
	for _, r := range records {
		timeStr := ""
		if r.CreatedAt > 0 {
			timeStr = time.Unix(r.CreatedAt, 0).Format("2006-01-02 15:04:05")
		}

		validStartStr := ""
		if r.ValidStart > 0 {
			validStartStr = time.Unix(r.ValidStart, 0).Format("2006-01-02")
		}

		validEndStr := ""
		if r.ValidEnd > 0 {
			validEndStr = time.Unix(r.ValidEnd, 0).Format("2006-01-02")
		}

		items = append(items, &adminv1.DistRecord{
			Id:         r.ID,
			UserId:     r.UserID,
			Source:     r.Source,
			Status:     distRecordStatusText(r.Status),
			UsedTime:   r.UsedTime,
			OrderId:    r.OrderID,
			Time:       timeStr,
			User:       "",
			CreateTime: r.CreatedAt,
			ValidStart: validStartStr,
			ValidEnd:   validEndStr,
		})
	}
	return items
}