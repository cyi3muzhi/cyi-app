package data

import (
	"context"
	"cyi-app/app/rabbit/internal/biz"
	"fmt"
)

type orderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{
		data: data,
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context) error {
	fmt.Println("data CreateOrder")
	return nil
}
