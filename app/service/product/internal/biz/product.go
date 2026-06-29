package biz

import (
	"context"
)

type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
}

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      int32
}

type ProductUseCase struct {
	repo      ProductRepo
	stockRepo StockRepo
}

func NewProductUseCase(repo ProductRepo, stockRepo StockRepo) *ProductUseCase {
	return &ProductUseCase{repo: repo, stockRepo: stockRepo}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.CreateProduct(ctx, in); err != nil {
		return err
	}
	// 创建商品后初始化库存
	return uc.stockRepo.CreateStock(ctx, &Stock{
		ProductID: in.ID,
	})
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	return uc.repo.UpdateProduct(ctx, in)
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	return uc.repo.GetProduct(ctx, productID)
}
