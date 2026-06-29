package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type Stock struct {
	ProductID int64
	Total     int32
	Available int32
	Reserved  int32
	Version   int32
}

type StockRepo interface {
	UpdateStock(ctx context.Context, stock *Stock) error
	ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
	DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
	GetStock(ctx context.Context, productID int64) (*Stock, error)
}

type StockUseCase struct {
	log  *log.Helper
	repo StockRepo
}

func NewStockUseCase(repo StockRepo, logger log.Logger) *StockUseCase {
	return &StockUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (u *StockUseCase) UpdateStock(ctx context.Context, stock *Stock) error {
	return u.repo.UpdateStock(ctx, stock)
}

func (u *StockUseCase) ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	return u.repo.ReserveStock(ctx, productID, count, version)
}

func (u *StockUseCase) DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	return u.repo.DeductStock(ctx, productID, count, version)
}

func (u *StockUseCase) GetStock(ctx context.Context, productID int64) (*Stock, error) {
	return u.repo.GetStock(ctx, productID)
}
