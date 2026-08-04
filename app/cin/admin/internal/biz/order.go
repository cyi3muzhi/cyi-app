package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Order 订单领域模型，匹配管理端 proto 中的 OrderItem
type Order struct {
	ID          int64
	OrderNo     string
	CreatedAt   string // 创建时间（格式化字符串）
	UserName    string // 用户名
	Amount      int64  // 订单金额（单位：分）
	Payment     string // 支付方式
	Status      string // 状态编码
	StatusText  string // 状态文本
	Products    []*OrderProduct
	Freight     int64 // 运费（单位：分）
	DiscountAmt int64 // 优惠金额（单位：分）

	// 物流信息（发货后展示）
	LogisticsCompany string // 物流公司
	LogisticsNo      string // 快递单号
	Receiver         string // 收货人
	Phone            string // 手机号
	Address          string // 完整收货地址
}

// OrderProduct 订单商品明细
type OrderProduct struct {
	Name  string
	Spec  string
	Qty   int32
	Price int64
	Image string
}

// OrderRepo 订单数据仓库接口
type OrderRepo interface {
	// ListOrder 获取订单列表
	ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Order, int32, error)
	// GetOrder 获取订单详情
	GetOrder(ctx context.Context, orderID int64) (*Order, error)
	// ShipOrder 发货
	ShipOrder(ctx context.Context, orderID int64, logisticsCompany string, logisticsNo string) error
	// CancelOrder 取消订单
	CancelOrder(ctx context.Context, orderID int64, reason string) error
	// OrderStats 获取订单统计数据
	OrderStats(ctx context.Context) (total, pending, today, refund int32, err error)
}

// OrderUseCase 订单管理业务用例
type OrderUseCase struct {
	log  *log.Helper
	repo OrderRepo
}

// NewOrderUseCase 创建订单管理业务用例
func NewOrderUseCase(repo OrderRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// List 获取订单列表
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Order, int32, error) {
	return uc.repo.ListOrder(ctx, page, pageSize, status, keyword, dateFrom, dateTo)
}

// Stats 获取订单统计数据
func (uc *OrderUseCase) Stats(ctx context.Context) (total, pending, today, refund int32, err error) {
	return uc.repo.OrderStats(ctx)
}

// Get 获取订单详情
func (uc *OrderUseCase) Get(ctx context.Context, orderID int64) (*Order, error) {
	return uc.repo.GetOrder(ctx, orderID)
}

// Ship 发货
func (uc *OrderUseCase) Ship(ctx context.Context, orderID int64, logisticsCompany string, logisticsNo string) error {
	return uc.repo.ShipOrder(ctx, orderID, logisticsCompany, logisticsNo)
}

// Cancel 取消订单
func (uc *OrderUseCase) Cancel(ctx context.Context, orderID int64, reason string) error {
	return uc.repo.CancelOrder(ctx, orderID, reason)
}
