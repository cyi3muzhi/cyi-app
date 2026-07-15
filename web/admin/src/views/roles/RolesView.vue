<template>
  <!-- 角色管理页面 - 统计卡片 + Tab切换（角色列表/权限配置） -->
  <div class="roles-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">角色管理</h2>
        <span class="page-subtitle">管理后台角色与权限分配</span>
      </div>
      <el-button type="primary" @click="openAddModal">
        <el-icon><Plus /></el-icon>
        添加角色
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in stats" :key="stat.label">
        <div class="stat-header">
          <span class="stat-label">{{ stat.label }}</span>
          <div :class="['stat-icon', stat.color]">
            <el-icon :size="20"><component :is="stat.icon" /></el-icon>
          </div>
        </div>
        <div class="stat-value">{{ stat.value }} <span class="stat-unit">{{ stat.unit }}</span></div>
      </div>
    </div>

    <!-- 角色列表 + 权限配置 Tab 卡片 -->
    <el-card class="perm-card" shadow="never">
      <!-- Tab 导航 -->
      <div class="perm-tabs">
        <div
          class="perm-tab"
          :class="{ active: activeTab === 'list' }"
          @click="activeTab = 'list'">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
          角色列表
          <span class="tab-badge">{{ rolesData.length }}</span>
        </div>
        <div
          class="perm-tab"
          :class="{ active: activeTab === 'permission' }"
          @click="activeTab = 'permission'">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          权限配置
          <span class="tab-badge">{{ permSelectedCount }}</span>
        </div>
      </div>

      <!-- Tab 内容: 角色列表 -->
      <div v-show="activeTab === 'list'" class="tab-content">
        <div class="tab-header">
          <span style="font-weight:600;font-size:0.9375rem;">所有角色</span>
          <span class="tab-total">共 {{ rolesData.length }} 个角色</span>
        </div>
        <el-table :data="rolesData" stripe highlight-current-row>
          <el-table-column prop="name" label="角色名称" min-width="120">
            <template #default="{ row }">
              <span class="role-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="slug" label="角色标识" min-width="140">
            <template #default="{ row }">
              <span class="role-slug">{{ row.slug }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="desc" label="描述" min-width="160" show-overflow-tooltip />
          <el-table-column prop="users" label="用户数" width="80" align="center" />
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small" effect="plain">
                {{ row.enabled ? '启用' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="110" align="center">
            <template #default="{ row }">{{ row.createdAt }}</template>
          </el-table-column>
          <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row, $index }">
              <el-button text size="small" type="primary" @click="editPermission($index)">编辑权限</el-button>
              <el-button text size="small" type="warning" @click="openEditModal($index)">编辑</el-button>
              <el-button
                v-if="!row.isSystem"
                text size="small" type="danger"
                @click="openDeleteModal($index)">删除</el-button>
              <el-tag v-else size="small" type="info" effect="plain" style="cursor:default;">系统</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Tab 内容: 权限配置 -->
      <div v-show="activeTab === 'permission'" class="tab-content">
        <div class="perm-header">
          <h3>权限配置</h3>
          <div class="perm-summary">
            当前角色：<strong>{{ currentRole.name }}</strong>
            &nbsp;|&nbsp; 已选 <strong>{{ permSelectedCount }}</strong> 项权限
          </div>
        </div>
        <el-alert
          title="选中表格中的角色以编辑权限。超级管理员拥有所有系统权限，且不可修改。"
          type="info" :closable="false" show-icon class="perm-hint" />
        <div class="perm-body">
          <div class="perm-grid">
            <div class="perm-module" v-for="mod in permModules" :key="mod.key">
              <div class="perm-module-header">
                <span class="perm-module-title">{{ mod.title }}</span>
                <el-button
                  v-if="currentRole.slug !== 'super_admin'"
                  text size="small" type="primary"
                  @click="toggleSelectAll(mod.key)">
                  {{ isModuleAllChecked(mod.key) ? '取消全选' : '全选' }}
                </el-button>
                <span v-else class="perm-readonly">不可修改</span>
              </div>
              <div class="perm-module-items">
                <label
                  class="perm-checkbox"
                  v-for="item in mod.items"
                  :key="item.key"
                  :class="{ disabled: currentRole.slug === 'super_admin' }">
                  <el-checkbox
                    :model-value="getPermValue(mod.key, item.key)"
                    :disabled="currentRole.slug === 'super_admin'"
                    size="small"
                    @change="(val: boolean) => onPermChange(mod.key, item.key, val)" />
                  <span class="perm-label">{{ item.label }}</span>
                </label>
              </div>
            </div>
          </div>
        </div>
        <div class="perm-footer">
          <span class="perm-count">已选择 <strong>{{ permSelectedCount }}</strong> / {{ totalPermCount }} 项权限</span>
          <el-button @click="resetPermissions" :disabled="currentRole.slug === 'super_admin'">重置</el-button>
          <el-button type="primary" @click="savePermissions" :disabled="currentRole.slug === 'super_admin'">保存权限</el-button>
        </div>
      </div>
    </el-card>

    <!-- 添加/编辑角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      :title="editingIndex === -1 ? '添加角色' : '编辑角色'"
      width="480px"
      :close-on-click-modal="false">
      <el-form ref="roleFormRef" :model="roleForm" :rules="roleFormRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" @input="autoGenerateSlug" />
        </el-form-item>
        <el-form-item label="角色标识" prop="slug">
          <el-input v-model="roleForm.slug" placeholder="自动生成或手动输入" />
          <span class="form-hint">仅支持英文小写字母、数字和下划线</span>
        </el-form-item>
        <el-form-item label="角色描述">
          <el-input v-model="roleForm.desc" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="roleForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRole">{{ editingIndex === -1 ? '保存' : '保存修改' }}</el-button>
      </template>
    </el-dialog>

    <!-- 删除确认对话框 -->
    <el-dialog
      v-model="deleteDialogVisible"
      title="确认删除"
      width="400px"
      :close-on-click-modal="false">
      <div class="delete-confirm">
        <div class="confirm-icon">
          <el-icon :size="28"><WarningFilled /></el-icon>
        </div>
        <p>确定要删除角色 <strong>{{ deleteRoleName }}</strong> 吗？</p>
        <p class="confirm-hint">此操作不可撤销，该角色下的用户将失去相关权限。</p>
      </div>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 角色管理页面
 * 支持 Tab 切换：角色列表 / 权限配置，角色的增删改查
 */
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, WarningFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'

// ---------- 类型定义 ----------
interface PermItem {
  key: string
  label: string
}

interface PermModule {
  key: string
  title: string
  items: PermItem[]
}

interface RolePermissions {
  [moduleKey: string]: { [permKey: string]: boolean }
}

interface Role {
  id: number
  name: string
  slug: string
  desc: string
  users: number
  enabled: boolean
  createdAt: string
  isSystem: boolean
  permissions: RolePermissions
}

// ---------- 权限模块 ----------
const permModules: PermModule[] = [
  { key: 'products', title: '商品管理', items: [
    { key: 'view', label: '查看商品列表' }, { key: 'add', label: '添加商品' },
    { key: 'edit', label: '编辑商品' }, { key: 'delete', label: '删除商品' },
    { key: 'publish', label: '上架/下架' }
  ]},
  { key: 'orders', title: '订单管理', items: [
    { key: 'view', label: '查看订单' }, { key: 'edit', label: '编辑订单' },
    { key: 'refund', label: '退款处理' }, { key: 'ship', label: '发货操作' },
    { key: 'cancel', label: '取消订单' }
  ]},
  { key: 'users', title: '用户管理', items: [
    { key: 'view', label: '查看用户' }, { key: 'edit', label: '编辑用户' },
    { key: 'freeze', label: '冻结/解冻' }, { key: 'export', label: '导出数据' }
  ]},
  { key: 'settings', title: '系统设置', items: [
    { key: 'basic', label: '基本设置' }, { key: 'security', label: '安全设置' },
    { key: 'notification', label: '通知设置' }, { key: 'logs', label: '操作日志查看' }
  ]},
  { key: 'roles', title: '权限管理', items: [
    { key: 'view', label: '查看角色' }, { key: 'add', label: '添加角色' },
    { key: 'edit', label: '编辑角色' }, { key: 'delete', label: '删除角色' }
  ]},
  { key: 'coupons', title: '优惠券管理', items: [
    { key: 'view', label: '查看优惠券' }, { key: 'create', label: '创建优惠券' },
    { key: 'edit', label: '编辑优惠券' }, { key: 'void', label: '作废优惠券' }
  ]},
  { key: 'data', title: '数据查看', items: [
    { key: 'dashboard', label: '查看仪表盘' }, { key: 'report', label: '查看报表' },
    { key: 'export', label: '导出报表' }
  ]}
]

function createEmptyPermissions(): RolePermissions {
  const perms: RolePermissions = {}
  for (const mod of permModules) {
    perms[mod.key] = {}
    for (const item of mod.items) {
      perms[mod.key][item.key] = false
    }
  }
  return perms
}

// ---------- 模拟数据 ----------
const rolesData = reactive<Role[]>([
  { id: 1, name: '超级管理员', slug: 'super_admin', desc: '拥有所有系统权限', users: 1, enabled: true, createdAt: '2024-06-01', isSystem: true,
    permissions: { products: { view: true, add: true, edit: true, delete: true, publish: true }, orders: { view: true, edit: true, refund: true, ship: true, cancel: true }, users: { view: true, edit: true, freeze: true, export: true }, settings: { basic: true, security: true, notification: true, logs: true }, roles: { view: true, add: true, edit: true, delete: true }, coupons: { view: true, create: true, edit: true, void: true }, data: { dashboard: true, report: true, export: true } } },
  { id: 2, name: '管理员', slug: 'admin', desc: '除系统设置外的所有管理权限', users: 2, enabled: true, createdAt: '2024-06-01', isSystem: true,
    permissions: { products: { view: true, add: true, edit: true, delete: true, publish: true }, orders: { view: true, edit: true, refund: true, ship: true, cancel: true }, users: { view: true, edit: true, freeze: true, export: true }, settings: { basic: false, security: false, notification: false, logs: true }, roles: { view: true, add: false, edit: false, delete: false }, coupons: { view: true, create: true, edit: true, void: true }, data: { dashboard: true, report: true, export: true } } },
  { id: 3, name: '运营专员', slug: 'operation', desc: '商品管理+订单管理+用户查看', users: 5, enabled: true, createdAt: '2024-06-15', isSystem: true,
    permissions: { products: { view: true, add: true, edit: true, delete: false, publish: true }, orders: { view: true, edit: true, refund: false, ship: true, cancel: false }, users: { view: true, edit: false, freeze: false, export: false }, settings: { basic: false, security: false, notification: false, logs: false }, roles: { view: false, add: false, edit: false, delete: false }, coupons: { view: true, create: false, edit: false, void: false }, data: { dashboard: true, report: false, export: false } } },
  { id: 4, name: '客服专员', slug: 'support', desc: '订单查看+用户查看+售后处理', users: 7, enabled: true, createdAt: '2024-08-01', isSystem: true,
    permissions: { products: { view: false, add: false, edit: false, delete: false, publish: false }, orders: { view: true, edit: false, refund: true, ship: false, cancel: false }, users: { view: true, edit: false, freeze: false, export: false }, settings: { basic: false, security: false, notification: false, logs: false }, roles: { view: false, add: false, edit: false, delete: false }, coupons: { view: false, create: false, edit: false, void: false }, data: { dashboard: true, report: false, export: false } } },
  { id: 5, name: '数据分析师', slug: 'analyst', desc: '数据查看+报表导出', users: 3, enabled: true, createdAt: '2025-01-10', isSystem: true,
    permissions: { products: { view: false, add: false, edit: false, delete: false, publish: false }, orders: { view: false, edit: false, refund: false, ship: false, cancel: false }, users: { view: false, edit: false, freeze: false, export: false }, settings: { basic: false, security: false, notification: false, logs: false }, roles: { view: false, add: false, edit: false, delete: false }, coupons: { view: false, create: false, edit: false, void: false }, data: { dashboard: true, report: true, export: true } } },
  { id: 6, name: '自定义角色', slug: 'custom_role', desc: '自定义权限组合', users: 1, enabled: false, createdAt: '2025-03-20', isSystem: false,
    permissions: { products: { view: true, add: false, edit: false, delete: false, publish: false }, orders: { view: true, edit: false, refund: false, ship: false, cancel: false }, users: { view: true, edit: false, freeze: false, export: false }, settings: { basic: false, security: false, notification: false, logs: false }, roles: { view: true, add: false, edit: false, delete: false }, coupons: { view: false, create: false, edit: false, void: false }, data: { dashboard: true, report: false, export: false } } }
])

// ---------- Tab 切换 ----------
const activeTab = ref<'list' | 'permission'>('list')

// ---------- 选中角色 ----------
const selectedIndex = ref(0)
const currentRole = computed(() => rolesData[selectedIndex.value])

// ---------- 统计卡片 ----------
const stats = computed(() => [
  { label: '角色总数', value: rolesData.length, unit: '个', color: 'blue', icon: 'Shield' },
  { label: '管理员', value: rolesData.filter(r => r.name === '超级管理员' || r.name === '管理员').reduce((s, r) => s + r.users, 0), unit: '人', color: 'green', icon: 'UserFilled' },
  { label: '运营人员', value: rolesData.filter(r => r.name === '运营专员' || r.name === '客服专员').reduce((s, r) => s + r.users, 0), unit: '人', color: 'orange', icon: 'Avatar' },
  { label: '自定义角色', value: rolesData.filter(r => !r.isSystem).length, unit: '个', color: 'purple', icon: 'SetUp' }
])

// ---------- 权限相关 ----------
const totalPermCount = computed(() => permModules.reduce((s, m) => s + m.items.length, 0))

function countPermissions(perms: RolePermissions): number {
  let count = 0
  for (const modKey in perms) {
    for (const permKey in perms[modKey]) {
      if (perms[modKey][permKey]) count++
    }
  }
  return count
}

const permSelectedCount = computed(() => countPermissions(currentRole.value.permissions))

function getPermValue(modKey: string, permKey: string): boolean {
  return currentRole.value.permissions[modKey]?.[permKey] ?? false
}

function onPermChange(modKey: string, permKey: string, val: boolean) {
  if (currentRole.value.slug === 'super_admin') return
  currentRole.value.permissions[modKey][permKey] = val
}

function isModuleAllChecked(modKey: string): boolean {
  const modPerms = currentRole.value.permissions[modKey]
  if (!modPerms) return false
  return permModules.find(m => m.key === modKey)?.items.every(item => modPerms[item.key] === true) ?? false
}

function toggleSelectAll(modKey: string) {
  if (currentRole.value.slug === 'super_admin') return
  const allChecked = isModuleAllChecked(modKey)
  const modPerms = currentRole.value.permissions[modKey]
  for (const item of permModules.find(m => m.key === modKey)?.items ?? []) {
    modPerms[item.key] = !allChecked
  }
}

function resetPermissions() {
  if (currentRole.value.slug === 'super_admin') return
  for (const mod of permModules) {
    for (const item of mod.items) {
      currentRole.value.permissions[mod.key][item.key] = false
    }
  }
}

function savePermissions() {
  ElMessage.success(`角色「${currentRole.value.name}」权限已保存（共 ${permSelectedCount.value} 项权限）`)
}

// ---------- 编辑权限（点击表格操作列） ----------
function editPermission(index: number) {
  selectedIndex.value = index
  activeTab.value = 'permission'
}

// ---------- 角色表单 ----------
const roleDialogVisible = ref(false)
const deleteDialogVisible = ref(false)
const editingIndex = ref(-1)
const deleteIndex = ref(-1)
const deleteRoleName = ref('')
const roleFormRef = ref<FormInstance>()

const roleForm = reactive({
  name: '',
  slug: '',
  desc: '',
  enabled: true
})

const roleFormRules: FormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入角色标识', trigger: 'blur' }]
}

function openAddModal() {
  editingIndex.value = -1
  roleForm.name = ''
  roleForm.slug = ''
  roleForm.desc = ''
  roleForm.enabled = true
  roleDialogVisible.value = true
}

function openEditModal(index: number) {
  const role = rolesData[index]
  editingIndex.value = index
  roleForm.name = role.name
  roleForm.slug = role.slug
  roleForm.desc = role.desc
  roleForm.enabled = role.enabled
  roleDialogVisible.value = true
}

function autoGenerateSlug() {
  const name = roleForm.name
  if (!name) return
  roleForm.slug = name.toLowerCase()
    .replace(/\s+/g, '_')
    .replace(/[^a-z0-9_\u4e00-\u9fa5]/g, '')
    .replace(/[\u4e00-\u9fa5]/g, 'x')
}

function saveRole() {
  roleFormRef.value?.validate((valid) => {
    if (!valid) return
    if (editingIndex.value >= 0) {
      const role = rolesData[editingIndex.value]
      role.name = roleForm.name
      role.slug = roleForm.slug
      role.desc = roleForm.desc
      role.enabled = roleForm.enabled
      ElMessage.success('角色已更新')
    } else {
      const newRole: Role = {
        id: rolesData.length + 1,
        name: roleForm.name,
        slug: roleForm.slug,
        desc: roleForm.desc,
        users: 0,
        enabled: roleForm.enabled,
        createdAt: new Date().toISOString().slice(0, 10),
        isSystem: false,
        permissions: createEmptyPermissions()
      }
      rolesData.push(newRole)
      ElMessage.success('角色已添加')
    }
    roleDialogVisible.value = false
  })
}

function openDeleteModal(index: number) {
  deleteIndex.value = index
  deleteRoleName.value = rolesData[index].name
  deleteDialogVisible.value = true
}

function confirmDelete() {
  if (deleteIndex.value >= 0) {
    rolesData.splice(deleteIndex.value, 1)
    if (selectedIndex.value >= rolesData.length) {
      selectedIndex.value = Math.max(0, rolesData.length - 1)
    }
    ElMessage.success('角色已删除')
  }
  deleteDialogVisible.value = false
}
</script>

<style scoped>
.roles-page {
  padding-bottom: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0;
}

.page-subtitle {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin-top: 2px;
  display: block;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 18px 20px;
  transition: box-shadow 0.2s, transform 0.2s;
}
.stat-card:hover {
  box-shadow: var(--shop-shadow-md);
  transform: translateY(-1px);
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon.blue { background: var(--shop-primary-bg); color: var(--shop-primary); }
.stat-icon.green { background: #ecfdf5; color: #059669; }
.stat-icon.orange { background: #fff7ed; color: #d97706; }
.stat-icon.purple { background: #f5f3ff; color: #7c3aed; }

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
}
.stat-unit {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
}

/* Tab 卡片 */
.perm-card {
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
}

.perm-tabs {
  display: flex;
  border-bottom: 1px solid var(--shop-border);
  background: #f8fafc;
}

.perm-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  user-select: none;
}

.perm-tab:hover {
  color: var(--shop-foreground);
  background: rgba(0,0,0,0.02);
}

.perm-tab.active {
  color: var(--shop-primary);
  border-bottom-color: var(--shop-primary);
  background: #fff;
}

.tab-badge {
  font-size: 0.6875rem;
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
  padding: 1px 7px;
  border-radius: 999px;
  font-weight: 600;
}

.perm-tab.active .tab-badge {
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

.tab-content {
  padding: 0;
}

.tab-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--shop-border);
}

.tab-total {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  font-weight: 400;
}

/* 角色表格内样式 */
.role-name {
  font-weight: 600;
}

.role-slug {
  font-family: var(--shop-font-mono);
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  background: var(--shop-muted);
  padding: 2px 8px;
  border-radius: var(--shop-radius-sm);
}

/* 权限区域 */
.perm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--shop-border);
}

.perm-header h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
}

.perm-summary {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  font-weight: 400;
}

.perm-summary strong {
  color: var(--shop-primary);
  font-weight: 700;
}

.perm-hint {
  margin: 16px 20px 0;
}

.perm-body {
  padding: 16px 20px;
}

.perm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.perm-module {
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
  transition: box-shadow 0.2s;
}
.perm-module:hover {
  box-shadow: var(--shop-shadow-sm);
}

.perm-module-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--shop-muted);
  border-bottom: 1px solid var(--shop-border);
}

.perm-module-title {
  font-size: 0.8125rem;
  font-weight: 700;
}

.perm-readonly {
  font-size: 0.6875rem;
  color: #94a3b8;
}

.perm-module-items {
  padding: 8px 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}

.perm-checkbox {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: var(--shop-radius-md);
  cursor: pointer;
  transition: background 0.15s;
  user-select: none;
  font-size: 0.8125rem;
}
.perm-checkbox:hover:not(.disabled) { background: var(--shop-muted); }
.perm-checkbox.disabled { cursor: not-allowed; opacity: 0.7; }

.perm-label {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.perm-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--shop-border);
}

.perm-count {
  margin-right: auto;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.perm-count strong {
  color: var(--shop-primary);
  font-weight: 700;
}

/* 删除确认 */
.delete-confirm {
  text-align: center;
  padding: 8px 0;
}
.confirm-icon {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  background: #fef2f2;
  color: var(--shop-destructive);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;
}
.delete-confirm p {
  font-size: 0.875rem;
  color: var(--shop-foreground);
  margin: 4px 0;
}
.confirm-hint {
  font-size: 0.75rem !important;
  color: var(--shop-muted-foreground) !important;
}

.form-hint {
  display: block;
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
  line-height: 1.4;
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr 1fr; }
  .page-header { flex-direction: column; align-items: flex-start; gap: 12px; }
}
</style>
