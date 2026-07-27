/**
 * 认证相关 API
 */
import request from './index'

interface LoginParams {
  username: string
  password: string
}

interface LoginResult {
  token: string
}

interface RegisterParams {
  username: string
  password: string
}

/** 登录 */
export function login(data: LoginParams) {
  return request.post<any, LoginResult>('/user/v1/login', data)
}

/** 注册 */
export function register(data: RegisterParams) {
  return request.post<any, { success: boolean }>('/user/v1/register', data)
}
