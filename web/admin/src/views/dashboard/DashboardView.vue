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
              <span class="donut-total">&yen;45.6K</span>
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
        v-for="action in quickActions"
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
 */
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Document, Money, User, Goods,
  Plus, ShoppingCart, Avatar, Bell
} from '@element-plus/icons-vue'
import {
  dashboardStats, orderTrend, revenueCategories,
  latestOrders, quickActions
} from '../../api/mock'

const router = useRouter()

/** 统计数据映射 */
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

/** 订单趋势最大值 */
const maxTrendVal = computed(() => Math.max(...orderTrend.values))

/** 环形图渐变色 */
const donutGradient = computed(() => {
  const segments = revenueCategories.map(cat => `${cat.color} ${cat.percent}%`)
  // 改为 conic-gradient 形式
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
const actionIcons = {
  add: Plus,
  order: ShoppingCart,
  user: Avatar,
  notice: Bell
}

/**
 * 获取订单状态对应的 el-tag 类型
 * @param {string} status 订单状态标识
 * @returns {string} tag 类型
 */
function statusTagType(status) {
  const map = {
    completed: 'success',
    processing: 'warning',
    cancelled: 'info'
  }
  return map[status] || 'info'
}

/**
 * 处理快捷操作点击
 * @param {Object} action 快捷操作项
 */
function handleQuickAction(action) {
  if (action.route) {
    router.push(action.route)
  } else {
    ElMessage.info(`点击了 ${action.label}`)
  }
}
</script>

<style scoped>
.dashboard {
  width: 100%;
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
}

.stat-card:hover {
  box-shadow: var(--shop-shadow-md);
  transform: translateY(-1px);
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

/* ---- 图表区域 ---- */
.charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

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
}

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
}

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
