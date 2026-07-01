package biz

import "context"

type StockRepo interface {
	CreateStock(ctx context.Context, in *Stock) error
	UpdateStock(ctx context.Context, in *Stock) error
	ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
	DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error)
	GetStock(ctx context.Context, productID int64) (*Stock, error)
}

type StockUseCase struct {
	repo StockRepo
}

func NewStockUseCase(repo StockRepo) *StockUseCase {
	return &StockUseCase{repo: repo}
}

func (uc *StockUseCase) CreateStock(ctx context.Context, in *Stock) error {
	return uc.repo.CreateStock(ctx, in)
}

func (uc *StockUseCase) UpdateStock(ctx context.Context, in *Stock) error {
	return uc.repo.UpdateStock(ctx, in)
}

func (uc *StockUseCase) ReserveStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	return uc.repo.ReserveStock(ctx, productID, count, version)
}

func (uc *StockUseCase) DeductStock(ctx context.Context, productID int64, count int32, version int32) (int32, error) {
	return uc.repo.DeductStock(ctx, productID, count, version)
}

func (uc *StockUseCase) GetStock(ctx context.Context, productID int64) (*Stock, error) {
	return uc.repo.GetStock(ctx, productID)
}
