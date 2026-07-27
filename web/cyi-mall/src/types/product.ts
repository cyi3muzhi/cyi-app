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
