<template>
  <!-- 员工管理页面 - 统计卡片 + 筛选栏 + 员工表格 + 新增/编辑对话框 -->
  <div class="staff-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">员工管理</h2>
        <span class="page-subtitle">管理所有后台账号与员工信息</span>
      </div>
      <el-button type="primary" @click="openAddModal">
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
          <el-input v-model="searchKeyword" placeholder="姓名 / 手机号 / 邮箱" clearable size="default" style="width:200px" @input="applyFilters" />
        </div>
        <div class="filter-group">
          <label>角色</label>
          <el-select v-model="roleFilter" placeholder="全部" clearable size="default" style="width:150px" @change="applyFilters">
            <el-option v-for="r in roleOptions" :key="r" :label="r" :value="r" />
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
      <el-table :data="filteredStaff" stripe style="width:100%">
        <el-table-column label="姓名" min-width="160">
          <template #default="{ row }">
            <div class="staff-cell">
              <div class="staff-avatar" :style="{ background: row.color }">{{ row.avatar }}</div>
              <div>
                <div class="staff-name" @click="editStaff(row.id)">{{ row.name }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="账号" min-width="180">
          <template #default="{ row }">
            <div class="staff-phone">{{ row.phone }}</div>
            <div class="staff-email">{{ row.email }}</div>
          </template>
        </el-table-column>
        <el-table-column label="所属角色" width="140">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" size="small" effect="plain">{{ row.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="dept" label="部门" width="100" />
        <el-table-column label="最后登录" width="150">
          <template #default="{ row }">
            <span class="login-time">{{ row.lastLogin }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small" effect="plain">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="editStaff(row.id)">编辑</el-button>
            <el-button
              text size="small"
              :type="row.status === 'active' ? 'danger' : 'success'"
              @click="toggleStaff(row.id)">
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑员工对话框 -->
    <el-dialog
      v-model="staffDialogVisible"
      :title="editingId ? '编辑员工' : '添加员工'"
      width="560px"
      :close-on-click-modal="false">
      <el-form ref="staffFormRef" :model="staffForm" :rules="staffFormRules" label-width="80px">
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
                <el-option v-for="r in roleOptions" :key="r" :label="r" :value="r" />
              </el-select>
              <div class="form-tip">
                <el-icon><InfoFilled /></el-icon>
                角色权限可在 <router-link to="/roles" class="tip-link">角色管理</router-link> 中配置
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属部门">
              <el-select v-model="staffForm.dept" style="width:100%">
                <el-option v-for="d in deptOptions" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="初始密码">
          <div class="password-row">
            <el-input v-model="staffForm.password" placeholder="请设置登录密码" />
            <el-button size="default" @click="generatePassword">自动生成</el-button>
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="staffForm.status" active-value="active" inactive-value="inactive" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="staffDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveStaff">{{ editingId ? '保存修改' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 员工管理页面
 * 支持员工的增删改查、按角色/状态筛选、启用/禁用切换
 */
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'

// ---------- 类型定义 ----------
interface Staff {
  id: number
  name: string
  phone: string
  email: string
  role: string
  dept: string
  lastLogin: string
  status: 'active' | 'inactive'
  avatar: string
  color: string
}

// ---------- 模拟数据 ----------
const staffData = reactive<Staff[]>([
  { id: 1, name: '张明', phone: '138****0001', email: 'admin@yuexuan.com', role: '超级管理员', dept: '技术部', lastLogin: '2025-07-14 08:30', status: 'active', avatar: '张', color: '#f59e0b' },
  { id: 2, name: '李华', phone: '138****0002', email: 'lihua@yuexuan.com', role: '运营专员', dept: '运营部', lastLogin: '2025-07-13 16:45', status: 'active', avatar: '李', color: '#3b82f6' },
  { id: 3, name: '王芳', phone: '138****0003', email: 'wangfang@yuexuan.com', role: '客服专员', dept: '客服部', lastLogin: '2025-07-13 14:20', status: 'active', avatar: '王', color: '#10b981' },
  { id: 4, name: '赵强', phone: '138****0004', email: 'zhaoqiang@yuexuan.com', role: '管理员', dept: '技术部', lastLogin: '2025-07-12 09:15', status: 'active', avatar: '赵', color: '#8b5cf6' },
  { id: 5, name: '刘洋', phone: '138****0005', email: 'liuyang@yuexuan.com', role: '运营专员', dept: '运营部', lastLogin: '2025-07-11 11:30', status: 'inactive', avatar: '刘', color: '#ec4899' },
  { id: 6, name: '陈丽', phone: '138****0006', email: 'chenli@yuexuan.com', role: '客服专员', dept: '客服部', lastLogin: '2025-07-10 17:00', status: 'active', avatar: '陈', color: '#0d9488' },
  { id: 7, name: '孙杰', phone: '138****0007', email: 'sunjie@yuexuan.com', role: '数据分析师', dept: '数据部', lastLogin: '2025-07-09 08:50', status: 'active', avatar: '孙', color: '#4f46e5' },
  { id: 8, name: '周婷', phone: '138****0008', email: 'zhouting@yuexuan.com', role: '运营专员', dept: '运营部', lastLogin: '2025-06-28 15:10', status: 'inactive', avatar: '周', color: '#db2777' }
])

// ---------- 下拉选项 ----------
const roleOptions = ['超级管理员', '管理员', '运营专员', '客服专员', '数据分析师', '自定义角色']
const deptOptions = ['技术部', '运营部', '客服部', '数据部', '市场部', '人力部', '财务部']

// ---------- 筛选 ----------
const searchKeyword = ref('')
const roleFilter = ref('')
const statusFilter = ref('')

const filteredStaff = computed(() => {
  return staffData.filter(s => {
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      if (!s.name.toLowerCase().includes(kw) && !s.phone.includes(kw) && !s.email.toLowerCase().includes(kw)) {
        return false
      }
    }
    if (roleFilter.value && s.role !== roleFilter.value) return false
    if (statusFilter.value && s.status !== statusFilter.value) return false
    return true
  })
})

function applyFilters() { /* 计算属性已自动响应 */ }
function resetFilters() {
  searchKeyword.value = ''
  roleFilter.value = ''
  statusFilter.value = ''
}

// ---------- 统计卡片 ----------
const stats = computed(() => [
  { label: '员工总数', value: staffData.length, unit: '人', color: 'blue', icon: 'UserFilled', change: 1, changeText: '较上月 +2' },
  { label: '在线中', value: staffData.filter(s => s.status === 'active').length, unit: '人', color: 'green', icon: 'Clock', change: null, changeText: '当前在线' },
  { label: '今日登录', value: '8', unit: '人', color: 'orange', icon: 'Right', change: 1, changeText: '昨日 6 人' }
])

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
  name: '',
  phone: '',
  email: '',
  role: '运营专员',
  dept: '运营部',
  password: 'yuexuan123',
  status: 'active' as 'active' | 'inactive'
})

const staffFormRules: FormRules = {
  name: [{ required: true, message: '请输入员工姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  email: [{ required: true, message: '请输入邮箱地址', trigger: 'blur' }]
}

function openAddModal() {
  editingId.value = null
  staffForm.name = ''
  staffForm.phone = ''
  staffForm.email = ''
  staffForm.role = '运营专员'
  staffForm.dept = '运营部'
  staffForm.password = 'yuexuan123'
  staffForm.status = 'active'
  staffDialogVisible.value = true
}

function editStaff(id: number) {
  const staff = staffData.find(s => s.id === id)
  if (!staff) return
  editingId.value = id
  staffForm.name = staff.name
  staffForm.phone = staff.phone
  staffForm.email = staff.email
  staffForm.role = staff.role
  staffForm.dept = staff.dept
  staffForm.password = '********'
  staffForm.status = staff.status
  staffDialogVisible.value = true
}

function saveStaff() {
  staffFormRef.value?.validate((valid) => {
    if (!valid) return
    if (editingId.value) {
      const staff = staffData.find(s => s.id === editingId.value)
      if (staff) {
        staff.name = staffForm.name
        staff.phone = staffForm.phone
        staff.email = staffForm.email
        staff.role = staffForm.role
        staff.dept = staffForm.dept
        staff.status = staffForm.status
      }
      ElMessage.success('员工信息已更新')
    } else {
      const colors = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#0d9488', '#4f46e5', '#db2777']
      const newStaff: Staff = {
        id: staffData.length > 0 ? staffData[staffData.length - 1].id + 1 : 1,
        name: staffForm.name,
        phone: staffForm.phone,
        email: staffForm.email,
        role: staffForm.role,
        dept: staffForm.dept,
        lastLogin: '尚未登录',
        status: staffForm.status,
        avatar: staffForm.name.charAt(0),
        color: colors[staffData.length % colors.length]
      }
      staffData.push(newStaff)
      ElMessage.success('员工已添加')
    }
    staffDialogVisible.value = false
  })
}

function generatePassword() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let pwd = ''
  for (let i = 0; i < 10; i++) {
    pwd += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  staffForm.password = pwd
}

function toggleStaff(id: number) {
  const staff = staffData.find(s => s.id === id)
  if (!staff) return
  const newStatus = staff.status === 'active' ? 'inactive' : 'active'
  const actionText = newStatus === 'active' ? '启用' : '禁用'
  ElMessageBox.confirm(
    `确定要${actionText}员工「${staff.name}」吗？`,
    `确认${actionText}`,
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    staff.status = newStatus
    ElMessage.success(`员工「${staff.name}」已${actionText}`)
  }).catch(() => {})
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
