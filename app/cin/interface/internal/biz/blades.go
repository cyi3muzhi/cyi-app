package biz

import (
	"context"
	"time"

	"cyi-app/app/cin/interface/internal/conf"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
)

// BladesUseCase AI 聊天用例
type BladesUseCase struct {
	agent blades.Agent
}

// NewBladesUseCase 创建 BladesUseCase，从 conf.Blades 读取配置
// 如果 c 为 nil，则使用默认配置和 DEEPSEEK_API_KEY 环境变量
func NewBladesUseCase(c *conf.Blades) *BladesUseCase {
	cfg := &pkgblades.AgentConfig{}
	if c != nil {
		cfg.APIKey = c.GetApiKey()
		cfg.APIEndpoint = c.GetApiEndpoint()
		cfg.Model = c.GetModel()
		cfg.SystemPrompt = c.GetSystemPrompt()
		cfg.MaxTokens = c.GetMaxTokens()
		cfg.Temperature = c.GetTemperature()
	}
	agent, err := pkgblades.CreateAgent(cfg)
	if err != nil {
		panic("failed to create blades agent: " + err.Error())
	}
	return &BladesUseCase{agent: agent}
}

// Chat 发送聊天消息（非流式，等待完整响应）
func (uc *BladesUseCase) Chat(ctx context.Context, message string) (string, error) {
	return pkgblades.RunAgent(ctx, uc.agent, message)
}

// ChatStream 流式聊天，通过 chunkCb 逐块返回
func (uc *BladesUseCase) ChatStream(ctx context.Context, message string, chunkCb func(chunk string)) (string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	invocation := &blades.Invocation{
		Message:    blades.UserMessage(message),
		Streamable: true,
	}

	stream := uc.agent.Run(ctxWithTimeout, invocation)

	var fullResponse string
	for msg, err := range stream {
		if err != nil {
			return fullResponse, err
		}
		if msg != nil {
			chunk := msg.Text()
			fullResponse += chunk
			chunkCb(chunk)
		}
	}

	return fullResponse, nil
}
