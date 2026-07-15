package biz

import (
	"context"
	"errors"
	"io"
	"testing"

	productRpc "cyi-app/api/service/product/v1"
	v1 "cyi-app/api/service/order/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// 库存操作边界测试 - 验证库存相关的错误处理和异常情况
// ---------------------------------------------------------------------------

// mockStockRepoWithErrors 用于测试库存操作错误的 mock
type mockStockRepoWithErrors struct {
	getStockErr    error
	releaseStockErr error
	addStockErr    error
}

func (m *mockStockRepoWithErrors) GetStock(_ context.Context, _ int64) (*productRpc.GetStockReply, error) {
	if m.getStockErr != nil {
		return nil, m.getStockErr
	}
	return &productRpc.GetStockReply{
		Total:     100,
		Available: 50,
		Reserved:  50,
		Version:   1,
	}, nil
}

func (m *mockStockRepoWithErrors) ReserveStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}

func (m *mockStockRepoWithErrors) DeductStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	return 2, nil
}

func (m *mockStockRepoWithErrors) ReleaseStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	if m.releaseStockErr != nil {
		return 0, m.releaseStockErr
	}
	return 2, nil
}

func (m *mockStockRepoWithErrors) AddStock(_ context.Context, _ int64, _ int32, _ int32) (int32, error) {
	if m.addStockErr != nil {
		return 0, m.addStockErr
	}
	return 2, nil
}

// TestWaitPay_Cancel_StockReleaseFailure 测试取消订单时库存释放失败
func TestWaitPay_Cancel_StockReleaseFailure(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepoWithErrors{
		releaseStockErr: errors.New("库存服务不可用"),
	}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 取消订单应该失败，因为库存释放失败
	err := order.Cancel(context.Background())
	assert.Error(t, err, "库存释放失败时应返回错误")
	assert.Contains(t, err.Error(), "库存服务不可用", "错误消息应包含原始错误")
	assertStatus(t, order.Status, v1.OrderStatus_WaitPay)
}

// TestPay_Cancel_StockReleaseFailure 测试已支付订单取消时库存释放失败
func TestPay_Cancel_StockReleaseFailure(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepoWithErrors{
		releaseStockErr: errors.New("释放库存失败"),
	}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_Pay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 取消订单应该失败，因为库存释放失败
	err := order.Cancel(context.Background())
	assert.Error(t, err, "库存释放失败时应返回错误")
	assertStatus(t, order.Status, v1.OrderStatus_Pay)
}

// TestPay_Cancel_AddStockFailure 测试已支付订单取消时增加库存失败
func TestPay_Cancel_AddStockFailure(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepoWithErrors{
		addStockErr: errors.New("增加库存失败"),
	}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_Pay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 取消订单应该失败，因为增加库存失败
	err := order.Cancel(context.Background())
	assert.Error(t, err, "增加库存失败时应返回错误")
	assert.Contains(t, err.Error(), "增加库存失败", "错误消息应包含原始错误")
	assertStatus(t, order.Status, v1.OrderStatus_Pay)
}

// TestWaitPay_Cancel_GetStockFailure 测试获取库存信息失败的情况
func TestWaitPay_Cancel_GetStockFailure(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepoWithErrors{
		getStockErr: errors.New("库存查询失败"),
	}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 取消订单应该失败，因为无法获取库存信息
	err := order.Cancel(context.Background())
	assert.Error(t, err, "获取库存失败时应返回错误")
	assert.Contains(t, err.Error(), "库存查询失败", "错误消息应包含原始错误")
}

// TestOrder_Cancel_MultipleItems 测试多商品订单的取消操作
func TestOrder_Cancel_MultipleItems(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 2, ProductName: "商品1", Price: 100},
			{ProductID: 2, Quantity: 3, ProductName: "商品2", Price: 200},
			{ProductID: 3, Quantity: 1, ProductName: "商品3", Price: 300},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 取消订单应成功处理多个商品
	err := order.Cancel(context.Background())
	assert.NoError(t, err, "多商品订单取消应成功")
	assertStatus(t, order.Status, v1.OrderStatus_Cancel)
}

// TestOrder_Cancel_EmptyItems 测试空商品列表的订单取消
func TestOrder_Cancel_EmptyItems(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}

	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items:  []*OrderItem{}, // 空商品列表
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 空商品列表的订单取消应成功
	err := order.Cancel(context.Background())
	assert.NoError(t, err, "空商品列表订单取消应成功")
	assertStatus(t, order.Status, v1.OrderStatus_Cancel)
}

// TestStatusBase_ReleaseStockForOrder_ConcurrentCalls 测试并发释放库存的安全性
func TestStatusBase_ReleaseStockForOrder_ConcurrentCalls(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}

	// 创建10个订单，并发执行取消操作
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			order := &Order{
				ID:     int64(id),
				UserID: 100,
				Status: v1.OrderStatus_WaitPay,
				Items: []*OrderItem{
					{ProductID: 1, Quantity: 1, ProductName: "测试商品", Price: 100},
				},
			}
			order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

			err := order.Cancel(context.Background())
			assert.NoError(t, err)

			done <- true
		}(i)
	}

	// 等待所有并发操作完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestOrder_StateTransitions_Complete 测试完整的状态转换路径
func TestOrder_StateTransitions_Complete(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}

	// 完整的生命周期：WaitPay → Pay → Ship → Complete
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 1, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 支付
	err := order.Pay(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Pay)

	// 发货
	err = order.Ship(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Ship)

	// 收货
	err = order.Receive(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Complete)

	// 已完成的订单不能取消
	err = order.Cancel(context.Background())
	assert.Error(t, err, "已完成的订单不能取消")
	assertStatus(t, order.Status, v1.OrderStatus_Complete)
}

// TestOrder_StateTransitions_Refund 测试退款流程的状态转换
func TestOrder_StateTransitions_Refund(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}

	// 退款流程：WaitPay → Pay → Refunding → Refunded
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		Items: []*OrderItem{
			{ProductID: 1, Quantity: 1, ProductName: "测试商品", Price: 100},
		},
	}
	order.SetStatus(order.Status, repo, stockRepo, log.NewStdLogger(io.Discard))

	// 支付
	err := order.Pay(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Pay)

	// 申请退款
	err = order.Refunding(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Refunding)

	// 完成退款
	err = order.Refunded(context.Background())
	assert.NoError(t, err)
	assertStatus(t, order.Status, v1.OrderStatus_Refunded)

	// 已退款的订单不能再退款
	err = order.Refunded(context.Background())
	assert.Error(t, err, "已退款的订单不能再次退款")
}

// TestGetStatusName_AllStatuses 测试状态名称映射的完整性
func TestGetStatusName_AllStatuses(t *testing.T) {
	tests := []struct {
		status   v1.OrderStatus
		expected string
	}{
		{v1.OrderStatus_WaitPay, "未支付"},
		{v1.OrderStatus_Pay, "已支付"},
		{v1.OrderStatus_WaitShip, "已支付"},
		{v1.OrderStatus_Ship, "已发货"},
		{v1.OrderStatus_WaitReceive, "已发货"},
		{v1.OrderStatus_Complete, "已完成"},
		{v1.OrderStatus_Cancel, "已取消"},
		{v1.OrderStatus_Refunding, "退款中"},
		{v1.OrderStatus_Refunded, "已退款"},
		{v1.OrderStatus_OrderStatusUnspecified, "状态异常"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := getStatusName(tt.status)
			assert.Equal(t, tt.expected, name)
		})
	}
}

// TestOrder_InvalidOperation_ErrorMessages 测试无效操作的错误消息格式
func TestOrder_InvalidOperation_ErrorMessages(t *testing.T) {
	repo := &mockOrderRepo{}
	order := createTestOrder(v1.OrderStatus_WaitPay, repo)

	// 尝试执行不允许的操作
	err := order.Ship(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未支付", "错误消息应包含当前状态")
	assert.Contains(t, err.Error(), "发货", "错误消息应包含尝试的操作")
	assert.Contains(t, err.Error(), "orderID=1", "错误消息应包含订单ID")
}

// TestNewOrderStatusHandler_AllStatuses 测试状态处理器工厂方法的完整性
func TestNewOrderStatusHandler_AllStatuses(t *testing.T) {
	repo := &mockOrderRepo{}
	stockRepo := &mockStockRepo{}
	logger := log.NewStdLogger(io.Discard)

	statuses := []v1.OrderStatus{
		v1.OrderStatus_WaitPay,
		v1.OrderStatus_Pay,
		v1.OrderStatus_WaitShip,
		v1.OrderStatus_Ship,
		v1.OrderStatus_WaitReceive,
		v1.OrderStatus_Complete,
		v1.OrderStatus_Cancel,
		v1.OrderStatus_Refunding,
		v1.OrderStatus_Refunded,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			handler := NewOrderStatusHandler(status, repo, stockRepo, logger)
			assert.NotNil(t, handler, "状态 %v 应有对应的处理器", status)
		})
	}

	// 测试未指定状态
	handler := NewOrderStatusHandler(v1.OrderStatus_OrderStatusUnspecified, repo, stockRepo, logger)
	assert.Nil(t, handler, "未指定状态应返回 nil 处理器")
}

// TestOrder_NilStatusHandler 测试未初始化状态处理器的订单操作
func TestOrder_NilStatusHandler(t *testing.T) {
	order := &Order{
		ID:     1,
		UserID: 100,
		Status: v1.OrderStatus_WaitPay,
		// 不设置 statusHandler，模拟未初始化状态
	}

	operations := []struct {
		name string
		fn   func() error
	}{
		{"Pay", func() error { return order.Pay(context.Background()) }},
		{"Ship", func() error { return order.Ship(context.Background()) }},
		{"Receive", func() error { return order.Receive(context.Background()) }},
		{"Complete", func() error { return order.Complete(context.Background()) }},
		{"Cancel", func() error { return order.Cancel(context.Background()) }},
		{"Refunding", func() error { return order.Refunding(context.Background()) }},
		{"Refunded", func() error { return order.Refunded(context.Background()) }},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			err := op.fn()
			assert.Error(t, err, "未初始化状态处理器时所有操作应返回错误")
			assert.Contains(t, err.Error(), "状态无效", "错误消息应说明状态无效")
		})
	}
}