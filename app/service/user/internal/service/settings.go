package service

import (
	"context"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
)

// AdminSettingsService 系统设置 gRPC 服务
type AdminSettingsService struct {
	v1.UnimplementedAdminSettingsServer
	uc *biz.SettingsUseCase
}

// NewAdminSettingsService 创建 AdminSettingsService
func NewAdminSettingsService(uc *biz.SettingsUseCase) *AdminSettingsService {
	return &AdminSettingsService{uc: uc}
}

// bizToProto 将 biz 结构体转换为 proto reply
func bizToProto(s *biz.SystemSettings) *v1.GetSettingsReply {
	if s == nil {
		return nil
	}
	payments := s.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &v1.GetSettingsReply{
		ShopName:               s.ShopName,
		ShopLogo:               s.ShopLogo,
		ShopDescription:        s.ShopDescription,
		ContactPhone:           s.ContactPhone,
		ContactEmail:           s.ContactEmail,
		IcpRecord:              s.IcpRecord,
		AutoConfirmDays:        s.AutoConfirmDays,
		AutoCloseMinutes:       s.AutoCloseMinutes,
		RefundTimeoutDays:      s.RefundTimeoutDays,
		StockAlertThreshold:    s.StockAlertThreshold,
		NewProductOnline:       s.NewProductOnline,
		ProductApproval:        s.ProductApproval,
		NewOrderNotify:         s.NewOrderNotify,
		RefundNotify:           s.RefundNotify,
		UserRegisterNotify:     s.UserRegisterNotify,
		StockAlertNotify:       s.StockAlertNotify,
		LoginAttemptLimit:      s.LoginAttemptLimit,
		LoginCaptcha:           s.LoginCaptcha,
		SessionTimeoutMinutes:  s.SessionTimeoutMinutes,
		EnabledPayments:        payments,
		DefaultPayment:         s.DefaultPayment,
		FreeShippingThreshold:  s.FreeShippingThreshold,
		DefaultShippingCompany: s.DefaultShippingCompany,
	}
}

// protoToBiz 将 proto request 转换为 biz 结构体
func protoToBiz(req *v1.UpdateSettingsRequest) *biz.SystemSettings {
	payments := req.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &biz.SystemSettings{
		ShopName:               req.ShopName,
		ShopLogo:               req.ShopLogo,
		ShopDescription:        req.ShopDescription,
		ContactPhone:           req.ContactPhone,
		ContactEmail:           req.ContactEmail,
		IcpRecord:              req.IcpRecord,
		AutoConfirmDays:        req.AutoConfirmDays,
		AutoCloseMinutes:       req.AutoCloseMinutes,
		RefundTimeoutDays:      req.RefundTimeoutDays,
		StockAlertThreshold:    req.StockAlertThreshold,
		NewProductOnline:       req.NewProductOnline,
		ProductApproval:        req.ProductApproval,
		NewOrderNotify:         req.NewOrderNotify,
		RefundNotify:           req.RefundNotify,
		UserRegisterNotify:     req.UserRegisterNotify,
		StockAlertNotify:       req.StockAlertNotify,
		LoginAttemptLimit:      req.LoginAttemptLimit,
		LoginCaptcha:           req.LoginCaptcha,
		SessionTimeoutMinutes:  req.SessionTimeoutMinutes,
		EnabledPayments:        payments,
		DefaultPayment:         req.DefaultPayment,
		FreeShippingThreshold:  req.FreeShippingThreshold,
		DefaultShippingCompany: req.DefaultShippingCompany,
	}
}

// GetSettings 获取系统设置
func (s *AdminSettingsService) GetSettings(ctx context.Context, req *v1.GetSettingsRequest) (*v1.GetSettingsReply, error) {
	settings, err := s.uc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return bizToProto(settings), nil
}

// UpdateSettings 更新系统设置
func (s *AdminSettingsService) UpdateSettings(ctx context.Context, req *v1.UpdateSettingsRequest) (*v1.UpdateSettingsReply, error) {
	if err := s.uc.Update(ctx, protoToBiz(req)); err != nil {
		return nil, err
	}
	return &v1.UpdateSettingsReply{Success: true}, nil
}
