<template>
  <!-- 操作日志页面 - 统计卡片 + 筛选栏 + 日志表格 + 详情弹窗 + 分页 -->
  <div class="logs-view">
    <!-- 页面标题与描述 -->
    <div class="page-header">
      <h2 class="page-title">操作日志</h2>
      <p class="page-desc">记录管理员的关键操作，追溯所有后台行为</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in logStats" :key="stat.label">
        <div class="stat-header">
          <div>
            <div class="stat-label">{{ stat.label }}</div>
            <div class="stat-value">{{ stat.value }}</div>
          </div>
          <div :class="['stat-icon', stat.color]">
            <el-icon :size="20"><component :is="stat.icon" /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
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
        />
      </div>
      <div class="filter-item">
        <label class="filter-label">操作人</label>
        <el-select v-model="filterOperator" placeholder="全部操作人" size="small" clearable style="width: 130px">
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
        <el-select v-model="filterType" placeholder="全部类型" size="small" clearable style="width: 120px">
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
        />
      </div>
      <div class="filter-actions">
        <el-button size="small" @click="handleReset">
          <el-icon><Refresh /></el-icon>
          重置
        </el-button>
        <el-button size="small" type="primary" @click="handleFilter">
          <el-icon><Search /></el-icon>
          筛选
        </el-button>
        <el-button size="small" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </div>

    <!-- 日志表格 -->
    <el-card class="table-card" shadow="never">
      <el-table
        :data="filteredLogs"
        stripe
        style="width: 100%"
        size="small"
        v-loading="loading"
      >
        <el-table-column prop="time" label="操作时间" width="160" />
        <el-table-column label="操作人" width="140">
          <template #default="{ row }">
            <div class="operator-cell">
              <span class="operator-name">{{ row.operator }}</span>
              <el-tag :type="roleTagType(row.role)" size="small" effect="plain">
                {{ row.role }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag
              :type="typeTagType(row.type)"
              size="small"
              effect="dark"
              class="type-tag"
            >
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP 地址" width="140">
          <template #default="{ row }">
            <span class="mono-text">{{ row.ip }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="操作内容" min-width="260" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
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
        :total="filteredLogs.length"
        layout="total, prev, pager, next, jumper"
        background
        small
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="日志详情"
      width="640px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="detail-grid" v-if="currentLog">
        <div class="detail-row">
          <span class="detail-label">操作时间</span>
          <span class="detail-value">{{ currentLog.time }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">操作人</span>
          <span class="detail-value">{{ currentLog.operator }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">角色</span>
          <span class="detail-value">
            <el-tag :type="roleTagType(currentLog.role)" size="small" effect="plain">
              {{ currentLog.role }}
            </el-tag>
          </span>
        </div>
        <div class="detail-row">
          <span class="detail-label">操作类型</span>
          <span class="detail-value">
            <el-tag
              :type="typeTagType(currentLog.type)"
              size="small"
              effect="dark"
            >
              {{ typeLabel(currentLog.type) }}
            </el-tag>
          </span>
        </div>
        <div class="detail-row">
          <span class="detail-label">IP 地址</span>
          <span class="detail-value mono-text">{{ currentLog.ip }}</span>
        </div>
        <div class="detail-row detail-row-full">
          <span class="detail-label">操作内容</span>
          <span class="detail-value">{{ currentLog.content }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">请求方法</span>
          <span class="detail-value">
            <el-tag :type="methodTagType(currentLog.method)" size="small" effect="plain">
              {{ currentLog.method }}
            </el-tag>
          </span>
        </div>
        <div class="detail-row">
          <span class="detail-label">请求路径</span>
          <span class="detail-value mono-text">{{ currentLog.path }}</span>
        </div>
        <div class="detail-row detail-row-full">
          <span class="detail-label">User Agent</span>
          <span class="detail-value mono-text ua-text">{{ currentLog.userAgent }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">记录 ID</span>
          <span class="detail-value mono-text">{{ currentLog.recordId }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">执行结果</span>
          <span class="detail-value">
            <el-tag
              :type="currentLog.result === '成功' ? 'success' : 'danger'"
              size="small"
              effect="plain"
            >
              {{ currentLog.result }}
            </el-tag>
          </span>
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
 * 支持按日期范围、操作人、操作类型、关键词筛选日志，查看完整日志详情弹窗，导出 CSV
 */
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download, Refresh, View, Clock, Calendar, List } from '@element-plus/icons-vue'

// ==================== Mock 数据 ====================

/** 12 条模拟日志数据 */
const mockLogs = [
  {
    id: 1,
    time: '2026-07-14 10:23:15',
    operator: '张明',
    role: '超级管理员',
    type: 'product',
    ip: '192.168.1.100',
    content: '新增商品「Apple iPhone 15 Pro Max 256GB 原色钛金属」',
    method: 'POST',
    path: '/api/v1/products',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'PROD-20260714-0001',
    result: '成功'
  },
  {
    id: 2,
    time: '2026-07-14 09:45:32',
    operator: '李芳',
    role: '运营',
    type: 'order',
    ip: '192.168.1.102',
    content: '取消订单 #ORD-20260713-0089（买家申请退款）',
    method: 'PUT',
    path: '/api/v1/orders/ORD-20260713-0089/cancel',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'ORD-20260713-0089',
    result: '成功'
  },
  {
    id: 3,
    time: '2026-07-14 08:30:00',
    operator: '王磊',
    role: '管理员',
    type: 'user',
    ip: '10.0.0.45',
    content: '冻结用户「test_user_001」（多次违规操作）',
    method: 'PUT',
    path: '/api/v1/users/test_user_001/status',
    userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15',
    recordId: 'USR-20260714-0003',
    result: '成功'
  },
  {
    id: 4,
    time: '2026-07-13 18:12:08',
    operator: '张明',
    role: '超级管理员',
    type: 'setting',
    ip: '192.168.1.100',
    content: '修改运费模板「江浙沪包邮」配送范围及费用',
    method: 'PUT',
    path: '/api/v1/settings/shipping/TPL-001',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'STG-20260713-0041',
    result: '成功'
  },
  {
    id: 5,
    time: '2026-07-13 16:55:21',
    operator: '陈雪',
    role: '运营',
    type: 'coupon',
    ip: '192.168.1.105',
    content: '创建满减优惠券「满 300 减 50」限新人专享',
    method: 'POST',
    path: '/api/v1/coupons',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    recordId: 'CPN-20260713-0012',
    result: '成功'
  },
  {
    id: 6,
    time: '2026-07-13 14:20:44',
    operator: '李芳',
    role: '运营',
    type: 'product',
    ip: '192.168.1.102',
    content: '批量上架 15 件商品（分类：数码配件）',
    method: 'POST',
    path: '/api/v1/products/batch/online',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'BATCH-20260713-0003',
    result: '成功'
  },
  {
    id: 7,
    time: '2026-07-12 11:05:37',
    operator: '王磊',
    role: '管理员',
    type: 'user',
    ip: '10.0.0.45',
    content: '重置用户「mobile_138****5678」登录密码',
    method: 'PUT',
    path: '/api/v1/users/mobile_138****5678/password',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'USR-20260712-0008',
    result: '成功'
  },
  {
    id: 8,
    time: '2026-07-12 09:30:12',
    operator: '张明',
    role: '超级管理员',
    type: 'setting',
    ip: '192.168.1.100',
    content: '开启系统维护模式，设置维护提示文案',
    method: 'PUT',
    path: '/api/v1/settings/maintenance',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'STG-20260712-0035',
    result: '成功'
  },
  {
    id: 9,
    time: '2026-07-11 20:15:03',
    operator: '陈雪',
    role: '运营',
    type: 'order',
    ip: '192.168.1.105',
    content: '手动退款订单 #ORD-20260710-0056（金额 ¥899.00）',
    method: 'POST',
    path: '/api/v1/orders/ORD-20260710-0056/refund',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    recordId: 'RFN-20260711-0002',
    result: '成功'
  },
  {
    id: 10,
    time: '2026-07-11 15:40:55',
    operator: '李芳',
    role: '运营',
    type: 'coupon',
    ip: '192.168.1.102',
    content: '手动发放优惠券「全场 8 折」给 50 位活跃用户',
    method: 'POST',
    path: '/api/v1/coupons/CPN-20260710-0009/assign',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'BATCH-20260711-0005',
    result: '成功'
  },
  {
    id: 11,
    time: '2026-07-10 14:22:38',
    operator: '王磊',
    role: '管理员',
    type: 'product',
    ip: '10.0.0.45',
    content: '删除商品「华为 MatePad SE 11 英寸 2026 款」（违规下架）',
    method: 'DELETE',
    path: '/api/v1/products/PROD-20260709-0088',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'PROD-20260709-0088',
    result: '成功'
  },
  {
    id: 12,
    time: '2026-07-10 10:08:19',
    operator: '张明',
    role: '超级管理员',
    type: 'setting',
    ip: '192.168.1.100',
    content: '配置短信网关参数（切换到阿里云短信服务）',
    method: 'PUT',
    path: '/api/v1/settings/sms/gateway',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/128.0.0.0',
    recordId: 'STG-20260710-0029',
    result: '失败'
  }
]

// ==================== 统计卡片配置 ====================

/** 统计卡片数据 */
const logStats = [
  { label: '今日操作', value: '268', icon: Clock, color: 'blue' },
  { label: '本周操作', value: '1,842', icon: Calendar, color: 'green' },
  { label: '本月操作', value: '6,351', icon: List, color: 'orange' },
  { label: '累计操作', value: '23,480', icon: View, color: 'purple' }
]

// ==================== 筛选选项 ====================

/** 操作人下拉选项 */
const operatorOptions = [
  { label: '张明', value: '张明' },
  { label: '李芳', value: '李芳' },
  { label: '王磊', value: '王磊' },
  { label: '陈雪', value: '陈雪' }
]

/** 操作类型下拉选项 */
const typeOptions = [
  { label: '商品', value: 'product' },
  { label: '订单', value: 'order' },
  { label: '用户', value: 'user' },
  { label: '设置', value: 'setting' },
  { label: '优惠券', value: 'coupon' }
]

/** 操作类型中文标签映射 */
const typeLabelMap = {
  product: '商品',
  order: '订单',
  user: '用户',
  setting: '设置',
  coupon: '优惠券'
}

// ==================== 响应式状态 ====================

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

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 加载状态 */
const loading = ref(false)

/** 详情弹窗可见性 */
const detailVisible = ref(false)

/** 当前查看的日志 */
const currentLog = ref(null)

// ==================== 计算属性 ====================

/** 过滤后的日志列表 */
const filteredLogs = computed(() => {
  let list = [...mockLogs]

  // 开始日期筛选
  if (filterStartDate.value) {
    list = list.filter(log => log.time >= filterStartDate.value)
  }

  // 结束日期筛选（包含当天）
  if (filterEndDate.value) {
    list = list.filter(log => log.time <= filterEndDate.value + ' 23:59:59')
  }

  // 操作人筛选
  if (filterOperator.value) {
    list = list.filter(log => log.operator === filterOperator.value)
  }

  // 操作类型筛选
  if (filterType.value) {
    list = list.filter(log => log.type === filterType.value)
  }

  // 关键词搜索（匹配操作内容）
  if (filterKeyword.value) {
    const kw = filterKeyword.value.toLowerCase()
    list = list.filter(log => log.content.toLowerCase().includes(kw))
  }

  return list
})

// ==================== 辅助函数 ====================

/**
 * 获取操作类型对应的 el-tag 颜色类型
 * @param {string} type 操作类型标识
 * @returns {string} tag 类型（success/warning/danger/info/primary）
 */
function typeTagType(type) {
  const map = {
    product: 'primary',
    order: 'success',
    user: 'warning',
    setting: 'info',
    coupon: 'danger'
  }
  return map[type] || 'info'
}

/**
 * 获取角色对应的 el-tag 颜色类型
 * @param {string} role 角色名称
 * @returns {string} tag 类型
 */
function roleTagType(role) {
  const map = {
    '超级管理员': 'danger',
    '管理员': 'primary',
    '运营': 'success'
  }
  return map[role] || 'info'
}

/**
 * 获取操作类型的中文标签
 * @param {string} type 操作类型标识
 * @returns {string} 中文标签
 */
function typeLabel(type) {
  return typeLabelMap[type] || type
}

/**
 * 获取 HTTP 方法对应的 el-tag 类型
 * @param {string} method HTTP 方法
 * @returns {string} tag 类型
 */
function methodTagType(method) {
  const map = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'info'
  }
  return map[method] || 'info'
}

// ==================== 事件处理 ====================

/**
 * 执行筛选
 */
function handleFilter() {
  currentPage.value = 1
  ElMessage.success('筛选完成')
}

/**
 * 重置所有筛选条件
 */
function handleReset() {
  filterStartDate.value = ''
  filterEndDate.value = ''
  filterOperator.value = ''
  filterType.value = ''
  filterKeyword.value = ''
  currentPage.value = 1
  ElMessage.success('已重置筛选条件')
}

/**
 * 查看日志详情
 * @param {Object} row 日志行数据
 */
function handleViewDetail(row) {
  currentLog.value = row
  detailVisible.value = true
}

/**
 * 导出日志 CSV
 */
function handleExport() {
  const headers = ['操作时间', '操作人', '角色', '操作类型', 'IP地址', '操作内容', '请求方法', '请求路径', '结果']
  const rows = filteredLogs.value.map(log => [
    log.time,
    log.operator,
    log.role,
    typeLabel(log.type),
    log.ip,
    log.content,
    log.method,
    log.path,
    log.result
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

.logs-view {
  width: 100%;
  animation: logsFadeUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

/* ---- 页面标题 ---- */
.page-header {
  margin-bottom: 1.25rem;
}

.page-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0 0 0.25rem;
}

.page-desc {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin: 0;
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
  animation: logsCardIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.stat-card:nth-child(1) { animation-delay: 0.05s; }
.stat-card:nth-child(2) { animation-delay: 0.1s; }
.stat-card:nth-child(3) { animation-delay: 0.15s; }
.stat-card:nth-child(4) { animation-delay: 0.2s; }

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

/* ---- 筛选栏 ---- */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 16px;
  animation: logsCardIn 0.4s ease both;
  animation-delay: 0.25s;
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
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
  animation: logsCardIn 0.4s ease both;
  animation-delay: 0.35s;
}

.table-card :deep(.el-card__body) {
  padding: 0;
}

.operator-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.operator-name {
  font-size: 0.8125rem;
  font-weight: 500;
}

/* ---- 标签样式 ---- */
.type-tag {
  font-weight: 600;
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

/* ---- 分页 ---- */
.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

/* ---- 详情弹窗 ---- */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
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
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.detail-value {
  font-size: 0.875rem;
  color: var(--shop-foreground);
  line-height: 1.5;
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
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
