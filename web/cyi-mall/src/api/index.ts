/**
 * Axios 实例配置
 */
import axios from 'axios'
import type { ApiResponse } from '@/types/api'
import router from '@/router'

const instance = axios.create({
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 公开接口路径前缀，401 时不跳登录
const publicPaths = ['/settings/v1/', '/product/v1/suggest']

/** 判断是否为公开接口 */
function isPublicPath(url: string = ''): boolean {
  return publicPaths.some(p => url.startsWith(p))
}

// 请求拦截器 - 添加令牌
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 - 统一错误处理
instance.interceptors.response.use(
  (response) => {
    const body = response.data
    // Kratos proto 接口直接返回数据（无 code 字段），视为成功
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code !== 0) {
        return Promise.reject(new Error(body.message || '请求失败'))
      }
      return body.data
    }
    // 直接返回对象（如 settings、商品列表等）
    return body
  },
  (error) => {
    if (error.response?.status === 401 && !isPublicPath(error.config?.url)) {
      localStorage.removeItem('token')
      // 仅当当前页面需要登录时才跳转登录页（访问受版权限页/token失效）；
      // 游客浏览首页/商品等公开页时，接口 401 只静默清理 token，不强制跳登录，保证商城可正常逛
      if (router.currentRoute.value.meta?.requiresAuth) {
        router.push('/login')
      }
    }
    return Promise.reject(error)
  }
)

export default instance
