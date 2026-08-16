package biz

import (
	"context"
	"fmt"
	"strings"

	productRpc "cyi-app/api/service/product/v1"
	"github.com/go-kratos/blades/tools"
)

// RegisterProductTools 注册商品工具
func RegisterProductTools(r *ToolRegistry, client productRpc.ProductClient) {
	// search_products 商品搜索
	t1, _ := tools.NewFunc("search_products", "搜索查询商品列表，获取商品总数、价格、上下架状态。当用户问有多少商品、商品列表、查商品时使用此工具",
		func(ctx context.Context, input struct {
			Keyword    string `json:"keyword"`
			CategoryID int64  `json:"category_id"`
			Page       int32  `json:"page"`
			PageSize   int32  `json:"page_size"`
		}) (string, error) {
			reply, err := client.ListProduct(ctx, &productRpc.ListProductRequest{
				Keyword:    input.Keyword,
				CategoryId: input.CategoryID,
				Page:       input.Page,
				PageSize:   input.PageSize,
			})
			if err != nil {
				return errText(err), nil
			}
			if reply.Total == 0 {
				return "没有找到匹配的商品", nil
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("共 %d 条商品：\n\n", reply.Total))
			sb.WriteString("| 商品名 | 价格 | 状态 |\n")
			sb.WriteString("|---|---|---|\n")
			for _, item := range reply.Items {
				statusText := "上架"
				if item.Status != 1 {
					statusText = "下架"
				}
				sb.WriteString(fmt.Sprintf("| %s | ￥%.2f | %s |\n", item.Name, float64(item.Price)/100, statusText))
			}
			return sb.String(), nil
		})
	r.Add(t1)

	// get_product_detail 商品详情
	t2, _ := tools.NewFunc("get_product_detail", "查询单个商品的详细信息，包含所有规格组和每个SKU的详细信息（含SKU编号）。当用户想查看商品规格、价格、库存或询问具体某个规格组合时使用此工具",
		func(ctx context.Context, input struct {
			ProductID int64 `json:"product_id"`
		}) (string, error) {
			reply, err := client.GetProduct(ctx, &productRpc.GetProductRequest{ProductId: input.ProductID})
			if err != nil {
				return errText(err), nil
			}
			p := reply.Product
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("商品 #%d 详情：\n\n", p.ProductId))
			sb.WriteString("| 字段 | 值 |\n")
			sb.WriteString("|---|---|\n")
			sb.WriteString(fmt.Sprintf("| 商品名 | %s |\n", p.Name))
			sb.WriteString(fmt.Sprintf("| 价格 | ￥%.2f |\n", float64(p.Price)/100))

			// 输出规格组
			if len(p.Specs) > 0 {
				for _, spec := range p.Specs {
					vals := ""
					for i, v := range spec.Values {
						if i > 0 {
							vals += "/"
						}
						vals += v
					}
					sb.WriteString(fmt.Sprintf("| 规格·%s | %s |\n", spec.Name, vals))
				}
			}
			sb.WriteString("\n")

			// 输出 SKU 列表（含 SKU ID）
			if len(p.Skus) > 0 {
				sb.WriteString("**SKU 列表：**\n\n")
				sb.WriteString("| SKU ID | 规格 | 价格 | 库存 | 销量 |\n")
				sb.WriteString("|---|---|---|---|---|\n")
				for _, sku := range p.Skus {
					specDesc := ""
					for k, v := range sku.Spec {
						if specDesc != "" {
							specDesc += " "
						}
						specDesc += fmt.Sprintf("%s:%s", k, v)
					}
					sb.WriteString(fmt.Sprintf("| %d", sku.Id))
					if sku.SkuCode != "" {
						sb.WriteString(fmt.Sprintf(" (编码:%s)", sku.SkuCode))
					}
					sb.WriteString(fmt.Sprintf(" | %s | ￥%.2f | %d | %d |\n",
						specDesc, float64(sku.Price)/100, sku.Stock, sku.Sales))
				}
			}
			return sb.String(), nil
		})
	r.Add(t2)
}
