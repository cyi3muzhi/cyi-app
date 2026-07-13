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

  // 管理员信息
  const adminInfo = ref({
    name: '管理员',
    avatar: '管',
    role: '超级管理员',
    email: 'admin@yuexuan.com'
  })

  // 通知数量
  const notificationCount = ref(3)

  // 计算侧边栏宽度
  const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '240px')

  /**
   * 切换侧边栏折叠状态
   */
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  /**
   * 登录
   */
  function login(userInfo) {
    adminInfo.value = userInfo
    localStorage.setItem('admin_token', 'mock_token_' + Date.now())
  }

  /**
   * 登出
   */
  function logout() {
    localStorage.removeItem('admin_token')
  }

  return {
    sidebarCollapsed,
    activeMenu,
    adminInfo,
    notificationCount,
    sidebarWidth,
    toggleSidebar,
    login,
    logout
  }
})
