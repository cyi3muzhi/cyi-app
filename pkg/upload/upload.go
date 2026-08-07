// Package upload 提供文件上传安全校验工具（扩展名白名单、目录参数校验）
// 供 admin/interface 的 OSS 预签名直传接口复用，防止上传活性文件（HTML/SVG 等）与路径注入。
package upload

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// allowedExts 允许上传的文件扩展名白名单
// 只放行图片与常见文档，排除 .html/.svg/.js 等可在浏览器执行的活性文件
var allowedExts = map[string]struct{}{
	"jpg": {}, "jpeg": {}, "png": {}, "gif": {}, "webp": {},
	"pdf": {}, "doc": {}, "docx": {}, "xls": {}, "xlsx": {}, "csv": {}, "txt": {},
}

// maxFileSize 预签名直传模式下服务端无法预知文件大小，
// 实际大小由部署层 Nginx client_max_body_size 与签名条件约束，此处仅做扩展名校验。
const maxFileSize = 10 << 20

// dirPattern 目录参数允许的字符集：字母数字、下划线、连字符（防路径穿越）
var dirPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)

// ValidateFilename 校验文件名扩展名是否在白名单内
// 返回规范化后的小写扩展名（不含点），非法返回错误
func ValidateFilename(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", fmt.Errorf("文件名缺少扩展名: %s", filename)
	}
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if _, ok := allowedExts[ext]; !ok {
		return "", fmt.Errorf("不支持的文件类型: .%s（仅支持图片及常见文档）", ext)
	}
	return ext, nil
}

// ValidateDirectory 校验上传目录参数是否合法，防止路径穿越与注入
func ValidateDirectory(directory string) error {
	if !dirPattern.MatchString(directory) {
		return fmt.Errorf("非法的目录参数: %s（仅允许字母、数字、下划线、连字符）", directory)
	}
	return nil
}

// MaxFileSize 返回允许的最大文件大小（字节），供调用方展示/校验参考
func MaxFileSize() int {
	return maxFileSize
}
