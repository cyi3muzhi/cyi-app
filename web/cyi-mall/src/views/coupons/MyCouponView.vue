<template>
  <div class="my-coupon">
    <!-- 面包屑 -->
    <div class="container-shop py-3">
      <div class="my-breadcrumb">
        <router-link to="/" class="breadcrumb-link">首页</router-link>
        <span class="separator">/</span>
        <router-link to="/user" class="breadcrumb-link">个人中心</router-link>
        <span class="separator">/</span>
        <span class="current">我的优惠券</span>
      </div>
    </div>

    <div class="container-shop pb-12">
      <!-- 三 Tab（下划线高亮 + 数量角标） -->
      <div class="coupon-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="coupon-tab"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
          <span class="tab-count">{{ tabCount(tab.key) }}</span>
        </button>
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="section-state">
        <span class="loading-dot"></span>正在加载优惠券...
      </div>

      <!-- 空状态 -->
      <div v-else-if="currentList.length === 0" class="empty-state">
        <svg class="illu" viewBox="0 0 120 120" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="28" y="26" width="64" height="64" rx="12"/>
          <circle cx="60" cy="58" r="16"/>
          <path d="M28 40h64M40 26v14M80 26v14"/>
          <path d="M50 76l7 7 14-14"/>
        </svg>
        <h4>{{ emptyTitle }}</h4>
        <p>{{ emptyDesc }}</p>
        <button v-if="activeTab !== 'used'" class="go-claim" @click="goClaim">去领券</button>
      </div>

      <!-- 券列表 -->
      <div v-else class="coupon-list">
        <article
          v-for="(c, idx) in currentList"
          :key="c.recordId"
          class="coupon-card"
          :class="{ 'is-used': c.recordStatus === 1, 'is-expired': c.recordStatus === 2 }"
          :style="{ animationDelay: (idx * 0.05) + 's' }"
        >
          <!-- 已使用：水印；已过期：角标 -->
          <span v-if="c.recordStatus === 1" class="watermark">已使用</span>
          <span v-else-if="c.recordStatus === 2" class="corner-badge">已过期</span>

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
              <h3 class="coupon-name">{{ c.name }}</h3>
              <!-- 未使用：有效期至 -->
              <p v-if="c.recordStatus === 0" class="coupon-valid">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                </svg>
                有效期至 {{ c.validEnd || '永久有效' }}
              </p>
              <!-- 已使用：核销订单 -->
              <p v-else-if="c.recordStatus === 1" class="coupon-valid">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                {{ c.orderId > 0 ? `订单 #${c.orderId} 使用` : '已使用' }}
              </p>
              <!-- 已过期：过期时间 -->
              <p v-else class="coupon-valid">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                </svg>
                已于 {{ c.validEnd || '—' }} 过期
              </p>
            </div>
            <div class="coupon-action">
              <!-- 未使用：去使用 -->
              <button v-if="c.recordStatus === 0" class="use-btn" @click="goUse">去使用</button>
              <!-- 已使用：使用时间 -->
              <div v-else-if="c.recordStatus === 1" class="coupon-extra">
                使用时间<br />
                <span class="used-time">{{ c.usedTime ? formatDateTime(c.usedTime) : '—' }}</span>
              </div>
              <!-- 已过期 -->
              <div v-else class="coupon-extra">已过期无法使用</div>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 我的优惠券页面
 * 三个 Tab（未使用/已使用/已过期）+ 券卡状态展示 + 空状态
 * 数据来源：GET /api/v1/coupons/my?status=&page=&page_size=
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listMyCoupons } from '@/api/coupon'
import type { MyCouponItem } from '@/types/coupon'

const router = useRouter()

/** Tab 定义：status 1=未使用 2=已使用 3=已过期 */
interface CouponTab {
  key: string
  label: string
  status: number
}
const tabs: CouponTab[] = [
  { key: 'unused', label: '未使用', status: 1 },
  { key: 'used', label: '已使用', status: 2 },
  { key: 'expired', label: '已过期', status: 3 }
]

const activeTab = ref('unused')
const loading = ref(true)

/** 各 Tab 的列表缓存 */
const listMap = ref<Record<string, MyCouponItem[]>>({})
/** 各 Tab 的总数 */
const totalMap = ref<Record<string, number>>({})

/** 当前 Tab 的列表 */
const currentList = computed(() => listMap.value[activeTab.value] || [])

/** Tab 角标数量（未加载时为 0 则不显示） */
function tabCount(key: string): number {
  return totalMap.value[key] ?? 0
}

/** 空状态文案 */
const emptyTitle = computed(() => {
  switch (activeTab.value) {
    case 'used': return '暂无已使用的优惠券'
    case 'expired': return '暂无已过期的优惠券'
    default: return '暂无优惠券'
  }
})
const emptyDesc = computed(() => {
  switch (activeTab.value) {
    case 'used': return '使用过的券会出现在这里'
    case 'expired': return '过期的券无法再使用'
    default: return '去领券中心逛逛，好券等你拿'
  }
})

/**
 * 加载指定状态（status 1/2/3）的优惠券列表
 * @param key Tab 键
 */
async function loadTab(key: string) {
  const tab = tabs.find(t => t.key === key)!
  loading.value = true
  try {
    const res = await listMyCoupons({ status: tab.status, page: 1, pageSize: 50 })
    listMap.value[key] = res.items || []
    totalMap.value[key] = res.total ?? 0
  } catch {
    listMap.value[key] = []
    totalMap.value[key] = 0
  } finally {
    loading.value = false
  }
}

/** 切换 Tab */
function switchTab(key: string) {
  activeTab.value = key
  // 已加载过则直接展示缓存，否则请求
  if (!(key in listMap.value)) {
    loadTab(key)
  }
}

/** 去使用：跳转商品列表 */
function goUse() {
  router.push('/products')
}

/** 去领券：跳转领券中心 */
function goClaim() {
  router.push('/coupons')
}

/** 面值块配色：与管理端统一（立减红、折扣橙、满减蓝），过期灰色 */
function faceClass(c: MyCouponItem): string {
  if (c.recordStatus === 2) return 'face-gray'
  if (c.type === 2) return 'face-orange'
  if (c.type === 3) return 'face-blue'
  return 'face-red'
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
function discountText(c: MyCouponItem): string {
  const num = c.value / 10
  const text = Number.isInteger(num) ? String(num) : num.toFixed(1)
  return `${text}折`
}

/** 使用门槛：满 x 可用 / 全场通用 */
function thresholdText(c: MyCouponItem): string {
  return c.minAmount > 0 ? `满${formatMoney(c.minAmount)}可用` : '全场通用'
}

/**
 * Unix 秒时间戳格式化为日期时间
 * @param ts 时间戳（秒）
 */
function formatDateTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${y}-${m}-${day} ${h}:${min}`
}

// 页面加载：并行加载三个 Tab，保证角标计数齐全
onMounted(() => {
  loadTab('unused')
  loadTab('used')
  loadTab('expired')
})
</script>

<style scoped>
/* ---- 面包屑 ---- */
.my-breadcrumb {
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

/* ---- Tab ---- */
.coupon-tabs {
  display: flex;
  border-bottom: 1px solid var(--shop-border);
  margin-bottom: 18px;
}
.coupon-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  transition: color 0.2s, border-color 0.2s;
  cursor: pointer;
  font-family: var(--shop-font-sans);
}
.coupon-tab:hover {
  color: var(--shop-foreground);
}
.coupon-tab.active {
  color: var(--shop-primary);
  border-bottom-color: var(--shop-primary);
  font-weight: 600;
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
.coupon-tab.active .tab-count {
  background: var(--shop-primary);
  color: #fff;
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
.go-claim {
  padding: 9px 26px;
  border: none;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--shop-primary), var(--shop-primary-dark));
  color: #fff;
  font-size: 0.8125rem;
  font-weight: 600;
  box-shadow: 0 3px 10px hsla(217, 91%, 52%, 0.35);
  transition: all var(--shop-transition);
  cursor: pointer;
  font-family: var(--shop-font-sans);
}
.go-claim:hover {
  transform: translateY(-1px);
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
.coupon-card.is-used,
.coupon-card.is-expired {
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
/* 面值块类型配色：与管理端一致（立减红、折扣橙、满减蓝） */
.face-blue {
  background: linear-gradient(135deg, #667eea 0%, #409eff 100%);
}
.face-red {
  background: linear-gradient(135deg, #f56c6c 0%, #e74c3c 100%);
}
.face-orange {
  background: linear-gradient(135deg, #e6a23c 0%, #d4890a 100%);
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

/* ---- 按钮/附加区 ---- */
.coupon-action {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}
.use-btn {
  border: none;
  padding: 9px 22px;
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
.use-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 5px 14px hsla(217, 91%, 52%, 0.45);
}
.coupon-extra {
  font-size: 0.7rem;
  color: var(--shop-muted-foreground);
  text-align: right;
}
.coupon-extra .used-time {
  color: hsl(0, 84%, 50%);
  font-weight: 600;
}

/* ---- 角标/水印 ---- */
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
.watermark {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.8rem;
  font-weight: 900;
  letter-spacing: 0.25em;
  color: rgb(0 0 0 / 0.07);
  pointer-events: none;
  transform: rotate(-12deg);
  z-index: 1;
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
  .coupon-action .coupon-extra {
    text-align: left;
  }
  .coupon-tab {
    padding: 10px 12px;
  }
}
</style>
