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

// Refund 退款单（用户端展示）
type Refund struct {
	ID           int64  `json:"id"`            // 退款单ID
	OrderID      int64  `json:"order_id"`      // 订单ID
	OrderNo      string `json:"order_no"`      // 订单号
	RefundAmount int64  `json:"refund_amount"` // 退款金额（分）
	Reason       string `json:"reason"`        // 退款原因
	Status       int32  `json:"status"`        // 状态: 1-待审核 2-已通过 3-已拒绝
	StatusText   string `json:"status_text"`   // 状态文本
	RejectReason string `json:"reject_reason"` // 拒绝原因
	CreatedAt    int64  `json:"created_at"`    // 申请时间（Unix 秒）
}

// LogisticsTrace 物流轨迹节点（用户端展示）
type LogisticsTrace struct {
	Time     int64  // 节点时间（Unix 秒）
	TimeText string // 节点时间（格式化字符串）
	Desc     string // 节点描述
	Status   int32  // 节点状态: 1-已发货 2-运输中 3-派送中 4-已签收
}

// Logistics 物流信息（用户端展示）
type Logistics struct {
	OrderID          int64             // 订单ID
	LogisticsCompany string            // 物流公司
	LogisticsNo      string            // 快递单号
	ShipTime         int64             // 发货时间（Unix 秒，0=未发货）
	ShipTimeText     string            // 发货时间（格式化字符串）
	Traces           []*LogisticsTrace // 轨迹节点（按时间升序）
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
	// ApplyRefund 申请退款（校验订单归属，仅已支付/待发货订单可申请）
	ApplyRefund(ctx context.Context, orderID int64, userID int64, reason string) (int64, error)
	// ListRefund 查询当前用户退款记录（支持按状态筛选）
	ListRefund(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*Refund, int32, error)
	// GetLogistics 查询当前用户订单物流轨迹（校验订单归属）
	GetLogistics(ctx context.Context, orderID int64, userID int64) (*Logistics, error)
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

// ApplyRefund 申请退款（仅已支付/待发货订单可申请）
func (u *OrderUseCase) ApplyRefund(ctx context.Context, orderID int64, userID int64, reason string) (int64, error) {
	if orderID == 0 {
		return 0, v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.ApplyRefund(ctx, orderID, userID, reason)
}

// ListRefund 查询当前用户退款记录（支持按状态筛选）
func (u *OrderUseCase) ListRefund(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*Refund, int32, error) {
	return u.repo.ListRefund(ctx, userID, page, pageSize, status)
}

// GetLogistics 查询当前用户订单物流轨迹（校验订单归属）
func (u *OrderUseCase) GetLogistics(ctx context.Context, orderID int64, userID int64) (*Logistics, error) {
	if orderID == 0 {
		return nil, v1.ErrorOrderIdNotExist("order id not exist")
	}
	return u.repo.GetLogistics(ctx, orderID, userID)
}
