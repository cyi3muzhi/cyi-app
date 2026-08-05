<template>
  <!-- ============================================================ -->
  <!-- 1. 页面标题栏                                                  -->
  <!-- ============================================================ -->
  <div class="page-header">
    <div class="container-shop">
      <div class="page-header-inner">
        <BackButton fallback="/user" />
        <h1 class="page-title">
          消息中心
          <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }} 条未读</span>
        </h1>
        <button class="read-all-btn" :disabled="unreadCount === 0 || loading" @click="onMarkAllRead">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
          </svg>
          全部已读
        </button>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 2. 类型筛选 Tab                                               -->
  <!-- ============================================================ -->
  <div class="msg-tabs">
    <div class="container-shop">
      <div class="tabs-inner">
        <button
          v-for="tab in tabs"
          :key="tab.value"
          class="tab-btn"
          :class="{ active: activeTab === tab.value }"
          @click="switchTab(tab.value)"
        >
          {{ tab.label }}
          <span v-if="tab.value === 0 && unreadCount > 0" class="tab-dot"></span>
        </button>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 3. 消息列表                                                   -->
  <!-- ============================================================ -->
  <div class="msg-list">
    <div class="container-shop">
      <!-- 加载中 -->
      <div v-if="loading" class="state-wrap">
        <p class="state-text">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="messages.length === 0" class="state-wrap">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="var(--shop-border)" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
        <p class="state-text">暂无{{ activeTab === 0 ? '' : tabs.find(t => t.value === activeTab)?.label }}消息</p>
      </div>

      <!-- 消息卡片 -->
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="msg-card"
        :class="{ unread: msg.isRead === 0 }"
        @click="onOpenMessage(msg)"
      >
        <!-- 类型图标 -->
        <div class="msg-icon" :class="`type-${msg.msgType}`">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <template v-if="msg.msgType === 1">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
            </template>
            <template v-else-if="msg.msgType === 3">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.83z"/><line x1="7" y1="7" x2="7.01" y2="7"/>
            </template>
            <template v-else>
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </template>
          </svg>
        </div>

        <!-- 消息内容 -->
        <div class="msg-main">
          <div class="msg-title-row">
            <span class="msg-type-tag">{{ typeLabel(msg.msgType) }}</span>
            <h3 class="msg-title">{{ msg.title }}</h3>
            <span v-if="msg.isRead === 0" class="unread-dot"></span>
          </div>
          <p class="msg-excerpt" v-if="msg.content">{{ msg.content }}</p>
          <div class="msg-meta">
            <span class="msg-time">{{ formatTime(msg.createdAt) }}</span>
            <span v-if="msg.isRead === 1" class="msg-read-state">已读</span>
          </div>
        </div>

        <!-- 展开箭头 -->
        <svg class="msg-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </div>

      <!-- 加载更多 -->
      <div v-if="!loading && messages.length > 0 && hasMore" class="load-more-wrap">
        <button class="load-more-btn" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? '加载中...' : '加载更多' }}
        </button>
      </div>
      <div v-else-if="!loading && messages.length > 0" class="load-more-wrap">
        <span class="no-more-text">没有更多了</span>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 4. 消息详情弹层                                               -->
  <!-- ============================================================ -->
  <div v-if="detailMsg" class="modal-mask" @click.self="closeDetail">
    <div class="modal-card">
      <div class="modal-head">
        <div class="modal-icon" :class="`type-${detailMsg.msgType}`">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <template v-if="detailMsg.msgType === 1">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
            </template>
            <template v-else-if="detailMsg.msgType === 3">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.83z"/><line x1="7" y1="7" x2="7.01" y2="7"/>
            </template>
            <template v-else>
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </template>
          </svg>
        </div>
        <div class="modal-title-wrap">
          <span class="modal-type-tag">{{ typeLabel(detailMsg.msgType) }}</span>
          <h2 class="modal-title">{{ detailMsg.title }}</h2>
        </div>
        <button class="modal-close" @click="closeDetail">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="modal-time">{{ formatTime(detailMsg.createdAt) }}</div>
      <div class="modal-content">{{ detailMsg.content }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 消息中心页面
 * 类型 Tab（全部/订单/系统/营销）+ 消息列表 + 详情弹层 + 标记已读/全部已读
 * 数据来源：/message/v1/*
 */
import { ref, onMounted } from 'vue'
import BackButton from '@/components/BackButton.vue'
import {
  listMessages,
  getUnreadCount,
  markMessageRead,
  markAllMessageRead,
  MSG_TYPE_LABEL,
  type MessageItem
} from '@/api/message'

// ---- Tab 配置 ----
const tabs = [
  { label: '全部', value: 0 },
  { label: '订单', value: 1 },
  { label: '系统', value: 2 },
  { label: '营销', value: 3 }
]
const activeTab = ref(0)

// ---- 列表状态 ----
const messages = ref<MessageItem[]>([])
const unreadCount = ref(0)
const loading = ref(true)
const loadingMore = ref(false)
const page = ref(1)
const pageSize = 10
const total = ref(0)
const hasMore = ref(false)

// ---- 详情弹层 ----
const detailMsg = ref<MessageItem | null>(null)

/** 消息类型展示名 */
function typeLabel(type: number): string {
  return MSG_TYPE_LABEL[type] || '通知'
}

/** 格式化时间 */
function formatTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 加载消息列表 */
async function fetchMessages(reset = false) {
  if (reset) {
    page.value = 1
    loading.value = true
  } else {
    loadingMore.value = true
  }
  try {
    const res = await listMessages(page.value, pageSize, activeTab.value)
    messages.value = reset ? res.items : [...messages.value, ...res.items]
    total.value = res.total
    hasMore.value = messages.value.length < res.total
    unreadCount.value = res.unreadCount
  } catch {
    if (reset) messages.value = []
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

/** 刷新未读数 */
async function refreshUnread() {
  try {
    const res = await getUnreadCount()
    unreadCount.value = res.count || 0
  } catch {
    // 未读数获取失败不阻塞页面
  }
}

/** 切换 Tab */
function switchTab(value: number) {
  if (activeTab.value === value) return
  activeTab.value = value
  fetchMessages(true)
}

/** 加载更多 */
function loadMore() {
  page.value += 1
  fetchMessages(false)
}

/** 打开消息详情（未读则标记已读） */
async function onOpenMessage(msg: MessageItem) {
  detailMsg.value = msg
  if (msg.isRead === 0) {
    try {
      await markMessageRead(msg.id)
      msg.isRead = 1
      if (unreadCount.value > 0) unreadCount.value -= 1
    } catch {
      // 标记已读失败不影响查看
    }
  }
}

/** 关闭详情 */
function closeDetail() {
  detailMsg.value = null
}

/** 全部标记已读 */
async function onMarkAllRead() {
  try {
    await markAllMessageRead()
    messages.value.forEach(m => (m.isRead = 1))
    unreadCount.value = 0
  } catch (e: any) {
    alert(e?.response?.data?.message || e?.message || '操作失败，请重试')
  }
}

onMounted(() => fetchMessages(true))
</script>

<style scoped>
/* ============================================
   消息中心页面样式
   ============================================ */

/* ---- 页面标题 ---- */
.page-header {
  padding: 24px 0 0;
}

.page-header-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.unread-badge {
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.read-all-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
  background: var(--shop-card);
  color: var(--shop-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.read-all-btn:hover:not(:disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.read-all-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---- 类型 Tab ---- */
.msg-tabs {
  padding: 20px 0 0;
}

.tabs-inner {
  display: flex;
  gap: 8px;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-xl);
  padding: 4px;
  width: fit-content;
}

.tab-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: transparent;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn.active {
  background: var(--shop-card);
  color: var(--shop-primary);
  font-weight: 600;
  box-shadow: var(--shop-shadow-sm);
}

.tab-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--shop-destructive);
}

/* ---- 消息列表 ---- */
.msg-list {
  padding: 20px 0 48px;
}

.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 24px;
}

.state-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.msg-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  margin-bottom: 12px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  cursor: pointer;
  transition: box-shadow 0.2s, transform 0.2s, border-color 0.2s;
}

.msg-card:hover {
  box-shadow: var(--shop-shadow-lg);
  transform: translateY(-1px);
}

.msg-card.unread {
  border-color: hsla(217, 91%, 60%, 0.35);
  background: linear-gradient(90deg, hsla(217, 91%, 60%, 0.04), var(--shop-card) 40%);
}

/* 类型图标 */
.msg-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 12px;
  flex-shrink: 0;
}

.msg-icon.type-1 {
  color: #2563eb;
  background: hsla(217, 91%, 60%, 0.12);
}

.msg-icon.type-2 {
  color: #059669;
  background: hsla(160, 84%, 39%, 0.12);
}

.msg-icon.type-3 {
  color: #d97706;
  background: hsla(38, 92%, 50%, 0.14);
}

/* 消息主体 */
.msg-main {
  flex: 1;
  min-width: 0;
}

.msg-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.msg-type-tag {
  flex-shrink: 0;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.msg-title {
  flex: 1;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--shop-destructive);
  flex-shrink: 0;
}

.msg-excerpt {
  margin-top: 4px;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--shop-muted-foreground);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.msg-meta {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.msg-time {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.msg-read-state {
  font-size: 0.6875rem;
  color: var(--shop-muted);
}

.msg-arrow {
  flex-shrink: 0;
  color: var(--shop-muted);
}

/* ---- 加载更多 ---- */
.load-more-wrap {
  display: flex;
  justify-content: center;
  padding: 16px 0 8px;
}

.load-more-btn {
  padding: 8px 32px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
  background: var(--shop-card);
  color: var(--shop-foreground);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.load-more-btn:hover:not(:disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.load-more-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.no-more-text {
  font-size: 0.75rem;
  color: var(--shop-muted);
}

/* ---- 详情弹层 ---- */
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgba(0, 0, 0, 0.45);
}

.modal-card {
  width: 100%;
  max-width: 520px;
  max-height: calc(100vh - 80px);
  overflow-y: auto;
  padding: 24px;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-xl);
}

.modal-head {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.modal-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 12px;
  flex-shrink: 0;
}

.modal-icon.type-1 {
  color: #2563eb;
  background: hsla(217, 91%, 60%, 0.12);
}

.modal-icon.type-2 {
  color: #059669;
  background: hsla(160, 84%, 39%, 0.12);
}

.modal-icon.type-3 {
  color: #d97706;
  background: hsla(38, 92%, 50%, 0.14);
}

.modal-title-wrap {
  flex: 1;
  min-width: 0;
}

.modal-type-tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.modal-title {
  margin-top: 6px;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
  word-break: break-word;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: background 0.2s;
}

.modal-close:hover {
  background: var(--shop-border);
}

.modal-time {
  margin: 16px 0 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--shop-border);
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.modal-content {
  font-size: 0.9375rem;
  line-height: 1.8;
  color: var(--shop-foreground);
  white-space: pre-wrap;
  word-break: break-word;
}

/* ---- 响应式 ---- */
@media (max-width: 640px) {
  .msg-card {
    padding: 14px 14px;
  }

  .msg-icon {
    width: 38px;
    height: 38px;
  }

  .read-all-btn {
    padding: 6px 12px;
    font-size: 0.8125rem;
  }

  .tabs-inner {
    width: 100%;
  }

  .tab-btn {
    flex: 1;
    justify-content: center;
    padding: 8px 8px;
  }
}
</style>
