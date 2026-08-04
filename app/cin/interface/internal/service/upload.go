package service

import (
	"context"

	v1 "cyi-app/api/cin/interface/v1"
)

// GetUploadToken 获取上传凭证（前端直传 OSS）
// 根据文件名和目录生成预签名 PUT URL 和公开访问 URL
func (s *CinInterface) GetUploadToken(ctx context.Context, req *v1.GetUploadTokenRequest) (*v1.GetUploadTokenReply, error) {
	uploadURL, objectKey, publicURL, err := s.uploadR.GetUploadToken(ctx, req.Filename, req.Directory)
	if err != nil {
		s.log.Errorf("获取上传凭证失败: %v", err)
		return nil, err
	}
	return &v1.GetUploadTokenReply{
		UploadUrl: uploadURL,
		ObjectKey: objectKey,
		PublicUrl: publicURL,
	}, nil
}
