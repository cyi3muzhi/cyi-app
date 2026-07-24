/**
 * 悦选商城管理后台 - 系统设置 API
 */
import request from './index'

/** 系统设置接口定义（camelCase 匹配 protobuf JSON 序列化） */
export interface SystemSettings {
  shopName: string
  shopLogo: string
  shopDescription: string
  contactPhone: string
  contactEmail: string
  icpRecord: string
  autoConfirmDays: number
  autoCloseMinutes: number
  refundTimeoutDays: number
  stockAlertThreshold: number
  newProductOnline: boolean
  productApproval: boolean
  newOrderNotify: boolean
  refundNotify: boolean
  userRegisterNotify: boolean
  stockAlertNotify: boolean
  loginAttemptLimit: boolean
  loginCaptcha: boolean
  sessionTimeoutMinutes: number
  enabledPayments: string[]
  defaultPayment: string
  freeShippingThreshold: number
  defaultShippingCompany: string
}

/** 获取系统设置 */
export function getSettings(): Promise<SystemSettings> {
  return request.get('/admin/v1/settings')
}

/** 更新系统设置 */
export function updateSettings(data: Partial<SystemSettings>): Promise<{ success: boolean }> {
  return request.put('/admin/v1/settings', data)
}
