/**
 * 悦选商城管理后台 - 文件上传 API（前端直传 OSS）
 */
import request from './index'

export interface UploadToken {
  uploadUrl: string
  objectKey: string
  publicUrl: string
}

/** 获取上传凭证 */
export function getUploadToken(filename: string, directory: string = 'product'): Promise<UploadToken> {
  return request.post('/admin/v1/upload/token', { filename, directory })
}
