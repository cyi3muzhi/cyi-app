package biz

import (
	"context"
	"time"

	"cyi-app/app/cin/interface/internal/conf"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
	"github.com/redis/go-redis/v9"
)

// BladesUseCase AI 聊天用例
type BladesUseCase struct {
	agent blades.Agent
	rdb   redis.UniversalClient
}

// NewBladesUseCase 创建 BladesUseCase
func NewBladesUseCase(c *conf.Blades, rdb redis.UniversalClient) *BladesUseCase {
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
	return &BladesUseCase{agent: agent, rdb: rdb}
}

// Chat 发送聊天消息（非流式，等待完整响应）
func (uc *BladesUseCase) Chat(ctx context.Context, message, sessionID string) (string, error) {
	history, err := pkgblades.LoadHistory(ctx, uc.rdb, sessionID)
	if err != nil {
		return "", err
	}
	fullResponse, err := pkgblades.RunAgent(ctx, uc.agent, message, history...)
	if err != nil {
		return "", err
	}
	pkgblades.SaveHistory(ctx, uc.rdb, sessionID, message, fullResponse)
	return fullResponse, nil
}

// ChatStream 流式聊天，通过 chunkCb 逐块返回
func (uc *BladesUseCase) ChatStream(ctx context.Context, message, sessionID string, chunkCb func(chunk string)) (string, error) {
	history, err := pkgblades.LoadHistory(ctx, uc.rdb, sessionID)
	if err != nil {
		return "", err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	invocation := &blades.Invocation{
		Message:    blades.UserMessage(message),
		Streamable: true,
	}
	if len(history) > 0 {
		invocation.History = history
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

	pkgblades.SaveHistory(ctx, uc.rdb, sessionID, message, fullResponse)
	return fullResponse, nil
}
