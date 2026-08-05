<template>
  <div class="order-detail-page">
    <!-- ============================================================ -->
    <!-- 1. 页面标题栏                                                  -->
    <!-- ============================================================ -->
    <div class="page-header">
      <div class="container-shop">
        <div class="page-header-inner">
          <BackButton fallback="/user/orders" />
          <h1 class="page-title">订单详情</h1>
        </div>
      </div>
    </div>

    <!-- ============================================================ -->
    <!-- 2. 加载 / 错误 / 内容                                          -->
    <!-- ============================================================ -->
    <div class="container-shop pb-8">
      <!-- 加载中 -->
      <div v-if="loading" class="state-box">
        <div class="loading-spinner"></div>
        <p class="state-text">正在加载订单详情...</p>
      </div>

      <!-- 加载失败 -->
      <div v-else-if="error" class="state-box">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        <p class="state-text">{{ error }}</p>
        <button class="retry-btn" @click="loadDetail">重试</button>
      </div>

      <!-- 详情内容 -->
      <template v-else-if="order">
        <!-- 状态横幅 -->
        <div class="status-banner">
          <span class="status-dot" :style="{ background: statusColor }"></span>
          <span class="status-text" :style="{ color: statusColor }">{{ order.statusText || statusLabel(order.status) }}</span>
          <span class="status-order-no">{{ order.orderNo }}</span>
        </div>

        <!-- 超时自动取消提示（仅待付款订单，含支付倒计时） -->
        <div v-if="order.status === 'pending_payment'" class="timeout-tip">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
          <span>剩余支付时间 <b class="countdown">{{ countdownText() }}</b>，超时未支付订单将自动取消</span>
        </div>

        <!-- 商品明细 -->
        <section class="detail-card">
          <div class="card-header">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
            <span class="card-title">商品明细</span>
            <span class="card-subtitle">共 {{ order.products.length }} 件商品</span>
          </div>

          <div class="goods-list">
            <div v-for="(p, i) in order.products" :key="i" class="goods-row">
              <div class="goods-image-wrap">
                <img :src="p.image" :alt="p.name" @error="onImageError" />
              </div>
              <div class="goods-info">
                <p class="goods-name">{{ p.name }}</p>
                <p class="goods-spec">{{ p.spec || '默认规格' }}</p>
              </div>
              <div class="goods-price">&yen;{{ (p.price / 100).toFixed(2) }}</div>
              <div class="goods-qty">x{{ p.qty }}</div>
            </div>
          </div>

          <!-- 金额汇总 -->
          <div class="amount-summary">
            <div class="amount-row">
              <span class="amount-label">商品总额</span>
              <span class="amount-value">&yen;{{ productTotal.toFixed(2) }}</span>
            </div>
            <div class="amount-row total-row">
              <span class="amount-label">实付金额</span>
              <span class="amount-value total-amount">&yen;{{ (order.amount / 100).toFixed(2) }}</span>
            </div>
          </div>
        </section>

        <!-- 收货与物流信息 -->
        <section class="detail-card">
          <div class="card-header">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
            </svg>
            <span class="card-title">收货与物流</span>
          </div>

          <div class="info-list">
            <div class="info-row">
              <span class="info-label">收货人</span>
              <span class="info-value">{{ order.receiver || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">联系电话</span>
              <span class="info-value">{{ order.phone || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">收货地址</span>
              <span class="info-value">{{ order.address || '-' }}</span>
            </div>
          </div>

          <!-- 物流轨迹时间线（发货后展示） -->
          <template v-if="order.logisticsCompany">
            <div class="info-divider"></div>
            <div class="info-row">
              <span class="info-label">物流公司</span>
              <span class="info-value">{{ order.logisticsCompany }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">快递单号</span>
              <span class="info-value mono-text">{{ order.logisticsNo || '-' }}</span>
            </div>

            <!-- 轨迹加载中 -->
            <div v-if="logisticsLoading" class="logistics-loading">
              <span class="mini-spinner"></span> 正在加载物流轨迹...
            </div>

            <!-- 轨迹时间线 -->
            <div v-else-if="logistics && logistics.traces.length" class="logistics-timeline">
              <div
                v-for="(t, i) in [...logistics.traces].reverse()"
                :key="i"
                class="trace-item"
              >
                <div class="trace-node">
                  <span
                    class="trace-dot"
                    :style="{ background: i === 0 ? logisticsStatusColor(t.status) : 'var(--shop-border)' }"
                  ></span>
                  <span v-if="i !== [...logistics.traces].length - 1" class="trace-line"></span>
                </div>
                <div class="trace-content">
                  <p class="trace-desc" :style="{ color: i === 0 ? logisticsStatusColor(t.status) : 'var(--shop-foreground)' }">
                    {{ t.desc }}
                    <span v-if="i === 0" class="trace-current">最新</span>
                  </p>
                  <p class="trace-status">
                    <span :style="{ color: logisticsStatusColor(t.status) }">{{ logisticsStatusText(t.status) }}</span>
                    <span class="trace-time">{{ t.timeText }}</span>
                  </p>
                </div>
              </div>
            </div>

            <!-- 无轨迹数据兜底 -->
            <div v-else class="logistics-empty">
              <p>物流轨迹更新中，请稍后刷新查看</p>
              <button class="logistics-refresh" @click="loadLogistics">刷新轨迹</button>
            </div>
          </template>
        </section>

        <!-- 订单信息 -->
        <section class="detail-card">
          <div class="card-header">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>
            </svg>
            <span class="card-title">订单信息</span>
          </div>

          <div class="info-list">
            <div class="info-row">
              <span class="info-label">订单号</span>
              <span class="info-value">{{ order.orderNo }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">下单时间</span>
              <span class="info-value">{{ order.createdAt }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">订单状态</span>
              <span class="info-value" :style="{ color: statusColor }">{{ order.statusText || statusLabel(order.status) }}</span>
            </div>
          </div>
        </section>

        <!-- 操作区（按状态展示可执行操作） -->
        <div v-if="orderActions.length" class="action-bar">
          <button
            v-for="act in orderActions"
            :key="act.key"
            :class="['detail-action-btn', act.primary ? 'action-primary' : 'action-secondary']"
            @click="onAction(act.key)"
          >
            {{ act.label }}
          </button>
        </div>
      </template>
    </div>
  </div>

  <!-- 支付弹窗（待付款订单去付款） -->
  <div v-if="payVisible" class="pay-overlay" @click.self="closePayDialog">
    <div class="pay-modal">
      <button class="pay-close" aria-label="关闭" @click="closePayDialog">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
      <div class="pay-icon">
        <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>
        </svg>
      </div>
      <h3 class="pay-title">订单支付</h3>
      <p class="pay-amount">&yen;{{ ((order?.amount || 0) / 100).toFixed(2) }}</p>
      <p class="pay-order-no">订单号：{{ order?.orderNo }}</p>
      <button class="pay-now-btn" :disabled="paying" @click="onPay">
        <span v-if="paying" class="btn-spinner"></span>
        {{ paying ? '支付中...' : '立即支付' }}
      </button>
      <button class="pay-later-btn" :disabled="paying" @click="closePayDialog">稍后支付</button>
    </div>
  </div>

  <!-- 退款申请弹窗（待发货订单申请退款） -->
  <div v-if="refundVisible" class="pay-overlay" @click.self="closeRefundDialog">
    <div class="pay-modal">
      <button class="pay-close" aria-label="关闭" @click="closeRefundDialog">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
      <div class="pay-icon refund-icon">
        <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/>
        </svg>
      </div>
      <h3 class="pay-title">申请退款</h3>
      <p class="refund-tip">订单支付后商家尚未发货，可申请全额退款</p>
      <div class="refund-amount">退款金额 <b>&yen;{{ ((order?.amount || 0) / 100).toFixed(2) }}</b></div>

      <!-- 退款原因选择 -->
      <div class="refund-reasons">
        <button
          v-for="opt in refundReasons"
          :key="opt"
          :class="['refund-reason', refundReason === opt ? 'refund-reason-active' : '']"
          @click="refundReason = opt"
        >{{ opt }}</button>
      </div>

      <button class="pay-now-btn" :disabled="submitting" @click="onSubmitRefund">
        <span v-if="submitting" class="btn-spinner"></span>
        {{ submitting ? '提交中...' : '提交申请' }}
      </button>
      <button class="pay-later-btn" :disabled="submitting" @click="closeRefundDialog">取消</button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 订单详情页面
 * 展示订单状态、商品明细、金额、收货与物流信息
 * 数据来源：GET /order/v1/{order_id}（当前登录用户订单）
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BackButton from '@/components/BackButton.vue'
import { getOrderDetail, payOrder, receiveOrder, cancelOrder, applyRefund, getLogistics, PAY_TIMEOUT_SECONDS, type OrderItemDTO, type LogisticsDTO } from '@/api/order'
import { toast } from '@/utils/toast'
import type { OrderStatus } from '@/types/order'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref('')
const order = ref<OrderItemDTO | null>(null)
/** 物流信息（发货后加载） */
const logistics = ref<LogisticsDTO | null>(null)
/** 物流加载中 */
const logisticsLoading = ref(false)

/** 当前时间（每秒刷新，驱动待付款倒计时渲染） */
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
 * 计算支付截止时间戳（毫秒）
 * 优先使用后端返回的 payDeadline（服务端时钟），兼容旧数据回退为 下单时间 + 30 分钟
 */
function payDeadlineOf(): number {
  const o = order.value
  if (!o) return 0
  if (o.payDeadline) return o.payDeadline * 1000
  return parseOrderTime(o.createdAt) + PAY_TIMEOUT_SECONDS * 1000
}

/**
 * 格式化剩余支付时间（MM:SS）
 * @returns 倒计时文本，已超时返回 00:00
 */
function countdownText(): string {
  if (!order.value?.createdAt) return '00:00'
  const remain = Math.max(0, payDeadlineOf() - now.value)
  const totalSec = Math.floor(remain / 1000)
  const mm = String(Math.floor(totalSec / 60)).padStart(2, '0')
  const ss = String(totalSec % 60).padStart(2, '0')
  return `${mm}:${ss}`
}

/** 商品总额（分 → 元） */
const productTotal = computed(() => {
  if (!order.value) return 0
  return order.value.products.reduce((sum, p) => sum + (p.price * p.qty) / 100, 0)
})

/**
 * 加载订单详情
 */
async function loadDetail() {
  loading.value = true
  error.value = ''
  try {
    const res = await getOrderDetail(Number(route.params.id))
    order.value = res.order
    // 已发货/待收货/已完成订单加载物流轨迹
    loadLogistics()
  } catch (e: any) {
    error.value = e?.response?.data?.message || e?.message || '订单不存在或加载失败'
  } finally {
    loading.value = false
  }
}

/**
 * 加载物流轨迹（发货后展示）
 */
async function loadLogistics() {
  const o = order.value
  if (!o || !o.logisticsCompany || o.status === 'pending_ship') {
    logistics.value = null
    return
  }
  logisticsLoading.value = true
  try {
    const res = await getLogistics(Number(route.params.id))
    logistics.value = res
  } catch {
    // 物流接口失败不影响订单详情展示
    logistics.value = null
  } finally {
    logisticsLoading.value = false
  }
}

/** 物流节点状态文字 */
function logisticsStatusText(status: number): string {
  const map: Record<number, string> = {
    1: '已发货',
    2: '运输中',
    3: '派送中',
    4: '已签收'
  }
  return map[status] || ''
}

/** 物流节点状态颜色 */
function logisticsStatusColor(status: number): string {
  const map: Record<number, string> = {
    1: 'var(--shop-primary)',
    2: 'var(--shop-primary)',
    3: 'var(--shop-warning)',
    4: 'var(--shop-success)'
  }
  return map[status] || 'var(--shop-muted-foreground)'
}

/**
 * 订单状态标签文字（接口未返回 statusText 时兜底）
 * @param status - 状态字符串
 */
function statusLabel(status: string): string {
  const map: Record<string, string> = {
    pending_payment: '待付款',
    pending_ship: '待发货',
    pending_receive: '待收货',
    pending_review: '待评价',
    completed: '已完成',
    cancelled: '已取消',
    refunding: '退款中',
    refunded: '已退款',
    refund: '退款中'
  }
  return map[status] || '未知'
}

/**
 * 订单状态颜色
 * @param status - 状态字符串
 */
function statusColorOf(status: string): string {
  const map: Record<string, string> = {
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

const statusColor = computed(() => statusColorOf(order.value?.status || ''))

/**
 * 操作区按钮（按订单状态动态生成）
 */
const orderActions = computed<Array<{ key: string; label: string; primary: boolean }>>(() => {
  const status = order.value?.status || ''
  const map: Record<string, Array<{ key: string; label: string; primary: boolean }>> = {
    pending_payment: [
      { key: 'cancel', label: '取消订单', primary: false },
      { key: 'pay', label: '去付款', primary: true }
    ],
    pending_ship: [
      { key: 'refund', label: '申请退款', primary: false },
      { key: 'remind', label: '提醒发货', primary: false }
    ],
    pending_receive: [
      { key: 'confirm', label: '确认收货', primary: true }
    ],
    pending_review: [
      { key: 'review', label: '去评价', primary: true }
    ],
    completed: [
      { key: 'again', label: '再次购买', primary: true }
    ]
  }
  return map[status] || []
})

/**
 * 处理操作区按钮点击
 * @param action - 操作标识
 */
function onAction(action: string) {
  if (action === 'pay') {
    openPayDialog()
    return
  }
  if (action === 'cancel') {
    onCancelOrder()
    return
  }
  if (action === 'confirm') {
    onConfirmReceive()
    return
  }
  if (action === 'again') {
    router.push('/cart')
    return
  }
  if (action === 'refund') {
    openRefundDialog()
    return
  }
  if (action === 'review') {
    toast('评价功能开发中，敬请期待')
    return
  }
  if (action === 'remind') {
    toast('已提醒商家尽快发货')
  }
}

/**
 * 取消订单（二次确认）
 */
async function onCancelOrder() {
  if (!order.value) return
  if (!window.confirm('确定取消该订单吗？取消后将释放库存。')) return
  try {
    await cancelOrder(order.value.id)
    toast('订单已取消', 'success')
    loadDetail()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '取消订单失败，请重试', 'error')
  }
}

/**
 * 确认收货（二次确认）
 */
async function onConfirmReceive() {
  if (!order.value) return
  if (!window.confirm('请确认已收到货品？确认后将完成订单。')) return
  try {
    await receiveOrder(order.value.id)
    toast('确认收货成功', 'success')
    loadDetail()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '确认收货失败，请重试', 'error')
  }
}

// ---- 支付弹窗（去付款） ----
const payVisible = ref(false)
const paying = ref(false)

/**
 * 打开支付弹窗（从 ?action=pay 或操作区进入）
 */
function openPayDialog() {
  // 非待付款状态不允许支付
  if (order.value?.status !== 'pending_payment') return
  payVisible.value = true
}

/**
 * 关闭支付弹窗（支付中禁止关闭）
 */
function closePayDialog() {
  if (paying.value) return
  payVisible.value = false
}

/**
 * 发起支付（带防抖，重复点击不会重复请求）
 */
async function onPay() {
  if (paying.value || !order.value) return
  paying.value = true
  try {
    await payOrder(order.value.id)
    toast('支付成功', 'success')
    payVisible.value = false
    loadDetail()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '支付失败，请重试', 'error')
  } finally {
    paying.value = false
  }
}

/** 商品图片加载失败占位 */
function onImageError(e: Event) {
  ;(e.target as HTMLImageElement).style.display = 'none'
}

// ---- 退款申请弹窗 ----
const refundVisible = ref(false)
const submitting = ref(false)
const refundReason = ref('')
/** 退款原因选项 */
const refundReasons = ['不想要了', '拍错了', '商品与描述不符', '多拍/重复下单', '其他原因']

/**
 * 打开退款申请弹窗（仅待发货订单可申请）
 */
function openRefundDialog() {
  if (order.value?.status !== 'pending_ship') return
  refundReason.value = ''
  refundVisible.value = true
}

/**
 * 关闭退款申请弹窗（提交中禁止关闭）
 */
function closeRefundDialog() {
  if (submitting.value) return
  refundVisible.value = false
}

/**
 * 提交退款申请（带防抖）
 */
async function onSubmitRefund() {
  if (submitting.value || !order.value) return
  if (!refundReason.value) {
    toast('请选择退款原因', 'error')
    return
  }
  submitting.value = true
  try {
    await applyRefund(order.value.id, refundReason.value)
    toast('退款申请已提交，等待商家审核', 'success')
    refundVisible.value = false
    loadDetail()
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '退款申请失败，请重试', 'error')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadDetail()
  // 从订单列表"去付款"跳转进入时（?action=pay）自动打开支付弹窗
  if (route.query.action === 'pay') {
    openPayDialog()
  }
  // 启动倒计时计时器（每秒刷新）
  countdownTimer = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

// 页面卸载时清理倒计时计时器，避免内存泄漏
onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})
</script>

<style scoped>
/* ============================================
   订单详情页面样式
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

/* ---- 操作区 ---- */
.action-bar {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 4px 0 24px;
}

.detail-action-btn {
  padding: 10px 28px;
  border-radius: var(--shop-radius-lg);
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.detail-action-btn.action-secondary {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  color: var(--shop-foreground);
}

.detail-action-btn.action-secondary:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.detail-action-btn.action-primary {
  background: linear-gradient(135deg, var(--shop-primary) 0%, var(--shop-primary-dark) 100%);
  border: 1px solid var(--shop-primary);
  color: var(--shop-primary-foreground);
}

.detail-action-btn.action-primary:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

/* ---- 支付弹窗 ---- */
.pay-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
  animation: payFade 0.2s ease;
}

@keyframes payFade {
  from { opacity: 0; }
  to { opacity: 1; }
}

.pay-modal {
  position: relative;
  width: 100%;
  max-width: 380px;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: 0 20px 60px rgba(15, 23, 42, 0.35);
  padding: 32px 24px 24px;
  text-align: center;
  animation: payModalIn 0.25s ease;
}

@keyframes payModalIn {
  from { opacity: 0; transform: translateY(16px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.pay-close {
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
  transition: background 0.2s, color 0.2s;
}

.pay-close:hover {
  background: var(--shop-border);
  color: var(--shop-foreground);
}

.pay-icon {
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

.pay-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.pay-amount {
  margin-top: 10px;
  font-size: 1.75rem;
  font-weight: 800;
  color: var(--shop-destructive);
}

.pay-order-no {
  margin-top: 4px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.pay-now-btn {
  width: 100%;
  height: 46px;
  margin-top: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
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

.pay-now-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
}

.pay-now-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.pay-later-btn {
  width: 100%;
  height: 44px;
  margin-top: 10px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  background: transparent;
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
  font-family: var(--shop-font-sans);
}

.pay-later-btn:hover:not(:disabled) {
  color: var(--shop-primary);
  border-color: var(--shop-primary);
}

.pay-later-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  border-radius: 50%;
  animation: paySpin 0.7s linear infinite;
}

@keyframes paySpin {
  to { transform: rotate(360deg); }
}

/* ---- 退款申请弹窗 ---- */
.refund-icon {
  background: var(--shop-destructive-bg, #fef2f2);
  color: var(--shop-destructive);
}
.refund-tip {
  margin-top: 8px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.refund-amount {
  margin-top: 12px;
  font-size: 0.875rem;
  color: var(--shop-foreground);
}
.refund-amount b {
  font-size: 1.375rem;
  color: var(--shop-destructive);
  margin-left: 4px;
}
.refund-reasons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 16px 0 20px;
}
.refund-reason {
  padding: 7px 14px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  background: var(--shop-card);
  color: var(--shop-foreground);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--shop-font-sans);
}
.refund-reason:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}
.refund-reason-active {
  border-color: var(--shop-destructive);
  background: var(--shop-destructive-bg, #fef2f2);
  color: var(--shop-destructive);
  font-weight: 600;
}

/* ---- 加载/错误状态 ---- */
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
  animation: detailSpin 0.8s linear infinite;
}
@keyframes detailSpin {
  to { transform: rotate(360deg); }
}
.state-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
.retry-btn {
  padding: 8px 28px;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: opacity 0.2s;
}
.retry-btn:hover {
  opacity: 0.9;
}

/* ---- 状态横幅 ---- */
.status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 20px;
  margin-bottom: 16px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
}
.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-text {
  font-size: 1.0625rem;
  font-weight: 700;
}
.status-order-no {
  margin-left: auto;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ---- 超时自动取消提示 ---- */
.timeout-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  margin-bottom: 16px;
  background: var(--shop-primary-bg);
  border: 1px dashed var(--shop-primary);
  border-radius: var(--shop-radius-lg);
  color: var(--shop-primary);
  font-size: 0.8125rem;
}

.countdown {
  font-family: var(--shop-font-mono, monospace);
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--shop-destructive);
}

/* ---- 卡片通用 ---- */
.detail-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
  padding: 18px 20px;
  margin-bottom: 16px;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 14px;
  margin-bottom: 14px;
  border-bottom: 1px solid var(--shop-border);
}
.card-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--shop-foreground);
}
.card-subtitle {
  margin-left: auto;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 商品明细 ---- */
.goods-list {
  display: flex;
  flex-direction: column;
}
.goods-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 0;
  border-bottom: 1px solid var(--shop-border);
}
.goods-row:last-child {
  border-bottom: none;
}
.goods-image-wrap {
  width: 60px;
  height: 60px;
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
  background: var(--shop-muted);
  flex-shrink: 0;
}
.goods-image-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.goods-info {
  flex: 1;
  min-width: 0;
}
.goods-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.goods-spec {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
}
.goods-price {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
  white-space: nowrap;
}
.goods-qty {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  min-width: 40px;
  text-align: right;
}

/* ---- 金额汇总 ---- */
.amount-summary {
  padding-top: 14px;
}
.amount-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 0;
  font-size: 0.875rem;
}
.amount-label {
  color: var(--shop-muted-foreground);
}
.amount-value {
  font-weight: 500;
  color: var(--shop-foreground);
}
.total-row {
  padding-top: 10px;
  border-top: 1px solid var(--shop-border);
}
.total-amount {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-destructive);
}

/* ---- 信息列表 ---- */
.info-list {
  display: flex;
  flex-direction: column;
}
.info-row {
  display: flex;
  align-items: flex-start;
  padding: 7px 0;
  font-size: 0.875rem;
}
.info-label {
  width: 84px;
  flex-shrink: 0;
  color: var(--shop-muted-foreground);
}
.info-value {
  flex: 1;
  color: var(--shop-foreground);
  font-weight: 500;
  word-break: break-all;
  line-height: 1.5;
}
.info-divider {
  height: 1px;
  background: var(--shop-border);
  margin: 8px 0;
}

/* ---- 物流轨迹时间线 ---- */
.mono-text {
  font-family: var(--shop-font-mono, monospace);
  letter-spacing: 0.02em;
}

.logistics-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 0 4px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.mini-spinner {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid var(--shop-border);
  border-top-color: var(--shop-primary);
  animation: traceSpin 0.7s linear infinite;
}

@keyframes traceSpin {
  to { transform: rotate(360deg); }
}

.logistics-timeline {
  padding: 6px 0 2px;
}

.trace-item {
  display: flex;
  gap: 12px;
}

.trace-node {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 3px;
}

.trace-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid var(--shop-card);
  box-shadow: 0 0 0 2px var(--shop-border);
  z-index: 1;
}

.trace-item:first-child .trace-dot {
  box-shadow: 0 0 0 3px var(--shop-primary-bg);
}

.trace-line {
  position: absolute;
  top: 16px;
  bottom: -12px;
  width: 2px;
  background: var(--shop-border);
}

.trace-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 14px;
}

.trace-desc {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
  line-height: 1.4;
  display: flex;
  align-items: center;
  gap: 6px;
}

.trace-current {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  font-size: 0.6875rem;
  font-weight: 600;
}

.trace-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 3px;
  font-size: 0.75rem;
}

.trace-time {
  color: var(--shop-muted-foreground);
  font-family: var(--shop-font-mono, monospace);
}

.logistics-empty {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 10px 0 4px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.logistics-refresh {
  padding: 4px 12px;
  border: 1px solid var(--shop-primary);
  border-radius: var(--shop-radius-lg);
  background: transparent;
  color: var(--shop-primary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--shop-font-sans);
}

.logistics-refresh:hover {
  background: var(--shop-primary-bg);
}
</style>
