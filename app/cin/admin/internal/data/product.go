package data

import (
	"context"

	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type productRepo struct {
	log  *log.Helper
	data *Data
}

// NewProductRepo 创建商品管理数据仓库，返回 mock 数据
func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}

// ListProduct 获取商品列表（mock 数据）
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword, category, status string) ([]*biz.Product, int32, error) {
	items := []*biz.Product{
		{ID: 1, Name: "iPhone 15 Pro", Category: "电子产品", Price: 899900, Stock: 128, Sales: 2356, Status: "上架", Image: "https://picsum.photos/200/200?random=1", SKU: "IP15P-256-BK"},
		{ID: 2, Name: "MacBook Air M3", Category: "电子产品", Price: 1099900, Stock: 56, Sales: 1892, Status: "上架", Image: "https://picsum.photos/200/200?random=2", SKU: "MBA-M3-512-SV"},
		{ID: 3, Name: "AirPods Pro 2", Category: "电子产品", Price: 189900, Stock: 320, Sales: 4568, Status: "上架", Image: "https://picsum.photos/200/200?random=3", SKU: "APP2-WL"},
		{ID: 4, Name: "纯棉T恤", Category: "服装鞋帽", Price: 12900, Stock: 500, Sales: 3280, Status: "上架", Image: "https://picsum.photos/200/200?random=4", SKU: "CT-S-WH"},
		{ID: 5, Name: "运动跑鞋", Category: "服装鞋帽", Price: 45900, Stock: 200, Sales: 1560, Status: "下架", Image: "https://picsum.photos/200/200?random=5", SKU: "RN-BLK-42"},
	}
	return items, int32(len(items)), nil
}

// CreateProduct 创建商品（mock 实现）
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) (int64, error) {
	in.ID = 100
	return 100, nil
}

// UpdateProduct 更新商品（mock 实现）
func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	return nil
}

// DeleteProduct 删除商品（mock 实现）
func (r *productRepo) DeleteProduct(ctx context.Context, id int64) error {
	return nil
}

// GetProduct 获取商品详情（mock 数据）
func (r *productRepo) GetProduct(ctx context.Context, id int64) (*biz.Product, error) {
	return &biz.Product{
		ID:          id,
		Name:        "iPhone 15 Pro",
		Category:    "电子产品",
		Price:       899900,
		Stock:       128,
		Sales:       2356,
		Status:      "上架",
		Image:       "https://picsum.photos/200/200?random=1",
		SKU:         "IP15P-256-BK",
		Description: "Apple iPhone 15 Pro 256GB 黑色钛金属",
	}, nil
}

// BatchOperation 批量操作商品（mock 实现）
func (r *productRepo) BatchOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	return int32(len(ids)), nil
}
