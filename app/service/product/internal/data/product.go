package data

import (
	"context"
	"errors"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type productRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

// CreateProduct 创建商品
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	// 设置新增字段
	product.Category = in.Category
	product.Image = in.Image
	product.Sales = in.Sales
	if err := curd.Product.WithContext(ctx).Create(product); err != nil {
		return err
	}
	in.ID = product.ID
	return nil
}

// UpdateProduct 更新商品
func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	convert.Convert(in, product)
	product.ID = in.ID
	product.Category = in.Category
	product.Image = in.Image
	product.Sales = in.Sales
	// 显式指定更新字段，确保 Status 等零值字段参与更新
	if err := curd.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).Select(
		"name", "description", "price", "status", "category", "image", "sales",
	).Where("id = ?", in.ID).Updates(product).Error; err != nil {
		return err
	}
	return nil
}

// DeleteProduct 删除商品
func (r *productRepo) DeleteProduct(ctx context.Context, productID int64) error {
	_, err := curd.Product.WithContext(ctx).Where(curd.Product.ID.Eq(productID)).Delete()
	return err
}

// GetProduct 获取商品详情
func (r *productRepo) GetProduct(ctx context.Context, productID int64) (*biz.Product, error) {
	product, err := curd.Product.WithContext(ctx).Where(curd.Product.ID.Eq(productID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out := &biz.Product{}
	convert.Convert(product, out)
	out.ID = product.ID
	out.Category = product.Category
	out.Image = product.Image
	out.Sales = product.Sales
	return out, nil
}

// ListProduct 获取商品列表（分页 + 关键词搜索 + 分类筛选）
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword, category, status string) ([]*biz.Product, int32, error) {
	// 从 gorm gen 获取底层 *gorm.DB，确保 OR 条件正确分组
	baseQuery := curd.Product.WithContext(ctx)
	db := baseQuery.UnderlyingDB().Model(&mod.Product{})

	// 关键词搜索：(name LIKE ? OR description LIKE ?)
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", kw, kw)
	}

	// 分类筛选
	if category != "" {
		db = db.Where("category = ?", category)
	}

	// 状态筛选
	if status != "" {
		var statusInt int32
		if status == "上架" || status == "1" {
			statusInt = 1
		} else {
			statusInt = 0
		}
		db = db.Where("status = ?", statusInt)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var products []*mod.Product
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	if err := db.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	// 转换为 biz.Product
	items := make([]*biz.Product, 0, len(products))
	for _, p := range products {
		item := &biz.Product{}
		convert.Convert(p, item)
		item.ID = p.ID
		item.Category = p.Category
		item.Image = p.Image
		item.Sales = p.Sales
		item.Status = biz.ProductStatus(p.Status)
		items = append(items, item)
	}

	return items, int32(total), nil
}
