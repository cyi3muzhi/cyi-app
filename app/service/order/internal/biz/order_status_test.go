package biz

import (
	"context"
	"io"
	"testing"

	productRpc "cyi-app/api/service/product/v1"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// mockOrderRepo 用于状态机测试的内存 mock 仓储
type mockOrderRepo struct {
	updateOrderFn     func(ctx context.Context, in *Order) error
	updateStatusIfFn  func(ctx context.Context, id int64, expectStatus, newStatus int32) error
}

func (m *mockOrderRepo) CreateOrder(_ context.Context, _ *Order) error { return nil }
func (m *mockOrderRepo) UpdateOrder(ctx context.Context, in *Order) error {
	if m.updateOrderFn != nil {
		return m.updateOrderFn(ctx, in)
	}
	return nil
}
func (m *mockOrderRepo) GetOrder(_ context.Context, _ int64) (*Order, error) { return nil, nil }

// BatchGetOrder 批量获取订单详情（含商品明细），状态机测试不依赖该方法，返回空切片以满足接口契约
func (m *mockOrderRepo) BatchGetOrder(_ context.Context, _ []int64) ([]*Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) ListOrder(_ context.Context, _ int32, _ int32, _ string, _ string, _ string, _ string, _ int64) ([]*Order, int32, error) {
	return nil, 0, nil
}

func (m *mockOrderRepo) OrderStats(_ context.Context) (int32, int32, int32, int32, error) {
	return 0, 0, 0, 0, nil
}

func (m *mockOrderRepo) GetOrderTrend(_ context.Context, _ int32) ([]*TrendItem, error) {
	return nil, nil
}

func (m *mockOrderRepo) GetOrderRevenueStats(_ context.Context) (int64, int64, error) {
	return 0, 0, nil
}

func (m *mockOrderRepo) GetUserOrderStats(_ context.Context, _ int64) (*UserOrderStats, error) {
	return &UserOrderStats{}, nil
}

func (m *mockOrderRepo) UpdateLogistics(_ context.Context, _ int64, _ string, _ string) error {
	return nil
}

// UpdateStatusIf 条件更新订单状态（状态机测试可注入 updateStatusIfFn 模拟 DB 失败/冲突）
func (m *mockOrderRepo) UpdateStatusIf(ctx context.Context, id int64, expectStatus, newStatus int32) error {
	if m.updateStatusIfFn != nil {
		return m.updateStatusIfFn(ctx, id, expectStatus, newStatus)
	}
	return nil
}

func (m *mockOrderRepo) CreatePaymentLog(_ context.Context, _ *PaymentLog) error {
	return nil
}

func (m *mockOrderRepo) ReleaseCoupon(_ context.Context, _ int64, _ int64, _ int64) error {
	return nil
}

func (m *mockOrderRepo) ListOverdueWaitPayOrders(_ context.Context, _ int32, _ int32, _ int64) ([]*Order, int32, error) {
	return nil, 0, nil
}

// mockStockRepo 用于状态机测试的 mock 库存仓储
type mockStockRepo struct{}

func (m *mockStockRepo) GetStock(_ context.Context, _ int64) (*productRpc.GetStockReply, error) {
	return &productRpc.GetStockReply{
		Total:     100,
		Available: 50,
		Reserved:  50,
		Version:   1,
	}, nil
}
func (m *mockStockRepo) ReserveStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}
func (m *mockStockRepo) DeductStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}
func (m *mockStockRepo) ReleaseStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}
func (m *mockStockRepo) AddStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}

// createTestOrder 创建指定状态的订单并初始化状态处理器
func createTestOrder(status v1.OrderStatus, repo *mockOrderRepo) *Order {
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: status,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 1, ProductName: "test", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, &mockStockRepo{}, log.NewStdLogger(io.Discard))
	return order
}

// assertStatus 辅助断言状态
func assertStatus(t *testing.T, got, want v1.OrderStatus) {
	t.Helper()
	if got != want {
		t.Errorf("期望状态 %v，实际 %v", want, got)
	}
}

// assertError 辅助断言有错误
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("期望有错误，实际没有")
	}
}

// assertNoError 辅助断言无错误
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("期望无错误，实际: %v", err)
	}
}

// ---------------------------------------------------------------------------
// WaitPay 状态测试
// ---------------------------------------------------------------------------

func TestOrder_WaitPay_Allowed(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_WaitPay, repo)

	t.Run("Pay → Pay", func(t *testing.T) {
		err := order.Pay(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Pay)
	})

	t.Run("Cancel → Cancel", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_WaitPay, repo)
		err := order.Cancel(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Cancel)
	})
}

func TestOrder_WaitPay_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Refunding 拒绝", func(o *Order) error { return o.Refunding(context.Background()) }},
		{"Refunded 拒绝", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_WaitPay, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_WaitPay) // 状态不变
		})
	}
}

// ---------------------------------------------------------------------------
// Pay / WaitShip 状态测试
// ---------------------------------------------------------------------------

func TestOrder_Pay_Allowed(t *testing.T) {
	repo := &mockOrderRepo{}

	t.Run("Ship → Ship", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Pay, repo)
		err := order.Ship(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Ship)
	})

	t.Run("Cancel → Cancel", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Pay, repo)
		err := order.Cancel(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Cancel)
	})

	t.Run("Refunding → Refunding", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Pay, repo)
		err := order.Refunding(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Refunding)
	})
}

func TestOrder_Pay_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝（重复支付）", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Refunded 拒绝（尚未退款）", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Pay, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Pay)
		})
	}
}

// WaitShip 共享 PayStatus 的相同行为
func TestOrder_WaitShip_Shared(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_WaitShip, repo)

	err := order.Ship(context.Background())
	assertNoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Ship)
}

// ---------------------------------------------------------------------------
// Ship / WaitReceive 状态测试
// ---------------------------------------------------------------------------

func TestOrder_Ship_Allowed(t *testing.T) {
	repo := &mockOrderRepo{}

	t.Run("Receive → Complete", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Ship, repo)
		err := order.Receive(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Complete)
	})

	t.Run("Refunding → Refunding", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Ship, repo)
		err := order.Refunding(context.Background())
		assertNoError(t, err)
		assertStatus(t, order.Status, v1.OrderStatus_Refunding)
	})
}

func TestOrder_Ship_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝（重复发货）", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunded 拒绝", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Ship, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Ship)
		})
	}
}

// WaitReceive 共享 ShipStatus 的相同行为
func TestOrder_WaitReceive_Shared(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_WaitReceive, repo)

	err := order.Receive(context.Background())
	assertNoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Complete)
}

// ---------------------------------------------------------------------------
// Complete 状态测试
// ---------------------------------------------------------------------------

func TestOrder_Complete_Allowed(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_Complete, repo)

	err := order.Refunding(context.Background())
	assertNoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Refunding)
}

func TestOrder_Complete_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝（重复完成）", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunded 拒绝", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Complete, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Complete)
		})
	}
}

// ---------------------------------------------------------------------------
// Cancel 状态测试（终态）
// ---------------------------------------------------------------------------

func TestOrder_Cancel_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝（重复取消）", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunding 拒绝", func(o *Order) error { return o.Refunding(context.Background()) }},
		{"Refunded 拒绝", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Cancel, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Cancel)
		})
	}
}

// ---------------------------------------------------------------------------
// Refunding 状态测试
// ---------------------------------------------------------------------------

func TestOrder_Refunding_Allowed(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_Refunding, repo)

	err := order.Refunded(context.Background())
	assertNoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Refunded)
}

func TestOrder_Refunding_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunding 拒绝（重复退款）", func(o *Order) error { return o.Refunding(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Refunding, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Refunding)
		})
	}
}

// ---------------------------------------------------------------------------
// Refunded 状态测试（终态）
// ---------------------------------------------------------------------------

func TestOrder_Refunded_Denied(t *testing.T) {
	repo := &mockOrderRepo{}

	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunding 拒绝", func(o *Order) error { return o.Refunding(context.Background()) }},
		{"Refunded 拒绝（重复退款）", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(v1.OrderStatus_Refunded, repo)
			assertError(t, tt.operation(order))
			assertStatus(t, order.Status, v1.OrderStatus_Refunded)
		})
	}
}

// ---------------------------------------------------------------------------
// 完整生命周期测试（正向流程 + 退款流程 + 取消流程）
// ---------------------------------------------------------------------------

func TestOrder_FullLifecycle_Normal(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_WaitPay, repo)

	// WaitPay → Pay → Ship → Complete
	assertNoError(t, order.Pay(context.Background()))
	assertStatus(t, order.Status, v1.OrderStatus_Pay)

	assertNoError(t, order.Ship(context.Background()))
	assertStatus(t, order.Status, v1.OrderStatus_Ship)

	assertNoError(t, order.Receive(context.Background()))
	assertStatus(t, order.Status, v1.OrderStatus_Complete)
}

func TestOrder_FullLifecycle_Refund(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_Pay, repo)

	// Pay → Refunding → Refunded
	assertNoError(t, order.Refunding(context.Background()))
	assertStatus(t, order.Status, v1.OrderStatus_Refunding)

	assertNoError(t, order.Refunded(context.Background()))
	assertStatus(t, order.Status, v1.OrderStatus_Refunded)
}

func TestOrder_FullLifecycle_Cancel(t *testing.T) {
	repo := &mockOrderRepo{}

	// 待支付时取消
	t.Run("WaitPay→Cancel", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_WaitPay, repo)
		assertNoError(t, order.Cancel(context.Background()))
		assertStatus(t, order.Status, v1.OrderStatus_Cancel)
	})

	// 已支付时取消
	t.Run("Pay→Cancel", func(t *testing.T) {
		order := createTestOrder(v1.OrderStatus_Pay, repo)
		assertNoError(t, order.Cancel(context.Background()))
		assertStatus(t, order.Status, v1.OrderStatus_Cancel)
	})
}

// ---------------------------------------------------------------------------
// 边界情况测试
// ---------------------------------------------------------------------------

func TestOrder_UnspecifiedStatus_Denied(t *testing.T) {
	// Unspecified(0) 应导致 statusHandler 为 nil，所有操作被拒绝
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_OrderStatusUnspecified,
	}
	// 不调用 SetStatus，模拟从未初始化

	// statusHandler 为 nil 时所有操作应返回错误
	tests := []struct {
		name      string
		operation func(order *Order) error
	}{
		{"Pay 拒绝", func(o *Order) error { return o.Pay(context.Background()) }},
		{"Ship 拒绝", func(o *Order) error { return o.Ship(context.Background()) }},
		{"Receive 拒绝", func(o *Order) error { return o.Receive(context.Background()) }},
		{"Complete 拒绝", func(o *Order) error { return o.Complete(context.Background()) }},
		{"Cancel 拒绝", func(o *Order) error { return o.Cancel(context.Background()) }},
		{"Refunding 拒绝", func(o *Order) error { return o.Refunding(context.Background()) }},
		{"Refunded 拒绝", func(o *Order) error { return o.Refunded(context.Background()) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, tt.operation(order))
		})
	}
}

// TestOrder_DefaultHandlerRejectsUnknown 验证 default 分支返回 nil，而非静默转为 WaitPay
func TestOrder_DefaultHandlerRejectsUnknown(t *testing.T) {
	repo := &mockOrderRepo{}
	logger := log.NewStdLogger(io.Discard)

	handler := NewOrderStatusHandler(v1.OrderStatus_OrderStatusUnspecified, repo, &mockStockRepo{}, logger)
	if handler != nil {
		t.Error("未知状态应返回 nil handler")
	}
}

// TestOrder_TransitionDBFailure 验证数据库失败时状态转换返回错误，且内存状态不被污染
func TestOrder_TransitionDBFailure(t *testing.T) {
	// 模拟 UpdateStatusIf（条件更新）失败
	repo := &mockOrderRepo{
		updateStatusIfFn: func(_ context.Context, _ int64, _ int32, _ int32) error {
			return v1.ErrorOrderStatusInvalid("db error")
		},
	}
	order := createTestOrder(v1.OrderStatus_WaitPay, repo)

	// 状态转换应该失败，且内存状态保持原状（成功后才切换状态处理器）
	oldStatus := order.Status
	err := order.Pay(context.Background())
	assertError(t, err)
	assertStatus(t, order.Status, oldStatus)
}

// TestOrder_SetStatus 验证 SetStatus 正确切换处理器
func TestOrder_SetStatus(t *testing.T) {
	repo := &mockOrderRepo{}
	logger := log.NewStdLogger(io.Discard)

	order := &Order{ID: 1, Status: v1.OrderStatus_WaitPay}
	order.SetStatus(order.Status, repo, &mockStockRepo{}, logger)

	if order.statusHandler == nil {
		t.Fatal("SetStatus 后 statusHandler 不应为 nil")
	}

	// 切换为 Cancel
	order.SetStatus(v1.OrderStatus_Cancel, repo, &mockStockRepo{}, logger)
	assertStatus(t, order.Status, v1.OrderStatus_Cancel)
	if order.statusHandler == nil {
		t.Fatal("切换到 Cancel 后 statusHandler 不应为 nil")
	}
}
