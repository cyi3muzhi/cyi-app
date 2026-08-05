/**
 * cyi商城管理后台 - 订单管理 API
 */
import request from './index'
import type { OrderListParams, OrderListResponse, OrderDetailResponse, ShipOrderData, ExportOrderParams, ExportOrderResponse, ExportTaskStatusResponse, OrderStatsResponse, RefundListParams, RefundListResponse } from '../types/order'

/** 订单管理 API 基础路径 */
const BASE = '/admin/v1/orders'

/** 退款/售后 API 基础路径 */
const REFUND_BASE = '/admin/v1/refunds'

/**
 * 获取订单列表
 */
export async function listOrder(params: OrderListParams): Promise<OrderListResponse> {
  const res = await request.get<any, OrderListResponse>(BASE, { params })
  return res
}

/**
 * 获取订单详情
 */
export async function getOrder(orderId: number): Promise<OrderDetailResponse> {
  const res = await request.get<any, OrderDetailResponse>(`${BASE}/${orderId}`)
  return res
}

/**
 * 发货
 */
export async function shipOrder(orderId: number, data: ShipOrderData): Promise<{ success: boolean }> {
  const res = await request.post<any, { success: boolean }>(`${BASE}/${orderId}/ship`, data)
  return res
}

/**
 * 取消订单
 */
export async function cancelOrder(orderId: number, reason: string): Promise<{ success: boolean }> {
  const res = await request.post<any, { success: boolean }>(`${BASE}/${orderId}/cancel`, { reason })
  return res
}

/**
 * 导出订单 - 投递异步任务，返回任务ID
 */
export async function exportOrder(params: ExportOrderParams): Promise<ExportOrderResponse> {
  const res = await request.get<any, ExportOrderResponse>(`${BASE}/export`, { params })
  return res
}

/**
 * 查询导出任务状态
 */
export async function getExportTaskStatus(taskId: string): Promise<ExportTaskStatusResponse> {
  const res = await request.get<any, ExportTaskStatusResponse>(`${BASE}/export/${taskId}`)
  return res
}

/**
 * 订单统计（顶部统计卡片数据）
 */
export async function orderStats(): Promise<OrderStatsResponse> {
  const res = await request.get<any, OrderStatsResponse>(`${BASE}/stats`)
  return res
}

/**
 * 获取退款/售后单列表
 */
export async function listRefunds(params: RefundListParams): Promise<RefundListResponse> {
  const res = await request.get<any, RefundListResponse>(REFUND_BASE, { params })
  return res
}

/**
 * 审核退款（通过/拒绝）
 * @param refundId - 退款单ID
 * @param approve - true=通过（退款）；false=拒绝
 * @param rejectReason - 拒绝原因（拒绝时必填）
 */
export async function auditRefund(refundId: number, approve: boolean, rejectReason = ''): Promise<{ success: boolean }> {
  const res = await request.post<any, { success: boolean }>(`${REFUND_BASE}/${refundId}/audit`, {
    refund_id: refundId,
    approve,
    reject_reason: rejectReason
  })
  return res
}
