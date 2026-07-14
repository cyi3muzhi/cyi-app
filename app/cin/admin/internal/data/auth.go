package data

import (
	"context"
	"fmt"
	"time"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// authRepo 认证数据仓库（通过 gRPC 调用 user service）
type authRepo struct {
	data *Data
	log  *log.Helper
}

// NewAuthRepo 创建认证数据仓库
func NewAuthRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &authRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// VerifyLogin 验证管理员登录，返回 admin + token（token 由 user 微服务生成并返回）
func (r *authRepo) VerifyLogin(ctx context.Context, username, password string) (*biz.AdminAuth, string, error) {
	resp, err := r.data.auc.Login(ctx, &userRpc.AdminLoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, "", fmt.Errorf("登录失败: %w", err)
	}
	if !resp.Valid {
		return nil, "", fmt.Errorf("账号或密码错误")
	}
	return &biz.AdminAuth{
		ID:       resp.AdminId,
		Name:     resp.Name,
		Username: resp.Username,
		Role:     resp.Role,
	}, resp.Token, nil
}

// RemoveToken 登出，调用 user service 移除 Redis 中的 token
func (r *authRepo) RemoveToken(ctx context.Context, token string) error {
	_, err := r.data.auc.Logout(ctx, &userRpc.AdminLogoutRequest{Token: token})
	return err
}

// SaveToken 保存 token
func (r *authRepo) SaveToken(ctx context.Context, userID int64, token string) error {
	_ = token
	return nil
}

// GetByToken 根据 token 获取管理员信息
func (r *authRepo) GetByToken(ctx context.Context, token string) (*biz.AdminAuth, error) {
	_ = token
	return nil, nil
}

// Ensure time import is used
var _ = time.Now
