package biz

import (
	"testing"
)

// TestMoney_Add 测试金额加法
func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name     string
		m1       Money
		m2       Money
		expected int64
	}{
		{"简单加法", Money{Amount: 100, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, 300},
		{"加零", Money{Amount: 100, Currency: "CNY"}, Money{Amount: 0, Currency: "CNY"}, 100},
		{"负数相加", Money{Amount: -100, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, 100},
		{"零加零", Money{Amount: 0, Currency: "CNY"}, Money{Amount: 0, Currency: "CNY"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m1.Add(tt.m2)
			if result.Amount != tt.expected {
				t.Errorf("expected amount %d, got %d", tt.expected, result.Amount)
			}
			if result.Currency != tt.m1.Currency {
				t.Errorf("expected currency %s, got %s", tt.m1.Currency, result.Currency)
			}
		})
	}
}

// TestMoney_Subtract 测试金额减法
func TestMoney_Subtract(t *testing.T) {
	tests := []struct {
		name     string
		m1       Money
		m2       Money
		expected int64
	}{
		{"简单减法", Money{Amount: 300, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, 100},
		{"减零", Money{Amount: 100, Currency: "CNY"}, Money{Amount: 0, Currency: "CNY"}, 100},
		{"结果为负", Money{Amount: 100, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, -100},
		{"零减零", Money{Amount: 0, Currency: "CNY"}, Money{Amount: 0, Currency: "CNY"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m1.Subtract(tt.m2)
			if result.Amount != tt.expected {
				t.Errorf("expected amount %d, got %d", tt.expected, result.Amount)
			}
			if result.Currency != tt.m1.Currency {
				t.Errorf("expected currency %s, got %s", tt.m1.Currency, result.Currency)
			}
		})
	}
}

// TestMoney_Multiply 测试金额乘法
func TestMoney_Multiply(t *testing.T) {
	tests := []struct {
		name     string
		m        Money
		n        int
		expected int64
	}{
		{"简单乘法", Money{Amount: 100, Currency: "CNY"}, 3, 300},
		{"乘以零", Money{Amount: 100, Currency: "CNY"}, 0, 0},
		{"乘以一", Money{Amount: 100, Currency: "CNY"}, 1, 100},
		{"负数相乘", Money{Amount: -100, Currency: "CNY"}, 3, -300},
		{"零乘零", Money{Amount: 0, Currency: "CNY"}, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m.Multiply(tt.n)
			if result.Amount != tt.expected {
				t.Errorf("expected amount %d, got %d", tt.expected, result.Amount)
			}
			if result.Currency != tt.m.Currency {
				t.Errorf("expected currency %s, got %s", tt.m.Currency, result.Currency)
			}
		})
	}
}

// TestMoney_LessThan 测试金额比较
func TestMoney_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		m1       Money
		m2       Money
		expected bool
	}{
		{"小于", Money{Amount: 100, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, true},
		{"大于", Money{Amount: 300, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, false},
		{"等于", Money{Amount: 200, Currency: "CNY"}, Money{Amount: 200, Currency: "CNY"}, false},
		{"负数小于正数", Money{Amount: -100, Currency: "CNY"}, Money{Amount: 100, Currency: "CNY"}, true},
		{"零小于正数", Money{Amount: 0, Currency: "CNY"}, Money{Amount: 1, Currency: "CNY"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m1.LessThan(tt.m2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMoney_IsZero 测试金额是否为零
func TestMoney_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		m        Money
		expected bool
	}{
		{"零", Money{Amount: 0, Currency: "CNY"}, true},
		{"正数", Money{Amount: 100, Currency: "CNY"}, false},
		{"负数", Money{Amount: -100, Currency: "CNY"}, false},
		{"一分钱", Money{Amount: 1, Currency: "CNY"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.m.IsZero()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMoney_Immutability 测试Money值对象的不可变性
func TestMoney_Immutability(t *testing.T) {
	m := Money{Amount: 100, Currency: "CNY"}
	other := Money{Amount: 50, Currency: "CNY"}

	// Add操作不应该改变原始值
	result := m.Add(other)
	if m.Amount != 100 {
		t.Error("Add should not modify original Money value")
	}
	if result.Amount != 150 {
		t.Errorf("expected result 150, got %d", result.Amount)
	}

	// Subtract操作不应该改变原始值
	result = m.Subtract(other)
	if m.Amount != 100 {
		t.Error("Subtract should not modify original Money value")
	}
	if result.Amount != 50 {
		t.Errorf("expected result 50, got %d", result.Amount)
	}

	// Multiply操作不应该改变原始值
	result = m.Multiply(2)
	if m.Amount != 100 {
		t.Error("Multiply should not modify original Money value")
	}
	if result.Amount != 200 {
		t.Errorf("expected result 200, got %d", result.Amount)
	}
}

// TestOrder_AddItem 测试订单添加商品
func TestOrder_AddItem(t *testing.T) {
	order := &Order{
		TotalPrice: Money{Amount: 0, Currency: "CNY"},
	}

	item1 := &OrderItem{
		ProductID: 1,
		Quantity:  2,
		Price:     100,
		Currency:  "CNY",
	}

	item2 := &OrderItem{
		ProductID: 2,
		Quantity:  1,
		Price:     50,
		Currency:  "CNY",
	}

	order.AddItem(item1)
	if len(order.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(order.Items))
	}
	if order.TotalPrice.Amount != 200 {
		t.Errorf("expected total 200, got %d", order.TotalPrice.Amount)
	}

	order.AddItem(item2)
	if len(order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(order.Items))
	}
	if order.TotalPrice.Amount != 250 {
		t.Errorf("expected total 250, got %d", order.TotalPrice.Amount)
	}
}

// TestOrder_VerifyOwner 测试订单归属校验
func TestOrder_VerifyOwner(t *testing.T) {
	order := &Order{
		UserID: 123,
	}

	// 正确的用户
	err := order.VerifyOwner(123)
	if err != nil {
		t.Errorf("expected no error for correct owner, got: %v", err)
	}

	// 错误的用户
	err = order.VerifyOwner(456)
	if err == nil {
		t.Error("expected error for wrong owner, got nil")
	}
}

// TestOrder_GenerateOrderNo 测试订单号生成
func TestOrder_GenerateOrderNo(t *testing.T) {
	order := &Order{
		UserID: 123,
	}

	order.GenerateOrderNo()
	if order.OrderNo == "" {
		t.Error("expected non-empty order number")
	}

	// 订单号应该以日期开头（长度至少为14位日期 + 用户ID + 8位随机）
	if len(order.OrderNo) < 14 {
		t.Errorf("expected order number length >= 14, got %d", len(order.OrderNo))
	}

	// 生成两个订单号应该不同
	order2 := &Order{UserID: 123}
	order2.GenerateOrderNo()
	if order.OrderNo == order2.OrderNo {
		t.Error("expected different order numbers for different calls")
	}
}
