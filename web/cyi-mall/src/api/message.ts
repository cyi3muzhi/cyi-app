/**
 * 消息中心相关 API（用户端）
 */
import request from './index'

/** 消息项 */
export interface MessageItem {
  id: number
  title: string // 标题
  content: string // 内容
  msgType: number // 消息类型 1=订单 2=系统 3=营销
  isRead: number // 是否已读 0=未读 1=已读
  createdAt: number // 创建时间（Unix 秒）
  readAt: number // 已读时间（Unix 秒）
  conversationId: number // 关联会话ID（IM 客服离线提醒，0=普通通知）
}

/** 消息列表响应 */
export interface MessageListResult {
  items: MessageItem[]
  total: number
  unreadCount: number
}

/** 消息类型标签（展示用） */
export const MSG_TYPE_LABEL: Record<number, string> = {
  1: '订单',
  2: '系统',
  3: '营销'
}

/**
 * 归一化消息项：Kratos protojson 输出 camelCase，统一转为组件可用结构
 */
function normalizeMessage(item: any): MessageItem {
  return {
    id: Number(item.id ?? 0),
    title: item.title ?? '',
    content: item.content ?? '',
    msgType: Number(item.msgType ?? item.msg_type ?? 0),
    isRead: Number(item.isRead ?? item.is_read ?? 0),
    createdAt: Number(item.createdAt ?? item.created_at ?? 0),
    readAt: Number(item.readAt ?? item.read_at ?? 0),
    conversationId: Number(item.conversationId ?? item.conversation_id ?? 0)
  }
}

/** 我的消息列表（分页 + 类型筛选） */
export async function listMessages(page = 1, pageSize = 10, msgType = 0) {
  const res = await request.get<any, any>('/message/v1/list', {
    params: { page, page_size: pageSize, msg_type: msgType }
  })
  return {
    items: (res.items || []).map(normalizeMessage),
    total: res.total || 0,
    unreadCount: res.unreadCount ?? res.unread_count ?? 0
  } as MessageListResult
}

/** 我的未读消息数 */
export function getUnreadCount() {
  return request.get<any, { count: number }>('/message/v1/unread_count')
}

/** 标记单条消息已读 */
export function markMessageRead(id: number) {
  return request.post<any, { success: boolean }>(`/message/v1/read/${id}`)
}

/** 全部标记已读 */
export function markAllMessageRead() {
  return request.post<any, { success: boolean }>('/message/v1/read_all')
}
