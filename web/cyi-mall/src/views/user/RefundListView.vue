<template>
  <div class="refund-list-page">
    <!-- ============================================================ -->
    <!-- 1. 页面标题栏                                                  -->
    <!-- ============================================================ -->
    <div class="page-header">
      <div class="container-shop">
        <div class="page-header-inner">
          <BackButton fallback="/user" />
          <h1 class="page-title">退款/售后</h1>
        </div>
      </div>
    </div>

    <div class="container-shop pb-8">
      <!-- 状态筛选 -->
      <div class="filter-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.value"
          :class="['filter-tab', activeTab === tab.value ? 'filter-tab-active' : '']"
          @click="switchTab(tab.value)"
        >{{ tab.label }}</button>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="state-box">
        <div class="loading-spinner"></div>
        <p class="state-text">正在加载退款记录...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="!items.length" class="state-box">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/>
        </svg>
        <p class="state-text">{{ activeTab === '0' ? '暂无退款记录' : '该状态下暂无退款记录' }}</p>
      </div>

      <!-- 退款记录列表 -->
      <template v-else>
        <div v-for="item in items" :key="item.refundId" class="refund-card">
          <div class="refund-card-header">
            <span class="refund-order-no">订单号 {{ item.orderNo }}</span>
            <span :class="['refund-status', statusClass(item.status)]">{{ item.statusText }}</span>
          </div>
          <div class="refund-card-body">
            <div class="refund-row">
              <span class="refund-label">退款金额</span>
              <span class="refund-money">&yen;{{ (item.refundAmount / 100).toFixed(2) }}</span>
            </div>
            <div class="refund-row">
              <span class="refund-label">退款原因</span>
              <span class="refund-value">{{ item.reason || '-' }}</span>
            </div>
            <div class="refund-row">
              <span class="refund-label">申请时间</span>
              <span class="refund-value">{{ item.createdAt }}</span>
            </div>
            <div v-if="item.status === 3 && item.rejectReason" class="refund-row">
              <span class="refund-label">拒绝原因</span>
              <span class="refund-value reject-text">{{ item.rejectReason }}</span>
            </div>
          </div>
        </div>

        <!-- 加载更多/分页 -->
        <div v-if="hasMore" class="load-more-wrap">
          <button class="load-more-btn" :disabled="loadingMore" @click="loadMore">
            {{ loadingMore ? '加载中...' : '加载更多' }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 我的退款/售后页面
 * 展示当前用户的退款记录列表（按状态筛选、分页加载）
 * 数据来源：GET /order/v1/refunds
 */
import { ref, computed, onMounted } from 'vue'
import BackButton from '@/components/BackButton.vue'
import { getRefunds, type RefundItemDTO } from '@/api/order'

/** 状态筛选标签（value 与后端 refund status 对应，0=全部） */
const tabs = [
  { label: '全部', value: '0' },
  { label: '待审核', value: '1' },
  { label: '已通过', value: '2' },
  { label: '已拒绝', value: '3' }
]

const activeTab = ref('0')
const items = ref<RefundItemDTO[]>([])
const loading = ref(true)
const loadingMore = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 10

/** 是否还有更多数据 */
const hasMore = computed(() => items.value.length < total.value)

/**
 * 切换状态筛选
 * @param value - 状态值（0=全部）
 */
function switchTab(value: string) {
  if (activeTab.value === value) return
  activeTab.value = value
  page.value = 1
  items.value = []
  loadList()
}

/**
 * 加载退款列表（按当前筛选状态）
 */
async function loadList() {
  loading.value = true
  try {
    const res = await getRefunds({
      page: page.value,
      pageSize,
      ...(activeTab.value !== '0' ? { status: activeTab.value } : {})
    })
    items.value = res.items || []
    total.value = res.total || 0
  } catch {
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

/**
 * 加载更多（追加下一页数据）
 */
async function loadMore() {
  if (loadingMore.value) return
  loadingMore.value = true
  try {
    page.value += 1
    const res = await getRefunds({
      page: page.value,
      pageSize,
      ...(activeTab.value !== '0' ? { status: activeTab.value } : {})
    })
    items.value = items.value.concat(res.items || [])
    total.value = res.total || 0
  } finally {
    loadingMore.value = false
  }
}

/**
 * 退款状态样式类
 * @param status - 状态: 1-待审核 2-已通过 3-已拒绝
 */
function statusClass(status: number): string {
  const map: Record<number, string> = {
    1: 'status-pending',
    2: 'status-approved',
    3: 'status-rejected'
  }
  return map[status] || ''
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
/* ============================================
   我的退款/售后页面样式
   ============================================ */

/* ---- 页面标题 ---- */
.page-header {
  padding: 24px 0 0;
}
.page-header-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}
.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

/* ---- 状态筛选 ---- */
.filter-tabs {
  display: flex;
  gap: 8px;
  padding: 18px 0 14px;
  overflow-x: auto;
}
.filter-tab {
  flex-shrink: 0;
  padding: 7px 18px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  background: var(--shop-card);
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--shop-font-sans);
}
.filter-tab:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}
.filter-tab-active {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  font-weight: 600;
}

/* ---- 退款记录卡片 ---- */
.refund-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
  margin-bottom: 14px;
  overflow: hidden;
}
.refund-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--shop-border);
}
.refund-order-no {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.refund-status {
  font-size: 0.8125rem;
  font-weight: 700;
}
.status-pending {
  color: var(--shop-warning);
}
.status-approved {
  color: var(--shop-success);
}
.status-rejected {
  color: var(--shop-destructive);
}
.refund-card-body {
  padding: 6px 18px 14px;
}
.refund-row {
  display: flex;
  align-items: flex-start;
  padding: 7px 0;
  font-size: 0.875rem;
}
.refund-label {
  width: 84px;
  flex-shrink: 0;
  color: var(--shop-muted-foreground);
}
.refund-value {
  flex: 1;
  color: var(--shop-foreground);
  word-break: break-all;
  line-height: 1.5;
}
.refund-money {
  font-size: 1.0625rem;
  font-weight: 700;
  color: var(--shop-destructive);
}
.reject-text {
  color: var(--shop-destructive);
}

/* ---- 加载更多 ---- */
.load-more-wrap {
  display: flex;
  justify-content: center;
  padding: 12px 0 24px;
}
.load-more-btn {
  padding: 9px 32px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  background: var(--shop-card);
  color: var(--shop-foreground);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--shop-font-sans);
}
.load-more-btn:hover:not(:disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}
.load-more-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ---- 加载/空状态 ---- */
.state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 80px 0;
}
.loading-spinner {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 3px solid var(--shop-border);
  border-top-color: var(--shop-primary);
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.state-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
</style>
