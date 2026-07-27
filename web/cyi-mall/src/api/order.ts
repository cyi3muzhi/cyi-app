/**
 * 订单相关 API
 */
import request from './index'
import type { Order } from '@/types/order'
import type { PaginationResult } from '@/types/api'

/** 获取订单列表 */
export function getOrders(params: { page?: number; pageSize?: number; status?: string }) {
  return request.get<any, PaginationResult<Order>>('/orders', { params })
}

/** 获取订单详情 */
export function getOrderDetail(id: number) {
  return request.get<any, Order>(`/orders/${id}`)
}

/** 创建订单 */
export function createOrder(data: {
  items: { productId: number; quantity: number }[]
  addressId: number
  couponId?: number
  remark?: string
}) {
  return request.post<any, { orderId: number }>('/orders', data)
}

/** 取消订单 */
export function cancelOrder(id: number) {
  return request.put<any, void>(`/orders/${id}/cancel`)
}
