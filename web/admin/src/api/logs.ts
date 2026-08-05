/**
 * cyi商城管理后台 - 操作日志 API
 */
import request from './index'

/** 操作日志项 */
export interface LogItem {
  id: number
  operator: string // 操作人姓名
  role: string // 角色标识
  module: string // 操作模块
  action: string // 动作
  content: string // 操作内容
  method: string // HTTP 方法
  path: string // 请求路径
  ip: string // 来源 IP
  user_agent: string // User-Agent
  result: number // 1-成功 0-失败
  created_at: number // 操作时间（Unix 秒）
  avatar: string // 操作人头像（公开 URL，可能为空）
}

/** 日志统计 */
export interface LogStats {
  today: number
  week: number
  month: number
  total: number
}

/** 查询参数 */
export interface LogQuery {
  page?: number
  page_size?: number
  start_date?: string
  end_date?: string
  operator?: string
  module?: string
  keyword?: string
}

/**
 * 归一化日志项：Kratos protojson 输出 camelCase（createdAt/userAgent），
 * 统一转为 snake_case 供组件使用
 */
function normalizeLog(item: any): LogItem {
  return {
    id: Number(item.id ?? 0),
    operator: item.operator ?? '',
    role: item.role ?? '',
    module: item.module ?? '',
    action: item.action ?? '',
    content: item.content ?? '',
    method: item.method ?? '',
    path: item.path ?? '',
    ip: item.ip ?? '',
    user_agent: item.userAgent ?? item.user_agent ?? '',
    result: Number(item.result ?? 1),
    created_at: Number(item.createdAt ?? item.created_at ?? 0),
    avatar: item.avatar ?? ''
  }
}

/** 分页查询操作日志 */
export async function getLogs(params: LogQuery = {}) {
  const res = await request.get<any, { items: any[]; total: number }>('/admin/v1/logs', { params })
  return {
    items: (res.items || []).map(normalizeLog),
    total: res.total || 0
  }
}

/** 操作日志统计 */
export function getLogStats() {
  return request.get<any, LogStats>('/admin/v1/logs/stats')
}

/** 操作人列表 */
export function getLogOperators() {
  return request.get<any, { operators: string[] }>('/admin/v1/logs/operators')
}
