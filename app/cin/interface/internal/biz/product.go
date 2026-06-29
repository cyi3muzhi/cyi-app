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
}

type ProductRepo interface {
	CreateProduct(ctx context.Context, product *Product) error
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
