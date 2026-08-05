/** cyi商城管理后台 - 订单管理类型定义 */

/** 订单商品明细 */
export interface OrderProduct {
  name: string
  spec: string
  qty: number
  price: number // 分
  image?: string // 商品图片
}

/** 订单列表查询参数 */
export interface OrderListParams {
  page?: number
  page_size?: number
  status?: string      // 状态筛选
  keyword?: string     // 订单号搜索
  date_from?: string   // 开始日期 "2006-01-02"
  date_to?: string     // 结束日期 "2006-01-02"
}

/** 订单列表项（字段为 camelCase，与 protobuf JSON 序列化一致） */
export interface OrderItem {
  id: number
  orderNo: string
  createdAt: string   // "2006-01-02 15:04:05"
  userName: string
  amount: number       // 实际支付金额（分）
  payment: string      // 支付方式
  status: string       // pending|paid|shipped|completed|cancelled|refunding|refunded
  statusText: string  // 待支付|已支付|已发货|已完成|已取消|退款中|已退款
  products: OrderProduct[]
  logisticsCompany?: string // 物流公司（发货后）
  logisticsNo?: string // 快递单号（发货后）
  receiver?: string // 收货人
  phone?: string // 手机号
  address?: string // 完整收货地址
}

/** 订单列表响应 */
export interface OrderListResponse {
  items: OrderItem[]
  total: number
}

/** 订单详情响应 */
export interface OrderDetailResponse {
  order: OrderItem
}

/** 发货请求 */
export interface ShipOrderData {
  logistics_company: string
  logistics_no: string
}

/** 取消订单请求 */
export interface CancelOrderData {
  reason: string
}

/** 导出订单参数 */
export interface ExportOrderParams {
  status?: string
  date_from?: string
  date_to?: string
}

/** 导出订单响应（投递异步任务后返回） */
export interface ExportOrderResponse {
  taskId: string      // 异步任务ID，用于轮询状态
  status: ExportStatus // 当前状态（初始为 PENDING）
}

/** 导出任务状态查询响应 */
export interface ExportTaskStatusResponse {
  taskId: string
  status: ExportStatus
  downloadUrl?: string   // 完成后的下载链接
  createdAt: number      // 任务创建时间（Unix秒）
  finishedAt?: number    // 任务完成时间（Unix秒）
  error?: string         // 失败原因
  totalCount?: number    // 总订单数
  processedCount?: number // 已处理订单数
}

/** 导出任务状态枚举 */
export type ExportStatus = 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'FAILED'

/** 订单统计响应（顶部统计卡片数据） */
export interface OrderStatsResponse {
  totalCount: number    // 总订单数
  pendingCount: number  // 待处理数（待支付+待发货）
  todayCount: number    // 今日订单数
  refundCount: number   // 退款/售后数（退款中+已退款）
}
