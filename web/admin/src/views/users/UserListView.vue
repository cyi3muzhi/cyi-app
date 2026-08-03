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
          placeholder="搜索用户名/手机号/昵称..."
          :prefix-icon="Search"
          size="small"
          style="width: 260px"
          clearable
          @keyup.enter="handleSearch"
        />
      </div>
      <div class="filter-right">
        <div class="filter-group">
          <span class="filter-label">会员等级:</span>
          <el-radio-group v-model="levelFilter" size="small" @change="handleSearch">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="regular">普通会员</el-radio-button>
            <el-radio-button value="gold">黄金会员</el-radio-button>
            <el-radio-button value="diamond">钻石会员</el-radio-button>
          </el-radio-group>
        </div>
        <div class="filter-group">
          <span class="filter-label">状态:</span>
          <el-radio-group v-model="statusFilter" size="small" @change="handleSearch">
            <el-radio-button :value="0">全部</el-radio-button>
            <el-radio-button :value="1">正常</el-radio-button>
            <el-radio-button :value="2">冻结</el-radio-button>
          </el-radio-group>
        </div>
        <el-button size="small" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出数据
        </el-button>
      </div>
    </div>

    <!-- 用户表格 -->
    <el-table :data="userList" stripe style="width: 100%" size="small" v-loading="loading">
      <el-table-column label="用户" width="64">
        <template #default="{ row }">
          <div class="user-avatar" :style="{ background: avatarColor(row.nickname || row.username) }">
            {{ avatarLetter(row) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="昵称" min-width="100">
        <template #default="{ row }">
          <span>{{ row.nickname || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" min-width="110" />
      <el-table-column prop="phone" label="手机号" min-width="130" />
      <el-table-column label="会员等级" min-width="100">
        <template #default="{ row }">
          <el-tag
            :type="levelTagType(row.level)"
            size="small"
            effect="plain"
          >
            {{ getLevelText(row.level) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册时间" min-width="160">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="最近登录" min-width="160">
        <template #default="{ row }">
          {{ row.lastLoginAt ? formatTime(row.lastLoginAt) : '—' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" min-width="80">
        <template #default="{ row }">
          <span :class="['status-indicator', statusClass(row.status)]">
            <span class="status-dot"></span>
            {{ getStatusText(row.status) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button
            text
            size="small"
            :type="row.status === 1 ? 'danger' : 'success'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === 1 ? '冻结' : '解冻' }}
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
        :total="total"
        layout="total, prev, pager, next, jumper"
        background
        small
        @current-change="fetchUserList"
        @size-change="handleSizeChange"
      />
    </div>

    <!-- 编辑用户弹窗 -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="440px" align-center append-to-body>
      <el-form :model="editForm" label-position="top" class="edit-form">
        <el-form-item label="昵称">
          <el-input v-model="editForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="editForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="性别">
          <el-radio-group v-model="editForm.gender">
            <el-radio :value="0">未知</el-radio>
            <el-radio :value="1">男</el-radio>
            <el-radio :value="2">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="会员等级">
          <el-select v-model="editForm.level" style="width: 100%">
            <el-option label="普通会员" value="regular" />
            <el-option label="黄金会员" value="gold" />
            <el-option label="钻石会员" value="diamond" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitEdit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 用户管理列表页面
 * 查询用户端的用户（user service），支持搜索、会员等级筛选、状态筛选、冻结/解冻操作、导出数据
 */
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, User, TrendCharts, Avatar, Medal } from '@element-plus/icons-vue'
import { listUser, getUserStats, toggleUserStatus, updateUser, type UserItem, type UserStatsResponse } from '../../api/user'

/** 搜索关键词 */
const searchKeyword = ref('')

/** 会员等级筛选（空字符串=全部） */
const levelFilter = ref('')

/** 状态筛选（0=全部 1=正常 2=冻结） */
const statusFilter = ref(0)

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 总数 */
const total = ref(0)

/** 加载状态 */
const loading = ref(false)

/** 提交状态 */
const submitting = ref(false)

/** 用户列表 */
const userList = ref<UserItem[]>([])

/** 用户统计数据 */
const userStats = ref<UserStatsResponse>({
  totalUsers: 0,
  todayNew: 0,
  todayNewGrowth: 0,
  activeUsers: 0,
  memberUsers: 0,
  memberRatio: 0
})

/** 编辑弹窗表单 */
const showEditDialog = ref(false)
const editForm = ref({
  id: 0,
  nickname: '',
  phone: '',
  email: '',
  gender: 0,
  level: 'regular'
})

/** 用户统计卡片配置 */
const userStatList = computed(() => [
  {
    label: '用户总数',
    displayValue: userStats.value.totalUsers.toLocaleString(),
    icon: User,
    color: 'blue'
  },
  {
    label: '今日新增',
    displayValue: `+${userStats.value.todayNew}`,
    sub: `${(userStats.value.todayNewGrowth * 100).toFixed(1)}% 增长`,
    icon: TrendCharts,
    color: 'green'
  },
  {
    label: '活跃用户',
    displayValue: userStats.value.activeUsers.toLocaleString(),
    sub: '最近30天登录',
    icon: Avatar,
    color: 'orange'
  },
  {
    label: '会员用户',
    displayValue: userStats.value.memberUsers.toLocaleString(),
    sub: `占比 ${(userStats.value.memberRatio * 100).toFixed(1)}%`,
    icon: Medal,
    color: 'purple'
  }
])

/**
 * 获取用户列表
 */
async function fetchUserList() {
  loading.value = true
  try {
    const res = await listUser({
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value || undefined,
      level: levelFilter.value || undefined,
      status: statusFilter.value || undefined
    })
    userList.value = res.items
    total.value = res.total
  } catch (e: any) {
    ElMessage.error('获取用户列表失败: ' + (e?.message || ''))
  } finally {
    loading.value = false
  }
}

/**
 * 获取用户统计数据
 */
async function fetchUserStats() {
  try {
    const res = await getUserStats()
    userStats.value = res
  } catch (e: any) {
    console.error('获取用户统计失败:', e)
  }
}

/**
 * 搜索处理（重置到第一页）
 */
function handleSearch() {
  currentPage.value = 1
  fetchUserList()
}

/**
 * 每页条数变更
 */
function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  fetchUserList()
}

/**
 * 获取会员等级对应的 el-tag 类型
 */
function levelTagType(level: string) {
  const map: Record<string, string> = {
    regular: 'info',
    gold: 'warning',
    diamond: 'primary'
  }
  return map[level] || 'info'
}

/**
 * 获取会员等级文本
 */
function getLevelText(level: string) {
  const map: Record<string, string> = {
    regular: '普通会员',
    gold: '黄金会员',
    diamond: '钻石会员'
  }
  return map[level] || '未知'
}

/**
 * 获取状态文本（1=正常 2=冻结）
 */
function getStatusText(status: number) {
  if (status === 1) return '正常'
  if (status === 2) return '冻结'
  return '未知'
}

/**
 * 获取状态样式类名
 */
function statusClass(status: number) {
  return status === 1 ? 'active' : 'frozen'
}

/**
 * 将 Unix 秒时间戳格式化为 YYYY-MM-DD HH:mm:ss
 */
function formatTime(ts: number) {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/**
 * 根据名称生成头像背景色
 */
function avatarColor(name: string) {
  const colors = ['#4338ca', '#d97706', '#64748b', '#059669', '#dc2626', '#7c3aed', '#0891b2']
  const idx = (name?.charCodeAt(0) || 0) % colors.length
  return colors[idx]
}

/**
 * 获取头像首字母（优先昵称，其次用户名）
 */
function avatarLetter(row: UserItem) {
  const name = row.nickname || row.username || ''
  return name ? name.charAt(0).toUpperCase() : 'U'
}

/**
 * 编辑用户（打开弹窗并填充表单）
 */
function handleEdit(row: UserItem) {
  editForm.value = {
    id: row.id,
    nickname: row.nickname,
    phone: row.phone,
    email: row.email,
    gender: row.gender,
    level: row.level
  }
  showEditDialog.value = true
}

/**
 * 提交编辑
 */
async function submitEdit() {
  submitting.value = true
  try {
    await updateUser(editForm.value.id, {
      nickname: editForm.value.nickname,
      phone: editForm.value.phone,
      email: editForm.value.email,
      gender: editForm.value.gender,
      level: editForm.value.level
    })
    ElMessage.success('更新成功')
    showEditDialog.value = false
    fetchUserList()
  } catch (e: any) {
    ElMessage.error('更新失败: ' + (e?.message || ''))
  } finally {
    submitting.value = false
  }
}

/**
 * 切换用户冻结/解冻状态
 */
function handleToggleStatus(row: UserItem) {
  const actionText = row.status === 1 ? '冻结' : '解冻'
  const displayName = row.nickname || row.username
  ElMessageBox.confirm(
    `确定${actionText}用户「${displayName}」吗？`,
    '操作确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(async () => {
    try {
      await toggleUserStatus(row.id)
      ElMessage.success(`用户已${actionText}`)
      fetchUserList()
    } catch (e: any) {
      ElMessage.error('操作失败: ' + (e?.message || ''))
    }
  }).catch(() => {})
}

/**
 * 查看用户订单（占位，后续对接订单详情页）
 */
function handleViewOrders(row: UserItem) {
  const displayName = row.nickname || row.username
  ElMessage.info(`查看用户 ${displayName} 的订单`)
}

/**
 * 导出用户数据为 CSV
 */
function handleExport() {
  const headers = ['昵称', '用户名', '手机号', '会员等级', '注册时间', '最近登录', '状态']
  const rows = userList.value.map(u => [
    u.nickname, u.username, u.phone, getLevelText(u.level),
    formatTime(u.createdAt), u.lastLoginAt ? formatTime(u.lastLoginAt) : '—',
    getStatusText(u.status)
  ])

  const csv = [headers, ...rows]
    .map(r => r.map(c => `"${c ?? ''}"`).join(','))
    .join('\n')

  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `用户数据_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)

  ElMessage.success('用户数据导出成功')
}

/** 页面加载时获取数据 */
onMounted(() => {
  fetchUserList()
  fetchUserStats()
})
</script>

<style scoped>
/* ---- 动效 ---- */
@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.user-list {
  animation: fadeSlideUp 0.35s ease both;
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
  animation: fadeSlideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.stat-card:nth-child(1) { animation-delay: 0s; }
.stat-card:nth-child(2) { animation-delay: 0.08s; }
.stat-card:nth-child(3) { animation-delay: 0.16s; }
.stat-card:nth-child(4) { animation-delay: 0.24s; }

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

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

/* 编辑表单 */
.edit-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

/* 表格行入场动画 */
.el-table__body-wrapper tbody tr {
  animation: fadeSlideUp 0.3s ease both;
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
