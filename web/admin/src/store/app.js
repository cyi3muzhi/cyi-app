/**
 * 悦选商城管理后台 - 全局状态
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
  // 侧边栏折叠状态
  const sidebarCollapsed = ref(false)

  // 当前高亮菜单
  const activeMenu = ref('')

  // 管理员信息 - 从 localStorage 读取持久化数据
  const adminInfo = ref(loadAdminInfo())

  // 通知数量
  const notificationCount = ref(3)

  // 计算侧边栏宽度
  const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '240px')

  /**
   * 从 localStorage 加载管理员信息
   */
  function loadAdminInfo() {
    try {
      const saved = localStorage.getItem('admin_info')
      if (saved) {
        return JSON.parse(saved)
      }
    } catch {
      // ignore
    }
    return null
  }

  /**
   * 检查是否已登录
   */
  const isLoggedIn = computed(() => {
    return !!localStorage.getItem('admin_token') && !!adminInfo.value
  })

  /**
   * 登录 - 保存 token 和管理员信息
   */
  function login(userInfo) {
    adminInfo.value = {
      name: userInfo.name || '管理员',
      avatar: userInfo.avatar || '管',
      role: userInfo.role || '超级管理员',
      email: userInfo.email || '',
      username: userInfo.username || ''
    }
    if (userInfo.token) {
      localStorage.setItem('admin_token', userInfo.token)
    }
    localStorage.setItem('admin_info', JSON.stringify(adminInfo.value))
  }

  /**
   * 登出 - 清除所有登录状态
   */
  function logout() {
    adminInfo.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
  }

  return {
    sidebarCollapsed,
    activeMenu,
    adminInfo,
    notificationCount,
    sidebarWidth,
    isLoggedIn,
    toggleSidebar() { sidebarCollapsed.value = !sidebarCollapsed.value },
    login,
    logout
  }
})
