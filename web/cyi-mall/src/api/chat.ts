/**
 * IM 客服相关 API（用户端）
 */
import request from './index'

/** 会话项（用户端视角） */
export interface ConversationItem {
  id: number
  agentId: number // 客服ID（0=未分配）
  status: number // 1=待接入 2=进行中 3=已结束
  serviceType: number // 1=人工 2=AI接待中
  lastMessage: string // 最后消息摘要
  lastMsgType: number // 最后消息类型
  lastMessageAt: number // 最后消息时间
  unreadCount: number // 未读数（客服发的未读）
  agentName: string // 客服姓名
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

/** 会话状态标签 */
export const CONV_STATUS_LABEL: Record<number, string> = {
  1: '待接入',
  2: '进行中',
  3: '已结束'
}

/** 归一化会话项 */
function normalizeConv(item: any): ConversationItem {
  return {
    id: Number(item.id ?? 0),
    agentId: Number(item.agentId ?? item.agent_id ?? 0),
    status: Number(item.status ?? 0),
    serviceType: Number(item.serviceType ?? item.service_type ?? 1),
    lastMessage: item.lastMessage ?? item.last_message ?? '',
    lastMsgType: Number(item.lastMsgType ?? item.last_msg_type ?? 1),
    lastMessageAt: Number(item.lastMessageAt ?? item.last_message_at ?? 0),
    unreadCount: Number(item.unreadCount ?? item.unread_count ?? 0),
    agentName: item.agentName ?? item.agent_name ?? '',
    createdAt: Number(item.createdAt ?? item.created_at ?? 0)
  }
}

/** 归一化消息项（供 REST 响应与 WS 推送共用；WS 帧为后端 snake_case 字段） */
export function normalizeMsg(item: any): ChatMessageItem {
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

/** 创建会话（复用未结束会话） */
export async function createConversation() {
  const res = await request.post<any, any>('/chat/v1/conversation', {})
  return normalizeConv(res.conversation ?? res)
}

/** 我的会话列表 */
export async function listMyConversations(params: { page?: number; page_size?: number; status?: number } = {}) {
  const res = await request.get<any, { items: any[]; total: number }>('/chat/v1/conversations', { params })
  return { items: (res.items || []).map(normalizeConv), total: res.total || 0 }
}

/** 会话历史消息 */
export async function getChatMessages(conversationId: number, params: { cursor_id?: number; limit?: number; direction?: number } = {}) {
  const res = await request.get<any, { items: any[]; has_more: boolean }>(`/chat/v1/conversations/${conversationId}/messages`, { params })
  return { items: (res.items || []).map(normalizeMsg), hasMore: !!res.has_more }
}

/** 用户发送消息 */
export function sendChatMessage(conversationId: number, data: { msg_type?: number; content: string; extra?: string }) {
  return request.post<any, { message_id: number; created_at: number; messageId?: number }>(`/chat/v1/conversations/${conversationId}/messages`, {
    msg_type: data.msg_type ?? 1,
    content: data.content,
    extra: data.extra ?? ''
  })
}

/** 用户标记会话已读 */
export function markChatRead(conversationId: number) {
  return request.post<any, { success: boolean }>(`/chat/v1/conversations/${conversationId}/read`, {})
}

/** 用户结束会话 */
export function closeChatConversation(conversationId: number) {
  return request.post<any, { success: boolean }>(`/chat/v1/conversations/${conversationId}/close`, {})
}
