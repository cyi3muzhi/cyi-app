/**
 * 商品相关 API
 */
import request from './index'
import type { CategoryNode } from '@/types/product'

/** 规格组信息 */
export interface SpecGroupInfo {
  id: number
  name: string
  values: string[]
}

/** SKU 信息 */
export interface SkuInfo {
  id: number
  sku_code: string
  price: number | string
  stock: number
  sales: number
  spec: Record<string, string>
  image: string
}

/** 商品列表项（匹配 proto ProductItem） */
export interface ProductItem {
  id: number | string
  name: string
  description: string
  price: number | string
  stock: number
  category: string
  status: string
  image: string
  originalPrice?: number | string
  sales?: number
  rating?: number
  tags?: string[]
  category_id?: number
  images?: string[]
  specs?: SpecGroupInfo[]
  skus?: SkuInfo[]
}

/** 获取商品列表 */
export function getProducts(params: {
  page?: number
  pageSize?: number
  keyword?: string
  category?: string
}) {
  return request.get<any, { items: ProductItem[]; total: number }>('/product/v1/list', { params })
}

/** 获取商品详情 */
export function getProductDetail(id: number) {
  return request.get<any, { product: ProductItem }>(`/product/v1/${id}`)
}

/** 获取分类树 */
export function getCategoryTree(parentId?: number) {
  return request.get<any, { items: CategoryNode[]; total: number }>('/category/v1/tree', {
    params: parentId !== undefined ? { parent_id: parentId } : {}
  })
}

/** 获取分类详情 */
export function getCategory(id: number) {
  return request.get<any, { item: CategoryNode }>(`/category/v1/${id}`)
}
