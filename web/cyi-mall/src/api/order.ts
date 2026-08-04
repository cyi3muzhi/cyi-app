/**
 * 订单相关 API
 */
import request from './index'

/** 接口返回的订单项结构（后端 Kratos proto JSON 为 camelCase） */
export interface OrderItemDTO {
  id: number
  orderNo: string
  createdAt: string
  userName: string
  amount: number
  payment: string
  status: string
  statusText: string
  products: Array<{ name: string; spec: string; qty: number; price: number; image: string; productId: number; skuId: number }>
  logisticsCompany?: string // 物流公司（发货后）
  logisticsNo?: string // 快递单号（发货后）
  receiver?: string // 收货人
  phone?: string // 手机号
  address?: string // 完整收货地址
  payDeadline?: number // 支付截止时间（Unix 秒，仅待付款订单有值）
}

/** 待付款订单支付超时时长（秒）：下单后 30 分钟内未支付由 task 服务自动取消 */
export const PAY_TIMEOUT_SECONDS = 30 * 60

/** 创建订单 */
export function createOrder(data: {
  productId: number
  count: number
  skuId?: number
}) {
  return request.post<any, { orderId: number }>('/order/v1/create', data)
}

/** 支付订单 */
export function payOrder(orderId: number) {
  return request.post<any, { success: boolean }>('/order/v1/pay', { orderId })
}

/** 确认收货 */
export function receiveOrder(orderId: number) {
  return request.post<any, { success: boolean }>('/order/v1/receive', { orderId })
}

/** 取消订单 */
export function cancelOrder(orderId: number, reason = '') {
  return request.post<any, { success: boolean }>('/order/v1/cancel', { orderId, reason })
}

/** 获取订单列表 */
export function getOrders(params: { page?: number; pageSize?: number; status?: string; keyword?: string }) {
  return request.get<any, { items: OrderItemDTO[]; total: number }>('/order/v1/list', { params })
}

/** 获取订单详情 */
export function getOrderDetail(orderId: number) {
  return request.get<any, { order: OrderItemDTO }>(`/order/v1/${orderId}`)
}

/** 购物车结算下单（请求体为 proto snake_case，后端可解析） */
export function cartCheckout(data: {
  items: Array<{ product_id: number; sku_id: number; quantity: number }>
  address_id: number
  receiver: string
  phone: string
  address: string
  freight: number // 单位：分
  coupon_id?: number // 优惠券发放记录 ID（0=不使用优惠券）
}) {
  return request.post<any, { orderId: number }>('/order/v1/cart_checkout', data)
}
