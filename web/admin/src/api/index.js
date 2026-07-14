/**
 * 悦选商城管理后台 - API 请求封装
 */
import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'
import { useAppStore } from '../store/app'

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
        const store = useAppStore()
        const message = error.response.data?.message || '登录已过期，请重新登录'
        ElMessage.error(message)
        store.logout()
        setTimeout(() => {
          router.push('/login')
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
