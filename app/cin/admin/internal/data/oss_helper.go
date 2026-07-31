package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"cyi-app/pkg/util/rustfs"
)

// ensureBucket 检查 OSS bucket 是否存在，不存在则创建
// 该模式在 upload/product/auth/export 等多个 repo 中重复出现，抽取为公共方法
func (d *Data) ensureBucket(ctx context.Context, bucketName string) error {
	exists, err := d.oss.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := d.oss.CreateBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
	}
	return nil
}

// ensureAdminBucket 确保 Admin 服务使用的 OSS bucket 存在
// 大多数场景使用固定的 AdminServiceBucket，封装为便捷方法
func (d *Data) ensureAdminBucket(ctx context.Context) error {
	return d.ensureBucket(ctx, string(rustfs.AdminServiceBucket))
}

// resolvePresignedURL 解析图片路径为可访问的 presigned URL
// 支持两种输入格式：
//   - "bucketName/objectKey"（OSS 内部路径）：直接生成 presigned URL
//   - "http(s)://..."（完整 URL）：尝试从 URL 路径中提取 bucket/objectKey 再转换，失败则原样返回
//   - 空字符串或已是完整 URL（无法解析路径）：原样返回
//
// presigned URL 的 host 来自 public_endpoint 配置
// 部署时 public_endpoint 设为 https://admin.cyi.dev/oss，通过 Nginx 代理访问 RustFS
// Nginx 代理时必须保持 Host=rustfs:9000，否则签名验证失败
func (d *Data) resolvePresignedURL(ctx context.Context, imageURL string, expiry int) string {
	if imageURL == "" || strings.HasPrefix(imageURL, "data:image") {
		return imageURL
	}

	var bucketName, objectKey string

	if strings.HasPrefix(imageURL, "http") {
		// 已是完整 URL，尝试解析路径提取 bucket/objectKey 后生成 presigned URL
		parsedURL, err := url.Parse(imageURL)
		if err != nil || parsedURL.Path == "" {
			return imageURL
		}
		pathParts := strings.SplitN(strings.TrimPrefix(parsedURL.Path, "/"), "/", 2)
		if len(pathParts) == 2 {
			bucketName, objectKey = pathParts[0], pathParts[1]
		} else {
			return imageURL
		}
	} else {
		// bucket/key 格式
		parts := strings.SplitN(imageURL, "/", 2)
		if len(parts) != 2 {
			return imageURL
		}
		bucketName, objectKey = parts[0], parts[1]
	}

	presigned, err := d.oss.GetPresignedURL(ctx, bucketName, objectKey, expiry)
	if err != nil {
		return imageURL
	}
	return presigned
}
