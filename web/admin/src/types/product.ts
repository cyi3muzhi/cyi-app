// web/admin/src/types/product.ts
// 商品管理相关 TypeScript 类型定义

/** 规格组（API 返回格式） */
export interface SpecGroup {
  id: number
  name: string
  values: string[]
}

/** SKU（API 返回格式） */
export interface SkuInfo {
  id: number
  sku_code: string
  price: number    // 分
  stock: number
  sales: number
  spec: Record<string, string>  // {"颜色":"红色","尺寸":"M"}
}

/** 商品列表项 */
export interface ProductItem {
  id: number
  name: string
  categoryId: number
  price: number
  stock: number
  sales: number
  status: string   // "上架" | "下架"
  image: string
  description: string
  specs: SpecGroup[]
  skus: SkuInfo[]
}

/** 商品列表查询参数 */
export interface ProductListParams {
  page?: number
  page_size?: number
  keyword?: string
  category_id?: number
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
  price: number     // 分
  stock: number     // 库存
  spec: Record<string, string>
}

/** 创建商品请求体 */
export interface CreateProductData {
  name: string
  category_id: number
  price: number
  description?: string
  image?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/** 更新商品请求体 */
export interface UpdateProductData {
  name?: string
  category_id?: number
  price?: number
  description?: string
  image?: string
  status?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/** 批量操作请求 */
export interface BatchOperationData {
  action: '上架' | '下架' | '删除'
  ids: number[]
}

/** 商品列表响应 */
export interface ProductListResponse {
  items: ProductItem[]
  total: number
}
