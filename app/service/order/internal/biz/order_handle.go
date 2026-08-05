package biz

import (
	"context"
	"sort"
	"strings"
	"time"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// UserOrderStats 用户订单统计（个人中心使用）
type UserOrderStats struct {
	TotalOrders    int64 // 总订单数（不含已取消）
	PendingPay     int64 // 待付款
	PendingShip    int64 // 待发货
	PendingReceive int64 // 待收货
	Complete       int64 // 已完成/待评价
	RefundCount    int64 // 退款/售后
	TotalSpent     int64 // 累计消费（已完成订单实付金额之和，单位：分）
}

// PaymentLog 支付流水记录（Pay 成功后写入，供对账/售后追溯）
type PaymentLog struct {
	OrderID       int64
	PayType       string // 支付方式（当前为模拟支付 mock）
	TransactionID string // 第三方支付单号（模拟支付为空）
	Amount        int64  // 支付金额（单位：分）
	PayStatus     int32  // 支付状态：2=已支付
	PayMsg        string // 支付结果描述
}

// TrendItem 订单趋势项（管理端仪表盘按日统计）
type TrendItem struct {
	Date       string // 日期（MM-DD）
	OrderCount int32  // 当日订单数
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
	GetOrder(ctx context.Context, id int64) (*Order, error)
	// BatchGetOrder 批量获取订单详情（含商品明细），1 次查询替代 N 次 GetOrder
	BatchGetOrder(ctx context.Context, ids []int64) ([]*Order, error)
	// ListOrder 获取订单列表，支持按订单号模糊搜索、按状态筛选、按日期范围筛选、按用户过滤，分页返回并按ID降序
	ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string, userID int64) ([]*Order, int32, error)
	// OrderStats 统计订单数据：总订单数、待处理数、今日订单数、退款售后数
	OrderStats(ctx context.Context) (total, pending, today, refund int32, err error)
	// GetOrderTrend 获取近 N 日每日订单数（管理端仪表盘，按创建时间分组）
	GetOrderTrend(ctx context.Context, days int32) ([]*TrendItem, error)
	// GetOrderRevenueStats 获取今日/昨日营收（已完成订单实付金额之和，单位：分）
	GetOrderRevenueStats(ctx context.Context) (todayRevenue, yesterdayRevenue int64, err error)
	// GetUserOrderStats 统计单个用户的订单数据（个人中心使用）
	GetUserOrderStats(ctx context.Context, userID int64) (*UserOrderStats, error)
	// UpdateLogistics 更新订单物流信息（发货时写入物流公司、快递单号和发货时间）
	UpdateLogistics(ctx context.Context, orderID int64, logisticsCompany, logisticsNo string, shipTime int64) error
	// GetLogistics 查询订单物流信息（含轨迹节点）
	GetLogistics(ctx context.Context, orderID int64) (*Logistics, error)
	// UpdateStatusIf 条件更新订单状态（乐观并发控制）：仅当当前状态为 expectStatus 时更新为 newStatus
	// 解决支付/取消并发时的竞态：落库状态转换原子化，行数=0 表示状态已被并发变更
	UpdateStatusIf(ctx context.Context, id int64, expectStatus, newStatus int32) error
	// CreatePaymentLog 写入支付流水（Pay 成功后调用）
	CreatePaymentLog(ctx context.Context, log *PaymentLog) error
	// ReleaseCoupon 返还优惠券（订单取消/退款时调用，解除优惠券核销，经 marketing gRPC）
	ReleaseCoupon(ctx context.Context, recordID, userID, orderID int64) error
	// ListOverdueWaitPayOrders 分页查询超时未支付的待付款订单（status=1 且创建时间早于 now-timeoutSeconds）
	ListOverdueWaitPayOrders(ctx context.Context, page, pageSize int32, timeoutSeconds int64) ([]*Order, int32, error)
}

type OrderUseCase struct {
	repo      OrderRepo
	productRe ProductRepo
	stockRe   StockRepo
	taskRe    OrderTimeoutTaskRepo
	log       *log.Helper
	logger    log.Logger
}

func NewOrderUseCase(repo OrderRepo, productRe ProductRepo, stockRe StockRepo, taskRe OrderTimeoutTaskRepo, logger log.Logger) *OrderUseCase {
	return &OrderUseCase{repo: repo, productRe: productRe, stockRe: stockRe, taskRe: taskRe, log: log.NewHelper(logger), logger: logger}
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

	// 6. 投递订单超时取消延迟任务（30 分钟未支付自动取消）
	// 投递失败不影响下单主流程：记录日志即可，消费端会先校验订单状态保证幂等
	if err := uc.taskRe.EnqueueTimeoutCancel(ctx, in.ID); err != nil {
		uc.log.Errorf("投递订单超时任务失败, orderID: %d: %v", in.ID, err)
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
func (uc *OrderUseCase) List(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string, userID int64) ([]*Order, int32, error) {
	return uc.repo.ListOrder(ctx, page, pageSize, status, keyword, dateFrom, dateTo, userID)
}

// Stats 获取订单统计数据
func (uc *OrderUseCase) Stats(ctx context.Context) (total, pending, today, refund int32, err error) {
	return uc.repo.OrderStats(ctx)
}

// OrderTrend 获取近 N 日订单趋势（管理端仪表盘）
func (uc *OrderUseCase) OrderTrend(ctx context.Context, days int32) ([]*TrendItem, error) {
	return uc.repo.GetOrderTrend(ctx, days)
}

// RevenueStats 获取今日/昨日营收（管理端仪表盘）
func (uc *OrderUseCase) RevenueStats(ctx context.Context) (todayRevenue, yesterdayRevenue int64, err error) {
	return uc.repo.GetOrderRevenueStats(ctx)
}

// UserStats 获取单个用户的订单统计（个人中心使用）
func (uc *OrderUseCase) UserStats(ctx context.Context, userID int64) (*UserOrderStats, error) {
	return uc.repo.GetUserOrderStats(ctx, userID)
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

// BatchGet 批量查询订单详情（含商品明细），用于导出等场景
func (uc *OrderUseCase) BatchGet(ctx context.Context, ids []int64) ([]*Order, error) {
	return uc.repo.BatchGetOrder(ctx, ids)
}

// Pay 支付订单（先扣库存，后改状态，避免已支付未扣库存的数据不一致）
// 幂等：已支付（status=2/3）订单重复支付直接返回成功，避免重复扣库存
func (uc *OrderUseCase) Pay(ctx context.Context, orderID int64, userID int64) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.VerifyOwner(userID); err != nil {
		return err
	}

	// 幂等：已支付/待发货状态重复支付直接返回成功
	if order.Status == v1.OrderStatus_Pay || order.Status == v1.OrderStatus_WaitShip {
		uc.log.Infof("订单已支付，重复支付幂等跳过, orderID: %d", orderID)
		return nil
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

	// 2. 再更新订单状态（条件更新保证原子性；失败说明状态已被并发变更，如超时取消，需补偿库存）
	if err := order.Pay(ctx); err != nil {
		uc.log.Errorf("order pay failed after stock deducted, compensating stock, orderID: %d: %v", orderID, err)
		uc.compensateDeductedStock(ctx, order)
		return err
	}

	// 3. 写入支付流水（供对账/售后追溯；失败仅记日志不影响支付结果）
	if err := uc.repo.CreatePaymentLog(ctx, &PaymentLog{
		OrderID:   orderID,
		PayType:   "mock", // 当前为模拟支付，后续接入微信/支付宝时替换
		Amount:    order.PayAmount.Amount,
		PayStatus: 2, // 已支付
		PayMsg:    "支付成功",
	}); err != nil {
		uc.log.Errorf("create payment log failed, orderID: %d: %v", orderID, err)
	}
	return nil
}

// compensateDeductedStock 支付状态更新失败时，补偿已扣减的库存（AddStock 为 DeductStock 的逆操作）
func (uc *OrderUseCase) compensateDeductedStock(ctx context.Context, order *Order) {
	for _, item := range order.Items {
		skuID := item.SkuID
		for retry := 0; retry < 3; retry++ {
			s, e := uc.stockRe.GetStock(ctx, skuID)
			if e != nil {
				uc.log.Errorf("compensate get stock failed for skuID %d: %v", skuID, e)
				continue
			}
			if _, e = uc.stockRe.AddStock(ctx, skuID, item.Quantity, s.Version); e != nil {
				uc.log.Errorf("compensate add stock failed for skuID %d (retry %d): %v", skuID, retry, e)
				continue
			}
			uc.log.Infof("compensate add stock success for skuID %d, qty %d", skuID, item.Quantity)
			break
		}
	}
}

// Ship 发货（管理端调用）：状态流转（待发货 -> 已发货）+ 写入物流信息和发货时间
func (uc *OrderUseCase) Ship(ctx context.Context, orderID int64, logisticsCompany, logisticsNo string) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.Ship(ctx); err != nil {
		uc.log.Errorf("order ship failed, orderID: %d: %v", orderID, err)
		return err
	}
	return uc.repo.UpdateLogistics(ctx, orderID, logisticsCompany, logisticsNo, time.Now().Unix())
}

// GetLogistics 查询订单物流信息（含轨迹节点，按发货时间动态推算）
func (uc *OrderUseCase) GetLogistics(ctx context.Context, orderID int64) (*Logistics, error) {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}
	logistics, err := uc.repo.GetLogistics(ctx, orderID)
	if err != nil {
		return nil, err
	}
	// 未发货时直接返回（轨迹为空）
	if logistics.ShipTime == 0 {
		return logistics, nil
	}
	// 订单已完成时强制补全全部轨迹节点（含签收）
	logistics.Traces = BuildLogisticsTraces(logistics.ShipTime, time.Now().Unix(), order.Status == v1.OrderStatus_Complete)
	return logistics, nil
}

// Receive 确认收货（用户端调用）：状态流转（已发货/待收货 -> 已完成）
func (uc *OrderUseCase) Receive(ctx context.Context, orderID int64, userID int64) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if err := order.VerifyOwner(userID); err != nil {
		return err
	}
	if err := order.Receive(ctx); err != nil {
		uc.log.Errorf("order receive failed, orderID: %d: %v", orderID, err)
		return err
	}
	return nil
}

// Cancel 取消订单（用户端/管理端调用）：待支付/已支付 -> 已取消
// userID > 0 时校验订单归属（用户端），userID = 0 表示管理端操作
// 取消成功后自动返还下单时核销的优惠券
func (uc *OrderUseCase) Cancel(ctx context.Context, orderID int64, userID int64, reason string) error {
	order, err := uc.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if userID > 0 {
		if err := order.VerifyOwner(userID); err != nil {
			return err
		}
	}
	if reason != "" {
		uc.log.Infof("cancel order, orderID=%d, userID=%d, reason=%s", orderID, userID, reason)
	}
	if err := order.Cancel(ctx); err != nil {
		uc.log.Errorf("order cancel failed, orderID: %d: %v", orderID, err)
		return err
	}
	// 取消成功后返还优惠券（若下单时核销过；失败仅记日志，可由人工/定时任务补偿）
	if order.CouponRecordID > 0 {
		if err := uc.repo.ReleaseCoupon(ctx, order.CouponRecordID, order.UserID, orderID); err != nil {
			uc.log.Errorf("release coupon failed, orderID: %d, recordID: %d: %v", orderID, order.CouponRecordID, err)
		}
	}
	return nil
}

// OverdueWaitPayOrders 分页查询超时未支付的待付款订单（task 服务 cron 兜底扫描用）
// timeoutSeconds 为超时阈值：创建时间早于 now-timeoutSeconds 且仍为待付款的订单
func (uc *OrderUseCase) OverdueWaitPayOrders(ctx context.Context, page, pageSize int32, timeoutSeconds int64) ([]*Order, int32, error) {
	return uc.repo.ListOverdueWaitPayOrders(ctx, page, pageSize, timeoutSeconds)
}

// formatSpecDesc 将规格 map 转换为可读字符串
// 如 {"颜色":"红色","尺码":"XL"} → "红色/XL"
// 先对 key 排序再拼接，避免 Go map 遍历顺序随机导致规格顺序跳动
func formatSpecDesc(spec map[string]string) string {
	if len(spec) == 0 {
		return ""
	}
	keys := make([]string, 0, len(spec))
	for k := range spec {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]string, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, spec[k])
	}
	return strings.Join(vals, "/")
}
