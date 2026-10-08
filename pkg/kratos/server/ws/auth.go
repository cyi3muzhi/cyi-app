package ws

import (
	"fmt"

	mid "cyi-app/pkg/kratos/server/middleware"

	"github.com/golang-jwt/jwt/v5"
)

// ParseToken 解析 JWT 并返回 UserClaims（HS256）
// 用户端与客服端 WS 均通过 query 参数携带 token（浏览器 WS 无法自定义 header）。
// secret 与 httproute.JWTAuthMiddleware / AuthServiceMiddleware 使用同一密钥。
func ParseToken(token string, secret []byte) (*mid.UserClaims, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("jwt secret is empty")
	}
	parsed, err := jwt.ParseWithClaims(token, &mid.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*mid.UserClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
