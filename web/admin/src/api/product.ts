// web/admin/src/api/product.ts
// 商品管理 API 方法

import request from './index'
import type {
  ProductListParams, ProductListResponse,
  CreateProductData, UpdateProductData, BatchOperationData
} from '../types/product'

/** 获取分页商品列表 */
export function listProducts(params: ProductListParams): Promise<ProductListResponse> {
  return request.get('/admin/v1/products', { params })
}

/** 创建商品 */
export function createProduct(data: CreateProductData): Promise<{ id: number }> {
  return request.post('/admin/v1/products', data)
}

/** 更新商品 */
export function updateProduct(id: number, data: UpdateProductData): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/products/${id}`, data)
}

/** 删除商品 */
export function deleteProduct(id: number): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/products/${id}`)
}

/** 获取商品详情 */
export function getProduct(id: number): Promise<{ product: import('../types/product').ProductItem }> {
  return request.get(`/admin/v1/products/${id}`)
}

/** 批量操作 */
export function batchProductOperation(action: string, ids: number[]): Promise<{ success: boolean; affected_count: number }> {
  return request.post('/admin/v1/products/batch', { action, ids })
}
