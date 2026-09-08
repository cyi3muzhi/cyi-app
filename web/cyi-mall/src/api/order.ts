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

/**
 * 生成客户端幂等键（32 位无横线 UUID）
 * 每次下单/结算动作生成一次，重复提交（双击/网络重试）时复用同一键，后端据此返回同一订单，防止重复下单
 */
export function generateIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID().replace(/-/g, '')
  }
  // 兼容不支持 crypto.randomUUID 的环境：时间戳 + 两段随机数
  return (
    Date.now().toString(16) +
    Math.random().toString(16).slice(2, 10) +
    Math.random().toString(16).slice(2, 10)
  )
}

/** 创建订单 */
export function createOrder(data: {
  productId: number
  count: number
  skuId?: number
  idempotencyKey?: string // 客户端幂等键（防重复下单）
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
  freight: number // 单位：分（服务端会按规则重算，前端传值忽略）
  coupon_id?: number // 优惠券发放记录 ID（0=不使用优惠券）
  idempotency_key?: string // 客户端幂等键（防重复下单）
}) {
  return request.post<any, { orderId: number }>('/order/v1/cart_checkout', data)
}

/** 退款记录项结构 */
export interface RefundItemDTO {
  refundId: number // 退款单ID
  orderId: number // 订单ID
  orderNo: string // 订单号
  refundAmount: number // 退款金额（单位：分）
  reason: string // 退款原因
  status: number // 状态: 1-待审核 2-已通过 3-已拒绝
  statusText: string // 状态文本
  rejectReason: string // 拒绝原因
  createdAt: string // 申请时间（格式化字符串）
}

/** 申请退款（仅已支付/待发货订单可申请） */
export function applyRefund(orderId: number, reason: string) {
  return request.post<any, { refundId: number }>('/order/v1/refund/apply', { orderId, reason })
}

/** 我的退款记录 */
export function getRefunds(params: { page?: number; pageSize?: number; status?: string }) {
  return request.get<any, { items: RefundItemDTO[]; total: number }>('/order/v1/refunds', { params })
}

/** 物流轨迹节点 */
export interface LogisticsTraceDTO {
  time: number // 节点时间（Unix 秒）
  timeText: string // 节点时间（格式化字符串）
  desc: string // 节点描述
  status: number // 节点状态: 1-已发货 2-运输中 3-派送中 4-已签收
}

/** 物流信息（发货后展示） */
export interface LogisticsDTO {
  orderId: number
  logisticsCompany: string // 物流公司
  logisticsNo: string // 快递单号
  shipTime: number // 发货时间（Unix 秒，0=未发货）
  shipTimeText: string // 发货时间（格式化字符串）
  traces: LogisticsTraceDTO[] // 轨迹节点（按时间升序）
}

/** 查询订单物流轨迹 */
export function getLogistics(orderId: number) {
  return request.get<any, LogisticsDTO>(`/order/v1/${orderId}/logistics`)
}
