<template>
  <header class="topbar">
    <div class="topbar-left">
      <el-button class="hamburger-btn" @click="store.toggleSidebar" text>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </el-button>
      <h1 class="page-title">{{ route.meta?.title || '管理后台' }}</h1>
    </div>

    <div class="topbar-right">
      <div class="datetime">{{ currentTime }}</div>

      <el-button class="notif-btn" circle @click="handleNotif">
        <el-badge :value="store.notificationCount" :hidden="store.notificationCount === 0">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>
          </svg>
        </el-badge>
      </el-button>

      <el-dropdown trigger="click">
        <span class="admin-avatar-wrap">
          <span class="avatar-circle">{{ store.adminInfo.avatar }}</span>
          <span class="avatar-name">{{ store.adminInfo.name }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <div style="text-align: center;">
                <div style="font-weight:600;">{{ store.adminInfo.name }}</div>
                <div style="font-size:12px;color:var(--shop-muted-foreground);">{{ store.adminInfo.role }}</div>
              </div>
            </el-dropdown-item>
            <el-dropdown-item divided @click="goProfile">个人信息</el-dropdown-item>
            <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup>
/**
 * 顶部栏组件
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { useAppStore } from '../store/app'
import { logout as logoutApi } from '../api/auth'

const route = useRoute()
const router = useRouter()
const store = useAppStore()

const currentTime = ref('')
let timer = null

/**
 * 更新时间显示
 */
function updateTime() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  const weekdays = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']
  currentTime.value = `${year}-${month}-${day} ${weekdays[now.getDay()]} ${hours}:${minutes}:${seconds}`
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  clearInterval(timer)
})

function handleNotif() {
  // TODO: 打开通知面板
}

/**
 * 退出登录
 */
async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', { type: 'warning' })
    try {
      const token = localStorage.getItem('admin_token')
      await logoutApi(token || '')
    } catch {
    }
    store.logout()
    router.push('/login')
  } catch {
  }
}

/** 跳转个人信息页 */
function goProfile() {
  router.push('/profile')
}
</script>

<style scoped>
.topbar {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid var(--shop-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 50;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hamburger-btn {
  font-size: 1.25rem;
  color: var(--shop-muted-foreground);
}

.page-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--shop-foreground);
  white-space: nowrap;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.datetime {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

.notif-btn {
  border: none;
  font-size: 1.25rem;
  color: var(--shop-muted-foreground);
}

.admin-avatar-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px 4px 4px;
  border-radius: 24px;
  background: var(--shop-muted);
  cursor: pointer;
  transition: background 0.2s ease;
}

.admin-avatar-wrap:hover {
  background: var(--shop-border);
}

.avatar-circle {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--shop-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
}

.avatar-name {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

@media (max-width: 768px) {
  .datetime { display: none; }
  .topbar { padding: 0 16px; }
}
</style>
