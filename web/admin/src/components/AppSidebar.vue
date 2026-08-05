<template>
  <aside class="sidebar" :style="{ width: store.sidebarWidth }">
    <!-- Logo区域 -->
    <div class="sidebar-logo">
      <router-link to="/dashboard" class="logo-link">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
        <span v-show="!store.sidebarCollapsed" class="logo-text">{{ store.shopName }}</span>
      </router-link>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav sidebar-scrollbar">
      <template v-for="group in menuGroups" :key="group.label">
        <template v-if="group.items.length > 0">
          <div v-show="!store.sidebarCollapsed" class="nav-label">{{ group.label }}</div>
          <router-link
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
          >
            <span class="nav-icon" v-html="item.icon"></span>
            <span v-show="!store.sidebarCollapsed" class="nav-text">{{ item.label }}</span>
            <span v-show="!store.sidebarCollapsed && item.badge" class="nav-badge">{{ item.badge }}</span>
          </router-link>
        </template>
      </template>
    </nav>

    <!-- 底部 - 返回前台 -->
    <div class="sidebar-footer">
      <router-link to="/dashboard" class="nav-item back-link">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
        </svg>
        <span v-show="!store.sidebarCollapsed">返回前台</span>
      </router-link>
    </div>
  </aside>
</template>

<script setup lang="ts">
/**
 * 侧边栏导航组件 - 支持分组菜单
 */
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '../store/app'

const route = useRoute()
const store = useAppStore()

/** 侧边栏从 store 读取已上架商品数量 */
const productBadge = computed(() => store.productOnlineCount)

/** 侧边栏从 store 读取待处理订单数量 */
const orderBadge = computed(() => store.orderPendingCount)

/** 侧边栏从 store 读取低库存预警数量 */
const stockAlertBadge = computed(() => store.stockAlertCount)

function onProductChanged() { store.fetchProductOnlineCount(); store.fetchStockAlertCount() }
window.addEventListener('product-changed', onProductChanged)
onUnmounted(() => window.removeEventListener('product-changed', onProductChanged))

function onOrderChanged() { store.fetchOrderPendingCount() }
window.addEventListener('order-changed', onOrderChanged)
onUnmounted(() => window.removeEventListener('order-changed', onOrderChanged))

onMounted(() => {
  store.fetchProductOnlineCount()
  store.fetchOrderPendingCount()
  store.fetchStockAlertCount()
  store.fetchShopName()
})

/** 菜单分组定义 */
interface MenuItem {
  path: string
  label: string
  icon: string
  badge?: string | number
  perm?: string
  /** 匹配时检查子路径 */
  matchPrefix?: boolean
}

interface MenuGroup {
  label: string
  items: MenuItem[]
}

const menuGroups = computed<MenuGroup[]>(() => [
  {
    label: '概览',
    items: [
      {
        path: '/dashboard',
        label: '仪表盘',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>',
        perm: ''
      }
    ]
  },
  {
    label: '商品',
    items: filterByPerm([
      {
        path: '/products',
        label: '商品管理',
        matchPrefix: true,
        badge: productBadge.value || undefined,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
        perm: 'products.view'
      },
      {
        path: '/categories',
        label: '商品分类',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>',
        perm: 'categories.view'
      },
      {
        path: '/stock-alert',
        label: '库存预警',
        badge: stockAlertBadge.value || undefined,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
        perm: 'products.view'
      }
    ])
  },
  {
    label: '订单',
    items: filterByPerm([
      {
        path: '/orders',
        label: '订单管理',
        matchPrefix: true,
        badge: orderBadge.value || undefined,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
        perm: 'orders.view'
      },
      {
        path: '/refunds',
        label: '售后管理',
        matchPrefix: true,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>',
        perm: 'orders.view'
      }
    ])
  },
  {
    label: '用户',
    items: filterByPerm([
      {
        path: '/users',
        label: '用户管理',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>',
        perm: 'users.view'
      },
      {
        path: '/members',
        label: '会员管理',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>',
        perm: 'users.view'
      },
      {
        path: '/roles',
        label: '角色管理',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M9 12l2 2 4-4"/></svg>',
        perm: 'roles.view'
      },
      {
        path: '/staff',
        label: '员工管理',
        matchPrefix: true,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/><polyline points="16 8 18 10 22 6"/></svg>',
        perm: 'staff.view'
      }
    ])
  },
  {
    label: '营销',
    items: filterByPerm([
      {
        path: '/banners',
        label: 'Banner 管理',
        matchPrefix: true,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="14" rx="2" ry="2"/><line x1="21" y1="9" x2="3" y2="9"/><line x1="21" y1="15" x2="3" y2="15"/><line x1="12" y1="3" x2="12" y2="21"/></svg>',
        perm: 'banners.view'
      },
      {
        path: '/coupons',
        label: '优惠券管理',
        matchPrefix: true,
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>',
        perm: 'coupons.view'
      }
    ])
  },
  {
    label: '系统',
    items: filterByPerm([
      {
        path: '/logs',
        label: '操作日志',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>',
        perm: 'settings.logs'
      },
      {
        path: '/knowledge',
        label: '知识库',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/><path d="M8 7h6"/><path d="M8 11h8"/></svg>',
        perm: 'agent.knowledge'
      },
      {
        path: '/settings',
        label: '系统设置',
        icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>',
        perm: 'settings.basic'
      }
    ])
  }
])

/** 按权限过滤菜单项 */
function filterByPerm(items: MenuItem[]): MenuItem[] {
  return items.filter(item => {
    if (!item.perm) return true
    const [mod, act] = item.perm.split('.')
    return store.hasPerm(mod, act)
  })
}

/** 判断菜单是否活跃 */
function isActive(path: string): boolean {
  // 带 matchPrefix 的项按前缀匹配
  for (const group of menuGroups.value) {
    for (const item of group.items) {
      if (item.path === path && item.matchPrefix) {
        return route.path.startsWith(path)
      }
    }
  }
  return route.path === path
}
</script>

<style scoped>
/* ---- 动效 ---- */
@keyframes sidebarFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  background: #1e293b;
  color: #cbd5e1;
  z-index: 100;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
}

.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.logo-link {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.logo-text {
  color: #fff;
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  white-space: nowrap;
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-label {
  font-size: 0.625rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #64748b;
  padding: 0.5rem 0.75rem 0.5rem;
  font-weight: 600;
  margin-top: 12px;
}

.nav-label:first-of-type {
  margin-top: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-radius: 8px;
  color: #94a3b8;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  margin-bottom: 2px;
  white-space: nowrap;
  position: relative;
}

/* 激活指示条 - 激活时从左侧滑入 */
.nav-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  width: 3px;
  height: 0;
  border-radius: 0 2px 2px 0;
  background: var(--shop-primary);
  transition: height 0.25s cubic-bezier(0.34, 1.56, 0.64, 1),
              top 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.nav-item.active::before {
  top: 6px;
  height: calc(100% - 12px);
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #e2e8f0;
  transform: translateX(2px);
}

.nav-item.active {
  background: var(--shop-primary);
  color: #fff;
  box-shadow: 0 4px 12px hsla(217, 91%, 52%, 0.3);
  transform: translateX(0);
}

.nav-item .nav-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nav-badge {
  margin-left: auto;
  background: rgba(255, 255, 255, 0.12);
  color: #cbd5e1;
  font-size: 0.6875rem;
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  font-weight: 600;
}

.nav-item.active .nav-badge {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.sidebar-footer {
  padding: 16px 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.back-link {
  font-size: 0.8125rem;
  color: #64748b;
}
</style>
