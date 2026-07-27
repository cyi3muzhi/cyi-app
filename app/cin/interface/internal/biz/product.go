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
	Stock       int32
	Category    string
	CategoryID  int64
	Status      int32
	Image       string
	Sales       int32
	Images      []string
	Specs       []*SpecGroupInfo
	Skus        []*SkuInfo
}

type SpecGroupInfo struct {
	ID     int64
	Name   string
	Values []string
}

type SkuInfo struct {
	ID       int64
	SkuCode  string
	Price    int64
	Stock    int32
	Sales    int32
	Spec     map[string]string
	Image    string
}

type ListProductRequest struct {
	Page       int32
	PageSize   int32
	Keyword    string
	CategoryID int64
	Status     string
}

type ProductRepo interface {
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
	ListProduct(ctx context.Context, req *ListProductRequest) ([]*Product, int32, error)
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

func (u *ProductUseCase) ListProduct(ctx context.Context, req *ListProductRequest) ([]*Product, int32, error) {
	return u.repo.ListProduct(ctx, req)
}
