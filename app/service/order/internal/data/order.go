package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	marketingRpc "cyi-app/api/service/marketing/v1"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// parseInt32 将字符串转换为 int32，转换失败时返回错误
func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

// normalizePagination 归一化分页参数：page 最小为 1，pageSize 默认 10、上限 100
// 先做边界校验再转 int，防止 int32 溢出或负数导致 offset 异常
func normalizePagination(page, pageSize int32) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return int(page-1) * int(pageSize), int(pageSize)
}

type orderRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{data: data, log: log.NewHelper(logger)}
}

func (r *orderRepo) CreateOrder(ctx context.Context, in *biz.Order) error {
	var couponUsed bool
	err := curd.Q.Transaction(func(tx *curd.Query) error {
		var txErr error
		couponUsed, txErr = createOrderTx(ctx, tx, in, r.data.mc, r.log)
		return txErr
	})
	if err != nil && couponUsed {
		// 事务回滚（订单不落库）但营销侧优惠券已核销：调用 marketing 服务 ReleaseCoupon 补偿释放
		// 补偿失败仅记录日志，可由取消订单流程/人工重试兜底
		if _, rerr := r.data.mc.ReleaseCoupon(ctx, &marketingRpc.ReleaseCouponRequest{
			RecordId: in.CouponRecordID,
			UserId:   in.UserID,
			OrderId:  in.ID,
		}); rerr != nil {
			r.log.Errorf("release coupon compensation failed, recordID=%d, orderID=%d: %v", in.CouponRecordID, in.ID, rerr)
		} else {
			r.log.Infof("release coupon compensation success, recordID=%d, orderID=%d", in.CouponRecordID, in.ID)
		}
	}
	return err
}

// createOrderTx 在指定事务内创建订单（订单主表 + 收货地址快照 + 优惠券核销 + 商品明细）
// 返回是否已核销优惠券，供调用方在事务回滚后对营销侧做补偿；供 CreateOrder 与 DTM TCC 订单分支复用
func createOrderTx(ctx context.Context, q *curd.Query, in *biz.Order, mc marketingRpc.CouponServiceClient, log *log.Helper) (couponUsed bool, err error) {
	// 幂等键：客户端未传时填充 auto-{orderNo}，保证 (user_id, idempotency_key) 唯一索引永不冲突
	idempotencyKey := in.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = "auto-" + in.OrderNo
	}
	order := &mod.Order{
		UserID:         in.UserID,
		OrderNo:        in.OrderNo,
		TotalPrice:     in.TotalPrice.Amount,
		PayAmount:      in.PayAmount.Amount,
		Freight:        in.Freight.Amount,
		DiscountAmount: in.Discount.Amount,
		Status:         int32(in.Status),
		CouponRecordID: in.CouponRecordID,
		IdempotencyKey: idempotencyKey,
	}
	if err := q.Order.WithContext(ctx).Create(order); err != nil {
		return false, errors.New("create order failed")
	}
	in.ID = order.ID

	// 收货地址快照写入物流表（结算下单时携带地址才写入）
	if in.Receiver != "" || in.Phone != "" || in.Address != "" {
		if err := q.OrderLogistics.WithContext(ctx).Create(&mod.OrderLogistics{
			OrderID:  order.ID,
			Receiver: in.Receiver,
			Phone:    in.Phone,
			Address:  in.Address,
		}); err != nil {
			log.Errorf("create order logistics failed, orderID=%d: %v", order.ID, err)
			return false, errors.New("create order logistics failed")
		}
	}

	// 如果有优惠券，调用 marketing 服务的 UseCoupon 核销并更新优惠金额
	if in.CouponRecordID > 0 {
		reply, err := mc.UseCoupon(ctx, &marketingRpc.UseCouponRequest{
			RecordId:    in.CouponRecordID,
			UserId:      in.UserID,
			OrderId:     order.ID,
			OrderAmount: order.TotalPrice,
		})
		if err != nil {
			log.Errorf("use coupon failed, recordID=%d, orderID=%d: %v", in.CouponRecordID, order.ID, err)
			return false, errors.New("use coupon failed")
		}
		couponUsed = true
		discountAmount := reply.DiscountAmount
		// 更新数据库中的优惠金额和实付金额
		order.PayAmount = order.TotalPrice + order.Freight - discountAmount
		order.DiscountAmount = discountAmount
		if _, err := q.Order.WithContext(ctx).Where(q.Order.ID.Eq(order.ID)).
			UpdateColumn(q.Order.PayAmount, order.PayAmount); err != nil {
			return false, errors.New("update order pay_amount failed")
		}
		if _, err := q.Order.WithContext(ctx).Where(q.Order.ID.Eq(order.ID)).
			UpdateColumn(q.Order.DiscountAmount, order.DiscountAmount); err != nil {
			return false, errors.New("update order discount_amount failed")
		}
		// 回写到 biz.Order，让调用方感知实际优惠和实付金额
		in.Discount.Amount = discountAmount
		in.PayAmount.Amount = order.PayAmount
	}

	for _, item := range in.Items {
		oi := &mod.OrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			SkuID:       item.SkuID,
			SpecDesc:    item.SpecDesc,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
		if err := q.OrderItem.WithContext(ctx).Create(oi); err != nil {
			return false, errors.New("create order item failed")
		}
	}
	return couponUsed, nil
}

// GetOrderByOrderNo 按订单号查询订单（order_no 唯一索引），DTM TCC 下单成功后回查订单 ID
func (r *orderRepo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*biz.Order, error) {
	order, err := curd.Order.WithContext(ctx).Where(curd.Order.OrderNo.Eq(orderNo)).First()
	if err != nil {
		return nil, err
	}
	if order == nil || order.ID == 0 {
		return nil, errors.New("order not found")
	}
	return &biz.Order{
		ID:      order.ID,
		UserID:  order.UserID,
		OrderNo: order.OrderNo,
		Status:  v1.OrderStatus(order.Status),
	}, nil
}

// GetOrderByIdempotencyKey 按客户端幂等键查询订单（(user_id, idempotency_key) 联合唯一索引）
// 用于下单去重：同一用户用相同幂等键重复提交时直接返回已创建的订单。
// 未命中返回 (nil, nil)；使用原生 GORM 查询（curd 生成代码不含新增的 idempotency_key 列）
func (r *orderRepo) GetOrderByIdempotencyKey(ctx context.Context, userID int64, key string) (*biz.Order, error) {
	var order mod.Order
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND idempotency_key = ?", userID, key).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未命中：首次下单，正常继续
		}
		r.log.Errorf("get order by idempotency key failed, userID=%d, key=%s: %v", userID, key, err)
		return nil, err
	}
	return &biz.Order{
		ID:             order.ID,
		UserID:         order.UserID,
		OrderNo:        order.OrderNo,
		IdempotencyKey: order.IdempotencyKey,
		Status:         v1.OrderStatus(order.Status),
	}, nil
}

func (r *orderRepo) UpdateOrder(ctx context.Context, in *biz.Order) error {
	order := &mod.Order{
		UserID:         in.UserID,
		OrderNo:        in.OrderNo,
		TotalPrice:     in.TotalPrice.Amount,
		PayAmount:      in.PayAmount.Amount,
		Freight:        in.Freight.Amount,
		DiscountAmount: in.Discount.Amount,
		Status:         int32(in.Status),
	}
	order.ID = in.ID
	if line, err := curd.Order.Debug().WithContext(ctx).Where(curd.Order.ID.Eq(in.ID)).Updates(order); err != nil {
		return errors.New("update order failed")
	} else if line.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

// ListOrder 实现订单列表查询，支持按订单号模糊搜索、按状态筛选（支持逗号分隔多状态）、按日期范围筛选、按用户过滤，分页返回并按ID降序
func (r *orderRepo) ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string, userID int64) ([]*biz.Order, int32, error) {
	// 构建查询条件
	do := curd.Order.WithContext(ctx)

	// 按用户过滤（>0 时生效，个人中心使用）
	if userID > 0 {
		do = do.Where(curd.Order.UserID.Eq(userID))
	}

	// 按状态筛选（status 支持逗号分隔的多状态，如 "2,3"）
	if status != "" {
		statuses := strings.Split(status, ",")
		var statusInts []int32
		for _, s := range statuses {
			if sInt, err := parseInt32(s); err == nil {
				statusInts = append(statusInts, sInt)
			}
		}
		if len(statusInts) > 0 {
			if len(statusInts) == 1 {
				do = do.Where(curd.Order.Status.Eq(statusInts[0]))
			} else {
				do = do.Where(curd.Order.Status.In(statusInts...))
			}
		}
	}

	// 按订单号模糊搜索
	if keyword != "" {
		do = do.Where(curd.Order.OrderNo.Like("%" + keyword + "%"))
	}

	// 按日期范围筛选（格式：2006-01-02，使用中国时区 CST）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if dateFrom != "" {
		fromTime, err := time.ParseInLocation("2006-01-02", dateFrom, loc)
		if err == nil {
			do = do.Where(curd.Order.CreateTime.Gte(fromTime.Unix()))
		}
	}
	if dateTo != "" {
		toTime, err := time.ParseInLocation("2006-01-02", dateTo, loc)
		if err == nil {
			toTime = toTime.Add(24*time.Hour - time.Second)
			do = do.Where(curd.Order.CreateTime.Lte(toTime.Unix()))
		}
	}

	// 查询总数
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询，按ID降序（page/pageSize 经 normalizePagination 归一化，防止溢出）
	offset, limit := normalizePagination(page, pageSize)
	orders, err := do.Offset(offset).Limit(limit).Order(curd.Order.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	// 转换为 biz.Order
	items := make([]*biz.Order, 0, len(orders))
	for _, order := range orders {
		items = append(items, &biz.Order{
			ID:      order.ID,
			UserID:  order.UserID,
			OrderNo: order.OrderNo,
			TotalPrice: biz.Money{
				Amount: order.TotalPrice,
			},
			PayAmount: biz.Money{
				Amount: order.PayAmount,
			},
			Freight: biz.Money{
				Amount: order.Freight,
			},
			Discount: biz.Money{
				Amount: order.DiscountAmount,
			},
			Status:     v1.OrderStatus(order.Status),
			CreateTime: order.CreateTime,
		})
	}
	return items, int32(total), nil
}

// OrderStats 统计订单数据：总订单数、待处理数、今日订单数、退款售后数
func (r *orderRepo) OrderStats(ctx context.Context) (total, pending, today, refund int32, err error) {
	// 总订单数
	totalCount, err := curd.Order.WithContext(ctx).Count()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	total = int32(totalCount)

	// 待处理数（待支付1 + 已支付2 + 待发货3）
	pendingCount, err := curd.Order.WithContext(ctx).
		Where(curd.Order.Status.In(
			int32(v1.OrderStatus_WaitPay),
			int32(v1.OrderStatus_Pay),
			int32(v1.OrderStatus_WaitShip),
		)).
		Count()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	pending = int32(pendingCount)

	// 今日订单数（基于 CreateTime Unix 时间戳，使用中国时区 CST）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	endToday := startToday + 24*60*60 - 1
	todayCount, err := curd.Order.WithContext(ctx).
		Where(curd.Order.CreateTime.Gte(startToday)).
		Where(curd.Order.CreateTime.Lte(endToday)).
		Count()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	today = int32(todayCount)

	// 退款/售后数（退款中8 + 已退款9）
	refundCount, err := curd.Order.WithContext(ctx).
		Where(curd.Order.Status.In(
			int32(v1.OrderStatus_Refunding),
			int32(v1.OrderStatus_Refunded),
		)).
		Count()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	refund = int32(refundCount)

	return total, pending, today, refund, nil
}

// GetOrderTrend 获取近 N 日每日订单数（管理端仪表盘）
// 按订单创建时间分组统计，缺失日期补 0，返回近 N 天数据（日期升序）
func (r *orderRepo) GetOrderTrend(ctx context.Context, days int32) ([]*biz.TrendItem, error) {
	if days <= 0 {
		days = 14
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -int(days)+1).Unix()

	// 按天分组统计订单数（原生 SQL 按日期分组）
	type dayCount struct {
		Day string
		Cnt int32
	}
	var rows []dayCount
	if err := curd.Order.WithContext(ctx).UnderlyingDB().Model(&mod.Order{}).
		Select("DATE_FORMAT(FROM_UNIXTIME(create_time), '%Y-%m-%d') AS day, COUNT(*) AS cnt").
		Where("create_time >= ? AND deleted_at IS NULL", start).
		Group("day").
		Order("day").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 组装近 N 天数据，无订单的日期补 0
	countMap := make(map[string]int32, len(rows))
	for _, row := range rows {
		countMap[row.Day] = row.Cnt
	}
	items := make([]*biz.TrendItem, 0, days)
	for i := 0; i < int(days); i++ {
		day := now.AddDate(0, 0, -int(days)+1+i)
		key := day.Format("2006-01-02")
		items = append(items, &biz.TrendItem{
			Date:       day.Format("01-02"),
			OrderCount: countMap[key],
		})
	}
	return items, nil
}

// GetOrderRevenueStats 获取今日/昨日营收（已完成订单实付金额之和，单位：分）
func (r *orderRepo) GetOrderRevenueStats(ctx context.Context) (todayRevenue, yesterdayRevenue int64, err error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	endToday := startToday + 24*60*60 - 1
	startYesterday := startToday - 24*60*60
	endYesterday := startToday - 1

	todayRevenue, err = r.revenueInRange(ctx, startToday, endToday)
	if err != nil {
		return 0, 0, err
	}
	yesterdayRevenue, err = r.revenueInRange(ctx, startYesterday, endYesterday)
	if err != nil {
		return 0, 0, err
	}
	return todayRevenue, yesterdayRevenue, nil
}

// revenueInRange 统计指定时间范围内已完成订单（status=6）的实付金额之和
func (r *orderRepo) revenueInRange(ctx context.Context, start, end int64) (int64, error) {
	var total int64
	if err := curd.Order.WithContext(ctx).UnderlyingDB().Model(&mod.Order{}).
		Select("COALESCE(SUM(pay_amount), 0)").
		Where("status = ? AND create_time >= ? AND create_time <= ? AND deleted_at IS NULL",
			int32(v1.OrderStatus_Complete), start, end).
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// GetUserOrderStats 统计单个用户的订单数据（个人中心使用）
// 各状态计数 + 累计消费（已完成订单实付金额之和）
func (r *orderRepo) GetUserOrderStats(ctx context.Context, userID int64) (*biz.UserOrderStats, error) {
	stats := &biz.UserOrderStats{}
	// 基础查询：按用户过滤（gen 查询链为克隆式，base 可复用）
	base := curd.Order.WithContext(ctx).Where(curd.Order.UserID.Eq(userID))

	// 总订单数（不含已取消）
	total, err := base.Where(curd.Order.Status.Neq(int32(v1.OrderStatus_Cancel))).Count()
	if err != nil {
		return nil, err
	}
	stats.TotalOrders = total

	// 待付款（status=1）
	pendingPay, err := base.Where(curd.Order.Status.Eq(int32(v1.OrderStatus_WaitPay))).Count()
	if err != nil {
		return nil, err
	}
	stats.PendingPay = pendingPay

	// 待发货（status=2 已支付 / 3 待发货）
	pendingShip, err := base.Where(curd.Order.Status.In(
		int32(v1.OrderStatus_Pay), int32(v1.OrderStatus_WaitShip))).Count()
	if err != nil {
		return nil, err
	}
	stats.PendingShip = pendingShip

	// 待收货（status=4 已发货 / 5 待收货）
	pendingReceive, err := base.Where(curd.Order.Status.In(
		int32(v1.OrderStatus_Ship), int32(v1.OrderStatus_WaitReceive))).Count()
	if err != nil {
		return nil, err
	}
	stats.PendingReceive = pendingReceive

	// 已完成/待评价（status=6）
	complete, err := base.Where(curd.Order.Status.Eq(int32(v1.OrderStatus_Complete))).Count()
	if err != nil {
		return nil, err
	}
	stats.Complete = complete

	// 退款/售后（status=8,9）
	refundCount, err := base.Where(curd.Order.Status.In(
		int32(v1.OrderStatus_Refunding), int32(v1.OrderStatus_Refunded))).Count()
	if err != nil {
		return nil, err
	}
	stats.RefundCount = refundCount

	// 累计消费（已完成订单实付金额之和；无已完成订单时 SUM 为 NULL，需用 NullInt64 兼容）
	var totalSpent sql.NullInt64
	if err := curd.Order.WithContext(ctx).
		Select(curd.Order.PayAmount.Sum()).
		Where(curd.Order.UserID.Eq(userID)).
		Where(curd.Order.Status.Eq(int32(v1.OrderStatus_Complete))).
		Scan(&totalSpent); err != nil {
		return nil, err
	}
	stats.TotalSpent = totalSpent.Int64

	return stats, nil
}

func (r *orderRepo) GetOrder(ctx context.Context, id int64) (*biz.Order, error) {
	order, err := curd.Order.WithContext(ctx).Where(curd.Order.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	if order == nil || order.ID == 0 {
		return nil, errors.New("order not found")
	}
	out := &biz.Order{
		ID:      order.ID,
		UserID:  order.UserID,
		OrderNo: order.OrderNo,
		TotalPrice: biz.Money{
			Amount: order.TotalPrice,
		},
		PayAmount: biz.Money{
			Amount: order.PayAmount,
		},
		Freight: biz.Money{
			Amount: order.Freight,
		},
		Discount: biz.Money{
			Amount: order.DiscountAmount,
		},
		Status:     v1.OrderStatus(order.Status),
		CreateTime: order.CreateTime,
	}

	items, err := curd.OrderItem.WithContext(ctx).Where(curd.OrderItem.OrderID.Eq(id)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		out.Items = append(out.Items, &biz.OrderItem{
			ProductID:   item.ProductID,
			SkuID:       item.SkuID,
			SpecDesc:    item.SpecDesc,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	// 查询物流信息（地址快照 + 发货后的物流公司和快递单号）
	logistics, err := curd.OrderLogistics.WithContext(ctx).
		Where(curd.OrderLogistics.OrderID.Eq(id)).First()
	if err == nil && logistics != nil {
		out.Receiver = logistics.Receiver
		out.Phone = logistics.Phone
		out.Address = logistics.Address
		out.LogisticsCompany = logistics.Company
		out.LogisticsNo = logistics.ExpressNo
		out.ShipTime = logistics.ShipTime
	}
	return out, nil
}

// UpdateLogistics 更新订单物流信息（发货时写入物流公司、快递单号和发货时间）
// 使用原生 GORM 更新（gen 生成的 curd 代码不包含新增的 Company 字段）
func (r *orderRepo) UpdateLogistics(ctx context.Context, orderID int64, logisticsCompany, logisticsNo string, shipTime int64) error {
	if err := r.data.db.Model(&mod.OrderLogistics{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{"company": logisticsCompany, "express_no": logisticsNo, "ship_time": shipTime}).Error; err != nil {
		r.log.Errorf("update logistics failed, orderID=%d: %v", orderID, err)
		return errors.New("update logistics failed")
	}
	return nil
}

// GetLogistics 查询订单物流信息（含发货时间）
// 使用原生 GORM 查询（gen 生成的 curd 代码不包含新增的 ShipTime 字段）
func (r *orderRepo) GetLogistics(ctx context.Context, orderID int64) (*biz.Logistics, error) {
	var logistics mod.OrderLogistics
	if err := r.data.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&logistics).Error; err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorOrderNotFound("order logistics not found")
		}
		r.log.Errorf("get logistics failed, orderID=%d: %v", orderID, err)
		return nil, errors.New("get logistics failed")
	}
	return &biz.Logistics{
		OrderID:          orderID,
		LogisticsCompany: logistics.Company,
		LogisticsNo:      logistics.ExpressNo,
		ShipTime:         logistics.ShipTime,
	}, nil
}

// BatchGetOrder 批量获取订单详情（含商品明细）
// 用 1 次 orders 表查询 + 1 次 order_items 表查询替代 N 次 GetOrder，消除 N+1 查询。
// 商品明细按 order_id 分组后回填到对应订单。
func (r *orderRepo) BatchGetOrder(ctx context.Context, ids []int64) ([]*biz.Order, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// 批量查询订单主表
	orders, err := curd.Order.WithContext(ctx).Where(curd.Order.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	// 批量查询所有订单的商品明细（1 次查询替代 N 次）
	items, err := curd.OrderItem.WithContext(ctx).Where(curd.OrderItem.OrderID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	// 按 order_id 分组商品明细
	itemsMap := make(map[int64][]*biz.OrderItem, len(orders))
	for _, item := range items {
		itemsMap[item.OrderID] = append(itemsMap[item.OrderID], &biz.OrderItem{
			ProductID:   item.ProductID,
			SkuID:       item.SkuID,
			SpecDesc:    item.SpecDesc,
			ProductName: item.ProductName,
			Image:       item.Image,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	// 组装返回（含商品明细）
	result := make([]*biz.Order, 0, len(orders))
	for _, order := range orders {
		result = append(result, &biz.Order{
			ID:         order.ID,
			UserID:     order.UserID,
			OrderNo:    order.OrderNo,
			TotalPrice: biz.Money{Amount: order.TotalPrice},
			PayAmount:  biz.Money{Amount: order.PayAmount},
			Freight:    biz.Money{Amount: order.Freight},
			Discount:   biz.Money{Amount: order.DiscountAmount},
			Status:     v1.OrderStatus(order.Status),
			CreateTime: order.CreateTime,
			Items:      itemsMap[order.ID],
		})
	}
	return result, nil
}

// UpdateStatusIf 条件更新订单状态（乐观并发控制）
// 仅当当前状态仍为 expectStatus 时更新为 newStatus，行数=0 说明状态已被并发变更（如支付时被超时取消）
func (r *orderRepo) UpdateStatusIf(ctx context.Context, id int64, expectStatus, newStatus int32) error {
	result, err := curd.Order.WithContext(ctx).
		Where(curd.Order.ID.Eq(id)).
		Where(curd.Order.Status.Eq(expectStatus)).
		UpdateColumn(curd.Order.Status, newStatus)
	if err != nil {
		r.log.Errorf("update order status failed, orderID=%d, expect=%d, new=%d: %v", id, expectStatus, newStatus, err)
		return errors.New("update order status failed")
	}
	if result.RowsAffected == 0 {
		// 状态已被并发变更，返回明确的业务错误供调用方补偿
		return v1.ErrorOrderStatusInvalid("订单状态已变更，操作失败, orderID=%d", id)
	}
	return nil
}

// CreatePaymentLog 写入支付流水
func (r *orderRepo) CreatePaymentLog(ctx context.Context, log *biz.PaymentLog) error {
	if err := curd.PaymentLog.WithContext(ctx).Create(&mod.PaymentLog{
		OrderID:       log.OrderID,
		PayType:       log.PayType,
		TransactionID: log.TransactionID,
		Amount:        log.Amount,
		PayStatus:     log.PayStatus,
		PayMsg:        log.PayMsg,
	}); err != nil {
		r.log.Errorf("create payment log failed, orderID=%d: %v", log.OrderID, err)
		return errors.New("create payment log failed")
	}
	return nil
}

// ReleaseCoupon 返还优惠券（订单取消/退款时调用，解除核销，经 marketing gRPC）
func (r *orderRepo) ReleaseCoupon(ctx context.Context, recordID, userID, orderID int64) error {
	_, err := r.data.mc.ReleaseCoupon(ctx, &marketingRpc.ReleaseCouponRequest{
		RecordId: recordID,
		UserId:   userID,
		OrderId:  orderID,
	})
	if err != nil {
		r.log.Errorf("release coupon failed, recordID=%d, orderID=%d: %v", recordID, orderID, err)
		return err
	}
	return nil
}

// ListOverdueWaitPayOrders 分页查询超时未支付的待付款订单（status=1 且创建时间早于 now-timeoutSeconds）
func (r *orderRepo) ListOverdueWaitPayOrders(ctx context.Context, page, pageSize int32, timeoutSeconds int64) ([]*biz.Order, int32, error) {
	deadline := time.Now().Unix() - timeoutSeconds
	do := curd.Order.WithContext(ctx).
		Where(curd.Order.Status.Eq(int32(v1.OrderStatus_WaitPay))). // 待付款（v1.OrderStatus_WaitPay=1）
		Where(curd.Order.CreateTime.Lt(deadline))                   // 创建时间早于超时阈值

	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询（page/pageSize 经 normalizePagination 归一化，防止溢出）
	offset, limit := normalizePagination(page, pageSize)
	orders, err := do.Offset(offset).Limit(limit).Order(curd.Order.ID.Desc()).Find()
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Order, 0, len(orders))
	for _, order := range orders {
		items = append(items, &biz.Order{
			ID:         order.ID,
			UserID:     order.UserID,
			OrderNo:    order.OrderNo,
			TotalPrice: biz.Money{Amount: order.TotalPrice},
			PayAmount:  biz.Money{Amount: order.PayAmount},
			Status:     v1.OrderStatus(order.Status),
			CreateTime: order.CreateTime,
		})
	}
	return items, int32(total), nil
}
