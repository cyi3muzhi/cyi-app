package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"cyi-app/pkg/util/rustfs"
)

// ensureBucket 检查 OSS bucket 是否存在，不存在则创建
// 同时确保 bucket 为公开读（允许匿名 GET 对象）
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
	// 无论新建还是已存在，都设置一次公开读 policy（幂等操作）
	// 这样旧 bucket 也能通过 Nginx /oss/ 代理匿名访问
	if err := d.oss.SetBucketPublicRead(ctx, bucketName); err != nil {
		// policy 设置失败不阻断主流程，只记日志（可能是权限问题或已设置）
		fmt.Printf("warning: 设置 bucket %s 公开读 policy 失败: %v\n", bucketName, err)
	}
	return nil
}

// ensureAdminBucket 确保 Admin 服务使用的 OSS bucket 存在
// 大多数场景使用固定的 AdminServiceBucket，封装为便捷方法
func (d *Data) ensureAdminBucket(ctx context.Context) error {
	return d.ensureBucket(ctx, string(rustfs.AdminServiceBucket))
}

// resolvePresignedURL 解析图片路径为可访问的公开 URL
// 支持两种输入格式：
//   - "bucketName/objectKey"（OSS 内部路径）：生成公开访问 URL
//   - "http(s)://..."（完整 URL）：尝试从 URL 路径中提取 bucket/objectKey 再转换，失败则原样返回
//   - 空字符串或已是完整 URL（无法解析路径）：原样返回
//
// 使用 GetPublicURL 生成不带签名的 URL，通过 Nginx /oss/ 代理访问 RustFS
// 优势：
//  1. 避免 presigned URL 经过 Nginx 代理后路径变化导致签名验证失败
//  2. URL 更简洁，不会暴露签名信息
//  3. 需要将 bucket 设为公开读（通过 RustFS 控制台或 API 设置 bucket policy）
func (d *Data) resolvePresignedURL(ctx context.Context, imageURL string, expiry int) string {
	_ = expiry // 公开 URL 不需要 expiry，保留参数兼容现有调用

	if imageURL == "" || strings.HasPrefix(imageURL, "data:image") {
		return imageURL
	}

	var bucketName, objectKey string

	if strings.HasPrefix(imageURL, "http") {
		// 已是完整 URL，尝试解析路径提取 bucket/objectKey 后生成公开 URL
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

	return d.oss.GetPublicURL(bucketName, objectKey)
}
