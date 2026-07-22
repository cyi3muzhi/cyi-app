package data

import (
	"context"
	"fmt"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
)

// categoryRepo 商品分类仓储实现
type categoryRepo struct {
	data *Data
	log  *log.Helper
}

// NewCategoryRepo 创建分类仓储
func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{data: data, log: log.NewHelper(logger)}
}

// ListCategoryTree 获取全量扁平分类列表，按排序值升序，并统计每个分类下的商品数
func (r *categoryRepo) ListCategoryTree(ctx context.Context, parentID int64) ([]*biz.Category, error) {
	categories, err := curd.Category.WithContext(ctx).
		Order(curd.Category.Sort).
		Find()
	if err != nil {
		return nil, err
	}

	// 统计每个分类下的商品数（不含已软删除的商品）
	categoryIDs := make([]int64, 0, len(categories))
	for _, cat := range categories {
		categoryIDs = append(categoryIDs, cat.ID)
	}
	type CatCount struct {
		CategoryID int64
		Count      int64
	}
	var counts []CatCount
	curd.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).
		Select("category_id, COUNT(*) as count").
		Where("category_id IN ?", categoryIDs).
		Group("category_id").
		Scan(&counts) // GORM Model() 会自动处理软删除条件
	countMap := make(map[int64]int64, len(counts))
	for _, c := range counts {
		countMap[c.CategoryID] = c.Count
	}

	out := make([]*biz.Category, 0, len(categories))
	for _, cat := range categories {
		bizCat := &biz.Category{}
		convert.Convert(cat, bizCat)
		bizCat.ID = cat.ID
		bizCat.Count = int32(countMap[cat.ID])
		out = append(out, bizCat)
	}
	return out, nil
}

// GetCategory 获取单个分类
func (r *categoryRepo) GetCategory(ctx context.Context, id int64) (*biz.Category, error) {
	cat, err := curd.Category.WithContext(ctx).Where(curd.Category.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	out := &biz.Category{}
	convert.Convert(cat, out)
	out.ID = cat.ID
	return out, nil
}

// CreateCategory 创建分类
func (r *categoryRepo) CreateCategory(ctx context.Context, in *biz.Category) error {
	cat := &mod.Category{}
	convert.Convert(in, cat)
	if err := curd.Category.WithContext(ctx).Create(cat); err != nil {
		return err
	}
	in.ID = cat.ID
	return nil
}

// UpdateCategory 更新分类
func (r *categoryRepo) UpdateCategory(ctx context.Context, in *biz.Category) error {
	cat := &mod.Category{}
	convert.Convert(in, cat)
	cat.ID = in.ID
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.Eq(in.ID)).
		UpdateSimple(
			curd.Category.ParentID.Value(in.ParentID),
			curd.Category.Name.Value(in.Name),
			curd.Category.Slug.Value(in.Slug),
			curd.Category.Icon.Value(in.Icon),
			curd.Category.Image.Value(in.Image),
			curd.Category.Sort.Value(in.Sort),
			curd.Category.Status.Value(in.Status),
			curd.Category.Showcase.Value(in.Showcase),
			curd.Category.SeoTitle.Value(in.SeoTitle),
			curd.Category.SeoKeywords.Value(in.SeoKeywords),
			curd.Category.SeoDesc.Value(in.SeoDesc),
		)
	return err
}

// DeleteCategory 删除分类（级联删除所有子分类，防止产生孤儿数据）
// 若分类（含子分类）下存在商品，则拒绝删除，避免产生 category 字段悬空的孤儿商品
func (r *categoryRepo) DeleteCategory(ctx context.Context, id int64) error {
	// 递归收集所有需要删除的分类 ID
	ids := []int64{id}
	if err := r.collectChildIDs(ctx, id, &ids); err != nil {
		return err
	}
	// 检查这些分类下是否存在商品，存在则拒绝删除
	if err := r.assertNoProductsInCategories(ctx, ids); err != nil {
		return err
	}
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.In(ids...)).
		Delete()
	return err
}

// collectChildIDs 递归收集指定父分类下所有子分类 ID
func (r *categoryRepo) collectChildIDs(ctx context.Context, parentID int64, ids *[]int64) error {
	children, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ParentID.Eq(parentID)).
		Find()
	if err != nil {
		return err
	}
	for _, child := range children {
		*ids = append(*ids, child.ID)
		if err := r.collectChildIDs(ctx, child.ID, ids); err != nil {
			return err
		}
	}
	return nil
}

// assertNoProductsInCategories 校验指定分类（含子分类）下是否存在商品
// 存在则返回错误，避免删除后产生 category_id 字段悬空的孤儿商品
func (r *categoryRepo) assertNoProductsInCategories(ctx context.Context, categoryIDs []int64) error {
	if len(categoryIDs) == 0 {
		return nil
	}
	// 直接按 category_id 检查商品表中是否仍有引用
	var count int64
	if err := curd.Product.WithContext(ctx).UnderlyingDB().Model(&mod.Product{}).
		Where("category_id IN ?", categoryIDs).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("分类下存在 %d 个商品，请先处理商品后再删除", count)
	}
	return nil
}

// ListChildCategories 获取直接子分类列表
func (r *categoryRepo) ListChildCategories(ctx context.Context, parentID int64) ([]*biz.Category, error) {
	categories, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ParentID.Eq(parentID)).
		Order(curd.Category.Sort).
		Find()
	if err != nil {
		return nil, err
	}
	out := make([]*biz.Category, 0, len(categories))
	for _, cat := range categories {
		bizCat := &biz.Category{}
		convert.Convert(cat, bizCat)
		bizCat.ID = cat.ID
		out = append(out, bizCat)
	}
	return out, nil
}

// UpdateCategorySort 更新排序
func (r *categoryRepo) UpdateCategorySort(ctx context.Context, id int64, sort int32) error {
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.Eq(id)).
		UpdateSimple(curd.Category.Sort.Value(sort))
	return err
}

// UpdateCategoryStatus 更新状态
func (r *categoryRepo) UpdateCategoryStatus(ctx context.Context, id int64, status bool) error {
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.Eq(id)).
		UpdateSimple(curd.Category.Status.Value(status))
	return err
}

// ToggleCategoryShowcase 切换首页展示
func (r *categoryRepo) ToggleCategoryShowcase(ctx context.Context, id int64, showcase bool) error {
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.Eq(id)).
		UpdateSimple(curd.Category.Showcase.Value(showcase))
	return err
}

// BatchOperation 批量操作（action: delete/disable/enable）
func (r *categoryRepo) BatchOperation(ctx context.Context, action string, ids []int64) (int32, error) {
	switch action {
	case "delete":
		// 级联收集所有子分类 ID，防止产生孤儿数据
		allIDs := make([]int64, 0, len(ids)*2)
		for _, id := range ids {
			allIDs = append(allIDs, id)
			if err := r.collectChildIDs(ctx, id, &allIDs); err != nil {
				return 0, err
			}
		}
		// 检查这些分类下是否存在商品，存在则拒绝删除，避免悬空 category 字段
		if err := r.assertNoProductsInCategories(ctx, allIDs); err != nil {
			return 0, err
		}
		result, err := curd.Category.WithContext(ctx).
			Where(curd.Category.ID.In(allIDs...)).
			Delete()
		if err != nil {
			return 0, err
		}
		return int32(result.RowsAffected), nil
	case "disable":
		result, err := curd.Category.WithContext(ctx).
			Where(curd.Category.ID.In(ids...)).
			UpdateSimple(curd.Category.Status.Value(false))
		if err != nil {
			return 0, err
		}
		return int32(result.RowsAffected), nil
	case "enable":
		result, err := curd.Category.WithContext(ctx).
			Where(curd.Category.ID.In(ids...)).
			UpdateSimple(curd.Category.Status.Value(true))
		if err != nil {
			return 0, err
		}
		return int32(result.RowsAffected), nil
	default:
		return 0, nil
	}
}
