package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// SystemSettings 系统设置业务模型
type SystemSettings struct {
	ShopName               string
	ShopLogo               string
	ShopDescription        string
	ContactPhone           string
	ContactEmail           string
	IcpRecord              string
	AutoConfirmDays        int32
	AutoCloseMinutes       int32
	RefundTimeoutDays      int32
	StockAlertThreshold    int32
	NewProductOnline       bool
	ProductApproval        bool
	NewOrderNotify         bool
	RefundNotify           bool
	UserRegisterNotify     bool
	StockAlertNotify       bool
	LoginAttemptLimit      bool
	LoginCaptcha           bool
	SessionTimeoutMinutes  int32
	EnabledPayments        []string
	DefaultPayment         string
	FreeShippingThreshold  float64
	DefaultShippingCompany string
}

// SettingsRepo 系统设置仓库接口
type SettingsRepo interface {
	Get(ctx context.Context) (*SystemSettings, error)
	Update(ctx context.Context, s *SystemSettings) error
}

// SettingsUseCase 系统设置用例
type SettingsUseCase struct {
	repo SettingsRepo
	log  *log.Helper
}

// NewSettingsUseCase 创建 SettingsUseCase
func NewSettingsUseCase(repo SettingsRepo, logger log.Logger) *SettingsUseCase {
	return &SettingsUseCase{repo: repo, log: log.NewHelper(logger)}
}

// Get 获取系统设置
func (uc *SettingsUseCase) Get(ctx context.Context) (*SystemSettings, error) {
	return uc.repo.Get(ctx)
}

// Update 更新系统设置
func (uc *SettingsUseCase) Update(ctx context.Context, s *SystemSettings) error {
	return uc.repo.Update(ctx, s)
}
