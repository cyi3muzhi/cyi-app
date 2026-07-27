/**
 * 订单相关 API
 */
import request from './index'
import type { Order } from '@/types/order'
import type { PaginationResult } from '@/types/api'

/** 创建订单 */
export function createOrder(data: {
  productId: number
  count: number
  skuId?: number
}) {
  return request.post<any, { orderId: number }>('/order/v1/create', data)
}

/** 支付订单 */
export function payOrder(orderId: number) {
  return request.post<any, { success: boolean }>('/order/v1/pay', { orderId })
}

/** 获取订单列表 */
export function getOrders(params: { page?: number; pageSize?: number; status?: string; keyword?: string }) {
  return request.get<any, { items: Order[]; total: number }>('/order/v1/list', { params })
}

/** 获取订单详情 */
export function getOrderDetail(orderId: number) {
  return request.get<any, { order: Order }>(`/order/v1/${orderId}`)
}
