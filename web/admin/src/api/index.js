/**
 * 悦选商城管理后台 - API 请求封装
 */
import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useAppStore } from '../store/app'

/** 防止 401 重复处理（多个请求同时失败时只处理一次） */
let isHandling401 = false

const request = axios.create({
  baseURL: '',
  timeout: 15000
})

// 请求拦截器 - 添加 token
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器 - 统一错误处理
request.interceptors.response.use(
  response => {
    const data = response.data
    if (data.code !== 0 && data.code !== undefined) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }
    return data
  },
  error => {
    if (error.response) {
      const { status } = error.response
      if (status === 401) {
        // 防止多个请求同时 401 时重复弹窗和跳转
        if (isHandling401) {
          return Promise.reject(error)
        }
        isHandling401 = true

        const message = error.response.data?.message || '登录已过期，请重新登录'
        ElMessage.error(message)

        // 延迟清除状态，确保 ElMessage 先渲染完成
        setTimeout(() => {
          const store = useAppStore()
          store.logout()
          router.push('/login')
          isHandling401 = false
        }, 1500)
      } else if (status === 403) {
        ElMessage.error('没有权限执行此操作')
      } else {
        ElMessage.error(error.response.data?.message || '服务器错误')
      }
    } else {
      ElMessage.error('网络连接失败，请检查网络')
    }
    return Promise.reject(error)
  }
)

export default request
