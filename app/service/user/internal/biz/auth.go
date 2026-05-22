package biz

import (
	"context"
	"time"

	v1 "cyi-app/api/service/user/v1"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 注册用户
func (uc *UserUseCase) Register(ctx context.Context, u *User) error {
	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.repo.FindByUsername(ctx, u.Username)
	if err == nil && user != nil && user.ID != 0 {
		uc.log.Infof("User already exists: %v", user)
		// 用户已存在
		return v1.ErrorUserAlreadyRegistered("username already exists")
	}

	// 检查是否是其他错误（非用户不存在错误）
	if err != nil && !v1.IsUserNotFound(err) {
		// 记录错误日志
		uc.log.Errorf("Failed to check user existence: %v", err)
		return v1.ErrorUserRegisterFailed("check user existence failed")
	}

	// 添加用户
	_, err = uc.repo.CreateUser(ctx, u)
	if err != nil {
		uc.log.Errorf("Failed to add user: %v", err)
		return v1.ErrorUserRegisterFailed("add user failed")
	}
	return nil
}

func (uc *UserUseCase) Login(ctx context.Context, in *User, apiKey string) (*v1.LoginReply, error) {

	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		// 用户不存在
		uc.log.Errorf("User not found: %v", err)
		return nil, v1.ErrorUserNotFound("user not found")
	}

	err = uc.repo.VerifyPassword(ctx, user)
	if err != nil {
		// 密码错误
		uc.log.Errorf("Password error: %v", err)
		return nil, v1.ErrorPasswordError("password not match")
	}
	// generate token
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, server.UserClaims{
		ID:       user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cin",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	})
	signedString, err := claims.SignedString([]byte(apiKey))
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: signedString,
	}, nil
}
