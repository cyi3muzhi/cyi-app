<template>
  <div class="admin-layout">
    <!-- 侧边栏遮罩（移动端） -->
    <div
      class="sidebar-overlay"
      :class="{ open: !store.sidebarCollapsed && windowWidth <= 768 }"
      @click="store.toggleSidebar()"
    ></div>

    <!-- 侧边栏 -->
    <AppSidebar />

    <!-- 主区域 -->
    <div class="main-wrapper" :style="{ marginLeft: store.sidebarWidth }">
      <AppTopbar />

      <main class="content-area">
        <router-view />
      </main>

      <footer class="admin-footer">
        &copy; 2026 悦选商城 - All Rights Reserved. 管理后台 v2.0
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 管理后台布局组件
 * 包含侧边栏 + 顶部栏 + 内容区
 */
import { ref, onMounted, onUnmounted } from 'vue'
import AppSidebar from '../components/AppSidebar.vue'
import AppTopbar from '../components/AppTopbar.vue'
import { useAppStore } from '../store/app'
import { getMyPermissions } from '../api/auth'

const store = useAppStore()
const windowWidth = ref(window.innerWidth)

function handleResize() {
  windowWidth.value = window.innerWidth
  if (window.innerWidth <= 768) {
    store.sidebarCollapsed = true
  }
}

// 页面加载/刷新时加载权限
async function loadPermissions() {
  if (!store.adminInfo?.role) return
  try {
    const permRes = await getMyPermissions()
    const perms: Record<string, Record<string, boolean>> = {}
    for (const modKey in permRes.permissions) {
      perms[modKey] = permRes.permissions[modKey]?.values || {}
    }
    store.setPermissions(perms)
  } catch {
    // 权限加载失败不影响页面展示
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  handleResize()
  loadPermissions()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.admin-layout {
  min-height: 100vh;
}

.main-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  transition: margin-left 0.3s ease;
}

.content-area {
  flex: 1;
  padding: 24px;
}

.admin-footer {
  text-align: center;
  padding: 16px 24px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  border-top: 1px solid var(--shop-border);
  background: #fff;
  margin-top: auto;
}

.sidebar-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 90;
}

.sidebar-overlay.open {
  display: block;
}

@media (max-width: 768px) {
  .content-area { padding: 16px; }
}
</style>
