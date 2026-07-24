export interface CouponItem {
  id: number
  name: string
  type: string
  value: number
  min_amount: number
  quantity: number
  issued_count: number
  used_count: number
  valid_start: string
  valid_end: string
  status: string
  enabled: boolean
  create_time: number
  update_time: number
  dist_records?: DistRecord[]
}

export interface DistRecord {
  id: number
  user_id: number
  source: string
  status: string
  used_time: number
  order_id: number
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
  min_amount: number
  quantity: number
  valid_start: string
  valid_end: string
}

export interface CouponUpdateData {
  name?: string
  type?: string
  value?: number
  min_amount?: number
  quantity?: number
  valid_start?: string
  valid_end?: string
  enabled?: boolean
}

export interface CouponStatsResponse {
  total_count: number
  active_count: number
  issued_count: number
  expired_count: number
}