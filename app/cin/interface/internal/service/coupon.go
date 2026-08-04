package service

import (
	"context"
	"time"

	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	service "cyi-app/pkg/kratos/service"
)

// UnixToDateStr 将时间戳转换为日期字符串
func UnixToDateStr(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

// ClaimCoupon 用户领取优惠券
func (s *CinInterface) ClaimCoupon(ctx context.Context, req *v1.ClaimCouponRequest) (*v1.ClaimCouponReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	recordID, err := s.cc.ClaimCoupon(ctx, req.CouponId, userID, req.Source)
	if err != nil {
		return nil, err
	}

	return &v1.ClaimCouponReply{
		RecordId: recordID,
	}, nil
}

// ListMyCoupons 我的优惠券列表
func (s *CinInterface) ListMyCoupons(ctx context.Context, req *v1.ListMyCouponsRequest) (*v1.ListMyCouponsReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	items, total, err := s.cc.ListMyCoupons(ctx, userID, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	reply := &v1.ListMyCouponsReply{
		Total: total,
	}
	for _, item := range items {
		reply.Items = append(reply.Items, couponItemToReply(item))
	}
	return reply, nil
}

// couponItemToReply 将 biz 层的优惠券记录转换为响应格式
func couponItemToReply(item *biz.CouponItem) *v1.MyCouponItem {
	return &v1.MyCouponItem{
		RecordId:     item.RecordID,
		CouponId:     item.CouponID,
		Name:         item.Name,
		Type:         item.Type,
		Value:        item.Value,
		MinAmount:    item.MinAmount,
		ValidStart:   UnixToDateStr(item.ValidStart),
		ValidEnd:     UnixToDateStr(item.ValidEnd),
		RecordStatus: item.RecordStatus,
		UsedTime:     item.UsedTime,
		OrderId:      item.OrderID,
		Source:       item.Source,
		CreateTime:   UnixToDateStr(item.CreateTime),
	}
}

// ListAvailableCoupons 领券中心：可领取优惠券列表
func (s *CinInterface) ListAvailableCoupons(ctx context.Context, req *v1.ListAvailableCouponsRequest) (*v1.ListAvailableCouponsReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.cc.ListAvailableCoupons(ctx, userID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	reply := &v1.ListAvailableCouponsReply{Total: total}
	for _, item := range items {
		reply.Items = append(reply.Items, &v1.AvailableCouponItem{
			CouponId:     item.CouponID,
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
	return reply, nil
}

// ListUsableCoupons 结算可用优惠券：按订单金额筛选用户未使用且有效的券
func (s *CinInterface) ListUsableCoupons(ctx context.Context, req *v1.ListUsableCouponsRequest) (*v1.ListUsableCouponsReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.cc.ListUsableCoupons(ctx, userID, req.Amount)
	if err != nil {
		return nil, err
	}
	reply := &v1.ListUsableCouponsReply{}
	for _, item := range items {
		reply.Items = append(reply.Items, &v1.UsableCouponItem{
			RecordId:  item.RecordID,
			CouponId:  item.CouponID,
			Name:      item.Name,
			Type:      item.Type,
			Value:     item.Value,
			MinAmount: item.MinAmount,
			ValidEnd:  UnixToDateStr(item.ValidEnd),
			Usable:    item.Usable,
			Discount:  item.Discount,
			GapAmount: item.GapAmount,
		})
	}
	return reply, nil
}
