/**
 * 悦选商城管理后台 - 认证 API
 */
import request from './index'

/**
 * 管理员登录
 * @param {string} username - 管理员账号
 * @param {string} password - 管理员密码
 * @returns {Promise<{token: string, admin: object}>}
 */
export function login(username, password) {
  return request.post('/admin/v1/auth/login', { username, password })
}

/**
 * 管理员登出
 * token 通过请求拦截器自动添加到 Authorization 头
 * @returns {Promise}
 */
export function logout() {
  return request.post('/admin/v1/auth/logout')
}

/**
 * 获取当前管理员信息
 * @returns {Promise<{admin: object}>}
 */
export function getAdminInfo() {
  return request.get('/admin/v1/auth/info')
}
