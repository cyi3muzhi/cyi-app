package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "cyi-app/api/service/product/v1"
)

type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
}

type ProductUseCase struct {
	repo      ProductRepo
	stockRepo StockRepo
	log       *log.Helper
}

func NewProductUseCase(repo ProductRepo, stockRepo StockRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{repo: repo, stockRepo: stockRepo, log: log.NewHelper(logger)}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.CreateProduct(ctx, in); err != nil {
		return v1.ErrorCreateProductFailed("create product failed")
	}
	in.InitStock(0)
	// 创建商品后初始化库存
	if err := uc.stockRepo.CreateStock(ctx, in.Stock); err != nil {
		return v1.ErrorCreateStockFailed("create stock failed")
	}
	return nil
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	return uc.repo.UpdateProduct(ctx, in)
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	return uc.repo.GetProduct(ctx, productID)
}
