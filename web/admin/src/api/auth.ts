/**
 * cyi商城管理后台 - 认证 API
 */
import request from './index'
import type { LoginReply, GetAdminInfoReply, ChangeAvatarReply, UpdateAdminInfoReply } from '../types/api'

/**
 * 管理员登录
 * @param username - 管理员账号
 * @param password - 管理员密码
 */
export function login(username: string, password: string): Promise<LoginReply> {
  return request.post('/admin/v1/auth/login', { username, password })
}

/**
 * 管理员登出
 * @param token - 要登出的 token
 */
export function logout(token?: string): Promise<void> {
  return request.post('/admin/v1/auth/logout', token ? { token } : undefined)
}

/** 获取当前用户的角色权限（无需角色管理权限） */
export function getMyPermissions(): Promise<{ permissions: Record<string, { values: Record<string, boolean> }> }> {
  return request.get('/admin/v1/auth/my-permissions')
}

/**
 * 获取当前管理员信息
 */
export function getAdminInfo(): Promise<GetAdminInfoReply> {
  return request.get('/admin/v1/auth/info')
}

/**
 * 更换头像（上传 Base64 图片）
 * @param avatarBase64 - Base64 编码的图片数据
 */
export function changeAvatar(avatarBase64: string): Promise<ChangeAvatarReply> {
  return request.post('/admin/v1/auth/avatar', { avatar_base64: avatarBase64 })
}

/**
 * 更新管理员基本信息
 * @param data - 基本信息 {name, email, phone, department, position, bio}
 */
export function updateAdminInfo(data: {
  name: string
  email: string
  phone: string
  department: string
  position: string
  bio: string
}): Promise<UpdateAdminInfoReply> {
  return request.put('/admin/v1/auth/info', data)
}
