package biz

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// CartItem 购物车商品项
type CartItem struct {
	ProductID int64
	Quantity  int32
	SkuID     int64 // SKU ID
}

// OrderProduct 订单商品（列表展示）
type OrderProduct struct {
	ProductID int64
	SkuID     int64
	Name      string
	Spec      string
	Image     string
	Price     int64 // 单价（分）
	Quantity  int32
}

type Order struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	ProductID int64       `json:"product_id"`
	Count     int32       `json:"count"`
	Price     int64       `json:"price"`
	Status    int32       `json:"status"`
	SkuID     int64       `json:"sku_id"` // SKU ID
	Items     []*CartItem // 购物车结算：多个商品

	// ---- 列表/详情展示字段 ----
	OrderNo     string          `json:"order_no"`     // 订单号
	CreatedAt   int64           `json:"created_at"`   // 下单时间（Unix 秒）
	TotalAmount int64           `json:"total_amount"` // 订单总金额（分）
	PayAmount   int64           `json:"pay_amount"`   // 实付金额（分）
	Freight     int64           `json:"freight"`      // 运费（分）
	Discount    int64           `json:"discount"`     // 优惠金额（分）
	Products    []*OrderProduct `json:"products"`     // 商品明细

	// ---- 结算地址快照 ----
	AddressID int64  `json:"address_id"` // 收货地址ID
	Receiver  string `json:"receiver"`   // 收货人
	Phone     string `json:"phone"`      // 手机号
	Address   string `json:"address"`    // 完整收货地址

	// 优惠券发放记录ID（结算下单时使用，0=不使用优惠券）
	CouponRecordID int64 `json:"coupon_record_id"`

	// ---- 物流信息（发货后） ----
	LogisticsCompany string `json:"logistics_company"` // 物流公司
	LogisticsNo      string `json:"logistics_no"`      // 快递单号
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order) error
	PayOrder(ctx context.Context, orderID int64, userID int64) error
	// ReceiveOrder 确认收货（校验订单归属）
	ReceiveOrder(ctx context.Context, orderID int64, userID int64) error
	// CancelOrder 取消订单（校验订单归属）
	CancelOrder(ctx context.Context, orderID int64, userID int64, reason string) error
	GetOrder(ctx context.Context, orderID int64) (*Order, error)
	// ListOrder 查询当前用户订单列表（含商品明细）
	ListOrder(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*Order, int32, error)
	// GetOrderDetail 查询当前用户订单详情（含商品明细）
	GetOrderDetail(ctx context.Context, orderID int64, userID int64) (*Order, error)
}

type OrderUseCase struct {
	log  *log.Helper
	repo OrderRepo
}

func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (u *OrderUseCase) CreateOrder(ctx context.Context, order *Order) error {
	return u.repo.CreateOrder(ctx, order)
}

func (u *OrderUseCase) PayOrder(ctx context.Context, orderID int64, userID int64) error {
	if orderID == 0 {
		return v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.PayOrder(ctx, orderID, userID)
}

// ReceiveOrder 确认收货（校验订单归属）
func (u *OrderUseCase) ReceiveOrder(ctx context.Context, orderID int64, userID int64) error {
	if orderID == 0 {
		return v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.ReceiveOrder(ctx, orderID, userID)
}

// CancelOrder 取消订单（校验订单归属）
func (u *OrderUseCase) CancelOrder(ctx context.Context, orderID int64, userID int64, reason string) error {
	if orderID == 0 {
		return v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.CancelOrder(ctx, orderID, userID, reason)
}

func (u *OrderUseCase) GetOrder(ctx context.Context, orderID int64) (order *Order, err error) {
	return u.repo.GetOrder(ctx, orderID)
}

// ListOrder 查询当前用户订单列表
func (u *OrderUseCase) ListOrder(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*Order, int32, error) {
	return u.repo.ListOrder(ctx, userID, page, pageSize, status)
}

// GetOrderDetail 查询当前用户订单详情
func (u *OrderUseCase) GetOrderDetail(ctx context.Context, orderID int64, userID int64) (*Order, error) {
	if orderID == 0 {
		return nil, v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.GetOrderDetail(ctx, orderID, userID)
}
