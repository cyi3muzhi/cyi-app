package service

import (
	"context"
	"encoding/json"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
	"cyi-app/pkg/kratos/service"
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
	hotKeywords := settings.HotKeywords
	if hotKeywords == nil {
		hotKeywords = []string{}
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
		HotKeywords:            hotKeywords,
	}, nil
}

// UpdateSettings 更新系统设置
func (s *AdminService) UpdateSettings(ctx context.Context, req *adminv1.UpdateSettingsRequest) (*adminv1.UpdateSettingsReply, error) {
	payments := req.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	hotKeywords := req.HotKeywords
	if hotKeywords == nil {
		hotKeywords = []string{}
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
		HotKeywords:            hotKeywords,
	}
	if err := s.settUC.Update(ctx, settings); err != nil {
		return nil, err
	}
	return &adminv1.UpdateSettingsReply{Success: true}, nil
}

// getPreference 获取用户偏好设置（内部方法，proto 接口通过 proto 方法接入）
func (s *AdminService) getPreference(ctx context.Context, userID int64) (json.RawMessage, error) {
	return s.settUC.GetPreference(ctx, userID)
}

// savePreference 保存用户偏好设置（内部方法，proto 接口通过 proto 方法接入）
func (s *AdminService) savePreference(ctx context.Context, userID int64, data json.RawMessage) error {
	return s.settUC.SavePreference(ctx, userID, data)
}

// GetPreference 获取当前用户偏好设置（proto 接口实现）
func (s *AdminService) GetPreference(ctx context.Context, req *adminv1.GetPreferenceRequest) (*adminv1.GetPreferenceReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}
	data, err := s.getPreference(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &adminv1.GetPreferenceReply{Data: data}, nil
}

// SavePreference 保存当前用户偏好设置（proto 接口实现）
func (s *AdminService) SavePreference(ctx context.Context, req *adminv1.SavePreferenceRequest) (*adminv1.SavePreferenceReply, error) {
	userID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}
	if err := s.savePreference(ctx, userID, req.Data); err != nil {
		return nil, err
	}
	return &adminv1.SavePreferenceReply{Success: true}, nil
}
