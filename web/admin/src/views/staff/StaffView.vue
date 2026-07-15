<template>
  <!-- 员工管理页面 - 统计卡片 + 筛选栏 + 员工表格 + 新增/编辑对话框 -->
  <div class="staff-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">员工管理</h2>
        <span class="page-subtitle">管理所有后台账号与员工信息</span>
      </div>
      <el-button v-if="store.hasPerm('staff', 'add')" type="primary" @click="openAddModal">
        <el-icon><Plus /></el-icon>
        添加员工
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
        <div v-if="stat.change" :class="['stat-change', stat.change > 0 ? 'up' : 'down']">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <polyline v-if="stat.change > 0" points="18 15 12 9 6 15" />
            <polyline v-else points="6 9 12 15 18 9" />
          </svg>
          {{ stat.changeText }}
        </div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-bar">
        <div class="filter-group">
          <label>搜索</label>
          <el-input v-model="searchKeyword" placeholder="姓名 / 手机号 / 邮箱" clearable size="default" style="width:200px" @input="onSearchInput" />
        </div>
        <div class="filter-group">
          <label>角色</label>
          <el-select v-model="roleFilter" placeholder="全部" clearable size="default" style="width:150px" @change="applyFilters">
            <el-option v-for="r in roleOptions" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
        </div>
        <div class="filter-group">
          <label>状态</label>
          <el-select v-model="statusFilter" placeholder="全部" clearable size="default" style="width:120px" @change="applyFilters">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </div>
        <el-button type="primary" size="default" @click="applyFilters">搜索</el-button>
        <el-button size="default" @click="resetFilters">重置</el-button>
      </div>
    </el-card>

    <!-- 员工表格 -->
    <el-card class="table-card" shadow="never">
      <el-table :data="staffList" stripe style="width:100%" v-loading="loading">
        <el-table-column label="姓名" min-width="180">
          <template #default="{ row }">
            <div class="staff-cell">
              <div class="staff-avatar" :style="{ background: row.avatar?.startsWith('http') ? 'transparent' : getAvatarColor(row.name) }">
                <img v-if="row.avatar?.startsWith('http')" :src="row.avatar" class="staff-avatar-img" />
                <span v-else>{{ row.name?.charAt(0) || '?' }}</span>
              </div>
              <div>
                <div class="staff-name" @click="editStaff(row.id)">{{ row.name }}</div>
                <div class="staff-online">
                  <span :class="['online-dot', onlineSet.has(row.id) ? 'on' : 'off']" />
                  {{ onlineSet.has(row.id) ? '在线' : '离线' }}
                </div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="账号" width="130">
          <template #default="{ row }">
            <span>{{ row.username }}</span>
          </template>
        </el-table-column>
        <el-table-column label="联系方式" min-width="180">
          <template #default="{ row }">
            <div class="staff-phone">{{ row.phone || '-' }}</div>
            <div class="staff-email">{{ row.email }}</div>
          </template>
        </el-table-column>
        <el-table-column label="所属角色" width="140">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" size="small" effect="plain">{{ roleDisplayName(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="department" label="部门" width="100" />
        <el-table-column label="最后登录" width="150">
          <template #default="{ row }">
            <span class="login-time">{{ row.lastLogin || '尚未登录' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="plain">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="row.role === 'super_admin' || row.role === '超级管理员'">
               <span class="staff-system-tag">系统账号</span>
             </template>
             <template v-else>
              <el-button v-if="store.hasPerm('staff', 'edit')" text size="small" type="primary" @click="editStaff(row.id)">编辑</el-button>
              <el-button
                v-if="store.hasPerm('staff', 'toggle')"
                text size="small"
                :type="row.status === 1 ? 'danger' : 'success'"
                @click="toggleStaff(row.id)">
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @change="loadStaffList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑员工对话框 -->
    <el-dialog
      v-model="staffDialogVisible"
      :title="editingId ? '编辑员工' : '添加员工'"
      width="560px"
      :close-on-click-modal="false"
      @close="resetStaffForm">
      <el-form ref="staffFormRef" :model="staffForm" :rules="staffFormRules" label-width="80px">
        <el-form-item label="登录账号" prop="username">
          <el-input v-model="staffForm.username" placeholder="请输入登录账号" :disabled="!!editingId" />
          <span v-if="editingId" class="form-hint">创建后不可修改</span>
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="staffForm.name" placeholder="请输入员工姓名" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="staffForm.phone" placeholder="请输入手机号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="staffForm.email" placeholder="请输入邮箱地址" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="所属角色">
              <el-select v-model="staffForm.role" style="width:100%">
                <el-option v-for="r in formRoleOptions" :key="r.value" :label="r.label" :value="r.value" />
              </el-select>
              <div class="form-tip">
                <el-icon><InfoFilled /></el-icon>
                角色权限可在 <router-link to="/roles" class="tip-link">角色管理</router-link> 中配置
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属部门">
              <el-select v-model="staffForm.department" style="width:100%">
                <el-option v-for="d in deptOptions" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="初始密码" v-if="!editingId">
          <div class="password-row">
            <el-input v-model="staffForm.password" placeholder="请设置登录密码" />
            <el-button size="default" @click="generatePassword">自动生成</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="staffDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveStaff" :loading="saving">{{ editingId ? '保存修改' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 员工管理页面
 * 支持员工的增删改查、按角色/状态筛选、启用/禁用切换
 * 数据通过 API 实时获取，在线状态通过 Redis 检测
 */
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { listStaff, createStaff, updateStaff, toggleStaffStatus, getOnlineAdminIds } from '../../api/staff'
import type { StaffInfo } from '../../api/staff'
import { listRoles } from '../../api/role'
import { useAppStore } from '../../store/app'

const store = useAppStore()

// ---------- 状态 ----------
const loading = ref(false)
const saving = ref(false)
const staffList = ref<StaffInfo[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const onlineSet = ref<Set<number>>(new Set())

// ---------- 角色映射 ----------
interface RoleOption { label: string; value: string }
const roleOptions = ref<RoleOption[]>([])
const roleMap = ref<Record<string, string>>({}) // slug → name

async function loadRoles() {
  try {
    const res = await listRoles()
    const items = res.items || []
    const options: RoleOption[] = []
    const map: Record<string, string> = {}
    for (const r of items) {
      const slug = r.slug || r.name
      const name = r.name || slug
      options.push({ label: name, value: slug })
      map[slug] = name
    }
    roleOptions.value = options
    roleMap.value = map
  } catch {
    // 加载失败时保持默认
  }
}

// 获取角色显示名
function roleDisplayName(slug: string): string {
  return roleMap.value[slug] || slug
}

// 新增/编辑时可选的角色（过滤掉超级管理员）
const formRoleOptions = computed(() => roleOptions.value.filter(r => r.value !== 'super_admin'))

const deptOptions = ['技术部', '运营部', '客服部', '数据部', '市场部', '人力部', '财务部']

// ---------- 筛选 ----------
const searchKeyword = ref('')
const roleFilter = ref('')
const statusFilter = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

/** 搜索输入防抖 */
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    loadStaffList()
  }, 300)
}

function applyFilters() {
  page.value = 1
  loadStaffList()
}

function resetFilters() {
  searchKeyword.value = ''
  roleFilter.value = ''
  statusFilter.value = ''
  page.value = 1
  loadStaffList()
}

// ---------- 数据加载 ----------
/** 加载员工列表 */
async function loadStaffList() {
  loading.value = true
  try {
    const params: { page?: number; page_size?: number; keyword?: string; status?: string; role?: string } = {
      page: page.value,
      page_size: pageSize.value
    }
    if (searchKeyword.value) params.keyword = searchKeyword.value
    if (statusFilter.value) params.status = statusFilter.value
    if (roleFilter.value) params.role = roleFilter.value
    const res = await listStaff(params)
    staffList.value = res.items || []
    total.value = res.total || 0
  } catch {
    ElMessage.error('获取员工列表失败')
  } finally {
    loading.value = false
  }
}

// ---------- 在线状态 ----------
let onlineTimer: ReturnType<typeof setInterval> | null = null

async function refreshOnlineStatus() {
  try {
    const res = await getOnlineAdminIds()
    onlineSet.value = new Set(res.ids || [])
  } catch {
    // API 失败时保持上次状态
  }
}

onMounted(() => {
  loadRoles()
  loadStaffList()
  refreshOnlineStatus()
  onlineTimer = setInterval(refreshOnlineStatus, 60000)
})

onUnmounted(() => {
  if (onlineTimer) clearInterval(onlineTimer)
})

// ---------- 统计卡片 ----------
const stats = computed(() => {
  const activeCount = staffList.value.filter(s => s.status === 1).length
  const onlineCount = staffList.value.filter(s => onlineSet.value.has(s.id)).length
  return [
    { label: '员工总数', value: total.value, unit: '人', color: 'blue', icon: 'UserFilled', change: 1, changeText: '含全部状态' },
    { label: '已启用', value: activeCount, unit: '人', color: 'green', icon: 'Clock', change: null, changeText: '正常状态' },
    { label: '在线中', value: onlineCount, unit: '人', color: 'orange', icon: 'Right', change: null, changeText: '当前在线' }
  ]
})

// ---------- 头像颜色 ----------
const avatarColors = ['#f59e0b', '#3b82f6', '#10b981', '#8b5cf6', '#ec4899', '#0d9488', '#4f46e5', '#db2777']

function getAvatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

// ---------- 角色 Tag 颜色 ----------
function getRoleTagType(role: string): string {
  const map: Record<string, string> = {
    '超级管理员': 'warning',
    '管理员': 'primary',
    '运营专员': 'success',
    '客服专员': 'info',
    '数据分析师': 'danger'
  }
  return map[role] || ''
}

// ---------- 员工表单 ----------
const staffDialogVisible = ref(false)
const editingId = ref<number | null>(null)
const staffFormRef = ref<FormInstance>()

const staffForm = reactive({
  username: '',
  name: '',
  phone: '',
  email: '',
  role: '',
  department: '',
  password: 'yuexuan123'
})

const staffFormRules: FormRules = {
  name: [{ required: true, message: '请输入员工姓名', trigger: 'blur' }],
  email: [{ required: true, message: '请输入邮箱地址', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

function openAddModal() {
  editingId.value = null
  resetStaffForm()
  staffDialogVisible.value = true
}

function resetStaffForm() {
  staffForm.username = ''
  staffForm.name = ''
  staffForm.phone = ''
  staffForm.email = ''
  staffForm.role = formRoleOptions.value[0]?.value || ''
  staffForm.department = ''
  staffForm.password = 'yuexuan123'
}

async function editStaff(id: number) {
  editingId.value = id
  const staff = staffList.value.find(s => s.id === id)
  if (!staff) return
  staffForm.username = staff.username
  staffForm.name = staff.name
  staffForm.phone = staff.phone
  staffForm.email = staff.email
  staffForm.role = staff.role
  staffForm.department = staff.department
  staffForm.password = ''
  staffDialogVisible.value = true
}

async function saveStaff() {
  const valid = await staffFormRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    if (editingId.value) {
      await updateStaff(editingId.value, {
        name: staffForm.name,
        role: staffForm.role,
        email: staffForm.email,
        phone: staffForm.phone || undefined,
        department: staffForm.department || undefined
      })
      ElMessage.success('员工信息已更新')
    } else {
      await createStaff({
        username: staffForm.username,
        password: staffForm.password,
        name: staffForm.name,
        role: staffForm.role,
        email: staffForm.email,
        phone: staffForm.phone || undefined,
        department: staffForm.department || undefined
      })
      ElMessage.success('员工已添加')
    }
    staffDialogVisible.value = false
    loadStaffList()
  } catch {
    ElMessage.error(editingId.value ? '更新员工信息失败' : '添加员工失败')
  } finally {
    saving.value = false
  }
}

function generatePassword() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let pwd = ''
  for (let i = 0; i < 10; i++) {
    pwd += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  staffForm.password = pwd
}

async function toggleStaff(id: number) {
  const staff = staffList.value.find(s => s.id === id)
  if (!staff) return
  const actionText = staff.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${actionText}员工「${staff.name}」吗？`,
      `确认${actionText}`,
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await toggleStaffStatus(id)
    ElMessage.success(`员工「${staff.name}」已${actionText}`)
    loadStaffList()
  } catch {
    // 取消操作不做处理
  }
}
</script>

<style scoped>
.staff-page {
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
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 16px;
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

.stat-change {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  margin-top: 4px;
}
.stat-change.up { color: #059669; }
.stat-change.down { color: var(--shop-destructive); }

/* 筛选栏 */
.filter-card {
  margin-bottom: 16px;
  border-radius: var(--shop-radius-lg);
}

.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-group label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

/* 表格卡片 */
.table-card {
  border-radius: var(--shop-radius-lg);
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}

.staff-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.staff-avatar {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
  overflow: hidden;
}

.staff-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.staff-name {
  font-weight: 600;
  color: var(--shop-primary);
  cursor: pointer;
  transition: color 0.2s;
}
.staff-name:hover {
  color: var(--shop-primary-dark);
}

.staff-online {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 2px;
}

.online-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  display: inline-block;
}
.online-dot.on { background: #10b981; }
.online-dot.off { background: #94a3b8; }

.staff-phone {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}
.staff-email {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.login-time {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.staff-system-tag {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

/* 表单 */
.password-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.form-tip {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.tip-link {
  color: var(--shop-primary);
  text-decoration: none;
  font-weight: 500;
}
.tip-link:hover {
  text-decoration: underline;
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr 1fr; }
  .page-header { flex-direction: column; align-items: flex-start; gap: 12px; }
  .filter-bar { flex-direction: column; align-items: stretch; }
  .filter-group { flex-direction: column; align-items: stretch; }
}
</style>
