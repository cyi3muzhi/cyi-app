package biz

import (
	"context"
	"io"
	"testing"

	v1 "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// OrderUseCase 测试专用 mock
// ---------------------------------------------------------------------------

// mockProductRepoForOrder 用于 OrderUseCase 测试的商品 mock
type mockProductRepoForOrder struct {
	getProductFn func(ctx context.Context, productID int64) (*productRpc.GetProductReply, error)
}

func (m *mockProductRepoForOrder) GetProduct(ctx context.Context, productID int64) (*productRpc.GetProductReply, error) {
	if m.getProductFn != nil {
		return m.getProductFn(ctx, productID)
	}
	return nil, nil
}

// mockStockRepoForOrder 用于 OrderUseCase 测试的库存 mock
type mockStockRepoForOrder struct {
	getStockFn     func(ctx context.Context, productID int64) (*productRpc.GetStockReply, error)
	reserveStockFn func(ctx context.Context, productID int64, quantity int32, version int32) (int32, error)
	deductStockFn  func(ctx context.Context, productID int64, quantity int32, version int32) (int32, error)
}

func (m *mockStockRepoForOrder) GetStock(ctx context.Context, productID int64) (*productRpc.GetStockReply, error) {
	if m.getStockFn != nil {
		return m.getStockFn(ctx, productID)
	}
	return &productRpc.GetStockReply{Total: 100, Available: 50, Reserved: 50, Version: 1}, nil
}

func (m *mockStockRepoForOrder) ReserveStock(ctx context.Context, productID int64, quantity int32, version int32) (int32, error) {
	if m.reserveStockFn != nil {
		return m.reserveStockFn(ctx, productID, quantity, version)
	}
	return 2, nil
}

func (m *mockStockRepoForOrder) DeductStock(ctx context.Context, productID int64, quantity int32, version int32) (int32, error) {
	if m.deductStockFn != nil {
		return m.deductStockFn(ctx, productID, quantity, version)
	}
	return 2, nil
}

func (m *mockStockRepoForOrder) ReleaseStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}

func (m *mockStockRepoForOrder) AddStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}

// mockOrderRepoForOrder 用于 OrderUseCase 测试的订单 mock
type mockOrderRepoForOrder struct {
	createOrderFn func(ctx context.Context, in *Order) error
	updateOrderFn func(ctx context.Context, in *Order) error
	getOrderFn    func(ctx context.Context, id int64) (*Order, error)
}

func (m *mockOrderRepoForOrder) CreateOrder(ctx context.Context, in *Order) error {
	if m.createOrderFn != nil {
		return m.createOrderFn(ctx, in)
	}
	return nil
}

func (m *mockOrderRepoForOrder) UpdateOrder(ctx context.Context, in *Order) error {
	if m.updateOrderFn != nil {
		return m.updateOrderFn(ctx, in)
	}
	return nil
}

func (m *mockOrderRepoForOrder) GetOrder(ctx context.Context, id int64) (*Order, error) {
	if m.getOrderFn != nil {
		return m.getOrderFn(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderRepoForOrder) ListOrder(_ context.Context, _ int32, _ int32, _ string, _ string, _ string, _ string) ([]*Order, int32, error) {
	return nil, 0, nil
}

func (m *mockOrderRepoForOrder) OrderStats(_ context.Context) (int32, int32, int32, int32, error) {
	return 0, 0, 0, 0, nil
}

// newTestOrderUseCase 快速构建测试用 OrderUseCase
func newTestOrderUseCase(
	repo *mockOrderRepoForOrder,
	productRe *mockProductRepoForOrder,
	stockRe *mockStockRepoForOrder,
) *OrderUseCase {
	logger := log.NewStdLogger(io.Discard)
	return NewOrderUseCase(repo, productRe, stockRe, logger)
}

// ---------------------------------------------------------------------------
// Create 测试
// ---------------------------------------------------------------------------

// TestOrderUseCase_Create_Success 验证正常创建订单流程：商品校验通过、库存充足、预留成功
func TestOrderUseCase_Create_Success(t *testing.T) {
	repo := &mockOrderRepoForOrder{}
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{
					ProductId: 1,
					Name:      "测试商品",
					Price:     9900,
					Status:    1,
				},
			}, nil
		},
	}
	stockRe := &mockStockRepoForOrder{}
	uc := newTestOrderUseCase(repo, productRe, stockRe)

	order := &Order{
		UserID: 100,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2},
		},
	}

	err := uc.Create(context.Background(), order)
	assert.NoError(t, err)
	assert.Equal(t, v1.OrderStatus_WaitPay, order.Status)
}

// TestOrderUseCase_Create_EmptyItems 验证空商品列表返回 OrderProductEmpty 错误
func TestOrderUseCase_Create_EmptyItems(t *testing.T) {
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, &mockProductRepoForOrder{}, &mockStockRepoForOrder{})

	order := &Order{UserID: 100, Items: []*OrderItem{}}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsOrderProductEmpty(err))
}

// TestOrderUseCase_Create_ProductNotFound 验证商品查询失败返回 ProductNotAvailable 错误
func TestOrderUseCase_Create_ProductNotFound(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return nil, v1.ErrorProductNotAvailable("product not found")
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, &mockStockRepoForOrder{})

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsProductNotAvailable(err))
}

// TestOrderUseCase_Create_ProductOffShelf 验证商品已下架返回 ProductNotAvailable 错误
func TestOrderUseCase_Create_ProductOffShelf(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{
					ProductId: 1,
					Name:      "下架商品",
					Price:     9900,
					Status:    0, // 已下架
				},
			}, nil
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, &mockStockRepoForOrder{})

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsProductNotAvailable(err))
}

// TestOrderUseCase_Create_StockNotFound 验证库存查询失败返回 ProductNotAvailable 错误
func TestOrderUseCase_Create_StockNotFound(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{ProductId: 1, Name: "商品", Price: 9900, Status: 1},
			}, nil
		},
	}
	stockRe := &mockStockRepoForOrder{
		getStockFn: func(_ context.Context, _ int64) (*productRpc.GetStockReply, error) {
			return nil, v1.ErrorProductNotAvailable("stock not found")
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, stockRe)

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsProductNotAvailable(err))
}

// TestOrderUseCase_Create_InsufficientStock 验证库存不足返回 InsufficientStock 错误
func TestOrderUseCase_Create_InsufficientStock(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{ProductId: 1, Name: "商品", Price: 9900, Status: 1},
			}, nil
		},
	}
	stockRe := &mockStockRepoForOrder{
		getStockFn: func(_ context.Context, _ int64) (*productRpc.GetStockReply, error) {
			return &productRpc.GetStockReply{Total: 100, Available: 3, Reserved: 97, Version: 1}, nil
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, stockRe)

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 5}}, // 需要 5 个但只有 3 个可用
	}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsInsufficientStock(err))
}

// TestOrderUseCase_Create_ReserveStockFailed 验证预留库存失败返回 StockOperationFailed 错误
func TestOrderUseCase_Create_ReserveStockFailed(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{ProductId: 1, Name: "商品", Price: 9900, Status: 1},
			}, nil
		},
	}
	stockRe := &mockStockRepoForOrder{
		reserveStockFn: func(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
			return 0, v1.ErrorStockOperationFailed("concurrent conflict")
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, stockRe)

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := uc.Create(context.Background(), order)
	assert.True(t, v1.IsStockOperationFailed(err))
}

// TestOrderUseCase_Create_SetsItemInfo 验证创建订单时将商品的名称和价格复制到订单项
func TestOrderUseCase_Create_SetsItemInfo(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, _ int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{ProductId: 1, Name: "精品咖啡", Price: 29900, Status: 1},
			}, nil
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, &mockStockRepoForOrder{})

	order := &Order{
		UserID: 100,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := uc.Create(context.Background(), order)
	assert.NoError(t, err)
	assert.Equal(t, "精品咖啡", order.Items[0].ProductName)
	assert.Equal(t, int64(29900), order.Items[0].Price)
}

// TestOrderUseCase_Create_MultipleItems 验证多商品订单全部校验通过并预留库存
func TestOrderUseCase_Create_MultipleItems(t *testing.T) {
	productRe := &mockProductRepoForOrder{
		getProductFn: func(_ context.Context, productID int64) (*productRpc.GetProductReply, error) {
			return &productRpc.GetProductReply{
				Product: &productRpc.ProductItem{
					ProductId: productID,
					Name:      "商品A",
					Price:     1000,
					Status:    1,
				},
			}, nil
		},
	}
	reserveCalls := make([]int32, 0)
	stockRe := &mockStockRepoForOrder{
		reserveStockFn: func(_ context.Context, _ int64, quantity int32, _ int32) (int32, error) {
			reserveCalls = append(reserveCalls, quantity)
			return 2, nil
		},
	}
	uc := newTestOrderUseCase(&mockOrderRepoForOrder{}, productRe, stockRe)

	order := &Order{
		UserID: 100,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 3},
		},
	}
	err := uc.Create(context.Background(), order)
	assert.NoError(t, err)
	assert.Equal(t, v1.OrderStatus_WaitPay, order.Status)
	assert.Len(t, reserveCalls, 2)
	assert.Equal(t, int32(2), reserveCalls[0])
	assert.Equal(t, int32(3), reserveCalls[1])
}

// ---------------------------------------------------------------------------
// Pay 测试
// ---------------------------------------------------------------------------

// TestOrderUseCase_Pay_Success 验证正常支付订单并扣减库存
func TestOrderUseCase_Pay_Success(t *testing.T) {
	// 准备一个 WaitPay 状态的订单
	testOrder := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "商品", Price: 1000},
		},
	}

	repo := &mockOrderRepoForOrder{
		getOrderFn: func(_ context.Context, _ int64) (*Order, error) {
			// 每次返回一个新的 WaitPay 订单副本，模拟数据库查询
			o := &Order{
				ID:     testOrder.ID,
				UserID: testOrder.UserID,
				Status: v1.OrderStatus_WaitPay,
				Items:  testOrder.Items,
			}
			return o, nil
		},
	}
	stockRe := &mockStockRepoForOrder{}
	uc := newTestOrderUseCase(repo, &mockProductRepoForOrder{}, stockRe)

	err := uc.Pay(context.Background(), 1, 100)
	assert.NoError(t, err)
}

// TestOrderUseCase_Pay_OrderNotFound 验证订单不存在时支付返回错误
func TestOrderUseCase_Pay_OrderNotFound(t *testing.T) {
	repo := &mockOrderRepoForOrder{
		getOrderFn: func(_ context.Context, _ int64) (*Order, error) {
			return nil, v1.ErrorOrderNotFound("order not found")
		},
	}
	uc := newTestOrderUseCase(repo, &mockProductRepoForOrder{}, &mockStockRepoForOrder{})

	err := uc.Pay(context.Background(), 999, 100)
	assert.True(t, v1.IsOrderNotFound(err))
}

// TestOrderUseCase_Pay_WrongOwner 验证非订单所有者支付返回 UserNotOrder 错误
func TestOrderUseCase_Pay_WrongOwner(t *testing.T) {
	repo := &mockOrderRepoForOrder{
		getOrderFn: func(_ context.Context, _ int64) (*Order, error) {
			o := &Order{
				ID:     1,
				UserID: 100, // 订单属于用户 100
				Status: v1.OrderStatus_WaitPay,
				Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
			}
			return o, nil
		},
	}
	uc := newTestOrderUseCase(repo, &mockProductRepoForOrder{}, &mockStockRepoForOrder{})

	// 用户 200 尝试支付属于用户 100 的订单
	err := uc.Pay(context.Background(), 1, 200)
	assert.True(t, v1.IsUserNotOrder(err))
}

// TestOrderUseCase_Pay_StockDeductFailed 验证支付成功但扣减库存失败时返回 StockOperationFailed 错误
func TestOrderUseCase_Pay_StockDeductFailed(t *testing.T) {
	repo := &mockOrderRepoForOrder{
		getOrderFn: func(_ context.Context, _ int64) (*Order, error) {
			o := &Order{
				ID:     1,
				UserID: 100,
				Status: v1.OrderStatus_WaitPay,
				Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
			}
			return o, nil
		},
	}
	stockRe := &mockStockRepoForOrder{
		deductStockFn: func(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
			return 0, v1.ErrorStockOperationFailed("deduct conflict")
		},
	}
	uc := newTestOrderUseCase(repo, &mockProductRepoForOrder{}, stockRe)

	err := uc.Pay(context.Background(), 1, 100)
	assert.True(t, v1.IsStockOperationFailed(err))
}

// ---------------------------------------------------------------------------
// Update 测试
// ---------------------------------------------------------------------------

// TestOrderUseCase_Update_SetsHandlerWhenNil 验证 Update 在 statusHandler 为 nil 时自动调用 SetStatus
func TestOrderUseCase_Update_SetsHandlerWhenNil(t *testing.T) {
	updateCalled := false
	repo := &mockOrderRepoForOrder{
		updateOrderFn: func(_ context.Context, in *Order) error {
			updateCalled = true
			// 验证此时 statusHandler 已被初始化
			if in.statusHandler == nil {
				t.Error("UpdateOrder 调用时 statusHandler 不应为 nil")
			}
			return nil
		},
	}
	uc := newTestOrderUseCase(repo, &mockProductRepoForOrder{}, &mockStockRepoForOrder{})

	// 构造一个 statusHandler 为 nil 的订单
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items:  []*OrderItem{{ProductID: 1, Quantity: 1}},
	}
	// statusHandler 初始为 nil
	assert.Nil(t, order.statusHandler)

	err := uc.Update(context.Background(), order)
	assert.NoError(t, err)
	assert.True(t, updateCalled)
	// Update 后 statusHandler 应已被 SetStatus 初始化
	assert.NotNil(t, order.statusHandler)
}
