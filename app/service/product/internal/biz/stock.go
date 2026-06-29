package biz


import (
	"context"
)

type StockRepo interface {
	CreateStock(ctx context.Context, in *Stock) error
	UpdateStock(ctx context.Context, in *Stock) error
}


type Stock struct {
}


type StockUseCase struct {
	repo StockRepo
}


func NewStockUseCase(repo StockRepo) *StockUseCase {
	return &StockUseCase{repo: repo}
}


func (uc *StockUseCase) CreateStock(ctx context.Context, in *Stock) error {
	return nil
}


func (uc *StockUseCase) UpdateStock(ctx context.Context, in *Stock) error {	
	return nil
}