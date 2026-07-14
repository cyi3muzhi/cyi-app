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
            <el-tag class="role-badge" size="small" color="#fff" effect="dark">超级管理员</el-tag>
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
            zhangming@yuexuan.com
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="手机号码" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Iphone /></el-icon>
            138****8888
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="注册日期" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Clock /></el-icon>
            2024-06-15
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="最后登录" width="150px">
          <div class="meta-cell">
            <el-icon class="meta-icon"><Timer /></el-icon>
            2026-07-14 09:32:18
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
      <el-form
        ref="basicFormRef"
        :model="basicForm"
        :rules="basicRules"
        label-position="top"
        class="basic-form"
      >
        <el-row :gutter="24">
          <el-col :span="12">
            <el-form-item label="真实姓名" prop="realName">
              <el-input v-model="basicForm.realName" placeholder="请输入真实姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号码" prop="phone">
              <el-input v-model="basicForm.phone" placeholder="请输入手机号码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电子邮箱" prop="email">
              <el-input v-model="basicForm.email" placeholder="请输入电子邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属部门" prop="department">
              <el-input v-model="basicForm.department" placeholder="请输入所属部门" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="职位" prop="position">
              <el-input v-model="basicForm.position" placeholder="请输入职位" />
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
          <el-tag size="small" effect="dark" color="var(--shop-primary)">超级管理员</el-tag>
        </div>
      </template>

      <p class="perm-note">
        当前角色拥有以下系统权限，权限变更需联系上级管理员操作。
      </p>

      <div class="perm-grid">
        <div class="perm-item" v-for="perm in permissions" :key="perm.label">
          <div class="perm-check">
            <svg
              class="perm-check-icon"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M20 6 9 17l-5-5" />
            </svg>
          </div>
          <div class="perm-info">
            <span class="perm-label">{{ perm.label }}</span>
            <span class="perm-desc">{{ perm.desc }}</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
/**
 * 个人信息页面组件
 * 展示用户概览、基本资料编辑、安全设置和角色权限管理
 */
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Message, Iphone, Clock, Timer,
  EditPen, Lock, UserFilled
} from '@element-plus/icons-vue'
import { useAppStore } from '../../store/app'
import { getAdminInfo, changeAvatar } from '../../api/auth'
import request from '../../api/index'

/** 基本信息表单引用 */
const basicFormRef = ref(null)

/** 头像文件输入引用 */
const avatarInputRef = ref(null)

/** 应用状态 */
const store = useAppStore()

/** 头像 URL */
const avatarUrl = ref('')

/** 基本信息表单数据 */
const basicForm = reactive({
  realName: '',
  phone: '',
  email: '',
  department: '',
  position: '',
  bio: ''
})

/**
 * 页面加载时获取管理员信息
 */
async function loadAdminInfo() {
  try {
    const res = await getAdminInfo()
    if (res.admin) {
      basicForm.realName = res.admin.name || ''
      basicForm.email = res.admin.email || ''
      basicForm.phone = res.admin.phone || ''
      basicForm.bio = res.admin.bio || ''
      avatarUrl.value = res.admin.avatar || ''
    }
  } catch (error) {
    console.error('获取管理员信息失败:', error)
  }
}

onMounted(() => {
  loadAdminInfo()
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
const permissions = [
  { label: '商品管理', desc: '商品的创建、编辑、上下架及分类管理' },
  { label: '订单管理', desc: '订单的查询、编辑、退款及物流跟踪' },
  { label: '用户管理', desc: '用户信息的查看、编辑及账户状态管理' },
  { label: '系统设置', desc: '系统基本配置、安全策略及通知管理' },
  { label: '数据查看', desc: '运营数据、财务报表的查看与导出' }
]

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
    ElMessage.info('正在上传头像...')

    const reader = new FileReader()
    reader.onload = async (e) => {
      const base64Data = e.target.result.split(',')[1]

      const avatarResponse = await changeAvatar(base64Data)

      if (avatarResponse.success) {
        avatarUrl.value = avatarResponse.avatar_url
        ElMessage.success('头像上传成功')
      } else {
        ElMessage.error('头像保存失败')
      }
    }
    reader.readAsDataURL(file)
  } catch (error) {
    console.error('头像上传失败:', error)
    ElMessage.error('头像上传失败，请重试')
  }

  event.target.value = ''
}

/**
 * 保存基本信息（模拟）
 */
function handleSaveBasic() {
  basicFormRef.value.validate((valid) => {
    if (!valid) return
    ElMessage.success('基本信息已保存')
  })
}

/**
 * 重置基本信息表单
 */
function handleResetBasic() {
  basicForm.realName = '张明'
  basicForm.phone = '138****8888'
  basicForm.email = 'zhangming@yuexuan.com'
  basicForm.department = '技术研发部'
  basicForm.position = '技术总监'
  basicForm.bio = '全栈工程师，专注于电商平台架构设计与团队管理。拥有十余年互联网行业经验，擅长高并发系统设计与微服务架构。'
  ElMessage.success('表单已重置')
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
  font-weight: 600;
  border: 1px solid var(--shop-primary);
  background: var(--shop-primary) !important;
  color: #fff !important;
  border-radius: 999px;
  padding: 0 10px;
  line-height: 22px;
}

.user-bio {
  margin: 0.375rem 0 0;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ---- 元数据卡片 ---- */
.meta-card {
  margin-bottom: 1.25rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.meta-cell {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.meta-icon {
  font-size: 0.9375rem;
  color: var(--shop-primary);
}

/* ---- 通用分区卡片 ---- */
.section-card {
  margin-bottom: 1.25rem;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.section-title-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-icon {
  font-size: 1rem;
  color: var(--shop-primary);
}

/* ---- 基本信息表单 ---- */
.basic-form {
  width: 100%;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  padding-top: 0.25rem;
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
  padding-right: 1rem;
}

.security-item-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
  margin-bottom: 0.25rem;
}

.security-item-desc {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pwd-strength-tag {
  font-weight: 600;
}

.pwd-hint {
  color: var(--shop-muted-foreground);
}

/* ---- 角色与权限 ---- */
.perm-note {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin: 0 0 1rem;
  line-height: 1.5;
}

.perm-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.perm-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-md);
  transition: background 0.2s ease;
}

.perm-item:hover {
  background: hsl(220, 14%, 92%);
}

.perm-check {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  background: var(--shop-success);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 1px;
}

.perm-check-icon {
  color: #fff;
  stroke-width: 3;
}

.perm-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.perm-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.perm-desc {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  line-height: 1.4;
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
