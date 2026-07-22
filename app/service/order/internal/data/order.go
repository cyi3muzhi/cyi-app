package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"
	"errors"
	"strconv"

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
		}
		if err := tx.Order.WithContext(ctx).Create(order); err != nil {
			return errors.New("create order failed")
		}
		in.ID = order.ID

		for _, item := range in.Items {
			oi := &mod.OrderItem{
				OrderID:     order.ID,
				ProductID:   item.ProductID,
				SkuID:       item.SkuID,
				SpecDesc:    item.SpecDesc,
				ProductName: item.ProductName,
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

// ListOrder 实现订单列表查询，支持按订单号模糊搜索、按状态筛选，分页返回并按ID降序
func (r *orderRepo) ListOrder(ctx context.Context, page, pageSize int32, status string, keyword string) ([]*biz.Order, int32, error) {
	// 构建查询条件
	do := curd.Order.WithContext(ctx)

	// 按状态筛选（status 为 proto 中 OrderStatus 的数字字符串）
	if status != "" {
		statusInt, err := parseInt32(status)
		if err == nil {
			do = do.Where(curd.Order.Status.Eq(statusInt))
		}
	}

	// 按订单号模糊搜索
	if keyword != "" {
		do = do.Where(curd.Order.OrderNo.Like("%" + keyword + "%"))
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
		Status: v1.OrderStatus(order.Status),
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
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return out, nil
}
