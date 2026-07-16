package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
	GetOrder(ctx context.Context, id int64) (*Order, error)
}

type OrderUseCase struct {
	repo      OrderRepo
	productRe ProductRepo
	stockRe   StockRepo
	log       *log.Helper
	logger    log.Logger
}

func NewOrderUseCase(repo OrderRepo, productRe ProductRepo, stockRe StockRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, productRe: productRe, stockRe: stockRe, log: log.NewHelper(logger), logger: logger}
}

// Create 创建订单（含商品校验、库存校验、预留库存）
func (uc *OrderUseCase) Create(ctx context.Context, in *Order) error {
	if len(in.Items) == 0 {
		return v1.ErrorOrderProductEmpty("order item is empty")
	}

	// 循环校验每个商品：查商品、查库存、预留库存
	for _, item := range in.Items {
		// 1. 查找商品是否存在且上架
		product, err := uc.productRe.GetProduct(ctx, item.ProductID)
		if err != nil {
			uc.log.Errorf("product not found: %v", err)
			return v1.ErrorProductNotAvailable("product not found")
		}
		if product.Product.Status != 1 {
			return v1.ErrorProductNotAvailable("product is off shelf")
		}

		// 从商品服务获取商品名称和价格
		item.ProductName = product.Product.Name
		item.Price = product.Product.Price

		// 2. 检查库存是否充足
		stock, err := uc.stockRe.GetStock(ctx, item.ProductID)
		if err != nil {
			uc.log.Errorf("stock not found: %v", err)
			return v1.ErrorProductNotAvailable("stock not found")
		}
		if stock.Available < item.Quantity {
			return v1.ErrorInsufficientStock("insufficient stock")
		}

		// 3. 预留库存（乐观锁）
		_, err = uc.stockRe.ReserveStock(ctx, item.ProductID, item.Quantity, stock.Version)
		if err != nil {
			uc.log.Errorf("reserve stock failed: %v", err)
			return v1.ErrorStockOperationFailed("reserve stock failed, please retry")
		}
	}

	// 4. 设置状态并创建订单
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo, uc.stockRe, uc.logger)
	return uc.repo.CreateOrder(ctx, in)
}

// Update 更新订单
func (uc *OrderUseCase) Update(ctx context.Context, in *Order) error {
	if in.statusHandler == nil {
		in.SetStatus(in.Status, uc.repo, uc.stockRe, uc.logger)
	}
	return uc.repo.UpdateOrder(ctx, in)
}

// Get 查询订单
func (uc *OrderUseCase) Get(ctx context.Context, orderID int64) (*Order, error) {
	order, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.SetStatus(order.Status, uc.repo, uc.stockRe, uc.logger)
	return order, nil
}

// Pay 支付订单（含扣减库存）
func (uc *OrderUseCase) Pay(ctx context.Context, orderID int64, userID int64) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.VerifyOwner(userID); err != nil {
		return err
	}

	// 1. 支付订单（状态流转）—— 先做，失败则库存不动
	if err := order.Pay(ctx); err != nil {
		return err
	}

	// 2. 实际扣减库存（Total-、Reserved-、乐观锁）
	for _, item := range order.Items {
		stock, err := uc.stockRe.GetStock(ctx, item.ProductID)
		if err != nil {
			uc.log.Errorf("stock not found: %v", err)
			return v1.ErrorProductNotAvailable("stock not found")
		}
		_, err = uc.stockRe.DeductStock(ctx, item.ProductID, item.Quantity, stock.Version)
		if err != nil {
			uc.log.Errorf("deduct stock failed, order already paid, need manual fix: %v", err)
			return v1.ErrorStockOperationFailed("deduct stock failed, already paid")
		}
	}

	return nil
}
