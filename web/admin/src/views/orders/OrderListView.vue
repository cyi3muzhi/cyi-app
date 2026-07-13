<template>
  <!-- 订单管理页面 - 状态标签页 + 筛选栏 + 订单表格（可展开）+ 分页 -->
  <div class="order-list">
    <!-- 状态标签页 -->
    <div class="status-tabs">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane
          v-for="tab in tabs"
          :key="tab.key"
          :label="`${tab.label} (${tab.count})`"
          :name="tab.key"
        />
      </el-tabs>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-item">
        <label class="filter-label">下单时间</label>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="small"
          style="width: 240px"
        />
      </div>
      <div class="filter-item">
        <label class="filter-label">状态</label>
        <el-select v-model="statusFilter" placeholder="全部状态" size="small" clearable style="width: 120px">
          <el-option v-for="tab in tabs" :key="tab.key" :label="tab.label" :value="tab.key" />
        </el-select>
      </div>
      <div class="filter-item">
        <el-input
          v-model="orderSearch"
          placeholder="搜索订单号..."
          :prefix-icon="Search"
          size="small"
          style="width: 180px"
          clearable
        />
      </div>
      <div class="filter-actions">
        <el-button size="small" type="primary" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出订单
        </el-button>
      </div>
    </div>

    <!-- 订单表格 -->
    <el-table
      :data="filteredOrders"
      stripe
      style="width: 100%"
      size="small"
      v-loading="loading"
      :row-class-name="tableRowClassName"
    >
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="order-detail">
            <div class="detail-header">订单明细（共 {{ row.items.length }} 件商品）</div>
            <div class="detail-items">
              <div class="detail-item" v-for="(item, idx) in row.items" :key="idx">
                <div class="item-info">
                  <span class="item-name">{{ item.name }}</span>
                  <span class="item-spec">{{ item.spec }}</span>
                  <span class="item-qty">x{{ item.qty }}</span>
                </div>
                <span class="item-price">&yen;{{ item.price.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="id" label="订单号" width="170">
        <template #default="{ row }">
          <span class="order-id-text">{{ row.id }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="time" label="下单时间" width="150" />
      <el-table-column prop="user" label="用户" width="90" />
      <el-table-column label="商品" min-width="200">
        <template #default="{ row }">
          <span class="product-summary">{{ getProductSummary(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="100">
        <template #default="{ row }">
          <span class="amount-text">&yen;{{ row.amount.toFixed(2) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="payment" label="支付方式" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small" effect="plain">
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="handleViewDetail(row)">
            查看详情
          </el-button>
          <el-button
            v-if="row.status === 'pending_ship'"
            text size="small" type="success"
            @click="handleShip(row)"
          >
            立即发货
          </el-button>
          <el-button
            v-if="row.status === 'pending_payment'"
            text size="small" type="danger"
            @click="handleCancel(row)"
          >
            取消订单
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="filteredOrders.length"
        layout="total, prev, pager, next, jumper"
        background
        small
      />
    </div>
  </div>
</template>

<script setup>
/**
 * 订单管理列表页面
 * 支持状态筛选、日期范围搜索、订单号搜索、展开明细、导出 CSV、发货/取消操作
 */
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download } from '@element-plus/icons-vue'
import { orders } from '../../api/mock'

/** 当前激活的标签页 */
const activeTab = ref('all')

/** 日期范围 */
const dateRange = ref(null)

/** 状态下拉筛选 */
const statusFilter = ref('')

/** 订单号搜索 */
const orderSearch = ref('')

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 加载状态 */
const loading = ref(false)

/** 订单标签页配置 */
const tabs = [
  { key: 'all', label: '全部订单', count: 128 },
  { key: 'pending_payment', label: '待付款', count: 18 },
  { key: 'pending_ship', label: '待发货', count: 25 },
  { key: 'pending_receive', label: '待收货', count: 32 },
  { key: 'completed', label: '已完成', count: 41 },
  { key: 'cancelled', label: '已取消', count: 12 }
]

/** 过滤后的订单列表 */
const filteredOrders = computed(() => {
  let list = [...orders]

  // 标签页筛选
  if (activeTab.value !== 'all') {
    list = list.filter(o => o.status === activeTab.value)
  }

  // 状态下拉筛选
  if (statusFilter.value) {
    list = list.filter(o => o.status === statusFilter.value)
  }

  // 订单号搜索
  if (orderSearch.value) {
    const kw = orderSearch.value.toLowerCase()
    list = list.filter(o => o.id.toLowerCase().includes(kw))
  }

  return list
})

/**
 * 获取商品摘要文本
 * @param {Object} order 订单数据
 * @returns {string} 商品摘要
 */
function getProductSummary(order) {
  const names = order.items.slice(0, 2).map(item => item.name)
  let summary = names.join('、')
  if (order.items.length > 2) {
    summary += ` 等${order.items.length}件`
  }
  return summary
}

/**
 * 获取订单状态对应的 el-tag 类型
 * @param {string} status 订单状态
 * @returns {string} tag 类型
 */
function statusTagType(status) {
  const map = {
    pending_payment: 'danger',
    pending_ship: 'warning',
    pending_receive: 'primary',
    completed: 'success',
    cancelled: 'info'
  }
  return map[status] || 'info'
}

/**
 * 表格行样式类名
 */
function tableRowClassName() {
  return 'order-row'
}

/**
 * 处理标签页切换
 */
function handleTabChange() {
  currentPage.value = 1
}

/**
 * 查看订单详情
 * @param {Object} row 订单行数据
 */
function handleViewDetail(row) {
  ElMessage.info(`查看订单详情：${row.id}`)
}

/**
 * 立即发货
 * @param {Object} row 订单行数据
 */
function handleShip(row) {
  ElMessageBox.confirm(
    `确定要将订单 ${row.id} 标记为已发货吗？`,
    '发货确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'success' }
  ).then(() => {
    row.status = 'pending_receive'
    row.statusText = '待收货'
    ElMessage.success('已标记为发货状态')
  }).catch(() => {})
}

/**
 * 取消订单
 * @param {Object} row 订单行数据
 */
function handleCancel(row) {
  ElMessageBox.confirm(
    `确定要取消订单 ${row.id} 吗？`,
    '取消确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
  ).then(() => {
    row.status = 'cancelled'
    row.statusText = '已取消'
    ElMessage.success('订单已取消')
  }).catch(() => {})
}

/**
 * 导出订单 CSV
 */
function handleExport() {
  const headers = ['订单号', '下单时间', '用户', '商品', '金额', '支付方式', '状态']
  const rows = filteredOrders.value.map(o => {
    const products = o.items.map(item =>
      `${item.name}(${item.spec})x${item.qty}`
    ).join(';')
    return [o.id, o.time, o.user, products, o.amount.toFixed(2), o.payment, o.statusText]
  })

  const csv = [headers, ...rows]
    .map(r => r.map(c => `"${c}"`).join(','))
    .join('\n')

  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `订单导出_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)

  ElMessage.success('订单导出成功')
}
</script>

<style scoped>
.order-list {
  max-width: 1200px;
}

.status-tabs {
  background: #fff;
  border-radius: var(--shop-radius-lg) var(--shop-radius-lg) 0 0;
  border: 1px solid var(--shop-border);
  border-bottom: none;
  padding: 0 1rem;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 16px;
  padding: 16px;
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
}

.order-detail {
  padding: 16px 24px;
  background: hsl(220, 14%, 97%);
}

.detail-header {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  margin-bottom: 12px;
}

.detail-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #fff;
  border-radius: var(--shop-radius-md);
  border: 1px solid var(--shop-border);
}

.item-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.item-name {
  font-size: 0.875rem;
  font-weight: 500;
}

.item-spec {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.item-qty {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.item-price {
  font-size: 0.875rem;
  font-weight: 600;
}

.order-id-text {
  font-weight: 600;
  color: var(--shop-primary);
  font-family: var(--shop-font-mono);
  font-size: 0.8125rem;
}

.product-summary {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
}

.amount-text {
  font-weight: 600;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}
</style>
