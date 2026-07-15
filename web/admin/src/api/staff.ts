/**
 * 悦选商城管理后台 - 员工管理 API
 */
import request from './index'

/** 员工信息 */
export interface StaffInfo {
  id: number
  name: string
  username: string
  phone: string
  email: string
  role: string
  department: string
  position: string
  bio: string
  avatar: string
  status: number
  last_login: string
  created_at: string
}

/** 获取员工列表 */
export function listStaff(params: { page?: number; page_size?: number; keyword?: string; status?: string }): Promise<{ items: StaffInfo[]; total: number }> {
  return request.get('/admin/v1/staff', { params })
}

/** 创建员工 */
export function createStaff(data: { username: string; password: string; name: string; role: string; email: string; phone?: string; department?: string }): Promise<{ id: number }> {
  return request.post('/admin/v1/staff', data)
}

/** 更新员工 */
export function updateStaff(id: number, data: { name: string; role: string; email: string; phone?: string; department?: string; position?: string; bio?: string }): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/staff/${id}`, data)
}

/** 切换员工状态 */
export function toggleStaffStatus(id: number): Promise<{ success: boolean; new_status: number }> {
  return request.post(`/admin/v1/staff/${id}/toggle-status`, {})
}

/** 获取在线管理员 ID 列表 */
export function getOnlineAdminIds(): Promise<{ ids: number[] }> {
  return request.get('/admin/v1/auth/online-ids')
}

/** 获取角色名称映射（无需角色管理权限） */
export function getRoleNameMap(): Promise<{ items: Record<string, string> }> {
  return request.get('/admin/v1/role-name-map')
}
