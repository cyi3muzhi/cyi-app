/**
 * 悦选商城管理后台 - 商品管理 API
 */
import request from './index'

/** 规格组接口（API 返回格式） */
export interface SpecGroup {
  id: number
  name: string
  values: string[]
}

/** SKU 接口（API 返回格式） */
export interface SkuInfo {
  id: number
  sku_code: string
  price: number
  stock: number
  sales: number
  spec: Record<string, string>
}

/** 商品列表项（API 返回格式） */
export interface ProductItem {
  id: number
  name: string
  category: string
  price: number
  stock: number
  sales: number
  status: string
  image: string
  sku: string
  specs: SpecGroup[]
  skus: SkuInfo[]
}

/** 商品列表查询参数 */
export interface ProductListParams {
  page?: number
  page_size?: number
  keyword?: string
  category?: string
  status?: string
}

/** 规格组提交格式 */
export interface SpecGroupInput {
  name: string
  values: string[]
}

/** SKU 提交格式 */
export interface SkuInput {
  sku_code: string
  price: number
  stock: number
  sales: number
  spec: Record<string, string>
}

/** 创建商品请求体 */
export interface CreateProductData {
  name: string
  category: string
  price: number
  stock: number
  description?: string
  image?: string
  sku?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/** 更新商品请求体 */
export interface UpdateProductData {
  name?: string
  category?: string
  price?: number
  stock?: number
  description?: string
  image?: string
  sku?: string
  status?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/**
 * 获取分页商品列表
 * @param params 查询参数（分页、搜索关键字、分类、状态筛选）
 */
export function listProducts(params: ProductListParams): Promise<{ items: ProductItem[]; total: number }> {
  return request.get('/admin/v1/products', { params })
}

/**
 * 创建商品
 * @param data 商品信息
 */
export function createProduct(data: CreateProductData): Promise<{ id: number }> {
  return request.post('/admin/v1/products', data)
}

/**
 * 更新商品
 * @param id 商品 ID
 * @param data 待更新字段
 */
export function updateProduct(id: number, data: UpdateProductData): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/products/${id}`, data)
}

/**
 * 删除商品
 * @param id 商品 ID
 */
export function deleteProduct(id: number): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/products/${id}`)
}

/**
 * 获取商品详情
 * @param id 商品 ID
 */
export function getProduct(id: number): Promise<{ product: ProductItem }> {
  return request.get(`/admin/v1/products/${id}`)
}

/**
 * 批量操作商品
 * @param action 操作类型："上架" | "下架" | "删除"
 * @param ids 商品 ID 列表
 */
export function batchProductOperation(action: string, ids: number[]): Promise<{ success: boolean; affected_count: number }> {
  return request.post('/admin/v1/products/batch', { action, ids })
}
