package biz

import (
	"context"
	v1 "cyi-app/api/service/order/v1"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
	GetOrder(ctx context.Context, id int64) (*Order, error)
	// ListOrder 获取订单列表，支持按订单号模糊搜索、按状态筛选、按日期范围筛选，分页返回并按ID降序
	ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Order, int32, error)
	// OrderStats 统计订单数据：总订单数、待处理数、今日订单数、退款售后数
	OrderStats(ctx context.Context) (total, pending, today, refund int32, err error)
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

// Create 创建订单（两阶段：先校验全部商品，再统一预留库存）
func (uc *OrderUseCase) Create(ctx context.Context, in *Order) error {
	if len(in.Items) == 0 {
		return v1.ErrorOrderProductEmpty("order item is empty")
	}

	// 阶段一：校验全部商品（查商品上架、查库存充足），先不扣库存
	type stockInfo struct {
		SkuID int64
	}
	stockInfos := make([]stockInfo, 0, len(in.Items))

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

		// 从商品服务获取商品名称、价格和图片
		item.ProductName = product.Product.Name
		item.Price = product.Product.Price
		item.Image = product.Product.Image

		// 查找匹配的SKU，填充规格描述快照；SKU有独立图片则覆盖SPU主图
		if item.SkuID > 0 {
			for _, sku := range product.Product.Skus {
				if sku.Id == item.SkuID {
					item.SpecDesc = formatSpecDesc(sku.Spec)
					if item.Price == 0 {
						item.Price = sku.Price
					}
					if sku.Image != "" {
						item.Image = sku.Image
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
			uc.log.Errorf("insufficient stock for skuID %d: available=%d, required=%d", skuID, stock.Available, item.Quantity)
			return v1.ErrorInsufficientStock("insufficient stock")
		}

		stockInfos = append(stockInfos, stockInfo{
			SkuID: skuID,
		})
	}

	// 阶段二：全部校验通过，统一预留库存
	// 预留时重新读取最新版本号，避免阶段一记录的版本过期
	// 预留失败或库存不足时，释放已预留并返回对应错误
	var reserveErr error
	reserved := make([]int, 0, len(stockInfos))
	for i, si := range stockInfos {
		// 预留前重新读取最新库存
		latestStock, err := uc.stockRe.GetStock(ctx, si.SkuID)
		if err != nil {
			uc.log.Errorf("stock not found for skuID %d: %v", si.SkuID, err)
			reserveErr = v1.ErrorProductNotAvailable("stock not found")
			goto rollback
		}
		if latestStock.Available < in.Items[i].Quantity {
			uc.log.Errorf("insufficient stock for skuID %d: available=%d, required=%d", si.SkuID, latestStock.Available, in.Items[i].Quantity)
			reserveErr = v1.ErrorInsufficientStock("insufficient stock")
			goto rollback
		}
		if _, err = uc.stockRe.ReserveStock(ctx, si.SkuID, in.Items[i].Quantity, latestStock.Version); err != nil {
			uc.log.Errorf("reserve stock failed for skuID %d: %v", si.SkuID, err)
			reserveErr = v1.ErrorStockOperationFailed("reserve stock failed, please retry")
			goto rollback
		}
		reserved = append(reserved, i)
	}

	goto done

rollback:
	// 释放已成功预留的库存，失败时最多重试3次
	for _, ri := range reserved {
		skuID := stockInfos[ri].SkuID
		qty := in.Items[ri].Quantity
		for retry := 0; retry < 3; retry++ {
			s, e := uc.stockRe.GetStock(ctx, skuID)
			if e != nil {
				uc.log.Errorf("rollback get stock failed for skuID %d: %v", skuID, e)
				continue
			}
			if _, e = uc.stockRe.ReleaseStock(ctx, skuID, qty, s.Version); e != nil {
				uc.log.Errorf("rollback release stock failed for skuID %d (retry %d): %v", skuID, retry, e)
				continue
			}
			uc.log.Infof("rollback release stock success for skuID %d, qty %d", skuID, qty)
			break
		}
	}
	return reserveErr

done:

	// 3. 重新计算订单总金额
	in.TotalPrice = Money{Amount: 0, Currency: "CNY"}
	for _, item := range in.Items {
		in.TotalPrice = in.TotalPrice.Add(item.SubTotal())
	}

	// 4. 计算实付金额
	in.PayAmount = in.TotalPrice.Add(in.Freight).Subtract(in.Discount)

	// 5. 设置状态并创建订单
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo, uc.stockRe, uc.logger)
	if err := uc.repo.CreateOrder(ctx, in); err != nil {
		// Saga 补偿：订单创建失败，释放已预留库存
		uc.log.Errorf("order creation failed after stock reserved, executing saga compensation: %v", err)
		for _, ri := range reserved {
			skuID := stockInfos[ri].SkuID
			qty := in.Items[ri].Quantity
			for retry := 0; retry < 3; retry++ {
				s, e := uc.stockRe.GetStock(ctx, skuID)
				if e != nil {
					uc.log.Errorf("saga compensate get stock failed for skuID %d: %v", skuID, e)
					continue
				}
				if _, e = uc.stockRe.ReleaseStock(ctx, skuID, qty, s.Version); e != nil {
					uc.log.Errorf("saga compensate release stock failed for skuID %d (retry %d): %v", skuID, retry, e)
					continue
				}
				uc.log.Infof("saga compensate release stock success for skuID %d, qty %d", skuID, qty)
				break
			}
		}
		return err
	}
	return nil
}

// Update 更新订单
func (uc *OrderUseCase) Update(ctx context.Context, in *Order) error {
	if in.statusHandler == nil {
		in.SetStatus(in.Status, uc.repo, uc.stockRe, uc.logger)
	}
	return uc.repo.UpdateOrder(ctx, in)
}

// List 获取订单列表
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*Order, int32, error) {
	return uc.repo.ListOrder(ctx, page, pageSize, status, keyword, dateFrom, dateTo)
}

// Stats 获取订单统计数据
func (uc *OrderUseCase) Stats(ctx context.Context) (total, pending, today, refund int32, err error) {
	return uc.repo.OrderStats(ctx)
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

// formatSpecDesc 将规格 map 转换为可读字符串
// 如 {"颜色":"红色","尺码":"XL"} → "红色/XL"
func formatSpecDesc(spec map[string]string) string {
	if len(spec) == 0 {
		return ""
	}
	vals := make([]string, 0, len(spec))
	for _, v := range spec {
		vals = append(vals, v)
	}
	return strings.Join(vals, "/")
}
