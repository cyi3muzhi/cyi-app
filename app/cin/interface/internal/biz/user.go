package biz

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Address 收货地址领域对象
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
}

// CenterStats 个人中心统计
type CenterStats struct {
	TotalOrders    int64 // 总订单数
	PendingPay     int64 // 待付款
	PendingShip    int64 // 待发货
	PendingReceive int64 // 待收货
	PendingReview  int64 // 待评价（已完成）
	RefundCount    int64 // 退款/售后
	CouponCount    int64 // 可用优惠券数
	TotalSpent     int64 // 累计消费（分）
}

// UserCenter 个人中心概览（用户信息 + 统计）
type UserCenter struct {
	User  *User
	Stats *CenterStats
}

type UserRepo interface {
	// 添加用户
	AddUser(ctx context.Context, user *User) error
	// 移除用户
	RemoveUser(ctx context.Context, id int64) error
	// 修改用户
	ModifyUser(ctx context.Context, user *User) error
	// 获取用户
	GetUser(ctx context.Context, id int64) (*User, error)
	// 查询用户列表
	FindUserList(ctx context.Context) ([]*User, error)

	FindByUsername(ctx context.Context, username string) (*User, error)
	// 校验密码是否正确
	VerifyPassword(ctx context.Context, user *User, password string) error
	// 获取用户信息

	// 注册用户
	Register(ctx context.Context, req *v1.RegisterReq) error
	// 登录用户
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error)
	// 修改密码
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	// 个人中心概览（聚合用户信息 + 订单/优惠券统计）
	GetUserCenter(ctx context.Context, userID int64) (*UserCenter, error)
	// 收货地址列表
	ListAddress(ctx context.Context, userID int64) ([]*Address, error)
	// 新增收货地址
	CreateAddress(ctx context.Context, a *Address) (*Address, error)
	// 更新收货地址
	UpdateAddress(ctx context.Context, a *Address) error
	// 删除收货地址
	DeleteAddress(ctx context.Context, id int64, userID int64) error
	// 设置默认收货地址
	SetDefaultAddress(ctx context.Context, id int64, userID int64) error
}

type UserUseCase struct {
	log  *log.Helper
	repo UserRepo
}

func NewUserUseCase(repo UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// GetUser 获取用户
// TODO: 待业务需求明确后实现
func (u *UserUseCase) GetUser(ctx context.Context, id int64) (*User, error) {
	return nil, nil
}

// 获取用户信息
func (u *UserUseCase) UserInfo(ctx context.Context, id int64) (*User, error) {
	user, err := u.repo.GetUser(ctx, id)
	if err != nil {
		u.log.Errorf("UserInfo err: %v", err)
		return nil, v1.ErrorUserNotExist("user not exist")
	}
	return user, nil
}

// 更新用户信息
func (u *UserUseCase) UpdateUserInfo(ctx context.Context, user *User) error {
	err := u.repo.ModifyUser(ctx, user)
	if err != nil {
		u.log.Errorf("UpdateUserInfo err: %v", err)
		return v1.ErrorUserUpdateFailed("user update failed")
	}
	return nil
}

// 注册用户
func (u *UserUseCase) Register(ctx context.Context, req *v1.RegisterReq) error {
	return u.repo.Register(ctx, req)
}

func (u *UserUseCase) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return u.repo.Login(ctx, req)
}

// ChangePassword 修改密码
func (u *UserUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	err := u.repo.ChangePassword(ctx, userID, oldPassword, newPassword)
	if err != nil {
		u.log.Errorf("ChangePassword err: %v", err)
		return err
	}
	return nil
}

// GetUserCenter 获取个人中心概览
func (u *UserUseCase) GetUserCenter(ctx context.Context, userID int64) (*UserCenter, error) {
	center, err := u.repo.GetUserCenter(ctx, userID)
	if err != nil {
		u.log.Errorf("GetUserCenter err: %v", err)
		return nil, err
	}
	return center, nil
}

// ListAddress 获取收货地址列表
func (u *UserUseCase) ListAddress(ctx context.Context, userID int64) ([]*Address, error) {
	list, err := u.repo.ListAddress(ctx, userID)
	if err != nil {
		u.log.Errorf("ListAddress err: %v", err)
		return nil, err
	}
	return list, nil
}

// CreateAddress 新增收货地址
func (u *UserUseCase) CreateAddress(ctx context.Context, a *Address) (*Address, error) {
	created, err := u.repo.CreateAddress(ctx, a)
	if err != nil {
		u.log.Errorf("CreateAddress err: %v", err)
		return nil, err
	}
	return created, nil
}

// UpdateAddress 更新收货地址
func (u *UserUseCase) UpdateAddress(ctx context.Context, a *Address) error {
	if err := u.repo.UpdateAddress(ctx, a); err != nil {
		u.log.Errorf("UpdateAddress err: %v", err)
		return err
	}
	return nil
}

// DeleteAddress 删除收货地址
func (u *UserUseCase) DeleteAddress(ctx context.Context, id int64, userID int64) error {
	if err := u.repo.DeleteAddress(ctx, id, userID); err != nil {
		u.log.Errorf("DeleteAddress err: %v", err)
		return err
	}
	return nil
}

// SetDefaultAddress 设置默认收货地址
func (u *UserUseCase) SetDefaultAddress(ctx context.Context, id int64, userID int64) error {
	if err := u.repo.SetDefaultAddress(ctx, id, userID); err != nil {
		u.log.Errorf("SetDefaultAddress err: %v", err)
		return err
	}
	return nil
}
