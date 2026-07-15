<template>
  <aside class="sidebar" :style="{ width: store.sidebarWidth }">
    <!-- Logo区域 -->
    <div class="sidebar-logo">
      <router-link to="/dashboard" class="logo-link">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
        <span v-show="!store.sidebarCollapsed" class="logo-text">悦选商城</span>
      </router-link>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav sidebar-scrollbar">
      <div v-show="!store.sidebarCollapsed" class="nav-label">导航菜单</div>
      <router-link
        v-for="item in menuItems"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <span class="nav-icon" v-html="item.icon"></span>
        <span v-show="!store.sidebarCollapsed" class="nav-text">{{ item.label }}</span>
        <span v-show="!store.sidebarCollapsed && item.badge" class="nav-badge">{{ item.badge }}</span>
      </router-link>
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
 * 侧边栏导航组件
 */
import { useRoute } from 'vue-router'
import { useAppStore } from '../store/app'

const route = useRoute()
const store = useAppStore()

const menuItems = [
  {
    path: '/dashboard',
    label: '仪表盘',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>'
  },
  {
    path: '/products',
    label: '商品管理',
    badge: '1,286',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>'
  },
  {
    path: '/categories',
    label: '商品分类',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>'
  },
  {
    path: '/orders',
    label: '订单管理',
    badge: '48',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>'
  },
  {
    path: '/users',
    label: '用户管理',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>'
  },
  {
    path: '/roles',
    label: '角色管理',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M9 12l2 2 4-4"/></svg>'
  },
  {
    path: '/staff',
    label: '员工管理',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/><polyline points="16 8 18 10 22 6"/></svg>'
  },
  {
    path: '/coupons',
    label: '优惠券管理',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>'
  },
  {
    path: '/logs',
    label: '操作日志',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>'
  },
  {
    path: '/settings',
    label: '系统设置',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>'
  }
]

/**
 * 判断菜单项是否活跃
 */
function isActive(path) {
  if (['/products', '/categories', '/roles', '/staff'].includes(path)) {
    return route.path.startsWith(path)
  }
  return route.path === path
}
</script>

<style scoped>
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
  transition: all 0.2s ease;
  cursor: pointer;
  margin-bottom: 2px;
  white-space: nowrap;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #e2e8f0;
}

.nav-item.active {
  background: var(--shop-primary);
  color: #fff;
  box-shadow: 0 4px 12px hsla(217, 91%, 52%, 0.3);
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
