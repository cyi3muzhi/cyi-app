/** 订单 */
export interface Order {
  id: number
  orderNo: string
  status: OrderStatus
  totalAmount: number
  actualAmount: number
  quantity: number
  items: OrderItem[]
  createdAt: string
  payDeadline?: number // 支付截止时间（Unix 秒，仅待付款订单有值）
  address: string
  receiver: string
  phone: string
}

/** 订单商品项 */
export interface OrderItem {
  productId: number
  skuId: number
  name: string
  image: string
  price: number
  quantity: number
}

/** 订单状态 */
export type OrderStatus =
  | 'pending_payment'   // 待付款
  | 'pending_ship'      // 待发货
  | 'pending_receive'   // 待收货
  | 'pending_review'    // 待评价
  | 'completed'         // 已完成
  | 'cancelled'         // 已取消
  | 'refund'            // 退款/售后
