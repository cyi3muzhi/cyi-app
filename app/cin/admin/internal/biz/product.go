package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Product 商品领域模型
type Product struct {
	ID          int64
	Name        string
	CategoryID  int64
	Price       int64
	Stock       int32
	Sales       int32
	Status      string
	Image       string
	Images      []string // 轮播图列表
	SKU         string
	Description string
	Specs       []*SpecGroup // 规格组列表
	Skus        []*SkuItem   // SKU 列表
}

// SpecGroup 规格组领域模型
type SpecGroup struct {
	ID     int64
	Name   string
	Values []string
}

// SkuItem SKU 领域模型
type SkuItem struct {
	ID      int64
	SkuCode string
	Price   int64
	Stock   int32
	Image   string
	Spec    map[string]string
}

// ProductRepo 商品数据仓库接口
type ProductRepo interface {
	// ListProduct 获取商品列表
	ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*Product, int32, error)
	// CreateProduct 创建商品
	CreateProduct(ctx context.Context, in *Product) (int64, error)
	// UpdateProduct 更新商品
	UpdateProduct(ctx context.Context, in *Product) error
	// DeleteProduct 删除商品
	DeleteProduct(ctx context.Context, id int64) error
	// GetProduct 获取商品详情
	GetProduct(ctx context.Context, id int64) (*Product, error)
	// BatchOperation 批量操作商品
	BatchOperation(ctx context.Context, action string, ids []int64) (int32, error)
}

// ProductUseCase 商品管理业务用例
type ProductUseCase struct {
	log  *log.Helper
	repo ProductRepo
}

// NewProductUseCase 创建商品管理业务用例
func NewProductUseCase(repo ProductRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{
		log:  log.NewHelper(logger),
		repo: repo,
	}
}

// ListProduct 获取商品列表
func (uc *ProductUseCase) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*Product, int32, error) {
	return uc.repo.ListProduct(ctx, page, pageSize, keyword, categoryID, status, stockEQ)
}

// CreateProduct 创建商品
func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) (int64, error) {
	return uc.repo.CreateProduct(ctx, in)
}

// UpdateProduct 更新商品
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	return uc.repo.UpdateProduct(ctx, in)
}

// DeleteProduct 删除商品
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id int64) error {
	return uc.repo.DeleteProduct(ctx, id)
}

// GetProduct 获取商品详情
func (uc *ProductUseCase) GetProduct(ctx context.Context, id int64) (*Product, error) {
	return uc.repo.GetProduct(ctx, id)
}

// BatchOperation 批量操作商品
func (uc *ProductUseCase) BatchOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	return uc.repo.BatchOperation(ctx, action, ids)
}
