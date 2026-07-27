/**
 * 商品相关 API
 */
import request from './index'
import type { Product, Category } from '@/types/product'
import type { PaginationResult } from '@/types/api'

/** 获取商品列表 */
export function getProducts(params: {
  page?: number
  pageSize?: number
  categoryId?: number
  keyword?: string
  sort?: string
}) {
  return request.get<any, PaginationResult<Product>>('/products', { params })
}

/** 获取商品详情 */
export function getProductDetail(id: number) {
  return request.get<any, Product>(`/products/${id}`)
}

/** 获取分类列表 */
export function getCategories() {
  return request.get<any, Category[]>('/categories')
}

/** 获取推荐商品 */
export function getRecommendProducts() {
  return request.get<any, Product[]>('/products/recommend')
}
