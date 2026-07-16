package service

import (
	"context"

	adminv1 "cyi-app/api/cin/admin/v1"
)

// GetUploadToken 获取上传凭证（前端直传 OSS）
// 根据文件名和目录生成预签名 PUT URL 和公开访问 URL
func (s *AdminService) GetUploadToken(ctx context.Context, req *adminv1.GetUploadTokenRequest) (*adminv1.GetUploadTokenReply, error) {
	s.log.Infof("获取上传凭证: filename=%s, directory=%s", req.Filename, req.Directory)

	uploadURL, objectKey, publicURL, err := s.ur.GetUploadToken(ctx, req.Filename, req.Directory)
	if err != nil {
		s.log.Errorf("获取上传凭证失败: %v", err)
		return nil, err
	}

	return &adminv1.GetUploadTokenReply{
		UploadUrl: uploadURL,
		ObjectKey: objectKey,
		PublicUrl: publicURL,
	}, nil
}
