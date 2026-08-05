/**
 * cyi商城管理后台 - 消息中心 API
 */
import request from './index'

/** 消息项 */
export interface MessageItem {
  id: number
  title: string // 标题
  content: string // 内容
  msgType: number // 消息类型 1=订单 2=系统 3=营销
  createdAt: number // 创建时间（Unix 秒）
  createdBy: number // 创建人（管理员ID）
  targetType: number // 发送范围 0=全体用户 1=指定用户
  targetUserIds: string // 指定用户ID列表（JSON 数组字符串）
  status: number // 1=启用 2=下线
}

/** 查询参数 */
export interface MessageQuery {
  page?: number
  page_size?: number
  msg_type?: number
  status?: number
  keyword?: string
}

/** 发送消息参数 */
export interface SendMessagePayload {
  title: string
  content: string
  msg_type: number // 1=订单 2=系统 3=营销
  target_type: number // 0=全体用户 1=指定用户
  target_user_ids?: number[]
}

/** 消息类型标签 */
export const MSG_TYPE_LABEL: Record<number, string> = {
  1: '订单',
  2: '系统',
  3: '营销'
}

/** 消息类型 tag 颜色 */
export const MSG_TYPE_TAG: Record<number, string> = {
  1: 'blue',
  2: 'green',
  3: 'orange'
}

/**
 * 归一化消息项：Kratos protojson 输出 camelCase，统一转为 snake_case
 */
function normalizeMessage(item: any): MessageItem {
  return {
    id: Number(item.id ?? 0),
    title: item.title ?? '',
    content: item.content ?? '',
    msgType: Number(item.msgType ?? item.msg_type ?? 0),
    createdAt: Number(item.createdAt ?? item.created_at ?? 0),
    createdBy: Number(item.createdBy ?? item.created_by ?? 0),
    targetType: Number(item.targetType ?? item.target_type ?? 0),
    targetUserIds: item.targetUserIds ?? item.target_user_ids ?? '',
    status: Number(item.status ?? 1)
  }
}

/** 消息列表（分页筛选） */
export async function getMessages(params: MessageQuery = {}) {
  const res = await request.get<any, { items: any[]; total: number }>('/admin/v1/messages', { params })
  return {
    items: (res.items || []).map(normalizeMessage),
    total: res.total || 0
  }
}

/** 发送消息（全体/指定用户） */
export function sendMessage(data: SendMessagePayload) {
  return request.post<any, { id: number }>('/admin/v1/messages', data)
}

/** 下线/删除消息 */
export function deleteMessage(id: number) {
  return request.delete<any, { success: boolean }>(`/admin/v1/messages/${id}`)
}
