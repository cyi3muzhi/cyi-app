package middleware

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/middleware/selector"
)

func NewWhiteListMatcher(awl []string) selector.MatchFunc {
	whiteList := make(map[string]struct{})
	if awl == nil {
		whiteList[v1.OperationCinInterfaceLogin] = struct{}{}
		whiteList[v1.OperationCinInterfaceRegister] = struct{}{}
		return func(ctx context.Context, operation string) bool {
			return false
		}
	}
	for _, op := range awl {
		whiteList[op] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}
