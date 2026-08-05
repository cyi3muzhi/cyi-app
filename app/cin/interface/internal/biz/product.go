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
	// SuggestProduct 搜索联想：返回匹配关键词的商品名列表
	SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error)
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

// SuggestProduct 搜索联想：返回匹配关键词的商品名列表
// @param keyword 联想关键词
// @param limit 返回条数上限（默认 8，最大 20）
// @return 商品名称列表
func (u *ProductUseCase) SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error) {
	if keyword == "" {
		return nil, nil
	}
	// limit 默认值 8，上限 20
	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}
	return u.repo.SuggestProduct(ctx, keyword, limit)
}
