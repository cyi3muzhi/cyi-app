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
	// ListOrder 获取订单列表，支持按订单号模糊搜索和按状态筛选，分页返回并按ID降序
	ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string) ([]*Order, int32, error)
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

	// 循环校验每个商品：查商品、查SKU、查库存、预留库存
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

		// 查找匹配的SKU，填充规格描述快照
		if item.SkuID > 0 {
			for _, sku := range product.Product.Skus {
				if sku.Id == item.SkuID {
					item.SpecDesc = sku.Spec["spec_desc"]
					if item.Price == 0 {
						item.Price = sku.Price // SKU级别价格优先
					}
					break
				}
			}
		}

		// 2. 检查库存是否充足（按SKU维度）
		skuID := item.SkuID
		stock, err := uc.stockRe.GetStock(ctx, skuID)
		if err != nil {
			uc.log.Errorf("stock not found for skuID %d: %v", skuID, err)
			return v1.ErrorProductNotAvailable("stock not found")
		}
		if stock.Available < item.Quantity {
			return v1.ErrorInsufficientStock("insufficient stock")
		}

		// 3. 预留库存（乐观锁，按SKU维度）
		_, err = uc.stockRe.ReserveStock(ctx, skuID, item.Quantity, stock.Version)
		if err != nil {
			uc.log.Errorf("reserve stock failed for skuID %d: %v", skuID, err)
			return v1.ErrorStockOperationFailed("reserve stock failed, please retry")
		}
	}

	// 4. 计算实付金额：pay_amount = total_price + freight - discount
	in.PayAmount = in.TotalPrice.Add(in.Freight).Subtract(in.Discount)

	// 5. 设置状态并创建订单
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

// List 获取订单列表
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int32, status string, keyword string) ([]*Order, int32, error) {
	return uc.repo.ListOrder(ctx, page, pageSize, status, keyword)
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

// Pay 支付订单（先扣库存，后改状态，避免已支付未扣库存的数据不一致）
func (uc *OrderUseCase) Pay(ctx context.Context, orderID int64, userID int64) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.VerifyOwner(userID); err != nil {
		return err
	}

	// 1. 先扣减库存（Total-、Reserved-、乐观锁，按SKU维度）
	// 库存扣减失败时订单状态不变，调用方可安全重试
	for _, item := range order.Items {
		skuID := item.SkuID
		stock, err := uc.stockRe.GetStock(ctx, skuID)
		if err != nil {
			uc.log.Errorf("stock not found for skuID %d: %v", skuID, err)
			return v1.ErrorProductNotAvailable("stock not found")
		}
		_, err = uc.stockRe.DeductStock(ctx, skuID, item.Quantity, stock.Version)
		if err != nil {
			uc.log.Errorf("deduct stock failed for skuID %d: %v", skuID, err)
			return v1.ErrorStockOperationFailed("deduct stock failed")
		}
	}

	// 2. 再更新订单状态（库存已扣，状态变更失败需人工介入补偿库存）
	if err := order.Pay(ctx); err != nil {
		uc.log.Errorf("order pay failed after stock deducted, need manual compensation, orderID: %d: %v", orderID, err)
		return err
	}

	return nil
}
