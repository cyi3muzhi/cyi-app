<template>
  <div class="order-page">
    <!-- 图片预览弹窗 -->
    <div v-if="previewImage" class="img-preview-modal" @click="closePreview">
      <img :src="previewImage" class="preview-img" />
    </div>

    <!-- ===== 顶部统计卡片 ===== -->
    <div class="stats-row">
      <div class="stat-card stat-card--total">
        <div class="stat-icon"><el-icon :size="22"><Tickets /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">总订单</span>
          <span class="stat-value">{{ statsData.totalCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--pending">
        <div class="stat-icon"><el-icon :size="22"><WarningFilled /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">待处理</span>
          <span class="stat-value">{{ pendingCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--today">
        <div class="stat-icon"><el-icon :size="22"><TrendCharts /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">今日订单</span>
          <span class="stat-value">{{ todayCount }}</span>
        </div>
      </div>
      <div class="stat-card stat-card--refund">
        <div class="stat-icon"><el-icon :size="22"><CircleCloseFilled /></el-icon></div>
        <div class="stat-body">
          <span class="stat-label">退款/售后</span>
          <span class="stat-value">{{ refundCount }}</span>
        </div>
      </div>
    </div>

    <!-- ===== 状态标签页 ===== -->
    <div class="tab-bar-wrap">
      <div
        v-for="tab in tabs"
        :key="tab.key"
        :class="['tab-item', { active: activeTab === tab.key }]"
        @click="switchTab(tab.key)"
      >
        <span class="tab-label">{{ tab.label }}</span>
      </div>
    </div>

    <!-- ===== 筛选栏 ===== -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索订单号..."
          :prefix-icon="Search"
          size="small"
          class="filter-input"
          clearable
          @input="onSearchDebounce"
          @clear="handleSearch"
        />
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="small"
          class="filter-date"
          value-format="YYYY-MM-DD"
          @change="handleDateChange"
        />
      </div>
      <div class="filter-right">
        <el-button size="small" plain :loading="exportLoading" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出订单
        </el-button>
      </div>
    </div>

    <!-- ===== 订单表格 ===== -->
    <el-table
      :data="orders"
      v-loading="loading"
      stripe
      style="width: 100%"
      size="small"
      @expand-change="onExpandChange"
    >
      <el-table-column type="expand" width="40">
        <template #default="{ row }">
          <div class="expand-detail">
            <!-- 加载中 -->
            <div v-if="expandLoading.has(row.id)" class="expand-loading">
              <el-icon class="is-loading"><Loading /></el-icon> 加载中...
            </div>
            <!-- 商品明细 -->
            <table v-else-if="expandProducts[row.id]?.length" class="product-table">
              <thead>
                <tr>
                  <th>商品</th>
                  <th>规格</th>
                  <th class="th-price">单价</th>
                  <th class="th-qty">数量</th>
                  <th class="th-price">小计</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(p, i) in expandProducts[row.id]" :key="i">
                  <td class="cell-product">
                    <img v-if="p.image" :src="p.image" class="cell-img" @error="onExpandImgError($event)" @click="openPreview(p.image)" />
                    <span v-else class="cell-avatar" :style="{ background: nameColor(p.name) }">{{ p.name.charAt(0) }}</span>
                    <span class="cell-name">{{ p.name }}</span>
                  </td>
                  <td class="cell-spec">{{ p.spec || '-' }}</td>
                  <td class="cell-price">&yen;{{ (p.price / 100).toFixed(2) }}</td>
                  <td class="cell-qty">{{ p.qty }}</td>
                  <td class="cell-price">&yen;{{ ((p.price * p.qty) / 100).toFixed(2) }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="expand-empty">暂无商品数据</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="订单号" width="180">
        <template #default="{ row }">
          <span class="order-no-text">{{ row.orderNo }}</span>
        </template>
      </el-table-column>
      <el-table-column label="下单时间" width="155">
        <template #default="{ row }">{{ row.createdAt }}</template>
      </el-table-column>
      <el-table-column label="用户" width="90" show-overflow-tooltip>
        <template #default="{ row }">{{ row.userName }}</template>
      </el-table-column>
      <el-table-column label="商品" min-width="180">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openDetail(row)">
            {{ expandProducts[row.id]?.length ? `${expandProducts[row.id].length}件商品` : '查看商品' }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column label="实付金额" width="120" align="right">
        <template #default="{ row }">
          <span class="amount-text">&yen;{{ (row.amount / 100).toFixed(2) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small" effect="plain">
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openDetail(row)">
            详情
          </el-button>
          <el-button
            v-if="row.status === 'paid'"
            link type="success" size="small"
            @click="openShipDialog(row)"
          >
            发货
          </el-button>
          <el-button
            v-if="row.status === 'pending'"
            link type="danger" size="small"
            @click="handleCancelClick(row)"
          >
            取消
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- ===== 空状态 ===== -->
    <el-empty v-if="!loading && orders.length === 0" description="暂无订单" />

    <!-- ===== 分页 ===== -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="totalCount"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        background
        small
        @current-change="loadOrders"
        @size-change="onSizeChange"
      />
    </div>

<!-- ===== 发货对话框 ===== -->
    <el-dialog
      v-model="shipDialogVisible"
      title="订单发货"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-form ref="shipFormRef" :model="shipForm" :rules="shipRules" label-width="100px">
        <el-form-item label="物流公司" prop="logistics_company">
          <el-select v-model="shipForm.logistics_company" placeholder="请选择物流公司" style="width: 100%">
            <el-option v-for="c in logisticsCompanies" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号" prop="logistics_no">
          <el-input v-model="shipForm.logistics_no" placeholder="请输入快递单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="confirmShip">
          确认发货
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 订单管理列表页面
 * 功能：统计卡片、状态标签页、筛选、列表、展开商品、发货、取消、导出
 */
import { ref, computed, onMounted, onUnmounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, Download,
  Tickets, WarningFilled, TrendCharts, CircleCloseFilled, Loading
} from '@element-plus/icons-vue'
import { listOrder, getOrder, shipOrder, cancelOrder, orderStats, exportOrder, getExportTaskStatus } from '../../api/order'
import type { OrderItem, OrderProduct, OrderStatsResponse } from '../../types/order'

const router = useRouter()

/* ==================== 状态 & 数据 ==================== */

const loading = ref(false)
const orders = ref<OrderItem[]>([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const activeTab = ref('')
const searchKeyword = ref('')
const dateRange = ref<[string, string] | null>(null)

/** 展开行缓存 { orderId: OrderProduct[] } */
const expandProducts = reactive<Record<number, OrderProduct[]>>({})
/** 展开行加载状态 */
const expandLoading = reactive<Set<number>>(new Set())

/** 图片预览 */
const previewImage = ref('')
function openPreview(src: string) {
  previewImage.value = src
}
function closePreview() {
  previewImage.value = ''
}

/** 搜索防抖定时器 */
let searchTimer: ReturnType<typeof setTimeout> | null = null

/* ==================== 统计卡片数据 ==================== */

/** 统计数据（独立接口获取，不随状态切换变化） */
const statsData = ref<OrderStatsResponse>({
  totalCount: 0,
  pendingCount: 0,
  todayCount: 0,
  refundCount: 0,
})

/** 加载统计数据 */
async function loadStats() {
  try {
    const res = await orderStats()
    statsData.value = res
  } catch {
    // 静默失败，统计只是辅助显示
  }
}

/** 待处理数 */
const pendingCount = computed(() => statsData.value.pendingCount)

/** 今日订单数 */
const todayCount = computed(() => statsData.value.todayCount)

/** 退款/售后数 */
const refundCount = computed(() => statsData.value.refundCount)

/* ==================== 标签页 ==================== */

const tabs = computed(() => [
  { key: '', label: '全部' },
  { key: 'pending', label: '待付款' },
  { key: 'paid', label: '待发货' },
  { key: 'shipped', label: '已发货' },
  { key: 'completed', label: '已完成' },
  { key: 'cancelled', label: '已取消' },
  { key: 'refunding', label: '退款中' },
])

/** 标签页切换 */
function switchTab(key: string) {
  activeTab.value = key
  currentPage.value = 1
  loadOrders()
}

/* ==================== 加载数据 ==================== */

/** 加载订单列表 */
async function loadOrders() {
  loading.value = true
  try {
    const res = await listOrder({
      page: currentPage.value,
      page_size: pageSize.value,
      status: activeTab.value,
      keyword: searchKeyword.value,
      date_from: dateRange.value?.[0] || undefined,
      date_to: dateRange.value?.[1] || undefined,
    })
    orders.value = res.items
    totalCount.value = res.total
  } catch {
    orders.value = []
    totalCount.value = 0
  } finally {
    loading.value = false
  }
}

/** 每页条数变更 */
function onSizeChange() {
  currentPage.value = 1
  loadOrders()
}

/* ==================== 筛选 ==================== */

/** 搜索防抖 */
function onSearchDebounce() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => handleSearch(), 300)
}

function handleSearch() {
  currentPage.value = 1
  loadOrders()
}

function handleDateChange() {
  currentPage.value = 1
  loadOrders()
}

/* ==================== 工具函数 ==================== */

/** 状态 → el-tag 类型 */
function statusTagType(status: string): string {
  const map: Record<string, string> = {
    pending: 'danger',
    paid: 'warning',
    shipped: 'primary',
    completed: 'success',
    cancelled: 'info',
    refunding: 'danger',
    refunded: 'info',
  }
  return map[status] || 'info'
}

/** 根据商品名称生成固定颜色 */
function nameColor(name: string): string {
  const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

/** 展开行图片加载失败 */
function onExpandImgError(e: Event) {
  const el = e.target as HTMLElement
  el.style.display = 'none'
}

/** 展开行时异步加载商品明细 */
async function onExpandChange(row: OrderItem, expanded: boolean) {
  if (!expanded || expandProducts[row.id]) return
  expandLoading.add(row.id)
  try {
    const res = await getOrder(row.id)
    expandProducts[row.id] = res.order.products || []
  } catch {
    expandProducts[row.id] = []
  } finally {
    expandLoading.delete(row.id)
  }
}

/** 跳转到订单详情页 */
function openDetail(row: OrderItem) {
  router.push(`/orders/${row.id}`)
}

/* ==================== 发货 ==================== */

const shipDialogVisible = ref(false)
const shipLoading = ref(false)
const shipTarget = ref<OrderItem | null>(null)
const shipFormRef = ref<any>(null)

const shipForm = ref({
  logistics_company: '',
  logistics_no: '',
})

const shipRules = {
  logistics_company: [{ required: true, message: '请选择物流公司', trigger: 'change' }],
  logistics_no: [{ required: true, message: '请输入物流单号', trigger: 'blur' }],
}

const logisticsCompanies = ['顺丰快递', '圆通速递', '中通快递', '韵达快递', '申通快递', '极兔速递', '邮政EMS', '德邦快递']

function openShipDialog(row: OrderItem) {
  shipTarget.value = row
  shipForm.value = { logistics_company: '', logistics_no: '' }
  shipDialogVisible.value = true
}

async function confirmShip() {
  if (!shipFormRef.value) return
  const valid = await shipFormRef.value.validate().catch(() => false)
  if (!valid || !shipTarget.value) return

  shipLoading.value = true
  try {
    await shipOrder(shipTarget.value.id, shipForm.value)
    ElMessage.success(`订单 ${shipTarget.value.orderNo} 发货成功`)
    shipDialogVisible.value = false
    loadOrders()
    loadStats()
    // 通知侧边栏刷新待处理订单徽标
    window.dispatchEvent(new Event('order-changed'))
  } catch {
    // 错误已由拦截器处理
  } finally {
    shipLoading.value = false
  }
}

/* ==================== 取消订单 ==================== */

async function handleCancelClick(row: OrderItem) {
  try {
    await ElMessageBox.confirm(
      `确定要取消订单 ${row.orderNo} 吗？`,
      '取消确认',
      { confirmButtonText: '确定取消', cancelButtonText: '再想想', type: 'warning' }
    )
    await cancelOrder(row.id, '管理员取消')
    ElMessage.success(`订单 ${row.orderNo} 已取消`)
    loadOrders()
    loadStats()
    // 通知侧边栏刷新待处理订单徽标
    window.dispatchEvent(new Event('order-changed'))
  } catch {
    // 用户取消或接口错误
  }
}

/* ==================== 导出 ==================== */

const exportLoading = ref(false)
let exportPollTimer: ReturnType<typeof setInterval> | null = null
let exportPollCount = 0
const MAX_POLL_TIMES = 100 // 最多轮询 100 次（约 5 分钟）

async function handleExport() {
  if (exportLoading.value) return
  exportLoading.value = true
  try {
    const params = {
      status: activeTab.value || undefined,
      date_from: dateRange.value?.[0] || undefined,
      date_to: dateRange.value?.[1] || undefined,
    }
    const res = await exportOrder(params)
    ElMessage.success('导出任务已提交，正在后台处理...')
    // 轮询期间保持 loading，防止重复提交（轮询结束/超时/失败时才解除）
    startExportPolling(res.taskId)
  } catch {
    // 投递失败时解除 loading（轮询未启动）
    exportLoading.value = false
  }
}

function startExportPolling(taskId: string) {
  if (exportPollTimer) clearInterval(exportPollTimer)
  exportPollCount = 0
  exportPollTimer = setInterval(async () => {
    exportPollCount++
    if (exportPollCount > MAX_POLL_TIMES) {
      clearInterval(exportPollTimer!)
      exportPollTimer = null
      exportLoading.value = false
      ElMessage.warning('导出超时，请稍后重试')
      return
    }
    try {
      const status = await getExportTaskStatus(taskId)
      if (status.status === 'COMPLETED' && status.downloadUrl) {
        clearInterval(exportPollTimer!)
        exportPollTimer = null
        exportLoading.value = false
        ElMessage.success('订单导出完成，正在下载...')
        // 用隐藏 <a> 标签触发下载，避免 window.open 打开新标签页导致页面闪烁
        triggerDownload(status.downloadUrl)
      } else if (status.status === 'FAILED') {
        clearInterval(exportPollTimer!)
        exportPollTimer = null
        exportLoading.value = false
        ElMessage.error(`导出失败: ${status.error || '未知错误'}`)
      }
    } catch {
      // 单次查询失败不中断轮询
    }
  }, 3000)
}

/**
 * 触发文件下载，不打开新标签页，避免页面闪烁
 * 原理：动态创建隐藏 <a> 标签并 click，浏览器在当前页直接发起下载
 */
function triggerDownload(url: string) {
  const link = document.createElement('a')
  link.href = url
  link.style.display = 'none'
  // 从 URL 中提取文件名作为下载名（跨域时浏览器可能忽略，但不影响下载行为）
  const fileName = url.split('/').pop()?.split('?')[0] || `orders_${Date.now()}.xlsx`
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onUnmounted(() => {
  if (exportPollTimer) {
    clearInterval(exportPollTimer)
    exportPollTimer = null
    exportLoading.value = false
  }
})

/* ==================== 初始化 ==================== */

onMounted(() => {
  loadOrders()
  loadStats()
})
</script>

<style scoped>
/* ========================================
   订单管理页面样式
   ======================================== */

.order-page {
  padding: 24px;
}

/* ---- 统计卡片 ---- */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: #fff;
  border-radius: var(--shop-radius-lg, 0.75rem);
  border: 1px solid var(--shop-border, #e8e8e8);
  transition: box-shadow var(--shop-transition, 0.2s);
}

.stat-card:hover {
  box-shadow: var(--shop-shadow-md, 0 4px 6px -1px rgb(0 0 0 / 0.07));
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--shop-radius-md, 0.5rem);
}

.stat-card--total .stat-icon {
  background: var(--shop-primary-bg, #eef4ff);
  color: var(--shop-primary, #3b82f6);
}
.stat-card--pending .stat-icon {
  background: hsl(38, 92%, 94%);
  color: var(--shop-warning, #f59e0b);
}
.stat-card--today .stat-icon {
  background: hsl(142, 71%, 94%);
  color: var(--shop-success, #10b981);
}
.stat-card--refund .stat-icon {
  background: hsl(0, 84%, 94%);
  color: var(--shop-destructive, #ef4444);
}

.stat-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground, #64748b);
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.2;
  color: var(--shop-foreground, #1e293b);
}

/* ---- 标签页 ---- */
.tab-bar-wrap {
  display: flex;
  gap: 2px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--shop-border, #e8e8e8);
  background: #fff;
  border-radius: var(--shop-radius-lg, 0.75rem) var(--shop-radius-lg, 0.75rem) 0 0;
  border: 1px solid var(--shop-border, #e8e8e8);
  border-bottom: none;
  padding: 0 4px;
  overflow-x: auto;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 18px;
  cursor: pointer;
  color: var(--shop-muted-foreground, #64748b);
  font-size: 0.875rem;
  border-bottom: 2px solid transparent;
  transition: all var(--shop-transition, 0.2s);
  user-select: none;
  white-space: nowrap;
}

.tab-item:hover {
  color: var(--shop-primary, #3b82f6);
}

.tab-item.active {
  color: var(--shop-primary, #3b82f6);
  border-bottom-color: var(--shop-primary, #3b82f6);
  font-weight: 600;
}

.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 18px;
  padding: 0 6px;
  font-size: 0.6875rem;
  font-weight: 600;
  border-radius: 10px;
  background: var(--shop-muted, #f1f5f9);
  color: var(--shop-muted-foreground, #64748b);
}

.tab-item.active .tab-count {
  background: var(--shop-primary-bg, #eef4ff);
  color: var(--shop-primary, #3b82f6);
}

/* ---- 筛选栏 ---- */
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #fff;
  border: 1px solid var(--shop-border, #e8e8e8);
  border-radius: var(--shop-radius-lg, 0.75rem);
}

.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-input {
  width: 240px;
}

.filter-date {
  width: 260px;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ---- 表格 ---- */
.order-no-text {
  font-weight: 600;
  color: var(--shop-primary, #3b82f6);
  font-family: var(--shop-font-mono, 'JetBrains Mono', monospace);
  font-size: 0.8125rem;
}

.product-summary {
  font-size: 0.8125rem;
  color: var(--shop-foreground, #1e293b);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  max-width: 220px;
}

.amount-text {
  font-weight: 600;
  font-family: var(--shop-font-mono, 'JetBrains Mono', monospace);
  font-size: 0.875rem;
}

/* ---- 展开行商品表格 ---- */
.expand-detail {
  padding: 16px 24px;
  background: var(--shop-muted, #f8fafc);
}

.expand-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 0;
  color: var(--shop-muted-foreground, #64748b);
  font-size: 0.8125rem;
}

.expand-empty {
  padding: 16px 0;
  color: var(--shop-muted-foreground, #64748b);
  font-size: 0.8125rem;
  text-align: center;
}

.product-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.product-table th {
  padding: 8px 12px;
  text-align: left;
  font-weight: 600;
  color: var(--shop-muted-foreground, #64748b);
  border-bottom: 1px solid var(--shop-border, #e8e8e8);
  background: #fff;
}

.product-table .th-price {
  text-align: right;
}

.product-table .th-qty {
  text-align: center;
}

.product-table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--shop-border, #e8e8e8);
}

.cell-product {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cell-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--shop-radius-sm, 0.375rem);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  flex-shrink: 0;
}

.cell-img {
  width: 32px;
  height: 32px;
  border-radius: var(--shop-radius-sm, 0.375rem);
  flex-shrink: 0;
  :deep(img) {
    object-fit: cover;
    width: 100%;
    height: 100%;
    border-radius: inherit;
  }
  /* 预览遮罩层提升层级，避免被表格内容遮挡 */
  :deep(.el-image__preview) {
    z-index: 9999;
  }
}

.cell-name {
  font-weight: 500;
  font-size: 0.875rem;
}

.cell-spec {
  color: var(--shop-muted-foreground, #64748b);
}

.cell-price {
  text-align: right;
  font-family: var(--shop-font-mono, 'JetBrains Mono', monospace);
}

.cell-qty {
  text-align: center;
}

/* ---- 分页 ---- */
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0 8px;
}

/* ---- 响应式 ---- */
@media (max-width: 900px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-left {
    flex-direction: column;
  }
  .filter-input,
  .filter-date {
    width: 100%;
  }
}

/* ---- 图片预览弹窗 ---- */
.img-preview-modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: zoom-out;
}

.preview-img {
  max-width: 90%;
  max-height: 90%;
  object-fit: contain;
  border-radius: 8px;
}
</style>
