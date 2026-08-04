/**
 * 用户相关 API（个人中心）
 */
import request from './index'
import type { Address, AddressPayload, User, UserCenter } from '@/types/user'

/** 获取个人信息 */
export function getMyInfo() {
  return request.get<any, User>('/user/v1/my_info')
}

/** 更新个人信息 */
export function updateUserInfo(data: Partial<Pick<User, 'nickname' | 'avatar' | 'email' | 'gender' | 'birthday'>>) {
  return request.put<any, { success: boolean }>('/user/v1/update', data)
}

/** 修改密码 */
export function changePassword(data: { old_password: string; new_password: string }) {
  return request.post<any, { success: boolean }>('/user/v1/change_password', data)
}

/** 个人中心概览（用户信息 + 订单/优惠券统计） */
export function getUserCenter() {
  return request.get<any, UserCenter>('/user/v1/center')
}

/** 收货地址列表 */
export function listAddress() {
  return request.get<any, { items: Address[] }>('/user/v1/address/list')
}

/** 新增收货地址 */
export function createAddress(data: AddressPayload) {
  return request.post<any, { id: number }>('/user/v1/address/create', data)
}

/** 更新收货地址 */
export function updateAddress(data: AddressPayload) {
  return request.put<any, { success: boolean }>('/user/v1/address/update', data)
}

/** 删除收货地址 */
export function deleteAddress(id: number) {
  return request.delete<any, { success: boolean }>(`/user/v1/address/${id}`)
}

/** 设置默认收货地址 */
export function setDefaultAddress(id: number) {
  return request.post<any, { success: boolean }>(`/user/v1/address/${id}/default`)
}
