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
  return request.get('/admin/v1/coupons/stats')
}