/**
 * 用户状态管理
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const nickname = ref('')
  const avatar = ref('')

  const isLoggedIn = computed(() => !!token.value)

  /** 设置登录态 */
  function setLogin(t: string, nick: string, ava: string) {
    token.value = t
    nickname.value = nick
    avatar.value = ava
    localStorage.setItem('token', t)
  }

  /** 退出登录 */
  function logout() {
    token.value = ''
    nickname.value = ''
    avatar.value = ''
    localStorage.removeItem('token')
  }

  return { token, nickname, avatar, isLoggedIn, setLogin, logout }
})
