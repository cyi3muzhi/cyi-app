package blades

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/tools"
)

type DeepSeekModel struct {
	apiKey      string
	apiEndpoint string
	model       string
}

func NewDeepSeekModel(apiKey string, apiEndpoint string, model string) *DeepSeekModel {
	if apiEndpoint == "" {
		apiEndpoint = "https://api.deepseek.com/v1"
	}
	if model == "" {
		model = "deepseek-v4-flash"
	}
	return &DeepSeekModel{
		apiKey:      apiKey,
		apiEndpoint: apiEndpoint,
		model:       model,
	}
}

func (m *DeepSeekModel) Name() string {
	return m.model
}

// deepseekTools 转为 DeepSeek API 工具格式
func deepseekTools(reqTools []tools.Tool) []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range reqTools {
		schema := t.InputSchema()
		result = append(result, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Name(),
				"description": t.Description(),
				"parameters":  schema,
			},
		})
	}
	return result
}

// convertToolCalls 将 DeepSeek 返回的 tool_calls 转为 API 请求格式
func convertToolCalls(toolCalls []deepSeekToolCall) []map[string]interface{} {
	var result []map[string]interface{}
	for _, tc := range toolCalls {
		tcType := tc.Type
		if tcType == "" {
			tcType = "function"
		}
		result = append(result, map[string]interface{}{
			"id":   tc.ID,
			"type": tcType,
			"function": map[string]interface{}{
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
			},
		})
	}
	return result
}

// deepSeekToolCall DeepSeek tool_calls 结构（解析用）
// Index SSE 流式中的索引号，用于区分多个并行工具调用
type deepSeekToolCall struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"` //nolint:govet
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// deepSeekChoice DeepSeek API 响应中的 choice 结构
type deepSeekChoice struct {
	Message struct {
		Role      string             `json:"role"`
		Content   string             `json:"content"`
		ToolCalls []deepSeekToolCall `json:"tool_calls"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// deepSeekSSEDelta SSE 流中的 delta 结构
type deepSeekSSEDelta struct {
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	ToolCalls []deepSeekToolCall `json:"tool_calls"`
}

// deepSeekSSEEvent SSE 事件结构
type deepSeekSSEEvent struct {
	Choices []struct {
		Delta        deepSeekSSEDelta `json:"delta"`
		FinishReason *string          `json:"finish_reason"`
	} `json:"choices"`
}

// buildMessages 构建发送给 DeepSeek API 的消息列表
func (m *DeepSeekModel) buildMessages(req *blades.ModelRequest) []map[string]interface{} {
	var messages []map[string]interface{}

	// 添加系统指令
	if req.Instruction != nil {
		text := req.Instruction.Text()
		if text != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "system",
				"content": text,
			})
		}
	}

	// 添加历史消息
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": msg.Text(),
		})
	}

	return messages
}

// executeTool 执行单个工具调用
func (m *DeepSeekModel) executeTool(ctx context.Context, toolList []tools.Tool, tc deepSeekToolCall) string {
	for _, t := range toolList {
		if t.Name() == tc.Function.Name {
			result, err := t.Handle(ctx, tc.Function.Arguments)
			if err != nil {
				return fmt.Sprintf("工具执行失败: %v", err)
			}
			return result
		}
	}
	return fmt.Sprintf("未找到工具: %s", tc.Function.Name)
}

// Generate 非流式生成（模型内部完成工具调用轮询）
func (m *DeepSeekModel) Generate(ctx context.Context, req *blades.ModelRequest) (*blades.ModelResponse, error) {
	apiURL := m.apiEndpoint + "/chat/completions"
	client := &http.Client{Timeout: 120 * time.Second}

	// 构建初始消息列表
	messages := m.buildMessages(req)

	for round := 0; round < 3; round++ {
		// 构建请求体
		requestBody := map[string]interface{}{
			"model":       m.model,
			"messages":    messages,
			"temperature": 0.7,
			"max_tokens":  4096,
		}
		if len(req.Tools) > 0 {
			requestBody["tools"] = deepseekTools(req.Tools)
		}

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

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("deepseek api error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
		}

		var apiResp struct {
			Choices []deepSeekChoice `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if len(apiResp.Choices) == 0 {
			return nil, fmt.Errorf("no response from DeepSeek API")
		}

		choice := apiResp.Choices[0]

		// 没有工具调用 → 返回最终文本响应
		if len(choice.Message.ToolCalls) == 0 {
			return &blades.ModelResponse{
				Message: &blades.Message{
					Role:   blades.RoleAssistant,
					Parts:  blades.NewMessageParts(choice.Message.Content),
					Status: blades.StatusCompleted,
				},
			}, nil
		}

		// 有工具调用 → 添加 assistant 消息（含 tool_calls）
		assistantMsg := map[string]interface{}{
			"role":       "assistant",
			"content":    choice.Message.Content,
			"tool_calls": convertToolCalls(choice.Message.ToolCalls),
		}
		messages = append(messages, assistantMsg)

		// 执行工具并添加 tool 消息
		for _, tc := range choice.Message.ToolCalls {
			result := m.executeTool(ctx, req.Tools, tc)
			messages = append(messages, map[string]interface{}{
				"role":         "tool",
				"content":      result,
				"tool_call_id": tc.ID,
			})
		}
	}

	return nil, fmt.Errorf("max tool call rounds exceeded")
}

// NewStreaming 流式生成（模型内部完成工具调用轮询）
func (m *DeepSeekModel) NewStreaming(ctx context.Context, req *blades.ModelRequest) blades.Generator[*blades.ModelResponse, error] {
	return func(yield func(*blades.ModelResponse, error) bool) {
		apiURL := m.apiEndpoint + "/chat/completions"
		client := &http.Client{Timeout: 120 * time.Second}

		// 构建初始消息列表
		messages := m.buildMessages(req)

		for round := 0; round < 3; round++ {
			// 构建请求体
			requestBody := map[string]interface{}{
				"model":       m.model,
				"messages":    messages,
				"temperature": 0.7,
				"max_tokens":  4096,
				"stream":      true,
			}
			if len(req.Tools) > 0 {
				requestBody["tools"] = deepseekTools(req.Tools)
			}

			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				yield(nil, fmt.Errorf("failed to marshal request: %w", err))
				return
			}

			httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				yield(nil, fmt.Errorf("failed to create request: %w", err))
				return
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
			httpReq.Header.Set("Accept", "text/event-stream")

			resp, err := client.Do(httpReq)
			if err != nil {
				yield(nil, fmt.Errorf("failed to send request: %w", err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				yield(nil, fmt.Errorf("deepseek api error: status=%d, body=%s", resp.StatusCode, string(bodyBytes)))
				return
			}

			// 流式读取 SSE 事件
			type toolCallAcc struct {
				ID      string
				Name    string
				ArgsBuf string
			}
			accumulated := make(map[int]*toolCallAcc)
			var fullContent string
			hasToolCalls := false

			scanner := NewSSEScanner(resp.Body)
			for scanner.Scan() {
				data := scanner.Data()
				if data == nil || string(data) == "[DONE]" {
					continue
				}

				var sseEvent deepSeekSSEEvent
				if err := json.Unmarshal(data, &sseEvent); err != nil {
					continue
				}
				if len(sseEvent.Choices) == 0 {
					continue
				}

				delta := sseEvent.Choices[0].Delta
				finishReason := sseEvent.Choices[0].FinishReason

				// 累积工具调用
				if len(delta.ToolCalls) > 0 {
					hasToolCalls = true
					for _, tc := range delta.ToolCalls {
						acc, ok := accumulated[tc.Index]
						if !ok {
							acc = &toolCallAcc{}
							accumulated[tc.Index] = acc
						}
						if tc.ID != "" {
							acc.ID = tc.ID
						}
						if tc.Function.Name != "" {
							acc.Name = tc.Function.Name
						}
						acc.ArgsBuf += tc.Function.Arguments
					}
					continue
				}

				// 处理文本内容
				content := delta.Content
				fullContent += content
				status := blades.StatusInProgress
				if finishReason != nil {
					status = blades.StatusCompleted
				}

				if content == "" && status != blades.StatusCompleted {
					continue
				}

				// 文本流式输出
				if !yield(&blades.ModelResponse{
					Message: &blades.Message{
						Role:   blades.RoleAssistant,
						Parts:  blades.NewMessageParts(content),
						Status: status,
					},
				}, nil) {
					resp.Body.Close()
					return
				}
			}
			resp.Body.Close()

			if err := scanner.Err(); err != nil {
				yield(nil, fmt.Errorf("sse read error: %w", err))
				return
			}

			// 没有工具调用 → 流结束
			if !hasToolCalls {
				return
			}

			// 有工具调用 → 添加 assistant 消息（含 tool_calls）
			var toolCalls []map[string]interface{}
			for _, acc := range accumulated {
				toolCalls = append(toolCalls, map[string]interface{}{
					"id":   acc.ID,
					"type": "function",
					"function": map[string]interface{}{
						"name":      acc.Name,
						"arguments": acc.ArgsBuf,
					},
				})
			}
			assistantMsg := map[string]interface{}{
				"role":       "assistant",
				"content":    fullContent,
				"tool_calls": toolCalls,
			}
			messages = append(messages, assistantMsg)

			// 执行工具并添加 tool 消息
			for _, acc := range accumulated {
				tc := deepSeekToolCall{
					ID:   acc.ID,
					Type: "function",
				}
				tc.Function.Name = acc.Name
				tc.Function.Arguments = acc.ArgsBuf
				result := m.executeTool(ctx, req.Tools, tc)
				messages = append(messages, map[string]interface{}{
					"role":         "tool",
					"content":      result,
					"tool_call_id": acc.ID,
				})
			}
		}

		yield(nil, fmt.Errorf("max tool call rounds exceeded"))
	}
}

func evaluateExpression(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)
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
	return strconv.ParseFloat(expr, 64)
}

type CalculatorInput struct {
	Expression string `json:"expression"`
}

type CalculatorOutput struct {
	Result string `json:"result"`
}

type AgentConfig struct {
	APIKey       string
	APIEndpoint  string
	Model        string
	SystemPrompt string
	MaxTokens    int32
	Temperature  float64
}

func CreateAgent(cfg *AgentConfig) (blades.Agent, error) {
	if cfg == nil {
		cfg = &AgentConfig{}
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("DEEPSEEK_API_KEY")
	}
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-v4-flash"
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = "You are a helpful assistant. Use the calculator tool for math questions."
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Temperature <= 0 {
		cfg.Temperature = 0.7
	}
	model := NewDeepSeekModel(cfg.APIKey, cfg.APIEndpoint, cfg.Model)
	calculatorTool, err := tools.NewFunc(
		"calculator",
		"A calculator tool that evaluates mathematical expressions",
		func(ctx context.Context, input CalculatorInput) (CalculatorOutput, error) {
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
	agent, err := blades.NewAgent(
		cfg.Model,
		blades.WithModel(model),
		blades.WithTools(calculatorTool),
		blades.WithInstruction(cfg.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	return agent, nil
}

func RunAgent(ctx context.Context, agent blades.Agent, prompt string, history ...*blades.Message) (string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	invocation := &blades.Invocation{
		Message: blades.UserMessage(prompt),
	}
	if len(history) > 0 {
		invocation.History = history
	}
	stream := agent.Run(ctxWithTimeout, invocation)
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
