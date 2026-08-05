/** cyi商城管理后台 - 仪表盘类型定义 */

/** 仪表盘统计数据响应 */
export interface DashboardStatsResponse {
  todayOrders: number
  todayRevenue: string
  totalUsers: number
  totalProducts: number
  ordersGrowth: number
  revenueGrowth: number
  usersGrowth: number
  productsGrowth: number
}

/** 订单趋势项 */
export interface TrendItem {
  date: string
  value: number
}

/** 订单趋势响应 */
export interface OrderTrendResponse {
  items: TrendItem[]
}

/** 营收分类 */
export interface RevenueCategory {
  name: string
  percentage: number
  color: string
}

/** 营收分类占比响应 */
export interface RevenueCategoriesResponse {
  categories: RevenueCategory[]
}
