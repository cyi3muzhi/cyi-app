/**
 * 认证相关 API
 */
import request from './index'

interface LoginParams {
  account: string
  password: string
}

interface LoginResult {
  token: string
  user: {
    id: number
    nickname: string
    avatar: string
  }
}

interface RegisterParams {
  nickname: string
  phone: string
  password: string
}

/** 登录 */
export function login(data: LoginParams) {
  return request.post<any, LoginResult>('/auth/login', data)
}

/** 注册 */
export function register(data: RegisterParams) {
  return request.post<any, void>('/auth/register', data)
}

/** 退出登录 */
export function logout() {
  return request.post<any, void>('/auth/logout')
}
