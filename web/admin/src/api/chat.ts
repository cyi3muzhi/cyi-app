/**
 * cyi商城管理后台 - 客服工作台 API
 */
import request from './index'

/** 会话项（客服端视角） */
export interface ConversationItem {
  id: number
  userId: number // 用户ID
  agentId: number // 客服ID（0=未分配）
  status: number // 1=待接入 2=进行中 3=已结束
  serviceType: number // 1=人工 2=AI接待中
  lastMessage: string // 最后消息摘要
  lastMsgType: number // 最后消息类型
  lastMessageAt: number // 最后消息时间
  unreadCount: number // 未读数（用户发的未读）
  userName: string // 用户昵称
  userOnline: number // 用户是否在线 0=否 1=是
  createdAt: number // 创建时间
}

/** 消息项 */
export interface ChatMessageItem {
  id: number
  conversationId: number
  senderType: number // 1=用户 2=客服 3=系统 4=AI
  senderId: number
  msgType: number // 1=文本 2=图片 3=商品卡片 4=订单卡片 5=系统
  content: string
  extra: string
  isRead: number
  createdAt: number
}

/** 接待统计 */
export interface ChatStats {
  totalToday: number
  processing: number
  resolvedToday: number
}

/** 归一化会话项（Kratos protojson 输出 camelCase，兼容 snake_case） */
function normalizeConv(item: any): ConversationItem {
  return {
    id: Number(item.id ?? 0),
    userId: Number(item.userId ?? item.user_id ?? 0),
    agentId: Number(item.agentId ?? item.agent_id ?? 0),
    status: Number(item.status ?? 0),
    serviceType: Number(item.serviceType ?? item.service_type ?? 1),
    lastMessage: item.lastMessage ?? item.last_message ?? '',
    lastMsgType: Number(item.lastMsgType ?? item.last_msg_type ?? 1),
    lastMessageAt: Number(item.lastMessageAt ?? item.last_message_at ?? 0),
    unreadCount: Number(item.unreadCount ?? item.unread_count ?? 0),
    userName: item.userName ?? item.user_name ?? '',
    userOnline: Number(item.userOnline ?? item.user_online ?? 0),
    createdAt: Number(item.createdAt ?? item.created_at ?? 0)
  }
}

/** 归一化消息项 */
function normalizeMsg(item: any): ChatMessageItem {
  return {
    id: Number(item.id ?? 0),
    conversationId: Number(item.conversationId ?? item.conversation_id ?? 0),
    senderType: Number(item.senderType ?? item.sender_type ?? 0),
    senderId: Number(item.senderId ?? item.sender_id ?? 0),
    msgType: Number(item.msgType ?? item.msg_type ?? 1),
    content: item.content ?? '',
    extra: item.extra ?? '',
    isRead: Number(item.isRead ?? item.is_read ?? 0),
    createdAt: Number(item.createdAt ?? item.created_at ?? 0)
  }
}

/** 待接入会话列表 */
export async function getPendingConversations(params: { page?: number; page_size?: number } = {}) {
  const res = await request.get<any, { items: any[]; total: number }>('/admin/v1/chat/conversations/pending', { params })
  return { items: (res.items || []).map(normalizeConv), total: res.total || 0 }
}

/** 我的会话列表 */
export async function getAgentConversations(params: { page?: number; page_size?: number; status?: number } = {}) {
  const res = await request.get<any, { items: any[]; total: number }>('/admin/v1/chat/conversations', { params })
  return { items: (res.items || []).map(normalizeConv), total: res.total || 0 }
}

/** 客服接入会话 */
export function acceptConversation(conversationId: number) {
  return request.post<any, { success: boolean }>(`/admin/v1/chat/conversations/${conversationId}/accept`, {})
}

/** 客服转接会话 */
export function transferConversation(conversationId: number, targetAgentId: number) {
  return request.post<any, { success: boolean }>(`/admin/v1/chat/conversations/${conversationId}/transfer`, { target_agent_id: targetAgentId })
}

/** 会话历史消息 */
export async function getChatMessages(conversationId: number, params: { cursor_id?: number; limit?: number; direction?: number } = {}) {
  const res = await request.get<any, { items: any[]; has_more: boolean }>(`/admin/v1/chat/conversations/${conversationId}/messages`, { params })
  return { items: (res.items || []).map(normalizeMsg), hasMore: !!res.has_more }
}

/** 客服发送消息 */
export function sendChatMessage(conversationId: number, data: { msg_type?: number; content: string; extra?: string }) {
  return request.post<any, { message_id: number; created_at: number; messageId?: number }>(`/admin/v1/chat/conversations/${conversationId}/messages`, {
    msg_type: data.msg_type ?? 1,
    content: data.content,
    extra: data.extra ?? ''
  })
}

/** 客服标记会话已读 */
export function markChatRead(conversationId: number) {
  return request.post<any, { success: boolean }>(`/admin/v1/chat/conversations/${conversationId}/read`, {})
}

/** 客服结束会话 */
export function closeChatConversation(conversationId: number) {
  return request.post<any, { success: boolean }>(`/admin/v1/chat/conversations/${conversationId}/close`, {})
}

/** 客服接待统计 */
export async function getChatStats() {
  const res = await request.get<any, any>('/admin/v1/chat/stats')
  return {
    totalToday: Number(res.totalToday ?? res.total_today ?? 0),
    processing: Number(res.processing ?? 0),
    resolvedToday: Number(res.resolvedToday ?? res.resolved_today ?? 0)
  }
}
