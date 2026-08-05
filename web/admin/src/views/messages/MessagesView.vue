<template>
  <!-- 消息中心页面 - 发送通知 + 消息列表 + 详情/下线 -->
  <div class="messages-view">
    <!-- 页面头部 -->
    <header class="page-head">
      <div class="head-left">
        <div class="title-row">
          <span class="title-mark"></span>
          <h2 class="page-title">消息中心</h2>
        </div>
        <p class="page-desc">向用户发送订单、系统、营销通知，支持全体或指定用户</p>
      </div>
      <div class="head-right">
        <el-button size="small" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button size="small" type="primary" @click="openSend">
          <el-icon><Promotion /></el-icon>
          发送消息
        </el-button>
      </div>
    </header>

    <!-- 筛选栏 -->
    <section class="filter-panel">
      <div class="filter-item">
        <label class="filter-label">消息类型</label>
        <el-select v-model="filterType" placeholder="全部类型" size="small" clearable style="width: 120px" @change="handleAutoFilter">
          <el-option label="订单" :value="1" />
          <el-option label="系统" :value="2" />
          <el-option label="营销" :value="3" />
        </el-select>
      </div>
      <div class="filter-item">
        <label class="filter-label">状态</label>
        <el-select v-model="filterStatus" placeholder="全部状态" size="small" clearable style="width: 120px" @change="handleAutoFilter">
          <el-option label="启用" :value="1" />
          <el-option label="已下线" :value="2" />
        </el-select>
      </div>
      <div class="filter-item">
        <label class="filter-label">关键词</label>
        <el-input
          v-model="filterKeyword"
          placeholder="搜索标题..."
          :prefix-icon="Search"
          size="small"
          style="width: 180px"
          clearable
          @input="handleKeywordInput"
        />
      </div>
      <div class="filter-actions">
        <el-button size="small" @click="handleReset">
          <el-icon><Refresh /></el-icon>
          重置
        </el-button>
      </div>
    </section>

    <!-- 消息表格 -->
    <el-card class="table-card" shadow="never">
      <el-table :data="messages" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.msgType)" size="small" effect="light">{{ typeLabel(row.msgType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="内容" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content }}</template>
        </el-table-column>
        <el-table-column label="发送范围" width="100">
          <template #default="{ row }">
            <span :class="row.targetType === 1 ? 'scope-tag scope-user' : 'scope-tag scope-all'">
              {{ row.targetType === 1 ? '指定用户' : '全体用户' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" effect="plain">
              {{ row.status === 1 ? '启用' : '已下线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送时间" width="165">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">查看</el-button>
            <el-button v-if="row.status === 1" link type="danger" size="small" @click="handleDelete(row)">下线</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="table-footer">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchMessages"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <!-- 发送消息弹窗 -->
    <el-dialog v-model="sendVisible" title="发送消息" width="560px" :close-on-click-modal="false">
      <el-form ref="sendFormRef" :model="sendForm" :rules="sendRules" label-width="90px">
        <el-form-item label="消息标题" prop="title">
          <el-input v-model="sendForm.title" placeholder="请输入消息标题" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="消息类型" prop="msg_type">
          <el-radio-group v-model="sendForm.msg_type">
            <el-radio :value="1">订单</el-radio>
            <el-radio :value="2">系统</el-radio>
            <el-radio :value="3">营销</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="发送范围" prop="target_type">
          <el-radio-group v-model="sendForm.target_type" @change="onTargetTypeChange">
            <el-radio :value="0">全体用户</el-radio>
            <el-radio :value="1">指定用户</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="sendForm.target_type === 1" label="用户ID" prop="target_user_ids">
          <el-select
            v-model="sendForm.target_user_ids"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="输入或选择用户ID（可多个）"
            style="width: 100%"
          >
            <el-option
              v-for="uid in recentUserIds"
              :key="uid"
              :label="`用户 ${uid}`"
              :value="uid"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="消息内容" prop="content">
          <el-input
            v-model="sendForm.content"
            type="textarea"
            :rows="5"
            maxlength="5000"
            show-word-limit
            placeholder="请输入消息内容，支持换行"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="sendVisible = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="handleSend">发送</el-button>
      </template>
    </el-dialog>

    <!-- 消息详情弹窗 -->
    <el-dialog v-model="detailVisible" title="消息详情" width="560px">
      <div v-if="detailMsg" class="detail-body">
        <div class="detail-meta">
          <el-tag :type="typeTag(detailMsg.msgType)" size="small">{{ typeLabel(detailMsg.msgType) }}</el-tag>
          <el-tag v-if="detailMsg.targetType === 1" size="small" type="warning" effect="plain">指定用户</el-tag>
          <span class="detail-time">发送时间：{{ formatTime(detailMsg.createdAt) }}</span>
        </div>
        <h3 class="detail-title">{{ detailMsg.title }}</h3>
        <div class="detail-content">{{ detailMsg.content }}</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 消息中心页面
 * 发送消息（全体/指定用户）+ 消息列表 + 详情查看 + 下线
 * 数据来源：/admin/v1/messages
 */
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Promotion, Search } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getMessages,
  sendMessage,
  deleteMessage,
  MSG_TYPE_LABEL,
  type MessageItem,
  type SendMessagePayload
} from '@/api/message'

// ---- 列表状态 ----
const messages = ref<MessageItem[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// ---- 筛选 ----
const filterType = ref<number | undefined>()
const filterStatus = ref<number | undefined>()
const filterKeyword = ref('')

/** 消息类型展示名 */
function typeLabel(type: number): string {
  return MSG_TYPE_LABEL[type] || '通知'
}

/** 消息类型 tag 颜色映射 */
function typeTag(type: number): 'primary' | 'success' | 'warning' | 'info' {
  const map: Record<number, any> = { 1: 'primary', 2: 'success', 3: 'warning' }
  return map[type] || 'info'
}

/** 格式化时间 */
function formatTime(ts: number): string {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 加载消息列表 */
async function fetchMessages() {
  loading.value = true
  try {
    const res = await getMessages({
      page: page.value,
      page_size: pageSize.value,
      msg_type: filterType.value,
      status: filterStatus.value,
      keyword: filterKeyword.value || undefined
    })
    messages.value = res.items
    total.value = res.total
  } catch (e: any) {
    ElMessage.error(e?.message || '加载消息列表失败')
  } finally {
    loading.value = false
  }
}

/** 刷新 */
function handleRefresh() {
  page.value = 1
  fetchMessages()
}

/** 自动筛选 */
function handleAutoFilter() {
  page.value = 1
  fetchMessages()
}

/** 关键词输入（防抖由调用频率控制，简单即时触发） */
function handleKeywordInput() {
  page.value = 1
  fetchMessages()
}

/** 重置筛选 */
function handleReset() {
  filterType.value = undefined
  filterStatus.value = undefined
  filterKeyword.value = ''
  page.value = 1
  fetchMessages()
}

/** 分页大小变化 */
function handleSizeChange() {
  page.value = 1
  fetchMessages()
}

// ---- 发送消息 ----
const sendVisible = ref(false)
const sending = ref(false)
const sendFormRef = ref<FormInstance>()
const sendForm = ref<SendMessagePayload>({
  title: '',
  content: '',
  msg_type: 2,
  target_type: 0,
  target_user_ids: []
})

/** 最近使用的用户ID（保留常用） */
const recentUserIds = ref<number[]>([])

const sendRules: FormRules = {
  title: [{ required: true, message: '请输入消息标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入消息内容', trigger: 'blur' }],
  target_user_ids: [
    {
      validator: (_rule, value: number[], callback) => {
        if (sendForm.value.target_type === 1 && (!value || value.length === 0)) {
          callback(new Error('请选择或输入至少一个用户ID'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

/** 打开发送弹窗 */
function openSend() {
  sendForm.value = { title: '', content: '', msg_type: 2, target_type: 0, target_user_ids: [] }
  sendVisible.value = true
}

/** 切换发送范围时清空用户ID */
function onTargetTypeChange() {
  if (sendForm.value.target_type === 0) {
    sendForm.value.target_user_ids = []
  }
}

/** 提交发送 */
async function handleSend() {
  if (!sendFormRef.value) return
  try {
    await sendFormRef.value.validate()
  } catch {
    return
  }
  sending.value = true
  try {
    const payload: SendMessagePayload = {
      title: sendForm.value.title.trim(),
      content: sendForm.value.content.trim(),
      msg_type: sendForm.value.msg_type,
      target_type: sendForm.value.target_type,
      target_user_ids: sendForm.value.target_type === 1 ? (sendForm.value.target_user_ids || []).map(Number) : []
    }
    await sendMessage(payload)
    // 记录最近使用的用户ID
    if (payload.target_user_ids.length > 0) {
      const merged = [...new Set([...payload.target_user_ids, ...recentUserIds.value])].slice(0, 20)
      recentUserIds.value = merged
    }
    ElMessage.success('消息发送成功')
    sendVisible.value = false
    fetchMessages()
  } catch (e: any) {
    ElMessage.error(e?.message || '发送失败，请重试')
  } finally {
    sending.value = false
  }
}

// ---- 详情 / 下线 ----
const detailVisible = ref(false)
const detailMsg = ref<MessageItem | null>(null)

/** 查看消息详情 */
function openDetail(row: MessageItem) {
  detailMsg.value = row
  detailVisible.value = true
}

/** 下线消息 */
async function handleDelete(row: MessageItem) {
  try {
    await ElMessageBox.confirm(
      `确定下线消息「${row.title}」吗？下线后用户将无法再看到该消息。`,
      '下线确认',
      { type: 'warning', confirmButtonText: '确认下线', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await deleteMessage(row.id)
    ElMessage.success('消息已下线')
    fetchMessages()
  } catch (e: any) {
    ElMessage.error(e?.message || '下线失败，请重试')
  }
}

onMounted(fetchMessages)
</script>

<style scoped>
/* ============================================
   消息中心页面样式
   ============================================ */

/* ---- 页面头部 ---- */
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-mark {
  width: 4px;
  height: 20px;
  border-radius: 2px;
  background: var(--shop-primary);
}

.page-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0;
}

.page-desc {
  margin: 6px 0 0 14px;
  font-size: 0.8125rem;
  color: var(--el-text-color-secondary);
}

.head-right {
  display: flex;
  gap: 10px;
}

/* ---- 筛选栏 ---- */
.filter-panel {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 16px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-label {
  font-size: 0.75rem;
  color: var(--el-text-color-secondary);
}

.filter-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

/* ---- 表格 ---- */
.table-card {
  border-radius: 10px;
}

.table-card :deep(.el-card__body) {
  padding: 12px 16px 8px;
}

.table-footer {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0 8px;
}

/* 发送范围标签 */
.scope-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.scope-all {
  color: var(--el-color-success);
  background: var(--el-color-success-light-9);
}

.scope-user {
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
}

/* ---- 详情弹窗 ---- */
.detail-body {
  padding: 4px 0;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.detail-time {
  font-size: 0.75rem;
  color: var(--el-text-color-secondary);
}

.detail-title {
  margin: 14px 0 10px;
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
  word-break: break-word;
}

.detail-content {
  font-size: 0.875rem;
  line-height: 1.8;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 14px 16px;
}
</style>
