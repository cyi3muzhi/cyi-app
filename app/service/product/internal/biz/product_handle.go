package biz

import (
	"context"

	v1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
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
	repo          ProductRepo
	stockRepo     StockRepo
	specGroupRepo SpecGroupRepo
	skuRepo       SkuRepo
	log           *log.Helper
}

// NewProductUseCase 创建商品用例
func NewProductUseCase(repo ProductRepo, stockRepo StockRepo, specGroupRepo SpecGroupRepo, skuRepo SkuRepo, logger log.Logger) *ProductUseCase {
	return &ProductUseCase{repo: repo, stockRepo: stockRepo, specGroupRepo: specGroupRepo, skuRepo: skuRepo, log: log.NewHelper(logger)}
}

// CreateProduct 创建商品（含规格组和SKU）
// TODO: 后续应改为数据库事务（curd.Q.Transaction），当前使用补偿回滚机制
func (uc *ProductUseCase) CreateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.CreateProduct(ctx, in); err != nil {
		return v1.ErrorCreateProductFailed("create product failed")
	}

	// 标记是否需要回滚商品
	needRollback := true
	defer func() {
		if needRollback {
			uc.log.Warnf("CreateProduct rollback: product_id=%d", in.ID)
			// 依次删除已创建的关联数据和商品主体，各自记录错误但不中断回滚流程
			if err := uc.specGroupRepo.DeleteSpecGroupsByProductID(ctx, in.ID); err != nil {
				uc.log.Errorf("rollback delete spec groups failed: %v", err)
			}
			if err := uc.skuRepo.DeleteSkusByProductID(ctx, in.ID); err != nil {
				uc.log.Errorf("rollback delete skus failed: %v", err)
			}
			if err := uc.stockRepo.DeleteStockByProductID(ctx, in.ID); err != nil {
				uc.log.Errorf("rollback delete stock failed: %v", err)
			}
			if err := uc.repo.DeleteProduct(ctx, in.ID); err != nil {
				uc.log.Errorf("rollback delete product failed: %v", err)
			}
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
	// 初始化库存（每个 SKU 一条库存记录；无 SKU 时创建 product 级记录 sku_id=0）
	skus := in.Skus
	if len(skus) == 0 {
		// 无 SKU 时创建一条空库存占位
		skus = []*Sku{{Stock: 0}}
	}
	for _, sku := range skus {
		stock := &Stock{
			ProductID: in.ID,
			SkuID:     sku.ID,
			Total:     sku.Stock,
			Available: sku.Stock,
		}
		if err := uc.stockRepo.CreateStock(ctx, stock); err != nil {
			return v1.ErrorCreateStockFailed("create stock failed")
		}
	}

	// 所有步骤成功，取消回滚
	needRollback = false
	return nil
}

// UpdateProduct 更新商品（含规格组、SKU和库存的替换）
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
	// 替换库存：先删旧库存，为每个 SKU 创建新库存
	_ = uc.stockRepo.DeleteStockByProductID(ctx, in.ID)
	skus := in.Skus
	if len(skus) == 0 {
		skus = []*Sku{{Stock: 0}}
	}
	for _, sku := range skus {
		stock := &Stock{
			ProductID: in.ID,
			SkuID:     sku.ID,
			Total:     sku.Stock,
			Available: sku.Stock,
		}
		if err := uc.stockRepo.CreateStock(ctx, stock); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProduct 删除商品（含关联的规格组、SKU、库存）
// 关联删除失败时记录警告日志，不阻塞主流程
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, productID int64) error {
	if err := uc.specGroupRepo.DeleteSpecGroupsByProductID(ctx, productID); err != nil {
		uc.log.Warnf("delete spec groups for product %d: %v", productID, err)
	}
	if err := uc.skuRepo.DeleteSkusByProductID(ctx, productID); err != nil {
		uc.log.Warnf("delete skus for product %d: %v", productID, err)
	}
	if err := uc.stockRepo.DeleteStockByProductID(ctx, productID); err != nil {
		uc.log.Warnf("delete stock for product %d: %v", productID, err)
	}
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
	// 加载每个 SKU 的库存
	for _, sku := range product.Skus {
		stock, err := uc.stockRepo.GetStock(ctx, sku.ID)
		if err == nil && stock != nil {
			sku.Stock = stock.Available
		}
	}
	return product, nil
}

// ListProduct 获取商品列表（使用批量加载避免 N+1 查询）
func (uc *ProductUseCase) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string) ([]*Product, int32, error) {
	products, total, err := uc.repo.ListProduct(ctx, page, pageSize, keyword, categoryID, status)
	if err != nil {
		return nil, 0, err
	}
	if len(products) == 0 {
		return products, total, nil
	}

	// 收集所有商品ID
	productIDs := make([]int64, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	// 批量加载规格组
	specGroupMap, _ := uc.specGroupRepo.GetSpecGroupsByProductIDs(ctx, productIDs)
	// 批量加载SKU
	skuMap, _ := uc.skuRepo.GetSkusByProductIDs(ctx, productIDs)

	// 收集所有SKU ID并回填规格组和SKU
	var allSkuIDs []int64
	for _, p := range products {
		if skus, ok := skuMap[p.ID]; ok {
			p.Skus = skus
			for _, sku := range skus {
				allSkuIDs = append(allSkuIDs, sku.ID)
			}
		}
		if groups, ok := specGroupMap[p.ID]; ok {
			p.SpecGroups = groups
		}
	}

	// 批量加载库存
	if len(allSkuIDs) > 0 {
		stockMap, _ := uc.stockRepo.GetStocksBySkuIDs(ctx, allSkuIDs)
		for _, p := range products {
			for _, sku := range p.Skus {
				if s, ok := stockMap[sku.ID]; ok {
					sku.Stock = s.Available
				}
			}
		}
	}

	return products, total, nil
}
