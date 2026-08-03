<template>
  <!-- ============================================================ -->
  <!-- 1. 顶部导航栏（最小化，独立页面）                               -->
  <!-- ============================================================ -->
  <header class="reg-header">
    <div class="header-inner">
      <!-- Logo -->
      <router-link to="/" class="logo-link">
        <span class="logo-text">{{ shopStore.shopName }}</span>
      </router-link>
      <!-- 登录入口 -->
      <div class="header-right">
        <span class="login-hint">已有账号？</span>
        <router-link to="/login" class="login-btn">立即登录</router-link>
      </div>
    </div>
  </header>

  <!-- ============================================================ -->
  <!-- 2. 注册区域（全屏居中）                                        -->
  <!-- ============================================================ -->
  <main class="reg-main">
    <div class="reg-container">
      <!-- 注册卡片 -->
      <div class="reg-card">

        <!-- 标题 -->
        <div class="card-header">
          <h1 class="card-title">创建账户</h1>
          <p class="card-subtitle">加入{{ shopStore.shopName }}，享受品质生活</p>
        </div>

        <!-- 步骤指示器 -->
        <div class="steps-indicator">
          <div
            v-for="(step, index) in steps"
            :key="index"
            class="step-item"
            :class="{
              active: currentStep === index,
              done: currentStep > index
            }"
          >
            <div class="step-circle">
              <template v-if="currentStep > index">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </template>
              <template v-else>{{ index + 1 }}</template>
            </div>
            <span class="step-label">{{ step.label }}</span>
          </div>
        </div>

        <!-- ===== 步骤1：填写信息 ===== -->
        <form v-show="currentStep === 0" class="reg-form" @submit.prevent="submitStep1" novalidate>
          <!-- 昵称 -->
          <div class="form-group">
            <label for="nickname" class="field-label">昵称</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
              </svg>
              <input
                id="nickname"
                v-model="form.nickname"
                type="text"
                placeholder="请输入昵称"
                class="input-field"
                :class="{ 'input-error': errors.nickname }"
                @input="errors.nickname = ''"
              />
            </div>
            <p v-if="errors.nickname" class="field-error">{{ errors.nickname }}</p>
          </div>

          <!-- 手机号 -->
          <div class="form-group">
            <label for="phone" class="field-label">手机号</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/>
              </svg>
              <input
                id="phone"
                v-model="form.phone"
                type="tel"
                placeholder="手机号"
                class="input-field"
                :class="{ 'input-error': errors.phone }"
                maxlength="11"
                @input="onPhoneInput"
              />
            </div>
            <p v-if="errors.phone" class="field-error">{{ errors.phone }}</p>
          </div>

          <!-- 密码 -->
          <div class="form-group">
            <label for="password" class="field-label">设置密码</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                id="password"
                v-model="form.password"
                :type="form.showPassword ? 'text' : 'password'"
                placeholder="设置密码（6-20位字符）"
                class="input-field"
                :class="{ 'input-error': errors.password }"
                autocomplete="new-password"
                @input="onPasswordInput"
              />
              <button
                type="button"
                class="pw-toggle"
                @click="form.showPassword = !form.showPassword"
              >
                <svg v-if="!form.showPassword" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </svg>
                <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                  <line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              </button>
            </div>
            <!-- 密码强度指示器 -->
            <div class="strength-area">
              <div class="strength-bar-wrap">
                <div
                  class="strength-bar"
                  :class="strengthClass"
                  :style="{ width: strengthWidth }"
                ></div>
              </div>
              <p class="strength-text" :class="strengthClass">{{ strengthLabel }}</p>
            </div>
            <p v-if="errors.password" class="field-error">{{ errors.password }}</p>
          </div>

          <!-- 确认密码 -->
          <div class="form-group">
            <label for="confirmPassword" class="field-label">确认密码</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                id="confirmPassword"
                v-model="form.confirmPassword"
                :type="form.showConfirmPw ? 'text' : 'password'"
                placeholder="请再次输入密码"
                class="input-field"
                :class="{ 'input-error': errors.confirmPassword }"
                autocomplete="new-password"
                @input="onConfirmInput"
              />
              <button
                type="button"
                class="pw-toggle"
                @click="form.showConfirmPw = !form.showConfirmPw"
              >
                <svg v-if="!form.showConfirmPw" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </svg>
                <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                  <line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              </button>
            </div>
            <p v-if="errors.confirmPassword" class="field-error">{{ errors.confirmPassword }}</p>
          </div>

          <!-- 协议同意 -->
          <div class="form-group agreement-group">
            <label class="agreement-label">
              <input
                v-model="form.agreed"
                type="checkbox"
                class="agreement-checkbox"
                :class="{ 'checkbox-error': errors.agreed }"
              />
              <span class="agreement-text">
                我已阅读并同意
                <a href="#" class="agreement-link">《用户协议》</a>
                和
                <a href="#" class="agreement-link">《隐私政策》</a>
              </span>
            </label>
            <p v-if="errors.agreed" class="field-error">{{ errors.agreed }}</p>
          </div>

          <!-- 提交按钮 -->
          <button type="submit" class="submit-btn" :disabled="step1Loading">
            <span v-if="!step1Loading">下一步</span>
            <span v-else class="loading-wrap">
              <svg class="loading-spinner" width="20" height="20" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round" opacity="0.3"/>
                <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
              </svg>
              <span>提交中...</span>
            </span>
          </button>
        </form>

        <!-- ===== 步骤2：验证手机 ===== -->
        <div v-show="currentStep === 1" class="step-verify">
          <div class="verify-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/>
            </svg>
          </div>
          <p class="verify-hint">验证码已发送至 <strong>{{ form.phone }}</strong></p>
          <div class="verify-code-wrap">
            <input
              v-for="(_, i) in 6"
              :key="i"
              :id="'code-' + i"
              v-model="verifyCode[i]"
              type="text"
              maxlength="1"
              class="code-input"
              :class="{ 'code-filled': verifyCode[i] }"
              @input="onCodeInput($event, i)"
              @keydown.backspace="onCodeBackspace($event, i)"
            />
          </div>
          <p v-if="errors.verifyCode" class="field-error" style="text-align: center;">{{ errors.verifyCode }}</p>
          <button class="submit-btn" :disabled="verifyLoading" @click="submitVerify">
            <span v-if="!verifyLoading">验证并注册</span>
            <span v-else class="loading-wrap">
              <svg class="loading-spinner" width="20" height="20" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round" opacity="0.3"/>
                <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
              </svg>
              <span>验证中...</span>
            </span>
          </button>
          <div class="resend-area">
            <span v-if="resendCountdown > 0" class="resend-timer">{{ resendCountdown }}s 后重新发送</span>
            <button v-else type="button" class="resend-btn" @click="resendCode">重新发送验证码</button>
          </div>
        </div>

        <!-- ===== 步骤3：注册成功 ===== -->
        <div v-show="currentStep === 2" class="step-success">
          <div class="success-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="var(--shop-success)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <h2 class="success-title">注册成功！</h2>
          <p class="success-subtitle">欢迎加入{{ shopStore.shopName }}</p>
          <router-link to="/" class="go-home-btn">去首页逛逛</router-link>
        </div>

        <!-- 登录链接 -->
        <div v-if="currentStep < 2" class="login-entry">
          <span class="login-text">已有账号？</span>
          <router-link to="/login" class="login-link">立即登录</router-link>
        </div>
      </div>
    </div>
  </main>

  <!-- ============================================================ -->
  <!-- 3. 页脚（最小化）                                              -->
  <!-- ============================================================ -->
  <footer class="reg-footer">
    <p>&copy; 2026 {{ shopStore.shopName }} 版权所有 | 京ICP备2026XXXXXXXX号-1</p>
  </footer>

  <!-- ============================================================ -->
  <!-- 4. Toast 通知                                                -->
  <!-- ============================================================ -->
  <Teleport to="body">
    <div class="toast" :class="{ 'toast-show': toastVisible }" :style="{ background: toastBg }">
      <span>{{ toastMessage }}</span>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * 注册页面
 * 独立页面（无 DefaultLayout），三步骤注册流程
 */
import { ref, computed, reactive, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useShopStore } from '@/store/shop'
import { register } from '@/api/auth'

const shopStore = useShopStore()
const route = useRoute()

// ---- 步骤定义 ----
const steps = [
  { label: '填写信息' },
  { label: '验证手机' },
  { label: '注册成功' }
]

const currentStep = ref(0)

// ---- 表单数据 ----
const form = reactive({
  nickname: '',
  phone: (route.query.phone as string) || '',
  password: '',
  confirmPassword: '',
  showPassword: false,
  showConfirmPw: false,
  agreed: false
})

// ---- 表单错误 ----
const errors = reactive({
  nickname: '',
  phone: '',
  password: '',
  confirmPassword: '',
  agreed: '',
  verifyCode: ''
})

// ---- 密码强度 ----
interface StrengthInfo {
  score: number
  label: string
  cssClass: string
}

const strengthInfo = ref<StrengthInfo>({ score: 0, label: '', cssClass: '' })

const strengthClass = computed(() => strengthInfo.value.cssClass)
const strengthWidth = computed(() => {
  if (!form.password) return '0'
  const s = strengthInfo.value.score
  if (s <= 1) return '33.33%'
  if (s <= 3) return '66.66%'
  return '100%'
})
const strengthLabel = computed(() => strengthInfo.value.label)

/**
 * 评估密码强度
 * @param pw - 密码字符串
 * @returns 强度分数 (0-5)
 */
function evaluateStrength(pw: string): number {
  let score = 0
  if (pw.length >= 6) score += 1
  if (pw.length >= 10) score += 1
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) score += 1
  if (/\d/.test(pw)) score += 1
  if (/[^a-zA-Z0-9]/.test(pw)) score += 1
  return score
}

/**
 * 更新密码强度显示
 */
function updateStrength() {
  const pw = form.password
  if (!pw) {
    strengthInfo.value = { score: 0, label: '', cssClass: '' }
    return
  }
  const score = evaluateStrength(pw)
  if (score <= 1) {
    strengthInfo.value = { score, label: '弱 - 建议增加字符类型', cssClass: 'weak' }
  } else if (score <= 3) {
    strengthInfo.value = { score, label: '中', cssClass: 'medium' }
  } else {
    strengthInfo.value = { score, label: '强', cssClass: 'strong' }
  }
}

/**
 * 处理密码输入
 */
function onPasswordInput() {
  errors.password = ''
  updateStrength()
}

/**
 * 处理确认密码输入
 */
function onConfirmInput() {
  errors.confirmPassword = ''
  if (form.confirmPassword && form.password !== form.confirmPassword) {
    errors.confirmPassword = '两次输入的密码不一致'
  }
}

/**
 * 处理手机号输入 - 只允许数字
 */
function onPhoneInput() {
  errors.phone = ''
  form.phone = form.phone.replace(/\D/g, '')
}

// ---- 加载状态 ----
const step1Loading = ref(false)
const verifyLoading = ref(false)

/**
 * 验证步骤1的表单
 * @returns 是否通过验证
 */
function validateStep1(): boolean {
  let valid = true
  Object.assign(errors, {
    nickname: '',
    phone: '',
    password: '',
    confirmPassword: '',
    agreed: '',
    verifyCode: ''
  })

  if (!form.nickname.trim()) {
    errors.nickname = '请输入昵称'
    valid = false
  }

  const phoneRe = /^1\d{10}$/
  if (!phoneRe.test(form.phone)) {
    errors.phone = '请输入有效的11位手机号'
    valid = false
  }

  if (form.password.length < 6 || form.password.length > 20) {
    errors.password = '密码长度需为6-20位'
    valid = false
  }

  if (form.password !== form.confirmPassword || !form.confirmPassword) {
    errors.confirmPassword = '两次输入的密码不一致'
    valid = false
  }

  if (!form.agreed) {
    errors.agreed = '请阅读并同意用户协议和隐私政策'
    valid = false
  }

  return valid
}

/**
 * 提交步骤1 - 模拟API调用
 */
async function submitStep1() {
  if (!validateStep1()) return

  step1Loading.value = true
  // 模拟API延迟
  await new Promise(resolve => setTimeout(resolve, 1500))
  step1Loading.value = false
  currentStep.value = 1
  showToast('信息提交成功！请验证手机号', 'var(--shop-success)')
  startResendCountdown()
}

// ---- 验证码 ----
const verifyCode = reactive<string[]>(Array(6).fill(''))
const resendCountdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

/**
 * 验证码输入处理 - 自动跳到下一格
 */
function onCodeInput(e: Event, index: number) {
  const input = e.target as HTMLInputElement
  const val = input.value.replace(/\D/g, '')
  verifyCode[index] = val
  errors.verifyCode = ''

  if (val && index < 5) {
    const nextInput = document.getElementById('code-' + (index + 1))
    nextInput?.focus()
  }
}

/**
 * 验证码退格处理 - 回到上一格
 */
function onCodeBackspace(e: KeyboardEvent, index: number) {
  if (!verifyCode[index] && index > 0) {
    const prevInput = document.getElementById('code-' + (index - 1))
    prevInput?.focus()
  }
}

/**
 * 提交验证码
 */
async function submitVerify() {
  const code = verifyCode.join('')
  if (code.length < 6) {
    errors.verifyCode = '请输入完整的验证码'
    return
  }

  verifyLoading.value = true
  try {
    // 暂时跳过验证码校验，直接调用注册 API
    // 验证码步骤后续对接短信服务后启用
    await register({
      username: form.phone,      // 手机号作为登录账号
      password: form.password,
      nickname: form.nickname,
      phone: form.phone,
    })
    currentStep.value = 2
    showToast('注册成功！', 'var(--shop-success)')
  } catch (e: any) {
    errors.verifyCode = e?.message || '注册失败，请稍后重试'
  } finally {
    verifyLoading.value = false
  }
}

/**
 * 启动重新发送倒计时
 */
function startResendCountdown() {
  resendCountdown.value = 60
  countdownTimer = setInterval(() => {
    if (resendCountdown.value > 0) {
      resendCountdown.value--
    } else {
      if (countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }
  }, 1000)
}

/**
 * 重新发送验证码
 */
async function resendCode() {
  showToast('验证码已重新发送', 'var(--shop-primary)')
  startResendCountdown()
}

// ---- Toast ----
const toastVisible = ref(false)
const toastMessage = ref('')
const toastBg = ref('var(--shop-destructive)')
let toastTimer: ReturnType<typeof setTimeout> | null = null

/**
 * 显示 Toast 通知
 * @param msg - 消息内容
 * @param bg - 背景色
 */
function showToast(msg: string, bg: string = 'var(--shop-destructive)') {
  if (toastTimer) clearTimeout(toastTimer)
  toastMessage.value = msg
  toastBg.value = bg
  toastVisible.value = true
  toastTimer = setTimeout(() => {
    toastVisible.value = false
  }, 3000)
}

// 清理定时器
onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
  if (toastTimer) clearTimeout(toastTimer)
})
</script>

<style scoped>
/* ============================================
   注册页面样式
   ============================================ */

/* ---- 顶部导航 ---- */
.reg-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: var(--shop-background);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  height: 64px;
}

.header-inner {
  max-width: 1280px;
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.logo-link {
  text-decoration: none;
}

.logo-text {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-primary);
  letter-spacing: -0.02em;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-hint {
  display: none;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

@media (min-width: 640px) {
  .login-hint {
    display: inline;
  }
}

.login-btn {
  padding: 8px 16px;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: opacity 0.2s;
}

.login-btn:hover {
  opacity: 0.9;
}

/* ---- 主内容区 ---- */
.reg-main {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 16px 60px;
  background: linear-gradient(135deg, hsl(217, 91%, 97%), hsl(217, 91%, 93%));
}

.reg-container {
  width: 100%;
  max-width: 440px;
}

/* ---- 注册卡片 ---- */
.reg-card {
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  padding: 40px;
  box-shadow: var(--shop-shadow-xl);
  animation: fadeInUp 0.5s ease both;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.card-header {
  text-align: center;
  margin-bottom: 24px;
}

.card-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  letter-spacing: -0.02em;
}

.card-subtitle {
  margin-top: 8px;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

/* ---- 步骤指示器 ---- */
.steps-indicator {
  display: flex;
  align-items: flex-start;
  margin-bottom: 32px;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  flex: 1;
  position: relative;
}

.step-item::after {
  content: '';
  position: absolute;
  top: 14px;
  left: calc(50% + 18px);
  width: calc(100% - 36px);
  height: 2px;
  background: var(--shop-border);
}

.step-item:last-child::after {
  display: none;
}

.step-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid var(--shop-border);
  background: var(--shop-card);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  transition: all var(--shop-transition);
  z-index: 1;
}

.step-item.active .step-circle {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  border-color: var(--shop-primary);
}

.step-item.done .step-circle {
  background: var(--shop-success);
  color: #fff;
  border-color: var(--shop-success);
}

.step-label {
  font-size: 0.75rem;
  text-align: center;
  color: var(--shop-muted-foreground);
  transition: color var(--shop-transition);
}

.step-item.active .step-label {
  color: var(--shop-primary);
  font-weight: 600;
}

.step-item.done .step-label {
  color: var(--shop-success);
}

/* ---- 表单 ---- */
.reg-form {
  width: 100%;
}

.form-group {
  margin-bottom: 20px;
}

.field-label {
  display: block;
  margin-bottom: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

.input-wrap {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--shop-muted-foreground);
  pointer-events: none;
}

.input-field {
  width: 100%;
  height: 48px;
  padding: 0 44px 0 44px;
  border-radius: var(--shop-radius-xl);
  border: 1px solid var(--shop-input);
  background: var(--shop-background);
  font-size: 0.875rem;
  color: var(--shop-foreground);
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-field:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(217, 91%, 52%, 0.15);
}

.input-error {
  border-color: var(--shop-destructive);
}

.input-error:focus {
  box-shadow: 0 0 0 3px hsla(0, 84%, 60%, 0.15);
}

.pw-toggle {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  background: none;
  border: none;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: color 0.2s;
}

.pw-toggle:hover {
  color: var(--shop-foreground);
}

.field-error {
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--shop-destructive);
}

/* ---- 密码强度 ---- */
.strength-area {
  margin-top: 8px;
}

.strength-bar-wrap {
  height: 4px;
  border-radius: 2px;
  background: var(--shop-border);
  overflow: hidden;
}

.strength-bar {
  height: 100%;
  border-radius: 2px;
  transition: all var(--shop-transition);
}

.strength-bar.weak {
  background: var(--shop-destructive);
}

.strength-bar.medium {
  background: var(--shop-warning);
}

.strength-bar.strong {
  background: var(--shop-success);
}

.strength-text {
  margin-top: 4px;
  font-size: 0.75rem;
}

.strength-text.weak {
  color: var(--shop-destructive);
}

.strength-text.medium {
  color: var(--shop-warning);
}

.strength-text.strong {
  color: var(--shop-success);
}

/* ---- 协议同意 ---- */
.agreement-group {
  margin-bottom: 24px;
}

.agreement-label {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.agreement-checkbox {
  margin-top: 2px;
  width: 16px;
  height: 16px;
  accent-color: var(--shop-primary);
  border: 2px solid var(--shop-border);
  border-radius: 3px;
}

.checkbox-error {
  outline: 2px solid var(--shop-destructive);
  outline-offset: 1px;
}

.agreement-text {
  font-size: 0.75rem;
  line-height: 1.5;
  color: var(--shop-muted-foreground);
}

.agreement-link {
  font-weight: 500;
  color: var(--shop-primary);
  text-decoration: none;
  transition: color 0.2s;
}

.agreement-link:hover {
  color: var(--shop-primary-dark);
  text-decoration: underline;
}

/* ---- 提交按钮 ---- */
.submit-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 48px;
  border-radius: var(--shop-radius-xl);
  border: none;
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
}

.submit-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.submit-btn:active:not(:disabled) {
  transform: scale(0.98);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ---- 验证手机步骤 ---- */
.step-verify {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.verify-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--shop-primary-bg);
}

.verify-hint {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  text-align: center;
}

.verify-hint strong {
  color: var(--shop-foreground);
}

.verify-code-wrap {
  display: flex;
  gap: 8px;
}

.code-input {
  width: 44px;
  height: 52px;
  text-align: center;
  font-size: 1.25rem;
  font-weight: 600;
  border: 2px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  background: var(--shop-background);
  color: var(--shop-foreground);
  outline: none;
  transition: border-color 0.2s;
}

.code-input:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(217, 91%, 52%, 0.15);
}

.code-filled {
  border-color: var(--shop-primary);
}

.resend-area {
  text-align: center;
}

.resend-timer {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.resend-btn {
  background: none;
  border: none;
  font-size: 0.8125rem;
  color: var(--shop-primary);
  cursor: pointer;
  transition: color 0.2s;
}

.resend-btn:hover {
  color: var(--shop-primary-dark);
  text-decoration: underline;
}

/* ---- 注册成功步骤 ---- */
.step-success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 24px 0;
}

.success-icon {
  animation: scaleIn 0.5s ease;
}

@keyframes scaleIn {
  from { transform: scale(0); }
  50% { transform: scale(1.2); }
  to { transform: scale(1); }
}

.success-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.success-subtitle {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.go-home-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 32px;
  border-radius: var(--shop-radius-xl);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  transition: opacity 0.2s;
}

.go-home-btn:hover {
  opacity: 0.9;
}

/* ---- 登录入口 ---- */
.login-entry {
  margin-top: 24px;
  text-align: center;
}

.login-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.login-link {
  margin-left: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-primary);
  text-decoration: none;
  transition: color 0.2s;
}

.login-link:hover {
  color: var(--shop-primary-dark);
}

/* ---- 页脚 ---- */
.reg-footer {
  text-align: center;
  padding: 16px 24px;
  border-top: 1px solid var(--shop-border);
  background: var(--shop-background);
}

.reg-footer p {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- Toast ---- */
.toast {
  position: fixed;
  top: 80px;
  left: 50%;
  transform: translateX(-50%) translateY(-20px);
  opacity: 0;
  visibility: hidden;
  transition: all 300ms ease;
  z-index: 9999;
  border-radius: var(--shop-radius-xl);
  padding: 12px 24px;
  font-size: 0.875rem;
  font-weight: 500;
  color: #fff;
  box-shadow: var(--shop-shadow-lg);
  pointer-events: none;
}

.toast-show {
  opacity: 1;
  visibility: visible;
  transform: translateX(-50%) translateY(0);
}

/* ---- 响应式 ---- */
@media (max-width: 480px) {
  .reg-card {
    padding: 24px 20px;
  }

  .verify-code-wrap {
    gap: 6px;
  }

  .code-input {
    width: 38px;
    height: 46px;
    font-size: 1.125rem;
  }
}
</style>