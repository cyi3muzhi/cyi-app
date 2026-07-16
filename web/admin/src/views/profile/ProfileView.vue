<template>
  <!-- 个人信息页面 - 包含概览卡片、基本信息表单、安全设置、角色权限 -->
  <div class="profile">
    <!-- 面包屑导航 -->
    <div class="profile-breadcrumb">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item>系统设置</el-breadcrumb-item>
        <el-breadcrumb-item>个人信息</el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 个人概览卡片 - 渐变头部 -->
    <el-card class="profile-overview" shadow="never">
      <div class="overview-header">
        <div class="overlay-pattern" />
      </div>
      <div class="overview-body">
        <div class="avatar-section">
          <div class="avatar-circle" @click="handleChangeAvatar">
            <img v-if="avatarUrl" :src="avatarUrl" class="avatar-image" />
            <span v-else class="avatar-text">{{ getDefaultAvatarText() }}</span>
          </div>
          <el-button class="avatar-btn" size="small" plain @click="handleChangeAvatar">
            更换头像
          </el-button>
        </div>
        <div class="overview-info">
          <div class="info-top">
            <h2 class="user-name">{{ basicForm.realName }}</h2>
            <el-tag class="role-badge color-tag" size="small" :color="roleColor" effect="dark">{{ roleDisplayName }}</el-tag>
          </div>
          <p class="user-bio">{{ basicForm.bio }}</p>
        </div>
      </div>
      <input
        ref="avatarInputRef"
        type="file"
        accept="image/*"
        style="display: none"
        @change="handleAvatarChange"
      />
    </el-card>

    <!-- 元数据信息网格 -->
    <el-card class="meta-card" shadow="never">
      <el-descriptions :column="4" border size="small">
        <el-descriptions-item label="电子邮箱" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Message /></el-icon>
            {{ basicForm.email || '未设置' }}
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="手机号码" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Iphone /></el-icon>
            {{ basicForm.phone || '未设置' }}
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="注册日期" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Clock /></el-icon>
            {{ basicForm.createdAt || '未知' }}
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="最后登录" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Timer /></el-icon>
            {{ basicForm.lastLogin || '未知' }}
          </div>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 基本信息编辑 -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="section-title-row">
          <div class="section-title-left">
            <el-icon class="section-icon"><EditPen /></el-icon>
            <span>基本信息</span>
          </div>
        </div>
      </template>

      <el-form ref="basicFormRef" :rules="basicRules" :model="basicForm" label-width="100px" class="basic-form">
        <el-row :gutter="24">
          <el-col :xs="24" :md="12">
            <el-form-item label="真实姓名" prop="realName">
              <el-input v-model="basicForm.realName" placeholder="请输入真实姓名" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="手机号码" prop="phone">
              <el-input v-model="basicForm.phone" placeholder="请输入手机号码" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="电子邮箱" prop="email">
              <el-input v-model="basicForm.email" placeholder="请输入电子邮箱" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="所属部门">
              <el-select v-model="basicForm.department" placeholder="请选择所属部门" style="width:100%">
                <el-option v-for="d in deptOptions" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="担任职位">
              <el-input v-model="basicForm.position" placeholder="请输入担任职位" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="个人简介" prop="bio">
              <el-input
                v-model="basicForm.bio"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="请简要介绍自己"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <div class="form-actions">
          <el-button type="primary" @click="handleSaveBasic">保存修改</el-button>
          <el-button @click="handleResetBasic">重置</el-button>
        </div>
      </el-form>
    </el-card>

    <!-- 安全设置 -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="section-title-row">
          <div class="section-title-left">
            <el-icon class="section-icon"><Lock /></el-icon>
            <span>安全设置</span>
          </div>
        </div>
      </template>

      <!-- 密码 -->
      <div class="security-item">
        <div class="security-item-left">
          <div class="security-item-title">登录密码</div>
          <div class="security-item-desc">
            密码强度：
            <el-tag size="small" type="success" effect="plain" class="pwd-strength-tag">强</el-tag>
            <span class="pwd-hint">建议定期更换密码以保障账户安全</span>
          </div>
        </div>
        <el-button size="small" @click="handleChangePwd">修改密码</el-button>
      </div>

      <!-- 双因素认证 -->
      <div class="security-item">
        <div class="security-item-left">
          <div class="security-item-title">双因素认证 (2FA)</div>
          <div class="security-item-desc">开启后，登录时将额外要求动态验证码，提升账户安全性</div>
        </div>
        <el-switch v-model="security.twoFactor" @change="handleToggle2FA" />
      </div>

      <!-- 登录设备管理 -->
      <div class="security-item">
        <div class="security-item-left">
          <div class="security-item-title">登录设备管理</div>
          <div class="security-item-desc">当前有 2 台设备处于活跃登录状态</div>
        </div>
        <el-button size="small" @click="handleManageDevices">管理设备</el-button>
      </div>

      <!-- 最近活动 -->
      <div class="security-item no-border">
        <div class="security-item-left">
          <div class="security-item-title">最近活动</div>
          <div class="security-item-desc">查看近期的登录和操作记录</div>
        </div>
        <el-button size="small" @click="handleViewActivity">查看详情</el-button>
      </div>
    </el-card>

    <!-- 角色与权限 -->
    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="section-title-row">
          <div class="section-title-left">
            <el-icon class="section-icon"><UserFilled /></el-icon>
            <span>角色与权限</span>
          </div>
          <el-tag size="small" effect="dark" :color="roleColor" class="color-tag">{{ roleDisplayName }}</el-tag>
        </div>
      </template>

      <p class="perm-note">
        当前角色拥有以下系统权限，权限变更需联系上级管理员操作。
      </p>

      <div class="perm-grid">
        <div class="perm-card" v-for="mod in permissions" :key="mod.key">
          <div class="perm-card-header">
            <el-icon class="perm-card-icon"><Check /></el-icon>
            <span class="perm-card-title">{{ mod.title }}</span>
          </div>
          <div class="perm-card-items">
            <el-tag size="small" type="success" effect="plain" v-for="item in mod.items" :key="item.key">{{ item.label }}</el-tag>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/**
 * 个人信息页面组件
 * 展示用户概览、基本资料编辑、安全设置和角色权限管理
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Message, Iphone, Clock, Timer,
  EditPen, Lock, UserFilled, Check
} from '@element-plus/icons-vue'
import { useAppStore } from '../../store/app'
import { getAdminInfo, changeAvatar, updateAdminInfo } from '../../api/auth'
import { getRoleNameMap } from '../../api/staff'
import { getMyPermissions } from '../../api/auth'
import { listAllPermissions } from '../../api/role'

/** 角色名称映射 */
const roleNameMap = ref<Record<string, string>>({})
/** 角色颜色映射 */
const roleNameMapColor = ref<Record<string, string>>({})

/** 角色显示名 */
const roleDisplayName = computed(() => {
  const slug = store.adminInfo?.role || ''
  return roleNameMap.value[slug] || slug
})

/** 角色标签颜色 */
const roleColor = computed(() => {
  const slug = store.adminInfo?.role || ''
  return roleNameMapColor.value[slug] || '#409EFF'
})

/** 基本信息表单引用 */
const basicFormRef = ref(null)

/** 头像文件输入引用 */
const avatarInputRef = ref(null)

/** 应用状态 */
const store = useAppStore()

/** 头像 URL */
const avatarUrl = ref('')

/** 部门选项 */
const deptOptions = ['技术部', '运营部', '客服部', '数据部', '市场部', '人力部', '财务部']

/** 基本信息表单数据 */
const basicForm = reactive({
  realName: '',
  phone: '',
  email: '',
  department: '',
  position: '',
  bio: '',
  createdAt: '',
  lastLogin: ''
})

/**
 * 页面加载时获取管理员信息
 */
async function loadAdminInfo() {
  try {
    const res = await getAdminInfo()
    console.log('getAdminInfo response:', res)
    if (res.admin) {
      basicForm.realName = res.admin.name || ''
      basicForm.email = res.admin.email || ''
      basicForm.phone = res.admin.phone || ''
      basicForm.department = res.admin.department || ''
      basicForm.position = res.admin.position || ''
      basicForm.bio = res.admin.bio || ''
      basicForm.createdAt = res.admin?.createdAt || ''
      basicForm.lastLogin = res.admin?.lastLogin || ''
      avatarUrl.value = res.admin.avatar || ''

      // 同步更新 store，使顶部栏头像保持最新
      if (res.admin.avatar && res.admin.avatar !== store.adminInfo?.avatar) {
        store.adminInfo = { ...store.adminInfo!, avatar: res.admin.avatar } as any
        localStorage.setItem('admin_info', JSON.stringify(store.adminInfo))
      }
    }
  } catch (error) {
    console.error('获取管理员信息失败:', error)
  }
}

onMounted(async () => {
  loadAdminInfo()
  // 加载角色显示名称映射
  try {
    const res = await getRoleNameMap()
    roleNameMap.value = res.items || {}
  } catch {
    // 忽略
  }
  // 加载角色颜色映射
  try {
    const roleRes = await import('../../api/role').then(m => m.listRoles())
    const colorMap: Record<string, string> = {}
    for (const r of roleRes.items || []) {
      if (r.color) colorMap[r.slug || r.name] = r.color
    }
    roleNameMapColor.value = colorMap
  } catch {
    // 忽略
  }
  // 加载当前角色权限
  await loadMyPermissions()
})

/**
 * 获取默认头像文字（名字的最后一个字）
 */
function getDefaultAvatarText() {
  const name = basicForm.realName || store.adminInfo?.name || ''
  if (name) {
    return name.slice(-1)
  }
  return '管'
}

/** 基本信息验证规则 */
const basicRules = {
  realName: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号码', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号码格式不正确', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入电子邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ]
}

/** 安全设置开关 */
const security = reactive({
  twoFactor: false
})

/** 角色权限列表 */
interface PermModuleInfo {
  key: string
  title: string
  items: { key: string; label: string }[]
}

/** 当前角色拥有的权限模块列表 */
const permissions = ref<PermModuleInfo[]>([])

/** 加载当前角色权限 */
async function loadMyPermissions() {
  try {
    const [permRes, modRes] = await Promise.all([
      getMyPermissions(),
      listAllPermissions()
    ])
    const userPerms = permRes.permissions || {}
    const moduleDefs = modRes.modules || []
    const list: PermModuleInfo[] = []
    for (const mod of moduleDefs) {
      const modPerms = userPerms[mod.key]?.values
      if (!modPerms) continue
      // 只保留有权限的具体项
      const enabledItems = mod.items.filter(item => modPerms[item.key])
      if (enabledItems.length === 0) continue
      list.push({ key: mod.key, title: mod.title, items: enabledItems })
    }
    permissions.value = list
  } catch {
    // 加载失败不显示
  }
}

/**
 * 处理更换头像 - 触发文件选择
 */
function handleChangeAvatar() {
  avatarInputRef.value?.click()
}

/**
 * 处理头像文件选择
 * @param {Event} event 文件选择事件
 */
async function handleAvatarChange(event) {
  const file = event.target.files?.[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }

  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('图片大小不能超过 2MB')
    return
  }

  try {
    const loadingMsg = ElMessage.info('正在上传头像...')

    const base64Data = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = (e) => {
        resolve((e.target!.result as string).split(',')[1])
      }
      reader.onerror = reject
      reader.readAsDataURL(file)
    })

    const avatarResponse = await changeAvatar(base64Data)

    loadingMsg.close()

    // Kratos v2.9.2 的 protojson 使用 camelCase，avatar_url 在 JSON 响应中为 avatarUrl
    const newAvatarUrl = avatarResponse.avatarUrl || avatarResponse.avatar_url
    if (avatarResponse.success && newAvatarUrl) {
      avatarUrl.value = newAvatarUrl
      // 同步更新 store 和 localStorage，使顶部栏头像立即更新
      store.adminInfo = {
        ...store.adminInfo!,
        avatar: newAvatarUrl
      } as any
      localStorage.setItem('admin_info', JSON.stringify(store.adminInfo))
      ElMessage.success('头像上传成功')
    } else {
      console.error('头像上传响应异常:', avatarResponse)
      ElMessage.error('头像保存失败')
    }
  } catch (error) {
    console.error('头像上传失败:', error)
    ElMessage.error('头像上传失败，请重试')
  }

  event.target.value = ''
}

/**
 * 保存基本信息
 */
async function handleSaveBasic() {
  basicFormRef.value.validate(async (valid) => {
    if (!valid) return

    try {
      const res = await updateAdminInfo({
        name: basicForm.realName,
        email: basicForm.email,
        phone: basicForm.phone,
        department: basicForm.department,
        position: basicForm.position,
        bio: basicForm.bio
      })

      if (res.success) {
        // 同步更新 store 中的姓名
        store.adminInfo = { ...store.adminInfo!, name: basicForm.realName } as any
        localStorage.setItem('admin_info', JSON.stringify(store.adminInfo))
        ElMessage.success('基本信息已保存')
      } else {
        ElMessage.error('保存失败')
      }
    } catch (error) {
      console.error('保存基本信息失败:', error)
      ElMessage.error('保存失败，请重试')
    }
  })
}

/**
 * 重置基本信息表单（重新加载服务器数据）
 */
async function handleResetBasic() {
  try {
    await loadAdminInfo()
    ElMessage.success('表单已重置')
  } catch {
    ElMessage.error('重置失败')
  }
}

/**
 * 修改密码（模拟）
 */
function handleChangePwd() {
  ElMessageBox.prompt('请输入当前密码以验证身份', '修改密码', {
    confirmButtonText: '下一步',
    cancelButtonText: '取消',
    inputType: 'password',
    inputPlaceholder: '请输入当前密码'
  }).then(() => {
    ElMessageBox.alert('密码修改链接已发送至您的绑定邮箱，请查收并按指引操作。', '提示', {
      confirmButtonText: '我知道了',
      type: 'success'
    })
  }).catch(() => {})
}

/**
 * 切换双因素认证（模拟）
 * @param {boolean} val 开关值
 */
function handleToggle2FA(val) {
  ElMessage.success(val ? '双因素认证已开启' : '双因素认证已关闭')
}

/**
 * 管理登录设备（模拟）
 */
function handleManageDevices() {
  ElMessage.info('登录设备管理（模拟）')
}

/**
 * 查看最近活动（模拟）
 */
function handleViewActivity() {
  ElMessage.info('最近活动记录（模拟）')
}
</script>

<style scoped>
/* 自定义颜色标签去掉 Element Plus 默认蓝色边框 */
.color-tag {
  border: none !important;
}
/* ---- 页面容器 ---- */
.profile {
  width: 100%;
}

/* ---- 面包屑 ---- */
.profile-breadcrumb {
  margin-bottom: 1.25rem;
}

/* ---- 个人概览卡片 ---- */
.profile-overview {
  position: relative;
  overflow: hidden;
  margin-bottom: 1.25rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.overview-header {
  height: 140px;
  background: linear-gradient(135deg, #4f6ef7 0%, #7c5cfc 50%, #a855f7 100%);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  pointer-events: none;
}

.overlay-pattern {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0.1;
  background-image:
    radial-gradient(circle at 20% 50%, #fff 1px, transparent 1px),
    radial-gradient(circle at 80% 50%, #fff 1px, transparent 1px);
  background-size: 40px 40px;
}

.overview-body {
  display: flex;
  align-items: flex-start;
  gap: 1.5rem;
  padding: 80px 1.5rem 1.5rem;
  position: relative;
  z-index: 1;
}

/* 头像 */
.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.625rem;
  flex-shrink: 0;
}

.avatar-circle {
  width: 96px;
  height: 96px;
  border-radius: 999px;
  background: linear-gradient(135deg, #4f6ef7, #a855f7);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 4px solid #fff;
  box-shadow: 0 4px 16px rgba(79, 110, 247, 0.25);
  cursor: pointer;
  transition: transform 0.2s ease;
}

.avatar-circle:hover {
  transform: scale(1.05);
}

.avatar-image {
  width: 100%;
  height: 100%;
  border-radius: 999px;
  object-fit: cover;
}

.avatar-text {
  font-size: 2rem;
  font-weight: 700;
  color: #fff;
  user-select: none;
}

.avatar-btn {
  font-size: 0.75rem;
  white-space: nowrap;
}

/* 概览信息 */
.overview-info {
  padding-top: 2.5rem;
  flex: 1;
  min-width: 0;
}

.info-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.user-name {
  font-size: 1.375rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0;
  line-height: 1.3;
}

.role-badge {
  font-size: 0.75rem;
}

.user-bio {
  margin-top: 0.5rem;
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ---- 元数据信息卡片 ---- */
.meta-card {
  margin-bottom: 1.25rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.meta-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
}

.meta-icon {
  color: var(--shop-muted-foreground);
  font-size: 14px;
}

.meta-card :deep(.el-descriptions__title) {
  font-weight: 600;
}

/* ---- 相关信息卡片 ---- */
.section-card {
  margin-bottom: 1.25rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-icon {
  color: var(--shop-primary);
  font-size: 16px;
}

/* ---- 基本信息表单 ---- */
.basic-form {
  max-width: 820px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--shop-border);
}

/* ---- 安全设置 ---- */
.security-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 0;
  border-bottom: 1px solid var(--shop-border);
}

.security-item.no-border {
  border-bottom: none;
}

.security-item-left {
  flex: 1;
  min-width: 0;
  padding-right: 2rem;
}

.security-item-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
  margin-bottom: 4px;
}

.security-item-desc {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  line-height: 1.4;
}

.pwd-strength-tag {
  margin-left: 4px;
}

.pwd-hint {
  margin-left: 8px;
}

/* ---- 角色与权限 ---- */
.perm-note {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--shop-border);
}

.perm-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.perm-card {
  padding: 1rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius);
  transition: border-color 0.2s ease;
}

.perm-card:hover {
  border-color: var(--shop-primary);
}

.perm-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.perm-card-icon {
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: #e8f5e9;
  color: #2e7d32;
  display: flex;
  align-items: center;
  justify-content: center;
}

.perm-card-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.perm-card-items {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .profile {
    max-width: 100%;
  }

  .perm-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .overview-header {
    height: 120px;
  }

  .overview-body {
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 60px 1rem 1rem;
    gap: 1rem;
  }

  .info-top {
    justify-content: center;
  }

  .overview-info {
    padding-top: 0;
  }

  .basic-form {
    max-width: 100%;
  }

  .meta-card :deep(.el-descriptions__body) {
    display: block;
  }

  .meta-card :deep(.el-descriptions__table) {
    width: 100%;
  }

  .meta-card :deep(.el-descriptions__row) {
    display: block;
  }

  .meta-card :deep(.el-descriptions__cell) {
    display: block;
    width: 100%;
    margin-bottom: 0.5rem;
  }

  .security-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .security-item-left {
    padding-right: 0;
  }
}

@media (max-width: 480px) {
  .overview-header {
    height: 100px;
  }

  .overview-body {
    padding: 45px 0.75rem 1rem;
    gap: 0.75rem;
  }

  .avatar-circle {
    width: 72px;
    height: 72px;
  }

  .avatar-text {
    font-size: 1.5rem;
  }

  .user-name {
    font-size: 1.125rem;
  }

  .role-badge {
    font-size: 0.6875rem;
    padding: 0 8px;
    line-height: 20px;
  }

  .user-bio {
    font-size: 0.75rem;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions .el-button {
    width: 100%;
  }
}
</style>
