/**
 * 文件上传相关 API（前端直传 OSS）
 * 流程：先向后端换取预签名 PUT URL → 前端直接把文件 PUT 到该地址 → 拿到公开访问 URL 保存
 * 注意：后端为 Kratos proto 接口，JSON 输出字段为 camelCase（uploadUrl/objectKey/publicUrl）
 */
import request from './index'

/** 上传凭证（后端返回，camelCase） */
export interface UploadToken {
  uploadUrl: string // 预签名 PUT 地址（前端直传用）
  objectKey: string // 对象存储路径
  publicUrl: string // 公开访问地址（保存到用户资料用）
}

/** 获取上传凭证 */
export function getUploadToken(data: { filename: string; directory: string }) {
  return request.post<any, UploadToken>('/user/v1/upload/token', data)
}

/**
 * 上传文件到 OSS（预签名直传）
 * @param file 要上传的文件
 * @param directory 目录名（如 avatar）
 * @returns 上传成功后的公开访问 URL
 */
export async function uploadFile(file: File, directory: string): Promise<string> {
  // 1. 向后端换取上传凭证（预签名 PUT URL）
  const token = await getUploadToken({ filename: file.name, directory })
  // 兼容 camelCase 标准输出；防御异常数据
  const uploadUrl = token.uploadUrl || (token as any)?.upload_url
  const publicUrl = token.publicUrl || (token as any)?.public_url
  if (!uploadUrl) {
    throw new Error('上传凭证无效，请重试')
  }
  // 2. 使用预签名 PUT URL 直传文件到 OSS
  const resp = await fetch(uploadUrl, {
    method: 'PUT',
    headers: { 'Content-Type': file.type || 'application/octet-stream' },
    body: file
  })
  if (!resp.ok) {
    throw new Error('文件上传失败，请重试')
  }
  // 3. 返回公开访问地址（保存到用户资料）
  return publicUrl || uploadUrl
}
