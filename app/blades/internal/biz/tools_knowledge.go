package biz

import (
	"context"

	"github.com/go-kratos/blades/tools"
)

// RegisterKnowledgeTools 注册知识库搜索工具
func RegisterKnowledgeTools(r *ToolRegistry, rag *RAGEngine) {
	t1, _ := tools.NewFunc("search_knowledge", "搜索知识库，获取平台规则、政策、商品说明等信息",
		func(ctx context.Context, input struct {
			Query string `json:"query"`
		}) (string, error) {
			results, err := rag.Retrieve(ctx, "", input.Query, 3)
			if err != nil {
				return "", err
			}
			if len(results) == 0 {
				return "知识库中没有找到相关信息", nil
			}
			return FormatResults(results), nil
		})
	r.Add(t1)
}
