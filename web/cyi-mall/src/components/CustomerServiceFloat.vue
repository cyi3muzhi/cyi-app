<template>
  <!-- 悬浮在线客服圆钮（收起态） -->
  <div v-if="!panel.open" class="cs-fab" @click="onFabClick">
    <div class="cs-btn">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
      </svg>
      <span v-if="panel.unreadTotal > 0" class="cs-badge">{{ panel.unreadTotal > 99 ? '99+' : panel.unreadTotal }}</span>
    </div>
    <span class="cs-tip">在线客服</span>
  </div>

  <!-- 悬浮小窗 / 全屏抽屉（展开态） -->
  <Teleport to="body">
    <div v-if="panel.open" class="cw-root">
      <div class="cw-mask" @click="onCollapse"></div>
      <section class="cw-window" :style="winStyle" @pointerdown.stop>
        <!-- 头部：左上角会话切换器 + 标题 + 最小化 -->
        <header class="cw-head" @pointerdown="startDrag">
          <button class="cw-switch" @click.stop="onSwitchView" :title="panel.view === 'chat' ? '选择会话' : '返回聊天'">
            <svg v-if="panel.view === 'chat'" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            <svg v-else width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
          </button>
          <div class="cw-title">{{ headTitle }}</div>
          <button class="cw-min" @click.stop="onCollapse" title="收起">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
        </header>

        <!-- 会话列表视图 -->
        <div v-if="panel.view === 'list'" class="cw-body">
          <div v-if="panel.conversations.length === 0" class="cw-empty">
            <p>暂无会话</p>
            <button class="cw-new" @click="onStartNew">发起咨询</button>
          </div>
          <button
            v-for="c in panel.conversations"
            :key="c.id"
            class="cw-cc"
            :class="{ active: c.id === panel.activeConversationId }"
            @click="onPickConversation(c.id)"
          >
            <span class="cw-cc-title">{{ c.status === 3 ? '已结束会话' : (c.agentName ? `客服 · ${c.agentName}` : '在线客服') }}</span>
            <span v-if="unreadOf(c.id) > 0" class="cw-cc-unread">{{ unreadOf(c.id) }}</span>
            <span class="cw-cc-last">{{ c.lastMessage || (c.status === 1 ? '等待客服接入' : '') }}</span>
          </button>
        </div>

        <!-- 聊天视图 -->
        <div v-else class="cw-chat">
          <!-- 状态条 -->
          <div class="cw-status" :class="statusClass">
            <span class="cw-status-dot"></span>
            <span class="cw-status-text">{{ statusText }}</span>
            <span class="cw-ws" :class="{ on: wsConnected }">
              <i></i>{{ wsConnected ? '实时连接' : '连接中…' }}
            </span>
            <button v-if="!isClosed" class="cw-end" @click="closeConversation">结束会话</button>
          </div>

          <!-- 消息列表 -->
          <div ref="msgBox" class="cw-msgs">
            <div v-if="messages.length === 0" class="cw-empty">
              <p>您好，请问有什么可以帮您？</p>
            </div>
            <div
              v-for="m in messages"
              :key="m.id"
              class="cw-row"
              :class="m.senderType === 1 ? 'mine' : 'theirs'"
            >
              <div v-if="m.msgType === 3" class="cw-card">
                <div class="cw-card-t">{{ cardOf(m).product_name || '商品详情' }}</div>
                <div class="cw-card-p">¥{{ cardOf(m).price || '--' }}</div>
                <div class="cw-card-s">商品ID：{{ cardOf(m).product_id || '--' }}</div>
              </div>
              <div v-else-if="m.msgType === 4" class="cw-card">
                <div class="cw-card-t">订单信息</div>
                <div class="cw-card-s">订单号：{{ cardOf(m).order_no || '--' }}</div>
              </div>
              <div v-else class="cw-bubble" :class="bubbleClass(m)">{{ m.content }}</div>
              <div class="cw-time">{{ formatTime(m.createdAt) }}</div>
            </div>
          </div>

          <!-- 输入区 -->
          <div class="cw-input">
            <div v-if="!isClosed" class="cw-input-wrap">
              <textarea
                v-model="inputText"
                class="cw-box"
                rows="1"
                placeholder="输入您的问题…"
                @keydown.enter.exact.prevent="send"
              ></textarea>
              <button class="cw-send" :disabled="sending || !inputText.trim()" @click="send">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 19V5"/><path d="M5 12l7-7 7 7"/></svg>
              </button>
            </div>
            <p v-else class="cw-closed">
              会话已结束
              <a class="cw-relink" @click="onStartNew">发起新咨询</a>
            </p>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * 悬浮在线客服：收起为圆钮，展开为可拖动小窗（桌面）/全屏抽屉（移动）。
 * 左上角切换器在「聊天 ⇄ 会话列表」间切换；收起断开 WebSocket，圆钮显示未读角标。
 */
import { computed, nextTick, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { useChatPanel } from '@/store/chatPanel'
import { toast } from '@/utils/toast'
import {
  createConversation,
  getChatMessages,
  listMyConversations,
  sendChatMessage,
  markChatRead,
  closeChatConversation,
  normalizeMsg,
  type ChatMessageItem
} from '@/api/chat'

const router = useRouter()
const userStore = useUserStore()
const panel = useChatPanel()

// ---- 拖拽状态（桌面小窗） ----
const winPos = ref({ x: 0, y: 0, init: false })
let drag: { sx: number; sy: number; ox: number; oy: number } | null = null
const WIDTH = 360
const HEIGHT = 560

const winStyle = computed(() => {
  if (panel.open) {
    if (!winPos.value.init) {
      winPos.value = { init: true, x: window.innerWidth - WIDTH - 20, y: window.innerHeight - HEIGHT - 96 }
    }
    return { left: `${winPos.value.x}px`, top: `${winPos.value.y}px`, width: `${WIDTH}px`, height: `${HEIGHT}px` }
  }
  return {}
})

// ---- 聊天状态 ----
const messages = ref<ChatMessageItem[]>([])
const inputText = ref('')
const sending = ref(false)
const msgBox = ref<HTMLDivElement | null>(null)
const convStatus = ref(1)
const convServiceType = ref(1)
const convAgentName = ref('')
const wsConnected = ref(false)
let ws: WebSocket | null = null
let disposed = false

const isClosed = computed(() => convStatus.value === 3)

const headTitle = computed(() =>
  panel.view === 'list' ? '选择会话' : convAgentName.value ? `客服 ${convAgentName.value}` : '在线客服'
)

const statusText = computed(() => {
  if (convStatus.value === 3) return '会话已结束'
  if (convServiceType.value === 2) return 'AI 客服接待中'
  if (convStatus.value === 2) return convAgentName.value ? `客服 ${convAgentName.value} 正在为您服务` : '客服正在为您服务'
  return '正在为您接入人工客服…'
})
const statusClass = computed(() => {
  if (convStatus.value === 3) return 'st-closed'
  if (convServiceType.value === 2) return 'st-ai'
  if (convStatus.value === 2) return 'st-live'
  return 'st-pending'
})

function unreadOf(id: number) {
  return panel.unread[id] || 0
}

function cardOf(m: ChatMessageItem): any {
  if (!m.extra) return {}
  try { return JSON.parse(m.extra) } catch { return {} }
}
function bubbleClass(m: ChatMessageItem) {
  if (m.msgType === 5) return 'b-system'
  if (m.senderType === 2) return 'b-agent'
  if (m.senderType === 4) return 'b-ai'
  return 'b-user'
}
function formatTime(ts: number) {
  if (!ts) return ''
  const d = new Date(Number(ts) * 1000)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}
function scrollToBottom() {
  nextTick(() => { if (msgBox.value) msgBox.value.scrollTop = msgBox.value.scrollHeight })
}

// ---- 会话/消息操作 ----
async function loadMessages(id: number) {
  try {
    const res = await getChatMessages(id, { limit: 50, direction: 1 })
    messages.value = res.items
    markChatRead(id).catch(() => {})
    scrollToBottom()
  } catch (err) { console.warn('[chat] loadMessages 失败', err) }
}

function applyConversation(c: { status: number; serviceType: number; agentName?: string }) {
  convStatus.value = c.status
  convServiceType.value = c.serviceType
  convAgentName.value = c.agentName || ''
}

function openChatAt(id: number, conv?: { status: number; serviceType: number; agentName?: string }) {
  // 初始化会话状态：优先用当前会话对象（若客服已接入但事件未推送，也能正确显示状态）
  const src = conv || panel.conversations.find(c => c.id === id)
  if (src) applyConversation({ status: src.status, serviceType: src.serviceType, agentName: src.agentName })
  loadMessages(id)
  connectWS()
}

async function resetConversation() {
  try {
    const conv = await createConversation()
    panel.setActive(conv.id)
    applyConversation(conv)
    loadMessages(conv.id)
    refreshList()
  } catch { /* 静默 */ }
}

async function refreshList() {
  try {
    const res = await listMyConversations({ page: 1, page_size: 50 })
    panel.conversations.length = 0
    panel.conversations.push(...res.items)
  } catch { /* 静默 */ }
}

async function send() {
  const id = panel.activeConversationId
  const text = inputText.value.trim()
  if (!text || sending.value || !id) return
  sending.value = true
  let messageId = 0
  try {
    const reply = await sendChatMessage(id, { msg_type: 1, content: text })
    console.log('[chat] REST reply:', reply)
    messageId = Number(reply.message_id ?? reply.messageId ?? 0)
    inputText.value = ''
    if (messageId > 0) {
      messages.value.push({ id: messageId, conversationId: id, senderType: 1, senderId: 0, msgType: 1, content: text, extra: '', isRead: 0, createdAt: Date.now() / 1000 })
      scrollToBottom()
    } else {
      console.warn('[chat] 发送成功但 messageId 为 0，未本地插入')
    }
  } catch (err) { console.error('[chat] send 失败', err) } finally {
    sending.value = false
  }
  // 兜底轮询：AI/客服回复生成需 2~5s，递增重拉历史覆盖（WS 失效或延迟时也能看到回复）。
  // 记录「发送后」最后一条已知消息的 id：期间若未通过 WS 收到更新（最后 id 不变）才拉取，避免重复覆盖
  const lastId = messages.value.length ? messages.value[messages.value.length - 1].id : 0
  ;[800, 2000, 3500, 5000].forEach(delay => {
    setTimeout(() => {
      if (disposed) return
      const nowLast = messages.value.length ? messages.value[messages.value.length - 1].id : 0
      if (nowLast <= lastId) loadMessages(id)
    }, delay)
  })
}

async function closeConversation() {
  const id = panel.activeConversationId
  if (!id || isClosed.value) return
  if (!window.confirm('确定结束本次会话吗？')) return
  try {
    await closeChatConversation(id)
    convStatus.value = 3
    refreshList()
  } catch { /* 已提示 */ }
}

// ---- WebSocket ----
function connectWS() {
  const token = localStorage.getItem('token')
  if (!token || ws || disposed) return
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws = new WebSocket(`${proto}://${location.host}/ws/chat/v2?token=${token}`)
  ws.onopen = () => { wsConnected.value = true; console.log('[WS] open', ws?.url) }
  ws.onclose = () => {
    wsConnected.value = false
    console.log('[WS] close')
    ws = null
    if (!disposed && panel.open) setTimeout(connectWS, 3000)
  }
  ws.onerror = (err) => { console.log('[WS] error', err); ws?.close() }
  ws.onmessage = (e) => {
    console.log('[WS] frame:', e.data)
    try {
      const frame = JSON.parse(e.data)
      if (frame.type === 'message') handleMessage(frame.data)
      else if (frame.type === 'conversation_status') handleConvStatus(frame.data)
    } catch { /* 忽略脏数据 */ }
  }
}

function handleMessage(raw: any) {
  // WS 帧为后端 snake_case 字段（conversation_id 等），统一归一化为前端 ChatMessageItem
  const msg = normalizeMsg(raw)
  const cvId = msg.conversationId
  if (cvId === panel.activeConversationId) {
    if (!messages.value.some(m => m.id === msg.id)) {
      messages.value.push(msg)
      scrollToBottom()
    }
    markChatRead(cvId).catch(() => {})
    panel.clearUnread(cvId)
  } else {
    panel.bumpUnread(cvId)
  }
}

function handleConvStatus(data: any) {
  const cvId = Number(data.conversation_id ?? data.conversationId ?? 0)
  const evt = Number(data.event_type ?? data.eventType ?? 0)
  if (cvId === panel.activeConversationId) {
    if (evt === 2) { convStatus.value = 2; convServiceType.value = 1 }
    else if (evt === 4) { convStatus.value = 3 }
  }
  refreshList()
}

// 统一入口：打开面板后由 watcher 负责加载消息与连接 WS；这里只做创建/复用会话
async function onFabClick() {
  if (!userStore.isLoggedIn) {
    toast('请先登录', 'error')
    router.push({ name: 'Login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }
  await panel.openChat()
}

// 面板打开且有目标会话时：加载该会话消息、连接 WS、刷新列表。
// 覆盖从个人中心 / 消息中心等 store 直达入口打开的场景（绕过了 onFabClick）。
watch(
  () => [panel.open, panel.activeConversationId] as const,
  ([open, id]) => {
    if (open && id && panel.view === 'chat' && !disposed) {
      openChatAt(id)
      refreshList()
    }
  }
)

async function onSwitchView() {
  await panel.switchView()
  if (panel.view === 'chat' && panel.activeConversationId) {
    openChatAt(panel.activeConversationId)
  }
}

function onPickConversation(id: number) {
  panel.showChat(id)
  openChatAt(id)
}

async function onStartNew() {
  await panel.openChat()
  await resetConversation()
}

function onCollapse() {
  panel.collapse()
  if (ws) { ws.close(); ws = null }
}

// ---- 拖拽（桌面） ----
function startDrag(e: PointerEvent) {
  // 仅桌面端可拖动；点击头部按钮不触发拖拽
  if (window.matchMedia('(max-width: 767px)').matches) return
  if ((e.target as HTMLElement).closest('button')) return
  drag = { sx: e.clientX, sy: e.clientY, ox: winPos.value.x, oy: winPos.value.y }
  document.body.style.userSelect = 'none'
  window.addEventListener('pointermove', onDragMove)
  window.addEventListener('pointerup', endDrag)
}
function onDragMove(e: PointerEvent) {
  if (!drag) return
  let x = drag.ox + (e.clientX - drag.sx)
  let y = drag.oy + (e.clientY - drag.sy)
  x = Math.min(Math.max(0, x), window.innerWidth - WIDTH)
  y = Math.min(Math.max(0, y), window.innerHeight - HEIGHT)
  winPos.value.x = x
  winPos.value.y = y
}
function endDrag() {
  if (!drag) return
  drag = null
  document.body.style.userSelect = ''
  window.removeEventListener('pointermove', onDragMove)
  window.removeEventListener('pointerup', endDrag)
}

onUnmounted(() => {
  disposed = true
  if (ws) { ws.close(); ws = null }
  window.removeEventListener('pointermove', onDragMove)
  window.removeEventListener('pointerup', endDrag)
})
</script>

<style scoped>
/* ---- 圆钮 ---- */
.cs-fab {
  position: fixed;
  right: 20px;
  bottom: 96px;
  z-index: 90;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.cs-btn {
  position: relative;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--shop-primary), #38bdf8);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.35);
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.2s;
}
.cs-btn:hover { transform: translateY(-3px) scale(1.05); box-shadow: 0 10px 26px rgba(37, 99, 235, 0.45); }
.cs-btn::before {
  content: '';
  position: absolute;
  inset: -6px;
  border-radius: 50%;
  border: 2px solid var(--shop-primary);
  opacity: 0.45;
  animation: csPulse 2.2s ease-out infinite;
}
@keyframes csPulse { 0% { transform: scale(0.9); opacity: 0.5; } 70% { transform: scale(1.25); opacity: 0; } 100% { transform: scale(1.25); opacity: 0; } }
.cs-tip {
  font-size: 11px;
  color: var(--shop-text-light);
  background: #fff;
  padding: 2px 8px;
  border-radius: 999px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  white-space: nowrap;
  opacity: 0;
  transform: translateY(2px);
  transition: all 0.2s;
}
.cs-fab:hover .cs-tip { opacity: 1; transform: translateY(0); }
.cs-badge {
  position: absolute;
  top: -2px;
  right: -2px;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: #ef4444;
  color: #fff;
  font-size: 11px;
  text-align: center;
  border: 2px solid #fff;
  box-sizing: content-box;
}

/* ---- 小窗根容器（桌面透明穿透 / 移动全屏遮罩） ---- */
.cw-root { position: fixed; inset: 0; z-index: 95; pointer-events: none; }
.cw-mask { position: absolute; inset: 0; pointer-events: none; background: transparent; transition: background 0.25s; }
.cw-window {
  position: fixed;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  pointer-events: auto;
  background: var(--shop-background);
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 16px;
  box-shadow: 0 24px 60px -12px rgba(15, 23, 42, 0.4), 0 8px 20px rgba(15, 23, 42, 0.15);
  animation: cwPop 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes cwPop { from { opacity: 0; transform: translateY(10px) scale(0.96); } to { opacity: 1; transform: translateY(0) scale(1); } }

/* ---- 头部 ---- */
.cw-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 12px;
  background: linear-gradient(180deg, #ffffff 0%, var(--shop-primary-bg) 140%);
  border-bottom: 1px solid var(--shop-border);
  cursor: grab;
  flex-shrink: 0;
  user-select: none;
}
.cw-head:active { cursor: grabbing; }
.cw-switch, .cw-min {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.7);
  color: var(--shop-text);
  cursor: pointer;
  transition: all 0.2s;
}
.cw-switch:hover, .cw-min:hover { background: #fff; color: var(--shop-primary); box-shadow: 0 2px 6px rgba(0,0,0,0.06); }
.cw-min { margin-left: auto; }
.cw-title { font-size: 14px; font-weight: 700; color: var(--shop-text); }

/* ---- 主体 ---- */
.cw-body, .cw-chat { flex: 1; min-height: 0; display: flex; flex-direction: column; background: var(--shop-background); }
.cw-body { overflow-y: auto; padding: 8px; }
.cw-cc {
  display: flex;
  flex-direction: column;
  gap: 4px;
  text-align: left;
  padding: 10px 12px;
  margin-bottom: 8px;
  border: 1px solid var(--shop-border);
  border-radius: 12px;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
}
.cw-cc:hover { border-color: var(--shop-primary); box-shadow: 0 4px 12px rgba(37,99,235,0.08); }
.cw-cc.active { border-color: var(--shop-primary); background: var(--shop-primary-bg); }
.cw-cc-title { font-size: 13px; font-weight: 600; color: var(--shop-text); }
.cw-cc-last { font-size: 12px; color: var(--shop-text-light); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cw-cc-unread {
  position: absolute;
  align-self: flex-end;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  text-align: center;
}
.cw-empty { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; padding: 40px 0; color: var(--shop-text-light); font-size: 13px; }
.cw-new {
  background: linear-gradient(135deg, var(--shop-primary), #38bdf8);
  color: #fff;
  border: none;
  border-radius: 20px;
  padding: 8px 22px;
  font-size: 13px;
  cursor: pointer;
}

/* ---- 状态条 ---- */
.cw-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  font-size: 12px;
  border-bottom: 1px solid var(--shop-border);
  flex-shrink: 0;
}
.cw-status-dot { width: 8px; height: 8px; border-radius: 50%; }
.cw-ws { margin-left: auto; font-size: 11px; color: var(--shop-text-light); display: inline-flex; align-items: center; gap: 4px; }
.cw-ws i { width: 6px; height: 6px; border-radius: 50%; background: #cbd5e1; }
.cw-ws.on { color: #059669; }
.cw-ws.on i { background: #059669; animation: blink 1.2s ease-in-out infinite; }
.cw-end {
  border: 1px solid var(--shop-border);
  background: #fff;
  color: var(--shop-text-light);
  border-radius: 999px;
  font-size: 11px;
  padding: 2px 8px;
  cursor: pointer;
  margin-left: 6px;
  transition: all 0.2s;
}
.cw-end:hover { color: #ef4444; border-color: #ef4444; }
@keyframes blink { 0%,100% { opacity: 1; } 50% { opacity: 0.3; } }
.st-live { color: #059669; } .st-live .cw-status-dot { background: #059669; box-shadow: 0 0 0 3px #05966922; }
.st-ai { color: #2563eb; } .st-ai .cw-status-dot { background: #2563eb; box-shadow: 0 0 0 3px #2563eb22; }
.st-pending { color: #ea580c; } .st-pending .cw-status-dot { background: #ea580c; box-shadow: 0 0 0 3px #ea580c22; }
.st-closed { color: #94a3b8; } .st-closed .cw-status-dot { background: #94a3b8; }

/* ---- 消息 ---- */
.cw-msgs { flex: 1; overflow-y: auto; padding: 12px; }
.cw-row { display: flex; flex-direction: column; align-items: flex-start; margin-bottom: 12px; max-width: 82%; }
.cw-row.mine { align-items: flex-end; margin-left: auto; }
.cw-bubble { padding: 9px 13px; border-radius: 12px; font-size: 13px; line-height: 1.5; word-break: break-word; white-space: pre-wrap; }
.b-user { background: linear-gradient(135deg, var(--shop-primary), #3b82f6); color: #fff; border-bottom-right-radius: 4px; }
.b-agent { background: #fff; color: var(--shop-text); border: 1px solid var(--shop-border); border-bottom-left-radius: 4px; }
.b-ai { background: #eef2f7; color: #475569; border-bottom-left-radius: 4px; }
.b-system { background: transparent; color: var(--shop-text-light); font-size: 11px; text-align: center; align-self: center; }
.cw-card { background: #fff; border: 1px solid var(--shop-border); border-radius: 12px; padding: 10px 14px; min-width: 180px; max-width: 260px; cursor: pointer; }
.cw-card-t { font-size: 13px; font-weight: 600; color: var(--shop-text); }
.cw-card-p { font-size: 15px; font-weight: 700; color: #ef4444; margin-top: 2px; }
.cw-card-s { font-size: 11px; color: var(--shop-text-light); margin-top: 2px; }
.cw-time { font-size: 10.5px; color: var(--shop-text-light); opacity: 0.8; margin-top: 3px; }

/* ---- 输入区 ---- */
.cw-input { flex-shrink: 0; border-top: 1px solid var(--shop-border); background: #fff; padding: 10px 12px calc(10px + env(safe-area-inset-bottom)); }
.cw-input-wrap { display: flex; align-items: flex-end; gap: 8px; }
.cw-box {
  flex: 1;
  border: 1px solid var(--shop-border);
  border-radius: 10px;
  padding: 9px 11px;
  font-size: 13px;
  font-family: inherit;
  resize: none;
  outline: none;
  min-height: 38px;
  max-height: 100px;
  line-height: 1.5;
  transition: border-color 0.2s;
}
.cw-box:focus { border-color: var(--shop-primary); }
.cw-send {
  width: 38px;
  height: 38px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--shop-primary), #38bdf8);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}
.cw-send:disabled { opacity: 0.5; cursor: not-allowed; }
.cw-closed { margin: 0; text-align: center; font-size: 12px; color: var(--shop-text-light); }
.cw-relink { color: var(--shop-primary); cursor: pointer; margin-left: 6px; }

@media (prefers-reduced-motion: reduce) { .cs-btn::before { animation: none; } }

/* ---- 移动端：全屏抽屉 ---- */
@media (max-width: 767px) {
  .cw-root { pointer-events: auto; display: flex; flex-direction: column; justify-content: flex-end; }
  .cw-mask { pointer-events: auto; background: rgba(15, 23, 42, 0.35); }
  .cw-window {
    left: 0 !important;
    right: 0 !important;
    top: auto !important;
    bottom: 0;
    width: 100% !important;
    height: 92vh !important;
    max-height: 92vh;
    border-radius: 16px 16px 0 0;
    animation: cwDrawer 0.28s cubic-bezier(0.16, 1, 0.3, 1);
  }
  @keyframes cwDrawer { from { transform: translateY(40px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
}
</style>