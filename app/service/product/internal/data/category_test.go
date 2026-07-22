package data

import (
	"context"
	"testing"

	"cyi-app/app/service/product/internal/biz"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)


// newCategoryTestRepo 创建测试用的 categoryRepo
func newCategoryTestRepo(t *testing.T, db *gorm.DB) *categoryRepo {
	return &categoryRepo{data: &Data{}, log: log.NewHelper(log.NewStdLogger(nil))}
}

// TestDeleteCategory_CascadeChildren 测试删除父分类时级联删除所有子分类
func TestDeleteCategory_CascadeChildren(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	parent := &mod.Category{Name: "父分类", ParentID: 0}
	requireNoError(t, db.Create(parent).Error)
	child := &mod.Category{Name: "子分类", ParentID: parent.ID}
	requireNoError(t, db.Create(child).Error)
	grandchild := &mod.Category{Name: "孙分类", ParentID: child.ID}
	requireNoError(t, db.Create(grandchild).Error)

	err := repo.DeleteCategory(ctx, parent.ID)
	assert.NoError(t, err)

	var count int64
	db.Model(&mod.Category{}).Count(&count)
	assert.Equal(t, int64(0), count, "父、子、孙分类应全部被级联删除")
}

// TestDeleteCategory_WithProductsBlocked 测试分类下存在商品时拒绝删除
func TestDeleteCategory_WithProductsBlocked(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	cat := &mod.Category{Name: "手机数码"}
	requireNoError(t, db.Create(cat).Error)
	product := &mod.Product{Name: "智能手机", Category: "手机数码"}
	requireNoError(t, db.Create(product).Error)

	err := repo.DeleteCategory(ctx, cat.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类下存在")

	var count int64
	db.Model(&mod.Category{}).Count(&count)
	assert.Equal(t, int64(1), count, "分类应保留")
}

// TestDeleteCategory_ChildHasProductsBlocked 测试子分类下存在商品时，父分类删除也应被拒绝
func TestDeleteCategory_ChildHasProductsBlocked(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	parent := &mod.Category{Name: "电子产品"}
	requireNoError(t, db.Create(parent).Error)
	child := &mod.Category{Name: "手机", ParentID: parent.ID}
	requireNoError(t, db.Create(child).Error)
	product := &mod.Product{Name: "智能手机", Category: "手机"}
	requireNoError(t, db.Create(product).Error)

	err := repo.DeleteCategory(ctx, parent.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类下存在")

	var count int64
	db.Model(&mod.Category{}).Count(&count)
	assert.Equal(t, int64(2), count, "父分类和子分类都应保留")
}

// TestBatchOperation_DeleteCascade 测试批量删除时级联收集子分类
func TestBatchOperation_DeleteCascade(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	parent := &mod.Category{Name: "服饰"}
	requireNoError(t, db.Create(parent).Error)
	child := &mod.Category{Name: "男装", ParentID: parent.ID}
	requireNoError(t, db.Create(child).Error)

	affected, err := repo.BatchOperation(ctx, "delete", []int64{parent.ID})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), affected)

	var count int64
	db.Model(&mod.Category{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestBatchOperation_DeleteWithProductsBlocked 测试批量删除时若分类下存在商品则拒绝
func TestBatchOperation_DeleteWithProductsBlocked(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	cat := &mod.Category{Name: "家电"}
	requireNoError(t, db.Create(cat).Error)
	product := &mod.Product{Name: "冰箱", Category: "家电"}
	requireNoError(t, db.Create(product).Error)

	affected, err := repo.BatchOperation(ctx, "delete", []int64{cat.ID})
	assert.Error(t, err)
	assert.Equal(t, int32(0), affected)
}

// TestBatchOperation_DisableAndEnable 测试批量启用与禁用分类
func TestBatchOperation_DisableAndEnable(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	cat1 := &mod.Category{Name: "分类1", Status: true}
	cat2 := &mod.Category{Name: "分类2", Status: true}
	requireNoError(t, db.Create(cat1).Error)
	requireNoError(t, db.Create(cat2).Error)

	affected, err := repo.BatchOperation(ctx, "disable", []int64{cat1.ID, cat2.ID})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), affected)

	var c1, c2 mod.Category
	db.First(&c1, cat1.ID)
	db.First(&c2, cat2.ID)
	assert.False(t, c1.Status)
	assert.False(t, c2.Status)

	affected, err = repo.BatchOperation(ctx, "enable", []int64{cat1.ID})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), affected)

	db.First(&c1, cat1.ID)
	assert.True(t, c1.Status)
}

// TestBatchOperation_UnknownAction 测试未知批量操作应返回 0 影响行数且无错误
func TestBatchOperation_UnknownAction(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	affected, err := repo.BatchOperation(ctx, "unknown", []int64{1})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), affected)
}

// TestUpdateCategory_Success 测试更新分类字段
func TestUpdateCategory_Success(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newCategoryTestRepo(t, db)
	ctx := context.Background()

	cat := &mod.Category{Name: "旧名称"}
	requireNoError(t, db.Create(cat).Error)

	err := repo.UpdateCategory(ctx, &biz.Category{
		ID:       cat.ID,
		Name:     "新名称",
		Slug:     "new-slug",
		Sort:     10,
		Status:   false,
		Showcase: true,
	})
	assert.NoError(t, err)

	var updated mod.Category
	db.First(&updated, cat.ID)
	assert.Equal(t, "新名称", updated.Name)
	assert.Equal(t, "new-slug", updated.Slug)
	assert.Equal(t, int32(10), updated.Sort)
	assert.False(t, updated.Status)
	assert.True(t, updated.Showcase)
}


