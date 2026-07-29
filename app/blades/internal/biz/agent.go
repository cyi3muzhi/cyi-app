package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	orderRpc "cyi-app/api/service/order/v1"
	productRpc "cyi-app/api/service/product/v1"
	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/blades/internal/conf"
	pkgblades "cyi-app/pkg/blades"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/emptypb"
)

// 管理员会话 Redis key 前缀
const agentSessionPrefix = "agent:chat:session:"

// AgentUseCase 智能体用例
type AgentUseCase struct {
	agent blades.Agent
	rdb   redis.UniversalClient
	pc    productRpc.ProductClient
	uc    userRpc.UserClient
	oc    orderRpc.OrderClient
	log   *log.Helper
}

// NewAgentUseCase 创建 AgentUseCase
func NewAgentUseCase(c *conf.Blades, rdb redis.UniversalClient, pc productRpc.ProductClient, uc userRpc.UserClient, oc orderRpc.OrderClient, logger log.Logger) (*AgentUseCase, error) {
	cfg := &pkgblades.AgentConfig{}
	if c != nil {
		cfg.APIKey = c.ApiKey
		cfg.Model = c.Model
		cfg.MaxTokens = c.MaxTokens
		cfg.Temperature = float64(c.Temperature)
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("blades api_key 未配置")
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = "你是汐商城的智能助手。你可以帮助管理员查询订单、商品、用户数据。请用友好的语气简洁回答问题。当用户询问数据时，直接给出数字即可，不要复述原始数据。"
	}

	model := pkgblades.NewDeepSeekModel(cfg.APIKey, cfg.APIEndpoint, cfg.Model)
	agent, err := blades.NewAgent(
		cfg.Model,
		blades.WithModel(model),
		blades.WithInstruction(cfg.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 blades agent 失败: %w", err)
	}

	return &AgentUseCase{
		agent: agent,
		rdb:   rdb,
		pc:    pc,
		uc:    uc,
		oc:    oc,
		log:   log.NewHelper(logger),
	}, nil
}

// ─── 数据查询 ────────────────────────────────────────────────────────────────

// isDataQuery 判断消息是否像在提问数据
func isDataQuery(message string) bool {
	m := strings.TrimSpace(message)
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

type dateRange struct {
	from string
	to   string
	desc string
}

func parseDateRange(message string) (dateRange, bool) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")
	dayBeforeStr := today.AddDate(0, 0, -2).Format("2006-01-02")
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
	reRange := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日(?:到|至|~)(\d{1,2})月(\d{1,2})日`)
	if m := reRange.FindStringSubmatch(message); len(m) == 5 {
		year := today.Year()
		from := fmt.Sprintf("%04d-%02d-%02d", year, atoi(m[1]), atoi(m[2]))
		to := fmt.Sprintf("%04d-%02d-%02d", year, atoi(m[3]), atoi(m[4]))
		return dateRange{from: from, to: to, desc: m[0]}, true
	}
	reISORange := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})(?:到|至|~)(\d{4}-\d{2}-\d{2})`)
	if m := reISORange.FindStringSubmatch(message); len(m) == 3 {
		return dateRange{from: m[1], to: m[2], desc: m[1] + "至" + m[2]}, true
	}
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

// queryAdminData 根据消息内容通过 gRPC 查询后台数据
func (uc *AgentUseCase) queryAdminData(ctx context.Context, message string) string {
	if !isDataQuery(message) {
		return ""
	}
	dim := detectDimensions(message)
	if dim == dimNone {
		return ""
	}

	parts := make([]string, 0, 3)

	if dim&dimOrder != 0 && uc.oc != nil {
		if dr, found := parseDateRange(message); found {
			// 按日期范围查订单数
			req := &orderRpc.ListOrderRequest{
				Page:     1,
				PageSize: 1,
				DateFrom: dr.from,
				DateTo:   dr.to,
			}
			rep, err := uc.oc.ListOrder(ctx, req)
			if err == nil {
				parts = append(parts, fmt.Sprintf("%s订单：%d 单", dr.desc, rep.Total))
			}
		} else {
			// 无日期时查今日订单统计
			stats, err := uc.oc.OrderStats(ctx, &orderRpc.OrderStatsRequest{})
			if err == nil {
				parts = append(parts, fmt.Sprintf("今日订单：%d 单", stats.TodayCount))
			}
		}
	}

	if dim&dimUser != 0 && uc.uc != nil {
		rep, err := uc.uc.UserCount(ctx, &emptypb.Empty{})
		if err == nil {
			parts = append(parts, fmt.Sprintf("总用户数：%d", rep.Count))
		}
	}

	if dim&dimProduct != 0 && uc.pc != nil {
		rep, err := uc.pc.ListProduct(ctx, &productRpc.ListProductRequest{
			Page:     1,
			PageSize: 1,
			Status:   "1",
		})
		if err == nil {
			parts = append(parts, fmt.Sprintf("总商品数：%d", rep.Total))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return "当前系统数据：\n- " + strings.Join(parts, "\n- ")
}

// ─── 会话管理 ────────────────────────────────────────────────────────────────

type agentSessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const sessionTTL = 2 * time.Hour

func agentSessionKey(adminID int64) string {
	return agentSessionPrefix + strconv.FormatInt(adminID, 10)
}

func (uc *AgentUseCase) loadHistory(ctx context.Context, adminID int64) ([]*blades.Message, error) {
	if uc.rdb == nil {
		return nil, nil
	}
	key := agentSessionKey(adminID)
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

func (uc *AgentUseCase) saveHistory(ctx context.Context, adminID int64, userMsg, assistantMsg string) error {
	if uc.rdb == nil {
		return nil
	}
	key := agentSessionKey(adminID)
	var msgs []agentSessionMessage
	data, err := uc.rdb.Get(ctx, key).Bytes()
	if err == nil {
		json.Unmarshal(data, &msgs)
	}
	msgs = append(msgs, agentSessionMessage{Role: "user", Content: userMsg})
	msgs = append(msgs, agentSessionMessage{Role: "assistant", Content: assistantMsg})
	if len(msgs) > 40 {
		msgs = msgs[len(msgs)-40:]
	}
	payload, err := json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	return uc.rdb.Set(ctx, key, payload, sessionTTL).Err()
}

// ClearHistory 清除会话历史
func (uc *AgentUseCase) ClearHistory(ctx context.Context, adminID int64) error {
	if uc.rdb == nil {
		return nil
	}
	return uc.rdb.Del(ctx, agentSessionKey(adminID)).Err()
}

// HistoryMessage 返回给前端的聊天历史消息结构
type HistoryMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// LoadChatHistory 加载聊天历史
func (uc *AgentUseCase) LoadChatHistory(ctx context.Context, adminID int64) ([]*HistoryMessage, error) {
	if uc.rdb == nil {
		return nil, nil
	}
	key := agentSessionKey(adminID)
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

// ─── AI 对话 ─────────────────────────────────────────────────────────────────

// Chat 非流式聊天
func (uc *AgentUseCase) Chat(ctx context.Context, adminID int64, message string) (string, error) {
	history, err := uc.loadHistory(ctx, adminID)
	if err != nil {
		return "", err
	}

	// 增强消息：注入数据上下文
	if dataStr := uc.queryAdminData(ctx, message); dataStr != "" {
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

	uc.saveHistory(ctx, adminID, message, fullResponse)
	return fullResponse, nil
}

// ChatStream 流式聊天
func (uc *AgentUseCase) ChatStream(ctx context.Context, adminID int64, message string, chunkCb func(chunk string, done bool)) error {
	history, err := uc.loadHistory(ctx, adminID)
	if err != nil {
		return err
	}

	// 增强消息：注入数据上下文
	if dataStr := uc.queryAdminData(ctx, message); dataStr != "" {
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
			chunkCb(fullResponse, true)
			return err
		}
		if msg != nil {
			chunk := msg.Text()
			fullResponse += chunk
			chunkCb(chunk, false)
		}
	}

	uc.saveHistory(ctx, adminID, message, fullResponse)
	chunkCb("", true)
	return nil
}
