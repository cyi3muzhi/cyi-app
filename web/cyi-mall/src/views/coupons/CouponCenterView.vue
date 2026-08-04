<template>
  <div class="coupon-center">
    <!-- 面包屑 -->
    <div class="container-shop py-3">
      <div class="center-breadcrumb">
        <router-link to="/" class="breadcrumb-link">首页</router-link>
        <span class="separator">/</span>
        <span class="current">领券中心</span>
      </div>
    </div>

    <div class="container-shop pb-12">
      <!-- 顶部宣传 banner -->
      <div class="center-banner">
        <div class="deco d1"></div>
        <div class="deco d2"></div>
        <div class="deco d3"></div>
        <div class="bb-title">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.83z"/><line x1="7" y1="7" x2="7.01" y2="7"/>
          </svg>
          领券中心
        </div>
        <p class="bb-sub">新人好礼 · 全场通用券限时发放，先到先得</p>
        <div class="bb-ticket">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          当前共 {{ total }} 张好券
        </div>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="section-state">
        <span class="loading-dot"></span>正在加载优惠券...
      </div>

      <!-- 空状态 -->
      <div v-else-if="coupons.length === 0" class="empty-state">
        <svg class="illu" viewBox="0 0 120 120" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="28" y="26" width="64" height="64" rx="12"/>
          <circle cx="60" cy="58" r="16"/>
          <path d="M28 40h64M40 26v14M80 26v14"/>
          <path d="M50 76l7 7 14-14"/>
        </svg>
        <h4>暂无发放中的优惠券</h4>
        <p>快去逛逛，好券不定期上新</p>
      </div>

      <!-- 券列表 -->
      <div v-else class="coupon-list">
        <article
          v-for="(c, idx) in coupons"
          :key="c.couponId"
          class="coupon-card"
          :class="{ 'is-soldout': isSoldOut(c) }"
          :style="{ animationDelay: (idx * 0.05) + 's' }"
        >
          <!-- 已领取：对勾角标 -->
          <span v-if="c.claimed && !isSoldOut(c)" class="corner-check">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
          </span>
          <!-- 已领完：角标 -->
          <span v-if="isSoldOut(c)" class="corner-badge gray">已领完</span>

          <!-- 面值块 -->
          <div class="coupon-face" :class="faceClass(c)">
            <div>
              <template v-if="c.type === 2">
                <span class="face-num">{{ discountText(c) }}</span>
              </template>
              <template v-else>
                <span class="face-symbol">¥</span>
                <span class="face-num">{{ formatMoney(c.value) }}</span>
              </template>
            </div>
            <div class="face-cond">{{ thresholdText(c) }}</div>
          </div>

          <!-- 信息区 -->
          <div class="coupon-body">
            <div class="coupon-info">
              <h3 class="coupon-name">
                {{ c.name }}
                <span v-if="c.type === 2" class="coupon-tag hot">折扣</span>
                <span v-else-if="c.minAmount > 0" class="coupon-tag">满减</span>
                <span v-else class="coupon-tag">立减</span>
              </h3>
              <p class="coupon-desc">{{ typeDesc(c) }}</p>
              <p class="coupon-valid">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/>
                </svg>
                {{ validText(c) }}
              </p>
            </div>
            <div class="coupon-action">
              <button
                v-if="!c.claimed && !isSoldOut(c)"
                class="claim-btn"
                :disabled="claimingId === c.couponId"
                @click="onClaim(c)"
              >
                {{ claimingId === c.couponId ? '领取中...' : '立即领取' }}
              </button>
              <button v-else class="claim-btn" disabled>
                {{ isSoldOut(c) ? '已领完' : '已领取' }}
              </button>
              <span v-if="c.remainCount > 0" class="remain-tip">仅剩 {{ c.remainCount }} 张</span>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 领券中心页面
 * 展示可领取的优惠券列表，支持一键领取
 * 数据来源：GET /api/v1/coupons/available，领取 POST /api/v1/coupons/{coupon_id}/claim
 */
import { ref, onMounted } from 'vue'
import { listAvailableCoupons, claimCoupon } from '@/api/coupon'
import { toast } from '@/utils/toast'
import type { AvailableCouponItem } from '@/types/coupon'

const coupons = ref<AvailableCouponItem[]>([])
const total = ref(0)
const loading = ref(true)
const claimingId = ref(0) // 领取中的券 ID

/** 加载可领取优惠券列表 */
async function loadCoupons() {
  loading.value = true
  try {
    const res = await listAvailableCoupons({ page: 1, pageSize: 50 })
    coupons.value = res.items || []
    total.value = res.total ?? 0
  } catch {
    coupons.value = []
  } finally {
    loading.value = false
  }
}

/** 是否已领完（剩余 0 或剩余量为空） */
function isSoldOut(c: AvailableCouponItem): boolean {
  return c.remainCount <= 0
}

/**
 * 领取优惠券
 * @param c 目标优惠券
 */
async function onClaim(c: AvailableCouponItem) {
  if (claimingId.value) return
  claimingId.value = c.couponId
  try {
    await claimCoupon(c.couponId, 'coupon_center')
    c.claimed = true
    toast('领取成功，去我的优惠券查看', 'success')
  } catch (e: any) {
    toast(e?.response?.data?.message || e?.message || '领取失败，请重试', 'error')
  } finally {
    claimingId.value = 0
  }
}

/** 面值块配色：折扣券红橙渐变，已领完灰色，其余蓝色 */
function faceClass(c: AvailableCouponItem): string {
  if (isSoldOut(c)) return 'face-gray'
  if (c.type === 2) return 'face-red'
  return 'face-blue'
}

/**
 * 金额格式化（分 → 元，整数不带小数）
 * @param cents 金额（分）
 */
function formatMoney(cents: number): string {
  const yuan = cents / 100
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
}

/** 折扣展示：value 为百分比（1-99），如 85 → 8.5折 */
function discountText(c: AvailableCouponItem): string {
  const num = c.value / 10
  const text = Number.isInteger(num) ? String(num) : num.toFixed(1)
  return `${text}折`
}

/** 使用门槛：满 x 可用 / 全场通用 */
function thresholdText(c: AvailableCouponItem): string {
  return c.minAmount > 0 ? `满${formatMoney(c.minAmount)}可用` : '全场通用'
}

/** 类型描述 */
function typeDesc(c: AvailableCouponItem): string {
  switch (c.type) {
    case 2: return '指定商品可用 · 按折扣立省'
    case 3: return '满减券 · 结账时自动抵扣'
    default: return '全场通用 · 不可与其他优惠券叠加'
  }
}

/**
 * 有效期文案
 * 固定日期显示起止，领取后 N 天显示天数，永久显示永久有效
 */
function validText(c: AvailableCouponItem): string {
  switch (c.validType) {
    case 2:
      return `领取后 ${c.validDays} 天内有效`
    case 3:
      return '永久有效'
    default:
      return `${formatDate(c.validStart)} ~ ${formatDate(c.validEnd)}`
  }
}

/**
 * Unix 秒时间戳格式化为日期
 * @param ts 时间戳（秒）
 */
function formatDate(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

onMounted(loadCoupons)
</script>

<style scoped>
/* ---- 面包屑 ---- */
.center-breadcrumb {
  display: flex;
  align-items: center;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}
.breadcrumb-link {
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color var(--shop-transition);
}
.breadcrumb-link:hover {
  color: var(--shop-primary);
}
.separator {
  margin: 0 6px;
  color: var(--shop-border);
}
.current {
  color: var(--shop-foreground);
}

/* ---- 顶部 banner ---- */
.center-banner {
  position: relative;
  overflow: hidden;
  border-radius: var(--shop-radius-xl);
  background: linear-gradient(135deg, var(--shop-primary) 0%, hsl(217, 91%, 62%) 55%, hsl(217, 91%, 72%) 100%);
  color: #fff;
  padding: 30px 28px;
  margin-bottom: 18px;
  box-shadow: 0 8px 24px hsla(217, 91%, 45%, 0.25);
}
.center-banner .deco {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.14);
  pointer-events: none;
}
.center-banner .deco.d1 { width: 220px; height: 220px; right: -60px; top: -70px; }
.center-banner .deco.d2 { width: 120px; height: 120px; right: 120px; bottom: -50px; background: rgba(255, 255, 255, 0.1); }
.center-banner .deco.d3 { width: 60px; height: 60px; left: 38%; top: -24px; background: rgba(255, 255, 255, 0.08); }
.center-banner .bb-title {
  position: relative;
  z-index: 1;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  display: flex;
  align-items: center;
  gap: 10px;
}
.center-banner .bb-sub {
  position: relative;
  z-index: 1;
  font-size: 0.8125rem;
  color: rgba(255, 255, 255, 0.85);
  margin-top: 6px;
}
.center-banner .bb-ticket {
  position: absolute;
  z-index: 1;
  right: 24px;
  bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.18);
  border: 1px dashed rgba(255, 255, 255, 0.5);
  border-radius: 12px;
  padding: 8px 14px;
  font-size: 0.75rem;
  font-weight: 600;
  backdrop-filter: blur(2px);
}

/* ---- 加载/空状态 ---- */
.section-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 48px 0;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
}
.loading-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid var(--shop-primary-bg);
  border-top-color: var(--shop-primary);
  animation: spin 0.8s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 44px 20px;
  text-align: center;
  background: var(--shop-card);
  border: 1px dashed var(--shop-border);
  border-radius: var(--shop-radius-xl);
}
.empty-state .illu {
  width: 120px;
  height: 120px;
  color: hsl(220, 8%, 72%);
}
.empty-state h4 {
  font-size: 0.9375rem;
  font-weight: 600;
}
.empty-state p {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ---- 券列表 ---- */
.coupon-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.coupon-card {
  position: relative;
  display: flex;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-md);
  overflow: hidden;
  border: 1px solid var(--shop-border);
  transition: transform var(--shop-transition), box-shadow var(--shop-transition);
  animation: reveal 0.5s cubic-bezier(0.2, 0.7, 0.3, 1) both;
}
@keyframes reveal {
  from { opacity: 0; transform: translateY(14px); }
  to { opacity: 1; transform: translateY(0); }
}
.coupon-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shop-shadow-lg);
}
.coupon-card.is-soldout {
  filter: saturate(0.5);
  opacity: 0.85;
}

/* ---- 面值块 ---- */
.coupon-face {
  position: relative;
  flex-shrink: 0;
  width: 118px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #fff;
  text-align: center;
}
.coupon-face::after {
  content: '';
  position: absolute;
  right: -7px;
  top: 50%;
  transform: translateY(-50%);
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--shop-background);
  box-shadow: inset 0 0 0 1px var(--shop-border);
}
.face-blue {
  background: linear-gradient(135deg, var(--shop-primary) 0%, var(--shop-primary-light) 100%);
}
.face-red {
  background: linear-gradient(135deg, hsl(0, 84%, 60%) 0%, hsl(24, 90%, 56%) 100%);
}
.face-gray {
  background: linear-gradient(135deg, hsl(220, 10%, 68%), hsl(220, 10%, 80%));
}
.coupon-face .face-symbol {
  font-size: 0.95rem;
  font-weight: 700;
  opacity: 0.95;
}
.coupon-face .face-num {
  font-size: 2.1rem;
  font-weight: 900;
  letter-spacing: -0.02em;
  line-height: 1;
}
.coupon-face .face-cond {
  font-size: 0.66rem;
  font-weight: 500;
  opacity: 0.94;
  background: rgba(255, 255, 255, 0.2);
  padding: 2px 8px;
  border-radius: 999px;
}

/* ---- 信息区 ---- */
.coupon-body {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 16px 16px 20px;
}
.coupon-info {
  flex: 1;
  min-width: 0;
}
.coupon-name {
  font-size: 0.9375rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.coupon-tag {
  font-size: 0.625rem;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 999px;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  border: 1px solid hsl(217, 91%, 84%);
}
.coupon-tag.hot {
  background: hsl(0, 84%, 95%);
  color: hsl(0, 84%, 55%);
  border-color: hsl(0, 84%, 85%);
}
.coupon-desc {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 5px;
}
.coupon-valid {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.72rem;
  color: var(--shop-muted-foreground);
  margin-top: 6px;
}
.coupon-valid svg {
  flex-shrink: 0;
}

/* ---- 按钮区 ---- */
.coupon-action {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}
.claim-btn {
  border: none;
  padding: 8px 20px;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--shop-primary), var(--shop-primary-dark));
  color: #fff;
  box-shadow: 0 3px 10px hsla(217, 91%, 52%, 0.35);
  transition: all var(--shop-transition);
  cursor: pointer;
  font-family: var(--shop-font-sans);
}
.claim-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 5px 14px hsla(217, 91%, 52%, 0.45);
}
.claim-btn:disabled {
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
  box-shadow: none;
  cursor: not-allowed;
  transform: none;
}
.remain-tip {
  font-size: 0.7rem;
  color: var(--shop-muted-foreground);
}

/* ---- 角标 ---- */
.corner-badge {
  position: absolute;
  top: 10px;
  right: -32px;
  transform: rotate(45deg);
  background: var(--shop-destructive);
  color: #fff;
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 3px 36px;
  box-shadow: 0 2px 6px rgb(0 0 0 / 0.18);
  z-index: 2;
}
.corner-badge.gray {
  background: hsl(220, 10%, 62%);
}
.corner-check {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--shop-success);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
  animation: badgePulse 2s ease-in-out infinite;
}
@keyframes badgePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.12); }
}

/* ---- 响应式：移动端券卡纵向布局 ---- */
@media (max-width: 640px) {
  .coupon-card {
    flex-direction: column;
  }
  .coupon-face {
    width: 100%;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    gap: 10px;
  }
  .coupon-face::after {
    display: none;
  }
  .coupon-face .face-num {
    font-size: 1.7rem;
  }
  .coupon-body {
    padding: 14px 16px;
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  .coupon-action {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
  .center-banner .bb-ticket {
    display: none;
  }
}
</style>
