/**
 * 用户相关 API
 */
import request from './index'

/** 获取个人信息 */
export function getMyInfo() {
  return request.get<any, { id: number; username: string }>('/user/v1/my_info')
}
