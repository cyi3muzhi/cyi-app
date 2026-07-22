package data

import (
	"context"
	"errors"
	"fmt"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type productRepo struct {
	data *Data
	log  *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, log: log.NewHelper(logger)}
}

// CreateProduct 创建商品
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	product := &mod.Product{}
	if err := convert.Convert(in, product); err != nil {
		return fmt.Errorf("convert product for create failed: %w", err)
	}
	// 设置新增字段
	product.CategoryID = in.CategoryID
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
	if err := convert.Convert(in, product); err != nil {
		return fmt.Errorf("convert product for update failed: %w", err)
	}
	product.ID = in.ID
	product.CategoryID = in.CategoryID
	product.Image = in.Image
	product.Sales = in.Sales
	// 显式指定更新字段，确保 Status 等零值字段参与更新
	if err := curd.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).Select(
		"name", "description", "price", "status", "category_id", "image", "sales",
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
	if err := convert.Convert(product, out); err != nil {
		return nil, fmt.Errorf("convert product for get failed: %w", err)
	}
	out.ID = product.ID
	out.CategoryID = product.CategoryID
	out.Image = product.Image
	out.Sales = product.Sales
	return out, nil
}

// ListProduct 获取商品列表（分页 + 关键词搜索 + 分类筛选 + 按ID降序）
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string) ([]*biz.Product, int32, error) {
	// 从 gorm gen 获取底层 *gorm.DB，确保 OR 条件正确分组
	baseQuery := curd.Product.WithContext(ctx)
	db := baseQuery.UnderlyingDB().Model(&mod.Product{})

	// 关键词搜索：(name LIKE ? OR description LIKE ?)
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", kw, kw)
	}

	// 分类筛选
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}

	// 状态筛选（支持 "上架"/"1"=上架, "下架"/"0"=下架）
	if status != "" {
		switch status {
		case "上架", "1":
			db = db.Where("status = ?", 1)
		case "下架", "0":
			db = db.Where("status = ?", 0)
		}
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 分页查询，按 ID 降序
	var products []*mod.Product
	offset := int((page - 1) * pageSize)
	limit := int(pageSize)
	if err := db.Order("id DESC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	// 转换为 biz.Product
	items := make([]*biz.Product, 0, len(products))
	for _, p := range products {
		item := &biz.Product{}
		if err := convert.Convert(p, item); err != nil {
			return nil, 0, fmt.Errorf("convert product for list failed: %w", err)
		}
		item.ID = p.ID
		item.CategoryID = p.CategoryID
		item.Image = p.Image
		item.Sales = p.Sales
		item.Status = biz.ProductStatus(p.Status)
		items = append(items, item)
	}

	return items, int32(total), nil
}
