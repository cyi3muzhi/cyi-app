/**
 * 悦选商城管理后台 - 角色权限 API
 */
import request from './index'
import type { RoleItem, PermModule } from '../types/api'

/** 获取角色列表 */
export function listRoles(): Promise<{ items: RoleItem[] }> {
  return request.get('/admin/v1/roles')
}

/** 创建角色 */
export function createRole(data: { name: string; slug: string; desc: string }): Promise<{ success: boolean }> {
  return request.post('/admin/v1/roles', data)
}

/** 更新角色 */
export function updateRole(slug: string, data: { name: string; desc: string }): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/roles/${slug}`, { slug, ...data })
}

/** 删除角色 */
export function deleteRole(slug: string): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/roles/${slug}`)
}

/** 权限值（proto PermissionValue 格式） */
interface PermissionValue { values: Record<string, boolean> }

/** 获取角色权限 */
export function getRolePermissions(role: string): Promise<{ permissions: Record<string, PermissionValue> }> {
  return request.get(`/admin/v1/roles/${encodeURIComponent(role)}/permissions`)
}

/** 更新角色权限 */
export function updateRolePermissions(role: string, permissions: Record<string, PermissionValue>): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/roles/${encodeURIComponent(role)}/permissions`, { role, permissions })
}

/** 获取所有权限模块定义 */
export function listAllPermissions(): Promise<{ modules: PermModule[] }> {
  return request.get('/admin/v1/permissions/modules')
}
