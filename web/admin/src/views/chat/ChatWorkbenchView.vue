<template>
  <div class="workbench-page">
    <!-- 顶部统计 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon blue"><el-icon :size="18"><ChatDotRound /></el-icon></div>
        <div class="stat-info">
          <div class="stat-label">今日接入</div>
          <div class="stat-value">{{ stats.totalToday }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green"><el-icon :size="18"><Headset /></el-icon></div>
        <div class="stat-info">
          <div class="stat-label">进行中会话</div>
          <div class="stat-value">{{ stats.processing }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon orange"><el-icon :size="18"><CircleCheck /></el-icon></div>
        <div class="stat-info">
          <div class="stat-label">今日已解决</div>
          <div class="stat-value">{{ stats.resolvedToday }}</div>
        </div>
      </div>
      <div class="stat-card ws-state" :class="wsConnected ? 'on' : 'off'">
        <div class="ws-dot"></div>
        <div class="stat-info">
          <div class="stat-label">实时连接</div>
          <div class="stat-value small">{{ wsConnected ? '已连接' : '重连中…' }}</div>
        </div>
      </div>
    </div>

    <div class="workbench-body">
      <!-- 左侧会话列表 -->
      <div class="conv-panel">
        <el-tabs v-model="activeTab" class="conv-tabs" @tab-change="switchTab">
          <el-tab-pane label="待接入" name="pending">
            <template #label>
              <span>待接入</span>
              <span v-if="pendingTotal > 0" class="tab-badge">{{ pendingTotal }}</span>
            </template>
          </el-tab-pane>
          <el-tab-pane label="我的会话" name="mine" />
        </el-tabs>
        <div class="conv-list">
          <template v-if="convList.length > 0">
            <div
              v-for="c in convList"
              :key="c.id"
              class="conv-item"
              :class="{ active: currentId === c.id }"
              @click="selectConv(c)"
            >
              <div class="conv-avatar" :style="{ background: avatarColor(c.userId) }">
                {{ (c.userName || '用户').charAt(0) }}
                <span :class="['online-dot', c.userOnline === 1 ? 'on' : 'off']"></span>
              </div>
              <div class="conv-main">
                <div class="conv-top">
                  <span class="conv-name">{{ c.userName || `用户 ${c.userId}` }}</span>
                  <span class="conv-time">{{ formatTime(c.lastMessageAt || c.createdAt) }}</span>
                </div>
                <div class="conv-bottom">
                  <span class="conv-last">{{ c.lastMessage || '暂无消息' }}</span>
                  <span v-if="c.unreadCount > 0" class="conv-unread">{{ c.unreadCount }}</span>
                </div>
              </div>
            </div>
          </template>
          <div v-else class="conv-empty">
            <el-icon :size="40" color="#c0c4cc"><ChatLineRound /></el-icon>
            <p>{{ activeTab === 'pending' ? '暂无待接入会话' : '暂无会话记录' }}</p>
          </div>
        </div>
      </div>

      <!-- 右侧聊天窗口 -->
      <div class="chat-panel">
        <template v-if="currentConv">
          <div class="chat-header">
            <div class="chat-user">
              <span class="chat-name">{{ currentConv.userName || `用户 ${currentConv.userId}` }}</span>
              <span class="chat-status" :class="currentConv.userOnline === 1 ? 'on' : 'off'">
                {{ currentConv.userOnline === 1 ? '在线' : '离线' }}
              </span>
            </div>
            <div class="chat-actions">
              <el-tag v-if="currentConv.status === 1" type="warning" size="small">待接入</el-tag>
              <el-tag v-else-if="currentConv.status === 2" type="success" size="small">进行中</el-tag>
              <el-tag v-else type="info" size="small">已结束</el-tag>
              <el-button v-if="currentConv.status === 1" type="primary" size="small" :loading="acting" @click="accept">
                接入
              </el-button>
              <el-button v-if="currentConv.status === 2" size="small" @click="openTransfer">转接</el-button>
              <el-button v-if="currentConv.status === 2" type="danger" plain size="small" :loading="acting" @click="closeConv">
                结束会话
              </el-button>
            </div>
          </div>
          <div ref="msgBox" class="chat-messages">
            <template v-if="messages.length > 0">
              <div v-for="m in messages" :key="m.id" class="msg-row" :class="m.senderType === 2 ? 'mine' : 'theirs'">
                <!-- 商品卡片消息 -->
                <div v-if="m.msgType === 3" class="msg-card product-card">
                  <div class="card-thumb" v-if="cardOf(m).image">
                    <img :src="cardOf(m).image" alt="" />
                  </div>
                  <div class="card-body">
                    <div class="card-title">{{ cardOf(m).product_name || '商品' }}</div>
                    <div class="card-price">¥{{ cardOf(m).price || '--' }}</div>
                    <div class="card-sub">商品ID：{{ cardOf(m).product_id || m.extra || '--' }}</div>
                  </div>
                </div>
                <!-- 订单卡片消息 -->
                <div v-else-if="m.msgType === 4" class="msg-card order-card">
                  <div class="card-body">
                    <div class="card-title">订单信息</div>
                    <div class="card-sub">订单号：{{ cardOf(m).order_no || m.extra || '--' }}</div>
                    <div class="card-sub">订单ID：{{ cardOf(m).order_id || '--' }}</div>
                  </div>
                </div>
                <!-- 文本/图片/系统消息 -->
                <div v-else class="msg-bubble" :class="[m.senderType === 1 ? 'user' : m.senderType === 2 ? 'agent' : m.senderType === 4 ? 'ai' : 'system']">
                  <span v-if="m.msgType === 5" class="system-text">{{ m.content }}</span>
                  <template v-else>{{ m.content }}</template>
                </div>
                <div class="msg-time">{{ formatTime(m.createdAt) }}</div>
              </div>
            </template>
            <div v-else class="msg-empty">暂无消息，发送第一条消息开始接待</div>
          </div>
          <div v-if="currentConv.status === 2" class="chat-input">
            <el-input
              v-model="inputText"
              type="textarea"
              :rows="2"
              resize="none"
              placeholder="输入回复内容，Enter 发送"
              @keydown.enter.prevent="send"
            />
            <el-button type="primary" :loading="sending" :disabled="!inputText.trim()" @click="send">
              发送
            </el-button>
          </div>
          <div v-else-if="currentConv.status === 1" class="chat-input-disabled">点击「接入」开始接待该会话</div>
          <div v-else class="chat-input-disabled">会话已结束</div>
        </template>
        <div v-else class="chat-empty">
          <el-icon :size="48" color="#c0c4cc"><ChatLineRound /></el-icon>
          <p>从左侧选择会话开始接待</p>
        </div>
      </div>
    </div>

    <!-- 转接对话框 -->
    <el-dialog v-model="transferOpen" title="转接会话" width="420px">
      <el-form label-width="90px">
        <el-form-item label="目标客服ID">
          <el-input v-model="targetAgentId" placeholder="输入接收客服的管理员ID" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="transferOpen = false">取消</el-button>
        <el-button type="primary" :loading="acting" @click="doTransfer">确认转接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 客服工作台：待接入/我的会话 + 实时聊天窗口
 */
import { onMounted, onUnmounted, ref, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Headset, CircleCheck, ChatLineRound } from '@element-plus/icons-vue'
import { useAppStore } from '../../store/app'
import {
  getPendingConversations,
  getAgentConversations,
  getChatMessages,
  getChatStats,
  acceptConversation,
  transferConversation,
  sendChatMessage,
  markChatRead,
  closeChatConversation,
  type ConversationItem,
  type ChatMessageItem
} from '../../api/chat'

const activeTab = ref('pending')
const convList = ref<ConversationItem[]>([])
const pendingTotal = ref(0)
const currentId = ref<number | null>(null)
const currentConv = ref<ConversationItem | null>(null)
const messages = ref<ChatMessageItem[]>([])
const inputText = ref('')
const sending = ref(false)
const acting = ref(false)
const stats = ref({ totalToday: 0, processing: 0, resolvedToday: 0 })
const msgBox = ref<HTMLDivElement | null>(null)
const transferOpen = ref(false)
const targetAgentId = ref('')

let ws: WebSocket | null = null
const wsConnected = ref(false)
const store = useAppStore()
const myAgentId = Number((store.adminInfo as any)?.id ?? 0)

/** 头像颜色 */
const avatarColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#9b59b6']
function avatarColor(seed: number) {
  return avatarColors[seed % avatarColors.length]
}

/** 格式化时间 */
function formatTime(ts: number) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

/** 解析消息扩展数据（商品/订单卡片） */
function cardOf(m: ChatMessageItem): any {
  if (!m.extra) return {}
  try {
    return JSON.parse(m.extra)
  } catch {
    return {}
  }
}

/** 加载列表 */
async function loadList() {
  if (activeTab.value === 'pending') {
    const res = await getPendingConversations({ page: 1, page_size: 100 })
    convList.value = res.items
    pendingTotal.value = res.total
  } else {
    const res = await getAgentConversations({ page: 1, page_size: 100 })
    convList.value = res.items
  }
  // 保持当前选中
  if (currentId.value && !convList.value.find(c => c.id === currentId.value)) {
    currentId.value = null
    currentConv.value = null
    messages.value = []
  }
}

/** 加载统计 */
async function loadStats() {
  try {
    stats.value = await getChatStats()
  } catch {
    // 辅助数据静默降级
  }
}

/** 切换 Tab */
function switchTab() {
  currentId.value = null
  currentConv.value = null
  messages.value = []
  loadList()
}

/** 选中会话 */
async function selectConv(c: ConversationItem) {
  currentId.value = c.id
  currentConv.value = c
  messages.value = []
  await loadMessages(c.id)
  if (c.unreadCount > 0) {
    markChatRead(c.id)
    c.unreadCount = 0
  }
}

/** 加载历史消息（最新一页） */
async function loadMessages(conversationId: number) {
  const res = await getChatMessages(conversationId, { limit: 50, direction: 1 })
  messages.value = res.items
  scrollToBottom()
}

/** 滚动到底部 */
function scrollToBottom() {
  nextTick(() => {
    if (msgBox.value) msgBox.value.scrollTop = msgBox.value.scrollHeight
  })
}

/** 接入会话 */
async function accept() {
  if (!currentId.value) return
  acting.value = true
  try {
    await acceptConversation(currentId.value)
    ElMessage.success('已接入会话')
    const c = convList.value.find(x => x.id === currentId.value)
    if (c) c.status = 2
    currentConv.value!.status = 2
    loadStats()
  } catch {
    // 全局拦截器已提示
  } finally {
    acting.value = false
  }
}

/** 转接 */
function openTransfer() {
  transferOpen.value = true
  targetAgentId.value = ''
}

async function doTransfer() {
  if (!currentId.value || !targetAgentId.value) return
  acting.value = true
  try {
    await transferConversation(currentId.value, Number(targetAgentId.value))
    ElMessage.success('已转接')
    transferOpen.value = false
    loadList()
  } catch {
    // 已提示
  } finally {
    acting.value = false
  }
}

/** 结束会话 */
async function closeConv() {
  if (!currentId.value) return
  try {
    await ElMessageBox.confirm('确定结束该会话吗？结束后将无法再发送消息。', '结束会话', { type: 'warning' })
    await closeChatConversation(currentId.value)
    ElMessage.success('会话已结束')
    currentConv.value!.status = 3
    loadStats()
    // 刷新列表：已结束会话从「待接入」移除（若停留残留会导致点开报 403 CONVERSATION_CLOSED）
    await loadList()
  } catch {
    // 取消或失败
  }
}

/** 发送消息（REST 发送，WS 推送统一渲染） */
async function send() {
  const text = inputText.value.trim()
  if (!text || !currentId.value) return
  sending.value = true
  let messageId = 0
  try {
    const reply = await sendChatMessage(currentId.value, { msg_type: 1, content: text })
    messageId = Number(reply.message_id ?? reply.messageId ?? 0)
    inputText.value = ''
    currentConv.value!.lastMessage = text.slice(0, 100)
    // 立即本地回显：后端只把推送事件发给对端，发送方不依赖 WS，避免"发出去没反应"
    if (messageId > 0) {
      messages.value.push({
        id: messageId,
        conversationId: currentId.value,
        senderType: 2,
        senderId: 0,
        msgType: 1,
        content: text,
        createdAt: Math.floor(Date.now() / 1000) // 秒级时间戳，与后端保持一致
      } as ChatMessageItem)
      scrollToBottom()
    }
  } catch {
    return
  } finally {
    sending.value = false
  }
  // 兜底轮询：AI/用户回复生成有延迟，递增重拉历史补齐（WS 失效或慢时也能看到回复）
  // 记录发送后最后一条已知消息 id：期间未通过 WS 收到更新才拉取，避免重复覆盖
  const lastId = messages.value.length ? messages.value[messages.value.length - 1].id : 0
  ;[800, 2000, 3500, 5000].forEach(delay => {
    setTimeout(() => {
      const nowLast = messages.value.length ? messages.value[messages.value.length - 1].id : 0
      if (nowLast <= lastId && currentId.value) loadMessages(currentId.value)
    }, delay)
  })
}

/** WebSocket 连接 */
function connectWS() {
  const token = localStorage.getItem('admin_token')
  if (!token) return
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws = new WebSocket(`${proto}://${location.host}/admin/v1/chat/ws?token=${token}`)
  ws.onopen = () => { wsConnected.value = true }
  ws.onclose = () => {
    wsConnected.value = false
    setTimeout(connectWS, 3000)
  }
  ws.onerror = () => { ws.close() }
  ws.onmessage = (e) => {
    try {
      const frame = JSON.parse(e.data)
      if (frame.type === 'message') handleMessage(frame.data)
      else if (frame.type === 'conversation_status') handleConvStatus(frame.data)
    } catch {
      // 忽略脏数据
    }
  }
}

/** 处理推送的新消息 */
function handleMessage(msg: ChatMessageItem) {
  // 渲染当前会话的新消息
  if (currentId.value === msg.conversationId) {
    if (!messages.value.some(m => m.id === msg.id)) {
      messages.value.push(msg)
      scrollToBottom()
    }
  } else {
    // 非当前会话：更新未读角标
    const c = convList.value.find(x => x.id === msg.conversationId)
    if (c) c.unreadCount = (c.unreadCount || 0) + 1
  }
}

/** 处理会话状态变更（新建/接入/转接/结束） */
async function handleConvStatus(data: any) {
  const evt = {
    conversationId: Number(data.conversation_id ?? data.conversationId ?? 0),
    eventType: Number(data.event_type ?? data.eventType ?? 0),
    agentId: Number(data.agent_id ?? data.agentId ?? 0)
  }
  // 刷新列表与统计（简单可靠）
  loadList()
  loadStats()
  // 会话被接入给自己：自动打开
  if (evt.eventType === 2 && evt.agentId === myAgentId && activeTab.value === 'pending') {
    activeTab.value = 'mine'
    await loadList()
    const c = convList.value.find(x => x.id === evt.conversationId)
    if (c) selectConv(c)
  }
}

onMounted(() => {
  loadList()
  loadStats()
  connectWS()
})
onUnmounted(() => {
  if (ws) ws.close()
})
</script>

<style scoped>
.workbench-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

/* 统计 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.stat-icon.blue { background: #409eff; }
.stat-icon.green { background: #67c23a; }
.stat-icon.orange { background: #e6a23c; }
.stat-label { font-size: 12px; color: #909399; }
.stat-value { font-size: 24px; font-weight: 700; color: #303133; line-height: 1.2; }
.stat-value.small { font-size: 15px; font-weight: 600; }
.ws-state .ws-dot {
  width: 10px; height: 10px; border-radius: 50%;
  background: #f56c6c;
  margin: 0 10px;
}
.ws-state.on .ws-dot { background: #67c23a; box-shadow: 0 0 0 4px rgba(103, 194, 58, 0.15); }

/* 主体 */
.workbench-body {
  flex: 1;
  display: flex;
  gap: 16px;
  min-height: 0;
}

/* 会话列表 */
.conv-panel {
  width: 320px;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.conv-tabs { padding: 0 12px; }
.conv-tabs :deep(.el-tabs__nav-wrap::after) { height: 1px; }
.tab-badge {
  display: inline-block;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  padding: 0 4px;
  margin-left: 4px;
  border-radius: 8px;
  background: #f56c6c;
  color: #fff;
  font-size: 11px;
  text-align: center;
}
.conv-list { flex: 1; overflow-y: auto; }
.conv-item {
  display: flex;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f5f5f5;
  transition: background 0.15s;
}
.conv-item:hover { background: #f7f9fc; }
.conv-item.active { background: #ecf5ff; }
.conv-avatar {
  position: relative;
  width: 42px; height: 42px;
  border-radius: 50%;
  color: #fff;
  font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.online-dot {
  position: absolute;
  right: 0; bottom: 0;
  width: 11px; height: 11px;
  border-radius: 50%;
  border: 2px solid #fff;
}
.online-dot.on { background: #67c23a; }
.online-dot.off { background: #c0c4cc; }
.conv-main { flex: 1; min-width: 0; }
.conv-top { display: flex; justify-content: space-between; align-items: center; }
.conv-name { font-size: 14px; font-weight: 600; color: #303133; }
.conv-time { font-size: 12px; color: #c0c4cc; }
.conv-bottom { display: flex; justify-content: space-between; align-items: center; margin-top: 4px; }
.conv-last {
  font-size: 12px; color: #909399;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  max-width: 190px;
}
.conv-unread {
  min-width: 18px; height: 18px; line-height: 18px;
  padding: 0 5px; border-radius: 9px;
  background: #f56c6c; color: #fff;
  font-size: 11px; text-align: center;
}
.conv-empty { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #c0c4cc; gap: 8px; }
.conv-empty p { font-size: 13px; }

/* 聊天窗口 */
.chat-panel {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  min-width: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #f0f2f5;
}
.chat-user { display: flex; align-items: center; gap: 10px; }
.chat-name { font-size: 15px; font-weight: 600; color: #303133; }
.chat-status { font-size: 12px; }
.chat-status.on { color: #67c23a; }
.chat-status.off { color: #c0c4cc; }
.chat-actions { display: flex; align-items: center; gap: 8px; }
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  background: #f7f9fc;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.msg-row { display: flex; flex-direction: column; align-items: flex-start; max-width: 70%; }
.msg-row.mine { align-items: flex-end; align-self: flex-end; }
.msg-bubble {
  padding: 9px 14px;
  border-radius: 10px;
  font-size: 14px;
  line-height: 1.55;
  word-break: break-word;
  white-space: pre-wrap;
}
.msg-bubble.user {
  background: #fff;
  color: #303133;
  border: 1px solid #e4e7ed;
  border-top-left-radius: 2px;
}
.msg-bubble.agent {
  background: #409eff;
  color: #fff;
  border-top-right-radius: 2px;
}
.msg-bubble.ai {
  background: #909399;
  color: #fff;
  border-top-left-radius: 2px;
}
.msg-bubble.system { background: #f0f2f5; color: #909399; font-size: 13px; }
.msg-card {
  display: flex;
  gap: 10px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 10px 12px;
  min-width: 220px;
  max-width: 320px;
}
.product-card { border-color: #e6a23c33; }
.order-card { border-color: #409eff33; }
.card-thumb {
  width: 56px; height: 56px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: #f5f7fa;
}
.card-thumb img { width: 100%; height: 100%; object-fit: cover; }
.card-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.card-title { font-size: 14px; font-weight: 600; color: #303133; }
.card-price { font-size: 15px; font-weight: 700; color: #f56c6c; }
.card-sub { font-size: 12px; color: #909399; word-break: break-all; }
.msg-time { font-size: 11px; color: #c0c4cc; margin-top: 3px; }
.msg-empty, .chat-empty, .chat-input-disabled {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
  font-size: 13px;
}
.chat-input {
  display: flex;
  gap: 10px;
  padding: 12px 16px;
  border-top: 1px solid #f0f2f5;
  align-items: flex-end;
}
.chat-input .el-input { flex: 1; }
</style>
