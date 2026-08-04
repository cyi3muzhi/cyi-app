package biz

import (
	"context"

	v1 "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Address 收货地址业务对象
type Address struct {
	ID         int64
	UserID     int64
	Receiver   string // 收货人姓名
	Phone      string // 收货人手机号
	Province   string // 省份
	City       string // 城市
	District   string // 区县
	Detail     string // 详细地址
	IsDefault  int32  // 是否默认地址 0=否 1=是
	CreateTime int64
	UpdateTime int64
}

// AddressRepo 收货地址数据仓库接口
type AddressRepo interface {
	// ListAddress 查询用户地址列表（默认地址排前）
	ListAddress(ctx context.Context, userID int64) ([]*Address, error)
	// CreateAddress 新增地址
	CreateAddress(ctx context.Context, a *Address) (*Address, error)
	// UpdateAddress 更新地址
	UpdateAddress(ctx context.Context, a *Address) error
	// DeleteAddress 删除地址
	DeleteAddress(ctx context.Context, id int64, userID int64) error
	// SetDefaultAddress 设置默认地址
	SetDefaultAddress(ctx context.Context, id int64, userID int64) error
}

// AddressUseCase 收货地址业务逻辑
type AddressUseCase struct {
	repo AddressRepo
	log  *log.Helper
}

// NewAddressUseCase 创建 AddressUseCase
func NewAddressUseCase(repo AddressRepo, logger log.Logger) *AddressUseCase {
	return &AddressUseCase{repo: repo, log: log.NewHelper(logger)}
}

// ListAddress 查询用户地址列表
func (uc *AddressUseCase) ListAddress(ctx context.Context, userID int64) ([]*Address, error) {
	return uc.repo.ListAddress(ctx, userID)
}

// CreateAddress 新增地址
func (uc *AddressUseCase) CreateAddress(ctx context.Context, a *Address) (*Address, error) {
	created, err := uc.repo.CreateAddress(ctx, a)
	if err != nil {
		uc.log.Errorf("CreateAddress err: %v", err)
		return nil, err
	}
	return created, nil
}

// UpdateAddress 更新地址
func (uc *AddressUseCase) UpdateAddress(ctx context.Context, a *Address) error {
	if err := uc.repo.UpdateAddress(ctx, a); err != nil {
		uc.log.Errorf("UpdateAddress err: %v", err)
		if err == ErrAddressNotFound {
			return v1.ErrorAddressNotFound("address not found")
		}
		return err
	}
	return nil
}

// DeleteAddress 删除地址
func (uc *AddressUseCase) DeleteAddress(ctx context.Context, id int64, userID int64) error {
	if err := uc.repo.DeleteAddress(ctx, id, userID); err != nil {
		uc.log.Errorf("DeleteAddress err: %v", err)
		if err == ErrAddressNotFound {
			return v1.ErrorAddressNotFound("address not found")
		}
		return err
	}
	return nil
}

// SetDefaultAddress 设置默认地址
func (uc *AddressUseCase) SetDefaultAddress(ctx context.Context, id int64, userID int64) error {
	if err := uc.repo.SetDefaultAddress(ctx, id, userID); err != nil {
		uc.log.Errorf("SetDefaultAddress err: %v", err)
		if err == ErrAddressNotFound {
			return v1.ErrorAddressNotFound("address not found")
		}
		return err
	}
	return nil
}
