package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      int32
}

type ProductRepo interface {
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
}

type ProductUseCase struct {
	log  *log.Helper
	repo ProductRepo
}

func NewProductUseCase(repo ProductRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

func (u *ProductUseCase) CreateProduct(ctx context.Context, product *Product) error {
	return u.repo.CreateProduct(ctx, product)
}

func (u *ProductUseCase) UpdateProduct(ctx context.Context, product *Product) error {
	return u.repo.UpdateProduct(ctx, product)
}

func (u *ProductUseCase) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	return u.repo.GetProduct(ctx, productID)
}
