/** 优惠券模块类型定义 */

/** 字段为 camelCase，与 protobuf JSON 序列化一致 */

export interface CouponItem {
  id: number
  name: string
  type: string
  value: number
  minAmount: number
  quantity: number
  issuedCount: number
  usedCount: number
  validStart: string
  validEnd: string
  status: string
  enabled: boolean
  createTime: number
  updateTime: number
  limitPerUser: number
  distRecords?: DistRecord[]
}

export interface DistRecord {
  id: number
  userId: number
  source: string
  status: string
  usedTime: number
  orderId: number
  time: string
  user: string
  createTime: number
}

export interface CouponListParams {
  page?: number
  page_size?: number
  status?: string
  keyword?: string
  date_from?: string
  date_to?: string
}

export interface CouponListResponse {
  items: CouponItem[]
  total: number
}

export interface CouponCreateData {
  name: string
  type: string
  value: number
  minAmount: number
  quantity: number
  limitPerUser: number
  validStart: string
  validEnd: string
}

export interface CouponUpdateData {
  name?: string
  type?: string
  value?: number
  minAmount?: number
  quantity?: number
  limitPerUser?: number
  validStart?: string
  validEnd?: string
  enabled?: boolean
}

export interface CouponStatsResponse {
  totalCount: number
  activeCount: number
  issuedCount: number
  expiredCount: number
}
