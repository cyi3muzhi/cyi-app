package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	userRpc "cyi-app/api/service/user/v1"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/tools"
	"github.com/go-kratos/kratos/v2/log"
)

// toolPerm 工具所需的模块权限
type toolPerm struct {
	module string
	action string
}

// toolModulePerms 工具名 → 所需最小模块权限（模块 key 与 user 服务 permModules 一致）
var toolModulePerms = map[string]toolPerm{
	"search_products":    {"products", "view"},
	"get_product_detail": {"products", "view"},
	"query_stock":        {"products", "view"},
	"query_orders":       {"orders", "view"},
	"query_order_detail": {"orders", "view"},
	"query_users":        {"users", "view"},
}

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
func (inst *AgentInstance) Chat(ctx context.Context, message string, enableRAG bool, history []*blades.Message, instruction *blades.Message) (string, error) {
	if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
		contextStr, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
		if err == nil && contextStr != "" {
			message = contextStr + "\n\n用户问题：" + message
		}
	}
	invocation := &blades.Invocation{
		Message:     blades.UserMessage(message),
		History:     history,
		Instruction: instruction,
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
func (inst *AgentInstance) ChatStream(ctx context.Context, message string, enableRAG bool, history []*blades.Message, instruction *blades.Message, chunkCb func(string, bool)) error {
	if enableRAG && inst.rag != nil && inst.rag.ShouldRetrieve(message) {
		contextStr, err := inst.rag.Retrieve(ctx, inst.knowledgeTag, message, 0)
		if err == nil && contextStr != "" {
			message = contextStr + "\n\n用户问题：" + message
		}
	}
	invocation := &blades.Invocation{
		Message:     blades.UserMessage(message),
		Streamable:  true,
		History:     history,
		Instruction: instruction,
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

// cacheEntry 带 TTL 的缓存条目
type cacheEntry struct {
	inst      *AgentInstance
	createdAt time.Time
}

const agentCacheTTL = 15 * time.Minute

// AgentFactory Agent 工厂，按 source+role 管理不同端的 Agent
type AgentFactory struct {
	mu         sync.RWMutex
	cache      map[string]*cacheEntry
	cfg        *BladesConfig
	rag        *RAGEngine
	reg        *ToolRegistry
	roleClient userRpc.AdminRoleClient
	logger     log.Logger
}

// NewAgentFactory 创建 Agent 工厂
func NewAgentFactory(cfg *BladesConfig, rag *RAGEngine, reg *ToolRegistry, rc userRpc.AdminRoleClient, logger log.Logger) *AgentFactory {
	return &AgentFactory{
		cache:      make(map[string]*cacheEntry),
		cfg:        cfg,
		rag:        rag,
		reg:        reg,
		roleClient: rc,
		logger:     logger,
	}
}

// GetOrCreate 获取或创建 Agent 实例（按 source+role 缓存，3分钟 TTL）
func (f *AgentFactory) GetOrCreate(ctx context.Context, source, role string) (*AgentInstance, error) {
	cacheKey := source + ":" + role
	f.mu.RLock()
	entry, ok := f.cache[cacheKey]
	f.mu.RUnlock()
	if ok && time.Since(entry.createdAt) < agentCacheTTL {
		return entry.inst, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// 双检：可能刚被其他协程更新
	if entry, ok := f.cache[cacheKey]; ok && time.Since(entry.createdAt) < agentCacheTTL {
		return entry.inst, nil
	}

	inst, err := f.createAgent(ctx, source, role)
	if err != nil {
		return nil, err
	}
	f.cache[cacheKey] = &cacheEntry{inst: inst, createdAt: time.Now()}
	return inst, nil
}

func (f *AgentFactory) createAgent(ctx context.Context, source, role string) (*AgentInstance, error) {
	prompt, knowledgeTag := f.getSourceConfig(source)

	opts := []blades.AgentOption{
		blades.WithInstruction(prompt),
		blades.WithModel(pkgblades.NewDeepSeekModel(
			f.cfg.ApiKey,
			"",
			f.cfg.Model,
		)),
	}

	// 从注册中心获取工具列表，按角色过滤
	var agentTools []tools.Tool
	allowedTools := f.getAllowedTools(ctx, source, role)
	allowed := make(map[string]bool, len(allowedTools))
	for _, name := range allowedTools {
		allowed[name] = true
	}
	for _, t := range f.reg.GetTools() {
		if allowed[t.Name()] {
			agentTools = append(agentTools, t)
		}
	}
	if len(agentTools) > 0 {
		opts = append(opts, blades.WithTools(agentTools...))
	}

	// 添加中间件
	opts = append(opts, blades.WithMiddleware(
		AgentRecovery(f.logger),
		AgentLogging(f.logger),
	))

	agent, err := blades.NewAgent(f.cfg.Model, opts...)
	if err != nil {
		return nil, fmt.Errorf("create agent for %s:%s failed: %w", source, role, err)
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

// getAllowedTools 调 user 服务查角色的模块权限，返回该角色可用的工具列表
func (f *AgentFactory) getAllowedTools(ctx context.Context, source, role string) []string {
	// 通用工具（所有角色都有）
	common := []string{"search_knowledge", "save_user_preference", "query_user_permissions", "check_user_permission"}

	if source != "admin" {
		return common
	}

	if f.roleClient == nil {
		return common
	}

	// 查角色的模块权限
	resp, err := f.roleClient.GetRolePermissions(ctx, &userRpc.GetRolePermissionsRequest{Role: role})
	if err != nil || resp == nil {
		return common
	}

	// 构建模块→操作映射
	modPerms := make(map[string]map[string]bool)
	for mod, pv := range resp.Permissions {
		if pv != nil {
			modPerms[mod] = pv.Values
		}
	}

	// 检查每个工具有无对应模块权限
	var tools_ []string
	for name, perm := range toolModulePerms {
		if actions, ok := modPerms[perm.module]; ok {
			if actions[perm.action] {
				tools_ = append(tools_, name)
			}
		}
	}
	return append(common, tools_...)
}

func (f *AgentFactory) getSourceConfig(source string) (prompt, knowledgeTag string) {
	dateStr := time.Now().Format("2006年1月2日")
	switch source {
	case "admin":
		return fmt.Sprintf(`当前日期：%s

你是商城管理助手。你可以使用工具查询订单、商品、用户等数据。

当用户询问数据（如"有多少商品"、"昨天订单多少"、"总用户数"）时，你必须使用工具来获取实时数据，不要编造答案。

用户可能会告诉你他们的偏好（如昵称、称呼方式等），当用户设置偏好时，请使用 save_user_preference 工具保存。

当用户询问权限相关问题（如"我有什么权限"、"我能做什么"、"我可以创建商品吗"）时，使用 query_user_permissions 或 check_user_permission 工具查询。

当用户询问商品库存（如"还有多少库存"、"还剩多少"、"有没有货"）时：
1. 先使用 get_product_detail 工具查询商品详细信息，获取 SKU 列表（每个 SKU 都有唯一的 SKU ID）
2. 从 SKU 列表中找出目标规格对应的 SKU ID
3. 使用 query_stock 工具传入 SKU ID 查询实时库存
4. 将结果告知用户

规则：
1. 用户问数据 → 立即使用对应的工具查询
2. 工具返回结果后，用简洁的语言告诉用户
3. 如果工具查询失败，请告知用户暂时无法查询
4. 当用户告诉你关于他们自己的信息（如昵称、称呼）时，使用 save_user_preference 工具保存
5. 用户问权限 → 使用权限查询工具
6. 请用友好的语气简洁回答问题%s`, dateStr, formatRules), "admin"
	case "user":
		return fmt.Sprintf(`当前日期：%s

你是商城购物助手，帮助用户解答购物问题。
你可以帮用户搜索商品、查询商品详情、查找平台规则和指南。
请用热情、友好的语气回答。对于不确定的信息，请引导用户联系客服。%s`, dateStr, formatRules), "user"
	default:
		return fmt.Sprintf(`当前日期：%s

你是商城通用助手。请用友好、简洁的方式回答用户的问题。%s`, dateStr, formatRules), "public"
	}
}

// formatRules 统一的输出格式规范，拼接在各场景提示词之后，保证 AI 回复结构清晰、易于阅读
const formatRules = `
输出格式要求（必须遵守）：
1. 使用 Markdown 结构化输出：重点小节用 ## 标题，要点用 - 或 1. 列表，强调用 **加粗**
2. 查询类结果（订单、商品、用户等数据）优先用 Markdown 表格展示。表格语法：每行以 | 开头和结尾，列之间用 | 分隔，表头下方必须有一行 |---|---| 分隔行，表头用中文，列数不超过 5 列。严禁使用 Tab 制表符或空格分隔的文本表格
3. 先直接给出结论，再补充细节；不要客套寒暄，结尾不要写"还有什么可以帮您"之类的套话
4. 内容分点表述，每点尽量一句话；避免一大段文字堆砌
5. 整个回答不要用代码块包裹，除非是给用户的命令示例`
