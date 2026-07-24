package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
)

// GetSettings 获取系统设置
func (s *AdminService) GetSettings(ctx context.Context, req *adminv1.GetSettingsRequest) (*adminv1.GetSettingsReply, error) {
	settings, err := s.settUC.Get(ctx)
	if err != nil {
		return nil, err
	}
	payments := settings.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &adminv1.GetSettingsReply{
		ShopName:               settings.ShopName,
		ShopLogo:               settings.ShopLogo,
		ShopDescription:        settings.ShopDescription,
		ContactPhone:           settings.ContactPhone,
		ContactEmail:           settings.ContactEmail,
		IcpRecord:              settings.IcpRecord,
		AutoConfirmDays:        settings.AutoConfirmDays,
		AutoCloseMinutes:       settings.AutoCloseMinutes,
		RefundTimeoutDays:      settings.RefundTimeoutDays,
		StockAlertThreshold:    settings.StockAlertThreshold,
		NewProductOnline:       settings.NewProductOnline,
		ProductApproval:        settings.ProductApproval,
		NewOrderNotify:         settings.NewOrderNotify,
		RefundNotify:           settings.RefundNotify,
		UserRegisterNotify:     settings.UserRegisterNotify,
		StockAlertNotify:       settings.StockAlertNotify,
		LoginAttemptLimit:      settings.LoginAttemptLimit,
		LoginCaptcha:           settings.LoginCaptcha,
		SessionTimeoutMinutes:  settings.SessionTimeoutMinutes,
		EnabledPayments:        payments,
		DefaultPayment:         settings.DefaultPayment,
		FreeShippingThreshold:  settings.FreeShippingThreshold,
		DefaultShippingCompany: settings.DefaultShippingCompany,
	}, nil
}

// UpdateSettings 更新系统设置
func (s *AdminService) UpdateSettings(ctx context.Context, req *adminv1.UpdateSettingsRequest) (*adminv1.UpdateSettingsReply, error) {
	payments := req.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	settings := &biz.SystemSettings{
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
	if err := s.settUC.Update(ctx, settings); err != nil {
		return nil, err
	}
	return &adminv1.UpdateSettingsReply{Success: true}, nil
}
