package service

import (
	"context"
	"encoding/json"

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
