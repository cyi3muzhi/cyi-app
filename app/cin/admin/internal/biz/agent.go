package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cyi-app/app/cin/admin/internal/conf"
	pkgblades "cyi-app/pkg/blades"
	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// 管理员会话 Redis key 前缀（与用户侧隔离）
const adminSessionPrefix = "admin:chat:session:"

// AgentUseCase 管理员智能体用例
type AgentUseCase struct {
	agent    blades.Agent
	rdb      redis.UniversalClient
	dc       *DashboardUseCase
	oc       *OrderUseCase
	enforcer mid.EnforcerProvider
	log      *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(c *conf.Blades, rdb redis.UniversalClient, dc *DashboardUseCase, oc *OrderUseCase, enforcer mid.EnforcerProvider, logger log.Logger) (*AgentUseCase, error) {
	cfg := &pkgblades.AgentConfig{}
	if c != nil {
		cfg.APIKey = c.ApiKey
		cfg.APIEndpoint = c.ApiEndpoint
		cfg.Model = c.Model
		cfg.SystemPrompt = c.SystemPrompt
		cfg.MaxTokens = c.MaxTokens
		cfg.Temperature = c.Temperature
	}
	// 默认系统提示词
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = "你是悦选商城管理后台的智能助手。你可以帮助管理员查询订单、商品、用户数据。请用友好的语气简洁回答问题。当用户询问数据时，直接给出数字即可，不要复述原始数据。"
	}

	// 创建 blades agent（仅文本生成，不注入任何工具）
	model := pkgblades.NewDeepSeekModel(cfg.APIKey, cfg.APIEndpoint)
	agent, err := blades.NewAgent(
		cfg.Model,
		blades.WithModel(model),
		blades.WithInstruction(cfg.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 blades agent 失败: %w", err)
	}

	return &AgentUseCase{
		agent:    agent,
		rdb:      rdb,
		dc:       dc,
		oc:       oc,
		enforcer: enforcer,
		log:      log.NewHelper(logger),
	}, nil
}

// 查询意图判断：消息是否像在提问数据
func isDataQuery(message string) bool {
	m := strings.TrimSpace(message)
	// 必须以疑问词开头或包含典型问句模式
	questionPrefixes := []string{"今天", "昨天", "本月", "上月", "最近", "当前", "目前", "现在", "查", "看", "多少", "几", "统计"}
	questionSuffixes := []string{"吗", "？", "?", "多少", "几个", "怎么样"}

	hasQuestion := false
	for _, p := range questionPrefixes {
		if strings.HasPrefix(m, p) {
			hasQuestion = true
			break
		}
	}
	if !hasQuestion {
		for _, s := range questionSuffixes {
			if strings.Contains(m, s) {
				hasQuestion = true
				break
			}
		}
	}
	return hasQuestion
}

// 数据维度判定：根据消息内容判断需要查询哪些数据
type queryDimension int

const (
	dimNone    queryDimension = 0
	dimOrder   queryDimension = 1
	dimUser    queryDimension = 2
	dimProduct queryDimension = 4
	dimAll     queryDimension = 7
)

func detectDimensions(message string) queryDimension {
	m := strings.ToLower(message)
	var dim queryDimension

	// 通用统计关键词 → 查全部
	generalKeywords := []string{"统计", "stats", "概况", "数据", "dashboard", "面板", "营收", "revenue"}
	for _, k := range generalKeywords {
		if strings.Contains(m, k) {
			return dimAll
		}
	}

	if strings.Contains(m, "订单") || strings.Contains(m, "order") || strings.Contains(m, "单") {
		dim |= dimOrder
	}
	if strings.Contains(m, "用户") || strings.Contains(m, "会员") || strings.Contains(m, "user") {
		dim |= dimUser
	}
	if strings.Contains(m, "商品") || strings.Contains(m, "产品") || strings.Contains(m, "product") {
		dim |= dimProduct
	}

	return dim
}

/* ────────── 日期范围解析 ────────── */

// dateRange 日期范围
type dateRange struct {
	from string // YYYY-MM-DD
	to   string // YYYY-MM-DD
	desc string // 自然语言描述，如"昨天"、"7月1日至7月24日"
}

// parseDateRange 从消息中解析日期范围，返回 (dateRange, true) 或 (zero, false)
func parseDateRange(message string) (dateRange, bool) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")
	dayBeforeStr := today.AddDate(0, 0, -2).Format("2006-01-02")

	// 精确关键词匹配
	switch {
	case strings.Contains(message, "昨天"):
		return dateRange{from: yesterdayStr, to: yesterdayStr, desc: "昨天"}, true
	case strings.Contains(message, "前天"):
		return dateRange{from: dayBeforeStr, to: dayBeforeStr, desc: "前天"}, true
	case strings.Contains(message, "今天"):
		return dateRange{from: todayStr, to: todayStr, desc: "今天"}, true
	case strings.Contains(message, "本周"):
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		return dateRange{from: monday.Format("2006-01-02"), to: todayStr, desc: "本周"}, true
	case strings.Contains(message, "本月"):
		firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		return dateRange{from: firstOfMonth.Format("2006-01-02"), to: todayStr, desc: "本月"}, true
	case strings.Contains(message, "上月"):
		firstOfLastMonth := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, today.Location())
		lastOfLastMonth := firstOfLastMonth.AddDate(0, 1, -1)
		return dateRange{from: firstOfLastMonth.Format("2006-01-02"), to: lastOfLastMonth.Format("2006-01-02"), desc: "上月"}, true
	}

	// 匹配 "X月Y日到X月Y日" 或 "X月Y日至X月Y日"
	reRange := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日(?:到|至|~)(\d{1,2})月(\d{1,2})日`)
	if m := reRange.FindStringSubmatch(message); len(m) == 5 {
		year := today.Year()
		from := fmt.Sprintf("%04d-%02d-%02d", year, atoi(m[1]), atoi(m[2]))
		to := fmt.Sprintf("%04d-%02d-%02d", year, atoi(m[3]), atoi(m[4]))
		return dateRange{from: from, to: to, desc: m[0]}, true
	}

	// 匹配单个 "X月Y日"
	reSingle := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`)
	if m := reSingle.FindStringSubmatch(message); len(m) == 3 {
		date := fmt.Sprintf("%04d-%02d-%02d", today.Year(), atoi(m[1]), atoi(m[2]))
		return dateRange{from: date, to: date, desc: m[0]}, true
	}

	// 匹配 "YYYY-MM-DD到YYYY-MM-DD"
	reISORange := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})(?:到|至|~)(\d{4}-\d{2}-\d{2})`)
	if m := reISORange.FindStringSubmatch(message); len(m) == 3 {
		return dateRange{from: m[1], to: m[2], desc: m[1] + "至" + m[2]}, true
	}

	// 匹配单个 "YYYY-MM-DD"
	reISO := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	if m := reISO.FindString(message); m != "" {
		return dateRange{from: m, to: m, desc: m}, true
	}

	return dateRange{}, false
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

/* ────────── 数据查询 ────────── */

// queryAdminData 根据消息内容查询后台数据，若匹配则返回格式化数据字符串，否则返回空字符串
func (uc *AgentUseCase) queryAdminData(ctx context.Context, message string, adminID int64, role string) string {
	// 第一步：判断是否真的是查询意图
	if !isDataQuery(message) {
		return ""
	}

	// 第二步：判断需要查哪些维度
	dim := detectDimensions(message)
	if dim == dimNone {
		return ""
	}

	// 第三步：权限校验 + 按需查询
	parts := make([]string, 0, 3)

	if dim&dimOrder != 0 {
		if ok, _ := uc.enforcer.Enforce(role, "/admin.order/list", "GET"); ok {
			// 尝试解析日期范围，有日期按日期查，无日期查今天
			if dr, found := parseDateRange(message); found {
				_, total, err := uc.oc.List(ctx, 1, 1, "", "", dr.from, dr.to)
				if err == nil {
					parts = append(parts, fmt.Sprintf("%s订单：%d 单", dr.desc, total))
				}
			} else {
				stats, err := uc.dc.GetDashboardStats(ctx)
				if err == nil {
					parts = append(parts, fmt.Sprintf("今日订单：%d 单", stats.TodayOrders))
				}
			}
		}
	}
	if dim&dimUser != 0 {
		if ok, _ := uc.enforcer.Enforce(role, "/admin.user/list", "GET"); ok {
			stats, err := uc.dc.GetDashboardStats(ctx)
			if err == nil {
				parts = append(parts, fmt.Sprintf("总用户数：%d", stats.TotalUsers))
			}
		}
	}
	if dim&dimProduct != 0 {
		if ok, _ := uc.enforcer.Enforce(role, "/admin.product/list", "GET"); ok {
			stats, err := uc.dc.GetDashboardStats(ctx)
			if err == nil {
				parts = append(parts, fmt.Sprintf("总商品数：%d", stats.TotalProducts))
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	result := "当前系统数据：\n- " + strings.Join(parts, "\n- ")
	return result
}

// 管理员的 session key，每个管理员独立会话
func adminSessionKey(adminID int64) string {
	return adminSessionPrefix + strconv.FormatInt(adminID, 10)
}

// agentSessionMessage Redis 中存储的管理员聊天记录
type agentSessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const sessionTTL = 2 * time.Hour

// loadHistory 从 Redis 加载管理员会话历史
func (uc *AgentUseCase) loadHistory(ctx context.Context, adminID int64) ([]*blades.Message, error) {
	key := adminSessionKey(adminID)
	data, err := uc.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var msgs []agentSessionMessage
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	history := make([]*blades.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case "user":
			history = append(history, blades.UserMessage(m.Content))
		case "assistant":
			history = append(history, &blades.Message{
				Role:   blades.RoleAssistant,
				Parts:  blades.NewMessageParts(m.Content),
				Status: blades.StatusCompleted,
			})
		}
	}
	return history, nil
}

// saveHistory 保存管理员会话历史
func (uc *AgentUseCase) saveHistory(ctx context.Context, adminID int64, userMsg, assistantMsg string) error {
	key := adminSessionKey(adminID)

	var msgs []agentSessionMessage
	data, err := uc.rdb.Get(ctx, key).Bytes()
	if err == nil {
		json.Unmarshal(data, &msgs)
	}

	msgs = append(msgs, agentSessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, agentSessionMessage{Role: "assistant", Content: assistantMsg})

	// 保留最近 20 轮对话
	if len(msgs) > 40 {
		msgs = msgs[len(msgs)-40:]
	}

	payload, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return uc.rdb.Set(ctx, key, payload, sessionTTL).Err()
}

// ClearHistory 清除管理员会话历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, adminID int64) error {
	return uc.rdb.Del(ctx, adminSessionKey(adminID)).Err()
}

// HistoryMessage 返回给前端的聊天历史消息结构
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// LoadChatHistory 加载管理员聊天历史（返回前端格式）
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, adminID int64) ([]*HistoryMessage, error) {
	key := adminSessionKey(adminID)
	data, err := uc.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	var msgs []agentSessionMessage
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}
	// 转为前端格式
	history := make([]*HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		history = append(history, &HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: time.Now().UnixMilli(),
		})
	}
	return history, nil
}

// Chat 发送聊天消息（非流式）
func (uc *AgentUseCase) Chat(ctx context.Context, adminID int64, role string, message string) (string, error) {
	history, err := uc.loadHistory(ctx, adminID)
	if err != nil {
		return "", err
	}

	// 保存原始消息用于历史记录，增强后的消息发给 AI
	originalMsg := message
	if dataStr := uc.queryAdminData(ctx, message, adminID, role); dataStr != "" {
		message = fmt.Sprintf("%s\n\n用户问题：%s", dataStr, message)
	}

	invocation := &blades.Invocation{
		Message: blades.UserMessage(message),
	}
	if len(history) > 0 {
		invocation.History = history
	}

	stream := uc.agent.Run(ctx, invocation)
	var fullResponse string
	for msg, err := range stream {
		if err != nil {
			return fullResponse, err
		}
		if msg != nil {
			fullResponse += msg.Text()
		}
	}

	// 保存原始消息（不含注入的数据），防止历史膨胀
	uc.saveHistory(ctx, adminID, originalMsg, fullResponse)
	return fullResponse, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, adminID int64, role string, message string, chunkCb func(chunk string)) (string, error) {
	history, err := uc.loadHistory(ctx, adminID)
	if err != nil {
		return "", err
	}

	// 保存原始消息用于历史记录，增强后的消息发给 AI
	originalMsg := message
	if dataStr := uc.queryAdminData(ctx, message, adminID, role); dataStr != "" {
		message = fmt.Sprintf("%s\n\n用户问题：%s", dataStr, message)
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

	// 保存原始消息（不含注入的数据），防止历史膨胀
	uc.saveHistory(ctx, adminID, originalMsg, fullResponse)
	return fullResponse, nil
}
