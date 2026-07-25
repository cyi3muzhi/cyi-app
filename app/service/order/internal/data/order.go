package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"
	"errors"
	"strconv"
	"strings"
	"time"

	marketingRpc "cyi-app/api/service/marketing/v1"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// parseInt32 将字符串转换为 int32，转换失败时返回错误
func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

type orderRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{data: data, log: log.NewHelper(logger)}
}

func (r *orderRepo) CreateOrder(ctx context.Context, in *biz.Order) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		order := &mod.Order{
			UserID:         in.UserID,
			OrderNo:        in.OrderNo,
			TotalPrice:     in.TotalPrice.Amount,
			PayAmount:      in.PayAmount.Amount,
			Freight:        in.Freight.Amount,
			DiscountAmount: in.Discount.Amount,
			Status:         int32(in.Status),
			CouponRecordID: in.CouponRecordID,
		}
		if err := tx.Order.WithContext(ctx).Create(order); err != nil {
			return errors.New("create order failed")
		}
		in.ID = order.ID

		// 如果有优惠券，调用 marketing 服务的 UseCoupon 核销并更新优惠金额
		if in.CouponRecordID > 0 {
			reply, err := r.data.mc.UseCoupon(ctx, &marketingRpc.UseCouponRequest{
				RecordId:    in.CouponRecordID,
				UserId:      in.UserID,
				OrderId:     order.ID,
				OrderAmount: order.TotalPrice,
			})
			if err != nil {
				r.log.Errorf("use coupon failed, recordID=%d, orderID=%d: %v", in.CouponRecordID, order.ID, err)
				return errors.New("use coupon failed")
			}
			discountAmount := reply.DiscountAmount
			// 更新数据库中的优惠金额和实付金额
			order.PayAmount = order.TotalPrice + order.Freight - discountAmount
			order.DiscountAmount = discountAmount
			if _, err := tx.Order.WithContext(ctx).Where(tx.Order.ID.Eq(order.ID)).
				UpdateColumn(tx.Order.PayAmount, order.PayAmount); err != nil {
				return errors.New("update order pay_amount failed")
			}
			if _, err := tx.Order.WithContext(ctx).Where(tx.Order.ID.Eq(order.ID)).
				UpdateColumn(tx.Order.DiscountAmount, order.DiscountAmount); err != nil {
				return errors.New("update order discount_amount failed")
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
			if err := tx.OrderItem.WithContext(ctx).Create(oi); err != nil {
				return errors.New("create order item failed")
			}
		}
		return nil
	})
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

// ListOrder 实现订单列表查询，支持按订单号模糊搜索、按状态筛选（支持逗号分隔多状态）、按日期范围筛选，分页返回并按ID降序
func (r *orderRepo) ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string, dateFrom, dateTo string) ([]*biz.Order, int32, error) {
	// 构建查询条件
	do := curd.Order.WithContext(ctx)

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

	// 分页查询，按ID降序
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
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
		Where(curd.Order.Status.In(1, 2, 3)).
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
		Where(curd.Order.Status.In(8, 9)).
		Count()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	refund = int32(refundCount)

	return total, pending, today, refund, nil
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
	return out, nil
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
