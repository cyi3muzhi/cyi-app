package data

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// UploadRepo 文件上传数据仓库
// 提供 OSS 预签名 URL 生成功能，支持前端直传
type UploadRepo struct {
	data *Data
	log  *log.Helper
}

// NewUploadRepo 创建 UploadRepo
func NewUploadRepo(data *Data, logger log.Logger) *UploadRepo {
	return &UploadRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetUploadToken 获取上传凭证
// 生成预签名 PUT URL（前端直传用）和预签名 GET URL（公开访问用）
// 返回 uploadURL（PUT地址）, objectKey（对象路径）, publicURL（公开访问地址）
func (r *UploadRepo) GetUploadToken(ctx context.Context, filename, directory string) (uploadURL, objectKey, publicURL string, err error) {
	// 从 filename 提取扩展名
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", "", "", fmt.Errorf("文件名缺少扩展名: %s", filename)
	}
	ext = strings.TrimPrefix(ext, ".")

	// 生成 objectKey: admin/{directory}/{uuid}.{ext}
	objectKey = fmt.Sprintf("admin/%s/%s.%s", directory, uuid.New().String(), ext)

	bucketName := string(rustfs.AdminServiceBucket)

	// 防御：OSS 客户端未初始化时直接返回明确错误，避免后续调用 nil 指针 panic
	// 该场景通常发生在配置缺失或测试环境未注入客户端时
	if r.data == nil || r.data.oss == nil {
		return "", "", "", fmt.Errorf("OSS 客户端未初始化，无法生成上传凭证")
	}

	// 确保 bucket 存在，不存在则创建
	exists, err := r.data.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return "", "", "", fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := r.data.oss.CreateBucket(ctx, bucketName); err != nil {
			return "", "", "", fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	// 生成预签名 PUT URL（1小时有效期，供前端直传）
	uploadURL, err = r.data.oss.GetPresignedPutURL(ctx, bucketName, objectKey, 3600)
	if err != nil {
		return "", "", "", fmt.Errorf("生成上传凭证失败: %w", err)
	}

	// 生成预签名 GET URL（30天有效期，供公开访问）
	publicURL, err = r.data.oss.GetPresignedURL(ctx, bucketName, objectKey, 3600*24*30)
	if err != nil {
		return "", "", "", fmt.Errorf("生成公开访问URL失败: %w", err)
	}

	return uploadURL, objectKey, publicURL, nil
}
