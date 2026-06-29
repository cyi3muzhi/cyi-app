package biz

import (
	"context"
)

type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
}

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Status      int32
}

type ProductUseCase struct {
	repo ProductRepo
}

func NewProductUseCase(repo ProductRepo) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	return uc.repo.CreateProduct(ctx, in)
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	return uc.repo.UpdateProduct(ctx, in)
}
