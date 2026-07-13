package middleware

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	jwt2.RegisteredClaims
}

// 创建白名单匹配器
func NewWhiteListMatcher(wl []string) selector.MatchFunc {
	whiteList := make(map[string]struct{})
	if wl == nil {

		return func(ctx context.Context, operation string) bool {
			return false
		}
	}
	for _, op := range wl {
		whiteList[op] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// 服务端认证中间件
func AuthServiceMiddleware(auth Auth) middleware.Middleware {
	if auth.ApiKey == "" {
		panic("AuthMiddleware: ApiKey is empty")
	}
	if len(auth.WhiteList) > 0 {
		return selector.Server(
			jwt.Server(func(token *jwt2.Token) (interface{}, error) {
				// 验证签名方法
				if _, ok := token.Method.(*jwt2.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(auth.ApiKey), nil
			}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
				return &UserClaims{}
			})),
		).
			Match(NewWhiteListMatcher(auth.WhiteList)).
			Build()
	}
	return jwt.Server(func(token *jwt2.Token) (interface{}, error) {
		return []byte(auth.ApiKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
		return &UserClaims{}
	}))
}

// 客户端认证中间件
func AuthClientMiddleware(auth Auth) middleware.Middleware {
	return jwt.Client(func(token *jwt2.Token) (interface{}, error) {
		return []byte(auth.ServiceKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256))
}
