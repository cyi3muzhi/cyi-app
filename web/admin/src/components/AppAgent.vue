<template>
  <div class="agent-panel" :class="{ open: store.agentPanelOpen }">
    <div class="agent-overlay" @click="store.toggleAgentPanel()"></div>

    <div class="agent-drawer">
      <div class="agent-header">
        <div class="floating-bg"></div>
        <div class="agent-title">
          <span class="agent-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <circle cx="12" cy="12" r="4"/>
              <line x1="21.17" y1="8" x2="12" y2="8"/>
              <line x1="3.95" y1="6.06" x2="8.54" y2="14"/>
              <line x1="10.88" y1="21.94" x2="15.46" y2="14"/>
            </svg>
            <span class="icon-glow"></span>
          </span>
          <div class="agent-info">
            <span class="agent-name">智能助手</span>
            <span class="agent-status">
              <span class="status-dot"></span>在线
            </span>
          </div>
        </div>
        <div class="agent-header-actions">
          <el-button class="agent-clear-btn" size="small" @click="handleClearHistory">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
            新对话
          </el-button>
          <button class="agent-close" @click="store.toggleAgentPanel()">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      </div>

      <div class="agent-messages" ref="messagesContainer">
        <div class="animated-bg">
          <div class="bg-blob blob-1"></div>
          <div class="bg-blob blob-2"></div>
          <div class="bg-blob blob-3"></div>
        </div>

        <div
          v-for="(msg, index) in messages"
          :key="msg.timestamp"
          :ref="(el: any) => setMessageRef(msg.timestamp!, el)"
          :class="['message-item', msg.role]"
          :style="{ animationDelay: `${index * 0.05}s` }"
        >
          <div class="message-avatar">
            <div v-if="msg.role === 'assistant'" class="avatar-assistant">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <circle cx="12" cy="12" r="4"/>
                <line x1="21.17" y1="8" x2="12" y2="8"/>
                <line x1="3.95" y1="6.06" x2="8.54" y2="14"/>
                <line x1="10.88" y1="21.94" x2="15.46" y2="14"/>
              </svg>
            </div>
            <div v-else class="avatar-user">
              <img
                v-if="store.adminInfo?.avatar && store.adminInfo.avatar.startsWith('http')"
                :src="store.adminInfo.avatar"
                class="avatar-user-img"
              />
              <span v-else>{{ store.adminInfo?.name?.charAt(0) || '管' }}</span>
            </div>
          </div>
          <div class="message-content">
            <div class="message-bubble">
              <div v-if="msg.role === 'assistant' && !msg.content && isLoading" class="typing-indicator">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
              </div>
              <div v-else-if="msg.role === 'assistant' && msg.content" class="message-text" v-html="formatMessage(msg.content)"></div>
              <span v-else class="message-text">{{ msg.content }}</span>
            </div>
            <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
          </div>
        </div>
      </div>

      <div class="agent-input-area">
        <div class="input-wrapper">
          <el-input
            v-model="inputMessage"
            placeholder="输入您的问题..."
            class="agent-input"
            @keyup.enter="handleSend"
            :disabled="isLoading"
          />
          <div class="input-actions">
            <el-button
              class="agent-send-btn"
              @click="handleSend"
              :disabled="!inputMessage.trim() || isLoading"
              circle
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="22" y1="2" x2="11" y2="13"/>
                <polygon points="22 2 15 22 11 13 2 9 22 2"/>
              </svg>
            </el-button>
          </div>
        </div>
      </div>

      <div class="agent-quick-actions">
        <span class="quick-label">💡 快捷提问</span>
        <div class="quick-tags">
          <el-tag
            v-for="action in quickActions"
            :key="action.label"
            class="quick-tag"
            @click="handleQuickAction(action.query)"
            :disabled="isLoading"
          >
            {{ action.label }}
          </el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch, onMounted } from 'vue'
import { useAppStore } from '../store/app'
import { sendChatMessageStream, clearChatHistory, getChatHistory, type ChatMessage } from '../api/agent'

const store = useAppStore()
const inputMessage = ref('')
const isLoading = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)
const messageRefs = ref<Map<number, HTMLElement>>(new Map())
const scrollVelocity = ref(0)
const isScrolling = ref(false)

const messages = ref<ChatMessage[]>([])

const quickActions = [
  { label: '今日订单统计', query: '今天有多少订单？' },
  { label: '待处理订单', query: '有哪些订单需要处理？' },
  { label: '热销商品', query: '最近哪些商品卖得最好？' },
  { label: '用户增长', query: '本月用户增长情况如何？' }
]

onMounted(async () => {
  // 尝试从后端加载聊天历史
  const history = await getChatHistory()
  if (history && history.length > 0) {
    messages.value = history
  } else {
    messages.value = [
      {
        role: 'assistant',
        content: `您好！我是${store.shopName}的智能助手，请问有什么可以帮助您的？`,
        timestamp: Date.now()
      }
    ]
  }

  const container = messagesContainer.value
  if (container) {
    container.addEventListener('scroll', handleScroll, { passive: true })
  }
})

watch(() => store.agentPanelOpen, (isOpen) => {
  if (isOpen) {
    nextTick(() => {
      scrollToBottom()
    })
  }
})

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

function formatMessage(content: string): string {
  return content
    .replace(/\n/g, '<br/>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
}

function formatTime(timestamp?: number): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

async function handleSend() {
  const msg = inputMessage.value.trim()
  if (!msg || isLoading.value) return

  inputMessage.value = ''
  messages.value.push({
    role: 'user',
    content: msg,
    timestamp: Date.now()
  })

  isLoading.value = true
  await nextTick()
  scrollToBottom()

  // 先插入一条空的助手消息，后续流式追加内容
  const assistantMsg: ChatMessage = {
    role: 'assistant',
    content: '',
    timestamp: Date.now()
  }
  messages.value.push(assistantMsg)

  // 批量更新：将多个 chunk 合并到同一帧更新，减少重排次数
  let pendingContent = ''
  let rafBatchId = 0

  function flushBatch() {
    if (!pendingContent) return
    assistantMsg.content += pendingContent
    pendingContent = ''
    messages.value = [...messages.value]
    scrollToBottom()
  }

  try {
    await sendChatMessageStream(msg, (chunk) => {
      pendingContent += chunk
      if (!rafBatchId) {
        rafBatchId = requestAnimationFrame(() => {
          rafBatchId = 0
          flushBatch()
        })
      }
    })
    // 流结束后，确保最后一批内容已刷新
    if (rafBatchId) {
      cancelAnimationFrame(rafBatchId)
      rafBatchId = 0
    }
    flushBatch()
  } catch (error) {
    // 如果完全没有内容，显示错误提示
    if (!assistantMsg.content) {
      assistantMsg.content = '抱歉，服务暂时不可用，请稍后再试。'
      messages.value = [...messages.value]
    }
  } finally {
    isLoading.value = false
    await nextTick()
    scrollToBottom()
  }
}

function handleQuickAction(query: string) {
  inputMessage.value = query
  handleSend()
}

/** 清除聊天历史，开始新对话 */
async function handleClearHistory() {
  messages.value = [
    {
      role: 'assistant',
      content: `您好！我是${store.shopName}的智能助手，请问有什么可以帮助您的？`,
      timestamp: Date.now()
    }
  ]
  try {
    await clearChatHistory()
  } catch {
    // 静默失败
  }
}

/* ────────── 消息元素 ref ────────── */

function setMessageRef(key: number, el: any) {
  if (el && el instanceof HTMLElement) {
    messageRefs.value.set(key, el)
  } else if (!el) {
    messageRefs.value.delete(key)
  }
}

/* ────────── 滚动拉扯动效 ────────── */

let lastScrollTop = 0
let lastScrollTime = 0
let velocity = 0
let rafId = 0
let offsets = new Map<number, number>() // msg.timestamp → current translateY

/**
 * 估算消息元素到滚动视口中心的距离系数（0~1）
 * 越靠近视口中心系数越大，拉扯越明显
 */
function getVisibilityFactor(el: HTMLElement, container: HTMLElement): number {
  const containerRect = container.getBoundingClientRect()
  const elRect = el.getBoundingClientRect()
  const containerCenter = containerRect.top + containerRect.height / 2
  const elCenter = elRect.top + elRect.height / 2
  const dist = Math.abs(elCenter - containerCenter)
  const maxDist = containerRect.height
  return Math.max(0, 1 - dist / maxDist)
}

/** 动画帧：不断衰减并回弹偏移 */
function tick() {
  const container = messagesContainer.value
  if (!container) {
    rafId = requestAnimationFrame(tick)
    return
  }

  // 平滑衰减速度
  velocity *= 0.92
  if (Math.abs(velocity) < 0.5) velocity = 0

  let hasOffset = false

  messageRefs.value.forEach((el, key) => {
    const factor = getVisibilityFactor(el, container)
    // 目标偏移 = 速度 * 距离系数 * 系数
    const target = velocity * factor * 8

    // 当前实际偏移
    const current = offsets.get(key) || 0
    // 向目标缓动（惯性 + 回弹）
    const next = current + (target - current) * 0.08

    if (Math.abs(next) > 0.3) {
      hasOffset = true
      el.style.transform = `translateY(${next.toFixed(1)}px)`
      el.style.transition = 'none'
      offsets.set(key, next)
    } else {
      if (Math.abs(current) > 0.3) {
        // 回正
        el.style.transform = ''
        el.style.transition = 'transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1)'
      }
      offsets.set(key, 0)
    }
  })

  // 如果没有偏移且速度为 0，停止动画循环
  if (!hasOffset && Math.abs(velocity) < 0.5) {
    isScrolling.value = false
    return
  }

  rafId = requestAnimationFrame(tick)
}

/** 容器滚动处理 */
function handleScroll() {
  const container = messagesContainer.value
  if (!container) return

  const now = performance.now()
  const deltaTime = now - lastScrollTime
  const deltaY = container.scrollTop - lastScrollTop

  // 计算速度（px/ms）
  if (deltaTime > 0) {
    velocity = deltaY / deltaTime * 16 // 归一化到约 60fps
    // 限幅
    velocity = Math.max(-30, Math.min(30, velocity))
  }

  lastScrollTop = container.scrollTop
  lastScrollTime = now

  if (!isScrolling.value) {
    isScrolling.value = true
    rafId = requestAnimationFrame(tick)
  }
}
</script>

<style scoped>
.agent-panel {
  position: fixed;
  inset: 0;
  z-index: 200;
  pointer-events: none;
}

.agent-panel.open {
  pointer-events: auto;
}

.agent-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.agent-panel.open .agent-overlay {
  background: rgba(0, 0, 0, 0.08);
}

.agent-drawer {
  position: absolute;
  right: -420px;
  top: 10%;
  bottom: 10%;
  width: 420px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  box-shadow: 
    -10px 0 40px rgba(102, 126, 234, 0.15),
    -20px 0 60px rgba(118, 75, 162, 0.1),
    0 0 0 1px rgba(255, 255, 255, 0.2) inset;
  display: flex;
  flex-direction: column;
  transition: right 0.4s cubic-bezier(0.32, 0.72, 0, 1);
  border-radius: 24px;
  overflow: hidden;
}

.agent-panel.open .agent-drawer {
  right: 20px;
}

.agent-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  background-size: 200% 200%;
  animation: headerGradient 8s ease infinite;
  position: relative;
  z-index: 10;
  overflow: hidden;
}

.agent-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  z-index: 2;
}

.agent-clear-btn {
  color: rgba(255, 255, 255, 0.75) !important;
  border-color: rgba(255, 255, 255, 0.25) !important;
  background: rgba(255, 255, 255, 0.1) !important;
  font-size: 0.75rem !important;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s ease;
}

.agent-clear-btn:hover {
  color: #fff !important;
  border-color: rgba(255, 255, 255, 0.5) !important;
  background: rgba(255, 255, 255, 0.2) !important;
}

@keyframes headerGradient {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.floating-bg {
  position: absolute;
  top: -100px;
  right: -100px;
  width: 250px;
  height: 250px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.15) 0%, transparent 70%);
  border-radius: 50%;
  animation: floatBg 6s ease-in-out infinite;
}

@keyframes floatBg {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(-20px, 20px) scale(1.1); }
}

.agent-title {
  display: flex;
  align-items: center;
  gap: 14px;
  position: relative;
  z-index: 1;
}

.agent-icon {
  width: 44px;
  height: 44px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  backdrop-filter: blur(10px);
  position: relative;
  animation: iconFloat 3s ease-in-out infinite;
}

@keyframes iconFloat {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

.icon-glow {
  position: absolute;
  inset: -6px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.4), transparent);
  border-radius: 20px;
  animation: glowPulse 2s ease-in-out infinite;
}

@keyframes glowPulse {
  0%, 100% { opacity: 0.5; transform: scale(1); }
  50% { opacity: 0.8; transform: scale(1.1); }
}

.agent-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.agent-name {
  font-size: 1.125rem;
  font-weight: 700;
  color: #fff;
  letter-spacing: 0.5px;
}

.agent-status {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.85);
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 6px;
  height: 6px;
  background: #34d399;
  border-radius: 50%;
  box-shadow: 0 0 8px #34d399;
  animation: statusPulse 2s ease-in-out infinite;
}

@keyframes statusPulse {
  0%, 100% { opacity: 1; transform: scale(1); box-shadow: 0 0 8px #34d399; }
  50% { opacity: 0.6; transform: scale(1.3); box-shadow: 0 0 16px #34d399; }
}

.agent-close {
  position: relative;
  z-index: 2;
  cursor: pointer;
  border: none;
  outline: none;
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.7);
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.agent-close:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.3);
  transform: rotate(90deg);
}

.agent-close:active {
  transform: rotate(90deg) scale(0.9);
}

.agent-messages {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  position: relative;
  background: linear-gradient(180deg, #fafbfc 0%, #f5f7fa 100%);
}

.animated-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  pointer-events: none;
}

.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.3;
}

.blob-1 {
  width: 200px;
  height: 200px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  top: -50px;
  left: -50px;
  animation: blobFloat1 8s ease-in-out infinite;
}

.blob-2 {
  width: 150px;
  height: 150px;
  background: linear-gradient(135deg, #f093fb, #f5576c);
  bottom: -30px;
  right: -30px;
  animation: blobFloat2 10s ease-in-out infinite;
}

.blob-3 {
  width: 120px;
  height: 120px;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  top: 50%;
  left: 50%;
  animation: blobFloat3 7s ease-in-out infinite;
}

@keyframes blobFloat1 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, 30px) scale(1.1); }
  66% { transform: translate(-20px, 20px) scale(0.9); }
}

@keyframes blobFloat2 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(-30px, -20px) scale(1.15); }
  66% { transform: translate(20px, -30px) scale(0.95); }
}

@keyframes blobFloat3 {
  0%, 100% { transform: translate(-50%, -50%) scale(1); }
  50% { transform: translate(-50%, -50%) scale(1.2); }
}

.message-item {
  display: flex;
  gap: 14px;
  margin-bottom: 24px;
  animation: messageSlideIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  will-change: transform;
}

@keyframes messageSlideIn {
  from { 
    opacity: 0; 
    transform: translateY(15px) scale(0.95); 
  }
  to { 
    opacity: 1; 
    transform: translateY(0) scale(1); 
  }
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.avatar-assistant {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 
    0 4px 12px rgba(102, 126, 234, 0.3),
    0 0 20px rgba(102, 126, 234, 0.1);
  animation: avatarGlow 4s ease-in-out infinite;
}

@keyframes avatarGlow {
  0%, 100% { box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3), 0 0 20px rgba(102, 126, 234, 0.1); }
  50% { box-shadow: 0 4px 16px rgba(102, 126, 234, 0.4), 0 0 30px rgba(102, 126, 234, 0.2); }
}

.avatar-user {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 0.9375rem;
  font-weight: 600;
  box-shadow: 
    0 4px 12px rgba(16, 185, 129, 0.3),
    0 0 20px rgba(16, 185, 129, 0.1);
  overflow: hidden;
}

.avatar-user-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 12px;
}

.message-content {
  max-width: 75%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message-bubble {
  padding: 14px 18px;
  border-radius: 20px;
  font-size: 0.875rem;
  line-height: 1.65;
  color: #1e293b;
  position: relative;
  transition: transform 0.2s ease;
}

.message-item.assistant .message-bubble {
  background: #fff;
  border-radius: 20px 20px 20px 6px;
  box-shadow: 
    0 4px 16px rgba(0, 0, 0, 0.06),
    0 0 0 1px rgba(0, 0, 0, 0.04);
}

.message-item.user .message-bubble {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-radius: 20px 20px 6px 20px;
  box-shadow: 
    0 4px 16px rgba(102, 126, 234, 0.3),
    0 0 24px rgba(102, 126, 234, 0.15);
}

.message-item:hover .message-bubble {
  transform: translateY(-2px);
}

.message-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.message-text strong {
  font-weight: 600;
}

.message-text em {
  font-style: italic;
}

.message-time {
  font-size: 0.6875rem;
  color: #94a3b8;
  margin-left: 10px;
}

.message-item.user .message-time {
  margin-left: 0;
  margin-right: 10px;
  color: rgba(255, 255, 255, 0.5);
}

.typing-indicator {
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 4px 0;
}

.typing-dot {
  width: 8px;
  height: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  animation: typingBounce 1.4s infinite ease-in-out both;
}

.typing-dot:nth-child(1) { animation-delay: -0.32s; }
.typing-dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes typingBounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

.agent-input-area {
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(0, 0, 0, 0.04);
}

.input-wrapper {
  display: flex;
  gap: 10px;
  align-items: center;
  background: #f1f5f9;
  border-radius: 16px;
  padding: 6px;
  border: 2px solid transparent;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.input-wrapper:focus-within {
  border-color: rgba(102, 126, 234, 0.4);
  box-shadow: 
    0 0 0 4px rgba(102, 126, 234, 0.1),
    0 8px 32px rgba(102, 126, 234, 0.1);
  background: #fff;
}

.agent-input {
  flex: 1;
}

.agent-input :deep(.el-input__wrapper) {
  background: transparent;
  border: none;
  box-shadow: none;
  padding: 12px 14px;
}

.agent-input :deep(.el-input__inner) {
  font-size: 0.875rem;
  color: #1e293b;
}

.input-actions {
  display: flex;
  align-items: center;
}

.agent-send-btn {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.agent-send-btn:hover:not(:disabled) {
  transform: scale(1.1) rotate(-5deg);
  box-shadow: 
    0 8px 24px rgba(102, 126, 234, 0.4),
    0 0 0 2px rgba(255, 255, 255, 0.3) inset;
}

.agent-send-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.agent-send-btn:disabled {
  background: #e2e8f0;
  color: #94a3b8;
}

.agent-quick-actions {
  padding: 16px 20px;
  background: rgba(248, 250, 252, 0.6);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(0, 0, 0, 0.03);
}

.quick-label {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 10px;
  display: block;
}

.quick-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.quick-tag {
  font-size: 0.75rem;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(102, 126, 234, 0.2);
  color: #667eea;
  border-radius: 24px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.quick-tag:hover:not(.is-disabled) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-color: transparent;
  transform: translateY(-3px) scale(1.05);
  box-shadow: 
    0 8px 24px rgba(102, 126, 234, 0.4),
    0 0 20px rgba(102, 126, 234, 0.2);
}

.quick-tag.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

::-webkit-scrollbar {
  width: 4px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.3);
  border-radius: 2px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.5);
}

@media (max-width: 768px) {
  .agent-drawer {
    width: 100%;
    right: -100%;
    top: 0;
    bottom: 0;
    border-radius: 0;
    box-shadow: none;
  }

  .agent-panel.open .agent-drawer {
    right: 0;
  }
}
</style>
