package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 注册用户
func (uc *UserUseCase) Register(ctx context.Context, u *User) error {
	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.repo.FindByUsername(ctx, u.Username)
	fmt.Println("user:", user)
	if err == nil && user != nil && user.ID != 0 {
		uc.log.Infof("User already exists: %v", user)
		// 用户已存在
		return fmt.Errorf("%w: username already exists", ErrUserAlreadyRegistered)
	}

	// 检查是否是其他错误（非用户不存在错误）
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		fmt.Printf("Register err: %v\n", err)
		// 记录错误日志
		uc.log.Errorf("Failed to check user existence: %v", err)
		return fmt.Errorf("%w: check user existence failed", ErrUserRegisterFailed)
	}

	// 注册即登录：记录最近登录时间和 IP（IP 由 service 层从请求中获取后传入）
	if u.LastLoginAt == 0 {
		u.LastLoginAt = time.Now().Unix()
	}

	// 添加用户
	_, err = uc.CreateUser(ctx, u)
	if err != nil {
		uc.log.Errorf("Failed to add user: %v", err)
		return fmt.Errorf("%w: add user failed", ErrUserRegisterFailed)
	}
	return nil
}

func (uc *UserUseCase) Login(ctx context.Context, in *User, apiKey string) (*LoginReply, error) {

	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		// 用户不存在
		uc.log.Errorf("User not found: %v", err)
		return nil, fmt.Errorf("%w: user not found", ErrUserNotFound)
	}

	err = uc.repo.VerifyPassword(ctx, in)
	if err != nil {
		// 密码错误
		uc.log.Errorf("Password error: %v", err)
		return nil, fmt.Errorf("%w: password not match", ErrPasswordError)
	}

	// 更新最近登录时间和 IP（失败仅告警，不影响登录）
	if err := uc.repo.UpdateLastLogin(ctx, user.ID, time.Now().Unix(), in.LastLoginIP); err != nil {
		uc.log.Warnf("Failed to update last login for user %d: %v", user.ID, err)
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
	return &LoginReply{
		Token: signedString,
	}, nil
}
