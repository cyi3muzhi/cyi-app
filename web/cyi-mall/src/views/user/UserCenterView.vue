<template>
  <!-- ============================================================ -->
  <!-- 1. 用户信息头部（渐变色背景）                                    -->
  <!-- ============================================================ -->
  <section class="user-hero">
    <div class="container-shop">
      <div class="hero-card">
        <!-- 装饰圆 -->
        <div class="hero-decoration"></div>

        <div class="hero-content">
          <!-- 头像 -->
          <div class="avatar-wrap">
            <div class="avatar">
              <img
                v-if="user.avatar && !avatarFailed"
                :src="user.avatar"
                alt="用户头像"
                class="avatar-img"
                @error="avatarFailed = true"
              />
              <svg v-else width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
            </div>
          </div>

          <!-- 用户信息 -->
          <div class="user-info">
            <div class="user-name-row">
              <h1 class="user-name">{{ loading ? '加载中…' : user.nickname }}</h1>
              <!-- 会员等级徽章 -->
              <span class="member-badge">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                  <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
                </svg>
                {{ loading ? '加载中…' : memberLevelLabel }}
              </span>
            </div>
            <p class="user-regtime">注册时间：{{ user.createdAt }}</p>

            <!-- 编辑资料按钮 -->
            <router-link to="/user/edit" class="edit-profile-btn">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
              </svg>
              编辑资料
            </router-link>

            <!-- 数据统计 -->
            <div class="stats-row">
              <div class="stat-item">
                <div class="stat-value">&yen;{{ formatAmount(user.totalSpent) }}</div>
                <div class="stat-label">累计消费</div>
              </div>
              <div class="stat-divider"></div>
              <div class="stat-item">
                <div class="stat-value">{{ user.growthValue }}</div>
                <div class="stat-label">成长值</div>
              </div>
              <div class="stat-divider"></div>
              <div class="stat-item clickable" @click="goMyCoupons">
                <div class="stat-value">{{ user.couponCount }}</div>
                <div class="stat-label">
                  优惠券
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="9 18 15 12 9 6"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- 右侧快捷操作（桌面端） -->
          <div class="hero-actions">
            <button class="action-btn" @click="onProfile">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
              个人主页
            </button>
            <button class="action-btn" @click="onSwitchAccount">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/>
                <polyline points="10 17 15 12 10 7"/>
                <line x1="15" y1="12" x2="3" y2="12"/>
              </svg>
              切换账户
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ============================================================ -->
  <!-- 2. 快捷功能入口（收货地址 / 领券中心）                        -->
  <!-- ============================================================ -->
  <section class="entry-section">
    <div class="container-shop entry-grid">
      <!-- 消息中心入口 -->
      <router-link to="/user/messages" class="entry-card message">
        <div class="entry-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
          </svg>
          <span v-if="msgUnread > 0" class="msg-badge">{{ msgUnread > 99 ? '99+' : msgUnread }}</span>
        </div>
        <div class="entry-info">
          <p class="entry-title">消息中心</p>
          <p class="entry-desc">{{ msgUnread > 0 ? `有 ${msgUnread} 条未读消息` : '查看订单与系统通知' }}</p>
        </div>
        <svg class="entry-arrow" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </router-link>

      <router-link to="/user/address" class="entry-card">
        <div class="entry-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
          </svg>
        </div>
        <div class="entry-info">
          <p class="entry-title">收货地址</p>
          <p class="entry-desc">管理常用收货地址</p>
        </div>
        <svg class="entry-arrow" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </router-link>

      <!-- 领券中心入口 -->
      <router-link to="/coupons" class="entry-card coupon">
        <div class="entry-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.83z"/><line x1="7" y1="7" x2="7.01" y2="7"/>
          </svg>
        </div>
        <div class="entry-info">
          <p class="entry-title">领券中心 <span v-if="hasNewCoupon" class="mini-new">NEW</span></p>
          <p class="entry-desc">每周好券上新，先到先得</p>
        </div>
        <svg class="entry-arrow" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </router-link>

      <!-- 退款/售后入口 -->
      <router-link to="/user/refunds" class="entry-card refund">
        <div class="entry-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/>
          </svg>
        </div>
        <div class="entry-info">
          <p class="entry-title">退款/售后</p>
          <p class="entry-desc">查看退款进度与处理结果</p>
        </div>
        <svg class="entry-arrow" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </router-link>
    </div>
  </section>

  <!-- ============================================================ -->
  <!-- 3. 订单管理区                                                  -->
  <!-- ============================================================ -->
  <section class="order-section">
    <div class="container-shop">
      <div class="order-card">
        <!-- 标题 -->
        <div class="order-card-header">
          <h2 class="order-card-title">我的订单</h2>
          <router-link to="/user/orders" class="view-all-link">
            查看全部订单
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
          </router-link>
        </div>

        <!-- 快速操作（图标+数字角标，点击跳转到订单列表页并携带状态筛选） -->
        <div class="quick-actions">
          <div
            v-for="action in quickActions"
            :key="action.key"
            class="quick-action-item"
            @click="onQuickAction(action.key)"
          >
            <div class="quick-action-icon-wrap">
              <div class="quick-action-icon" v-html="action.icon"></div>
              <span
                v-if="action.count > 0"
                class="quick-action-badge"
              >{{ action.count }}</span>
            </div>
            <span class="quick-action-label">{{ action.label }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
/**
 * 个人中心页面
 * 包含用户信息头部、订单快速操作和订单标签页
 * 数据来源：GET /user/v1/center（用户信息 + 订单/优惠券统计）
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getUserCenter } from '@/api/user'
import { getUnreadCount } from '@/api/message'
import { listAvailableCoupons } from '@/api/coupon'
import { useUserStore } from '@/store/user'
import { useCartStore } from '@/store/cart'
import type { User, UserCenter } from '@/types/user'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

/** 个人主页（暂无独立页面，提示开发中） */
function onProfile() {
  alert('个人主页功能开发中，敬请期待')
}

/** 切换账户：退出当前登录后跳转登录页 */
function onSwitchAccount() {
  if (!window.confirm('确定退出当前账户吗？')) return
  userStore.logout()
  // 清空购物车内存数据（未登录时重新读本地购物车）
  cartStore.reset()
  router.push({ path: '/login' })
}

/** 跳转我的优惠券 */
function goMyCoupons() {
  router.push('/user/coupons')
}

// ---- 个人中心数据（真实接口） ----
const center = ref<UserCenter | null>(null)
const loading = ref(true)

/** 头像加载失败标记（回退为默认人像图标） */
const avatarFailed = ref(false)

/** 用户信息（从接口数据组装，totalSpent 由分转为元） */
const user = computed<User>(() => {
  const u = center.value?.user
  return {
    id: u?.id ?? 0,
    nickname: u?.nickname || '未登录',
    avatar: u?.avatar || '',
    phone: u?.phone || '',
    email: u?.email || '',
    memberLevel: u?.level || 'regular',
    level: u?.level || 'regular',
    growthValue: u?.growthValue ?? 0,
    totalSpent: (center.value?.stats?.totalSpent ?? 0) / 100,
    points: 0,
    couponCount: center.value?.stats?.couponCount ?? 0,
    createdAt: formatDate(u?.createdAt)
  }
})

/**
 * 格式化 Unix 时间戳为日期（兼容数字或字符串）
 * @param ts - Unix 秒级时间戳
 * @returns YYYY-MM-DD
 */
function formatDate(ts?: number | string): string {
  if (!ts) return ''
  const num = typeof ts === 'string' ? Number(ts) : ts
  if (!num) return ''
  const d = new Date(num * 1000)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/**
 * 会员等级标签映射（level 取值 regular/gold/diamond）
 */
const memberLevelLabel = computed(() => {
  const levelMap: Record<string, string> = {
    regular: '普通会员',
    bronze: '青铜会员',
    silver: '白银会员',
    gold: '黄金会员',
    platinum: '铂金会员',
    diamond: '钻石会员'
  }
  return levelMap[user.value.level || user.value.memberLevel] || '普通会员'
})

/**
 * 格式化金额
 * @param amount - 金额数值（元）
 * @returns 格式化后的金额字符串
 */
function formatAmount(amount: number): string {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// ---- 快速操作（角标数字来自接口统计） ----
interface QuickAction {
  key: string
  label: string
  icon: string
  count: number
}

const quickActions = computed<QuickAction[]>(() => {
  const s = center.value?.stats
  return [
    {
      key: 'pending_payment',
      label: '待付款',
      count: s?.pendingPay ?? 0,
      icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>'
    },
    {
      key: 'pending_ship',
      label: '待发货',
      count: s?.pendingShip ?? 0,
      icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10V6a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v4"/><path d="M3 10v8a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-8"/><polyline points="16 14 12 18 8 14"/></svg>'
    },
    {
      key: 'pending_receive',
      label: '待收货',
      count: s?.pendingReceive ?? 0,
      icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>'
    },
    {
      key: 'pending_review',
      label: '待评价',
      count: s?.pendingReview ?? 0,
      icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><line x1="9" y1="10" x2="15" y2="10"/></svg>'
    },
    {
      key: 'refund',
      label: '退款/售后',
      count: s?.refundCount ?? 0,
      icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>'
    }
  ]
})

/**
 * 处理快速操作点击：跳转到订单列表页并带上状态参数
 * @param key - 操作标识
 */
function onQuickAction(key: string) {
  router.push({ path: '/user/orders', query: { status: key } })
}

// 页面加载时获取个人中心概览数据
async function loadCenter() {
  loading.value = true
  try {
    center.value = await getUserCenter()
  } catch {
    center.value = null
  } finally {
    loading.value = false
  }
}

/** 消息中心未读数 */
const msgUnread = ref(0)

/** 领券中心是否有未领取的新券（控制 NEW 标记显示） */
const hasNewCoupon = ref(false)

/** 获取消息未读数（失败静默，不影响页面） */
async function loadMsgUnread() {
  try {
    const res = await getUnreadCount()
    msgUnread.value = res.count || 0
  } catch {
    msgUnread.value = 0
  }
}

/** 判断是否有未领取的新券：领券中心列表中只要存在 claimed=false 的券即显示 NEW */
async function loadNewCoupon() {
  try {
    const res = await listAvailableCoupons({ page: 1, pageSize: 100 })
    hasNewCoupon.value = (res.items || []).some(i => !i.claimed)
  } catch {
    hasNewCoupon.value = false
  }
}

onMounted(() => {
  loadCenter()
  loadMsgUnread()
  loadNewCoupon()
})
</script>

<style scoped>
/* ============================================
   个人中心页面样式
   ============================================ */

/* ---- 用户信息头部 ---- */
.user-hero {
  padding: 24px 0 0;
}

.hero-card {
  position: relative;
  overflow: hidden;
  border-radius: var(--shop-radius-xl);
  border: 1px solid var(--shop-border);
  background: linear-gradient(135deg, var(--shop-primary) 0%, hsl(217, 91%, 65%) 50%, hsl(217, 91%, 75%) 100%);
}

.hero-decoration {
  pointer-events: none;
  position: absolute;
  right: -64px;
  top: -64px;
  width: 256px;
  height: 256px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
}

.hero-decoration::after {
  content: '';
  position: absolute;
  bottom: -80px;
  left: -160px;
  width: 320px;
  height: 320px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}

.hero-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  padding: 32px;
}

@media (min-width: 768px) {
  .hero-content {
    flex-direction: row;
    align-items: flex-start;
    gap: 32px;
    padding: 40px;
  }
}

/* ---- 头像 ---- */
.avatar-wrap {
  flex-shrink: 0;
}

.avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 96px;
  height: 96px;
  border-radius: 50%;
  border: 4px solid rgba(255, 255, 255, 0.7);
  background: linear-gradient(135deg, #fff, hsl(217, 91%, 90%));
  box-shadow: var(--shop-shadow-xl);
  overflow: hidden;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (min-width: 768px) {
  .avatar {
    width: 112px;
    height: 112px;
  }
}

/* ---- 用户信息 ---- */
.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}

@media (min-width: 768px) {
  .user-info {
    align-items: flex-start;
    text-align: left;
  }
}

.user-name-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

@media (min-width: 640px) {
  .user-name-row {
    flex-direction: row;
    flex-wrap: wrap;
  }
}

.user-name {
  font-size: 1.5rem;
  font-weight: 700;
  color: #fff;
}

@media (min-width: 768px) {
  .user-name {
    font-size: 1.75rem;
  }
}

.member-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 12px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: #fff;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
}

.user-regtime {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.8);
}

.edit-profile-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 20px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  background: transparent;
  font-size: 0.875rem;
  font-weight: 500;
  color: #fff;
  cursor: pointer;
  transition: background 0.2s;
}

.edit-profile-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* ---- 数据统计 ---- */
.stats-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 16px;
  margin-top: 8px;
}

@media (min-width: 768px) {
  .stats-row {
    justify-content: flex-start;
  }
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #fff;
}

.stat-label {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.7);
  margin-top: 2px;
}

.stat-divider {
  width: 1px;
  background: rgba(255, 255, 255, 0.2);
}

/* 可点击统计项（优惠券）：hover 高亮反馈 */
.stat-item.clickable {
  cursor: pointer;
  padding: 4px 10px;
  border-radius: var(--shop-radius-md);
  transition: background 0.2s;
}
.stat-item.clickable:hover {
  background: rgba(255, 255, 255, 0.16);
}
.stat-item.clickable .stat-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 3px;
  color: rgba(255, 255, 255, 0.95);
  font-weight: 600;
}

/* ---- 右侧快捷操作 ---- */
.hero-actions {
  display: none;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
  align-self: flex-start;
}

@media (min-width: 768px) {
  .hero-actions {
    display: flex;
  }
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--shop-radius-lg);
  border: none;
  background: rgba(255, 255, 255, 0.2);
  font-size: 0.875rem;
  font-weight: 500;
  color: #fff;
  cursor: pointer;
  backdrop-filter: blur(4px);
  transition: background 0.2s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ---- 快捷功能入口 ---- */
.entry-section {
  padding: 24px 0 0;
}

/* 入口卡片网格：桌面两列，移动端单列 */
.entry-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.entry-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  text-decoration: none;
  transition: box-shadow 0.2s, border-color 0.2s;
}

.entry-card:hover {
  border-color: var(--shop-primary);
  box-shadow: var(--shop-shadow-lg);
}

.entry-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  flex-shrink: 0;
}

/* 领券中心卡片：红橙渐变图标 */
.entry-card.coupon .entry-icon {
  background: linear-gradient(135deg, hsl(0, 84%, 60%), hsl(24, 90%, 56%));
  color: #fff;
  box-shadow: 0 3px 8px hsla(0, 84%, 55%, 0.35);
}

/* 退款/售后卡片：蓝紫渐变图标 */
.entry-card.refund .entry-icon {
  background: linear-gradient(135deg, hsl(240, 84%, 62%), hsl(270, 84%, 60%));
  color: #fff;
  box-shadow: 0 3px 8px hsla(255, 84%, 55%, 0.35);
}

/* 消息中心卡片：青绿渐变图标 */
.entry-card.message .entry-icon {
  background: linear-gradient(135deg, hsl(174, 84%, 42%), hsl(160, 84%, 45%));
  color: #fff;
  box-shadow: 0 3px 8px hsla(160, 84%, 40%, 0.35);
}

/* 消息未读角标 */
.msg-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--shop-destructive);
  color: #fff;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
  box-shadow: 0 2px 6px hsla(0, 84%, 50%, 0.4);
}

.entry-card.message .entry-icon {
  position: relative;
}

.entry-info {
  flex: 1;
  min-width: 0;
}

.entry-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
  display: flex;
  align-items: center;
  gap: 6px;
}

.entry-desc {
  margin-top: 2px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.entry-arrow {
  color: var(--shop-muted-foreground);
  flex-shrink: 0;
}

/* NEW 角标 */
.mini-new {
  font-size: 0.5625rem;
  font-weight: 700;
  color: #fff;
  background: hsl(0, 84%, 60%);
  border-radius: 4px;
  padding: 1px 5px;
  animation: badgePulse 2s ease-in-out infinite;
}

/* 移动端：入口卡片单列 */
@media (max-width: 640px) {
  .entry-grid {
    grid-template-columns: 1fr;
  }
}

/* ---- 订单管理区 ---- */
.order-section {
  padding: 32px 0;
}

.order-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  overflow: hidden;
}

.order-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--shop-border);
}

.order-card-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.view-all-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-primary);
  text-decoration: none;
  transition: opacity 0.2s;
}

.view-all-link:hover {
  opacity: 0.8;
}

/* ---- 快速操作 ---- */
.quick-actions {
  display: flex;
  gap: 8px;
  padding: 20px 24px 28px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.quick-actions::-webkit-scrollbar {
  display: none;
}

@media (min-width: 768px) {
  .quick-actions {
    justify-content: space-around;
    gap: 24px;
  }
}

.quick-action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  transition: transform 0.2s;
  flex-shrink: 0;
}

.quick-action-item:hover {
  transform: scale(1.05);
}

.quick-action-icon-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--shop-primary-bg);
}

.quick-action-badge {
  position: absolute;
  right: -4px;
  top: -4px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  background: var(--shop-destructive);
  color: #fff;
  font-size: 0.625rem;
  font-weight: 700;
  animation: badgePulse 2s ease-in-out infinite;
}

@keyframes badgePulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.quick-action-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--shop-foreground);
  white-space: nowrap;
}
</style>
