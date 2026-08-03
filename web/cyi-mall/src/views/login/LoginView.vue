<template>
  <!-- ============================================================ -->
  <!-- 1. 顶部导航栏（最小化，独立页面）                               -->
  <!-- ============================================================ -->
  <header class="login-header">
    <div class="header-inner">
      <!-- Logo -->
      <router-link to="/" class="logo-link">
        <span class="logo-text">{{ shopStore.shopName }}</span>
      </router-link>
      <!-- 右侧导航 -->
      <nav class="header-nav">
        <router-link to="/" class="nav-link">首页</router-link>
        <a href="#" class="nav-link">帮助</a>
      </nav>
    </div>
  </header>

  <!-- ============================================================ -->
  <!-- 2. 登录区域（全屏居中）                                        -->
  <!-- ============================================================ -->
  <main class="login-main">
    <div class="login-container">
      <!-- 登录卡片 -->
      <div class="login-card">

        <!-- 标题 -->
        <div class="card-header">
          <h1 class="card-title">欢迎回来</h1>
          <p class="card-subtitle">登录您的悦选账户</p>
        </div>

        <!-- 社交登录 -->
        <div class="social-login">
          <button type="button" class="social-btn" @click="onSocialLogin('wechat')">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178A1.17 1.17 0 0 1 4.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 0 1-1.162 1.178 1.17 1.17 0 0 1-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 0 1 .598.082l1.584.927a.272.272 0 0 0 .14.045.246.246 0 0 0 .242-.246c0-.06-.024-.12-.04-.178l-.325-1.233a.492.492 0 0 1 .178-.553C23.028 18.333 24 16.592 24 14.628c0-3.299-3.063-5.77-7.062-5.77zm-2.18 2.453c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982zm4.36 0c.535 0 .969.44.969.982a.976.976 0 0 1-.969.983.976.976 0 0 1-.969-.983c0-.542.434-.982.97-.982z"/>
            </svg>
            <span>微信登录</span>
          </button>
          <button type="button" class="social-btn" @click="onSocialLogin('alipay')">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M21.422 15.358c-3.22-1.386-6.852-2.927-9.839-4.177 1.215-2.336 2.26-4.674 2.26-4.674H9.764S9.28 6.54 8.63 8.26c-2.08.655-4.215 1.057-6.065 1.118-.353.012-.528-.203-.528-.528v-.529c0-.972.706-1.918 2.467-2.318.883-.2 1.926-.294 2.8-.376.442-.04.63-.176.63-.176s-2.338-.706-3.676-.706c-1.412 0-4.159.706-4.159 3.117 0 2.4 2.223 3.14 4.043 3.14 2.565 0 4.992-1.373 6.59-2.945-.824 1.745-1.946 3.592-3.005 5.224-3.363 1.716-6.167 3.745-6.709 6.81-.924 5.176 5.844 6.204 8.574 3.272 1.692-1.816 1.976-4.768 2.2-6.958.178-1.733.87-2.157 1.975-2.691.353-.17.672-.341.985-.51.996.773 2.213 1.434 3.315 2.079.942.55 1.66 1.155 1.544 2.05-.165 1.267-.837 1.974-2.141 2.36-1.01.3-2.12.293-3.13-.003-.844-.246-1.24-.963-1.952-1.165-.26-.074-.445.13-.361.374.277.796 1.06 1.624 1.812 2.067 1.564.922 3.846.95 5.663.375 1.975-.624 3.26-2.243 3.131-4.275-.074-1.208-1.025-2.022-2.154-2.624zM8.4 18.203c-2.035.449-4.18-.279-4.697-1.807-.184-.547-.09-1.086.278-1.528.773-.928 3.382-1.86 4.562-2.319-.523 1.614-1.045 3.288-1.045 4.954 0 .234.253.729.902.7z"/>
            </svg>
            <span>支付宝登录</span>
          </button>
        </div>

        <!-- 分隔线 -->
        <div class="divider">
          <span class="divider-text">或使用账号密码登录</span>
        </div>

        <!-- 登录表单 -->
        <form class="login-form" @submit.prevent="handleSubmit" novalidate>
          <!-- 错误提示 -->
          <div v-if="formError" class="form-error">{{ formError }}</div>

          <!-- 账号输入 -->
          <div class="form-group">
            <label for="account" class="sr-only">手机号 / 邮箱</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                <polyline points="22,6 12,13 2,6"/>
              </svg>
              <input
                id="account"
                ref="accountRef"
                v-model="account"
                type="text"
                placeholder="手机号"
                autocomplete="username"
                class="input-field"
                :class="{ 'input-error': accountError }"
                @input="clearErrors"
              />
            </div>
            <p v-if="accountError" class="field-error">{{ accountError }}</p>
          </div>

          <!-- 密码输入 -->
          <div class="form-group">
            <label for="password" class="sr-only">密码</label>
            <div class="input-wrap">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="密码"
                autocomplete="current-password"
                class="input-field"
                :class="{ 'input-error': passwordError }"
                @input="clearErrors"
              />
              <!-- 密码显示/隐藏切换 -->
              <button
                type="button"
                class="password-toggle"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                @click="showPassword = !showPassword"
              >
                <svg v-if="!showPassword" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                  <circle cx="12" cy="12" r="3"/>
                </svg>
                <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/>
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/>
                  <line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              </button>
            </div>
            <p v-if="passwordError" class="field-error">{{ passwordError }}</p>
          </div>

          <!-- 忘记密码 -->
          <div class="form-actions">
            <router-link to="#" class="forgot-link">忘记密码？</router-link>
          </div>

          <!-- 登录按钮 -->
          <button
            type="submit"
            class="submit-btn"
            :disabled="loading"
          >
            <span v-if="!loading">登录</span>
            <span v-else class="loading-wrap">
              <svg class="loading-spinner" width="20" height="20" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round" opacity="0.3"/>
                <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
              </svg>
              <span>登录中...</span>
            </span>
          </button>
        </form>

        <!-- 注册入口 -->
        <div class="register-entry">
          <span class="register-text">还没有账号？</span>
          <router-link to="/register" class="register-link">立即注册</router-link>
        </div>

        <!-- 协议提示 -->
        <p class="agreement-note">
          登录即代表同意
          <a href="#">《用户协议》</a>
          和
          <a href="#">《隐私政策》</a>
        </p>
      </div>
    </div>
  </main>

  <!-- ============================================================ -->
  <!-- 3. 页脚（最小化）                                              -->
  <!-- ============================================================ -->
  <footer class="login-footer">
    <p>&copy; 2026 {{ shopStore.shopName }} 版权所有 | 京ICP备2026XXXXXXXX号-1</p>
  </footer>
</template>

<script setup lang="ts">
/**
 * 登录页面
 * 独立页面（无 DefaultLayout），支持社交登录和账号密码登录
 */
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useShopStore } from '@/store/shop'
import { login } from '@/api/auth'

const router = useRouter()
const shopStore = useShopStore()

// ---- 表单数据 ----
const account = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const formError = ref('')
const accountError = ref('')
const passwordError = ref('')
const accountRef = ref<HTMLInputElement | null>(null)

/**
 * 清除所有错误提示
 */
function clearErrors() {
  formError.value = ''
  accountError.value = ''
  passwordError.value = ''
}

/**
 * 表单验证
 * @returns 是否通过验证
 */
function validateForm(): boolean {
  let valid = true
  clearErrors()

  if (!account.value.trim()) {
    accountError.value = '请输入手机号或邮箱'
    valid = false
    accountRef.value?.focus()
  }

  if (!password.value.trim()) {
    passwordError.value = '请输入密码'
    valid = false
    if (account.value.trim()) {
      // 如果账号已有值，聚焦到密码
      const pwInput = document.getElementById('password') as HTMLInputElement
      pwInput?.focus()
    }
  }

  return valid
}

/**
 * 处理表单提交
 */
async function handleSubmit() {
  if (!validateForm()) return

  loading.value = true
  formError.value = ''

  try {
    // 调用真实登录 API
    const res = await login({ username: account.value, password: password.value })
    localStorage.setItem('token', res.token)
    // 登录成功后跳转回原页面，没有 redirect 则去首页
    const redirect = (router.currentRoute.value.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e: any) {
    // 判断是否是"用户不存在"错误，自动跳转注册页
    const msg = e?.response?.data?.message || e?.message || ''
    if (msg.includes('not found') || msg.includes('不存在')) {
      // 自动跳转到注册页，带上手机号预填
      router.push({ path: '/register', query: { phone: account.value } })
      return
    }
    formError.value = msg || '登录失败，请检查账号和密码'
  } finally {
    loading.value = false
  }
}

/**
 * 处理社交登录
 * @param provider - 登录提供商
 */
function onSocialLogin(provider: string) {
  // 实际项目中跳转到 OAuth 授权页面
  console.log(`社交登录: ${provider}`)
}
</script>

<style scoped>
/* ============================================
   登录页面样式
   ============================================ */

/* ---- 顶部导航 ---- */
.login-header {
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

.header-nav {
  display: flex;
  align-items: center;
  gap: 20px;
}

.nav-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
}

.nav-link:hover {
  color: var(--shop-primary);
}

/* ---- 主内容区 ---- */
.login-main {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 16px 60px;
  background: linear-gradient(135deg,
    hsl(217, 91%, 97%) 0%,
    hsl(0, 0%, 100%) 50%,
    hsl(217, 60%, 95%) 100%);
}

.login-container {
  width: 100%;
  max-width: 440px;
}

/* ---- 登录卡片 ---- */
.login-card {
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
  margin-bottom: 32px;
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

/* ---- 社交登录 ---- */
.social-login {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.social-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
  background: var(--shop-background);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
}

.social-btn:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

/* ---- 分隔线 ---- */
.divider {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-top: 1px solid var(--shop-border);
}

.divider-text {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

/* ---- 表单 ---- */
.login-form {
  width: 100%;
}

.form-error {
  margin-bottom: 16px;
  padding: 12px 16px;
  border-radius: var(--shop-radius-lg);
  background: hsl(0, 84%, 95%);
  color: var(--shop-destructive);
  border: 1px solid hsl(0, 84%, 88%);
  font-size: 0.875rem;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.form-group {
  margin-bottom: 20px;
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
  border-radius: var(--shop-radius-lg);
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

.password-toggle {
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

.password-toggle:hover {
  color: var(--shop-foreground);
}

.field-error {
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--shop-destructive);
}

/* ---- 表单操作区 ---- */
.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 24px;
}

.forgot-link {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
}

.forgot-link:hover {
  color: var(--shop-primary);
}

/* ---- 提交按钮 ---- */
.submit-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 48px;
  border-radius: var(--shop-radius-lg);
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

/* ---- 注册入口 ---- */
.register-entry {
  margin-top: 24px;
  text-align: center;
}

.register-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.register-link {
  margin-left: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-primary);
  text-decoration: none;
  transition: color 0.2s;
}

.register-link:hover {
  color: var(--shop-primary-dark);
}

/* ---- 协议提示 ---- */
.agreement-note {
  margin-top: 16px;
  text-align: center;
  font-size: 0.75rem;
  line-height: 1.6;
  color: var(--shop-muted-foreground);
}

.agreement-note a {
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
}

.agreement-note a:hover {
  color: var(--shop-primary);
}

/* ---- 页脚 ---- */
.login-footer {
  text-align: center;
  padding: 16px 24px;
  border-top: 1px solid var(--shop-border);
  background: var(--shop-background);
}

.login-footer p {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 响应式 ---- */
@media (max-width: 480px) {
  .login-card {
    padding: 24px 20px;
  }

  .social-btn {
    font-size: 0.8125rem;
    padding: 8px 12px;
  }
}
</style>
