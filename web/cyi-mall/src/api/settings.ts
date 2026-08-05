/**
 * 系统设置相关 API（用户端可见的公开设置）
 */
import request from './index'

/** 公开系统设置 */
export interface PublicSettings {
  shopName: string
  shopDescription: string
  contactPhone: string
  contactEmail: string
  icpRecord: string
  hotKeywords: string[]  // 热门搜索词（搜索框联想展示）
}

/** 获取公开系统设置 */
export function getPublicSettings() {
  return request.get<any, PublicSettings>('/settings/v1/public')
}
