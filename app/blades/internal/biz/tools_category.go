package biz

import (
	"context"
	"fmt"
	"strings"

	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/blades/tools"
)

// RegisterCategoryTools 注册商品分类工具
func RegisterCategoryTools(r *ToolRegistry, client productRpc.CategoryClient) {
	// query_categories 分类树
	t1, _ := tools.NewFunc("query_categories", "查询商品分类树（含层级、启用状态、商品数量）。当用户问'有哪些分类'、'分类结构'时调用",
		func(ctx context.Context, _ struct{}) (string, error) {
			reply, err := client.ListCategoryTree(ctx, &productRpc.ListCategoryTreeRequest{})
			if err != nil {
				return errText(err), nil
			}
			if len(reply.Items) == 0 {
				return "暂无商品分类", nil
			}
			var sb strings.Builder
			sb.WriteString("商品分类树：\n\n")
			renderCategoryTree(reply.Items, 0, &sb)
			return sb.String(), nil
		})
	r.Add(t1)
}

// renderCategoryTree 递归渲染分类树为 Markdown 列表
func renderCategoryTree(nodes []*productRpc.CategoryNode, depth int, sb *strings.Builder) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		state := "启用"
		if !node.Status {
			state = "停用"
		}
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		sb.WriteString(fmt.Sprintf("%s- **%s**（%s，%d 个商品）\n", indent, node.Name, state, node.Count))
		if len(node.Children) > 0 {
			renderCategoryTree(node.Children, depth+1, sb)
		}
	}
}
