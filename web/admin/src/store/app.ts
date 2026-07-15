/**
 * 悦选商城管理后台 - 全局状态
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdminInfo } from '../types/api'

/** localStorage 中保存的管理员信息 */
interface StoredAdminInfo {
  name: string
  avatar: string
  role: string
  email: string
  username: string
}

export const useAppStore = defineStore('app', () => {
  // 侧边栏折叠状态
  const sidebarCollapsed = ref(false)

  // 当前高亮菜单
  const activeMenu = ref('')

  // 管理员信息 - 从 localStorage 读取持久化数据
  const adminInfo = ref<AdminInfo | StoredAdminInfo | null>(loadAdminInfo())

  // 用户权限映射（模块 → 权限列表）
  const userPermissions = ref<Record<string, Record<string, boolean>>>({})

  // 通知数量
  const notificationCount = ref(3)

  // 计算侧边栏宽度
  const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '240px')

  /**
   * 从 localStorage 加载管理员信息
   */
  function loadAdminInfo(): StoredAdminInfo | null {
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
  function login(userInfo: { token?: string; name?: string; avatar?: string; role?: string; email?: string; username?: string }) {
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
   * 设置用户权限
   */
  function setPermissions(perms: Record<string, Record<string, boolean>>) {
    userPermissions.value = perms
  }

  /**
   * 检查是否有某权限
   */
  function hasPerm(module: string, action: string): boolean {
    const mod = userPermissions.value[module]
    if (!mod) return false
    return !!mod[action]
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
    userPermissions,
    notificationCount,
    sidebarWidth,
    isLoggedIn,
    toggleSidebar() { sidebarCollapsed.value = !sidebarCollapsed.value },
    login,
    setPermissions,
    hasPerm,
    logout
  }
})
