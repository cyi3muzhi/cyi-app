/** 商品 */
export interface Product {
  id: number
  name: string
  description: string
  price: number
  originalPrice: number
  image: string
  images: string[]
  sales: number
  rating: number
  categoryId: number
  categoryName: string
  stock: number
  status: number
  tags: string[]
  createdAt: string
}

/** 商品分类 */
export interface Category {
  id: number
  name: string
  icon: string
}

/** 分类树节点 */
export interface CategoryNode {
  id: number
  name: string
  slug: string
  icon: string
  sort: number
  status: boolean
  showcase: boolean
  count: number
  image: string
  parent_id: number
  children: CategoryNode[]
}
