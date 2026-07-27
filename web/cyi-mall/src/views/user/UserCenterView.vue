<template>
  <!-- ============================================================ -->
  <!-- 1. 面包屑导航                                                  -->
  <!-- ============================================================ -->
  <div class="breadcrumb">
    <div class="container-shop">
      <router-link to="/" class="breadcrumb-link">首页</router-link>
      <svg class="breadcrumb-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="9 18 15 12 9 6"/>
      </svg>
      <span class="breadcrumb-current">个人中心</span>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 2. 用户信息头部（渐变色背景）                                    -->
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
              <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
            </div>
          </div>

          <!-- 用户信息 -->
          <div class="user-info">
            <div class="user-name-row">
              <h1 class="user-name">{{ user.nickname }}</h1>
              <!-- 会员等级徽章 -->
              <span class="member-badge">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                  <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
                </svg>
                {{ memberLevelLabel }}
              </span>
            </div>
            <p class="user-regtime">注册时间：{{ user.createdAt }}</p>

            <!-- 编辑资料按钮 -->
            <button class="edit-profile-btn">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
              </svg>
              编辑资料
            </button>

            <!-- 数据统计 -->
            <div class="stats-row">
              <div class="stat-item">
                <div class="stat-value">&yen;{{ formatAmount(user.totalSpent) }}</div>
                <div class="stat-label">累计消费</div>
              </div>
              <div class="stat-divider"></div>
              <div class="stat-item">
                <div class="stat-value">{{ user.points }}</div>
                <div class="stat-label">积分</div>
              </div>
              <div class="stat-divider"></div>
              <div class="stat-item">
                <div class="stat-value">{{ user.couponCount }}</div>
                <div class="stat-label">优惠券</div>
              </div>
            </div>
          </div>

          <!-- 右侧快捷操作（桌面端） -->
          <div class="hero-actions">
            <button class="action-btn">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
              个人主页
            </button>
            <button class="action-btn">
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

        <!-- 快速操作（图标+数字角标） -->
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

        <!-- 订单标签页 -->
        <div class="order-tabs">
          <button
            v-for="tab in orderTabs"
            :key="tab.key"
            class="order-tab"
            :class="{ 'tab-active': activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- 订单列表区域（后续扩展时在此处插入订单卡片列表） -->
        <div class="order-list-area">
          <div v-if="filteredOrders.length === 0" class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>
            </svg>
            <p>暂无相关订单</p>
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
 */
import { ref, computed } from 'vue'
import type { User } from '@/types/user'
import type { OrderStatus } from '@/types/order'

// ---- Mock 用户数据 ----
const user: User = {
  id: 1,
  nickname: '张小明',
  avatar: '',
  phone: '138****1234',
  email: 'zhangxm@example.com',
  memberLevel: 'gold',
  totalSpent: 12680,
  points: 2360,
  couponCount: 5,
  createdAt: '2025-01-15'
}

/**
 * 会员等级标签映射
 */
const memberLevelLabel = computed(() => {
  const levelMap: Record<string, string> = {
    bronze: '青铜会员',
    silver: '白银会员',
    gold: '黄金会员',
    platinum: '铂金会员',
    diamond: '钻石会员'
  }
  return levelMap[user.memberLevel] || '普通会员'
})

/**
 * 格式化金额
 * @param amount - 金额数值
 * @returns 格式化后的金额字符串
 */
function formatAmount(amount: number): string {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

// ---- 快速操作 ----
interface QuickAction {
  key: string
  label: string
  icon: string
  count: number
}

const quickActions: QuickAction[] = [
  {
    key: 'pending_payment',
    label: '待付款',
    count: 2,
    icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>'
  },
  {
    key: 'pending_ship',
    label: '待发货',
    count: 1,
    icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10V6a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v4"/><path d="M3 10v8a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-8"/><polyline points="16 14 12 18 8 14"/></svg>'
  },
  {
    key: 'pending_receive',
    label: '待收货',
    count: 0,
    icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>'
  },
  {
    key: 'pending_review',
    label: '待评价',
    count: 3,
    icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><line x1="9" y1="10" x2="15" y2="10"/></svg>'
  },
  {
    key: 'refund',
    label: '退款/售后',
    count: 0,
    icon: '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>'
  }
]

/**
 * 处理快速操作点击
 * @param key - 操作标识
 */
function onQuickAction(key: string) {
  // 点击快速操作时切换到对应标签页
  activeTab.value = key
}

// ---- 订单标签页 ----
interface OrderTab {
  key: string
  label: string
}

const orderTabs: OrderTab[] = [
  { key: 'all', label: '全部' },
  { key: 'pending_payment', label: '待付款' },
  { key: 'pending_ship', label: '待发货' },
  { key: 'pending_receive', label: '待收货' },
  { key: 'pending_review', label: '待评价' }
]

const activeTab = ref('all')

/**
 * 根据当前标签页过滤订单（占位，后续扩展）
 */
const filteredOrders = computed(() => {
  // 当前返回空数组占位，后续从 store 获取订单数据
  return [] as any[]
})
</script>

<style scoped>
/* ============================================
   个人中心页面样式
   ============================================ */

/* ---- 面包屑 ---- */
.breadcrumb {
  padding: 24px 0 0;
}

.breadcrumb-link {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
}

.breadcrumb-link:hover {
  color: var(--shop-primary);
  text-decoration: underline;
}

.breadcrumb-arrow {
  display: inline;
  margin: 0 8px;
  vertical-align: middle;
  color: var(--shop-muted-foreground);
}

.breadcrumb-current {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

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
  padding: 20px 24px;
  border-bottom: 1px solid var(--shop-border);
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

/* ---- 订单标签页 ---- */
.order-tabs {
  display: flex;
  gap: 0;
  padding: 0 24px;
  border-bottom: 1px solid var(--shop-border);
}

.order-tab {
  padding: 12px 16px;
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

.order-tab:hover {
  color: var(--shop-foreground);
}

.tab-active {
  color: var(--shop-primary);
  border-bottom-color: var(--shop-primary);
}

/* ---- 订单列表区域 ---- */
.order-list-area {
  min-height: 120px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
}

.empty-state p {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
</style>
