/**
 * 悦选商城管理后台 - 仪表盘 API
 */
import request from './index'
import type {
  DashboardStatsResponse,
  OrderTrendResponse,
  RevenueCategoriesResponse
} from '../types/dashboard'

/** 仪表盘 API 基础路径 */
const BASE = '/admin/v1/dashboard'

/**
 * 获取仪表盘统计数据（今日订单/营收、总用户数/总商品数等）
 */
export async function getDashboardStats(): Promise<DashboardStatsResponse> {
  const res = await request.get<any, DashboardStatsResponse>(`${BASE}/stats`)
  return res
}

/**
 * 获取近14日订单趋势
 * TODO: 目前后端返回 mock 数据，待 order service 提供按日统计 RPC 后接入真实数据
 */
export async function getOrderTrend(): Promise<OrderTrendResponse> {
  const res = await request.get<any, OrderTrendResponse>(`${BASE}/order_trend`)
  return res
}

/**
 * 获取营收分类占比
 * TODO: 目前后端返回 mock 数据，待 order service 提供分类营收 RPC 后接入真实数据
 */
export async function getRevenueCategories(): Promise<RevenueCategoriesResponse> {
  const res = await request.get<any, RevenueCategoriesResponse>(`${BASE}/revenue_categories`)
  return res
}
