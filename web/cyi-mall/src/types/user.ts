/** 用户 */
export interface User {
  id: number
  username?: string // 登录账号
  nickname: string
  avatar: string
  phone: string
  email: string
  gender?: number // 0=未知 1=男 2=女
  birthday?: string // YYYY-MM-DD
  memberLevel: string
  level?: string // 会员等级 regular/gold/diamond
  growthValue?: number // 成长值
  totalSpent: number // 累计消费（分）
  points: number // 积分
  couponCount: number // 优惠券数
  createdAt: string // 注册时间
}

/** 个人中心统计 */
export interface CenterStats {
  totalOrders: number // 总订单数
  pendingPay: number // 待付款
  pendingShip: number // 待发货
  pendingReceive: number // 待收货
  pendingReview: number // 待评价（已完成）
  refundCount: number // 退款/售后
  couponCount: number // 可用优惠券数
  totalSpent: number // 累计消费（分）
}

/** 个人中心概览（用户信息 + 统计） */
export interface UserCenter {
  user: User
  stats: CenterStats
}

/** 收货地址 */
export interface Address {
  id: number
  receiver: string // 收货人姓名
  phone: string // 收货人手机号
  province: string // 省份
  city: string // 城市
  district: string // 区县
  detail: string // 详细地址
  isDefault: number // 0=否 1=是
  createTime: number // 创建时间（Unix 秒）
}

/** 新增/更新地址参数 */
export interface AddressPayload {
  id?: number
  receiver: string
  phone: string
  province?: string
  city?: string
  district?: string
  detail: string
  isDefault?: number
}
