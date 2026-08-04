<template>
  <!-- ============================================================ -->
  <!-- 1. 页面标题栏                                                  -->
  <!-- ============================================================ -->
  <div class="page-header">
    <div class="container-shop">
      <div class="page-header-inner">
        <router-link to="/user" class="back-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </router-link>
        <h1 class="page-title">编辑资料</h1>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 2. 头像 Hero 卡片（渐变色背景 + 可编辑头像）                     -->
  <!-- ============================================================ -->
  <section class="hero-section">
    <div class="container-shop">
      <div class="hero-card">
        <!-- 装饰圆 -->
        <div class="hero-decoration"></div>

        <div class="hero-content">
          <!-- 头像（点击更换） -->
          <div class="avatar-block">
            <div class="avatar-wrap" @click="triggerUpload">
              <!-- 头像图片/占位 -->
              <img
                v-if="avatarDisplay"
                :src="avatarDisplay"
                alt="用户头像"
                class="avatar-img"
                @error="onAvatarError"
              />
              <div v-else class="avatar-placeholder">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="12" cy="7" r="4"/>
                </svg>
              </div>

              <!-- 悬停遮罩：更换头像 -->
              <div v-if="!uploading" class="avatar-mask">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/>
                  <circle cx="12" cy="13" r="4"/>
                </svg>
                <span>更换头像</span>
              </div>

              <!-- 上传中 loading -->
              <div v-else class="avatar-loading">
                <span class="spinner"></span>
              </div>

              <!-- 右下角小相机角标 -->
              <div v-if="!uploading" class="avatar-badge">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/>
                  <circle cx="12" cy="13" r="4"/>
                </svg>
              </div>
            </div>

            <!-- 隐藏的文件选择框 -->
            <input ref="fileInput" type="file" accept="image/*" class="hidden-input" @change="onFileChange" />

            <p class="avatar-tip">点击头像即可更换，支持 JPG / PNG / WebP，不超过 5MB</p>
          </div>

          <!-- 昵称 + 会员等级 -->
          <div class="hero-info">
            <h2 class="hero-name">{{ form.nickname || '未设置昵称' }}</h2>
            <span class="member-badge">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
              </svg>
              {{ memberLevelLabel }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ============================================================ -->
  <!-- 3. 资料编辑表单                                                -->
  <!-- ============================================================ -->
  <section class="form-section">
    <div class="container-shop">
      <div class="form-card">
        <div v-if="loading" class="state-wrap">
          <span class="spinner spinner-lg"></span>
          <p class="state-text">正在加载资料...</p>
        </div>

        <template v-else>
          <!-- 昵称 -->
          <div class="form-group">
            <label class="form-label">
              <svg class="label-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
              昵称
            </label>
            <input
              v-model="form.nickname"
              class="form-input"
              type="text"
              maxlength="24"
              placeholder="请输入昵称"
            />
            <p class="field-hint">昵称将展示在商品评价和社区中</p>
          </div>

          <!-- 邮箱 -->
          <div class="form-group">
            <label class="form-label">
              <svg class="label-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                <polyline points="22,6 12,13 2,6"/>
              </svg>
              邮箱
            </label>
            <input
              v-model="form.email"
              class="form-input"
              type="email"
              placeholder="请输入邮箱（用于接收通知）"
            />
          </div>

          <!-- 性别 -->
          <div class="form-group">
            <label class="form-label">
              <svg class="label-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <circle cx="12" cy="12" r="6"/>
                <circle cx="12" cy="12" r="2"/>
              </svg>
              性别
            </label>
            <div class="gender-options">
              <label
                v-for="g in genderOptions"
                :key="g.value"
                class="gender-item"
                :class="{ 'gender-active': form.gender === g.value }"
              >
                <input v-model="form.gender" type="radio" name="gender" :value="g.value" />
                <span class="gender-icon" v-html="g.icon"></span>
                <span>{{ g.label }}</span>
              </label>
            </div>
          </div>

          <!-- 生日 -->
          <div class="form-group">
            <label class="form-label">
              <svg class="label-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/>
                <line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
              生日
            </label>
            <div class="birthday-selects">
              <select v-model="birthYear" class="form-select" @change="onBirthdaySelect">
                <option value="">年</option>
                <option v-for="y in years" :key="y" :value="y">{{ y }} 年</option>
              </select>
              <select v-model="birthMonth" class="form-select" @change="onBirthdaySelect">
                <option value="">月</option>
                <option v-for="m in 12" :key="m" :value="m">{{ m }} 月</option>
              </select>
              <select v-model="birthDay" class="form-select" @change="onBirthdayDay">
                <option value="">日</option>
                <option v-for="d in dayCount" :key="d" :value="d">{{ d }} 日</option>
              </select>
            </div>
            <p class="field-hint">生日仅用于会员关怀，不会公开展示</p>
          </div>

          <!-- 错误提示 -->
          <transition name="fade">
            <div v-if="formError" class="form-error">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              {{ formError }}
            </div>
          </transition>

          <!-- 操作按钮 -->
          <div class="form-actions">
            <button class="save-btn" :disabled="saving || uploading" @click="onSave">
              <span v-if="saving" class="spinner spinner-sm"></span>
              {{ saving ? '保存中...' : '保存修改' }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
/**
 * 编辑资料页面
 * 支持修改头像/昵称/邮箱/性别/生日
 * 头像通过 OSS 预签名 URL 直传（接口：GET /user/v1/upload/token → PUT 直传）
 * 资料保存：PUT /user/v1/update
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getMyInfo, updateUserInfo } from '@/api/user'
import { uploadFile } from '@/api/upload'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const formError = ref('')
const form = ref({ nickname: '', email: '', gender: 0, birthday: '', avatar: '' })

/** 本地头像预览 URL（上传中/上传后未保存时优先展示） */
const avatarPreview = ref('')

/** 头像最终展示：优先本地预览，其次已保存的头像 URL */
const avatarDisplay = computed(() => avatarPreview.value || form.value.avatar)

/** 隐藏文件输入框引用 */
const fileInput = ref<HTMLInputElement>()

/** 会员等级标签映射 */
const memberLevelLabel = ref('普通会员')

// ---- 生日（年/月/日三列下拉） ----
/** 可选年份：当前年 ~ 1940 */
const years = computed(() => {
  const current = new Date().getFullYear()
  const list: number[] = []
  for (let y = current; y >= 1940; y--) list.push(y)
  return list
})

/** 年/月/日下拉选中值（空字符串表示未选择） */
const birthYear = ref('')
const birthMonth = ref('')
const birthDay = ref('')

/** 当月天数（随年月联动，处理 2 月、大小月） */
const dayCount = computed(() => {
  const y = Number(birthYear.value)
  const m = Number(birthMonth.value)
  if (!y || !m) return 31
  return new Date(y, m, 0).getDate()
})

/**
 * 年/月变化时：修正超出当月的日期并同步表单值
 */
function onBirthdaySelect() {
  if (birthDay.value && Number(birthDay.value) > dayCount.value) {
    birthDay.value = String(dayCount.value)
  }
  updateBirthday()
}

/**
 * 日变化时：同步表单值
 */
function onBirthdayDay() {
  updateBirthday()
}

/**
 * 将年/月/日拼装为 YYYY-MM-DD 存入表单（未选全则存空串）
 */
function updateBirthday() {
  if (birthYear.value && birthMonth.value && birthDay.value) {
    form.value.birthday = `${birthYear.value}-${String(birthMonth.value).padStart(2, '0')}-${String(birthDay.value).padStart(2, '0')}`
  } else {
    form.value.birthday = ''
  }
}

/** 性别选项（含图标） */
const genderOptions = [
  {
    value: 1,
    label: '男',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="7"/><line x1="12" y1="9" x2="12" y2="15"/><line x1="9" y1="12" x2="15" y2="12"/><line x1="15" y1="5" x2="19" y2="1"/><line x1="17" y1="3" x2="21" y2="7"/></svg>'
  },
  {
    value: 2,
    label: '女',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="5" r="3.5"/><path d="M12 8.5V20"/><path d="M8 16h8"/></svg>'
  },
  {
    value: 0,
    label: '保密',
    icon: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>'
  }
]

/**
 * 触发头像文件选择
 */
function triggerUpload() {
  if (uploading.value) return
  fileInput.value?.click()
}

/**
 * 头像图片加载失败时回退为占位图
 */
function onAvatarError() {
  form.value.avatar = ''
  avatarPreview.value = ''
}

/**
 * 选择头像文件：本地预览 + 直传 OSS
 * @param e - change 事件
 */
async function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  // 清空 input 值，允许重复选择同一文件
  input.value = ''
  if (!file) return

  // 校验文件类型
  if (!file.type.startsWith('image/')) {
    formError.value = '请选择图片文件（JPG / PNG / WebP）'
    return
  }
  // 校验文件大小（5MB 上限）
  if (file.size > 5 * 1024 * 1024) {
    formError.value = '图片大小不能超过 5MB'
    return
  }

  // 本地预览（上传中即时反馈）
  if (avatarPreview.value) URL.revokeObjectURL(avatarPreview.value)
  avatarPreview.value = URL.createObjectURL(file)

  // 直传 OSS
  uploading.value = true
  formError.value = ''
  try {
    const url = await uploadFile(file, 'avatar')
    form.value.avatar = url
    avatarPreview.value = ''
  } catch (err: any) {
    formError.value = err?.message || '头像上传失败，请重试'
    avatarPreview.value = ''
  } finally {
    uploading.value = false
  }
}

// 页面加载时回填当前用户信息
onMounted(async () => {
  try {
    const user = await getMyInfo()
    form.value = {
      nickname: user.nickname || '',
      email: user.email || '',
      gender: user.gender ?? 0,
      birthday: user.birthday || '',
      avatar: user.avatar || ''
    }
    // 回填生日（YYYY-MM-DD → 年/月/日下拉）
    if (user.birthday) {
      const [y, m, d] = user.birthday.split('-')
      birthYear.value = y || ''
      birthMonth.value = m ? String(Number(m)) : ''
      birthDay.value = d ? String(Number(d)) : ''
    }
    // 同步会员等级标签
    const levelMap: Record<string, string> = {
      regular: '普通会员',
      bronze: '青铜会员',
      silver: '白银会员',
      gold: '黄金会员',
      platinum: '铂金会员',
      diamond: '钻石会员'
    }
    memberLevelLabel.value = levelMap[user.level || 'regular'] || '普通会员'
  } catch {
    formError.value = '加载用户信息失败'
  } finally {
    loading.value = false
  }
})

/**
 * 保存编辑资料（成功后返回个人中心并同步全局用户状态）
 */
async function onSave() {
  const nickname = form.value.nickname.trim()
  if (nickname.length < 2 || nickname.length > 24) {
    formError.value = '昵称长度需为 2-24 个字符'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    // 未设置头像时不传 avatar 字段（避免空串触发后端 uri 校验）
    const payload: { nickname: string; avatar?: string; email: string; gender: number; birthday: string } = {
      nickname,
      email: form.value.email.trim(),
      gender: form.value.gender,
      birthday: form.value.birthday
    }
    if (form.value.avatar) {
      payload.avatar = form.value.avatar
    }
    await updateUserInfo(payload)
    // 同步全局用户状态（昵称/头像）
    userStore.fetchUserInfo()
    router.push('/user')
  } catch (e: any) {
    formError.value = e?.response?.data?.message || e?.message || '保存失败，请重试'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
/* ============================================
   编辑资料页面样式
   ============================================ */

/* ---- 页面标题 ---- */
.page-header {
  padding: 24px 0 0;
}

.page-header-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-foreground);
  transition: background 0.2s, transform 0.2s;
}

.back-link:hover {
  background: var(--shop-border);
  transform: translateX(-2px);
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

/* ---- Hero 卡片 ---- */
.hero-section {
  padding: 24px 0 0;
}

.hero-card {
  position: relative;
  overflow: hidden;
  border-radius: var(--shop-radius-xl);
  border: 1px solid var(--shop-border);
  background: linear-gradient(135deg, var(--shop-primary) 0%, hsl(217, 91%, 62%) 50%, hsl(217, 91%, 72%) 100%);
}

.hero-decoration {
  pointer-events: none;
  position: absolute;
  right: -64px;
  top: -64px;
  width: 256px;
  height: 256px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
}

.hero-decoration::after {
  content: '';
  position: absolute;
  bottom: -80px;
  left: -160px;
  width: 320px;
  height: 320px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}

.hero-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 36px 24px;
}

@media (min-width: 768px) {
  .hero-content {
    flex-direction: row;
    justify-content: center;
    gap: 36px;
    padding: 40px;
  }
}

/* ---- 头像编辑 ---- */
.avatar-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.avatar-wrap {
  position: relative;
  width: 104px;
  height: 104px;
  border-radius: 50%;
  border: 4px solid rgba(255, 255, 255, 0.8);
  background: linear-gradient(135deg, #fff, hsl(217, 91%, 90%));
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.18);
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.avatar-wrap:hover {
  transform: scale(1.04);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.24);
}

@media (min-width: 768px) {
  .avatar-wrap {
    width: 120px;
    height: 120px;
  }
}

.avatar-img,
.avatar-placeholder {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 悬停遮罩 */
.avatar-mask {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  background: rgba(15, 23, 42, 0.55);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 500;
  opacity: 0;
  transition: opacity 0.25s;
}

.avatar-wrap:hover .avatar-mask {
  opacity: 1;
}

/* 右下角相机角标（常驻，提示可编辑） */
.avatar-badge {
  position: absolute;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: 3px solid #fff;
  background: var(--shop-primary);
  color: #fff;
  transition: transform 0.2s;
}

.avatar-wrap:hover .avatar-badge {
  transform: scale(1.12);
}

/* 上传中 loading 遮罩 */
.avatar-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.45);
}

.spinner {
  display: inline-block;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.35);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border-width: 2px;
  margin-right: 8px;
  vertical-align: middle;
}

.spinner-lg {
  width: 32px;
  height: 32px;
  border-width: 3px;
  border-color: var(--shop-border);
  border-top-color: var(--shop-primary);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.hidden-input {
  display: none;
}

.avatar-tip {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.85);
  text-align: center;
}

/* ---- Hero 信息 ---- */
.hero-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  text-align: center;
}

@media (min-width: 768px) {
  .hero-info {
    align-items: flex-start;
    text-align: left;
  }
}

.hero-name {
  font-size: 1.5rem;
  font-weight: 700;
  color: #fff;
  word-break: break-all;
}

@media (min-width: 768px) {
  .hero-name {
    font-size: 1.625rem;
  }
}

.member-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 14px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: #fff;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.45);
}

/* ---- 表单区 ---- */
.form-section {
  padding: 24px 0 48px;
}

.form-card {
  max-width: 560px;
  margin: 0 auto;
  padding: 32px 28px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-lg);
}

@media (min-width: 768px) {
  .form-card {
    padding: 40px 48px;
  }
}

.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  min-height: 200px;
  justify-content: center;
}

.state-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

/* ---- 表单项 ---- */
.form-group {
  margin-bottom: 24px;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.label-icon {
  color: var(--shop-primary);
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-input);
  background: var(--shop-background);
  font-size: 0.875rem;
  color: var(--shop-foreground);
  outline: none;
  box-sizing: border-box;
  font-family: inherit;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-input:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(217, 91%, 52%, 0.15);
}

.field-hint {
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 生日下拉 ---- */
.birthday-selects {
  display: flex;
  gap: 12px;
}

.birthday-selects .form-select {
  flex: 1;
}

.form-select {
  appearance: none;
  width: 100%;
  padding: 12px 16px;
  padding-right: 36px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-input);
  background: var(--shop-background);
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%238890a0' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  font-size: 0.875rem;
  color: var(--shop-foreground);
  outline: none;
  box-sizing: border-box;
  font-family: inherit;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-select:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(217, 91%, 52%, 0.15);
}

/* 未选择时占位选项置灰 */
.form-select option[value=""] {
  color: var(--shop-muted-foreground);
}

/* ---- 性别选择卡片 ---- */
.gender-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.gender-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s, color 0.2s, transform 0.2s, box-shadow 0.2s;
}

.gender-item:hover {
  border-color: var(--shop-primary);
  transform: translateY(-2px);
}

.gender-item input {
  display: none;
}

.gender-icon {
  display: flex;
  color: var(--shop-muted-foreground);
  transition: color 0.2s;
}

.gender-active {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
  box-shadow: 0 4px 14px hsla(217, 91%, 52%, 0.18);
}

.gender-active .gender-icon {
  color: var(--shop-primary);
}

/* ---- 错误提示 ---- */
.form-error {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  padding: 12px 16px;
  border-radius: var(--shop-radius-lg);
  background: hsl(0, 84%, 95%);
  color: var(--shop-destructive);
  font-size: 0.8125rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s, transform 0.25s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ---- 保存按钮 ---- */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}

.save-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 44px;
  border-radius: var(--shop-radius-lg);
  border: none;
  background: linear-gradient(135deg, var(--shop-primary) 0%, var(--shop-primary-dark) 100%);
  color: var(--shop-primary-foreground);
  font-size: 0.9375rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 14px hsla(217, 91%, 52%, 0.35);
  transition: opacity 0.2s, transform 0.2s, box-shadow 0.2s;
}

.save-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-2px);
  box-shadow: 0 6px 18px hsla(217, 91%, 52%, 0.45);
}

.save-btn:active:not(:disabled) {
  transform: translateY(0);
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
