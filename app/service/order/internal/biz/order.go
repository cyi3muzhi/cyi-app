package biz

import (
	"context"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, in *Order) error
	UpdateOrder(ctx context.Context, in *Order) error
}


type Order struct {
}


type OrderUseCase struct {
	repo OrderRepo
}


func NewOrderUseCase(repo OrderRepo) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}


func (uc *OrderUseCase) CreateOrder(ctx context.Context, in *Order) error {
	return nil
}


func (uc *OrderUseCase) UpdateOrder(ctx context.Context, in *Order) error {	
	return nil
}