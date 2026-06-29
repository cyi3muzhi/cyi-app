package middleware

import (
	"context"

	com "cyi-app/api/common"

	"github.com/go-kratos/kratos/v2/middleware"
)

func Validate() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if v, ok := req.(interface{ Validate() error }); ok {
				if err := v.Validate(); err != nil {
					return nil, com.ErrorBadRequest("参数校验失败: %s", err.Error())
				}
			}
			return handler(ctx, req)
		}
	}
}
