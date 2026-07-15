<template>
  <!-- 用户管理页面 - 统计卡片 + 搜索筛选 + 用户表格 + 分页 -->
  <div class="user-list">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in userStatList" :key="stat.label">
        <div class="stat-header">
          <div>
            <div class="stat-label">{{ stat.label }}</div>
            <div class="stat-value">{{ stat.displayValue }}</div>
            <div v-if="stat.sub" class="stat-sub">{{ stat.sub }}</div>
          </div>
          <div class="stat-icon" :class="stat.color">
            <el-icon :size="22"><component :is="stat.icon" /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 搜索和筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索用户名/手机号..."
          :prefix-icon="Search"
          size="small"
          style="width: 240px"
          clearable
        />
      </div>
      <div class="filter-right">
        <div class="filter-group">
          <span class="filter-label">会员等级:</span>
          <el-radio-group v-model="levelFilter" size="small" @change="handleFilterChange">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="regular">普通会员</el-radio-button>
            <el-radio-button value="gold">黄金会员</el-radio-button>
            <el-radio-button value="diamond">钻石会员</el-radio-button>
          </el-radio-group>
        </div>
        <div class="filter-group">
          <span class="filter-label">状态:</span>
          <el-radio-group v-model="statusFilter" size="small" @change="handleFilterChange">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="active">正常</el-radio-button>
            <el-radio-button value="frozen">冻结</el-radio-button>
          </el-radio-group>
        </div>
        <el-button size="small" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出数据
        </el-button>
      </div>
    </div>

    <!-- 用户表格 -->
    <el-table :data="filteredUsers" stripe style="width: 100%" size="small" v-loading="loading">
      <el-table-column label="用户" width="80">
        <template #default="{ row }">
          <div class="user-avatar" :style="{ background: avatarColor(row.name) }">
            {{ row.name.charAt(0) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="用户名" width="100" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="会员等级" width="120">
        <template #default="{ row }">
          <el-tag
            :type="levelTagType(row.level)"
            size="small"
            effect="plain"
          >
            {{ row.levelText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="regTime" label="注册时间" width="120" />
      <el-table-column label="消费金额" width="120">
        <template #default="{ row }">
          <span class="amount-text">&yen;{{ row.amount.toLocaleString() }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="orders" label="订单数" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <span :class="['status-indicator', row.status]">
            <span class="status-dot"></span>
            {{ row.statusText }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button
            text
            size="small"
            :type="row.status === 'active' ? 'danger' : 'success'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === 'active' ? '冻结' : '解冻' }}
          </el-button>
          <el-button text size="small" @click="handleViewOrders(row)">
            订单
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="filteredUsers.length"
        layout="total, prev, pager, next, jumper"
        background
        small
      />
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 用户管理列表页面
 * 支持搜索、会员等级筛选、状态筛选、冻结/解冻操作、导出数据
 */
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, User, TrendCharts, Avatar, Medal } from '@element-plus/icons-vue'
import { users, userStats } from '../../api/mock'

/** 搜索关键词 */
const searchKeyword = ref('')

/** 会员等级筛选 */
const levelFilter = ref('')

/** 状态筛选 */
const statusFilter = ref('')

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 加载状态 */
const loading = ref(false)

/** 用户统计卡片配置 */
const userStatList = computed(() => [
  {
    label: userStats.totalUsers.label,
    displayValue: userStats.totalUsers.value.toLocaleString(),
    icon: User,
    color: 'blue'
  },
  {
    label: userStats.todayNew.label,
    displayValue: `+${userStats.todayNew.value}`,
    sub: userStats.todayNew.growth,
    icon: TrendCharts,
    color: 'green'
  },
  {
    label: userStats.activeUsers.label,
    displayValue: userStats.activeUsers.value.toLocaleString(),
    sub: userStats.activeUsers.sub,
    icon: Avatar,
    color: 'orange'
  },
  {
    label: userStats.memberUsers.label,
    displayValue: userStats.memberUsers.value.toLocaleString(),
    sub: userStats.memberUsers.sub,
    icon: Medal,
    color: 'purple'
  }
])

/** 过滤后的用户列表 */
const filteredUsers = computed(() => {
  let list = [...users]

  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(u =>
      u.name.toLowerCase().includes(kw) || u.phone.includes(kw)
    )
  }

  if (levelFilter.value) {
    list = list.filter(u => u.level === levelFilter.value)
  }

  if (statusFilter.value) {
    list = list.filter(u => u.status === statusFilter.value)
  }

  return list
})

/**
 * 获取会员等级对应的 el-tag 类型
 * @param {string} level 会员等级
 * @returns {string} tag 类型
 */
function levelTagType(level) {
  const map = {
    regular: 'info',
    gold: 'warning',
    diamond: 'primary'
  }
  return map[level] || 'info'
}

/**
 * 根据用户名生成头像背景色
 * @param {string} name 用户名
 * @returns {string} 颜色值
 */
function avatarColor(name) {
  const colors = ['#4338ca', '#d97706', '#64748b', '#059669', '#dc2626', '#7c3aed', '#0891b2']
  const idx = name.charCodeAt(0) % colors.length
  return colors[idx]
}

/**
 * 筛选条件变更时重置页码
 */
function handleFilterChange() {
  currentPage.value = 1
}

/**
 * 编辑用户（模拟弹窗）
 * @param {Object} row 用户行数据
 */
function handleEdit(row) {
  ElMessage.info(`编辑用户：${row.name}`)
}

/**
 * 切换用户冻结/解冻状态
 * @param {Object} row 用户行数据
 */
function handleToggleStatus(row) {
  const actionText = row.status === 'active' ? '冻结' : '解冻'
  ElMessageBox.confirm(
    `确定${actionText}用户「${row.name}」吗？`,
    '操作确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    if (row.status === 'active') {
      row.status = 'frozen'
      row.statusText = '冻结'
    } else {
      row.status = 'active'
      row.statusText = '正常'
    }
    ElMessage.success(`用户已${actionText}`)
  }).catch(() => {})
}

/**
 * 查看用户订单
 * @param {Object} row 用户行数据
 */
function handleViewOrders(row) {
  ElMessage.info(`查看用户 ${row.name} 的订单`)
}

/**
 * 导出用户数据
 */
function handleExport() {
  const headers = ['用户名', '手机号', '会员等级', '注册时间', '消费金额', '订单数', '状态']
  const rows = filteredUsers.value.map(u => [
    u.name, u.phone, u.levelText, u.regTime,
    `¥${u.amount.toFixed(2)}`, u.orders, u.statusText
  ])

  const csv = [headers, ...rows]
    .map(r => r.map(c => `"${c}"`).join(','))
    .join('\n')

  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `用户数据_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)

  ElMessage.success('用户数据导出成功')
}
</script>

<style scoped>
.user-list {
  width: 100%;
}

/* 统计卡片 */
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
  transform: translateY(-2px);
}

.stat-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.stat-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
}

.stat-sub {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 2px;
}

.stat-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.blue { background: var(--shop-primary-bg); color: var(--shop-primary); }
.stat-icon.green { background: #f0fdf4; color: var(--shop-success); }
.stat-icon.orange { background: #fef3c7; color: #d97706; }
.stat-icon.purple { background: #e0e7ff; color: #4338ca; }

/* 筛选栏 */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  padding: 16px 20px;
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

/* 用户头像 */
.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 0.875rem;
  font-weight: 600;
}

/* 状态指示器 */
.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-indicator.active .status-dot {
  background: var(--shop-success);
}

.status-indicator.frozen .status-dot {
  background: var(--shop-destructive);
}

.status-indicator.active {
  color: var(--shop-success);
}

.status-indicator.frozen {
  color: var(--shop-destructive);
}

.amount-text {
  font-weight: 600;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

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
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
