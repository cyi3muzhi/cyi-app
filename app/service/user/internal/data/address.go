package data

import (
	"context"

	"cyi-app/app/service/user/internal/biz"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// addressRepo 收货地址数据仓库实现
type addressRepo struct {
	data *Data
	log  *log.Helper
}

// NewAddressRepo 创建收货地址数据仓库
func NewAddressRepo(data *Data, logger log.Logger) biz.AddressRepo {
	return &addressRepo{data: data, log: log.NewHelper(logger)}
}

// ListAddress 查询用户的收货地址列表（默认地址排在最前）
func (r *addressRepo) ListAddress(ctx context.Context, userID int64) ([]*biz.Address, error) {
	var list []mod.Address
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, id DESC").
		Find(&list).Error; err != nil {
		r.log.Errorf("list address failed, userID=%d: %v", userID, err)
		return nil, err
	}
	items := make([]*biz.Address, 0, len(list))
	for i := range list {
		items = append(items, addressModelToBiz(&list[i]))
	}
	return items, nil
}

// CreateAddress 新增收货地址
// 当用户无任何地址或请求标记为默认时，自动将新地址设为默认
func (r *addressRepo) CreateAddress(ctx context.Context, a *biz.Address) (*biz.Address, error) {
	// 统计用户现有地址数
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&mod.Address{}).
		Where("user_id = ?", a.UserID).Count(&count).Error; err != nil {
		r.log.Errorf("count address failed, userID=%d: %v", a.UserID, err)
		return nil, err
	}
	// 第一条地址自动设为默认
	if count == 0 {
		a.IsDefault = 1
	}

	m := &mod.Address{}
	addressBizToModel(a, m)
	if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
		r.log.Errorf("create address failed, userID=%d: %v", a.UserID, err)
		return nil, err
	}
	a.ID = m.ID

	// 若设为默认，取消其他默认地址
	if a.IsDefault == 1 {
		if err := r.clearDefault(ctx, a.UserID, m.ID); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// UpdateAddress 更新收货地址（仅允许更新本人地址）
func (r *addressRepo) UpdateAddress(ctx context.Context, a *biz.Address) error {
	// 先校验地址归属
	exists, err := r.checkOwner(ctx, a.ID, a.UserID)
	if err != nil {
		return err
	}
	if !exists {
		return biz.ErrAddressNotFound
	}

	updates := map[string]interface{}{
		"receiver":   a.Receiver,
		"phone":      a.Phone,
		"province":   a.Province,
		"city":       a.City,
		"district":   a.District,
		"detail":     a.Detail,
		"is_default": a.IsDefault,
	}
	if err := r.data.db.WithContext(ctx).Model(&mod.Address{}).
		Where("id = ? AND user_id = ?", a.ID, a.UserID).
		Updates(updates).Error; err != nil {
		r.log.Errorf("update address failed, id=%d: %v", a.ID, err)
		return err
	}
	// 若设为默认，取消其他默认地址
	if a.IsDefault == 1 {
		if err := r.clearDefault(ctx, a.UserID, a.ID); err != nil {
			return err
		}
	}
	return nil
}

// DeleteAddress 删除收货地址（仅允许删除本人地址）
func (r *addressRepo) DeleteAddress(ctx context.Context, id int64, userID int64) error {
	res := r.data.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&mod.Address{})
	if res.Error != nil {
		r.log.Errorf("delete address failed, id=%d: %v", id, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		return biz.ErrAddressNotFound
	}
	return nil
}

// SetDefaultAddress 设置默认收货地址（仅允许设置本人地址）
func (r *addressRepo) SetDefaultAddress(ctx context.Context, id int64, userID int64) error {
	// 先校验地址归属
	exists, err := r.checkOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return biz.ErrAddressNotFound
	}
	return r.clearDefault(ctx, userID, id)
}

// checkOwner 校验地址是否存在且属于指定用户
func (r *addressRepo) checkOwner(ctx context.Context, id int64, userID int64) (bool, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&mod.Address{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// clearDefault 将指定地址设为默认，其余地址取消默认
func (r *addressRepo) clearDefault(ctx context.Context, userID int64, keepID int64) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先清除该用户全部默认标记
		if err := tx.Model(&mod.Address{}).
			Where("user_id = ? AND is_default = 1", userID).
			Update("is_default", 0).Error; err != nil {
			return err
		}
		// 再设置目标地址为默认
		return tx.Model(&mod.Address{}).
			Where("id = ? AND user_id = ?", keepID, userID).
			Update("is_default", 1).Error
	})
}

// addressModelToBiz models.Address -> biz.Address 转换
func addressModelToBiz(m *mod.Address) *biz.Address {
	if m == nil {
		return nil
	}
	return &biz.Address{
		ID:         m.ID,
		UserID:     m.UserID,
		Receiver:   m.Receiver,
		Phone:      m.Phone,
		Province:   m.Province,
		City:       m.City,
		District:   m.District,
		Detail:     m.Detail,
		IsDefault:  m.IsDefault,
		CreateTime: m.CreateTime,
		UpdateTime: m.UpdateTime,
	}
}

// addressBizToModel biz.Address -> models.Address 转换
func addressBizToModel(a *biz.Address, m *mod.Address) {
	if a == nil || m == nil {
		return
	}
	m.ID = a.ID
	m.UserID = a.UserID
	m.Receiver = a.Receiver
	m.Phone = a.Phone
	m.Province = a.Province
	m.City = a.City
	m.District = a.District
	m.Detail = a.Detail
	m.IsDefault = a.IsDefault
}

// 确保 addressRepo 实现 biz.AddressRepo（编译期断言）
var _ biz.AddressRepo = (*addressRepo)(nil)
