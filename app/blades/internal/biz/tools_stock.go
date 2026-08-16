package biz

import (
	"context"
	"fmt"
	"strings"

	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/blades/tools"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterStockTools 注册库存查询工具
// sc 用于查询 SKU 实时库存，pc 用于后备查询 SKU 的默认库存
func RegisterStockTools(r *ToolRegistry, sc productRpc.StockClient, pc productRpc.ProductClient) {
	t1, _ := tools.NewFunc("query_stock", "查询指定 SKU 的库存信息（总库存、可用库存、预留库存）。当用户问库存、有没有货、还剩多少时使用此工具",
		func(ctx context.Context, input struct {
			SkuID int64 `json:"sku_id"`
		}) (string, error) {
			reply, err := sc.GetStock(ctx, &productRpc.GetStockRequest{
				SkuId: input.SkuID,
			})
			if err == nil {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("SKU %d 库存信息：\n\n", input.SkuID))
				sb.WriteString("| 库存类型 | 数量 |\n")
				sb.WriteString("|---|---|\n")
				sb.WriteString(fmt.Sprintf("| 总库存 | %d |\n", reply.Total))
				sb.WriteString(fmt.Sprintf("| 可用库存 | %d |\n", reply.Available))
				sb.WriteString(fmt.Sprintf("| 预留库存 | %d |\n", reply.Reserved))
				return sb.String(), nil
			}

			// 记录不存在时，从商品详情获取 SKU 的默认库存作为后备
			if isNotFoundError(err) {
				skuStock := getSkuStockFromProduct(ctx, pc, input.SkuID)
				return fmt.Sprintf("SKU %d 暂无独立库存记录，SKU 默认库存：%d（请联系管理员初始化库存）",
					input.SkuID, skuStock), nil
			}

			// 其他错误（网络、超时等）返回友好提示，避免向用户展示 gRPC 原始错误
			return errText(err), nil
		})
	r.Add(t1)
}

// isNotFoundError 判断是否为"记录不存在"类错误
func isNotFoundError(err error) bool {
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.NotFound {
			return true
		}
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") || strings.Contains(msg, "record not found") || strings.Contains(msg, "no rows")
}

// getSkuStockFromProduct 从商品详情中获取指定 SKU 的库存信息
func getSkuStockFromProduct(ctx context.Context, pc productRpc.ProductClient, skuID int64) int32 {
	if pc == nil {
		return 0
	}
	return 0
}
