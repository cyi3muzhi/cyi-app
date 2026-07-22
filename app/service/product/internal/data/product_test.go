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


// newProductTestRepo 创建测试用的 productRepo
func newProductTestRepo(t *testing.T, db *gorm.DB) *productRepo {
	return &productRepo{data: &Data{cache: nil}, cache: nil, log: log.NewHelper(log.NewStdLogger(nil))}
}

// TestListProduct_StatusFilterOnSale 测试状态筛选 "上架" 只返回上架商品
func TestListProduct_StatusFilterOnSale(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	onSale := &mod.Product{Name: "上架商品", Status: 1}
	offSale := &mod.Product{Name: "下架商品", Status: 0}
	requireNoError(t, db.Create(onSale).Error)
	requireNoError(t, db.Create(offSale).Error)

	products, total, err := repo.ListProduct(ctx, 1, 10, "", "", "上架")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), total)
	assert.Len(t, products, 1)
	assert.Equal(t, "上架商品", products[0].Name)
}

// TestListProduct_StatusFilterNumericOne 测试状态筛选 "1" 也被识别为上架
func TestListProduct_StatusFilterNumericOne(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	onSale := &mod.Product{Name: "上架商品", Status: 1}
	requireNoError(t, db.Create(onSale).Error)

	products, total, err := repo.ListProduct(ctx, 1, 10, "", "", "1")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), total)
	assert.Len(t, products, 1)
}

// TestListProduct_StatusFilterOffSale 测试非上架状态字符串被识别为下架
func TestListProduct_StatusFilterOffSale(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	onSale := &mod.Product{Name: "上架商品", Status: 1}
	offSale := &mod.Product{Name: "下架商品", Status: 0}
	requireNoError(t, db.Create(onSale).Error)
	requireNoError(t, db.Create(offSale).Error)

	products, total, err := repo.ListProduct(ctx, 1, 10, "", "", "下架")
	assert.NoError(t, err)
	assert.Equal(t, int32(1), total)
	assert.Equal(t, "下架商品", products[0].Name)
}

// TestListProduct_KeywordSearch 测试关键词同时搜索名称和描述
func TestListProduct_KeywordSearch(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	p1 := &mod.Product{Name: "智能手机", Description: "高性能"}
	p2 := &mod.Product{Name: "普通商品", Description: "智能手环描述"}
	p3 := &mod.Product{Name: "其他", Description: "无关"}
	requireNoError(t, db.Create(p1).Error)
	requireNoError(t, db.Create(p2).Error)
	requireNoError(t, db.Create(p3).Error)

	products, total, err := repo.ListProduct(ctx, 1, 10, "智能", "", "")
	assert.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, products, 2)
}

// TestListProduct_CategoryFilter 测试分类筛选
func TestListProduct_CategoryFilter(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	p1 := &mod.Product{Name: "手机", Category: "数码"}
	p2 := &mod.Product{Name: "电脑", Category: "数码"}
	p3 := &mod.Product{Name: "衣服", Category: "服饰"}
	requireNoError(t, db.Create(p1).Error)
	requireNoError(t, db.Create(p2).Error)
	requireNoError(t, db.Create(p3).Error)

	products, total, err := repo.ListProduct(ctx, 1, 10, "", "数码", "")
	assert.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, products, 2)
}

// TestListProduct_Pagination 测试分页计算
func TestListProduct_Pagination(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		requireNoError(t, db.Create(&mod.Product{Name: "商品"}).Error)
	}

	products, total, err := repo.ListProduct(ctx, 2, 2, "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, int32(5), total)
	assert.Len(t, products, 2)
}

// TestCreateProduct_AssignsID 测试创建商品后会回写 ID
func TestCreateProduct_AssignsID(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	in := &biz.Product{Name: "测试商品", Price: 100, Status: biz.ProductStatusOnSale, Category: "数码", Image: "image.jpg", Sales: 10}
	err := repo.CreateProduct(ctx, in)
	assert.NoError(t, err)
	assert.Greater(t, in.ID, int64(0))

	var model mod.Product
	db.First(&model, in.ID)
	assert.Equal(t, "测试商品", model.Name)
	assert.Equal(t, int64(100), model.Price)
	assert.Equal(t, int32(1), model.Status)
	assert.Equal(t, "数码", model.Category)
	assert.Equal(t, "image.jpg", model.Image)
	assert.Equal(t, int32(10), model.Sales)
}

// TestGetProduct_NotFound 测试获取不存在的商品返回 nil
func TestGetProduct_NotFound(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	product, err := repo.GetProduct(ctx, 9999)
	assert.NoError(t, err)
	assert.Nil(t, product)
}

// TestCreateProduct_OffSaleStatusSaved 测试创建下架商品时，Status=0 被正确持久化，
// 防止 GORM 跳过零值导致商品被错误地标记为上架。
func TestCreateProduct_OffSaleStatusSaved(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	in := &biz.Product{Name: "下架商品", Status: biz.ProductStatusOffSale}
	err := repo.CreateProduct(ctx, in)
	assert.NoError(t, err)
	assert.Greater(t, in.ID, int64(0))

	var model mod.Product
	requireNoError(t, db.First(&model, in.ID).Error)
	assert.Equal(t, int32(0), model.Status)
}

// TestUpdateProduct_StatusToOffSale 测试将商品状态从上架更新为下架时，Status=0 被正确写入，
// 防止 GORM Updates 跳过零值导致状态更新不生效。
func TestUpdateProduct_StatusToOffSale(t *testing.T) {
	db := setupProductDataTestDB(t)
	repo := newProductTestRepo(t, db)
	ctx := context.Background()

	existing := &mod.Product{Name: "原商品", Status: 1}
	requireNoError(t, db.Create(existing).Error)

	err := repo.UpdateProduct(ctx, &biz.Product{
		ID:     existing.ID,
		Name:   "更新后商品",
		Status: biz.ProductStatusOffSale,
	})
	assert.NoError(t, err)

	var updated mod.Product
	requireNoError(t, db.First(&updated, existing.ID).Error)
	assert.Equal(t, "更新后商品", updated.Name)
	assert.Equal(t, int32(0), updated.Status)
}
