package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// FallbackFunc 降级处理函数：当下游调用失败（含被熔断拦截的快速失败）时执行，返回兜底响应。
// 注意：仅适用于可安全降级的只读/非关键接口（如商品详情、推荐位、公告等）；
// 涉及资金、库存、下单等强一致操作禁止降级，避免造成数据不一致。
type FallbackFunc func(ctx context.Context, req interface{}) (interface{}, error)

// fallbackMiddleware 客户端降级中间件
// 调用下游 gRPC 出错时，若该 operation 注册了降级函数，则返回兜底响应而不是把错误抛给上游，
// 实现「熔断 + 降级」的完整链路：熔断负责快速失败，降级负责失败后的兜底表现。
func fallbackMiddleware(fallbacks map[string]FallbackFunc) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			reply, err = handler(ctx, req)
			if err == nil {
				return reply, nil
			}
			if len(fallbacks) == 0 {
				return reply, err
			}
			// 根据被调用方法的 operation（如 /user.v1.User/GetUser）匹配降级策略
			if op := operationOf(ctx); op != "" {
				if fb, ok := fallbacks[op]; ok {
					return fb(ctx, req)
				}
			}
			return reply, err
		}
	}
}

// operationOf 从客户端上下文中取出被调用的 operation 全名
func operationOf(ctx context.Context) string {
	if tr, ok := transport.FromClientContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}
