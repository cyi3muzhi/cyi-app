/**
 * 购物车相关 API
 */
import request from './index'
import type { CartItem } from '@/types/cart'

/** 获取购物车列表 */
export function getCartList() {
  return request.get<any, CartItem[]>('/cart')
}

/** 添加商品到购物车 */
export function addToCart(data: { productId: number; quantity: number; sku?: string }) {
  return request.post<any, void>('/cart/add', data)
}

/** 更新购物车商品数量 */
export function updateCartQuantity(id: number, quantity: number) {
  return request.put<any, void>(`/cart/${id}`, { quantity })
}

/** 删除购物车商品 */
export function removeFromCart(id: number) {
  return request.delete<any, void>(`/cart/${id}`)
}
