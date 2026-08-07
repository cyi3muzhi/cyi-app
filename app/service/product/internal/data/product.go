package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	return updateProductByTx(ctx, curd.Q, in)
}

// updateProductByTx 在指定事务（或默认连接）内更新商品表
// 显式指定更新字段，确保 Status 等零值字段参与更新
func updateProductByTx(ctx context.Context, q *curd.Query, in *biz.Product) error {
	product := bizToModelProduct(in)
	product.ID = in.ID
	if err := q.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).Select(
		"name", "description", "price", "status", "category_id", "image", "images",
	).Where("id = ?", in.ID).Updates(product).Error; err != nil {
		return fmt.Errorf("update product failed: %w", err)
	}
	return nil
}

// UpdateProductTx 在单个 gorm 事务内更新商品及其关联数据（规格组、SKU、库存）
// 任一写操作失败则整体回滚，保证多表数据一致；由 biz 层 UpdateProduct 通过类型断言调用
func (r *productRepo) UpdateProductTx(ctx context.Context, in *biz.Product) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		// 1. 更新商品主体
		if err := updateProductByTx(ctx, tx, in); err != nil {
			return err
		}
		// 2. 替换规格组（先删后插），始终执行以支持清空规格
		if err := replaceSpecGroupsTx(ctx, tx, in.ID, in.SpecGroups); err != nil {
			return err
		}
		// 3. 事务内加载旧 SKU 与库存，建立 spec → SKU 映射（读与写在同一事务，保证一致性）
		oldSkus, err := getSkusByProductIDTx(ctx, tx, in.ID)
		if err != nil {
			return err
		}
		oldSkuBySpec := make(map[string]*biz.Sku, len(oldSkus))
		oldSkuIDs := make([]int64, 0, len(oldSkus))
		for _, s := range oldSkus {
			key := productSpecKey(s.Spec)
			oldSkuBySpec[key] = s
			oldSkuIDs = append(oldSkuIDs, s.ID)
		}
		oldStockBySkuID := make(map[int64]*biz.Stock)
		if len(oldSkuIDs) > 0 {
			stocks, err := getStocksBySkuIDsTx(ctx, tx, oldSkuIDs)
			if err != nil {
				return err
			}
			oldStockBySkuID = stocks
		}
		// 4. 为新 SKU 匹配旧 ID：spec 相同则复用旧 ID（ReplaceSkus 会带 ID 插入，保留主键与库存记录）
		newSkuSpecSet := make(map[string]bool, len(in.Skus))
		for _, sku := range in.Skus {
			sku.ProductID = in.ID
			key := productSpecKey(sku.Spec)
			newSkuSpecSet[key] = true
			if old, ok := oldSkuBySpec[key]; ok {
				sku.ID = old.ID // 复用旧 ID，库存记录不丢失
			}
		}
		// 5. 替换 SKU（先删后插，已匹配的 SKU 用旧 ID 插入，库存链接保留）
		if err := replaceSkusTx(ctx, tx, in.ID, in.Skus); err != nil {
			return err
		}
		// 6. 处理库存：匹配的更新总数，新增的创建
		skus := in.Skus
		if len(skus) == 0 {
			// 无 SKU 时创建一条 product 级库存占位
			skus = []*biz.Sku{{Stock: 0}}
		}
		for _, sku := range skus {
			if stock, ok := oldStockBySkuID[sku.ID]; ok {
				// 更新现有库存（只改 total/available，保留 reserved/version）
				stock.Total = sku.Stock
				stock.Available = sku.Stock
				if err := updateStockTx(ctx, tx, stock); err != nil {
					return fmt.Errorf("update stock for sku %d failed: %w", sku.ID, err)
				}
			} else {
				// 新 SKU，创建库存
				if err := createStockTx(ctx, tx, &biz.Stock{
					ProductID: in.ID,
					SkuID:     sku.ID,
					Total:     sku.Stock,
					Available: sku.Stock,
				}); err != nil {
					return fmt.Errorf("create stock for sku %d failed: %w", sku.ID, err)
				}
			}
		}
		// 7. 删除被移除 SKU 的库存记录（失败即回滚，避免遗留孤儿库存）
		for _, old := range oldSkus {
			key := productSpecKey(old.Spec)
			if !newSkuSpecSet[key] {
				if err := deleteStockBySkuIDTx(ctx, tx, old.ID); err != nil {
					return fmt.Errorf("delete stock for sku %d failed: %w", old.ID, err)
				}
			}
		}
		return nil
	})
}

// productSpecKey 将规格 map 序列化为可比较的字符串键（与 biz.specKey 保持一致的确定性 JSON 序列化）
// 用于事务内按规格匹配旧 SKU ID
func productSpecKey(spec map[string]string) string {
	if len(spec) == 0 {
		return "{}"
	}
	data, _ := json.Marshal(spec)
	return string(data)
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
func (r *productRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ, stockLTE int32) ([]*biz.Product, int32, error) {
	// 从 gorm gen 获取底层 *gorm.DB，确保 OR 条件正确分组
	baseQuery := curd.Product.WithContext(ctx)
	db := baseQuery.UnderlyingDB().Model(&mod.Product{})

	// 关键词搜索：使用 FULLTEXT 索引（ngram 分词器）+ BOOLEAN MODE
	// 相比 LIKE 全表扫描，FULLTEXT 走索引性能更优，且支持相关度排序
	// 注意：BOOLEAN MODE 中空格分隔的词默认是 OR 关系，需加 + 前缀强制 AND（所有词都要匹配），
	// 否则搜索 "HUAWEI Mate 80 Pro Max" 会因 "Pro"/"Max" 命中其他商品导致结果发散
	if keyword != "" {
		kw := escapeFulltextKeyword(keyword)
		if kw != "" {
			query := toFulltextAndQuery(kw)
			db = db.Where("MATCH(name, description) AGAINST(? IN BOOLEAN MODE)", query)
		}
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
	// product v1 proto 未定义商品状态常量，此处使用 biz.ProductStatusOffSale/OnSale 常量替代魔法数字
	if status != "" {
		switch status {
		case "上架", "1":
			db = db.Where("status = ?", int32(biz.ProductStatusOnSale))
		case "下架", "0":
			db = db.Where("status = ?", int32(biz.ProductStatusOffSale))
		}
	}

	// 库存筛选（通过子查询汇总 stock 表的可用库存）
	if stockEQ >= 0 {
		// 使用 COALESCE 处理无库存记录的情况（SUM(NULL)=NULL，COALESCE 转为 0）
		subquery := "(SELECT COALESCE(SUM(available), 0) FROM stock WHERE stock.product_id = product.id AND stock.deleted_at IS NULL) = ?"
		db = db.Where(subquery, stockEQ)
	}

	// 库存上限筛选（库存预警用：可用库存 <= 阈值）
	if stockLTE >= 0 {
		subquery := "(SELECT COALESCE(SUM(available), 0) FROM stock WHERE stock.product_id = product.id AND stock.deleted_at IS NULL) <= ?"
		db = db.Where(subquery, stockLTE)
	}

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// 分页参数归一化：page 最小为 1，pageSize 默认 10、上限 100
	// 先做边界校验再转 int，防止 int32 溢出或负数导致 offset 异常
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var products []*mod.Product
	offset := int(page-1) * int(pageSize)
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

// SuggestProduct 搜索联想：仅返回商品名称列表（去重），限制 limit 条
// 使用 name 字段的 FULLTEXT 索引（ngram 分词器）+ BOOLEAN MODE 查询
// 仅返回上架商品（status=上架，即 biz.ProductStatusOnSale）
func (r *productRepo) SuggestProduct(ctx context.Context, keyword string, limit int32) ([]string, error) {
	kw := escapeFulltextKeyword(keyword)
	if kw == "" {
		return nil, nil
	}
	query := toFulltextAndQuery(kw)
	var names []string
	err := curd.Product.WithContext(ctx).UnderlyingDB().
		Table("product").
		Select("DISTINCT name").
		Where("MATCH(name) AGAINST(? IN BOOLEAN MODE)", query).
		Where("status = ?", int32(biz.ProductStatusOnSale)). // 仅上架商品（product v1 proto 无状态常量，使用 biz 常量）
		Where("deleted_at IS NULL").
		Limit(int(limit)).
		Pluck("name", &names).Error
	if err != nil {
		return nil, err
	}
	return names, nil
}

// toFulltextAndQuery 将清理后的关键词转换为 BOOLEAN MODE 的 AND 查询
// FULLTEXT BOOLEAN MODE 中空格分隔的词默认是 OR 关系，
// 为每个词加 + 前缀可强制要求所有词都匹配，避免部分匹配导致结果发散。
// 例："HUAWEI Mate 80" → "+HUAWEI +Mate +80"
func toFulltextAndQuery(cleaned string) string {
	terms := strings.Fields(cleaned)
	if len(terms) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(terms))
	for _, t := range terms {
		quoted = append(quoted, "+"+t)
	}
	return strings.Join(quoted, " ")
}

// escapeFulltextKeyword 清理 FULLTEXT BOOLEAN MODE 查询的关键词
// 移除特殊字符 + - * < > ( ) ~ " @，避免被误解析为操作符
// 清理后若为空字符串，调用方应跳过查询
func escapeFulltextKeyword(keyword string) string {
	if keyword == "" {
		return ""
	}
	// BOOLEAN MODE 特殊字符替换为空格（避免误解析为操作符）
	replacer := strings.NewReplacer(
		"+", " ", "-", " ", "*", " ", "<", " ", ">", " ",
		"(", " ", ")", " ", "~", " ", "\"", " ", "@", " ",
	)
	cleaned := replacer.Replace(keyword)
	// 压缩多余空格并去首尾空格
	cleaned = strings.TrimSpace(cleaned)
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	return cleaned
}
