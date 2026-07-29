package biz

import (
	"context"
	"fmt"
	"sync"

	"cyi-app/app/blades/internal/conf"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/tools"
	"github.com/go-kratos/kratos/v2/log"
)

// AgentInstance 一个 Agent 实例（对应一个端）
type AgentInstance struct {
	source       string
	agent        blades.Agent
	rag          *RAGEngine
	prompt       string
	toolList     []tools.Tool
	knowledgeTag string
}

// Chat 发送消息（非流式）
func (inst *AgentInstance) Chat(ctx context.Context, message string, enableRAG bool) (string, error) {
	if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
		contextStr, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
		if err == nil && contextStr != "" {
			message = contextStr + "\n\n用户问题：" + message
		}
	}
	invocation := &blades.Invocation{
		Message: blades.UserMessage(message),
	}
	stream := inst.agent.Run(ctx, invocation)
	var response string
	for msg, err := range stream {
		if err != nil {
			return response, err
		}
		if msg != nil {
			response += msg.Text()
		}
	}
	return response, nil
}

// ChatStream 流式聊天
func (inst *AgentInstance) ChatStream(ctx context.Context, message string, enableRAG bool, chunkCb func(string, bool)) error {
	if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
		contextStr, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
		if err == nil && contextStr != "" {
			message = contextStr + "\n\n用户问题：" + message
		}
	}
	invocation := &blades.Invocation{
		Message:    blades.UserMessage(message),
		Streamable: true,
	}
	stream := inst.agent.Run(ctx, invocation)
	var fullResponse string
	for msg, err := range stream {
		if err != nil {
			chunkCb(fullResponse, true)
			return err
		}
		if msg != nil {
			chunk := msg.Text()
			fullResponse += chunk
			chunkCb(chunk, false)
		}
	}
	chunkCb("", true)
	return nil
}

// AgentFactory Agent 工厂，按 source 管理不同端的 Agent
type AgentFactory struct {
	mu     sync.RWMutex
	cache  map[string]*AgentInstance
	cfg    *conf.Blades
	rag    *RAGEngine
	reg    *ToolRegistry
	logger log.Logger
}

// NewAgentFactory 创建 Agent 工厂
func NewAgentFactory(cfg *conf.Blades, rag *RAGEngine, reg *ToolRegistry, logger log.Logger) *AgentFactory {
	return &AgentFactory{
		cache:  make(map[string]*AgentInstance),
		cfg:    cfg,
		rag:    rag,
		reg:    reg,
		logger: logger,
	}
}

// GetOrCreate 获取或创建 Agent 实例
func (f *AgentFactory) GetOrCreate(ctx context.Context, source string) (*AgentInstance, error) {
	f.mu.RLock()
	inst, ok := f.cache[source]
	f.mu.RUnlock()
	if ok {
		return inst, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if inst, ok := f.cache[source]; ok {
		return inst, nil
	}

	inst, err := f.createAgent(ctx, source)
	if err != nil {
		return nil, err
	}
	f.cache[source] = inst
	return inst, nil
}

func (f *AgentFactory) createAgent(ctx context.Context, source string) (*AgentInstance, error) {
	prompt, knowledgeTag := f.getSourceConfig(source)

	opts := []blades.AgentOption{
		blades.WithInstruction(prompt),
		blades.WithModel(pkgblades.NewDeepSeekModel(
			f.cfg.ApiKey,
			"",
			f.cfg.Model,
		)),
	}

	// 从注册中心获取工具列表
	var agentTools []tools.Tool
	for _, item := range f.reg.GetTools() {
		if t, ok := item.(tools.Tool); ok {
			agentTools = append(agentTools, t)
		}
	}
	if len(agentTools) > 0 {
		opts = append(opts, blades.WithTools(agentTools...))
	}

	agent, err := blades.NewAgent(f.cfg.Model, opts...)
	if err != nil {
		return nil, fmt.Errorf("create agent for %s failed: %w", source, err)
	}

	return &AgentInstance{
		source:       source,
		agent:        agent,
		rag:          f.rag,
		prompt:       prompt,
		toolList:     agentTools,
		knowledgeTag: knowledgeTag,
	}, nil
}

func (f *AgentFactory) getToolNamesForSource(source string) []string {
	switch source {
	case "admin":
		return []string{"query_orders", "query_order_detail", "search_products", "get_product_detail", "query_users", "search_knowledge"}
	case "user":
		return []string{"search_products", "get_product_detail", "search_knowledge"}
	default:
		return []string{"search_knowledge"}
	}
}

func (f *AgentFactory) getSourceConfig(source string) (prompt, knowledgeTag string) {
	switch source {
	case "admin":
		return `你是商城管理助手。你可以使用工具查询订单、商品、用户等数据。

当用户询问数据（如"有多少商品"、"昨天订单多少"、"总用户数"）时，你必须使用工具来获取实时数据，不要编造答案。

规则：
1. 用户问数据 → 立即使用对应的工具查询
2. 工具返回结果后，用简洁的语言告诉用户
3. 如果工具查询失败，请告知用户暂时无法查询
4. 请用友好的语气简洁回答问题`, "admin"
	case "user":
		return `你是商城购物助手，帮助用户解答购物问题。
你可以帮用户搜索商品、查询商品详情、查找平台规则和指南。
请用热情、友好的语气回答。对于不确定的信息，请引导用户联系客服。`, "user"
	default:
		return `你是商城通用助手。请用友好、简洁的方式回答用户的问题。`, "public"
	}
}
