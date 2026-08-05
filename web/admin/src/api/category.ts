/**
 * cyi商城管理后台 - 商品分类 API
 */
import request from './index'

/** 分类节点 */
export interface CategoryItem {
  id: number
  name: string
  slug: string
  icon: string
  sort: number
  status: boolean
  showcase: boolean
  count: number
  image: string
  seo_title: string
  seo_keywords: string
  seo_desc: string
  parent_id: number
  created_at: number
  children: CategoryItem[]
}

/** 获取分类树 */
export function listCategoryTree(): Promise<{ items: CategoryItem[]; total: number }> {
  return request.get('/admin/v1/categories/tree')
}

/** 查询子分类 */
export function listChildCategories(parentId: number): Promise<{ items: CategoryItem[]; total: number }> {
  return request.get('/admin/v1/categories', { params: { parent_id: parentId } })
}

/** 获取单个分类 */
export function getCategory(id: number): Promise<{ item: CategoryItem }> {
  return request.get(`/admin/v1/categories/${id}`)
}

/** 创建分类 */
export function createCategory(data: {
  name: string
  slug?: string
  icon?: string
  sort?: number
  status?: boolean
  showcase?: boolean
  image?: string
  seo_title?: string
  seo_keywords?: string
  seo_desc?: string
  parent_id?: number | null
}): Promise<{ id: number }> {
  return request.post('/admin/v1/categories', data)
}

/** 更新分类 */
export function updateCategory(id: number, data: {
  name?: string
  slug?: string
  icon?: string
  sort?: number
  status?: boolean
  showcase?: boolean
  image?: string
  seo_title?: string
  seo_keywords?: string
  seo_desc?: string
  parent_id?: number | null
}): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/categories/${id}`, data)
}

/** 删除分类 */
export function deleteCategory(id: number): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/categories/${id}`)
}

/** 更新分类排序 */
export function updateCategorySort(id: number, sort: number): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/categories/${id}/sort`, { sort })
}

/** 更新分类状态 */
export function updateCategoryStatus(id: number, status: boolean): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/categories/${id}/status`, { status })
}

/** 切换分类首页展示 */
export function toggleCategoryShowcase(id: number, showcase: boolean): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/categories/${id}/showcase`, { showcase })
}

/** 批量操作分类 */
export function batchCategoryOperation(action: string, ids: number[]): Promise<{ success: boolean; affected_count: number }> {
  return request.post('/admin/v1/categories/batch', { action, ids })
}
