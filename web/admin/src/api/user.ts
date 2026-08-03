/**
 * 悦选商城管理后台 - 用户管理 API
 * 对接 user service（用户端用户表，非管理员表）
 */
import request from './index'

/** 用户 API 基础路径 */
const BASE = '/admin/v1/users'

/** 用户信息（对齐后端 admin UserItem） */
export interface UserItem {
  id: number
  username: string        // 登录账号
  nickname: string        // 展示名
  phone: string
  email: string
  avatar: string
  gender: number          // 0=未知 1=男 2=女
  level: string           // 会员等级 regular/gold/diamond
  lastLoginAt: number     // 最近登录时间戳（Unix 秒）
  registerSource: string  // 注册来源 web/app/wx_mini
  status: number          // 1=正常 2=冻结
  createdAt: number       // 注册时间戳（Unix 秒）
}

/** 用户列表请求参数 */
export interface ListUserParams {
  page: number
  pageSize: number
  keyword?: string
  level?: string
  /** 0=全部 1=正常 2=冻结 */
  status?: number
}

/** 用户列表响应 */
export interface ListUserResponse {
  items: UserItem[]
  total: number
}

/** 用户统计数据 */
export interface UserStatsResponse {
  totalUsers: number
  todayNew: number
  todayNewGrowth: number
  activeUsers: number
  memberUsers: number
  memberRatio: number
}

/** 创建用户请求 */
export interface CreateUserParams {
  username: string
  phone: string
  email?: string
  password: string
  nickname?: string
  registerSource?: string
}

/** 更新用户请求 */
export interface UpdateUserParams {
  nickname?: string
  phone?: string
  email?: string
  gender?: number
  birthday?: string
  level?: string
}

/** 冻结/解冻响应 */
export interface ToggleUserStatusResponse {
  success: boolean
  /** 1=正常 2=冻结 */
  newStatus: number
}

/**
 * 获取用户列表
 */
export async function listUser(params: ListUserParams): Promise<ListUserResponse> {
  const res = await request.get<any, ListUserResponse>(BASE, { params })
  return res
}

/**
 * 获取用户详情
 */
export async function getUser(id: number): Promise<{ user: UserItem }> {
  const res = await request.get<any, { user: UserItem }>(`${BASE}/${id}`)
  return res
}

/**
 * 创建用户
 */
export async function createUser(data: CreateUserParams): Promise<{ id: number }> {
  const res = await request.post<any, { id: number }>(BASE, data)
  return res
}

/**
 * 更新用户资料
 */
export async function updateUser(id: number, data: UpdateUserParams): Promise<{ success: boolean }> {
  const res = await request.put<any, { success: boolean }>(`${BASE}/${id}`, data)
  return res
}

/**
 * 删除用户（软删除）
 */
export async function deleteUser(id: number): Promise<{ success: boolean }> {
  const res = await request.delete<any, { success: boolean }>(`${BASE}/${id}`)
  return res
}

/**
 * 冻结/解冻用户
 */
export async function toggleUserStatus(id: number): Promise<ToggleUserStatusResponse> {
  const res = await request.post<any, ToggleUserStatusResponse>(`${BASE}/${id}/toggle_status`)
  return res
}

/**
 * 获取用户统计数据
 * 注意：路径使用 /admin/v1/user-stats 而非 /admin/v1/users/stats，
 * 以避免与 /admin/v1/users/{id} 路由的 {id} 通配符冲突
 */
export async function getUserStats(): Promise<UserStatsResponse> {
  const res = await request.get<any, UserStatsResponse>('/admin/v1/user-stats')
  return res
}
