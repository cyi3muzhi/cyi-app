package blades

import (
	"testing"
)

// TestEvaluateExpression_Addition 测试加法运算
func TestEvaluateExpression_Addition(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"简单加法", "1+2", 3},
		{"多个数相加", "1+2+3", 6},
		{"带空格", " 1 + 2 + 3 ", 6},
		{"小数相加", "1.5+2.5", 4.0},
		{"零相加", "0+0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_Subtraction 测试减法运算
func TestEvaluateExpression_Subtraction(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"简单减法", "5-3", 2},
		{"多个数相减", "10-3-2", 5},
		{"带空格", " 10 - 3 - 2 ", 5},
		{"小数相减", "5.5-2.5", 3.0},
		{"结果为负", "3-5", -2},
		{"零减", "0-5", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_Multiplication 测试乘法运算
func TestEvaluateExpression_Multiplication(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"简单乘法", "3*4", 12},
		{"多个数相乘", "2*3*4", 24},
		{"带空格", " 2 * 3 * 4 ", 24},
		{"小数相乘", "2.5*4", 10.0},
		{"乘以零", "5*0", 0},
		{"乘以一", "5*1", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_Division 测试除法运算
func TestEvaluateExpression_Division(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"简单除法", "10/2", 5},
		{"多个数相除", "24/3/2", 4},
		{"带空格", " 24 / 3 / 2 ", 4},
		{"小数相除", "5.5/2.5", 2.2},
		{"零除以数", "0/5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_DivisionByZero 测试除零错误
func TestEvaluateExpression_DivisionByZero(t *testing.T) {
	_, err := evaluateExpression("10/0")
	if err == nil {
		t.Error("expected error for division by zero, got nil")
	}
	if err.Error() != "division by zero" {
		t.Errorf("expected 'division by zero' error, got: %v", err)
	}
}

// TestEvaluateExpression_SingleNumber 测试单个数字
func TestEvaluateExpression_SingleNumber(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"整数", "42", 42},
		{"小数", "3.14", 3.14},
		{"零", "0", 0},
		{"带空格", "  123  ", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluateExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_InvalidInput 测试无效输入
func TestEvaluateExpression_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"空字符串", ""},
		{"非数字", "abc"},
		{"混合无效", "1+abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluateExpression(tt.expr)
			if err == nil {
				t.Error("expected error for invalid input, got nil")
			}
		})
	}
}

// TestNewDeepSeekModel 测试DeepSeek模型创建
func TestNewDeepSeekModel(t *testing.T) {
	// 测试指定endpoint和model
	model := NewDeepSeekModel("test-key", "https://custom.api.com/v1", "deepseek-v4-flash")
	if model == nil {
		t.Fatal("expected non-nil model")
	}
	if model.Name() != "deepseek-v4-flash" {
		t.Errorf("expected model name 'deepseek-v4-flash', got '%s'", model.Name())
	}

	// 测试默认endpoint和model
	model2 := NewDeepSeekModel("test-key", "", "")
	if model2 == nil {
		t.Fatal("expected non-nil model")
	}
	// 验证默认endpoint被设置
	if model2.apiEndpoint != "https://api.deepseek.com/v1" {
		t.Errorf("expected default endpoint 'https://api.deepseek.com/v1', got '%s'", model2.apiEndpoint)
	}
	// 验证默认model被设置
	if model2.Name() != "deepseek-v4-flash" {
		t.Errorf("expected default model 'deepseek-v4-flash', got '%s'", model2.Name())
	}
}

// TestCreateAgent_NilConfig 测试使用nil配置创建Agent
func TestCreateAgent_NilConfig(t *testing.T) {
	// 使用nil配置应该不会panic，会使用默认值
	// 注意：实际创建会因为API key为空但使用环境变量，这里只测试不panic
	_, err := CreateAgent(nil)
	// 可能会因为blades.NewAgent内部失败而返回错误，但不应该panic
	// 这里我们只验证函数能正常执行
	t.Logf("CreateAgent with nil config returned err: %v", err)
}
