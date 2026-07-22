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
          <span class="tab-badge">{{ roles.length }}</span>
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
          <span class="tab-total">共 {{ roles.length }} 个角色</span>
        </div>
        <el-table :data="roles" stripe highlight-current-row>
          <el-table-column prop="name" label="角色名称" min-width="120">
            <template #default="{ row }">
              <el-tag :color="row.color" effect="dark" size="small" class="color-tag">{{ row.name }}</el-tag>
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
            <template #default="{ row }">{{ row.created_at }}</template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row, $index }">
              <template v-if="row.slug === 'super_admin'">
                <span class="role-system-tag">系统角色，不可操作</span>
              </template>
              <template v-else>
                <el-button text size="small" type="primary" @click="editPermission($index)">编辑权限</el-button>
                <el-button text size="small" type="warning" @click="openEditModal($index)">编辑</el-button>
                <el-button
                  v-if="!row.is_system"
                  text size="small" type="danger"
                  @click="openDeleteModal($index)">删除</el-button>
                <el-tag v-else size="small" type="info" effect="plain" style="cursor:default;">系统</el-tag>
              </template>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Tab 内容: 权限配置 -->
      <div v-if="currentRole && activeTab === 'permission'" class="tab-content">
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
                  v-if="!isSuperAdmin"
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
                  :class="{ disabled: isSuperAdmin }">
                  <el-checkbox
                    :model-value="getPermValue(mod.key, item.key)"
                    :disabled="isSuperAdmin"
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
          <el-button @click="resetPermissions" :disabled="isSuperAdmin">重置</el-button>
          <el-button type="primary" @click="savePermissions" :disabled="isSuperAdmin">保存权限</el-button>
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
          <el-input v-model="roleForm.slug" placeholder="自动生成或手动输入" :disabled="editingIndex >= 0" />
          <span class="form-hint">{{ editingIndex >= 0 ? '创建后不可修改' : '仅支持英文小写字母、数字和下划线' }}</span>
        </el-form-item>
        <el-form-item label="角色描述">
          <el-input v-model="roleForm.desc" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
        <el-form-item label="标签颜色">
          <el-color-picker v-model="roleForm.color" show-alpha :predefine="predefineColors" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="roleForm.enabled" active-text="启用" inactive-text="停用" :disabled="isEditingSuperAdmin" />
          <span v-if="isEditingSuperAdmin" class="form-hint">系统角色状态不可修改</span>
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
 * 对接后端 Casbin RBAC API
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, WarningFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { listRoles, createRole, updateRole, deleteRole, getRolePermissions, updateRolePermissions, listAllPermissions } from '@/api/role'

// ---------- 类型定义 ----------
interface PermModule {
  key: string
  title: string
  items: { key: string; label: string }[]
}

interface RolePermissions {
  [moduleKey: string]: { [permKey: string]: boolean }
}

interface Role {
  name: string
  slug: string
  desc: string
  color?: string
  users: number
  enabled: boolean
  is_system: boolean
  created_at: string
  permissions: RolePermissions
}

// ---------- 状态 ----------
const permModules = ref<PermModule[]>([])
const roles = ref<Role[]>([])
const loading = ref(false)

// ---------- Tab 切换 ----------
const activeTab = ref<'list' | 'permission'>('list')

// ---------- 选中角色 ----------
const selectedIndex = ref(0)
const currentRole = computed(() => roles.value[selectedIndex.value])
const isSuperAdmin = computed(() => {
  const role = currentRole.value
  return role?.slug === 'super_admin'
})
const isEditingSuperAdmin = computed(() => {
  if (editingIndex.value < 0) return false
  const role = roles.value[editingIndex.value]
  return role?.slug === 'super_admin'
})

// ---------- 数据加载 ----------
async function loadData() {
  loading.value = true
  try {
    const [roleRes, permRes] = await Promise.all([
      listRoles().catch((e: any) => {
        ElMessage.error('加载角色列表失败: ' + (e?.message || '未知错误'))
        return { items: [] }
      }),
      listAllPermissions().catch(() => ({ modules: [] }))
    ])
    // 转换后端角色数据为前端格式
    roles.value = (roleRes.items || []).map((r: any) => ({
      name: r.name,
      slug: r.slug || r.name,
      desc: r.desc || '',
      color: r.color || '',
      users: r.users || 0,
      enabled: r.enabled !== false,
      is_system: r.is_system === true,
      created_at: r.createdAt || r.created_at || '',
      permissions: {} as RolePermissions
    }))
    permModules.value = permRes.modules || []

    // 加载第一个角色的权限
    if (roles.value.length > 0) {
      await loadRolePermissions(0)
    }
  } finally {
    loading.value = false
  }
}

async function loadRolePermissions(index: number) {
  if (index < 0 || index >= roles.value.length) return
  const role = roles.value[index]
  try {
    const res = await getRolePermissions(role.slug)
    // 解包 PermissionValue.values
    const raw = res.permissions || {}
    const unwrapped: RolePermissions = {}
    for (const modKey in raw) {
      unwrapped[modKey] = (raw[modKey]?.values || {}) as Record<string, boolean>
    }
    role.permissions = unwrapped
  } catch {
    // API 失败时，超级管理员默认全选，其他角色全不选
    const perms: RolePermissions = {}
    for (const mod of permModules.value) {
      perms[mod.key] = {}
      for (const item of mod.items) {
        perms[mod.key][item.key] = role.slug === 'super_admin'
      }
    }
    role.permissions = perms
  }
}

onMounted(() => loadData())

// ---------- 统计卡片 ----------
const stats = computed(() => {
  const data = roles.value
  const cards: { label: string; value: number; unit: string; color: string; icon: string }[] = [
    { label: '角色总数', value: data.length, unit: '个', color: 'blue', icon: 'Shield' },
  ]
  // 每个角色一行（超级管理员固定不展示）
   const colors = ['green', 'orange', 'purple', 'pink', 'teal', 'indigo']
   data.filter((r: Role) => r.slug !== 'super_admin').forEach((r, idx) => {
     cards.push({
       label: r.name,
       value: r.users || 0,
       unit: '人',
       color: colors[idx % colors.length],
       icon: idx === 0 ? 'UserFilled' : idx === 1 ? 'Avatar' : idx === 2 ? 'SetUp' : 'User',
     })
   })
  return cards
})

// ---------- 权限相关 ----------
const totalPermCount = computed(() => permModules.value.reduce((s: number, m: PermModule) => s + m.items.length, 0))

function countPermissions(perms: RolePermissions): number {
  let count = 0
  for (const modKey in perms) {
    for (const permKey in perms[modKey]) {
      if (perms[modKey][permKey]) count++
    }
  }
  return count
}

const permSelectedCount = computed(() => countPermissions(currentRole.value?.permissions || {}))

function getPermValue(modKey: string, permKey: string): boolean {
  return currentRole.value?.permissions?.[modKey]?.[permKey] ?? false
}

function onPermChange(modKey: string, permKey: string, val: boolean) {
  if (!currentRole.value || isSuperAdmin.value) return
  currentRole.value.permissions[modKey][permKey] = val
}

function isModuleAllChecked(modKey: string): boolean {
  const modPerms = currentRole.value?.permissions?.[modKey]
  if (!modPerms) return false
  const mod = permModules.value.find(m => m.key === modKey)
  if (!mod) return false
  return mod.items.every(item => modPerms[item.key] === true)
}

function toggleSelectAll(modKey: string) {
  if (!currentRole.value || isSuperAdmin.value) return
  const allChecked = isModuleAllChecked(modKey)
  const modPerms = currentRole.value.permissions[modKey]
  const mod = permModules.value.find(m => m.key === modKey)
  if (!mod || !modPerms) return
  for (const item of mod.items) {
    modPerms[item.key] = !allChecked
  }
}

function resetPermissions() {
  if (!currentRole.value || isSuperAdmin.value) return
  for (const mod of permModules.value) {
    for (const item of mod.items) {
      currentRole.value.permissions[mod.key][item.key] = false
    }
  }
}

async function savePermissions() {
  if (!currentRole.value) return
  try {
    // 包裹成 proto 需要的 { values: {...} } 格式
    const wrapped: Record<string, { values: Record<string, boolean> }> = {}
    for (const modKey in currentRole.value.permissions) {
      wrapped[modKey] = { values: currentRole.value.permissions[modKey] }
    }
    await updateRolePermissions(currentRole.value.slug, wrapped)
    ElMessage.success(`角色「${currentRole.value.name}」权限已保存（共 ${permSelectedCount.value} 项权限）`)
  } catch (e: any) {
    ElMessage.error('保存权限失败: ' + (e?.message || ''))
  }
}

// ---------- 编辑权限 ----------
async function editPermission(index: number) {
  selectedIndex.value = index
  // 如果该角色还没加载过权限，加载它
  if (!roles.value[index]?.permissions || Object.keys(roles.value[index].permissions).length === 0) {
    await loadRolePermissions(index)
  }
  activeTab.value = 'permission'
}

// ---------- 颜色预定义 ----------
const predefineColors = [
  '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399',
  '#9b59b6', '#2ecc71', '#e74c3c', '#3498db', '#1abc9c',
  '#f39c12', '#8e44ad', '#2c3e50', '#d35400', '#16a085'
]

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
  enabled: true,
  color: '#409EFF'
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
  roleForm.color = '#409EFF'
  roleDialogVisible.value = true
}

function openEditModal(index: number) {
  const role = roles.value[index]
  editingIndex.value = index
  roleForm.name = role.name
  roleForm.slug = role.slug
  roleForm.desc = role.desc
  roleForm.enabled = role.enabled !== false
  roleForm.color = role.color || '#409EFF'
  roleDialogVisible.value = true
}

function autoGenerateSlug() {
  // 编辑模式不自动生成 slug（创建后不可修改）
  if (editingIndex.value >= 0) return
  const name = roleForm.name
  if (!name) return
  roleForm.slug = name.toLowerCase()
    .replace(/\s+/g, '_')
    .replace(/[^a-z0-9_\u4e00-\u9fa5]/g, '')
    .replace(/[\u4e00-\u9fa5]/g, 'x')
}

async function saveRole() {
  roleFormRef.value?.validate(async (valid) => {
    if (!valid) return
    try {
      if (editingIndex.value >= 0) {
        await updateRole(roleForm.slug, { name: roleForm.name, desc: roleForm.desc, color: roleForm.color })
        const role = roles.value[editingIndex.value]
        role.name = roleForm.name
        role.desc = roleForm.desc
        role.color = roleForm.color
        ElMessage.success('角色已更新')
      } else {
        await createRole({ name: roleForm.name, slug: roleForm.slug, desc: roleForm.desc, color: roleForm.color })
        // 重新加载列表
        await loadData()
        ElMessage.success('角色已添加')
      }
      roleDialogVisible.value = false
    } catch (e: any) {
      ElMessage.error('操作失败: ' + (e?.message || ''))
    }
  })
}

function openDeleteModal(index: number) {
  deleteIndex.value = index
  deleteRoleName.value = roles.value[index].name
  deleteDialogVisible.value = true
}

async function confirmDelete() {
  if (deleteIndex.value >= 0 && deleteIndex.value < roles.value.length) {
    try {
      await deleteRole(roles.value[deleteIndex.value].slug)
      roles.value.splice(deleteIndex.value, 1)
      if (selectedIndex.value >= roles.value.length) {
        selectedIndex.value = Math.max(0, roles.value.length - 1)
      }
      ElMessage.success('角色已删除')
    } catch (e: any) {
      ElMessage.error('删除失败: ' + (e?.message || ''))
    }
  }
  deleteDialogVisible.value = false
}
</script>

<style scoped>
/* 自定义颜色标签去掉 Element Plus 默认蓝色边框 */
.color-tag {
  border: none !important;
}
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

.role-system-tag {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
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
