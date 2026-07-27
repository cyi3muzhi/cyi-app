<template>
  <!-- ============================================================ -->
  <!-- 1. 页面标题栏                                                  -->
  <!-- ============================================================ -->
  <div class="page-header">
    <div class="container-shop">
      <div class="page-header-inner">
        <router-link to="/user" class="back-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </router-link>
        <h1 class="page-title">我的订单</h1>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 2. 订单状态标签页                                              -->
  <!-- ============================================================ -->
  <div class="tabs-bar">
    <div class="container-shop">
      <div class="tabs-scroll">
        <button
          v-for="tab in orderTabs"
          :key="tab.key"
          class="tab-btn"
          :class="{ 'tab-active': activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
          <span v-if="tab.count > 0" class="tab-count">{{ tab.count }}</span>
        </button>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 3. 订单卡片列表                                                -->
  <!-- ============================================================ -->
  <div class="order-list">
    <div class="container-shop">
      <!-- 空状态 -->
      <div v-if="filteredOrders.length === 0" class="empty-state">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="var(--shop-border)" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
          <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>
        </svg>
        <p class="empty-text">暂无相关订单</p>
        <router-link to="/products" class="go-shop-btn">去逛逛</router-link>
      </div>

      <!-- 订单卡片 -->
      <div v-for="order in pagedOrders" :key="order.id" class="order-card">
        <!-- 订单头部 -->
        <div class="order-card-header">
          <div class="order-meta">
            <span class="order-no">订单号：{{ order.orderNo }}</span>
            <span class="order-time">{{ order.createdAt }}</span>
          </div>
          <span class="order-status" :style="{ color: statusColor(order.status) }">
            {{ statusLabel(order.status) }}
          </span>
        </div>

        <!-- 商品列表 -->
        <div class="order-items">
          <div
            v-for="(item, index) in order.items"
            :key="index"
            class="order-item"
          >
            <div class="item-image">
              <img :src="item.image" :alt="item.name" />
            </div>
            <div class="item-info">
              <p class="item-name">{{ item.name }}</p>
              <p class="item-price">
                <span class="price">&yen;{{ formatAmount(item.price) }}</span>
                <span class="quantity">x{{ item.quantity }}</span>
              </p>
            </div>
          </div>
        </div>

        <!-- 订单底部 -->
        <div class="order-card-footer">
          <div class="order-total">
            共 {{ order.items.length }} 件商品 &nbsp; 合计：
            <span class="total-amount">&yen;{{ formatAmount(order.totalAmount) }}</span>
          </div>
          <div class="order-actions">
            <button
              v-for="action in getActions(order.status)"
              :key="action.label"
              :class="['action-btn', action.primary ? 'action-primary' : 'action-secondary']"
              @click="onOrderAction(order, action.key)"
            >
              {{ action.label }}
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="pagination">
        <button
          class="page-btn"
          :disabled="currentPage <= 1"
          @click="goPage(currentPage - 1)"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
          上一页
        </button>
        <button
          v-for="p in visiblePages"
          :key="p"
          class="page-btn"
          :class="{ 'page-active': p === currentPage }"
          @click="goPage(p)"
        >
          {{ p }}
        </button>
        <button
          class="page-btn"
          :disabled="currentPage >= totalPages"
          @click="goPage(currentPage + 1)"
        >
          下一页
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 订单列表页面
 * 展示用户所有订单，支持按状态筛选、分页、操作
 */
import { ref, computed } from 'vue'
import type { Order, OrderStatus } from '@/types/order'

// ---- 订单标签页 ----
interface OrderTab {
  key: string
  label: string
  count: number
}

const activeTab = ref('all')
const currentPage = ref(1)
const pageSize = 5

const orderTabs: OrderTab[] = [
  { key: 'all', label: '全部', count: 0 },
  { key: 'pending_payment', label: '待付款', count: 2 },
  { key: 'pending_ship', label: '待发货', count: 1 },
  { key: 'pending_receive', label: '待收货', count: 0 },
  { key: 'pending_review', label: '待评价', count: 3 },
  { key: 'completed', label: '已完成', count: 0 }
]

/**
 * 切换标签页
 * @param key - 标签页标识
 */
function switchTab(key: string) {
  activeTab.value = key
  currentPage.value = 1
}

// ---- Mock 订单数据 ----
const mockOrders: Order[] = [
  {
    id: 1,
    orderNo: 'YS202607270001',
    status: 'pending_payment',
    totalAmount: 2599.00,
    actualAmount: 2599.00,
    quantity: 1,
    items: [
      { productId: 101, name: 'Apple iPhone 15 Pro Max 256GB', image: 'https://via.placeholder.com/80x80/007bff/fff?text=iPhone', price: 2599.00, quantity: 1 }
    ],
    createdAt: '2026-07-27 14:30:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  },
  {
    id: 2,
    orderNo: 'YS202607250002',
    status: 'pending_ship',
    totalAmount: 1899.00,
    actualAmount: 1799.00,
    quantity: 2,
    items: [
      { productId: 201, name: 'Sony WH-1000XM5 无线降噪耳机', image: 'https://via.placeholder.com/80x80/28a745/fff?text=Sony', price: 1299.00, quantity: 1 },
      { productId: 202, name: 'JBL Flip 6 便携蓝牙音箱', image: 'https://via.placeholder.com/80x80/dc3545/fff?text=JBL', price: 600.00, quantity: 1 }
    ],
    createdAt: '2026-07-25 10:15:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  },
  {
    id: 3,
    orderNo: 'YS202607200003',
    status: 'pending_receive',
    totalAmount: 459.00,
    actualAmount: 459.00,
    quantity: 3,
    items: [
      { productId: 301, name: '无印良品 舒适家居服套装', image: 'https://via.placeholder.com/80x80/ffc107/000?text=MUJI', price: 199.00, quantity: 1 },
      { productId: 302, name: '网易严选 纯棉毛巾礼盒装', image: 'https://via.placeholder.com/80x80/17a2b8/fff?text=YY', price: 89.00, quantity: 2 }
    ],
    createdAt: '2026-07-20 09:00:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  },
  {
    id: 4,
    orderNo: 'YS202607180004',
    status: 'pending_review',
    totalAmount: 299.00,
    actualAmount: 279.00,
    quantity: 1,
    items: [
      { productId: 401, name: '米家 空气净化器 Pro H', image: 'https://via.placeholder.com/80x80/6f42c1/fff?text=MI', price: 279.00, quantity: 1 }
    ],
    createdAt: '2026-07-18 16:45:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  },
  {
    id: 5,
    orderNo: 'YS202607150005',
    status: 'pending_review',
    totalAmount: 189.00,
    actualAmount: 169.00,
    quantity: 1,
    items: [
      { productId: 501, name: '三只松鼠 坚果大礼包 1.5kg', image: 'https://via.placeholder.com/80x80/20c997/fff?text=3SS', price: 169.00, quantity: 1 }
    ],
    createdAt: '2026-07-15 11:20:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  },
  {
    id: 6,
    orderNo: 'YS202607100006',
    status: 'completed',
    totalAmount: 3999.00,
    actualAmount: 3799.00,
    quantity: 1,
    items: [
      { productId: 601, name: 'MacBook Air M3 13.6英寸 8GB 256GB', image: 'https://via.placeholder.com/80x80/fd7e14/fff?text=Mac', price: 3799.00, quantity: 1 }
    ],
    createdAt: '2026-07-10 08:00:00',
    address: '北京市朝阳区xxx路xxx号',
    receiver: '张小明',
    phone: '138****1234'
  }
]

/**
 * 根据当前标签页过滤订单
 */
const filteredOrders = computed(() => {
  if (activeTab.value === 'all') return mockOrders
  return mockOrders.filter(o => o.status === activeTab.value)
})

/**
 * 计算总页数
 */
const totalPages = computed(() => Math.ceil(filteredOrders.value.length / pageSize))

/**
 * 计算当前页显示的订单
 */
const pagedOrders = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredOrders.value.slice(start, start + pageSize)
})

/**
 * 计算显示的页码按钮
 */
const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value
  let start = Math.max(1, currentPage.value - 2)
  let end = Math.min(total, start + 4)
  if (end - start < 4) {
    start = Math.max(1, end - 4)
  }
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

/**
 * 跳转到指定页
 * @param page - 目标页码
 */
function goPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

/**
 * 获取订单状态标签文字
 * @param status - 订单状态
 * @returns 状态文字
 */
function statusLabel(status: OrderStatus): string {
  const map: Record<OrderStatus, string> = {
    pending_payment: '待付款',
    pending_ship: '待发货',
    pending_receive: '待收货',
    pending_review: '待评价',
    completed: '已完成',
    cancelled: '已取消',
    refund: '退款中'
  }
  return map[status] || '未知'
}

/**
 * 获取订单状态颜色
 * @param status - 订单状态
 * @returns 颜色值
 */
function statusColor(status: OrderStatus): string {
  const map: Record<OrderStatus, string> = {
    pending_payment: 'var(--shop-warning)',
    pending_ship: 'var(--shop-primary)',
    pending_receive: 'var(--shop-primary)',
    pending_review: 'var(--shop-warning)',
    completed: 'var(--shop-success)',
    cancelled: 'var(--shop-muted-foreground)',
    refund: 'var(--shop-destructive)'
  }
  return map[status] || 'var(--shop-muted-foreground)'
}

/**
 * 获取订单可执行操作按钮
 * @param status - 订单状态
 * @returns 操作按钮列表
 */
function getActions(status: OrderStatus): Array<{ key: string; label: string; primary: boolean }> {
  const map: Record<string, Array<{ key: string; label: string; primary: boolean }>> = {
    pending_payment: [
      { key: 'cancel', label: '取消订单', primary: false },
      { key: 'pay', label: '去付款', primary: true }
    ],
    pending_ship: [
      { key: 'remind', label: '提醒发货', primary: false }
    ],
    pending_receive: [
      { key: 'logistics', label: '查看物流', primary: false },
      { key: 'confirm', label: '确认收货', primary: true }
    ],
    pending_review: [
      { key: 'logistics', label: '查看物流', primary: false },
      { key: 'review', label: '去评价', primary: true }
    ],
    completed: [
      { key: 'detail', label: '查看详情', primary: false },
      { key: 'again', label: '再次购买', primary: true }
    ],
    cancelled: [
      { key: 'detail', label: '查看详情', primary: false }
    ],
    refund: [
      { key: 'detail', label: '查看详情', primary: false }
    ]
  }
  return map[status] || []
}

/**
 * 格式化金额
 * @param amount - 金额数值
 * @returns 格式化后的金额字符串
 */
function formatAmount(amount: number): string {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

/**
 * 处理订单操作
 * @param order - 订单对象
 * @param action - 操作标识
 */
function onOrderAction(order: Order, action: string) {
  console.log(`订单操作: ${order.orderNo} -> ${action}`)
  // 实际项目跳转到对应页面或弹出确认框
}
</script>

<style scoped>
/* ============================================
   订单列表页面样式
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

.back-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-foreground);
  transition: background 0.2s;
}

.back-link:hover {
  background: var(--shop-border);
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

/* ---- 标签页栏 ---- */
.tabs-bar {
  padding: 16px 0 0;
  border-bottom: 1px solid var(--shop-border);
}

.tabs-scroll {
  display: flex;
  gap: 0;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.tabs-scroll::-webkit-scrollbar {
  display: none;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  font-size: 0.875rem;
  font-weight: 500;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
  white-space: nowrap;
}

.tab-btn:hover {
  color: var(--shop-foreground);
}

.tab-active {
  color: var(--shop-primary);
  border-bottom-color: var(--shop-primary);
}

.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  font-size: 0.6875rem;
  font-weight: 700;
}

.tab-active .tab-count {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}

/* ---- 订单列表 ---- */
.order-list {
  padding: 16px 0 48px;
}

/* ---- 空状态 ---- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 64px 24px;
}

.empty-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.go-shop-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 24px;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: opacity 0.2s;
}

.go-shop-btn:hover {
  opacity: 0.9;
}

/* ---- 订单卡片 ---- */
.order-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  margin-bottom: 16px;
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.order-card:hover {
  box-shadow: var(--shop-shadow-md);
}

.order-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid var(--shop-border);
  background: var(--shop-muted);
}

.order-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.order-no {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.order-time {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.order-status {
  font-size: 0.875rem;
  font-weight: 600;
}

/* ---- 商品列表 ---- */
.order-items {
  padding: 16px 20px;
}

.order-item {
  display: flex;
  gap: 12px;
  padding: 8px 0;
}

.order-item + .order-item {
  border-top: 1px solid var(--shop-border);
  padding-top: 12px;
}

.item-image {
  flex-shrink: 0;
  width: 72px;
  height: 72px;
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
  background: var(--shop-muted);
}

.item-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-width: 0;
}

.item-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.item-price {
  display: flex;
  align-items: center;
  gap: 8px;
}

.price {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.quantity {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 订单底部 ---- */
.order-card-footer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid var(--shop-border);
}

@media (min-width: 640px) {
  .order-card-footer {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.order-total {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.total-amount {
  font-size: 1rem;
  font-weight: 700;
  color: var(--shop-destructive);
}

.order-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-btn {
  padding: 6px 16px;
  border-radius: var(--shop-radius-lg);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.action-secondary {
  background: var(--shop-background);
  border: 1px solid var(--shop-border);
  color: var(--shop-foreground);
}

.action-secondary:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.action-primary {
  background: var(--shop-primary);
  border: 1px solid var(--shop-primary);
  color: var(--shop-primary-foreground);
}

.action-primary:hover {
  opacity: 0.9;
}

/* ---- 分页 ---- */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 0;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  min-width: 36px;
  height: 36px;
  border-radius: var(--shop-radius-md);
  border: 1px solid var(--shop-border);
  background: var(--shop-card);
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: all 0.2s;
  justify-content: center;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-active {
  background: var(--shop-primary);
  border-color: var(--shop-primary);
  color: var(--shop-primary-foreground);
}

.page-active:hover {
  color: var(--shop-primary-foreground);
}

/* ---- 响应式 ---- */
@media (max-width: 480px) {
  .order-card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .order-actions {
    width: 100%;
  }

  .action-btn {
    flex: 1;
    text-align: center;
  }
}
</style>
