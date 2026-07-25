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

func (m *DeepSeekModel) Generate(ctx context.Context, req *blades.ModelRequest) (*blades.ModelResponse, error) {
	apiURL := m.apiEndpoint + "/chat/completions"
	var messages []map[string]interface{}
	for _, msg := range req.Messages {
		role := string(msg.Role)
		content := msg.Text()
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}
	requestBody := map[string]interface{}{
		"model":       m.model,
		"messages":    messages,
		"temperature": 0.7,
		"max_tokens":  1000,
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
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deepseek api error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

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
	return &blades.ModelResponse{
		Message: &blades.Message{
			Role:   blades.RoleAssistant,
			Parts:  blades.NewMessageParts(apiResp.Choices[0].Message.Content),
			Status: blades.StatusCompleted,
		},
	}, nil
}

func (m *DeepSeekModel) NewStreaming(ctx context.Context, req *blades.ModelRequest) blades.Generator[*blades.ModelResponse, error] {
	return func(yield func(*blades.ModelResponse, error) bool) {
		apiURL := m.apiEndpoint + "/chat/completions"
		var messages []map[string]interface{}
		for _, msg := range req.Messages {
			messages = append(messages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Text(),
			})
		}
		requestBody := map[string]interface{}{
			"model":       m.model,
			"messages":    messages,
			"temperature": 0.7,
			"max_tokens":  4096,
			"stream":      true,
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
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			yield(nil, fmt.Errorf("failed to send request: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			yield(nil, fmt.Errorf("deepseek api error: status=%d, body=%s", resp.StatusCode, string(bodyBytes)))
			return
		}

		scanner := NewSSEScanner(resp.Body)
		for scanner.Scan() {
			data := scanner.Data()
			if data == nil || string(data) == "[DONE]" {
				continue
			}
			var sseEvent struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal(data, &sseEvent); err != nil {
				continue
			}
			if len(sseEvent.Choices) == 0 {
				continue
			}
			content := sseEvent.Choices[0].Delta.Content
			status := blades.StatusInProgress
			if sseEvent.Choices[0].FinishReason != nil {
				status = blades.StatusCompleted
			}
			if content == "" && status != blades.StatusCompleted {
				continue
			}
			if !yield(&blades.ModelResponse{
				Message: &blades.Message{
					Role:   blades.RoleAssistant,
					Parts:  blades.NewMessageParts(content),
					Status: status,
				},
			}, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(nil, fmt.Errorf("sse read error: %w", err))
		}
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
