import request from './index'
import type {
  CouponListParams, CouponListResponse,
  CouponCreateData, CouponUpdateData, CouponStatsResponse
} from '../types/coupon'

export function listCoupons(params: CouponListParams): Promise<CouponListResponse> {
  return request.get('/admin/v1/coupons', { params })
}

export function getCoupon(id: number): Promise<{ coupon: import('../types/coupon').CouponItem }> {
  return request.get(`/admin/v1/coupons/${id}`)
}

export function createCoupon(data: CouponCreateData): Promise<{ coupon_id: number }> {
  return request.post('/admin/v1/coupons', data)
}

export function updateCoupon(id: number, data: CouponUpdateData): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/coupons/${id}`, data)
}

export function deleteCoupon(id: number): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/coupons/${id}`)
}

export function toggleCouponStatus(id: number, action: number): Promise<{ success: boolean }> {
  return request.post(`/admin/v1/coupons/${id}/toggle`, { action })
}

export function getCouponStats(): Promise<CouponStatsResponse> {
  // silent：统计卡片为辅助数据，失败时静默降级，避免与列表接口同时失败弹多条重复提示
  return request.get('/admin/v1/coupons/stats', { silent: true })
}

export function listCouponRecords(couponId: number, params?: { page?: number; page_size?: number }): Promise<{ items: import('../types/coupon').DistRecord[]; total: number }> {
  return request.get(`/admin/v1/coupons/${couponId}/records`, { params })
}