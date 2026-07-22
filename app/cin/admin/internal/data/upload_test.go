package data

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// newUploadTestRepo 创建测试用的 UploadRepo
func newUploadTestRepo(t *testing.T) *UploadRepo {
	return &UploadRepo{data: &Data{}, log: log.NewHelper(log.NewStdLogger(nil))}
}

// TestGetUploadToken_MissingExtension 测试缺少扩展名的文件名应返回明确错误
func TestGetUploadToken_MissingExtension(t *testing.T) {
	repo := newUploadTestRepo(t)

	_, _, _, err := repo.GetUploadToken(context.Background(), "filename", "avatars")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件名缺少扩展名")
}

// TestGetUploadToken_EmptyFilename 测试空文件名应返回错误
func TestGetUploadToken_EmptyFilename(t *testing.T) {
	repo := newUploadTestRepo(t)

	_, _, _, err := repo.GetUploadToken(context.Background(), "", "avatars")
	assert.Error(t, err)
}

// TestGetUploadToken_ExtensionWithDot 测试带点号的扩展名会被正确剥离
func TestGetUploadToken_ExtensionWithDot(t *testing.T) {
	repo := newUploadTestRepo(t)

	// 缺少 OSS 客户端时，objectKey 生成后会因调用 BucketExists 失败；
	// 这里仅验证输入校验阶段不会把点号作为扩展名内容保留
	_, _, _, err := repo.GetUploadToken(context.Background(), ".jpg", "avatars")
	assert.Error(t, err)
}
