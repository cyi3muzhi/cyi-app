package middleware

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwt2 "github.com/golang-jwt/jwt/v5"
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

func AuthServiceMiddleware(auth Auth) middleware.Middleware {
	if auth.ApiKey == "" {
		panic("AuthMiddleware: ApiKey is empty")
	}
	if len(auth.WhiteList) > 0 {
		return selector.Server(
			jwt.Server(func(token *jwt2.Token) (interface{}, error) {
				return []byte(auth.ApiKey), nil
			}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
				return &jwt2.MapClaims{}
			})),
		).
			Match(NewWhiteListMatcher(auth.WhiteList)).
			Build()
	}
	return jwt.Server(func(token *jwt2.Token) (interface{}, error) {
		return []byte(auth.ApiKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
		return &jwt2.MapClaims{}
	}))
}

func AuthClientMiddleware(auth Auth) middleware.Middleware {
	return jwt.Client(func(token *jwt2.Token) (interface{}, error) {
		return []byte(auth.ServiceKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256))
}
