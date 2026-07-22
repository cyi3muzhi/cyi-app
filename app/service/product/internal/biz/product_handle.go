package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "cyi-app/api/service/product/v1"
)

// ProductRepo 商品仓储接口
type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
	DeleteProduct(ctx context.Context, productID int64) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
	ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string) ([]*Product, int32, error)
}

// ProductUseCase 商品用例
type ProductUseCase struct {
	repo            ProductRepo
	stockRepo       StockRepo
	specGroupRepo   SpecGroupRepo
	skuRepo         SkuRepo
	log             *log.Helper
}

// NewProductUseCase 创建商品用例
func NewProductUseCase(repo ProductRepo, stockRepo StockRepo, specGroupRepo SpecGroupRepo, skuRepo SkuRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{repo: repo, stockRepo: stockRepo, specGroupRepo: specGroupRepo, skuRepo: skuRepo, log: log.NewHelper(logger)}
}

// CreateProduct 创建商品（含规格组和SKU）
// 若后续步骤失败，会回滚已创建的商品，保证数据一致性
func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.CreateProduct(ctx, in); err != nil {
		return v1.ErrorCreateProductFailed("create product failed")
	}

	// 标记是否需要回滚商品
	needRollback := true
	defer func() {
		if needRollback {
			// 回滚：删除已创建的商品及其关联数据
			_ = uc.specGroupRepo.DeleteSpecGroupsByProductID(ctx, in.ID)
			_ = uc.skuRepo.DeleteSkusByProductID(ctx, in.ID)
			_ = uc.repo.DeleteProduct(ctx, in.ID)
		}
	}()

	// 创建规格组
	if len(in.SpecGroups) > 0 {
		if err := uc.specGroupRepo.CreateSpecGroups(ctx, in.ID, in.SpecGroups); err != nil {
			return err
		}
	}
	// 创建SKU
	if len(in.Skus) > 0 {
		for _, sku := range in.Skus {
			sku.ProductID = in.ID
		}
		if err := uc.skuRepo.CreateSkus(ctx, in.Skus); err != nil {
			return err
		}
	}
	// 初始化库存（传统库存方案保留）
	stock := &Stock{
		ProductID: in.ID,
		Total:     0,
		Available: 0,
	}
	if err := uc.stockRepo.CreateStock(ctx, stock); err != nil {
		return v1.ErrorCreateStockFailed("create stock failed")
	}

	// 所有步骤成功，取消回滚
	needRollback = false
	return nil
}

// UpdateProduct 更新商品（含规格组和SKU的替换）
// 注意：规格组和SKU替换各自有独立的事务，若商品更新成功后规格组/SKU更新失败，
// 商品表的字段已更新但关联数据未被替换，需要手动重试。后续可引入跨仓储事务。
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.UpdateProduct(ctx, in); err != nil {
		return err
	}
	// 替换规格组（先删后插），始终执行以支持清空规格
	if err := uc.specGroupRepo.ReplaceSpecGroups(ctx, in.ID, in.SpecGroups); err != nil {
		return err
	}
	// 替换SKU（先删后插），始终执行以支持清空SKU
	for _, sku := range in.Skus {
		sku.ProductID = in.ID
	}
	if err := uc.skuRepo.ReplaceSkus(ctx, in.ID, in.Skus); err != nil {
		return err
	}
	return nil
}

// DeleteProduct 删除商品（含关联的规格组、SKU、库存）
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, productID int64) error {
	// 先删除关联的规格组和SKU
	_ = uc.specGroupRepo.DeleteSpecGroupsByProductID(ctx, productID)
	_ = uc.skuRepo.DeleteSkusByProductID(ctx, productID)
	// 删除库存记录
	_ = uc.stockRepo.DeleteStockByProductID(ctx, productID)
	return uc.repo.DeleteProduct(ctx, productID)
}

// GetProduct 获取商品详情
func (uc *ProductUseCase) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	product, err := uc.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	// 加载规格组和SKU
	specGroups, _ := uc.specGroupRepo.GetSpecGroupsByProductID(ctx, productID)
	skus, _ := uc.skuRepo.GetSkusByProductID(ctx, productID)
	product.SpecGroups = specGroups
	product.Skus = skus
	return product, nil
}

// ListProduct 获取商品列表
func (uc *ProductUseCase) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string) ([]*Product, int32, error) {
	products, total, err := uc.repo.ListProduct(ctx, page, pageSize, keyword, categoryID, status)
	if err != nil {
		return nil, 0, err
	}
	// 加载每个商品的规格组和SKU
	for _, p := range products {
		specGroups, _ := uc.specGroupRepo.GetSpecGroupsByProductID(ctx, p.ID)
		skus, _ := uc.skuRepo.GetSkusByProductID(ctx, p.ID)
		p.SpecGroups = specGroups
		p.Skus = skus
	}
	return products, total, nil
}
