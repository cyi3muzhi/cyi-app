<template>
  <!-- 操作日志页面 - 统计卡片 + 筛选栏 + 日志表格 + 详情弹窗 + 分页 -->
  <div class="logs-view">
    <!-- 页面头部 -->
    <header class="page-head">
      <div class="head-left">
        <div class="title-row">
          <span class="title-mark"></span>
          <h2 class="page-title">操作日志</h2>
        </div>
        <p class="page-desc">记录管理员的关键操作，追溯所有后台行为</p>
      </div>
      <div class="head-right">
        <span class="live-badge">
          <i class="live-dot"></i>
          实时记录中
        </span>
        <el-button size="small" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </header>

    <!-- 统计卡片 -->
    <section class="stats-grid">
      <article
        class="stat-card"
        v-for="stat in statsCards"
        :key="stat.label"
        :class="stat.color"
      >
        <div class="stat-top">
          <span class="stat-icon">
            <el-icon :size="17"><component :is="stat.icon" /></el-icon>
          </span>
          <span class="stat-caption">{{ stat.caption }}</span>
        </div>
        <div class="stat-value">{{ stat.value }}</div>
        <div class="stat-label">{{ stat.label }}</div>
        <div class="stat-bar"></div>
      </article>
    </section>

    <!-- 筛选栏 -->
    <section class="filter-panel">
      <div class="filter-item">
        <label class="filter-label">开始日期</label>
        <el-date-picker
          v-model="filterStartDate"
          type="date"
          placeholder="选择开始日期"
          size="small"
          style="width: 140px"
          value-format="YYYY-MM-DD"
          clearable
          @change="handleAutoFilter"
        />
      </div>
      <div class="filter-item">
        <label class="filter-label">结束日期</label>
        <el-date-picker
          v-model="filterEndDate"
          type="date"
          placeholder="选择结束日期"
          size="small"
          style="width: 140px"
          value-format="YYYY-MM-DD"
          clearable
          @change="handleAutoFilter"
        />
      </div>
      <div class="filter-item">
        <label class="filter-label">操作人</label>
        <el-select v-model="filterOperator" placeholder="全部操作人" size="small" clearable style="width: 130px" @change="handleAutoFilter">
          <el-option
            v-for="op in operatorOptions"
            :key="op.value"
            :label="op.label"
            :value="op.value"
          />
        </el-select>
      </div>
      <div class="filter-item">
        <label class="filter-label">操作类型</label>
        <el-select v-model="filterType" placeholder="全部类型" size="small" clearable style="width: 120px" @change="handleAutoFilter">
          <el-option
            v-for="tp in typeOptions"
            :key="tp.value"
            :label="tp.label"
            :value="tp.value"
          />
        </el-select>
      </div>
      <div class="filter-item">
        <label class="filter-label">关键词</label>
        <el-input
          v-model="filterKeyword"
          placeholder="搜索操作内容..."
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
        <el-button size="small" @click="handleExport" :disabled="logs.length === 0">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </section>

    <!-- 日志表格 -->
    <el-card class="table-card" shadow="never">
      <el-table
        :data="logs"
        stripe
        style="width: 100%"
        size="small"
        v-loading="loading"
      >
        <el-table-column label="操作时间" width="180">
          <template #default="{ row }">
            <div class="time-cell">
              <i v-if="isToday(row.created_at)" class="today-dot" title="今天"></i>
              <span class="time-text" :class="{ 'is-today': isToday(row.created_at) }">
                {{ formatTime(row.created_at) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作人" width="150">
          <template #default="{ row }">
            <div class="op-cell">
              <span class="op-avatar" :class="{ 'is-img': isHttpAvatar(row.avatar) }" :style="{ background: avatarColor(row.operator) }">
                <img v-if="isHttpAvatar(row.avatar)" :src="row.avatar" class="op-avatar-img" alt="" />
                <template v-else>{{ (row.operator || '?').slice(0, 1) }}</template>
              </span>
              <span class="op-name">{{ row.operator || '—' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="110">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small" effect="plain">
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作类型" width="110">
          <template #default="{ row }">
            <span class="module-tag" :class="moduleClass(row.module)">
              {{ typeLabel(row.module) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="IP 地址" width="140">
          <template #default="{ row }">
            <span class="mono-text">{{ row.ip || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作内容" min-width="280" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="content-text">{{ row.content || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="handleViewDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next, jumper"
        background
        small
        @current-change="loadLogs"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="日志详情"
      width="660px"
      align-center
      append-to-body
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="detail-body" v-if="currentLog">
        <!-- 基本信息 -->
        <div class="detail-group">
          <div class="group-title">基本信息</div>
          <div class="detail-grid">
            <div class="detail-row">
              <span class="detail-label">操作时间</span>
              <span class="detail-value">{{ formatTime(currentLog.created_at) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">操作人</span>
              <span class="detail-value">
                <div class="op-cell detail-op">
                  <span class="op-avatar sm" :class="{ 'is-img': isHttpAvatar(currentLog.avatar) }" :style="{ background: avatarColor(currentLog.operator) }">
                    <img v-if="isHttpAvatar(currentLog.avatar)" :src="currentLog.avatar" class="op-avatar-img" alt="" />
                    <template v-else>{{ (currentLog.operator || '?').slice(0, 1) }}</template>
                  </span>
                  <span class="op-name">{{ currentLog.operator || '—' }}</span>
                </div>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">角色</span>
              <span class="detail-value">
                <el-tag :type="roleTagType(currentLog.role)" size="small" effect="plain">
                  {{ roleLabel(currentLog.role) }}
                </el-tag>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">操作类型</span>
              <span class="detail-value">
                <span class="module-tag" :class="moduleClass(currentLog.module)">
                  {{ typeLabel(currentLog.module) }}
                </span>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">IP 地址</span>
              <span class="detail-value mono-text">{{ currentLog.ip || '—' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">执行结果</span>
              <span class="detail-value">
                <el-tag
                  :type="currentLog.result === 1 ? 'success' : 'danger'"
                  size="small"
                  effect="plain"
                >
                  {{ currentLog.result === 1 ? '成功' : '失败' }}
                </el-tag>
              </span>
            </div>
          </div>
        </div>

        <!-- 请求详情 -->
        <div class="detail-group">
          <div class="group-title">请求详情</div>
          <div class="detail-grid">
            <div class="detail-row detail-row-full">
              <span class="detail-label">操作内容</span>
              <span class="detail-value">{{ currentLog.content || '—' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">请求方法</span>
              <span class="detail-value">
                <span class="method-chip" :class="`method-${currentLog.method}`">
                  {{ currentLog.method || '—' }}
                </span>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">记录 ID</span>
              <span class="detail-value mono-text">{{ currentLog.id }}</span>
            </div>
            <div class="detail-row detail-row-full">
              <span class="detail-label">请求路径</span>
              <span class="detail-value mono-text">{{ currentLog.path || '—' }}</span>
            </div>
            <div class="detail-row detail-row-full">
              <span class="detail-label">User Agent</span>
              <span class="detail-value mono-text ua-text">{{ currentLog.user_agent || '—' }}</span>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button size="small" @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 操作日志管理页面
 * 数据来源：后端 NATS 落库，本页通过 GET /admin/v1/logs 系列接口查询
 * 支持按日期范围、操作人、操作类型、关键词筛选，查看详情弹窗，导出 CSV
 */
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download, Refresh, View, Clock, Calendar, List } from '@element-plus/icons-vue'
import { getLogs, getLogStats, getLogOperators, type LogItem } from '../../api/logs'

// ==================== 响应式状态 ====================

/** 日志列表 */
const logs = ref<LogItem[]>([])
/** 总数（服务端分页） */
const total = ref(0)
/** 加载状态 */
const loading = ref(false)
/** 当前页 */
const currentPage = ref(1)
/** 每页条数 */
const pageSize = ref(10)

/** 开始日期筛选 */
const filterStartDate = ref('')
/** 结束日期筛选 */
const filterEndDate = ref('')
/** 操作人筛选 */
const filterOperator = ref('')
/** 操作类型筛选 */
const filterType = ref('')
/** 关键词筛选 */
const filterKeyword = ref('')

/** 详情弹窗可见性 */
const detailVisible = ref(false)
/** 当前查看的日志 */
const currentLog = ref<LogItem | null>(null)

/** 操作人下拉选项 */
const operatorOptions = ref<{ label: string; value: string }[]>([])

/** 操作类型下拉选项（与后端 module 枚举一致） */
const typeOptions = [
  { label: '商品', value: 'product' },
  { label: '订单', value: 'order' },
  { label: '用户', value: 'user' },
  { label: '优惠券', value: 'coupon' },
  { label: '分类', value: 'category' },
  { label: 'Banner', value: 'banner' },
  { label: '员工', value: 'staff' },
  { label: '角色', value: 'role' },
  { label: '设置', value: 'setting' }
]

/** 操作类型中文标签映射 */
const typeLabelMap: Record<string, string> = {
  product: '商品',
  order: '订单',
  user: '用户',
  coupon: '优惠券',
  category: '分类',
  banner: 'Banner',
  staff: '员工',
  role: '角色',
  setting: '设置',
  other: '其他'
}

/** 角色标识 → 中文映射 */
const roleLabelMap: Record<string, string> = {
  super_admin: '超级管理员',
  admin: '管理员',
  operator: '运营'
}

// ==================== 统计卡片 ====================

const logStats = ref({ today: 0, week: 0, month: 0, total: 0 })

const statsCards = computed(() => [
  { label: '今日操作', caption: '今日零点起', value: logStats.value.today.toLocaleString(), icon: Clock, color: 'blue' },
  { label: '近7天操作', caption: '最近 7 天', value: logStats.value.week.toLocaleString(), icon: Calendar, color: 'green' },
  { label: '近30天操作', caption: '最近 30 天', value: logStats.value.month.toLocaleString(), icon: List, color: 'orange' },
  { label: '累计操作', caption: '全部历史记录', value: logStats.value.total.toLocaleString(), icon: View, color: 'purple' }
])

// ==================== 数据加载 ====================

/** 加载日志列表（服务端分页 + 筛选） */
async function loadLogs() {
  loading.value = true
  try {
    const res = await getLogs({
      page: currentPage.value,
      page_size: pageSize.value,
      start_date: filterStartDate.value || undefined,
      end_date: filterEndDate.value || undefined,
      operator: filterOperator.value || undefined,
      module: filterType.value || undefined,
      keyword: filterKeyword.value || undefined
    })
    logs.value = res.items || []
    total.value = res.total || 0
  } catch {
    ElMessage.error('获取操作日志失败')
  } finally {
    loading.value = false
  }
}

/** 加载统计卡片 */
async function loadStats() {
  try {
    logStats.value = await getLogStats()
  } catch {
    // 统计失败不阻塞主流程
  }
}

/** 加载操作人下拉选项 */
async function loadOperators() {
  try {
    const res = await getLogOperators()
    operatorOptions.value = (res.operators || []).map(name => ({ label: name, value: name }))
  } catch {
    // 忽略
  }
}

/** 手动刷新：重新拉取列表与统计 */
function handleRefresh() {
  loadLogs()
  loadStats()
}

// ==================== 辅助函数 ====================

/** 时间戳（秒）格式化为日期时间 */
function formatTime(ts: number): string {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/** 判断时间戳是否为今天 */
function isToday(ts: number): boolean {
  if (!ts) return false
  const d = new Date(ts * 1000)
  const n = new Date()
  return d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()
}

/** 判断头像是否为可访问的 HTTP 图片 URL */
function isHttpAvatar(avatar: string): boolean {
  return !!avatar && avatar.startsWith('http')
}

/** 根据名字生成稳定的头像底色（hash 到调色板） */
function avatarColor(name: string): string {
  if (!name) return '#94a3b8'
  const palette = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#f97316']
  let h = 0
  for (let i = 0; i < name.length; i++) h = (h * 31 + name.charCodeAt(i)) >>> 0
  return palette[h % palette.length]
}

/** 操作类型中文标签 */
function typeLabel(type: string): string {
  return typeLabelMap[type] || type || '—'
}

/** 模块标签样式类名（每种模块一套专属配色） */
function moduleClass(type: string): string {
  return 'mod-' + (type || 'other')
}

/** 角色中文名称 */
function roleLabel(role: string): string {
  return roleLabelMap[role] || role || '—'
}

/** 角色对应的 el-tag 颜色 */
function roleTagType(role: string): string {
  const map: Record<string, string> = {
    super_admin: 'danger',
    admin: 'primary',
    operator: 'success'
  }
  return map[role] || 'info'
}

// ==================== 事件处理 ====================

/** 关键词搜索防抖定时器 */
let searchTimer: ReturnType<typeof setTimeout> | null = null

/** 下拉/日期变化时自动筛选（回到第一页重新加载） */
function handleAutoFilter() {
  currentPage.value = 1
  loadLogs()
}

/** 关键词输入防抖 300ms 后自动搜索 */
function handleKeywordInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadLogs()
  }, 300)
}

/** 重置所有筛选条件 */
function handleReset() {
  filterStartDate.value = ''
  filterEndDate.value = ''
  filterOperator.value = ''
  filterType.value = ''
  filterKeyword.value = ''
  currentPage.value = 1
  loadLogs()
}

/** 每页条数变化时回到第一页并重新加载 */
function handlePageSizeChange() {
  currentPage.value = 1
  loadLogs()
}

/** 查看日志详情 */
function handleViewDetail(row: LogItem) {
  currentLog.value = row
  detailVisible.value = true
}

/** 导出日志 CSV（基于当前页数据） */
function handleExport() {
  const headers = ['操作时间', '操作人', '角色', '操作类型', 'IP地址', '操作内容', '请求方法', '请求路径', '结果']
  const rows = logs.value.map(log => [
    formatTime(log.created_at),
    log.operator,
    roleLabel(log.role),
    typeLabel(log.module),
    log.ip,
    log.content,
    log.method,
    log.path,
    log.result === 1 ? '成功' : '失败'
  ])

  const csv = [headers, ...rows]
    .map(r => r.map(c => `"${c}"`).join(','))
    .join('\n')

  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `操作日志_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)

  ElMessage.success('日志导出成功')
}

onMounted(() => {
  loadLogs()
  loadStats()
  loadOperators()
})
</script>

<style scoped>
/* ---- 动效 ---- */
@keyframes logsFadeUp {
  from { opacity: 0; transform: translateY(18px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes logsCardIn {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes livePulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.45); }
  50% { box-shadow: 0 0 0 5px rgba(16, 185, 129, 0); }
}

.logs-view {
  width: 100%;
  animation: logsFadeUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

/* ---- 页面头部 ---- */
.page-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 1.5rem;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-mark {
  width: 9px;
  height: 26px;
  border-radius: 5px;
  background: linear-gradient(180deg, var(--shop-primary), hsl(217, 91%, 68%));
  box-shadow: 0 2px 6px hsl(217 91% 52% / 0.35);
}

.page-title {
  font-size: 1.375rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  color: var(--shop-foreground);
  margin: 0;
}

.page-desc {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin: 4px 0 0 19px;
}

.head-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.live-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 6px 13px;
  border-radius: 999px;
  background: #fff;
  border: 1px solid var(--shop-border);
  box-shadow: var(--shop-shadow-sm);
  color: var(--shop-success);
  font-size: 0.75rem;
  font-weight: 600;
}

.live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--shop-success);
  animation: livePulse 1.8s ease-out infinite;
}

/* ---- 统计卡片 ---- */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  position: relative;
  overflow: hidden;
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 18px 20px 16px;
  box-shadow: var(--shop-shadow-sm);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  animation: logsCardIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.stat-card:nth-child(1) { animation-delay: 0.05s; }
.stat-card:nth-child(2) { animation-delay: 0.1s; }
.stat-card:nth-child(3) { animation-delay: 0.15s; }
.stat-card:nth-child(4) { animation-delay: 0.2s; }

.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shop-shadow-lg);
}

/* 底部色带 */
.stat-bar {
  position: absolute;
  left: 20px;
  right: 20px;
  bottom: 0;
  height: 3px;
  border-radius: 3px 3px 0 0;
  background: linear-gradient(90deg, var(--stat-c), transparent 85%);
  opacity: 0.9;
}

.stat-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.stat-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-caption {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 800;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  background: linear-gradient(135deg, var(--stat-c), hsl(220, 14%, 30%));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.stat-label {
  margin-top: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  letter-spacing: 0.03em;
}

.stat-card.blue   { --stat-c: hsl(217, 91%, 52%); }
.stat-card.green  { --stat-c: hsl(142, 71%, 40%); }
.stat-card.orange { --stat-c: hsl(38, 92%, 48%); }
.stat-card.purple { --stat-c: hsl(262, 83%, 55%); }

.stat-card.blue .stat-icon   { background: hsl(217, 91%, 96%); color: hsl(217, 91%, 52%); }
.stat-card.green .stat-icon  { background: hsl(142, 71%, 95%); color: hsl(142, 71%, 40%); }
.stat-card.orange .stat-icon { background: hsl(38, 92%, 95%); color: hsl(38, 92%, 48%); }
.stat-card.purple .stat-icon { background: hsl(262, 83%, 95%); color: hsl(262, 83%, 55%); }

/* ---- 筛选栏 ---- */
.filter-panel {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 14px 16px;
  margin-bottom: 16px;
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 14px 16px;
  box-shadow: var(--shop-shadow-sm);
  animation: logsCardIn 0.4s ease both;
  animation-delay: 0.25s;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.filter-label {
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  letter-spacing: 0.02em;
}

.filter-actions {
  margin-left: auto;
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

/* ---- 表格卡片 ---- */
.table-card {
  border-radius: var(--shop-radius-lg);
  box-shadow: var(--shop-shadow-sm);
  animation: logsCardIn 0.4s ease both;
  animation-delay: 0.35s;
}

.table-card :deep(.el-card__body) {
  padding: 0;
}

/* 表头 */
.table-card :deep(.el-table__header th) {
  background: #f8fafc;
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

/* 行 hover */
.table-card :deep(.el-table__row:hover > td) {
  background: hsl(217, 91%, 97%) !important;
}

/* ---- 操作时间 ---- */
.time-cell {
  display: flex;
  align-items: center;
  gap: 7px;
}

.time-text {
  font-family: var(--shop-font-mono);
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.time-text.is-today {
  color: var(--shop-primary);
  font-weight: 600;
}

.today-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--shop-primary);
  box-shadow: 0 0 0 3px hsl(217 91% 52% / 0.15);
}

/* ---- 操作人 ---- */
.op-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.op-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  flex-shrink: 0;
  user-select: none;
}

.op-avatar.sm {
  width: 22px;
  height: 22px;
  font-size: 0.66rem;
}

.op-avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.op-name {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

/* ---- 模块标签（每模块专属配色） ---- */
.module-tag {
  --mod-c: hsl(220, 10%, 45%);
  --mod-bg: hsl(220, 10%, 95%);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 600;
  line-height: 1.4;
  color: var(--mod-c);
  background: var(--mod-bg);
}

.module-tag::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--mod-c);
}

.mod-product  { --mod-c: hsl(217, 91%, 46%); --mod-bg: hsl(217, 91%, 96%); }
.mod-order    { --mod-c: hsl(142, 71%, 36%); --mod-bg: hsl(142, 71%, 95%); }
.mod-user     { --mod-c: hsl(262, 83%, 52%); --mod-bg: hsl(262, 83%, 95%); }
.mod-coupon   { --mod-c: hsl(0, 84%, 52%);   --mod-bg: hsl(0, 84%, 95%); }
.mod-category { --mod-c: hsl(190, 90%, 34%); --mod-bg: hsl(190, 90%, 95%); }
.mod-banner   { --mod-c: hsl(280, 60%, 52%); --mod-bg: hsl(280, 60%, 95%); }
.mod-staff    { --mod-c: hsl(330, 80%, 47%); --mod-bg: hsl(330, 80%, 95%); }
.mod-role     { --mod-c: hsl(38, 92%, 42%);  --mod-bg: hsl(38, 92%, 94%); }
.mod-setting  { --mod-c: hsl(220, 10%, 42%); --mod-bg: hsl(220, 10%, 95%); }
.mod-other    { --mod-c: hsl(220, 10%, 42%); --mod-bg: hsl(220, 10%, 95%); }

/* ---- 操作内容 ---- */
.content-text {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

/* ---- 等宽字体 ---- */
.mono-text {
  font-family: var(--shop-font-mono);
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.ua-text {
  font-size: 0.6875rem;
  word-break: break-all;
  line-height: 1.5;
}

/* ---- 方法徽章 ---- */
.method-chip {
  display: inline-flex;
  padding: 2px 9px;
  border-radius: 6px;
  font-family: var(--shop-font-mono);
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.method-GET    { color: #059669; background: #ecfdf5; }
.method-POST   { color: #2563eb; background: #eff6ff; }
.method-PUT    { color: #d97706; background: #fffbeb; }
.method-PATCH  { color: #7c3aed; background: #f5f3ff; }
.method-DELETE { color: #dc2626; background: #fef2f2; }

/* ---- 分页 ---- */
.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

/* ---- 详情弹窗 ---- */
.detail-body {
  max-height: calc(100vh - 230px);
  overflow-y: auto;
  padding-right: 6px;
}

.detail-group + .detail-group {
  margin-top: 20px;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  font-size: 0.78rem;
  font-weight: 700;
  color: #64748b;
  letter-spacing: 0.04em;
}

.group-title::before {
  content: '';
  width: 4px;
  height: 14px;
  border-radius: 2px;
  background: linear-gradient(180deg, var(--shop-primary), hsl(217, 91%, 68%));
}

.group-title::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--shop-border);
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px 24px;
}

.detail-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-row-full {
  grid-column: 1 / -1;
}

.detail-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  letter-spacing: 0.04em;
}

.detail-label::before {
  content: '';
  width: 3px;
  height: 11px;
  border-radius: 2px;
  background: hsl(217, 91%, 78%);
}

.detail-value {
  font-size: 0.875rem;
  color: var(--shop-foreground);
  line-height: 1.5;
}

.detail-op {
  padding-top: 2px;
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .filter-actions {
    margin-left: 0;
    width: 100%;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>