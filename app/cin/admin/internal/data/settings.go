package data

import (
	"context"
	"encoding/json"
	"fmt"

	userV1 "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// preferenceKeyPrefix 用户偏好设置的 Redis key 前缀
const preferenceKeyPrefix = "preference:"

// settingsRepo 系统设置数据仓库（gRPC 代理到 User Service + Redis 偏好设置）
type settingsRepo struct {
	client userV1.AdminSettingsClient
	rdb    redis.UniversalClient
	log    *log.Helper
}

// NewSettingsRepo 创建 SettingsRepo
func NewSettingsRepo(client userV1.AdminSettingsClient, rdb redis.UniversalClient, logger log.Logger) biz.SettingsRepo {
	return &settingsRepo{client: client, rdb: rdb, log: log.NewHelper(logger)}
}

// bizToProto 将 biz 结构体转换为 proto request
func bizToProto(s *biz.SystemSettings) *userV1.UpdateSettingsRequest {
	payments := s.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &userV1.UpdateSettingsRequest{
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

// protoToBiz 将 proto reply 转换为 biz 结构体
func protoToBiz(resp *userV1.GetSettingsReply) *biz.SystemSettings {
	payments := resp.EnabledPayments
	if payments == nil {
		payments = []string{}
	}
	return &biz.SystemSettings{
		ShopName:               resp.ShopName,
		ShopLogo:               resp.ShopLogo,
		ShopDescription:        resp.ShopDescription,
		ContactPhone:           resp.ContactPhone,
		ContactEmail:           resp.ContactEmail,
		IcpRecord:              resp.IcpRecord,
		AutoConfirmDays:        resp.AutoConfirmDays,
		AutoCloseMinutes:       resp.AutoCloseMinutes,
		RefundTimeoutDays:      resp.RefundTimeoutDays,
		StockAlertThreshold:    resp.StockAlertThreshold,
		NewProductOnline:       resp.NewProductOnline,
		ProductApproval:        resp.ProductApproval,
		NewOrderNotify:         resp.NewOrderNotify,
		RefundNotify:           resp.RefundNotify,
		UserRegisterNotify:     resp.UserRegisterNotify,
		StockAlertNotify:       resp.StockAlertNotify,
		LoginAttemptLimit:      resp.LoginAttemptLimit,
		LoginCaptcha:           resp.LoginCaptcha,
		SessionTimeoutMinutes:  resp.SessionTimeoutMinutes,
		EnabledPayments:        payments,
		DefaultPayment:         resp.DefaultPayment,
		FreeShippingThreshold:  resp.FreeShippingThreshold,
		DefaultShippingCompany: resp.DefaultShippingCompany,
	}
}

// Get 获取系统设置
func (r *settingsRepo) Get(ctx context.Context) (*biz.SystemSettings, error) {
	resp, err := r.client.GetSettings(ctx, &userV1.GetSettingsRequest{})
	if err != nil {
		return nil, err
	}
	return protoToBiz(resp), nil
}

// Update 更新系统设置
func (r *settingsRepo) Update(ctx context.Context, s *biz.SystemSettings) error {
	_, err := r.client.UpdateSettings(ctx, bizToProto(s))
	return err
}

// GetPreference 从 Redis 获取用户偏好设置（JSON 格式）
// 偏好设置按 userID 存储在 "preference:admin:{userID}" key 中
func (r *settingsRepo) GetPreference(ctx context.Context, userID int64) (json.RawMessage, error) {
	if r.rdb == nil {
		return json.RawMessage(`{}`), nil
	}
	key := fmt.Sprintf("%sadmin:%d", preferenceKeyPrefix, userID)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return json.RawMessage(`{}`), nil
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// SavePreference 将用户偏好设置保存到 Redis
func (r *settingsRepo) SavePreference(ctx context.Context, userID int64, data json.RawMessage) error {
	if r.rdb == nil {
		return fmt.Errorf("存储服务不可用")
	}
	key := fmt.Sprintf("%sadmin:%d", preferenceKeyPrefix, userID)
	return r.rdb.Set(ctx, key, []byte(data), 0).Err()
}
