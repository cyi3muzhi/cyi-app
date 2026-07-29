package biz

import (
	"context"
	"fmt"

	productRpc "cyi-app/api/service/product/v1"
	"github.com/go-kratos/blades/tools"
)

// RegisterProductTools 注册商品工具
func RegisterProductTools(r *ToolRegistry, client productRpc.ProductClient) {
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
				return "", err
			}
			if reply.Total == 0 {
				return "没有找到匹配的商品", nil
			}
			result := fmt.Sprintf("共 %d 条商品：\n", reply.Total)
			for _, item := range reply.Items {
				statusText := "上架"
				if item.Status != 1 {
					statusText = "下架"
				}
				result += fmt.Sprintf("- %s | ￥%.2f | %s\n", item.Name, float64(item.Price)/100, statusText)
			}
			return result, nil
		})
	r.Add(t1)

	t2, _ := tools.NewFunc("get_product_detail", "查询单个商品的详细信息",
		func(ctx context.Context, input struct {
			ProductID int64 `json:"product_id"`
		}) (string, error) {
			reply, err := client.GetProduct(ctx, &productRpc.GetProductRequest{ProductId: input.ProductID})
			if err != nil {
				return "", err
			}
			result := fmt.Sprintf("商品: %s\n价格: ￥%.2f\n", reply.Product.Name, float64(reply.Product.Price)/100)
			if len(reply.Product.Specs) > 0 {
				result += "规格:\n"
				for _, spec := range reply.Product.Specs {
					vals := ""
					for i, v := range spec.Values {
						if i > 0 {
							vals += "/"
						}
						vals += v
					}
					result += fmt.Sprintf("  - %s: %s\n", spec.Name, vals)
				}
			}
			return result, nil
		})
	r.Add(t2)
}
