/** 购物车商品 */
export interface CartItem {
  id: number
  productId: number
  name: string
  image: string
  price: number
  quantity: number
  selected: boolean
  sku: string
  stock: number
}
