/**
 * 悦选商城管理后台 - 通用 OSS 直传工具
 */
import { getUploadToken } from '../api/upload'

export interface UploadResult {
  objectKey: string  // OSS 对象路径，提交表单时用
  publicUrl: string  // 可访问的图片 URL，可直接用于 <img> 标签
}

/**
 * 上传文件到 OSS（前端直传，不走后端）
 * @param file 文件对象
 * @param directory OSS 目录，如 "product"、"avatar"
 * @returns UploadResult
 */
export async function uploadToOSS(file: File, directory: string = 'product'): Promise<UploadResult> {
  const token = await getUploadToken(file.name, directory)

  const response = await fetch(token.uploadUrl, {
    method: 'PUT',
    body: file,
    headers: { 'Content-Type': file.type || 'application/octet-stream' }
  })

  if (!response.ok) {
    throw new Error(`上传失败: ${response.statusText}`)
  }

  return {
    objectKey: token.objectKey,
    publicUrl: token.publicUrl
  }
}
