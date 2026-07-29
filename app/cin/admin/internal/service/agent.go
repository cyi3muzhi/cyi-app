package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"
	"cyi-app/pkg/kratos/service"
)

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatReply 聊天响应
type ChatReply struct {
	Response string `json:"response"`
}

// GenerateDescriptionRequest 商品描述生成请求
type GenerateDescriptionRequest struct {
	ProductName string `json:"product_name"`
	Category    string `json:"category,omitempty"`
	Specs       string `json:"specs,omitempty"`
}

// GenerateDescriptionReply 商品描述生成响应
type GenerateDescriptionReply struct {
	Description string `json:"description"`
}

// GenerateDescription AI 生成商品描述
func (s *AdminService) GenerateDescription(ctx context.Context, req *GenerateDescriptionRequest) (*GenerateDescriptionReply, error) {
	prompt := buildDescriptionPrompt(req.ProductName, req.Category, req.Specs)
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &GenerateDescriptionReply{Description: text}, nil
}

// ExpandDescriptionRequest 商品描述扩写请求
type ExpandDescriptionRequest struct {
	ProductName string `json:"product_name"`
	Content     string `json:"content"`
}

// ExpandDescriptionReply 商品描述扩写响应
type ExpandDescriptionReply struct {
	Description string `json:"description"`
}

// ExpandDescription AI 扩写商品描述
func (s *AdminService) ExpandDescription(ctx context.Context, req *ExpandDescriptionRequest) (*ExpandDescriptionReply, error) {
	prompt := buildExpandPrompt(req.ProductName, req.Content)
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &ExpandDescriptionReply{Description: text}, nil
}

// buildDescriptionPrompt 构造商品描述生成的提示词
func buildDescriptionPrompt(name, category, specs string) string {
	prompt := "请为商品" + name + "生成一段专业的商品描述。"
	if category != "" {
		prompt += "\n商品分类：" + category
	}
	if specs != "" {
		prompt += "\n规格信息：" + specs
	}
	prompt += `
要求：
1. 语言热情专业，突出商品卖点
2. 包含对目标用户的吸引力描述
3. 段落清晰，有层次感
4. 返回 HTML 格式，使用 <p> 分段、<strong> 加粗关键词、<ul><li> 列出卖点
5. 不要包含 <html>、<body>、<head> 等外层标签
6. 不要使用 Markdown 格式
7. 字数控制在 100-200 字之间
请直接输出 HTML 内容：`
	return prompt
}

// buildExpandPrompt 构造商品描述扩写的提示词
func buildExpandPrompt(name, content string) string {
	return `请扩写以下关于「` + name + `」的商品描述，使其更加丰富、详细、有吸引力：

---
` + content + `
---

要求：
1. 保留原有核心信息，补充更多卖点和细节
2. 语言热情专业，段落清晰
3. 返回 HTML 格式，使用 <p> 分段、<strong> 加粗关键词、<ul><li> 列出卖点
4. 不要包含 <html>、<body>、<head> 等外层标签
5. 不要使用 Markdown 格式
6. 扩写后控制在 200-300 字之间
请直接输出 HTML 内容：`
}

// GenerateBannerCopyRequest Banner 文案生成请求
type GenerateBannerCopyRequest struct {
	Topic    string `json:"topic"`    // 活动/商品主题
	Subtitle string `json:"subtitle"` // 补充信息（可选）
}

// GenerateBannerCopyReply Banner 文案生成响应
type GenerateBannerCopyReply struct {
	Options []BannerCopyOption `json:"options"`
}

// BannerCopyOption 一组 Banner 文案方案
type BannerCopyOption struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	BtnText  string `json:"btn_text"`
}

// GenerateBannerCopy AI 生成 Banner 文案
func (s *AdminService) GenerateBannerCopy(ctx context.Context, req *GenerateBannerCopyRequest) (*GenerateBannerCopyReply, error) {
	prompt := buildBannerCopyPrompt(req.Topic, req.Subtitle)
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	// 尝试解析 JSON，失败时放在一条里
	var options []BannerCopyOption
	if err := json.Unmarshal([]byte(text), &options); err != nil || len(options) == 0 {
		options = []BannerCopyOption{{Title: req.Topic, Subtitle: req.Subtitle, BtnText: "了解详情"}}
	}
	return &GenerateBannerCopyReply{Options: options}, nil
}

func buildBannerCopyPrompt(topic, subtitle string) string {
	prompt := "请为以下活动/商品生成3组Banner文案：\n主题：" + topic
	if subtitle != "" {
		prompt += "\n补充信息：" + subtitle
	}
	prompt += `
要求：
1. 每组包含 title（标题，10字内）、subtitle（副标题，20字内）、btn_text（按钮文字，6字内）
2. 标题要吸引眼球，副标题补充卖点，按钮文字有行动号召力
3. 直接返回 JSON 数组，格式：[{"title":"...","subtitle":"...","btn_text":"..."}]
4. 不要包含 Markdown 格式或多余文字
请直接输出 JSON：`
	return prompt
}

// GenerateCategorySEORequest 分类 SEO 生成请求
type GenerateCategorySEORequest struct {
	CategoryName string `json:"category_name"`
	Description  string `json:"description,omitempty"` // 分类描述（可选）
}

// GenerateCategorySEOReply 分类 SEO 生成响应
type GenerateCategorySEOReply struct {
	SeoTitle       string `json:"seo_title"`
	SeoKeywords    string `json:"seo_keywords"`
	SeoDescription string `json:"seo_description"`
}

// GenerateCategorySEO AI 生成分类 SEO 信息
func (s *AdminService) GenerateCategorySEO(ctx context.Context, req *GenerateCategorySEORequest) (*GenerateCategorySEOReply, error) {
	prompt := buildCategorySEOPrompt(req.CategoryName, req.Description)
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	// 尝试解析 JSON
	var result GenerateCategorySEOReply
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		result = GenerateCategorySEOReply{
			SeoTitle:       req.CategoryName,
			SeoKeywords:    req.CategoryName,
			SeoDescription: req.CategoryName + " - 精选优质商品",
		}
	}
	return &result, nil
}

func buildCategorySEOPrompt(name, desc string) string {
	prompt := "请为商品分类「" + name + "」生成SEO信息。"
	if desc != "" {
		prompt += "\n分类描述：" + desc
	}
	prompt += `
要求：
1. seo_title：SEO标题，30字内，包含分类名称和核心关键词
2. seo_keywords：SEO关键词，5-10个相关词，逗号分隔
3. seo_description：SEO描述，80-120字，自然流畅
4. 直接返回 JSON 格式：{"seo_title":"...","seo_keywords":"...","seo_description":"..."}
5. 不要包含 Markdown 格式或多余文字
请直接输出 JSON：`
	return prompt
}

// GenerateSummaryRequest 运营摘要生成请求
type GenerateSummaryRequest struct {
	TodayOrders   int32   `json:"today_orders"`
	TodayRevenue  float64 `json:"today_revenue"`
	TotalUsers    int32   `json:"total_users"`
	TotalProducts int32   `json:"total_products"`
	OrdersTrend   string  `json:"orders_trend,omitempty"` // 近14天订单趋势描述（可选）
}

// GenerateSummaryReply 运营摘要生成响应
type GenerateSummaryReply struct {
	Summary string `json:"summary"`
}

// GenerateSummary AI 生成运营摘要
func (s *AdminService) GenerateSummary(ctx context.Context, req *GenerateSummaryRequest) (*GenerateSummaryReply, error) {
	prompt := buildSummaryPrompt(req)
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return &GenerateSummaryReply{Summary: text}, nil
}

func buildSummaryPrompt(req *GenerateSummaryRequest) string {
	return `请根据以下商城经营数据生成一段简洁的运营日报摘要：

今日订单数：` + formatInt(req.TodayOrders) + `
今日营收：¥` + formatFloat(req.TodayRevenue) + `
总用户数：` + formatInt(req.TotalUsers) + `
总商品数：` + formatInt(req.TotalProducts) + `
` + optionalTrend(req.OrdersTrend) + `
要求：
1. 语言简洁自然，突出关键数据和变化
2. 如有异常数据（如订单突增/突降），给出提醒
3. 字数控制在 80-150 字之间
4. 不要包含 Markdown 格式
请直接输出摘要文本：`
}

func formatInt(v int32) string     { return fmt.Sprintf("%d", v) }
func formatFloat(v float64) string { return fmt.Sprintf("%.2f", v) }
func optionalTrend(trend string) string {
	if trend != "" {
		return "订单趋势：" + trend + "\n"
	}
	return ""
}

// SuggestQuestionsReply 智能推荐问题响应
type SuggestQuestionsReply struct {
	Questions []SuggestQuestion `json:"questions"`
}

// SuggestQuestion 一个推荐问题
type SuggestQuestion struct {
	Label string `json:"label"`
	Query string `json:"query"`
}

// SuggestQuestions AI 推荐快捷提问（根据角色权限）
func (s *AdminService) SuggestQuestions(ctx context.Context) (*SuggestQuestionsReply, error) {
	role, _ := mid.GetRoleFromClaims(ctx)

	// 获取角色的模块权限
	var allowedModules []string
	if role != "" && s.rc != nil {
		perms, err := s.rc.GetRolePermissions(ctx, role)
		if err == nil {
			for module, actions := range perms {
				hasView := false
				for _, v := range actions {
					if v {
						hasView = true
						break
					}
				}
				if hasView {
					allowedModules = append(allowedModules, module)
				}
			}
		}
	}

	prompt := "你是一个电商管理助手的智能推荐系统。"
	if len(allowedModules) > 0 {
		prompt += "当前管理员拥有以下模块的访问权限：" + strings.Join(allowedModules, "、") + "。"
	} else {
		prompt += "当前管理员尚未配置具体模块权限。"
	}
	prompt += `
请根据可访问的模块，推荐4个管理员最可能想问的问题。
要求：
1. 每个问题要简洁明了，适合作为快捷提问按钮
2. label 是按钮上显示的短文本（6字内），query 是实际发送的问题（30字内）
3. 只推荐当前管理员有权限访问的模块相关的问题
4. 直接返回 JSON 数组：[{"label":"...","query":"..."}]
5. 不要包含 Markdown 格式或多余文字
请直接输出 JSON：`
	text, err := s.agentUC.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var questions []SuggestQuestion
	if err := json.Unmarshal([]byte(text), &questions); err != nil || len(questions) == 0 {
		questions = []SuggestQuestion{
			{Label: "今日订单", Query: "今天有多少订单？"},
			{Label: "待处理订单", Query: "有哪些订单需要处理？"},
			{Label: "热销商品", Query: "最近哪些商品卖得最好？"},
			{Label: "用户增长", Query: "本月用户增长情况如何？"},
		}
	}
	return &SuggestQuestionsReply{Questions: questions}, nil
}

// Chat 发送聊天消息（非流式）
func (s *AdminService) Chat(ctx context.Context, req *ChatRequest) (*ChatReply, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}

	role, _ := mid.GetRoleFromClaims(ctx)
	response, err := s.agentUC.Chat(ctx, adminID, role, req.Message)
	if err != nil {
		return nil, err
	}
	return &ChatReply{Response: response}, nil
}

// ChatSSEHandler 处理 SSE 流式聊天请求，直接由 server/http.go 的 HandleFunc 注册
// message: 用户输入的消息
// adminID: 管理员 ID（已从 JWT 解析）
// flushFunc: 向客户端写入 SSE 数据并 flush
func (s *AdminService) ChatSSEHandler(ctx context.Context, message string, adminID int64, flushFunc func([]byte)) (string, error) {
	role, _ := mid.GetRoleFromClaims(ctx)
	return s.agentUC.ChatStream(ctx, adminID, role, message, func(chunk string) {
		data, _ := json.Marshal(map[string]string{"chunk": chunk})
		flushFunc(data)
	})
}

// ClearChatHistory 清除聊天历史
func (s *AdminService) ClearChatHistory(ctx context.Context) error {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return adminv1.ErrorUnauthorized("未登录")
	}
	return s.agentUC.ClearHistory(ctx, adminID)
}

// GetChatHistory 获取聊天历史
func (s *AdminService) GetChatHistory(ctx context.Context) ([]*biz.HistoryMessage, error) {
	adminID, err := service.GetUserID(ctx)
	if err != nil {
		return nil, adminv1.ErrorUnauthorized("未登录")
	}
	history, err := s.agentUC.LoadChatHistory(ctx, adminID)
	if err != nil {
		return nil, err
	}
	// 没有历史时返回空数组
	if history == nil {
		return []*biz.HistoryMessage{}, nil
	}
	return history, nil
}
