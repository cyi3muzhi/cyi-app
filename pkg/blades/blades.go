package blades

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/tools"
)

// DeepSeekModel 实现DeepSeek模型
type DeepSeekModel struct {
	apiKey string
}

// NewDeepSeekModel 创建DeepSeek模型
func NewDeepSeekModel(apiKey string) *DeepSeekModel {
	return &DeepSeekModel{
		apiKey: apiKey,
	}
}

// Name 返回模型名称
func (m *DeepSeekModel) Name() string {
	return "deepseek-chat"
}

// Generate 生成响应（实际调用DeepSeek API）
func (m *DeepSeekModel) Generate(ctx context.Context, req *blades.ModelRequest) (*blades.ModelResponse, error) {
	// 构建DeepSeek API请求
	apiURL := "https://api.deepseek.com/v1/chat/completions"

	// 构建消息
	var messages []map[string]interface{}

	// 添加历史消息
	for _, msg := range req.Messages {
		role := string(msg.Role)
		content := msg.Text()
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":       "deepseek-chat",
		"messages":    messages,
		"temperature": 0.7,
		"max_tokens":  1000,
	}

	// 发送HTTP请求
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	client := &http.Client{
		Timeout: 30 * time.Second, // 设置30秒超时
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from DeepSeek API")
	}

	// 构建Blades响应
	return &blades.ModelResponse{
		Message: &blades.Message{
			Role:   blades.RoleAssistant,
			Parts:  blades.NewMessageParts(apiResp.Choices[0].Message.Content),
			Status: blades.StatusCompleted,
		},
	}, nil
}

// NewStreaming 创建流式响应
func (m *DeepSeekModel) NewStreaming(ctx context.Context, req *blades.ModelRequest) blades.Generator[*blades.ModelResponse, error] {
	return func(yield func(*blades.ModelResponse, error) bool) {
		// 这里可以实现流式API调用
		// 现在返回模拟的流式响应
		yield(&blades.ModelResponse{
			Message: &blades.Message{
				Role:   blades.RoleAssistant,
				Parts:  blades.NewMessageParts("This is a streaming response from DeepSeek"),
				Status: blades.StatusCompleted,
			},
		}, nil)
	}
}

// evaluateExpression 计算数学表达式
func evaluateExpression(expr string) (float64, error) {
	// 简单的表达式计算实现
	expr = strings.TrimSpace(expr)

	// 处理简单的加减乘除
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		var sum float64
		for _, part := range parts {
			num, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				return 0, err
			}
			sum += num
		}
		return sum, nil
	} else if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		num1, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, err
		}
		for i := 1; i < len(parts); i++ {
			num, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
			if err != nil {
				return 0, err
			}
			num1 -= num
		}
		return num1, nil
	} else if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		product := 1.0
		for _, part := range parts {
			num, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				return 0, err
			}
			product *= num
		}
		return product, nil
	} else if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		num1, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, err
		}
		for i := 1; i < len(parts); i++ {
			num, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
			if err != nil {
				return 0, err
			}
			if num == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			num1 /= num
		}
		return num1, nil
	}

	// 纯数字
	return strconv.ParseFloat(expr, 64)
}

// CalculatorInput 计算器输入
type CalculatorInput struct {
	Expression string `json:"expression"`
}

// CalculatorOutput 计算器输出
type CalculatorOutput struct {
	Result string `json:"result"`
}

// CreateAgent 创建Blades Agent
func CreateAgent(apiKey string) (blades.Agent, error) {
	// 创建DeepSeek模型
	model := NewDeepSeekModel(apiKey)

	// 创建计算器工具
	calculatorTool, err := tools.NewFunc(
		"calculator",
		"A calculator tool that evaluates mathematical expressions",
		func(ctx context.Context, input CalculatorInput) (CalculatorOutput, error) {
			// 实际计算逻辑
			result, err := evaluateExpression(input.Expression)
			if err != nil {
				return CalculatorOutput{}, fmt.Errorf("failed to evaluate expression: %w", err)
			}
			return CalculatorOutput{
				Result: fmt.Sprintf("The result of %s is %v", input.Expression, result),
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create calculator tool: %w", err)
	}

	// 创建Agent
	agent, err := blades.NewAgent(
		"deepseek-agent",
		blades.WithModel(model),
		blades.WithTools(calculatorTool),
		blades.WithInstruction("You are a helpful assistant. Use the calculator tool for math questions."),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

// RunAgent 运行Agent
func RunAgent(ctx context.Context, agent blades.Agent, prompt string) (string, error) {
	// 创建带有超时的上下文
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// 创建调用
	invocation := &blades.Invocation{
		Message: blades.UserMessage(prompt),
	}

	// 运行Agent
	stream := agent.Run(ctxWithTimeout, invocation)

	// 收集结果
	var response string
	for msg, err := range stream {
		if err != nil {
			return "", err
		}
		if msg != nil {
			response += msg.Text()
		}
	}

	return response, nil
}
