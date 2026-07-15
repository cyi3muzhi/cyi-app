/**
 * 悦选商城管理后台 - API 类型定义
 */

/** 管理员信息 */
export interface AdminInfo {
  id: number
  name: string
  username: string
  role: string
  email: string
  phone: string
  department: string
  position: string
  bio: string
  avatar: string
  created_at?: string
  last_login?: string
}

/** 登录响应 */
export interface LoginReply {
  token: string
  admin: AdminInfo
}

/** 获取管理员信息响应 */
export interface GetAdminInfoReply {
  admin: AdminInfo
}

/** 更换头像响应 */
export interface ChangeAvatarReply {
  success: boolean
  avatar_url?: string
  avatarUrl?: string
}

/** 更新基本信息响应 */
export interface UpdateAdminInfoReply {
  success: boolean
}
