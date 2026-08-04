<template>
  <!-- 仪表盘页面 - 统计卡片 + 图表 + 最新订单 + 快捷操作 -->
  <div class="dashboard">
    <!-- 4个统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in stats" :key="stat.label">
        <div class="stat-header">
          <span class="stat-label">{{ stat.label }}</span>
          <div :class="['stat-icon', stat.iconClass]">
            <el-icon :size="20"><component :is="stat.icon" /></el-icon>
          </div>
        </div>
        <div class="stat-value">
          {{ stat.prefix }}{{ stat.displayValue }}
          <span v-if="stat.unit" class="stat-unit">{{ stat.unit }}</span>
        </div>
        <div :class="['stat-change', stat.growth >= 0 ? 'up' : 'down']">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <polyline v-if="stat.growth >= 0" points="18 15 12 9 6 15"/>
            <polyline v-else points="6 9 12 15 18 9"/>
          </svg>
          {{ Math.abs(stat.growth) }}%
        </div>
      </div>
    </div>

    <!-- AI 运营摘要 -->
    <el-card class="summary-card" shadow="never">
      <template #header>
        <div class="summary-header">
          <span class="summary-title">AI 运营摘要</span>
          <el-button type="primary" size="small" :loading="summaryLoading" @click="handleGenerateSummary">
            生成运营简报
          </el-button>
        </div>
      </template>
      <div class="summary-body">
        <div v-if="aiSummary" class="summary-text">{{ aiSummary }}</div>
        <div v-else class="summary-empty">
          <el-icon :size="32" color="#cbd5e1"><Document /></el-icon>
          <span class="summary-empty-text">点击「生成运营简报」获取 AI 智能分析</span>
        </div>
      </div>
    </el-card>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <!-- 近14日订单趋势 -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <span class="chart-title">近 14 日订单趋势</span>
        </template>
        <div class="line-chart">
          <div class="bar-wrap" v-for="(val, idx) in orderTrend.values" :key="idx">
            <div class="bar-value">{{ val }}</div>
            <div class="bar" :style="{ height: (val / maxTrendVal) * 100 + '%' }"></div>
            <div class="bar-label">{{ orderTrend.days[idx] }}</div>
          </div>
        </div>
      </el-card>

      <!-- 营收分类占比 -->
      <el-card class="chart-card" shadow="never">
        <template #header>
          <span class="chart-title">营收分类占比</span>
        </template>
        <div class="donut-wrap">
          <div class="donut" :style="{ background: donutGradient }">
            <div class="donut-hole">
              <span class="donut-total">&yen;{{ formatDonutRevenue(dashboardStats.todayRevenue.value) }}</span>
              <span class="donut-total-label">今日营收</span>
            </div>
          </div>
          <div class="donut-legend">
            <div class="legend-item" v-for="cat in revenueCategories" :key="cat.name">
              <span class="legend-dot" :style="{ background: cat.color }"></span>
              {{ cat.name }}
              <span class="legend-value">{{ cat.percent }}%</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 最新订单 -->
    <el-card class="orders-card" shadow="never">
      <template #header>
        <div class="orders-header">
          <span>最新订单</span>
          <router-link to="/orders" class="orders-link">查看全部 &rarr;</router-link>
        </div>
      </template>
      <el-table :data="latestOrders" stripe style="width: 100%" size="small">
        <el-table-column prop="id" label="订单号" width="180">
          <template #default="{ row }">
            <span class="order-id">{{ row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="user" label="用户" width="100" />
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }">
            <span class="order-amount">&yen;{{ row.amount.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="statusText" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">
              {{ row.statusText }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="time" label="时间" min-width="160" />
      </el-table>
    </el-card>

    <!-- 快捷操作 -->
    <div class="actions-grid">
      <div
        class="action-card"
        v-for="action in visibleActions"
        :key="action.label"
        @click="handleQuickAction(action)"
      >
        <div :class="['action-icon', action.color]">
          <el-icon :size="22"><component :is="actionIcons[action.icon]" /></el-icon>
        </div>
        <div class="action-label">{{ action.label }}</div>
        <div class="action-desc">{{ action.desc }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 仪表盘页面组件
 * 展示统计卡片、订单趋势图表、营收分类占比、最新订单列表、快捷操作
 * 优先从后端 API 获取真实数据，不可用的字段使用 mock 降级
 */
import { ref, computed, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Document, Money, User, Goods,
  Plus, ShoppingCart, Avatar, Bell
} from '@element-plus/icons-vue'
import { getDashboardStats, getOrderTrend, getRevenueCategories } from '../../api/dashboard'
import { listOrder } from '../../api/order'
import { quickActions } from '../../api/mock'
import { useAppStore } from '../../store/app'
import { generateSummary } from '../../api/agent'

const router = useRouter()
const store = useAppStore()

// ===================== Mock 降级数据 =====================
const MOCK_STATS = {
  todayOrders: { value: 128, unit: '单', growth: 12.5 },
  todayRevenue: { value: 45680, unit: '元', prefix: '¥', growth: 8.3 },
  totalUsers: { value: 23568, unit: '', growth: 5.2 },
  totalProducts: { value: 1286, unit: '', growth: 3.1 }
}
const MOCK_ORDER_TREND = {
  days: ['6/30', '7/1', '7/2', '7/3', '7/4', '7/5', '7/6', '7/7', '7/8', '7/9', '7/10', '7/11', '7/12', '7/13'],
  values: [85, 92, 78, 110, 95, 120, 105, 98, 115, 130, 108, 125, 140, 128]
}
const MOCK_REVENUE_CATEGORIES = [
  { name: '服装鞋帽', percent: 42, color: '#3b82f6' },
  { name: '数码电子', percent: 28, color: '#10b981' },
  { name: '家居生活', percent: 18, color: '#f59e0b' },
  { name: '其他品类', percent: 12, color: '#8b5cf6' }
]
const MOCK_LATEST_ORDERS = [
  { id: 'ORD-20260713-001', user: '王芳', amount: 1280.00, status: 'completed', statusText: '已完成', time: '2026-07-13 14:32' },
  { id: 'ORD-20260713-002', user: '李强', amount: 568.50, status: 'processing', statusText: '处理中', time: '2026-07-13 14:15' },
  { id: 'ORD-20260713-003', user: '赵丽', amount: 2450.00, status: 'processing', statusText: '处理中', time: '2026-07-13 13:58' },
  { id: 'ORD-20260713-004', user: '陈伟', amount: 399.00, status: 'completed', statusText: '已完成', time: '2026-07-13 12:40' },
  { id: 'ORD-20260713-005', user: '刘洋', amount: 1899.00, status: 'cancelled', statusText: '已取消', time: '2026-07-13 11:22' }
]

// ===================== 响应式状态 =====================
/** 统计卡片数据（初始为 mock，API 返回后覆盖真实字段） */
const dashboardStats = reactive({ ...MOCK_STATS })
/** 订单趋势数据 */
const orderTrend = reactive({ ...MOCK_ORDER_TREND })
/** 营收分类占比 */
const revenueCategories = reactive([...MOCK_REVENUE_CATEGORIES])
/** 最新订单 */
const latestOrders = reactive([...MOCK_LATEST_ORDERS])

/** AI 运营摘要 */
const aiSummary = ref('')
/** AI 运营摘要 loading */
const summaryLoading = ref(false)

// ===================== 数据获取 =====================
/**
 * 初始化：并行获取仪表盘所有数据，失败时保留 mock 降级
 */
onMounted(async () => {
  // 并行请求仪表盘统计 + 订单趋势 + 营收分类
  await Promise.all([
    fetchStats(),
    fetchTrend(),
    fetchCategories()
  ])

  // 获取最新订单（独立处理数据转换）
  await fetchLatestOrders()
})

/**
 * 获取仪表盘统计数据
 * 真实字段：todayOrders / todayRevenue / totalUsers / totalProducts / 增长率
 */
async function fetchStats() {
  try {
    const res = await getDashboardStats()
    if (!res) return
    dashboardStats.todayOrders.value = res.todayOrders ?? MOCK_STATS.todayOrders.value
    dashboardStats.todayRevenue.value = res.todayRevenue
      ? Number(res.todayRevenue)
      : MOCK_STATS.todayRevenue.value
    dashboardStats.totalUsers.value = res.totalUsers ?? MOCK_STATS.totalUsers.value
    dashboardStats.totalProducts.value = res.totalProducts ?? MOCK_STATS.totalProducts.value
    // 增长率优先使用后端真实值（今日 vs 昨日），未提供时保留 mock
    if (res.ordersGrowth != null) dashboardStats.todayOrders.growth = res.ordersGrowth
    if (res.revenueGrowth != null) dashboardStats.todayRevenue.growth = res.revenueGrowth
  } catch (e) {
    console.warn('获取仪表盘统计数据失败，使用 mock 降级', e)
  }
}

/** 今日营收环形图格式化：过千显示 K（如 45680 → 45.7K） */
function formatDonutRevenue(value: number): string {
  if (value >= 1000) return (value / 1000).toFixed(1) + 'K'
  return String(value)
}

/**
 * 获取近14日订单趋势（真实数据：order service 按日统计订单数）
 */
async function fetchTrend() {
  try {
    const res = await getOrderTrend()
    if (!res?.items?.length) return
    orderTrend.days = res.items.map(i => i.date)
    orderTrend.values = res.items.map(i => i.value)
  } catch (e) {
    console.warn('获取订单趋势失败，使用 mock 降级', e)
  }
}

/**
 * 获取营收分类占比（真实数据：近30天已完成订单按商品分类聚合）
 */
async function fetchCategories() {
  try {
    const res = await getRevenueCategories()
    if (!res?.categories?.length) return
    revenueCategories.length = 0
    res.categories.forEach(c => {
      revenueCategories.push({ name: c.name, percent: c.percentage, color: c.color })
    })
  } catch (e) {
    console.warn('获取营收分类失败，使用 mock 降级', e)
  }
}

/**
 * 获取最新订单（最近5条）
 * 通过已有 listOrder API 获取真实数据
 */
async function fetchLatestOrders() {
  try {
    const res = await listOrder({ page: 1, page_size: 5 })
    if (!res?.items?.length) return
    latestOrders.length = 0
    res.items.forEach(o => {
      latestOrders.push({
        id: o.orderNo,
        user: o.userName,
        amount: o.amount / 100, // 分→元
        status: o.status,
        statusText: o.statusText,
        time: o.createdAt
      })
    })
  } catch (e) {
    console.warn('获取最新订单失败，使用 mock 降级', e)
  }
}

// ===================== 计算属性 =====================
/** 统计卡片映射 */
const stats = computed(() => [
  {
    label: '今日订单',
    displayValue: dashboardStats.todayOrders.value,
    unit: dashboardStats.todayOrders.unit,
    growth: dashboardStats.todayOrders.growth,
    icon: Document,
    iconClass: 'blue',
    prefix: ''
  },
  {
    label: '今日营收',
    displayValue: dashboardStats.todayRevenue.value.toLocaleString(),
    unit: dashboardStats.todayRevenue.unit,
    prefix: dashboardStats.todayRevenue.prefix,
    growth: dashboardStats.todayRevenue.growth,
    icon: Money,
    iconClass: 'green'
  },
  {
    label: '总用户数',
    displayValue: dashboardStats.totalUsers.value.toLocaleString(),
    unit: dashboardStats.totalUsers.unit,
    prefix: '',
    growth: dashboardStats.totalUsers.growth,
    icon: User,
    iconClass: 'orange'
  },
  {
    label: '总商品数',
    displayValue: dashboardStats.totalProducts.value.toLocaleString(),
    unit: dashboardStats.totalProducts.unit,
    prefix: '',
    growth: dashboardStats.totalProducts.growth,
    icon: Goods,
    iconClass: 'purple'
  }
])

/** 订单趋势最大值（用于柱状图高度计算） */
const maxTrendVal = computed(() => Math.max(...orderTrend.values))

/** 环形图渐变色 */
const donutGradient = computed(() => {
  let currentDeg = 0
  const parts = revenueCategories.map(cat => {
    const start = currentDeg
    const end = currentDeg + (cat.percent / 100) * 360
    currentDeg = end
    return `${cat.color} ${start}deg ${end}deg`
  })
  return `conic-gradient(${parts.join(', ')})`
})

/** 快捷操作图标映射 */
const actionIcons: Record<string, any> = {
  add: Plus,
  order: ShoppingCart,
  user: Avatar,
  notice: Bell
}

/**
 * 快捷操作 → 权限映射（与侧边栏 AppSidebar.vue 保持一致）
 * key 为操作名称，value 为 [module, action] 元组，无权限时隐藏
 */
const ACTION_PERMS: Record<string, [string, string]> = {
  '系统公告': ['settings', 'basic']
}

/** 根据权限过滤后的可见快捷操作 */
const visibleActions = computed(() => {
  return quickActions.filter(action => {
    const perm = ACTION_PERMS[action.label]
    if (!perm) return true // 无权限要求，可见
    return store.hasPerm(perm[0], perm[1])
  })
})

/**
 * 获取订单状态对应的 el-tag 类型
 * @param status 订单状态标识
 */
function statusTagType(status: string): string {
  const map: Record<string, string> = {
    completed: 'success',
    processing: 'warning',
    cancelled: 'info'
  }
  return map[status] || 'info'
}

/**
 * 处理快捷操作点击
 */
function handleQuickAction(action: any) {
  if (action.route) {
    router.push(action.route)
  } else {
    ElMessage.info(`点击了 ${action.label}`)
  }
}

/**
 * 生成 AI 运营摘要
 * 从当前统计数据中获取今日订单、营收、用户、商品数，调用 AI 接口生成运营简报
 */
async function handleGenerateSummary() {
  summaryLoading.value = true
  try {
    const data = {
      today_orders: Number(dashboardStats.todayOrders.value) || 0,
      today_revenue: Number(dashboardStats.todayRevenue.value) || 0,
      total_users: Number(dashboardStats.totalUsers.value) || 0,
      total_products: Number(dashboardStats.totalProducts.value) || 0
    }
    const summary = await generateSummary(data)
    if (summary) {
      aiSummary.value = summary
    } else {
      ElMessage.warning('AI 运营摘要生成失败，请重试')
    }
  } catch {
    ElMessage.error('AI 运营摘要生成失败')
  } finally {
    summaryLoading.value = false
  }
}
</script>

<style scoped>
/* ---- 动效 ──── */
@keyframes dashFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes dashSlideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes dashScaleIn {
  from { opacity: 0; transform: scale(0.92); }
  to { opacity: 1; transform: scale(1); }
}
@keyframes barGrow {
  from { height: 0 !important; }
}
@keyframes countUp {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes iconBounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.15); }
}

.dashboard {
  width: 100%;
  animation: dashFadeIn 0.3s ease;
}

/* ---- 统计卡片 ---- */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: #fff;
  border-radius: var(--shop-radius-lg);
  padding: 1.25rem;
  box-shadow: var(--shop-shadow-sm);
  border: 1px solid var(--shop-border);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
  animation: dashSlideUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.stat-card:nth-child(1) { animation-delay: 0.05s; }
.stat-card:nth-child(2) { animation-delay: 0.12s; }
.stat-card:nth-child(3) { animation-delay: 0.19s; }
.stat-card:nth-child(4) { animation-delay: 0.26s; }

.stat-card:hover {
  box-shadow: var(--shop-shadow-md);
  transform: translateY(-1px);
}

.stat-card:hover .stat-icon {
  animation: iconBounce 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
}

.stat-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.stat-icon.blue { background: var(--shop-primary-bg); color: var(--shop-primary); }
.stat-icon.green { background: #ecfdf5; color: var(--shop-success); }
.stat-icon.orange { background: #fff7ed; color: var(--shop-warning); }
.stat-icon.purple { background: #f5f3ff; color: #7c3aed; }

.stat-value {
  font-size: 1.625rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
}

.stat-unit {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
}

.stat-change {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  margin-top: 0.375rem;
}

.stat-change.up { color: var(--shop-success); }
.stat-change.down { color: var(--shop-destructive); }

/* ---- AI 运营摘要 ---- */
.summary-card {
  margin-bottom: 1.5rem;
  animation: dashSlideUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: 0.1s;
}
.summary-card :deep(.el-card__header) {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--shop-border);
}
.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.summary-title {
  font-size: 0.9375rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.summary-title::before {
  content: '';
  width: 3px;
  height: 14px;
  background: linear-gradient(180deg, var(--shop-primary), #60a5fa);
  border-radius: 2px;
}
.summary-body {
  padding: 0.25rem 0;
}
.summary-text {
  font-size: 0.875rem;
  line-height: 1.8;
  color: var(--shop-foreground);
  white-space: pre-wrap;
  padding: 0.5rem 0;
}
.summary-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem 0;
}
.summary-empty-text {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ---- 图表区域 ---- */
.charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.chart-card {
  animation: dashSlideUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.chart-card:nth-child(1) { animation-delay: 0.18s; }
.chart-card:nth-child(2) { animation-delay: 0.25s; }

.chart-card :deep(.el-card__body) {
  padding: 1rem 1.25rem;
}

.chart-title {
  font-size: 0.9375rem;
  font-weight: 600;
}

/* 条形图 */
.line-chart {
  display: flex;
  align-items: flex-end;
  gap: 6px;
  height: 180px;
  padding-top: 0.5rem;
}

.bar-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.375rem;
  height: 100%;
  justify-content: flex-end;
}

.bar {
  width: 100%;
  max-width: 28px;
  border-radius: 4px 4px 0 0;
  background: linear-gradient(180deg, var(--shop-primary), var(--shop-primary-light));
  min-height: 4px;
  transition: height 0.6s ease;
  animation: barGrow 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.bar-wrap:nth-child(1) .bar { animation-delay: 0.30s; }
.bar-wrap:nth-child(2) .bar { animation-delay: 0.33s; }
.bar-wrap:nth-child(3) .bar { animation-delay: 0.36s; }
.bar-wrap:nth-child(4) .bar { animation-delay: 0.39s; }
.bar-wrap:nth-child(5) .bar { animation-delay: 0.42s; }
.bar-wrap:nth-child(6) .bar { animation-delay: 0.45s; }
.bar-wrap:nth-child(7) .bar { animation-delay: 0.48s; }
.bar-wrap:nth-child(8) .bar { animation-delay: 0.51s; }
.bar-wrap:nth-child(9) .bar { animation-delay: 0.54s; }
.bar-wrap:nth-child(10) .bar { animation-delay: 0.57s; }
.bar-wrap:nth-child(11) .bar { animation-delay: 0.60s; }
.bar-wrap:nth-child(12) .bar { animation-delay: 0.63s; }
.bar-wrap:nth-child(13) .bar { animation-delay: 0.66s; }
.bar-wrap:nth-child(14) .bar { animation-delay: 0.69s; }

.bar:hover {
  opacity: 0.8;
}

.bar-label {
  font-size: 0.625rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

.bar-value {
  font-size: 0.5625rem;
  color: var(--shop-muted-foreground);
  font-weight: 600;
}

/* 环形图 */
.donut-wrap {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.donut {
  width: 160px;
  height: 160px;
  border-radius: 999px;
  flex-shrink: 0;
  position: relative;
}

.donut-hole {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  background: #fff;
  border-radius: 999px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.donut-total {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1;
}

.donut-total-label {
  font-size: 0.625rem;
  color: var(--shop-muted-foreground);
  margin-top: 2px;
}

.donut-legend {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.legend-dot {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 999px;
  flex-shrink: 0;
}

.legend-value {
  margin-left: auto;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  font-size: 0.75rem;
}

/* ---- 最新订单 ---- */
.orders-card {
  margin-bottom: 1.5rem;
  animation: dashSlideUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: 0.32s;
}

.orders-card :deep(.el-card__header) {
  padding: 1.125rem 1.25rem;
  border-bottom: 1px solid var(--shop-border);
}

.orders-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.9375rem;
  font-weight: 600;
}

.orders-link {
  font-size: 0.8125rem;
  color: var(--shop-primary);
  font-weight: 500;
  text-decoration: none;
}

.orders-link:hover {
  text-decoration: underline;
}

.order-id {
  font-weight: 600;
  color: var(--shop-primary);
  font-family: var(--shop-font-mono);
  font-size: 0.75rem;
}

.order-amount {
  font-weight: 600;
}

/* ---- 快捷操作 ---- */
.actions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.action-card {
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 1.25rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  animation: dashScaleIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.action-card:nth-child(1) { animation-delay: 0.38s; }
.action-card:nth-child(2) { animation-delay: 0.43s; }
.action-card:nth-child(3) { animation-delay: 0.48s; }
.action-card:nth-child(4) { animation-delay: 0.53s; }

.action-card:hover {
  box-shadow: var(--shop-shadow-md);
  border-color: var(--shop-primary);
  transform: translateY(-2px);
}

.action-icon {
  width: 3rem;
  height: 3rem;
  border-radius: var(--shop-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.375rem;
  margin: 0 auto 0.625rem;
}

.action-icon.blue { background: var(--shop-primary-bg); color: var(--shop-primary); }
.action-icon.green { background: #ecfdf5; color: var(--shop-success); }
.action-icon.orange { background: #fff7ed; color: var(--shop-warning); }
.action-icon.purple { background: #f5f3ff; color: #7c3aed; }

.action-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.action-desc {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 0.25rem;
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-grid { grid-template-columns: 1fr; }
  .actions-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr 1fr; gap: 0.75rem; }
  .actions-grid { grid-template-columns: 1fr 1fr; }
  .donut-wrap { flex-direction: column; align-items: center; }
}

@media (max-width: 480px) {
  .stats-grid { grid-template-columns: 1fr; }
  .actions-grid { grid-template-columns: 1fr; }
}
</style>
