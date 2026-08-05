package biz

import (
	"context"
	"fmt"
	"strings"
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
	"search_products":       {"products", "view"},
	"get_product_detail":    {"products", "view"},
	"query_stock":           {"products", "view"},
	"query_orders":          {"orders", "view"},
	"query_order_detail":    {"orders", "view"},
	"query_order_stats":     {"orders", "view"},
	"query_order_trend":     {"orders", "view"},
	"query_dashboard_stats": {"orders", "view"},
	"query_users":           {"users", "view"},
	"query_user_stats":      {"users", "view"},
	"query_user_detail":     {"users", "view"},
	"query_coupons":         {"coupons", "view"},
	"query_coupon_stats":    {"coupons", "view"},
	"query_coupon_records":  {"coupons", "view"},
	"query_categories":      {"categories", "view"},
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

// Chat 发送消息（非流式）。RAG 上下文由调用方通过 instruction 传入，不再在此处拼接。
func (inst *AgentInstance) Chat(ctx context.Context, message string, instruction *blades.Message, history []*blades.Message) (string, error) {
	invocation := &blades.Invocation{
		Message:     blades.UserMessage(message),
		History:     history,
		Instruction: instruction,
	}
	stream := inst.agent.Run(ctx, invocation)
	var sb strings.Builder
	for msg, err := range stream {
		if err != nil {
			return sb.String(), err
		}
		if msg != nil {
			sb.WriteString(msg.Text())
		}
	}
	return sb.String(), nil
}

// ChatStream 流式聊天。RAG 上下文由调用方通过 instruction 传入，不再在此处拼接。
func (inst *AgentInstance) ChatStream(ctx context.Context, message string, instruction *blades.Message, history []*blades.Message, chunkCb func(string, bool)) error {
	invocation := &blades.Invocation{
		Message:     blades.UserMessage(message),
		Streamable:  true,
		History:     history,
		Instruction: instruction,
	}
	stream := inst.agent.Run(ctx, invocation)
	var sb strings.Builder
	for msg, err := range stream {
		if err != nil {
			chunkCb(sb.String(), true)
			return err
		}
		if msg != nil {
			chunk := msg.Text()
			sb.WriteString(chunk)
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

// GetOrCreate 获取或创建 Agent 实例（按 source+role 缓存，15 分钟 TTL）
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
		// 权限查询失败时降级为通用工具，并记录告警日志便于排查
		log.NewHelper(f.logger).Warnf("get role permissions failed (role=%s): %v, falling back to common tools only", role, err)
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
	var allowedToolNames []string
	for name, perm := range toolModulePerms {
		if actions, ok := modPerms[perm.module]; ok {
			if actions[perm.action] {
				allowedToolNames = append(allowedToolNames, name)
			}
		}
	}
	return append(common, allowedToolNames...)
}

func (f *AgentFactory) getSourceConfig(source string) (prompt, knowledgeTag string) {
	dateStr := time.Now().Format("2006年1月2日")
	switch source {
	case "admin":
		return fmt.Sprintf(`当前日期：%s

你是商城管理助手。你可以使用工具查询订单、商品、用户等数据。

当用户询问数据（如"有多少商品"、"昨天订单多少"、"总用户数"、"今天生意怎么样"）时，你必须使用工具来获取实时数据，不要编造答案。

常用数据查询：
1. "今天订单/待处理/营收" → 使用 query_order_stats
2. "运营概览/店铺数据" → 使用 query_dashboard_stats
3. "总用户/新增/活跃" → 使用 query_user_stats
4. "优惠券/发放记录" → 使用 query_coupons、query_coupon_stats、query_coupon_records
5. "分类结构" → 使用 query_categories

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

// formatRules 统一的输出格式规范，拼接在各场景提示词之后，保证 AI 回复结构清晰、层次感强
const formatRules = `
输出格式规范（严格遵守，层次分明）：

【结构层级】
- ## 用于主标题（如"运营概览"、"优惠券统计"）
- ### 用于分区子标题（如"核心指标"、"明细列表"）
- 正文用段落或列表，关键数据用 **加粗** 标注

【示例1 - 数据概览类】
## 今日运营概览

### 核心指标
| 指标 | 数值 | 变化 |
|---|---|---|
| 订单数 | 128 单 | ↑ 12% |
| 营收 | ￥3,680 | ↑ 8% |

### 重点关注
- 热销商品：**iPhone 17 Pro Max**（4 单）
- 待处理订单：**5 单**（超时 2 单）

【示例2 - 列表详情类】
## 用户管理

### 用户统计
- 总用户数：**520** 人
- 本月新增：**38** 人
- 活跃用户：**210** 人

### 用户详情
| 昵称 | 手机 | 会员等级 | 消费总额 |
|---|---|---|---|
| 张三 | 138****1234 | VIP2 | ￥1,280 |

【示例3 - 说明/规则类】
## 优惠券规则

### 领取条件
- 新人注册自动发放
- 满 100 元可用券自动触发

### 使用规则
- 每笔订单限用 **1 张** 优惠券
- 优惠券 **不可提现、不可叠加**
- 有效期以券面说明为准

【核心规则】
1. 主标题(##) → 子标题(###) → 正文 的层级结构
2. 关键数据、数字、强调内容用 **加粗**
3. 每个区块之间留一个空行
4. 先给总结性结论，再分区块展开细节
5. 不要客套寒暄，结尾不要写"还有什么可以帮您"
6. 整个回答不要用代码块包裹`
