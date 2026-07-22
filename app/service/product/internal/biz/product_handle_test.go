package biz

import (
	"context"
	"io"
	"testing"

	v1 "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockProductRepo 商品仓储模拟实现
type mockProductRepo struct {
	createProduct func(ctx context.Context, in *Product) error
	updateProduct func(ctx context.Context, in *Product) error
	deleteProduct func(ctx context.Context, productID int64) error
	getProduct    func(ctx context.Context, productID int64) (*Product, error)
	listProduct   func(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*Product, int32, error)
}

func (m *mockProductRepo) CreateProduct(ctx context.Context, in *Product) error {
	return m.createProduct(ctx, in)
}
func (m *mockProductRepo) UpdateProduct(ctx context.Context, in *Product) error {
	return m.updateProduct(ctx, in)
}
func (m *mockProductRepo) DeleteProduct(ctx context.Context, productID int64) error {
	return m.deleteProduct(ctx, productID)
}
func (m *mockProductRepo) GetProduct(ctx context.Context, productID int64) (*Product, error) {
	return m.getProduct(ctx, productID)
}
func (m *mockProductRepo) ListProduct(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*Product, int32, error) {
	return m.listProduct(ctx, page, pageSize, keyword, categoryID, status, stockEQ)
}

// mockStockRepo 库存仓储模拟实现
type mockStockRepo struct {
	createStock            func(ctx context.Context, in *Stock) error
	updateStock            func(ctx context.Context, in *Stock) error
	deleteStockByProductID func(ctx context.Context, productID int64) error
	deleteStockBySkuID     func(ctx context.Context, skuID int64) error
	reserveStock           func(ctx context.Context, skuID int64, count, version int32) (int32, error)
	deductStock            func(ctx context.Context, skuID int64, count, version int32) (int32, error)
	releaseStock           func(ctx context.Context, skuID int64, count, version int32) (int32, error)
	addStock               func(ctx context.Context, skuID int64, count, version int32) (int32, error)
	getStock               func(ctx context.Context, skuID int64) (*Stock, error)
	getStocksBySkuIDs      func(ctx context.Context, skuIDs []int64) (map[int64]*Stock, error)
}

func (m *mockStockRepo) CreateStock(ctx context.Context, in *Stock) error {
	return m.createStock(ctx, in)
}
func (m *mockStockRepo) UpdateStock(ctx context.Context, in *Stock) error {
	return m.updateStock(ctx, in)
}
func (m *mockStockRepo) DeleteStockByProductID(ctx context.Context, productID int64) error {
	return m.deleteStockByProductID(ctx, productID)
}
func (m *mockStockRepo) DeleteStockBySkuID(ctx context.Context, skuID int64) error {
	return m.deleteStockBySkuID(ctx, skuID)
}
func (m *mockStockRepo) ReserveStock(ctx context.Context, skuID int64, count, version int32) (int32, error) {
	return m.reserveStock(ctx, skuID, count, version)
}
func (m *mockStockRepo) DeductStock(ctx context.Context, skuID int64, count, version int32) (int32, error) {
	return m.deductStock(ctx, skuID, count, version)
}
func (m *mockStockRepo) ReleaseStock(ctx context.Context, skuID int64, count, version int32) (int32, error) {
	return m.releaseStock(ctx, skuID, count, version)
}
func (m *mockStockRepo) AddStock(ctx context.Context, skuID int64, count, version int32) (int32, error) {
	return m.addStock(ctx, skuID, count, version)
}
func (m *mockStockRepo) GetStock(ctx context.Context, skuID int64) (*Stock, error) {
	return m.getStock(ctx, skuID)
}
func (m *mockStockRepo) GetStocksBySkuIDs(ctx context.Context, skuIDs []int64) (map[int64]*Stock, error) {
	return m.getStocksBySkuIDs(ctx, skuIDs)
}

// mockSpecGroupRepo 规格组仓储模拟实现
type mockSpecGroupRepo struct {
	createSpecGroups            func(ctx context.Context, productID int64, groups []*SpecGroup) error
	replaceSpecGroups           func(ctx context.Context, productID int64, groups []*SpecGroup) error
	deleteSpecGroupsByProductID func(ctx context.Context, productID int64) error
	getSpecGroupsByProductID    func(ctx context.Context, productID int64) ([]*SpecGroup, error)
	getSpecGroupsByProductIDs   func(ctx context.Context, productIDs []int64) (map[int64][]*SpecGroup, error)
}

func (m *mockSpecGroupRepo) CreateSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error {
	return m.createSpecGroups(ctx, productID, groups)
}
func (m *mockSpecGroupRepo) ReplaceSpecGroups(ctx context.Context, productID int64, groups []*SpecGroup) error {
	return m.replaceSpecGroups(ctx, productID, groups)
}
func (m *mockSpecGroupRepo) DeleteSpecGroupsByProductID(ctx context.Context, productID int64) error {
	return m.deleteSpecGroupsByProductID(ctx, productID)
}
func (m *mockSpecGroupRepo) GetSpecGroupsByProductID(ctx context.Context, productID int64) ([]*SpecGroup, error) {
	return m.getSpecGroupsByProductID(ctx, productID)
}
func (m *mockSpecGroupRepo) GetSpecGroupsByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*SpecGroup, error) {
	return m.getSpecGroupsByProductIDs(ctx, productIDs)
}

// mockSkuRepo SKU仓储模拟实现
type mockSkuRepo struct {
	createSkus            func(ctx context.Context, skus []*Sku) error
	replaceSkus           func(ctx context.Context, productID int64, skus []*Sku) error
	deleteSkusByProductID func(ctx context.Context, productID int64) error
	getSkusByProductID    func(ctx context.Context, productID int64) ([]*Sku, error)
	getSkusByProductIDs   func(ctx context.Context, productIDs []int64) (map[int64][]*Sku, error)
}

func (m *mockSkuRepo) CreateSkus(ctx context.Context, skus []*Sku) error {
	return m.createSkus(ctx, skus)
}
func (m *mockSkuRepo) ReplaceSkus(ctx context.Context, productID int64, skus []*Sku) error {
	return m.replaceSkus(ctx, productID, skus)
}
func (m *mockSkuRepo) DeleteSkusByProductID(ctx context.Context, productID int64) error {
	return m.deleteSkusByProductID(ctx, productID)
}
func (m *mockSkuRepo) GetSkusByProductID(ctx context.Context, productID int64) ([]*Sku, error) {
	return m.getSkusByProductID(ctx, productID)
}
func (m *mockSkuRepo) GetSkusByProductIDs(ctx context.Context, productIDs []int64) (map[int64][]*Sku, error) {
	return m.getSkusByProductIDs(ctx, productIDs)
}

// newDefaultMockSkuRepo 创建默认的 mockSkuRepo，所有方法返回空结果（避免 nil 指针 panic）
func newDefaultMockSkuRepo() mockSkuRepo {
	return mockSkuRepo{
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			return []*Sku{}, nil
		},
	}
}

// newProductTestUseCase 创建测试用的ProductUseCase
func newProductTestUseCase(
	pr mockProductRepo,
	sr mockStockRepo,
	sgr mockSpecGroupRepo,
	skur mockSkuRepo,
) *ProductUseCase {
	return NewProductUseCase(&pr, &sr, &sgr, &skur, log.NewStdLogger(io.Discard))
}

// TestCreateProduct_Success 测试创建商品成功（含规格组和SKU）
func TestCreateProduct_Success(t *testing.T) {
	var specGroupCalled, skuCalled bool
	stockCallCount := 0
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 1 // 模拟数据库分配ID
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		createSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			specGroupCalled = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	skur := mockSkuRepo{
		createSkus: func(ctx context.Context, skus []*Sku) error {
			skuCalled = true
			// 模拟 SKU 分配 ID
			for i, sku := range skus {
				sku.ID = int64(i + 1)
			}
			return nil
		},
	}
	sr := mockStockRepo{
		createStock: func(ctx context.Context, in *Stock) error {
			stockCallCount++
			assert.Equal(t, int64(1), in.ProductID)
			if stockCallCount == 1 {
				assert.Equal(t, int64(1), in.SkuID, "第一个SKU的库存")
				assert.Equal(t, int32(0), in.Total)
				assert.Equal(t, int32(0), in.Available)
			} else {
				assert.Equal(t, int64(2), in.SkuID, "第二个SKU的库存")
				assert.Equal(t, int32(0), in.Total)
				assert.Equal(t, int32(0), in.Available)
			}
			return nil
		},
	}

	uc := newProductTestUseCase(pr, sr, sgr, skur)
	in := &Product{
		Name:       "测试商品",
		Price:      100,
		Status:     ProductStatusOnSale,
		SpecGroups: []*SpecGroup{{Name: "颜色", Values: []string{"红", "蓝"}}},
		Skus:       []*Sku{{Price: 100}, {Price: 120}},
	}

	err := uc.CreateProduct(context.Background(), in)
	assert.NoError(t, err)
	assert.True(t, specGroupCalled, "规格组应被创建")
	assert.True(t, skuCalled, "SKU应被创建")
	assert.Equal(t, 2, stockCallCount, "每个SKU应创建一条库存记录")
}

// TestCreateProduct_RepoError 测试创建商品时仓储层报错
func TestCreateProduct_RepoError(t *testing.T) {
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			return assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{}, mockSpecGroupRepo{}, mockSkuRepo{})

	err := uc.CreateProduct(context.Background(), &Product{Name: "测试"})
	assert.True(t, v1.IsCreateProductFailed(err), "应返回ErrorCreateProductFailed错误")
}

// TestCreateProduct_SpecGroupError 测试创建商品时规格组创建失败
func TestCreateProduct_SpecGroupError(t *testing.T) {
	var productDeleted bool
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 1
			return nil
		},
		deleteProduct: func(ctx context.Context, productID int64) error {
			productDeleted = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		createSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			return assert.AnError
		},
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	skur := mockSkuRepo{
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}, sgr, skur)

	err := uc.CreateProduct(context.Background(), &Product{
		SpecGroups: []*SpecGroup{{Name: "颜色"}},
	})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.True(t, productDeleted, "规格组创建失败时应回滚删除商品")
}

// TestCreateProduct_SkuError 测试创建商品时SKU创建失败
func TestCreateProduct_SkuError(t *testing.T) {
	var productDeleted bool
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 1
			return nil
		},
		deleteProduct: func(ctx context.Context, productID int64) error {
			productDeleted = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	skur := mockSkuRepo{
		createSkus: func(ctx context.Context, skus []*Sku) error {
			return assert.AnError
		},
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}, sgr, skur)

	err := uc.CreateProduct(context.Background(), &Product{
		Skus: []*Sku{{Price: 100}},
	})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.True(t, productDeleted, "SKU创建失败时应回滚删除商品")
}

// TestCreateProduct_StockError 测试创建商品时库存初始化失败
func TestCreateProduct_StockError(t *testing.T) {
	var productDeleted bool
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 1
			return nil
		},
		deleteProduct: func(ctx context.Context, productID int64) error {
			productDeleted = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	sr := mockStockRepo{
		createStock: func(ctx context.Context, in *Stock) error {
			return assert.AnError
		},
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}
	sgr := mockSpecGroupRepo{
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	skur := mockSkuRepo{
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			return nil
		},
	}
	uc := newProductTestUseCase(pr, sr, sgr, skur)

	err := uc.CreateProduct(context.Background(), &Product{Name: "测试"})
	assert.True(t, v1.IsCreateStockFailed(err), "应返回ErrorCreateStockFailed错误")
	assert.True(t, productDeleted, "库存初始化失败时应回滚删除商品")
}

// TestCreateProduct_NoSpecsNoSkus 测试创建商品时无规格组和SKU，创建 product 级库存
func TestCreateProduct_NoSpecsNoSkus(t *testing.T) {
	var specGroupCalled, skuCalled bool
	stockCallCount := 0
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 1
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		createSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			specGroupCalled = true
			return nil
		},
	}
	skur := mockSkuRepo{
		createSkus: func(ctx context.Context, skus []*Sku) error {
			skuCalled = true
			return nil
		},
	}
	sr := mockStockRepo{
		createStock: func(ctx context.Context, in *Stock) error {
			stockCallCount++
			assert.Equal(t, int64(0), in.SkuID, "无SKU时 sku_id 应为0")
			return nil
		},
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}
	uc := newProductTestUseCase(pr, sr, sgr, skur)

	err := uc.CreateProduct(context.Background(), &Product{Name: "简单商品"})
	assert.NoError(t, err)
	assert.False(t, specGroupCalled, "无规格组时不应调用CreateSpecGroups")
	assert.False(t, skuCalled, "无SKU时不应调用CreateSkus")
	assert.Equal(t, 1, stockCallCount, "无SKU时应创建一条 product 级库存")
}

// TestCreateProduct_SkuProductIDSet 测试创建商品时SKU的ProductID会被正确设置
func TestCreateProduct_SkuProductIDSet(t *testing.T) {
	pr := mockProductRepo{
		createProduct: func(ctx context.Context, in *Product) error {
			in.ID = 42
			return nil
		},
	}
	skur := mockSkuRepo{
		createSkus: func(ctx context.Context, skus []*Sku) error {
			for _, sku := range skus {
				assert.Equal(t, int64(42), sku.ProductID, "SKU的ProductID应被设置为商品ID")
			}
			return nil
		},
	}
	sr := mockStockRepo{
		createStock:            func(ctx context.Context, in *Stock) error { return nil },
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}
	uc := newProductTestUseCase(pr, sr, mockSpecGroupRepo{}, skur)

	in := &Product{
		Name: "测试商品",
		Skus: []*Sku{
			{Price: 100},
			{Price: 120},
		},
	}
	err := uc.CreateProduct(context.Background(), in)
	assert.NoError(t, err)
}

// TestUpdateProduct_Success 测试更新商品成功（含规格组和SKU替换）
func TestUpdateProduct_Success(t *testing.T) {
	var replaceSpecCalled, replaceSkuCalled bool
	pr := mockProductRepo{
		updateProduct: func(ctx context.Context, in *Product) error {
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		replaceSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			replaceSpecCalled = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	skur := mockSkuRepo{
		replaceSkus: func(ctx context.Context, productID int64, skus []*Sku) error {
			replaceSkuCalled = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			return []*Sku{}, nil
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
		createStock:            func(ctx context.Context, in *Stock) error { return nil },
	}, sgr, skur)

	err := uc.UpdateProduct(context.Background(), &Product{
		ID:         1,
		Name:       "更新商品",
		SpecGroups: []*SpecGroup{{Name: "尺码", Values: []string{"S", "M"}}},
		Skus:       []*Sku{{Price: 200}},
	})
	assert.NoError(t, err)
	assert.True(t, replaceSpecCalled, "规格组应被替换")
	assert.True(t, replaceSkuCalled, "SKU应被替换")
}

// TestUpdateProduct_RepoError 测试更新商品时仓储层报错
func TestUpdateProduct_RepoError(t *testing.T) {
	pr := mockProductRepo{
		updateProduct: func(ctx context.Context, in *Product) error {
			return assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{}, mockSpecGroupRepo{}, newDefaultMockSkuRepo())

	err := uc.UpdateProduct(context.Background(), &Product{ID: 1, Name: "测试"})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// TestUpdateProduct_ReplaceSpecGroupError 测试更新商品时规格组替换失败
func TestUpdateProduct_ReplaceSpecGroupError(t *testing.T) {
	pr := mockProductRepo{
		updateProduct: func(ctx context.Context, in *Product) error {
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		replaceSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			return assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{}, sgr, newDefaultMockSkuRepo())

	err := uc.UpdateProduct(context.Background(), &Product{
		ID:         1,
		SpecGroups: []*SpecGroup{{Name: "颜色"}},
	})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// TestUpdateProduct_ReplaceSkuError 测试更新商品时SKU替换失败
func TestUpdateProduct_ReplaceSkuError(t *testing.T) {
	pr := mockProductRepo{
		updateProduct: func(ctx context.Context, in *Product) error {
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		replaceSpecGroups: func(ctx context.Context, productID int64, groups []*SpecGroup) error {
			return nil // 规格组替换成功
		},
	}
	skur := mockSkuRepo{
		replaceSkus: func(ctx context.Context, productID int64, skus []*Sku) error {
			return assert.AnError
		},
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			return []*Sku{}, nil
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}, sgr, skur)

	err := uc.UpdateProduct(context.Background(), &Product{
		ID:   1,
		Skus: []*Sku{{Price: 100}},
	})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

// TestDeleteProduct_Success 测试删除商品成功，关联数据应先被清理
func TestDeleteProduct_Success(t *testing.T) {
	var specDeleted, skuDeleted, productDeleted bool
	pr := mockProductRepo{
		deleteProduct: func(ctx context.Context, productID int64) error {
			productDeleted = true
			assert.True(t, specDeleted, "规格组应在商品删除前被删除")
			assert.True(t, skuDeleted, "SKU应在商品删除前被删除")
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			specDeleted = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	skur := mockSkuRepo{
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			skuDeleted = true
			assert.Equal(t, int64(1), productID)
			return nil
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}, sgr, skur)

	err := uc.DeleteProduct(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, specDeleted, "规格组应被删除")
	assert.True(t, skuDeleted, "SKU应被删除")
	assert.True(t, productDeleted, "商品应被删除")
}

// TestDeleteProduct_AssociationError 测试删除商品时关联数据删除失败，商品仍应被删除
func TestDeleteProduct_AssociationError(t *testing.T) {
	var productDeleted bool
	pr := mockProductRepo{
		deleteProduct: func(ctx context.Context, productID int64) error {
			productDeleted = true
			return nil
		},
	}
	sgr := mockSpecGroupRepo{
		deleteSpecGroupsByProductID: func(ctx context.Context, productID int64) error {
			return assert.AnError
		},
	}
	skur := mockSkuRepo{
		deleteSkusByProductID: func(ctx context.Context, productID int64) error {
			return assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{
		deleteStockByProductID: func(ctx context.Context, productID int64) error { return nil },
	}, sgr, skur)

	err := uc.DeleteProduct(context.Background(), 1)
	assert.NoError(t, err, "关联数据删除失败不影响商品删除")
	assert.True(t, productDeleted, "即使关联数据删除失败，商品也应被删除")
}

// TestGetProduct_Success 测试获取商品详情成功，规格组、SKU 和库存应被加载
func TestGetProduct_Success(t *testing.T) {
	pr := mockProductRepo{
		getProduct: func(ctx context.Context, productID int64) (*Product, error) {
			return &Product{ID: 1, Name: "测试商品"}, nil
		},
	}
	sgr := mockSpecGroupRepo{
		getSpecGroupsByProductID: func(ctx context.Context, productID int64) ([]*SpecGroup, error) {
			return []*SpecGroup{{ProductID: 1, Name: "颜色"}}, nil
		},
	}
	skur := mockSkuRepo{
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			return []*Sku{{ProductID: 1, Price: 100, ID: 10}}, nil
		},
	}
	sr := mockStockRepo{
		getStock: func(ctx context.Context, skuID int64) (*Stock, error) {
			assert.Equal(t, int64(10), skuID)
			return &Stock{SkuID: 10, Available: 50}, nil
		},
	}
	uc := newProductTestUseCase(pr, sr, sgr, skur)

	product, err := uc.GetProduct(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), product.ID)
	assert.Len(t, product.SpecGroups, 1, "应加载规格组")
	assert.Equal(t, "颜色", product.SpecGroups[0].Name)
	assert.Len(t, product.Skus, 1, "应加载SKU")
	assert.Equal(t, int64(100), product.Skus[0].Price)
	assert.Equal(t, int32(50), product.Skus[0].Stock, "SKU库存应被加载")
}

// TestGetProduct_RepoError 测试获取商品时仓储层报错
func TestGetProduct_RepoError(t *testing.T) {
	pr := mockProductRepo{
		getProduct: func(ctx context.Context, productID int64) (*Product, error) {
			return nil, assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, mockStockRepo{}, mockSpecGroupRepo{}, mockSkuRepo{})

	product, err := uc.GetProduct(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, product)
}

// TestGetProduct_SpecGroupError 测试获取商品时规格组加载失败，错误被静默忽略
func TestGetProduct_SpecGroupError(t *testing.T) {
	pr := mockProductRepo{
		getProduct: func(ctx context.Context, productID int64) (*Product, error) {
			return &Product{ID: 1, Name: "测试商品"}, nil
		},
	}
	sgr := mockSpecGroupRepo{
		getSpecGroupsByProductID: func(ctx context.Context, productID int64) ([]*SpecGroup, error) {
			return nil, assert.AnError
		},
	}
	skur := mockSkuRepo{
		getSkusByProductID: func(ctx context.Context, productID int64) ([]*Sku, error) {
			return []*Sku{{ProductID: 1, Price: 100}}, nil
		},
	}
	sr := mockStockRepo{
		getStock: func(ctx context.Context, skuID int64) (*Stock, error) {
			// SKU没有ID，GetStock会返回错误，应被静默忽略
			return nil, assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, sr, sgr, skur)

	product, err := uc.GetProduct(context.Background(), 1)
	assert.NoError(t, err, "规格组加载错误应被静默忽略")
	assert.Nil(t, product.SpecGroups, "规格组加载失败时应为nil")
	assert.Len(t, product.Skus, 1, "SKU应正常加载")
}

// TestListProduct_Success 测试获取商品列表成功，使用批量加载避免N+1
func TestListProduct_Success(t *testing.T) {
	pr := mockProductRepo{
		listProduct: func(ctx context.Context, page, pageSize int32, keyword string, categoryID int64, status string, stockEQ int32) ([]*Product, int32, error) {
			return []*Product{
				{ID: 1, Name: "商品1"},
				{ID: 2, Name: "商品2"},
			}, 2, nil
		},
	}
	sgr := mockSpecGroupRepo{
		getSpecGroupsByProductIDs: func(ctx context.Context, productIDs []int64) (map[int64][]*SpecGroup, error) {
			result := make(map[int64][]*SpecGroup)
			for _, pid := range productIDs {
				result[pid] = []*SpecGroup{{ProductID: pid, Name: "规格"}}
			}
			return result, nil
		},
	}
	skur := mockSkuRepo{
		getSkusByProductIDs: func(ctx context.Context, productIDs []int64) (map[int64][]*Sku, error) {
			result := make(map[int64][]*Sku)
			for _, pid := range productIDs {
				result[pid] = []*Sku{{ProductID: pid, Price: 100}}
			}
			return result, nil
		},
	}
	sr := mockStockRepo{
		getStocksBySkuIDs: func(ctx context.Context, skuIDs []int64) (map[int64]*Stock, error) {
			// 库存加载失败应被静默忽略
			return nil, assert.AnError
		},
	}
	uc := newProductTestUseCase(pr, sr, sgr, skur)

	products, total, err := uc.ListProduct(context.Background(), 1, 10, "", 0, "", -1)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, products, 2)
	for _, p := range products {
		assert.Len(t, p.SpecGroups, 1, "每个商品应加载规格组")
		assert.Len(t, p.Skus, 1, "每个商品应加载SKU")
	}
}
