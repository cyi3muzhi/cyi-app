package data

import (
	"context"
	"fmt"

	"cyi-app/pkg/upload"
	"cyi-app/pkg/util/rustfs"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// UploadRepo 文件上传数据仓库
// 提供 OSS 预签名 URL 生成功能，支持前端直传（用户头像等）
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
// 生成预签名 PUT URL（前端直传用）和公开访问 URL
// objectKey 存放到用户服务 bucket（cyi-app-user）下，目录如 avatar/
func (r *UploadRepo) GetUploadToken(ctx context.Context, filename, directory string) (uploadURL, objectKey, publicURL string, err error) {
	// 校验文件扩展名白名单（防上传 HTML/SVG 等活性文件）
	ext, err := upload.ValidateFilename(filename)
	if err != nil {
		return "", "", "", err
	}
	// 校验目录参数，防止路径穿越与注入
	if err := upload.ValidateDirectory(directory); err != nil {
		return "", "", "", err
	}

	// 生成 objectKey: user/{directory}/{uuid}.{ext}
	objectKey = fmt.Sprintf("user/%s/%s.%s", directory, uuid.New().String(), ext)

	bucketName := string(rustfs.UserServiceBucket)

	// 防御：OSS 客户端未初始化时返回明确错误
	if r.data == nil || r.data.oss == nil {
		return "", "", "", fmt.Errorf("OSS 客户端未初始化，无法生成上传凭证")
	}

	// 确保 bucket 存在且公开读
	if err := r.ensureUserBucket(ctx); err != nil {
		return "", "", "", err
	}

	// 生成预签名 PUT URL（1小时有效期，供前端直传）
	uploadURL, err = r.data.oss.GetPresignedPutURL(ctx, bucketName, objectKey, 3600)
	if err != nil {
		return "", "", "", fmt.Errorf("生成上传凭证失败: %w", err)
	}

	// 生成公开访问 URL（通过 Nginx /oss/ 代理访问）
	publicURL = r.data.oss.GetPublicURL(bucketName, objectKey)

	return uploadURL, objectKey, publicURL, nil
}

// ensureUserBucket 确保用户服务 bucket 存在且公开读（幂等）
func (r *UploadRepo) ensureUserBucket(ctx context.Context) error {
	bucketName := string(rustfs.UserServiceBucket)
	exists, err := r.data.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := r.data.oss.CreateBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
	}
	// 设置公开读 policy（幂等）
	if err := r.data.oss.SetBucketPublicRead(ctx, bucketName); err != nil {
		r.log.Warnf("设置 bucket %s 公开读 policy 失败: %v", bucketName, err)
	}
	return nil
}
