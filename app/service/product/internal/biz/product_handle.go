package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

// ProductRepo 商品仓储接口
type ProductRepo interface {
	CreateProduct(ctx context.Context, in *Product) error
	UpdateProduct(ctx context.Context, in *Product) error
	DeleteProduct(ctx context.Context, productID int64) error
	GetProduct(ctx context.Context, productID int64) (*Product, error)
	ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ, stockLTE int32) ([]*Product, int32, error)
	SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error)
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
		return fmt.Errorf("%w: create product failed", ErrCreateProductFailed)
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
			return fmt.Errorf("%w: create stock failed", ErrCreateStockFailed)
		}
	}

	// 所有步骤成功，取消回滚
	needRollback = false
	return nil
}

// UpdateProduct 更新商品（含规格组、SKU和库存）
// 按规格组合匹配 SKU：匹配上的复用旧 ID（库存不重建），新增的创建，移除的清除
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, in *Product) error {
	if err := uc.repo.UpdateProduct(ctx, in); err != nil {
		return err
	}

	// 替换规格组（先删后插），始终执行以支持清空规格
	if err := uc.specGroupRepo.ReplaceSpecGroups(ctx, in.ID, in.SpecGroups); err != nil {
		return err
	}

	// 1. 加载旧 SKU 和库存，建立 spec → SKU 映射
	oldSkus, _ := uc.skuRepo.GetSkusByProductID(ctx, in.ID)
	oldSkuBySpec := make(map[string]*Sku, len(oldSkus))
	oldSkuIDs := make([]int64, 0, len(oldSkus))
	for _, s := range oldSkus {
		key := specKey(s.Spec)
		oldSkuBySpec[key] = s
		oldSkuIDs = append(oldSkuIDs, s.ID)
	}
	// 加载旧库存
	oldStockBySkuID := make(map[int64]*Stock)
	if len(oldSkuIDs) > 0 {
		if stocks, err := uc.stockRepo.GetStocksBySkuIDs(ctx, oldSkuIDs); err == nil {
			oldStockBySkuID = stocks
		}
	}

	// 2. 为新 SKU 匹配旧 ID：spec 相同则复用旧 ID（ReplaceSkus 会带 ID 插入，保留主键）
	newSkuSpecSet := make(map[string]bool, len(in.Skus))
	for _, sku := range in.Skus {
		sku.ProductID = in.ID
		key := specKey(sku.Spec)
		newSkuSpecSet[key] = true
		if old, ok := oldSkuBySpec[key]; ok {
			sku.ID = old.ID // 复用旧 ID，库存记录不丢失
		}
	}

	// 3. 替换 SKU（先删后插，已匹配的 SKU 用旧 ID 插入，库存链接保留）
	if err := uc.skuRepo.ReplaceSkus(ctx, in.ID, in.Skus); err != nil {
		return err
	}

	// 4. 处理库存：匹配的更新总数，新增的创建，移除的删除
	skus := in.Skus
	if len(skus) == 0 {
		skus = []*Sku{{Stock: 0}}
	}
	for _, sku := range skus {
		if stock, ok := oldStockBySkuID[sku.ID]; ok {
			// 更新现有库存（只改 total/available，保留 reserved/version）
			stock.Total = sku.Stock
			stock.Available = sku.Stock
			if err := uc.stockRepo.UpdateStock(ctx, stock); err != nil {
				return fmt.Errorf("update stock for sku %d failed: %w", sku.ID, err)
			}
		} else {
			// 新 SKU，创建库存
			if err := uc.stockRepo.CreateStock(ctx, &Stock{
				ProductID: in.ID,
				SkuID:     sku.ID,
				Total:     sku.Stock,
				Available: sku.Stock,
			}); err != nil {
				return fmt.Errorf("create stock for sku %d failed: %w", sku.ID, err)
			}
		}
	}

	// 5. 删除被移除 SKU 的库存记录
	for _, old := range oldSkus {
		key := specKey(old.Spec)
		if !newSkuSpecSet[key] {
			_ = uc.stockRepo.DeleteStockBySkuID(ctx, old.ID)
		}
	}

	return nil
}

// specKey 将 spec map 转为可比较的字符串键（JSON 序列化后 key 有序，保证确定性）
func specKey(spec map[string]string) string {
	if len(spec) == 0 {
		return "{}"
	}
	data, _ := json.Marshal(spec)
	return string(data)
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
// @param stockEQ 可用库存精确筛选（-1 表示不筛选）
// @param stockLTE 可用库存上限筛选，库存 <= stockLTE（-1 表示不筛选，用于库存预警）
func (uc *ProductUseCase) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ, stockLTE int32) ([]*Product, int32, error) {
	products, total, err := uc.repo.ListProduct(ctx, page, pageSize, keyword, categoryID, status, stockEQ, stockLTE)
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

// SuggestProduct 搜索联想：返回匹配关键词的商品名列表
// 仅做参数校验，实际查询由 repo 层实现（使用 FULLTEXT 索引）
// @param keyword 联想关键词
// @param limit 返回条数上限（默认 8，最大 20）
// @return 商品名称列表（去重）
func (uc *ProductUseCase) SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error) {
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
	return uc.repo.SuggestProduct(ctx, keyword, limit)
}
