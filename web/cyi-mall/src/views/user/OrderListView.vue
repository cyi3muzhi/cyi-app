<template>
  <!-- ============================================================ -->
  <!-- 1. 页面标题栏                                                  -->
  <!-- ============================================================ -->
  <div class="page-header">
    <div class="container-shop">
      <div class="page-header-inner">
        <BackButton fallback="/user" />
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
  <!-- 3. 搜索栏（按订单号模糊搜索）                                    -->
  <!-- ============================================================ -->
  <div class="search-bar">
    <div class="container-shop">
      <div class="search-inner">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="输入订单号搜索"
          @keyup.enter="onSearch"
        />
        <button v-if="keyword" class="search-clear" aria-label="清空搜索" @click="clearSearch">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
        <button class="search-btn" @click="onSearch">搜索</button>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 4. 订单卡片列表                                                -->
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
      <div v-for="order in pagedOrders" :key="order.id" class="order-card" @click="goDetail(order.id)">
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

        <!-- 超时自动取消提示（仅待付款订单，含支付倒计时） -->
        <div v-if="order.status === 'pending_payment'" class="timeout-tip">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
          <span>剩余支付时间 <b class="countdown">{{ countdownTextOf(order) }}</b>，超时未支付将自动取消</span>
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
              @click.stop="onOrderAction(order, action.key)"
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

  <!-- 物流信息弹窗 -->
  <div v-if="logisticsVisible" class="logistics-overlay" @click.self="closeLogistics">
    <div class="logistics-modal">
      <button class="logistics-close" aria-label="关闭" @click="closeLogistics">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>

      <div class="logistics-icon">
        <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <rect x="1" y="3" width="15" height="13" rx="1"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
        </svg>
      </div>
      <h3 class="logistics-title">查看物流</h3>

      <div v-if="logisticsLoading" class="logistics-state">正在加载...</div>
      <div v-else-if="logisticsInfo && logisticsInfo.logisticsCompany" class="logistics-body">
        <div class="logistics-row">
          <span class="logistics-label">物流公司</span>
          <span class="logistics-value">{{ logisticsInfo.logisticsCompany }}</span>
        </div>
        <div class="logistics-row">
          <span class="logistics-label">快递单号</span>
          <span class="logistics-value">{{ logisticsInfo.logisticsNo }}</span>
        </div>
        <div class="logistics-divider"></div>
        <div class="logistics-row">
          <span class="logistics-label">收货人</span>
          <span class="logistics-value">{{ logisticsInfo.receiver || '-' }}</span>
        </div>
        <div class="logistics-row">
          <span class="logistics-label">联系电话</span>
          <span class="logistics-value">{{ logisticsInfo.phone || '-' }}</span>
        </div>
        <div class="logistics-row address-row">
          <span class="logistics-label">收货地址</span>
          <span class="logistics-value">{{ logisticsInfo.address || '-' }}</span>
        </div>
      </div>
      <div v-else class="logistics-state">该订单暂未发货，暂无物流信息</div>

      <button class="logistics-confirm" @click="closeLogistics">知道了</button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 订单列表页面
 * 展示用户所有订单，支持按状态筛选、分页、操作
 * 数据来源：GET /order/v1/list（当前登录用户订单，含商品明细）
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BackButton from '@/components/BackButton.vue'
import { getOrders, getOrderDetail, receiveOrder, cancelOrder, PAY_TIMEOUT_SECONDS, type OrderItemDTO } from '@/api/order'
import { getUserCenter } from '@/api/user'
import { useCartStore } from '@/store/cart'
import { toast } from '@/utils/toast'
import type { Order, OrderStatus } from '@/types/order'

// ---- 订单标签页 ----
interface OrderTab {
  key: string
  label: string
  count: number
}

// tab key -> 后端订单状态数字（逗号分隔，空=全部）
// 状态 2 已支付/3 待发货 → 待发货；状态 4 已发货/5 待收货 → 待收货
const statusParamMap: Record<string, string> = {
  all: '',
  pending_payment: '1',
  pending_ship: '2,3',
  pending_receive: '4,5',
  pending_review: '6',
  completed: '6',
  cancelled: '7',
  refund: '8,9'
}

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const activeTab = ref('all')
const currentPage = ref(1)
const pageSize = 5
const orders = ref<Order[]>([])
const total = ref(0)
const loading = ref(true)
// 搜索关键字（按订单号模糊搜索）
const keyword = ref('')

const orderTabs = ref<OrderTab[]>([
  { key: 'all', label: '全部', count: 0 },
  { key: 'pending_payment', label: '待付款', count: 0 },
  { key: 'pending_ship', label: '待发货', count: 0 },
  { key: 'pending_receive', label: '待收货', count: 0 },
  { key: 'pending_review', label: '待评价', count: 0 },
  { key: 'completed', label: '已完成', count: 0 },
  { key: 'cancelled', label: '已取消', count: 0 },
  { key: 'refund', label: '退款/售后', count: 0 }
])

/**
 * 切换标签页：重置到第一页并重新请求
 * @param key - 标签页标识
 */
function switchTab(key: string) {
  if (activeTab.value === key) return
  activeTab.value = key
  currentPage.value = 1
  fetchOrders()
}

/**
 * 将接口返回的订单项转换为页面展示结构（金额由分转为元）
 * @param item - 接口订单项
 * @returns 页面订单对象
 */
function mapOrder(item: OrderItemDTO): Order {
  return {
    id: item.id,
    orderNo: item.orderNo,
    status: item.status as OrderStatus,
    totalAmount: item.amount / 100,
    actualAmount: item.amount / 100,
    quantity: item.products.reduce((sum, p) => sum + p.qty, 0),
    payDeadline: item.payDeadline,
    items: item.products.map(p => ({
      productId: p.productId ?? 0,
      skuId: p.skuId ?? 0,
      name: p.name,
      image: p.image,
      price: p.price / 100,
      quantity: p.qty
    })),
    createdAt: item.createdAt,
    address: '',
    receiver: '',
    phone: ''
  }
}

/**
 * 请求订单列表（接口按用户 + 状态过滤并分页）
 */
async function fetchOrders() {
  loading.value = true
  try {
    const status = statusParamMap[activeTab.value] ?? ''
    const res = await getOrders({ page: currentPage.value, pageSize, status, keyword: keyword.value.trim() })
    orders.value = (res.items || []).map(mapOrder)
    total.value = res.total || 0
  } catch {
    orders.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

/**
 * 执行搜索（按订单号模糊查询，重置到第一页）
 */
function onSearch() {
  currentPage.value = 1
  fetchOrders()
}

/**
 * 清空搜索关键字并刷新列表
 */
function clearSearch() {
  if (!keyword.value) return
  keyword.value = ''
  currentPage.value = 1
  fetchOrders()
}

/**
 * 根据当前标签页过滤订单（接口已按状态过滤，这里直接返回）
 */
const filteredOrders = computed(() => orders.value)

/**
 * 计算总页数
 */
const totalPages = computed(() => Math.ceil(total.value / pageSize))

/**
 * 计算当前页显示的订单（接口已分页）
 */
const pagedOrders = computed(() => orders.value)

/**
 * 计算显示的页码按钮
 */
const visiblePages = computed(() => {
  const pages: number[] = []
  const totalPagesCount = totalPages.value
  let start = Math.max(1, currentPage.value - 2)
  let end = Math.min(totalPagesCount, start + 4)
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
  fetchOrders()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

// 页面加载：优先使用个人中心跳转携带的状态参数，并拉取统计填充角标
onMounted(() => {
  const qStatus = route.query.status as string
  if (qStatus && orderTabs.value.some(t => t.key === qStatus)) {
    activeTab.value = qStatus
  }
  fetchOrders()
  // 填充各状态角标数字（来自个人中心统计接口，失败不影响列表）
  getUserCenter().then((center) => {
    const s = center.stats
    orderTabs.value = orderTabs.value.map(t => ({
      ...t,
      count: t.key === 'pending_payment' ? s.pendingPay
        : t.key === 'pending_ship' ? s.pendingShip
        : t.key === 'pending_receive' ? s.pendingReceive
        : t.key === 'pending_review' ? s.pendingReview
        : t.key === 'refund' ? s.refundCount
        : 0
    }))
  }).catch(() => { /* 忽略统计失败 */ })

  // 启动待付款倒计时计时器（每秒刷新）
  // 有待付款订单倒计时归零时自动刷新列表（后端 task 服务已自动取消该订单）
  countdownTimer = window.setInterval(() => {
    now.value = Date.now()
    const expired = orders.value.some(o => o.status === 'pending_payment' && payDeadlineOf(o) <= now.value)
    if (expired) fetchOrders()
  }, 1000)
})

// 页面卸载时清理倒计时计时器，避免内存泄漏
onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})

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
    refunding: '退款中',
    refunded: '已退款',
    refund: '退款/售后'
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
    refunding: 'var(--shop-warning)',
    refunded: 'var(--shop-muted-foreground)',
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
    refunding: [
      { key: 'detail', label: '查看详情', primary: false }
    ],
    refunded: [
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

// ---- 待付款订单支付倒计时 ----
/** 当前时间（每秒刷新，驱动倒计时渲染） */
const now = ref(Date.now())
let countdownTimer: number | undefined

/**
 * 解析后端时间字符串（YYYY-MM-DD HH:mm:ss）为时间戳
 * @param s - 时间字符串
 * @returns Unix 毫秒时间戳
 */
function parseOrderTime(s: string): number {
  // 替换 - 为 /，兼容 iOS Safari 的 Date 解析
  return new Date(s.replace(/-/g, '/')).getTime()
}

/**
 * 计算待付款订单的支付截止时间戳
 * 优先使用后端返回的 payDeadline（服务端时钟，与超时判定同源）；
 * 兼容旧数据回退为 下单时间 + 30 分钟
 * @param order - 订单对象
 * @returns 截止时间（毫秒）
 */
function payDeadlineOf(order: Order): number {
  if (order.payDeadline) return order.payDeadline * 1000
  return parseOrderTime(order.createdAt) + PAY_TIMEOUT_SECONDS * 1000
}

/**
 * 格式化剩余支付时间（MM:SS）
 * @param order - 订单对象
 * @returns 倒计时文本，已超时返回 00:00
 */
function countdownTextOf(order: Order): string {
  const remain = Math.max(0, payDeadlineOf(order) - now.value)
  const totalSec = Math.floor(remain / 1000)
  const mm = String(Math.floor(totalSec / 60)).padStart(2, '0')
  const ss = String(totalSec % 60).padStart(2, '0')
  return `${mm}:${ss}`
}

/**
 * 处理订单操作
 * @param order - 订单对象
 * @param action - 操作标识
 */
function onOrderAction(order: Order, action: string) {
  if (action === 'detail') {
    goDetail(order.id)
    return
  }
  if (action === 'logistics') {
    onViewLogistics(order)
    return
  }
  if (action === 'confirm') {
    onConfirmReceive(order)
    return
  }
  if (action === 'cancel') {
    onCancelOrder(order)
    return
  }
  if (action === 'pay') {
    onPayOrder(order)
    return
  }
  if (action === 'again') {
    onBuyAgain(order)
    return
  }
  if (action === 'review') {
    toast('评价功能开发中，敬请期待')
    return
  }
  if (action === 'remind') {
    onRemindShip(order)
    return
  }
}

/**
 * 去付款：跳转到订单详情页并自动打开支付弹窗（?action=pay）
 * @param order - 订单对象
 */
function onPayOrder(order: Order) {
  router.push(`/user/orders/${order.id}?action=pay`)
}

/**
 * 再次购买：把订单商品加入购物车后跳转购物车页
 * @param order - 订单对象
 */
async function onBuyAgain(order: Order) {
  try {
    for (const item of order.items) {
      await cartStore.addItem({
        productId: item.productId,
        skuId: item.skuId || 0,
        name: item.name,
        image: item.image,
        price: item.price,
        quantity: item.quantity
      })
    }
    toast('已加入购物车', 'success')
    router.push('/cart')
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '加入购物车失败', 'error')
  }
}

/**
 * 提醒发货（暂无独立接口，先给用户即时反馈）
 * @param order - 订单对象
 */
function onRemindShip(order: Order) {
  toast(`已提醒商家尽快发货`, 'success')
}

/**
 * 跳转订单详情页
 * @param orderId - 订单ID
 */
function goDetail(orderId: number) {
  router.push(`/user/orders/${orderId}`)
}

/**
 * 取消订单（二次确认）
 * @param order - 订单对象
 */
async function onCancelOrder(order: Order) {
  if (!window.confirm('确定取消该订单吗？取消后将释放库存。')) return
  try {
    await cancelOrder(order.id)
    toast('订单已取消', 'success')
    fetchOrders()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '取消订单失败，请重试', 'error')
  }
}

/**
 * 确认收货（二次确认）
 * @param order - 订单对象
 */
async function onConfirmReceive(order: Order) {
  if (!window.confirm('请确认已收到货品？确认后将完成订单。')) return
  try {
    await receiveOrder(order.id)
    toast('确认收货成功', 'success')
    fetchOrders()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '确认收货失败，请重试', 'error')
  }
}

// ---- 物流信息弹窗 ----
const logisticsVisible = ref(false)
const logisticsLoading = ref(false)
const logisticsInfo = ref<OrderItemDTO | null>(null)

/**
 * 查看物流：拉取订单详情展示物流公司与单号
 * @param order - 订单对象
 */
async function onViewLogistics(order: Order) {
  logisticsVisible.value = true
  logisticsLoading.value = true
  logisticsInfo.value = null
  try {
    const res = await getOrderDetail(order.id)
    logisticsInfo.value = res.order
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '获取物流信息失败', 'error')
    logisticsVisible.value = false
  } finally {
    logisticsLoading.value = false
  }
}

/** 关闭物流弹窗 */
function closeLogistics() {
  logisticsVisible.value = false
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

/* ---- 搜索栏 ---- */
.search-bar {
  padding: 16px 0 0;
}

.search-inner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 16px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.search-inner:focus-within {
  border-color: var(--shop-primary);
  box-shadow: var(--shop-shadow-sm);
}

.search-input {
  flex: 1;
  height: 38px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 0.875rem;
  color: var(--shop-foreground);
}

.search-input::placeholder {
  color: var(--shop-muted-foreground);
}

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--shop-muted-foreground);
  cursor: pointer;
}

.search-clear:hover {
  background: var(--shop-muted);
  color: var(--shop-foreground);
}

.search-btn {
  padding: 8px 20px;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.search-btn:hover {
  opacity: 0.9;
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
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s;
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

/* ---- 超时自动取消提示 ---- */
.timeout-tip {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  padding: 8px 20px 0;
  color: var(--shop-warning);
  font-size: 0.75rem;
}

.countdown {
  font-family: var(--shop-font-mono, monospace);
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--shop-destructive);
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
/* ---- 物流信息弹窗 ---- */
.logistics-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
  animation: logisticsFade 0.2s ease;
  padding: 20px;
}
@keyframes logisticsFade {
  from { opacity: 0; }
  to { opacity: 1; }
}
.logistics-modal {
  position: relative;
  width: 100%;
  max-width: 400px;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: 0 20px 60px rgba(15, 23, 42, 0.35);
  padding: 28px 24px 24px;
  text-align: center;
  animation: logisticsModalIn 0.25s ease;
}
@keyframes logisticsModalIn {
  from { opacity: 0; transform: translateY(16px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.logistics-close {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: background var(--shop-transition), color var(--shop-transition);
}
.logistics-close:hover {
  background: var(--shop-border);
  color: var(--shop-foreground);
}
.logistics-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}
.logistics-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin-bottom: 16px;
}
.logistics-state {
  padding: 20px 0;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
.logistics-body {
  background: var(--shop-muted);
  border-radius: var(--shop-radius-lg);
  padding: 12px 16px;
  margin-bottom: 18px;
  text-align: left;
}
.logistics-row {
  display: flex;
  align-items: flex-start;
  padding: 6px 0;
  font-size: 0.875rem;
}
.logistics-label {
  width: 76px;
  flex-shrink: 0;
  color: var(--shop-muted-foreground);
}
.logistics-value {
  flex: 1;
  color: var(--shop-foreground);
  font-weight: 500;
  word-break: break-all;
}
.logistics-divider {
  height: 1px;
  background: var(--shop-border);
  margin: 8px 0;
}
.address-row .logistics-value {
  line-height: 1.5;
}
.logistics-confirm {
  width: 100%;
  height: 44px;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: linear-gradient(135deg, var(--shop-primary) 0%, var(--shop-primary-dark) 100%);
  color: var(--shop-primary-foreground);
  font-size: 0.9375rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s;
  font-family: var(--shop-font-sans);
}
.logistics-confirm:hover {
  opacity: 0.92;
  transform: translateY(-2px);
}
</style>
