package biz

import (
	"context"
)

type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
}


type Product struct {
}


type ProductUseCase struct {
	repo ProductRepo
}


func NewProductUseCase(repo ProductRepo) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}


func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	return nil
}


func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {	
	return nil
}