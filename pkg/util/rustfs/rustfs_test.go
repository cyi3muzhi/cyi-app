package rustfs


import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ExampleRustFSClient 示例：使用RustFS客户端
func TestRustFSClient(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化客户端
	client, err := InitDefaultClient()
	if err != nil {
		fmt.Printf("初始化客户端失败: %v\n", err)
		return
	}

	// 2. 创建存储桶
	bucketName := "example-bucket"
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		fmt.Printf("检查存储桶失败: %v\n", err)
		return
	}

	if !exists {
		err = client.CreateBucket(ctx, bucketName)
		if err != nil {
			fmt.Printf("创建存储桶失败: %v\n", err)
			return
		}
		fmt.Println("存储桶创建成功")
	} else {
		fmt.Println("存储桶已存在")
	}

	// 3. 上传对象
	objectKey := "test.txt"
	body := strings.NewReader("Hello RustFS!")
	err = client.UploadObject(ctx, bucketName, objectKey, body)
	if err != nil {
		fmt.Printf("上传对象失败: %v\n", err)
		return
	}
	fmt.Println("对象上传成功")

	// 4. 列出存储桶中的对象
	objects, err := client.ListObjects(ctx, bucketName, "")
	if err != nil {
		fmt.Printf("列出对象失败: %v\n", err)
		return
	}

	fmt.Println("存储桶中的对象:")
	for _, obj := range objects {
		fmt.Printf("- %s\n", obj)
	}

	// 5. 下载对象
	data, err := client.DownloadObject(ctx, bucketName, objectKey)
	if err != nil {
		fmt.Printf("下载对象失败: %v\n", err)
		return
	}

	fmt.Printf("对象内容: %s\n", string(data))

	// 6. 获取预签名URL
	presignedURL, err := client.GetPresignedURL(ctx, bucketName, objectKey, 3600)
	if err != nil {
		fmt.Printf("获取预签名URL失败: %v\n", err)
		return
	}

	fmt.Printf("预签名URL: %s\n", presignedURL)

	// 7. 删除对象
	err = client.DeleteObject(ctx, bucketName, objectKey)
	if err != nil {
		fmt.Printf("删除对象失败: %v\n", err)
		return
	}

	fmt.Println("对象删除成功")

	// 8. 列出存储桶
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		fmt.Printf("列出存储桶失败: %v\n", err)
		return
	}

	fmt.Println("所有存储桶:")
	for _, b := range buckets {
		fmt.Printf("- %s\n", b)
	}
}

// ExampleUploadFile 示例：上传本地文件
func TestUploadFile(t *testing.T) {
	ctx := context.Background()

	// 初始化客户端
	client, err := InitDefaultClient()
	if err != nil {
		fmt.Printf("初始化客户端失败: %v\n", err)
		return
	}

	// 上传文件
	bucketName := "cyi-bucket"
	objectKey := "credentials.json"
	filePath := "./credentials.json"
	// 2. 创建存储桶
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		fmt.Printf("检查存储桶失败: %v\n", err)
		return
	}

	if !exists {
		err = client.CreateBucket(ctx, bucketName)
		if err != nil {
			fmt.Printf("创建存储桶失败: %v\n", err)
			return
		}
		fmt.Println("存储桶创建成功")
	} else {
		fmt.Println("存储桶已存在")
	}

	err = client.UploadFile(ctx, bucketName, objectKey, filePath)
	if err != nil {
		fmt.Printf("上传文件失败: %v\n", err)
		return
	}

	fmt.Println("文件上传成功")
}

// ExampleDownloadFile 示例：下载文件到本地
func TestDownloadFile(t *testing.T) {
	ctx := context.Background()

	// 初始化客户端
	client, err := InitDefaultClient()
	if err != nil {
		fmt.Printf("初始化客户端失败: %v\n", err)
		return
	}

	// 下载文件
	bucketName := "cyi-bucket"
	objectKey := "credentials.json"
	filePath := "./d-credentials.json"

	err = client.DownloadFile(ctx, bucketName, objectKey, filePath)
	if err != nil {
		fmt.Printf("下载文件失败: %v\n", err)
		return
	}

	fmt.Println("文件下载成功")
}
