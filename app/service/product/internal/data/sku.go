package data

import (
	"context"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/gorm"

	"github.com/go-kratos/kratos/v2/log"
)

type skuRepo struct {
	data *Data
	log  *log.Helper
}

// NewSkuRepo 创建SKU仓储
func NewSkuRepo(data *Data, logger log.Logger) biz.SkuRepo {
	return &skuRepo{data: data, log: log.NewHelper(logger)}
}

// CreateSkus 批量创建SKU（使用 GORM 批量 INSERT）
func (r *skuRepo) CreateSkus(ctx context.Context, skus []*biz.Sku) error {
	// 构建模型切片，GORM 的 Create 方法接受切片时会执行批量插入
	models := make([]*mod.Sku, 0, len(skus))
	for _, sku := range skus {
		models = append(models, &mod.Sku{
			ProductID: sku.ProductID,
			SkuCode:   sku.SkuCode,
			Price:     sku.Price,
			SpecDesc:  marshalSpec(sku.Spec),
			Image:     sku.Image,
		})
	}
	if len(models) == 0 {
		return nil
	}
	// 批量插入，GORM 会在一条 INSERT 语句中插入所有记录
	// 使用 ... 展开切片为变参，以确保 ID 回填的正确性
	if err := curd.Sku.WithContext(ctx).Create(models...); err != nil {
		return err
	}
	// 回写 ID（GORM 批量插入后会自动回填主键）
	for i, sku := range skus {
		sku.ID = models[i].ID
	}
	return nil
}

// ReplaceSkus 替换SKU（事务保证原子性：先删后插）
// 在事务内使用批量 INSERT 提升性能
func (r *skuRepo) ReplaceSkus(ctx context.Context, productID int64, skus []*biz.Sku) error {
	return curd.Q.Transaction(func(tx *curd.Query) error {
		// 物理删除旧的SKU（模型含 DeletedAt，不加 Unscoped 会软删除导致主键冲突）
		if _, err := tx.Sku.WithContext(ctx).Unscoped().Where(tx.Sku.ProductID.Eq(productID)).Delete(); err != nil {
			return err
		}
		// 批量创建新的SKU
		models := make([]*mod.Sku, 0, len(skus))
		for _, sku := range skus {
			models = append(models, &mod.Sku{
				Model:     gorm.Model{ID: sku.ID}, // 复用旧 ID（0=新建）
				ProductID: productID,
				SkuCode:   sku.SkuCode,
				Price:     sku.Price,
				SpecDesc:  marshalSpec(sku.Spec),
				Image:     sku.Image,
			})
		}
		if len(models) > 0 {
			// 使用 ... 展开切片为变参，GORM 会执行批量 INSERT
			if err := tx.Sku.WithContext(ctx).Create(models...); err != nil {
				return err
			}
		}
		// 回写 ID
		for i, sku := range skus {
			sku.ID = models[i].ID
			sku.ProductID = productID
		}
		return nil
	})
}

// skuModelToBiz 将数据层的SKU模型转换为业务层的SKU实体
func skuModelToBiz(m *mod.Sku) *biz.Sku {
	return &biz.Sku{
		ID:        m.ID,
		ProductID: m.ProductID,
		SkuCode:   m.SkuCode,
		Price:     m.Price,
		Spec:      unmarshalSpec(m.SpecDesc),
		Image:     m.Image,
	}
}

// GetSkusByProductID 获取商品的所有SKU
func (r *skuRepo) GetSkusByProductID(ctx context.Context, productID int64) ([]*biz.Sku, error) {
	models, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.Eq(productID)).Find()
	if err != nil {
		return nil, err
	}
	result := make([]*biz.Sku, 0, len(models))
	for _, m := range models {
		result = append(result, skuModelToBiz(m))
	}
	return result, nil
}

// GetSkusByProductIDs 批量获取多个商品的SKU，返回 map[productID][]*Sku
func (r *skuRepo) GetSkusByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*biz.Sku, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	models, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.In(productIDs...)).Find()
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]*biz.Sku)
	for _, m := range models {
		result[m.ProductID] = append(result[m.ProductID], skuModelToBiz(m))
	}
	return result, nil
}

// DeleteSkusByProductID 删除商品的所有SKU
func (r *skuRepo) DeleteSkusByProductID(ctx context.Context, productID int64) error {
	_, err := curd.Sku.WithContext(ctx).Where(curd.Sku.ProductID.Eq(productID)).Delete()
	return err
}
