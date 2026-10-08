package models

import (
	"encoding/json"

	"cyi-app/pkg/util/gorm"
)

// SystemSettings 系统设置，以 JSON 存储在 system_settings 表
type SystemSettings struct {
	ShopName               string   `json:"shop_name"`
	ShopLogo               string   `json:"shop_logo"`
	ShopDescription        string   `json:"shop_description"`
	ContactPhone           string   `json:"contact_phone"`
	ContactEmail           string   `json:"contact_email"`
	IcpRecord              string   `json:"icp_record"`
	AutoConfirmDays        int32    `json:"auto_confirm_days"`
	AutoCloseMinutes       int32    `json:"auto_close_minutes"`
	RefundTimeoutDays      int32    `json:"refund_timeout_days"`
	StockAlertThreshold    int32    `json:"stock_alert_threshold"`
	NewProductOnline       bool     `json:"new_product_online"`
	ProductApproval        bool     `json:"product_approval"`
	NewOrderNotify         bool     `json:"new_order_notify"`
	RefundNotify           bool     `json:"refund_notify"`
	UserRegisterNotify     bool     `json:"user_register_notify"`
	StockAlertNotify       bool     `json:"stock_alert_notify"`
	LoginAttemptLimit      bool     `json:"login_attempt_limit"`
	LoginCaptcha           bool     `json:"login_captcha"`
	SessionTimeoutMinutes  int32    `json:"session_timeout_minutes"`
	EnabledPayments        []string `json:"enabled_payments"`
	DefaultPayment         string   `json:"default_payment"`
	FreeShippingThreshold  float64  `json:"free_shipping_threshold"`
	DefaultShippingCompany string   `json:"default_shipping_company"`
}

// SettingsRow 数据库行记录，嵌入 gorm.Model 统一管理时间
type SettingsRow struct {
	gorm.Model
	Settings  json.RawMessage `gorm:"type:json;not null"`
	UpdatedBy string          `gorm:"size:64"`
}

// TableName 指定表名
func (SettingsRow) TableName() string {
	return "system_settings"
}
