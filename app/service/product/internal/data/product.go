package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

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

// ---------------------------------------------------------------------------
// 显式字段映射：不使用 convert.Convert，避开 mapstructure string→[]string 不兼容
// ---------------------------------------------------------------------------

// bizToModelProduct 将 biz.Product 转换为 model.Product（不含 SpecGroups/Skus）
func bizToModelProduct(in *biz.Product) *mod.Product {
	return &mod.Product{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		Status:      int32(in.Status),
		Sales:       in.Sales,
		CategoryID:  in.CategoryID,
		Image:       in.Image,
		Images:      marshalImages(in.Images),
	}
}

// modelToBizProduct 将 model.Product 转换为 biz.Product
func modelToBizProduct(p *mod.Product) *biz.Product {
	item := &biz.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Status:      biz.ProductStatus(p.Status),
		Sales:       p.Sales,
		CategoryID:  p.CategoryID,
		Image:       p.Image,
	}
	// 手动反序列化轮播图 JSON
	if p.Images != "" {
		json.Unmarshal([]byte(p.Images), &item.Images)
	}
	return item
}

// CreateProduct 创建商品
func (r *productRepo) CreateProduct(ctx context.Context, in *biz.Product) error {
	product := bizToModelProduct(in)
	if err := curd.Product.WithContext(ctx).Create(product); err != nil {
		return fmt.Errorf("create product failed: %w", err)
	}
	in.ID = product.ID
	return nil
}

// UpdateProduct 更新商品
func (r *productRepo) UpdateProduct(ctx context.Context, in *biz.Product) error {
	product := bizToModelProduct(in)
	product.ID = in.ID
	// 显式指定更新字段，确保 Status 等零值字段参与更新
	if err := curd.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).Select(
		"name", "description", "price", "status", "category_id", "image", "images",
	).Where("id = ?", in.ID).Updates(product).Error; err != nil {
		return fmt.Errorf("update product failed: %w", err)
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
	return modelToBizProduct(product), nil
}

// ListProduct 获取商品列表（分页 + 关键词搜索 + 分类筛选 + 库存筛选 + 按ID降序）
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*biz.Product, int32, error) {
	// 从 gorm gen 获取底层 *gorm.DB，确保 OR 条件正确分组
	baseQuery := curd.Product.WithContext(ctx)
	db := baseQuery.UnderlyingDB().Model(&mod.Product{})

	// 关键词搜索：(name LIKE ? OR description LIKE ?)
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", kw, kw)
	}

	// 分类筛选：匹配该分类及其所有子分类（层级分类下商品挂在叶子节点）
	if categoryID > 0 {
		ids, err := collectCategoryWithDescendants(ctx, categoryID)
		if err != nil {
			return nil, 0, err
		}
		if len(ids) == 0 {
			ids = []int64{categoryID}
		}
		db = db.Where("category_id IN ?", ids)
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

	// 库存筛选（通过子查询汇总 stock 表的可用库存）
	if stockEQ >= 0 {
		// 使用 COALESCE 处理无库存记录的情况（SUM(NULL)=NULL，COALESCE 转为 0）
		subquery := "(SELECT COALESCE(SUM(available), 0) FROM stock WHERE stock.product_id = product.id AND stock.deleted_at IS NULL) = ?"
		db = db.Where(subquery, stockEQ)
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

	// 转换为 biz.Product（使用显式字段映射，避免 mapstructure string→[]string 不兼容）
	items := make([]*biz.Product, 0, len(products))
	for _, p := range products {
		items = append(items, modelToBizProduct(p))
	}

	return items, int32(total), nil
}

// marshalImages 将轮播图列表序列化为 JSON 字符串
func marshalImages(images []string) string {
	if len(images) == 0 {
		return ""
	}
	data, _ := json.Marshal(images)
	return string(data)
}
