package rustfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

// RustFSClient RustFS客户端结构体
type RustFSClient struct {
	client         *rustfs.Client
	endpoint       string
	publicEndpoint string
}

// Config RustFS配置结构体
type Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
}

// NewRustFSClient 创建RustFS客户端实例
func NewRustFSClient(cfg *Config) (*RustFSClient, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("missing required config: Endpoint, AccessKey, SecretKey")
	}

	client, err := rustfs.New(cfg.Endpoint, &rustfs.Options{
		Credentials:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       false,
		BucketLookup: types.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create RustFS client: %v", err)
	}

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}

	return &RustFSClient{
		client:         client,
		endpoint:       cfg.Endpoint,
		publicEndpoint: publicEndpoint,
	}, nil
}

// GetEndpoint 获取后端上传用的 OSS 端点地址
func (r *RustFSClient) GetEndpoint() string {
	return r.endpoint
}

// GetPublicEndpoint 获取前端访问用的 OSS 端点地址
func (r *RustFSClient) GetPublicEndpoint() string {
	return r.publicEndpoint
}

func (r *RustFSClient) CreateBucket(ctx context.Context, bucketName string) error {
	if err := r.client.Bucket().Create(ctx, bucketName); err != nil {
		return err
	}
	// 设置 bucket 为公开读（匿名用户可 GET 对象）
	// 这样前端通过 Nginx /oss/ 代理访问时不需要 presigned 签名
	return r.SetBucketPublicRead(ctx, bucketName)
}

// SetBucketPublicRead 设置 bucket 为公开读
// 通过设置 S3 bucket policy 允许匿名用户执行 s3:GetObject
func (r *RustFSClient) SetBucketPublicRead(ctx context.Context, bucketName string) error {
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucketName)
	return r.client.Bucket().SetPolicy(ctx, bucketName, policy)
}

func (r *RustFSClient) DeleteBucket(ctx context.Context, bucketName string) error {
	return r.client.Bucket().Delete(ctx, bucketName)
}

func (r *RustFSClient) ListBuckets(ctx context.Context) ([]string, error) {
	buckets, err := r.client.Bucket().List(ctx)
	if err != nil {
		return nil, err
	}

	bucketNames := make([]string, 0, len(buckets))
	for _, b := range buckets {
		bucketNames = append(bucketNames, b.Name)
	}

	return bucketNames, nil
}

func (r *RustFSClient) ListObjects(ctx context.Context, bucketName string, prefix string) ([]string, error) {
	var objectNames []string

	for obj := range r.client.Object().List(ctx, bucketName) {
		objectNames = append(objectNames, obj.Key)
	}

	return objectNames, nil
}

func (r *RustFSClient) UploadObject(ctx context.Context, bucketName, objectKey string, body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	_, err = r.client.Object().Put(ctx, bucketName, objectKey, io.NopCloser(bytes.NewReader(data)), int64(len(data)))
	return err
}

func (r *RustFSClient) UploadFile(ctx context.Context, bucketName, objectKey, filePath string) error {
	_, err := r.client.Object().FPut(ctx, bucketName, objectKey, filePath)
	return err
}

func (r *RustFSClient) DownloadObject(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	body, _, err := r.client.Object().Get(ctx, bucketName, objectKey)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}

func (r *RustFSClient) DownloadFile(ctx context.Context, bucketName, objectKey, filePath string) error {
	_, err := r.client.Object().FGet(ctx, bucketName, objectKey, filePath)
	return err
}

func (r *RustFSClient) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	return r.client.Object().Delete(ctx, bucketName, objectKey)
}

// GetPresignedURL 生成预签名下载 URL（带签名验证，私有 bucket 用）
// presigned URL 的 host 来自 publicEndpoint 配置
// 部署时 publicEndpoint 设为 https://admin.cyi.dev/oss，通过 Nginx 代理访问 RustFS
// Nginx 代理时必须保持 Host=rustfs:9000，否则签名验证失败
//
// 注意：presigned URL 生成时用的是 endpoint（如 http://rustfs:9000）
// 替换 publicEndpoint 时要处理协议前缀，避免出现 http://https://... 这种错误
func (r *RustFSClient) GetPresignedURL(ctx context.Context, bucketName, objectKey string, expiry int) (string, error) {
	presignedURL, _, err := r.client.Object().PresignGet(ctx, bucketName, objectKey, time.Duration(expiry)*time.Second, nil)
	if err != nil {
		return "", err
	}

	urlStr := presignedURL.String()
	// 如果 publicEndpoint 和 endpoint 不同，替换 endpoint 为 publicEndpoint
	// 需要同时替换协议+host，避免出现 http://https://... 的错误
	if r.endpoint != r.publicEndpoint {
		// presigned URL 格式：http://rustfs:9000/bucket/key?signature=...
		// endpoint 格式：rustfs:9000（不含协议）
		// 替换 "http://rustfs:9000" 为 publicEndpoint（如 https://admin.cyi.dev/oss）
		urlStr = strings.Replace(urlStr, "http://"+r.endpoint, r.publicEndpoint, 1)
		urlStr = strings.Replace(urlStr, "https://"+r.endpoint, r.publicEndpoint, 1)
		// 兜底：如果 endpoint 本身含协议，直接替换
		urlStr = strings.Replace(urlStr, r.endpoint, r.publicEndpoint, 1)
	}

	return urlStr, nil
}

// GetPublicURL 生成公开访问 URL（不带签名，公开 bucket 用）
// 返回格式：{publicEndpoint}/{bucketName}/{objectKey}
// 适用于通过 Nginx 反向代理访问 OSS 的场景（避免签名 host 不匹配问题）
func (r *RustFSClient) GetPublicURL(bucketName, objectKey string) string {
	// 规范化 endpoint，去掉协议前缀和末尾斜杠
	endpoint := r.publicEndpoint
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
	endpoint = strings.TrimSuffix(endpoint, "/")

	// 如果 publicEndpoint 已包含协议，直接用；否则补 http://
	if strings.HasPrefix(r.publicEndpoint, "http://") || strings.HasPrefix(r.publicEndpoint, "https://") {
		return fmt.Sprintf("%s/%s/%s", r.publicEndpoint, bucketName, objectKey)
	}
	return fmt.Sprintf("http://%s/%s/%s", endpoint, bucketName, objectKey)
}

// GetPresignedPutURL 生成预签名上传 PUT URL，前端可直接上传文件到 OSS
func (r *RustFSClient) GetPresignedPutURL(ctx context.Context, bucketName, objectKey string, expiry int) (string, error) {
	presignedURL, _, err := r.client.Object().PresignPut(ctx, bucketName, objectKey, time.Duration(expiry)*time.Second, url.Values{})
	if err != nil {
		return "", err
	}

	urlStr := presignedURL.String()
	// 替换 endpoint 为 publicEndpoint，处理协议前缀避免 http://https://...
	if r.endpoint != r.publicEndpoint {
		urlStr = strings.Replace(urlStr, "http://"+r.endpoint, r.publicEndpoint, 1)
		urlStr = strings.Replace(urlStr, "https://"+r.endpoint, r.publicEndpoint, 1)
		urlStr = strings.Replace(urlStr, r.endpoint, r.publicEndpoint, 1)
	}

	return urlStr, nil
}

func (r *RustFSClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return r.client.Bucket().Exists(ctx, bucketName)
}

func (r *RustFSClient) ObjectExists(ctx context.Context, bucketName, objectKey string) (bool, error) {
	_, err := r.client.Object().Stat(ctx, bucketName, objectKey)
	if err != nil {
		return false, nil
	}

	return true, nil
}

type BucketName string

const (
	UserServiceBucket    BucketName = "cyi-app-user"
	AdminServiceBucket   BucketName = "cyi-app-admin"
	ProductServiceBucket BucketName = "cyi-app-product"
)

// 上传文件并返回链接
func (o *RustFSClient) UploadFileAndGetURL(ctx context.Context, bucketName BucketName, objectKey string, body io.Reader) (string, error) {
	// 2. 创建存储桶
	exists, err := o.BucketExists(ctx, string(bucketName))
	if err != nil {
		return "", err
	}

	if !exists {
		err = o.CreateBucket(ctx, string(bucketName))
		if err != nil {
			return "", err
		}
	}
	// 从路径中提取存储桶名和对象键
	err = o.UploadObject(ctx, string(bucketName), objectKey, body)
	if err != nil {
		return "", err
	}

	presignedURL, err := o.GetPresignedURL(ctx, string(bucketName), objectKey, 3600)
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}
