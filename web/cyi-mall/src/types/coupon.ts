/** 优惠券类型：1-立减 2-折扣 3-满减 */
export type CouponType = 1 | 2 | 3

/** 有效期类型：1-固定日期 2-领取后N天 3-永久 */
export type ValidType = 1 | 2 | 3

/** 我的优惠券项（GET /api/v1/coupons/my） */
export interface MyCouponItem {
  recordId: number // 发放记录 ID
  couponId: number // 优惠券 ID
  name: string // 名称
  type: CouponType // 类型：1-立减 2-折扣 3-满减
  value: number // 面值（分）或折扣百分比（1-99）
  minAmount: number // 使用门槛（分）
  validStart: string // 有效期开始（YYYY-MM-DD HH:mm:ss）
  validEnd: string // 有效期结束（YYYY-MM-DD HH:mm:ss）
  recordStatus: number // 0=未使用 1=已使用 2=已过期
  usedTime: number // 使用时间（Unix 秒）
  orderId: number // 核销订单 ID
  source: string // 领取来源
  createTime: string // 领取时间（YYYY-MM-DD HH:mm:ss）
}

/** 可领取优惠券项（领券中心 GET /api/v1/coupons/available） */
export interface AvailableCouponItem {
  couponId: number // 优惠券 ID
  name: string // 名称
  type: CouponType // 类型：1-立减 2-折扣 3-满减
  value: number // 面值（分）或折扣百分比（1-99）
  minAmount: number // 使用门槛（分）
  validType: ValidType // 有效期类型
  validDays: number // 领取后有效天数
  validStart: number // 有效期开始（Unix 秒）
  validEnd: number // 有效期结束（Unix 秒），0=永久
  remainCount: number // 剩余可领数量
  claimed: boolean // 当前用户是否已领取
  limitPerUser: number // 每人限领数量，0=不限制
}

/** 结算可用券项（GET /api/v1/coupons/usable） */
export interface UsableCouponItem {
  recordId: number // 发放记录 ID（下单时传给 cartCheckout）
  couponId: number // 优惠券 ID
  name: string // 名称
  type: CouponType // 类型：1-立减 2-折扣 3-满减
  value: number // 面值（分）或折扣百分比
  minAmount: number // 使用门槛（分）
  validEnd: string // 有效期截止（YYYY-MM-DD HH:mm:ss）
  usable: boolean // 是否可用
  discount: number // 可用时预估优惠金额（分）
  gapAmount: number // 不可用时还差多少金额可用（分）
}
