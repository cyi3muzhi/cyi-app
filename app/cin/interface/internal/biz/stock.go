package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type Stock struct {
	ProductID int64
	Total     int32
}

type StockRepo interface {
	UpdateStock(ctx context.Context, stock *Stock) error
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
