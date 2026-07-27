/**
 * 购物车相关 API
 */
import request from './index'

/** 加入购物车 */
export function addToCart(data: { productId: number; quantity: number }) {
  return request.post<any, { success: boolean }>('/cart/v1/add', data)
}

/** 获取购物车 */
export function getCartList() {
  return request.get<any, { items: any[] }>('/cart/v1/info')
}
