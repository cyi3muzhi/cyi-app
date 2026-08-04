/**
 * 购物车相关 API
 * 注意：后端 Kratos proto 接口，响应字段为 camelCase（productId/quantity/selected 等）
 */
import request from './index'

/** 购物车商品（后端返回，camelCase） */
export interface CartItem {
  productId: number // 商品ID
  skuId: number // SKU ID（0 表示商品默认 SKU）
  name: string // 商品名称
  price: number // 价格（单位：分）
  quantity: number // 购买数量
  image: string // 商品主图
  stock: number // 可用库存
  selected: boolean // 是否选中
  sku: string // 规格描述
  status: number // 商品状态 0-下架 1-上架
}

/** 加入购物车（同商品同规格数量自动累加） */
export function addToCart(data: { productId: number; skuId: number; quantity: number }) {
  return request.post<any, { success: boolean }>('/cart/v1/add', data)
}

/** 获取购物车 */
export function getCartList() {
  return request.get<any, { items: CartItem[] }>('/cart/v1/info')
}

/** 修改购物车商品数量 */
export function updateCartItem(data: { productId: number; skuId: number; quantity: number }) {
  return request.put<any, { success: boolean }>('/cart/v1/item', data)
}

/** 选中/取消选中单个商品 */
export function toggleCartItem(data: { productId: number; skuId: number; selected: boolean }) {
  return request.put<any, { success: boolean }>('/cart/v1/item/select', data)
}

/** 全选/取消全选 */
export function selectAllCart(selected: boolean) {
  return request.put<any, { success: boolean }>('/cart/v1/select_all', { selected })
}

/** 删除购物车商品（skuId 传 0 时删除该商品全部规格） */
export function deleteCartItem(productId: number, skuId = 0) {
  return request.delete<any, { success: boolean }>(`/cart/v1/item/${productId}`, { params: { sku_id: skuId } })
}

/** 清空购物车 */
export function clearCart() {
  return request.delete<any, { success: boolean }>('/cart/v1/clear')
}
