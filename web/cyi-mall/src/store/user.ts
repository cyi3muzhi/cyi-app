/**
 * 用户状态管理
 */
import { defineStore, acceptHMRUpdate } from 'pinia'
import { ref, computed } from 'vue'
import { getMyInfo } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const nickname = ref('')
  const avatar = ref('')
  /** 当前用户 ID（用于搜索历史等按用户隔离的本地数据） */
  const userId = ref<number | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  /** 设置登录态 */
  function setLogin(t: string, nick: string, ava: string) {
    token.value = t
    nickname.value = nick
    avatar.value = ava
    localStorage.setItem('token', t)
  }

  /** 拉取当前登录用户信息（更新昵称/头像），登录后或页面刷新时调用 */
  async function fetchUserInfo() {
    if (!token.value) return
    try {
      const user = await getMyInfo()
      nickname.value = user.nickname
      avatar.value = user.avatar || ''
      userId.value = user.id || null
    } catch {
      // 拉取失败静默处理（如 token 过期时由请求拦截器统一跳转登录页）
    }
  }

  /** 退出登录 */
  function logout() {
    token.value = ''
    nickname.value = ''
    avatar.value = ''
    userId.value = null
    localStorage.removeItem('token')
  }

  return { token, nickname, avatar, userId, isLoggedIn, setLogin, fetchUserInfo, logout }
})

// 启用 store 热更新（修改此文件时无需整页刷新）
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useUserStore, import.meta.hot))
}
