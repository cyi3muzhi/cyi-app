package data

import (
	"context"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// categoryRepo 商品分类仓储实现
type categoryRepo struct {
	data  *Data
	cache *redis.Client
	log   *log.Helper
}

// NewCategoryRepo 创建分类仓储
func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{data: data, cache: data.cache, log: log.NewHelper(logger)}
}

// ListCategoryTree 获取全量扁平分类列表，按排序值升序
func (r *categoryRepo) ListCategoryTree(ctx context.Context, parentID int64) ([]*biz.Category, error) {
	categories, err := curd.Category.WithContext(ctx).
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

// DeleteCategory 删除分类
func (r *categoryRepo) DeleteCategory(ctx context.Context, id int64) error {
	_, err := curd.Category.WithContext(ctx).
		Where(curd.Category.ID.Eq(id)).
		Delete()
	return err
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
		result, err := curd.Category.WithContext(ctx).
			Where(curd.Category.ID.In(ids...)).
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
