package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"` // 显示名（管理员姓名/用户昵称），可能为空
	Role     string `json:"role"`
	jwt2.RegisteredClaims
}

// TokenValidator 验证 token 是否有效（如检查 Redis 黑名单）
type TokenValidator func(ctx context.Context, rawToken string, claims *UserClaims) error

// tokenContextKey context 中存储 token 的 key
type tokenContextKey struct{}

// SetTokenToContext 将 token 存入 context
func SetTokenToContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

// GetTokenFromContext 从 context 中获取 token
func GetTokenFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	token, ok := ctx.Value(tokenContextKey{}).(string)
	return token, ok
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

	jwtMw := jwt.Server(func(token *jwt2.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt2.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(auth.ApiKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256), jwt.WithClaims(func() jwt2.Claims {
		return &UserClaims{}
	}))

	mainMw := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var rawToken string

			tr, ok := transport.FromServerContext(ctx)
			if ok {
				authHeader := tr.RequestHeader().Get("Authorization")
				rawToken = strings.TrimPrefix(authHeader, "Bearer ")
			}

			if rawToken != "" {
				ctx = SetTokenToContext(ctx, rawToken)
			}

			return jwtMw(func(ctx context.Context, req interface{}) (interface{}, error) {
				if auth.TokenValidator != nil {
					token, ok := GetTokenFromContext(ctx)
					if !ok || token == "" {
						return nil, fmt.Errorf("JWT token is missing")
					}

					claims, ok := jwt.FromContext(ctx)
					if !ok {
						return nil, fmt.Errorf("no jwt claims in context")
					}
					userClaims, ok := claims.(*UserClaims)
					if !ok {
						return nil, fmt.Errorf("invalid claims type")
					}
					if err := auth.TokenValidator(ctx, token, userClaims); err != nil {
						return nil, err
					}
				}
				return handler(ctx, req)
			})(ctx, req)
		}
	}

	if len(auth.WhiteList) > 0 {
		return selector.Server(mainMw).Match(NewWhiteListMatcher(auth.WhiteList)).Build()
	}
	return mainMw
}

// 客户端认证中间件
func AuthClientMiddleware(auth Auth) middleware.Middleware {
	return jwt.Client(func(token *jwt2.Token) (interface{}, error) {
		return []byte(auth.ServiceKey), nil
	}, jwt.WithSigningMethod(jwt2.SigningMethodHS256))
}
