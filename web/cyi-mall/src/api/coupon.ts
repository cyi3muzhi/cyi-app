/**
 * 优惠券相关 API
 * 注意：Kratos protojson 将 int64 序列化为字符串，这里统一转为 number
 */
import request from './index'
import type { AvailableCouponItem, MyCouponItem, UsableCouponItem } from '@/types/coupon'

/** 将字符串数字安全转为数字（空值/非法值转 0） */
function toNum(v: any): number {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

/**
 * 领取优惠券
 * @param couponId 优惠券 ID
 * @param source 领取来源（如 user_center / coupon_center）
 */
export function claimCoupon(couponId: number, source = 'coupon_center') {
  return request.post<any, { recordId: number }>(`/api/v1/coupons/${couponId}/claim`, { source }).then(res => ({
    recordId: toNum(res.recordId)
  }))
}

/**
 * 我的优惠券列表
 * @param params.status 0=全部 1=未使用 2=已使用 3=已过期
 */
export async function listMyCoupons(params: { status?: number; page?: number; pageSize?: number }) {
  const res = await request.get<any, { items: MyCouponItem[]; total: number }>('/api/v1/coupons/my', { params })
  return {
    items: (res.items || []).map<MyCouponItem>(i => ({
      ...i,
      recordId: toNum(i.recordId),
      couponId: toNum(i.couponId),
      type: toNum(i.type) as 1 | 2 | 3,
      value: toNum(i.value),
      minAmount: toNum(i.minAmount),
      recordStatus: toNum(i.recordStatus),
      usedTime: toNum(i.usedTime),
      orderId: toNum(i.orderId)
    })),
    total: toNum(res.total)
  }
}

/**
 * 领券中心：可领取优惠券列表
 */
export async function listAvailableCoupons(params: { page?: number; pageSize?: number }) {
  const res = await request.get<any, { items: AvailableCouponItem[]; total: number }>('/api/v1/coupons/available', { params })
  return {
    items: (res.items || []).map<AvailableCouponItem>(i => ({
      ...i,
      couponId: toNum(i.couponId),
      type: toNum(i.type) as 1 | 2 | 3,
      value: toNum(i.value),
      minAmount: toNum(i.minAmount),
      validType: toNum(i.validType) as 1 | 2 | 3,
      validDays: toNum(i.validDays),
      validStart: toNum(i.validStart),
      validEnd: toNum(i.validEnd),
      remainCount: toNum(i.remainCount),
      limitPerUser: toNum(i.limitPerUser)
    })),
    total: toNum(res.total)
  }
}

/**
 * 结算可用优惠券（按订单商品金额筛选）
 * @param amount 订单商品金额（分）
 */
export async function listUsableCoupons(amount: number) {
  const res = await request.get<any, { items: UsableCouponItem[] }>('/api/v1/coupons/usable', { params: { amount } })
  return {
    items: (res.items || []).map<UsableCouponItem>(i => ({
      ...i,
      recordId: toNum(i.recordId),
      couponId: toNum(i.couponId),
      type: toNum(i.type) as 1 | 2 | 3,
      value: toNum(i.value),
      minAmount: toNum(i.minAmount),
      discount: toNum(i.discount),
      gapAmount: toNum(i.gapAmount)
    }))
  }
}
