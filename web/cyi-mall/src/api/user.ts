/**
 * 用户相关 API
 */
import request from './index'
import type { User } from '@/types/user'

/** 获取用户信息 */
export function getUserInfo() {
  return request.get<any, User>('/user/info')
}

/** 更新用户信息 */
export function updateUserInfo(data: Partial<User>) {
  return request.put<any, void>('/user/info', data)
}
