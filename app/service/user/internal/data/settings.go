package data

import (
	"context"
	"encoding/json"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// settingsRepo 系统设置数据仓库
type settingsRepo struct {
	data *Data
	log  *log.Helper
}

// NewSettingsRepo 创建 SettingsRepo
func NewSettingsRepo(data *Data, logger log.Logger) biz.SettingsRepo {
	return &settingsRepo{data: data, log: log.NewHelper(logger)}
}

// defaultSettings 返回默认系统设置
func defaultSettings() *biz.SystemSettings {
	return &biz.SystemSettings{
		ShopName:               "悦选商城",
		ShopDescription:        "精选好物，品质生活",
		ContactPhone:           "400-888-8888",
		ContactEmail:           "service@yuexuan.com",
		IcpRecord:              "京ICP备2025XXXXXX号",
		AutoConfirmDays:        7,
		AutoCloseMinutes:       30,
		RefundTimeoutDays:      3,
		StockAlertThreshold:    10,
		NewProductOnline:       true,
		ProductApproval:        false,
		NewOrderNotify:         true,
		RefundNotify:           true,
		UserRegisterNotify:     false,
		StockAlertNotify:       true,
		LoginAttemptLimit:      true,
		LoginCaptcha:           true,
		SessionTimeoutMinutes:  120,
		EnabledPayments:        []string{"wechat", "alipay"},
		DefaultPayment:         "wechat",
		FreeShippingThreshold:  99.00,
		DefaultShippingCompany: "顺丰速运",
	}
}

// settingsToJSON 将 biz 结构体序列化为 JSON
func settingsToJSON(s *biz.SystemSettings) (json.RawMessage, error) {
	return json.Marshal(s)
}

// settingsFromJSON 将 JSON 反序列化为 biz 结构体
func settingsFromJSON(data json.RawMessage) (*biz.SystemSettings, error) {
	var s biz.SystemSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Get 获取系统设置，首次查询时自动创建默认配置
func (r *settingsRepo) Get(ctx context.Context) (*biz.SystemSettings, error) {
	var row models.SettingsRow
	err := r.data.db.WithContext(ctx).First(&row, 1).Error
	if err == gorm.ErrRecordNotFound {
		// 首次使用，创建默认配置（ID=1，嵌入 gorm.Model 自动管理时间）
		def := defaultSettings()
		jsonData, _ := settingsToJSON(def)
		row = models.SettingsRow{
			Settings:  jsonData,
			UpdatedBy: "system",
		}
		row.ID = 1
		if err := r.data.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return def, nil
	}
	if err != nil {
		return nil, err
	}
	return settingsFromJSON(row.Settings)
}

// Update 更新系统设置（全量覆盖，gorm.Model 自动更新 update_time）
func (r *settingsRepo) Update(ctx context.Context, s *biz.SystemSettings) error {
	jsonData, err := settingsToJSON(s)
	if err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&models.SettingsRow{}).
		Where("id = ?", 1).
		Update("settings", string(jsonData)).Error
}
