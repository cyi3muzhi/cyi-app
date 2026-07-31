package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	userRpc "cyi-app/api/service/user/v1"
	"cyi-app/app/cin/admin/internal/biz"
	mid "cyi-app/pkg/kratos/server/middleware"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

// VerifyLogin 验证管理员登录，只返回管理员信息（token 由 admin BFF 自己生成）
func (r *authRepo) VerifyLogin(ctx context.Context, username, password string) (*biz.AdminAuth, error) {
	resp, err := r.data.auc.Login(ctx, &userRpc.AdminLoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	if !resp.Valid {
		return nil, fmt.Errorf("账号或密码错误")
	}

	admin := &biz.AdminAuth{
		ID:       resp.AdminId,
		Name:     resp.Name,
		Username: resp.Username,
		Role:     resp.Role,
	}

	if resp.Avatar != "" {
		admin.Avatar = r.data.resolvePresignedURL(ctx, resp.Avatar, 3600*24*30)
	}

	return admin, nil
}

// SaveToken 保存 token 到本地 Redis（单点登录：一个账号只能一个人登录）
func (r *authRepo) SaveToken(ctx context.Context, userID int64, token string) error {
	return r.data.redis.Set(ctx, fmt.Sprintf("admin:token:%d", userID), token, 24*time.Hour).Err()
}

// RemoveToken 登出，从本地 Redis 移除 token
func (r *authRepo) RemoveToken(ctx context.Context, token string) error {
	claims := &mid.UserClaims{}
	_, _, err := jwt2.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return fmt.Errorf("解析token失败: %w", err)
	}
	key := fmt.Sprintf("admin:token:%d", claims.ID)
	return r.data.redis.Del(ctx, key).Err()
}

// GetAdminInfo 根据管理员 ID 获取管理员信息（通过 gRPC 调用 user service）
func (r *authRepo) GetAdminInfo(ctx context.Context, adminID int64) (*biz.AdminAuth, error) {
	resp, err := r.data.auc.GetAdminUser(ctx, &userRpc.GetAdminUserRequest{Id: adminID})
	if err != nil {
		return nil, fmt.Errorf("获取管理员信息失败: %w", err)
	}
	if resp.Admin == nil {
		return nil, fmt.Errorf("管理员不存在")
	}

	avatar := r.data.resolvePresignedURL(ctx, resp.Admin.Avatar, 3600*24*30)

	return &biz.AdminAuth{
		ID:         resp.Admin.Id,
		Name:       resp.Admin.Name,
		Username:   resp.Admin.Username,
		Role:       resp.Admin.Role,
		Email:      resp.Admin.Email,
		Phone:      resp.Admin.Phone,
		Department: resp.Admin.Department,
		Position:   resp.Admin.Position,
		Bio:        resp.Admin.Bio,
		Avatar:     avatar,
		CreatedAt:  resp.Admin.CreatedAt,
		LastLogin:  resp.Admin.LastLogin,
	}, nil
}

// UpdateAvatar 更新管理员头像（接收 Base64 图片，上传到 OSS 后保存路径）
func (r *authRepo) UpdateAvatar(ctx context.Context, adminID int64, avatarBase64 string) (string, error) {
	objectKey := fmt.Sprintf("admin/avatar/%d/%s.jpg", adminID, uuid.New().String())
	bucketName := string(rustfs.AdminServiceBucket)

	if err := r.data.ensureAdminBucket(ctx); err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(avatarBase64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	body := bytes.NewReader(data)
	if err := r.data.oss.UploadObject(ctx, bucketName, objectKey, body); err != nil {
		return "", fmt.Errorf("上传到OSS失败: %w", err)
	}

	avatarPath := fmt.Sprintf("%s/%s", bucketName, objectKey)
	if err := r.saveAvatarToDB(ctx, adminID, avatarPath); err != nil {
		return "", fmt.Errorf("保存头像路径失败: %w", err)
	}

	// 返回公开访问 URL（通过 Nginx /oss/ 代理访问，不需要 presigned 签名）
	// publicEndpoint 通过环境变量配置为 https://admin.cyi.dev/oss
	return r.data.oss.GetPublicURL(bucketName, objectKey), nil
}

// saveAvatarToDB 保存头像 URL 到数据库（通过 gRPC 调用 user service）
func (r *authRepo) saveAvatarToDB(ctx context.Context, adminID int64, avatarURL string) error {
	_, err := r.data.auc.UpdateAdminAvatar(ctx, &userRpc.UpdateAdminAvatarRequest{
		AdminId:   adminID,
		AvatarUrl: avatarURL,
	})
	return err
}

// UpdateAdminInfo 更新管理员基本信息（通过 gRPC 调用 user service）
func (r *authRepo) UpdateAdminInfo(ctx context.Context, adminID int64, info *biz.AdminAuth) error {
	_, err := r.data.auc.UpdateAdminUser(ctx, &userRpc.UpdateAdminUserRequest{
		Id:         adminID,
		Name:       info.Name,
		Email:      info.Email,
		Phone:      info.Phone,
		Department: info.Department,
		Position:   info.Position,
		Bio:        info.Bio,
	})
	return err
}

// GetOnlineAdminIds 获取当前在线管理员 ID 列表（通过扫描 Redis 中的 token key）
func (r *authRepo) GetOnlineAdminIds(ctx context.Context) ([]int64, error) {
	keys, err := r.data.redis.Keys(ctx, "admin:token:*").Result()
	if err != nil {
		return nil, fmt.Errorf("扫描在线用户失败: %w", err)
	}
	ids := make([]int64, 0, len(keys))
	for _, key := range keys {
		var id int64
		if _, err := fmt.Sscanf(key, "admin:token:%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
