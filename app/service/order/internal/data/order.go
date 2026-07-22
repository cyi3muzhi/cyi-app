package data

import (
	"context"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"
	"errors"

	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

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
			UserID:     in.UserID,
			OrderNo:    in.OrderNo,
			TotalPrice: in.TotalPrice.Amount,
			PayAmount:  in.PayAmount.Amount,
			Status:     int32(in.Status),
		}
		if err := tx.Order.WithContext(ctx).Create(order); err != nil {
			return errors.New("create order failed")
		}
		in.ID = order.ID

		for _, item := range in.Items {
			oi := &mod.OrderItem{
				OrderID:     order.ID,
				ProductID:   item.ProductID,
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
		UserID:     in.UserID,
		OrderNo:    in.OrderNo,
		TotalPrice: in.TotalPrice.Amount,
		PayAmount:  in.PayAmount.Amount,
		Status:     int32(in.Status),
	}
	order.ID = in.ID
	if line, err := curd.Order.Debug().WithContext(ctx).Where(curd.Order.ID.Eq(in.ID)).Updates(order); err != nil {
		return errors.New("update order failed")
	} else if line.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
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
		Status: v1.OrderStatus(order.Status),
	}

	items, err := curd.OrderItem.WithContext(ctx).Where(curd.OrderItem.OrderID.Eq(id)).Find()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		out.Items = append(out.Items, &biz.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}
	return out, nil
}
