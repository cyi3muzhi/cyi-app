<template>
  <div class="checkout-page">
    <!-- 返回上一页 -->
    <div class="container-shop py-3">
      <BackButton fallback="/cart" />
    </div>

    <div class="container-shop pb-12">
      <!-- 步骤指示器 -->
      <div class="step-indicator">
        <div class="step-item">
          <div class="step-circle done">✓</div>
          <span class="step-label done">我的购物车</span>
        </div>
        <div class="step-connector"></div>
        <div class="step-item">
          <div class="step-circle active">2</div>
          <span class="step-label active">确认订单</span>
        </div>
        <div class="step-connector"></div>
        <div class="step-item">
          <div class="step-circle pending">3</div>
          <span class="step-label pending">完成支付</span>
        </div>
      </div>

      <!-- 收货地址 -->
      <section class="checkout-card">
        <div class="card-header">
          <h2 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
            </svg>
            收货地址
          </h2>
          <router-link to="/user/address" class="add-address-link">+ 新增地址</router-link>
        </div>

        <div v-if="addressLoading" class="section-state">正在加载地址...</div>
        <div v-else-if="addresses.length === 0" class="empty-address">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
          </svg>
          <p>还没有收货地址，请先新增</p>
          <router-link to="/user/address" class="go-address-btn">去新增地址</router-link>
        </div>
        <div v-else class="address-list">
          <div
            v-for="addr in addresses"
            :key="addr.id"
            class="address-card"
            :class="{ 'address-active': selectedAddressId === addr.id }"
            @click="selectedAddressId = addr.id"
          >
            <div class="address-radio">
              <span class="radio-dot"></span>
            </div>
            <div class="address-info">
              <div class="address-name-row">
                <span class="address-receiver">{{ addr.receiver }}</span>
                <span class="address-phone">{{ addr.phone }}</span>
                <span v-if="addr.isDefault === 1" class="default-tag">默认</span>
              </div>
              <p class="address-detail">{{ addr.province }}{{ addr.city }}{{ addr.district }}{{ addr.detail }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 商品清单 -->
      <section class="checkout-card">
        <div class="card-header">
          <h2 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
            商品清单
          </h2>
          <span class="card-count">{{ selectedItems.length }} 种商品</span>
        </div>

        <div class="goods-list">
          <div v-for="item in selectedItems" :key="item.productId + '-' + item.skuId" class="goods-row">
            <div class="goods-image-wrap">
              <img :src="item.image" :alt="item.name" @error="onImageError" />
            </div>
            <div class="goods-info">
              <h3 class="goods-name">{{ item.name }}</h3>
              <p class="goods-sku">{{ item.sku }}</p>
            </div>
            <div class="goods-price">&yen;{{ item.price.toFixed(2) }}</div>
            <div class="goods-qty">x{{ item.quantity }}</div>
          </div>
        </div>
      </section>

      <!-- 金额汇总 -->
      <section class="checkout-card">
        <div class="summary-rows">
          <div class="summary-row">
            <span class="summary-label">商品小计</span>
            <span class="summary-value">&yen;{{ subtotal.toFixed(2) }}</span>
          </div>
          <!-- 优惠券行：点击打开选券弹层 -->
          <div class="summary-row coupon-summary-row" @click="openCouponPicker">
            <span class="summary-label coupon-label">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.83z"/><line x1="7" y1="7" x2="7.01" y2="7"/>
              </svg>
              优惠券
            </span>
            <span class="summary-value" :class="selectedCoupon ? 'text-discount' : 'text-placeholder'">
              <template v-if="selectedCoupon">-&yen;{{ selectedCoupon.discount / 100 }}</template>
              <template v-else>{{ usableCount > 0 ? `${usableCount} 张可用` : '暂无可用' }}</template>
              <svg class="chev" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </span>
          </div>
          <div class="summary-row">
            <span class="summary-label">运费</span>
            <span class="summary-value" :class="shipping === 0 ? 'text-success' : ''">
              {{ shipping === 0 ? '免运费' : '¥' + shipping.toFixed(2) }}
            </span>
          </div>
          <div class="summary-row" v-if="selectedCoupon">
            <span class="summary-label">优惠抵扣</span>
            <span class="summary-value text-discount">-&yen;{{ couponDiscount.toFixed(2) }}</span>
          </div>
          <div class="summary-row total-row">
            <span class="total-label">应付总额</span>
            <span class="total-price">&yen;{{ total.toFixed(2) }}</span>
          </div>
        </div>

        <button class="submit-btn" :disabled="submitting || selectedItems.length === 0" @click="onSubmit">
          <span v-if="submitting" class="btn-spinner"></span>
          {{ submitting ? '提交中...' : '提交订单' }}
        </button>
      </section>
    </div>

    <!-- 选券弹层：单选可用券/不使用，不可用券置灰提示差额 -->
    <div v-if="couponDialogVisible" class="pick-overlay" @click.self="closeCouponPicker">
      <div class="pick-sheet">
        <div class="pick-head">
          <h3>选择优惠券</h3>
          <button class="pick-close" aria-label="关闭" @click="closeCouponPicker">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <p class="pick-sub">已按当前订单（&yen;{{ subtotal.toFixed(2) }}）筛选优惠券</p>

        <div v-if="couponLoading" class="pick-loading">
          <span class="loading-dot"></span>加载中...
        </div>
        <template v-else>
          <div
            v-for="c in usableCoupons"
            :key="c.recordId"
            class="pick-item"
            :class="{ selected: pendingRecordId === c.recordId, unusable: !c.usable }"
            @click="selectCoupon(c)"
          >
            <span class="pick-radio"></span>
            <span class="pick-face">-&yen;{{ pickFace(c) }}</span>
            <div class="pick-info">
              <p class="name">{{ c.name }}</p>
              <p v-if="c.usable" class="cond">{{ thresholdText(c) }} · 有效期至 {{ c.validEnd || '永久' }}</p>
              <p v-else class="cond short">还差 &yen;{{ formatMoney(c.gapAmount) }} 可用</p>
            </div>
            <span v-if="c.usable" class="pick-tag">可用</span>
          </div>

          <!-- 不使用优惠券 -->
          <div class="pick-none" :class="{ selected: !pendingRecordId }" @click="pendingRecordId = 0">
            <span class="pick-radio"></span>
            <span style="font-size: 0.875rem; font-weight: 600;">不使用优惠券</span>
          </div>

          <button class="pick-confirm" :disabled="couponLoading" @click="confirmCoupon">确定</button>
        </template>
      </div>
    </div>

    <!-- 付款弹窗：下单成功后弹出，立即支付或稍后支付 -->
    <div v-if="payDialogVisible" class="pay-overlay" @click.self="closePayDialog">
      <div class="pay-modal">
        <button class="pay-close" aria-label="关闭" @click="closePayDialog">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>

        <div class="pay-icon">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
        </div>

        <h3 class="pay-title">订单提交成功</h3>
        <p class="pay-sub">请尽快完成支付，逾期未支付订单将自动取消</p>

        <div class="pay-info">
          <div class="pay-row">
            <span class="pay-row-label">订单号</span>
            <span class="pay-row-value">{{ payOrderNo || '生成中...' }}</span>
          </div>
          <div class="pay-row">
            <span class="pay-row-label">应付金额</span>
            <span class="pay-row-value pay-amount">&yen;{{ payAmount.toFixed(2) }}</span>
          </div>
        </div>

        <button class="pay-now-btn" :disabled="paying" @click="onPay">
          <span v-if="paying" class="btn-spinner"></span>
          {{ paying ? '支付中...' : '立即支付' }}
        </button>
        <button class="pay-later-btn" :disabled="paying" @click="closePayDialog">稍后支付</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 确认订单（结算）页面
 * 选择收货地址 → 确认商品清单 → 提交订单（购物车结算下单）
 * 下单成功后弹出付款弹窗：立即支付或稍后支付（稍后支付进入待付款列表）
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/store/cart'
import { listAddress } from '@/api/user'
import { cartCheckout, payOrder, getOrderDetail } from '@/api/order'
import { listUsableCoupons } from '@/api/coupon'
import { deleteCartItem } from '@/api/cart'
import BackButton from '@/components/BackButton.vue'
import { toast } from '@/utils/toast'
import type { Address } from '@/types/user'
import type { UsableCouponItem } from '@/types/coupon'

const router = useRouter()
const cartStore = useCartStore()

// ---- 地址 ----
const addresses = ref<Address[]>([])
const selectedAddressId = ref<number | null>(null)
const addressLoading = ref(true)

/** 选中的地址 */
const selectedAddress = computed(() => addresses.value.find(a => a.id === selectedAddressId.value))

/** 加载收货地址（默认地址优先选中） */
async function loadAddresses() {
  addressLoading.value = true
  try {
    const res = await listAddress()
    addresses.value = res.items || []
    const def = addresses.value.find(a => a.isDefault === 1) || addresses.value[0]
    if (def) selectedAddressId.value = def.id
  } catch {
    addresses.value = []
  } finally {
    addressLoading.value = false
  }
}

// ---- 商品与金额（来自购物车选中的商品） ----
const selectedItems = computed(() => cartStore.selectedItems)

const subtotal = computed(() => cartStore.subtotal)

/** 运费：满 199 免运费，否则 10 元（与购物车页一致） */
const shipping = computed(() => {
  if (selectedItems.value.length > 0 && subtotal.value < 199) return 10
  return 0
})

// ---- 优惠券（结算选券） ----
const usableCoupons = ref<UsableCouponItem[]>([])
const couponLoading = ref(false)
const couponDialogVisible = ref(false)
const couponLoadAmount = ref(0) // 已加载时对应的商品金额（分），金额变化需重新加载
const selectedRecordId = ref(0) // 确认选中的发放记录 ID（0=不使用）
const pendingRecordId = ref(0) // 弹层中临时选中（确定后才生效）

/** 当前选中的可用券 */
const selectedCoupon = computed<UsableCouponItem | null>(
  () => usableCoupons.value.find(c => c.recordId === selectedRecordId.value && c.usable) || null
)

/** 优惠金额（元） */
const couponDiscount = computed(() => (selectedCoupon.value ? selectedCoupon.value.discount / 100 : 0))

/** 可用券数量（满足门槛且在有效期内） */
const usableCount = computed(() => usableCoupons.value.filter(c => c.usable).length)

/** 应付总额（商品小计 + 运费 - 优惠） */
const total = computed(() => Math.max(0, subtotal.value + shipping.value - couponDiscount.value))

/**
 * 加载结算可用券（按当前商品金额筛选，金额变化时重新拉取）
 */
async function loadUsableCoupons() {
  const amountCents = Math.round(subtotal.value * 100)
  // 金额未变化且已加载过则复用
  if (couponLoadAmount.value === amountCents && usableCoupons.value.length > 0) return
  couponLoading.value = true
  try {
    const res = await listUsableCoupons(amountCents)
    usableCoupons.value = res.items || []
    couponLoadAmount.value = amountCents
  } catch {
    usableCoupons.value = []
  } finally {
    couponLoading.value = false
  }
}

/** 打开选券弹层 */
async function openCouponPicker() {
  await loadUsableCoupons()
  pendingRecordId.value = selectedRecordId.value
  couponDialogVisible.value = true
  document.body.style.overflow = 'hidden'
}

/** 关闭选券弹层（不保存临时选择） */
function closeCouponPicker() {
  couponDialogVisible.value = false
  document.body.style.overflow = ''
}

/** 选择券（仅可用券可点，不可用券置灰） */
function selectCoupon(c: UsableCouponItem) {
  if (!c.usable) return
  pendingRecordId.value = c.recordId
}

/** 确定：回写选中并关闭弹层 */
function confirmCoupon() {
  selectedRecordId.value = pendingRecordId.value
  closeCouponPicker()
}

/**
 * 弹层面值展示（分 → 元，整数不带小数）
 * @param c 可用券项
 */
function pickFace(c: UsableCouponItem): string {
  const cents = c.usable ? c.discount : (c.discount || c.value)
  return formatMoney(cents)
}

/** 使用门槛：满 x 可用 / 全场通用 */
function thresholdText(c: UsableCouponItem): string {
  return c.minAmount > 0 ? `满${formatMoney(c.minAmount)}可用` : '全场通用'
}

/**
 * 金额格式化（分 → 元，整数不带小数）
 * @param cents 金额（分）
 */
function formatMoney(cents: number): string {
  const yuan = cents / 100
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
}

// ---- 提交订单 ----
const submitting = ref(false)

// ---- 付款弹窗 ----
const payDialogVisible = ref(false)
const payOrderId = ref(0)
const payOrderNo = ref('')
const payAmount = ref(0) // 元
const paying = ref(false)

/**
 * 提交订单：购物车结算下单
 * 成功后将购物车中已结算商品移除，并弹出付款弹窗
 */
async function onSubmit() {
  if (!selectedAddress.value) {
    toast('请先选择收货地址', 'error')
    return
  }
  if (submitting.value) return
  submitting.value = true
  try {
    const addr = selectedAddress.value
    const res = await cartCheckout({
      items: selectedItems.value.map(i => ({ product_id: i.productId, sku_id: i.skuId, quantity: i.quantity })),
      address_id: addr.id,
      receiver: addr.receiver,
      phone: addr.phone,
      address: `${addr.province || ''}${addr.city || ''}${addr.district || ''}${addr.detail || ''}`,
      freight: Math.round(shipping.value * 100), // 元 → 分
      coupon_id: selectedRecordId.value // 优惠券发放记录 ID（0=不使用）
    })
    // 下单成功：先清理购物车中已结算商品，再弹出付款
    await Promise.all(selectedItems.value.map(i => deleteCartItem(i.productId, i.skuId)))
    await cartStore.loadCart()
    openPayDialog(res.orderId)
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '下单失败，请重试', 'error')
  } finally {
    submitting.value = false
  }
}

/**
 * 打开付款弹窗：先用结算页金额展示，异步拉取订单详情补充订单号与真实金额
 * @param orderId 订单ID
 */
async function openPayDialog(orderId: number) {
  payOrderId.value = orderId
  payOrderNo.value = ''
  payAmount.value = total.value // 兜底金额（元）
  payDialogVisible.value = true
  try {
    const detail = await getOrderDetail(orderId)
    if (detail.order?.orderNo) payOrderNo.value = detail.order.orderNo
    if (detail.order?.amount) payAmount.value = detail.order.amount / 100 // 分 → 元
  } catch {
    // 详情拉取失败时使用结算页金额兜底
  }
}

/**
 * 立即支付：调用支付接口，成功后跳转订单列表（待发货）
 */
async function onPay() {
  if (paying.value) return
  paying.value = true
  try {
    await payOrder(payOrderId.value)
    toast('支付成功', 'success')
    payDialogVisible.value = false
    router.push('/user/orders?status=pending_ship')
  } catch (e: any) {
    // 支付失败留在弹窗，可重试或稍后支付
    toast(e?.response?.data?.message || e?.message || '支付失败，请稍后重试', 'error')
  } finally {
    paying.value = false
  }
}

/**
 * 稍后支付/关闭弹窗：进入待付款列表
 */
function closePayDialog() {
  if (paying.value) return
  payDialogVisible.value = false
  router.push('/user/orders?status=pending_payment')
}

/** 商品图片加载失败占位 */
function onImageError(e: Event) {
  ;(e.target as HTMLImageElement).style.display = 'none'
}

// 页面加载：无选中商品则回到购物车，否则加载地址
onMounted(() => {
  if (!selectedItems.value.length) {
    router.replace('/cart')
    return
  }
  loadAddresses()
})
</script>

<style scoped>

/* ---- 步骤指示器 ---- */
.step-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 24px;
  overflow-x: auto;
  padding: 8px 0;
}
.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.step-circle {
  width: 32px;
  height: 32px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  font-weight: 600;
  flex-shrink: 0;
}
.step-circle.active {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}
.step-circle.done {
  background: var(--shop-success);
  color: var(--shop-success-foreground);
}
.step-circle.pending {
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
}
.step-label {
  font-size: 0.875rem;
  white-space: nowrap;
}
.step-label.active {
  color: var(--shop-primary);
  font-weight: 600;
}
.step-label.done {
  color: var(--shop-foreground);
}
.step-label.pending {
  color: var(--shop-muted-foreground);
}
.step-connector {
  width: 48px;
  height: 1px;
  background: var(--shop-border);
  margin: 0 8px;
}

/* ---- 卡片通用 ---- */
.checkout-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
  padding: 20px;
  margin-bottom: 20px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1rem;
  font-weight: 700;
  color: var(--shop-foreground);
}
.card-title svg {
  color: var(--shop-primary);
}
.add-address-link {
  font-size: 0.8125rem;
  color: var(--shop-primary);
  text-decoration: none;
  transition: opacity var(--shop-transition);
}
.add-address-link:hover {
  opacity: 0.8;
}
.card-count {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.section-state {
  padding: 24px 0;
  text-align: center;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

/* ---- 地址 ---- */
.empty-address {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 32px 0;
}
.empty-address p {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
.go-address-btn {
  padding: 8px 24px;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.8125rem;
  text-decoration: none;
  transition: opacity var(--shop-transition);
}
.go-address-btn:hover {
  opacity: 0.9;
}

.address-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.address-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  cursor: pointer;
  transition: border-color var(--shop-transition), background var(--shop-transition);
}
.address-card:hover {
  border-color: var(--shop-primary-light);
}
.address-active {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
}
.address-radio {
  display: flex;
  align-items: center;
  padding-top: 4px;
  flex-shrink: 0;
}
.radio-dot {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid var(--shop-border);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color var(--shop-transition);
}
.address-active .radio-dot {
  border-color: var(--shop-primary);
  background: radial-gradient(circle, var(--shop-primary) 45%, transparent 50%);
}
.address-info {
  flex: 1;
  min-width: 0;
}
.address-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}
.address-receiver {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}
.address-phone {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
.default-tag {
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 0.6875rem;
  font-weight: 600;
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}
.address-detail {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  line-height: 1.5;
}

/* ---- 商品清单 ---- */
.goods-list {
  display: flex;
  flex-direction: column;
}
.goods-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--shop-border);
}
.goods-row:last-child {
  border-bottom: none;
}
.goods-image-wrap {
  width: 64px;
  height: 64px;
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
.goods-sku {
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
  min-width: 48px;
  text-align: right;
}

/* ---- 金额汇总 ---- */
.summary-rows {
  margin-bottom: 16px;
}
.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 0.875rem;
}
.summary-label {
  color: var(--shop-muted-foreground);
}
.summary-value {
  font-weight: 500;
  color: var(--shop-foreground);
}
.text-success {
  color: var(--shop-success) !important;
}

/* ---- 优惠券行 ---- */
.coupon-summary-row {
  cursor: pointer;
  border-radius: var(--shop-radius-md);
  padding-left: 8px;
  padding-right: 8px;
  transition: background var(--shop-transition);
}
.coupon-summary-row:hover {
  background: var(--shop-muted);
}
.coupon-label {
  display: flex;
  align-items: center;
  gap: 6px;
}
.coupon-label svg {
  color: var(--shop-primary);
}
.text-discount {
  color: var(--shop-destructive) !important;
  font-weight: 700;
}
.text-placeholder {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--shop-muted-foreground);
  cursor: pointer;
}
.chev {
  color: var(--shop-muted-foreground);
  flex-shrink: 0;
}

/* ---- 选券弹层（底部弹层） ---- */
.pick-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(3px);
  animation: fadeIn 0.25s ease;
  padding: 0;
}
.pick-sheet {
  width: 100%;
  max-width: 560px;
  background: var(--shop-card);
  border-radius: 20px 20px 0 0;
  padding: 20px 20px 24px;
  max-height: 78vh;
  overflow: auto;
  box-shadow: 0 -10px 40px rgb(0 0 0 / 0.18);
  animation: sheetIn 0.3s cubic-bezier(0.2, 0.7, 0.3, 1);
}
@keyframes sheetIn {
  from { transform: translateY(30px); }
  to { transform: translateY(0); }
}
.pick-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.pick-head h3 {
  font-size: 1rem;
  font-weight: 700;
}
.pick-close {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background var(--shop-transition), color var(--shop-transition);
}
.pick-close:hover {
  background: var(--shop-border);
  color: var(--shop-foreground);
}
.pick-sub {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 12px;
}
.pick-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 32px 0;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.loading-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid var(--shop-primary-bg);
  border-top-color: var(--shop-primary);
  animation: spin 0.8s linear infinite;
}
.pick-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1.5px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  margin-bottom: 10px;
  cursor: pointer;
  transition: all var(--shop-transition);
  background: var(--shop-card);
}
.pick-item:hover {
  border-color: var(--shop-primary-light);
}
.pick-item.selected {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
  box-shadow: 0 0 0 1px var(--shop-primary);
}
.pick-item.unusable {
  opacity: 0.55;
  cursor: not-allowed;
}
.pick-radio {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--shop-border);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--shop-transition);
}
.pick-item.selected .pick-radio {
  border-color: var(--shop-primary);
}
.pick-item.selected .pick-radio::after {
  content: '';
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--shop-primary);
}
.pick-face {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--shop-destructive);
  width: 64px;
  flex-shrink: 0;
}
.pick-info {
  flex: 1;
  min-width: 0;
}
.pick-info .name {
  font-size: 0.875rem;
  font-weight: 600;
}
.pick-info .cond {
  font-size: 0.72rem;
  color: var(--shop-muted-foreground);
  margin-top: 3px;
}
.pick-info .cond.short {
  color: var(--shop-destructive);
  font-weight: 600;
}
.pick-tag {
  font-size: 0.625rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--shop-primary);
  color: #fff;
  flex-shrink: 0;
}
.pick-none {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: var(--shop-radius-lg);
  cursor: pointer;
  background: var(--shop-muted);
  margin-bottom: 4px;
}
.pick-none.selected {
  outline: 2px solid var(--shop-primary);
  background: var(--shop-primary-bg);
}
.pick-none.selected .pick-radio {
  border-color: var(--shop-primary);
}
.pick-none.selected .pick-radio::after {
  content: '';
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--shop-primary);
}
.pick-confirm {
  width: 100%;
  height: 46px;
  margin-top: 14px;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: linear-gradient(135deg, var(--shop-primary), var(--shop-primary-dark));
  color: #fff;
  font-size: 0.9375rem;
  font-weight: 600;
  box-shadow: 0 4px 14px hsla(217, 91%, 52%, 0.35);
  transition: opacity var(--shop-transition);
  cursor: pointer;
  font-family: var(--shop-font-sans);
}
.pick-confirm:hover:not(:disabled) {
  opacity: 0.92;
}
.pick-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.total-row {
  padding-top: 12px;
  border-top: 1px solid var(--shop-border);
}
.total-label {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}
.total-price {
  font-size: 1.375rem;
  font-weight: 700;
  color: var(--shop-destructive);
}

/* ---- 提交按钮 ---- */
.submit-btn {
  width: 100%;
  height: 48px;
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
  box-shadow: 0 4px 14px hsla(217, 91%, 52%, 0.35);
  transition: opacity 0.2s, transform 0.2s, box-shadow 0.2s;
  font-family: var(--shop-font-sans);
}
.submit-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-2px);
  box-shadow: 0 6px 18px hsla(217, 91%, 52%, 0.45);
}
.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-spinner {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ---- 付款弹窗 ---- */
.pay-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
  animation: fadeIn 0.2s ease;
  padding: 20px;
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
.pay-modal {
  position: relative;
  width: 100%;
  max-width: 400px;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: 0 20px 60px rgba(15, 23, 42, 0.35);
  padding: 32px 28px 28px;
  text-align: center;
  animation: modalIn 0.25s ease;
}
@keyframes modalIn {
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
  transition: background var(--shop-transition), color var(--shop-transition);
}
.pay-close:hover {
  background: var(--shop-border);
  color: var(--shop-foreground);
}
.pay-icon {
  width: 72px;
  height: 72px;
  margin: 0 auto 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}
.pay-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin-bottom: 8px;
}
.pay-sub {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 20px;
}
.pay-info {
  background: var(--shop-muted);
  border-radius: var(--shop-radius-lg);
  padding: 14px 16px;
  margin-bottom: 20px;
}
.pay-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 0.875rem;
}
.pay-row-label {
  color: var(--shop-muted-foreground);
}
.pay-row-value {
  font-weight: 600;
  color: var(--shop-foreground);
}
.pay-amount {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-destructive);
}
.pay-now-btn {
  width: 100%;
  height: 46px;
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
  box-shadow: 0 4px 14px hsla(217, 91%, 52%, 0.35);
  transition: opacity 0.2s, transform 0.2s;
  font-family: var(--shop-font-sans);
}
.pay-now-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-2px);
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
  transition: border-color var(--shop-transition), color var(--shop-transition);
  font-family: var(--shop-font-sans);
}
.pay-later-btn:hover:not(:disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}
.pay-later-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
