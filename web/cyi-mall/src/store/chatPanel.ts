/**
 * 聊天面板状态管理（悬浮客服小窗）
 * 承载面板开关、视图切换（聊天/会话列表）、目标会话、会话列表与未读角标。
 * 供全局浮动组件 CustomerServiceFloat 以及个人中心/消息中心入口共用。
 */
import { defineStore, acceptHMRUpdate } from 'pinia'
import { ref, computed } from 'vue'
import { createConversation, listMyConversations, type ConversationItem } from '@/api/chat'

export type ChatPanelView = 'chat' | 'list'

export const useChatPanel = defineStore('chatPanel', () => {
  /** 面板是否打开 */
  const open = ref(false)
  /** 当前视图：聊天 / 会话列表 */
  const view = ref<ChatPanelView>('chat')
  /** 当前会话 ID */
  const activeConversationId = ref(0)
  /** 会话列表（供会话切换与未读角标） */
  const conversations = ref<ConversationItem[]>([])
  /** 各会话未读计数 */
  const unread = ref<Record<number, number>>({})

  /** 悬浮圆钮未读角标总数 */
  const unreadTotal = computed(() =>
    Object.values(unread.value).reduce((sum, n) => sum + (n > 0 ? 1 : 0), 0)
  )

  /** 当前会话对象 */
  const activeConversation = computed(
    () => conversations.value.find(c => c.id === activeConversationId.value) || null
  )

  /** 打开面板进入聊天视图；无可复用会话则创建 */
  async function openChat() {
    open.value = true
    view.value = 'chat'
    try {
      // 幂等创建/复用未结束会话
      const conv = await createConversation()
      setActive(conv.id)
    } catch {
      // 失败静默，仅打开面板留空聊天
    }
  }

  /** 打开面板进入会话列表视图 */
  async function openList() {
    open.value = true
    view.value = 'list'
    await refreshList()
  }

  /** 切换到指定会话聊天视图并清未读 */
  function showChat(id: number) {
    setActive(id)
    view.value = 'chat'
    clearUnread(id)
  }

  /** 左上角切换器：chat ⇄ list 互切；切到列表时刷新 */
  async function switchView() {
    view.value = view.value === 'chat' ? 'list' : 'chat'
    if (view.value === 'list') {
      await refreshList()
    }
  }

  /** 收起面板，回到悬浮圆钮；下次打开默认聊天视图 */
  function collapse() {
    open.value = false
    view.value = 'chat'
  }

  /** 拉取会话列表并写入，同时将服务端未读数并入 */
  async function refreshList() {
    try {
      const res = await listMyConversations({ page: 1, page_size: 50 })
      conversations.value = res.items
      for (const c of res.items) {
        if (c.unreadCount > 0) unread.value[c.id] = c.unreadCount
      }
      // 若无当前会话，取活跃（未结束）的第一个
      if (!activeConversationId.value) {
        const active = res.items.find(c => c.status !== 3)
        if (active) setActive(active.id)
      }
    } catch {
      // 静默
    }
  }

  /** 设置当前会话 ID */
  function setActive(id: number) {
    activeConversationId.value = id
  }

  /** 非当前会话未读 +1 */
  function bumpUnread(id: number) {
    unread.value[id] = (unread.value[id] || 0) + 1
  }

  /** 清空某会话未读 */
  function clearUnread(id: number) {
    delete unread.value[id]
  }

  return {
    open,
    view,
    activeConversationId,
    activeConversation,
    conversations,
    unread,
    unreadTotal,
    openChat,
    openList,
    showChat,
    switchView,
    collapse,
    refreshList,
    setActive,
    bumpUnread,
    clearUnread
  }
})

// 启用 store 热更新
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useChatPanel, import.meta.hot))
}