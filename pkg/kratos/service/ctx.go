package service

import (
	"context"
	"errors"
	"fmt"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
)

// 返回用户id

func GetUserID(ctx context.Context) (int64, error) {
	claims, ok := jwt.FromContext(ctx)
	fmt.Println("claims:", claims)
	if !ok {
		return 0, errors.New("claims not found")
	}
	userClaims, ok := claims.(*server.UserClaims)
	if !ok {
		return 0, errors.New("claims not found")
	}
	return userClaims.ID, nil
}
