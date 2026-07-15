<template>
  <!-- 登录页 - 左侧深色品牌展示 + 右侧白色登录表单 -->
  <main class="login-container">
    <!-- 左侧品牌展示区（大屏显示） -->
    <aside class="brand-panel">
      <!-- 装饰圆 -->
      <div class="deco-circle deco-circle-1"></div>
      <div class="deco-circle deco-circle-2"></div>
      <div class="deco-circle deco-circle-3"></div>
      <div class="dots-bg"></div>

      <div class="brand-content">
        <!-- Logo 图标 -->
        <div class="brand-logo">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
        </div>
        <h1 class="brand-title">悦选商城</h1>
        <div class="brand-divider"></div>
        <p class="brand-subtitle">高效管理您的商城运营</p>

        <!-- 功能亮点 -->
        <div class="feature-list">
          <div class="feature-item">
            <span class="feature-icon">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>
              </svg>
            </span>
            <span>商品管理 · 订单处理 · 数据分析</span>
          </div>
          <div class="feature-item">
            <span class="feature-icon">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </span>
            <span>安全防护 · 权限管理 · 操作审计</span>
          </div>
        </div>

        <div class="version-badge">v3.0 · 管理后台</div>
      </div>
    </aside>

    <!-- 右侧登录表单区 -->
    <section class="form-panel">
      <div class="form-card">
        <!-- 移动端品牌提示 -->
        <div class="mobile-brand">
          <span class="mobile-brand-name">悦选商城</span>
          <p class="mobile-brand-sub">管理后台</p>
        </div>

        <!-- 标题 -->
        <div class="form-header">
          <h1 class="form-title">管理员登录</h1>
          <p class="form-desc">请输入您的管理员账号和密码</p>
        </div>

        <!-- 登录表单 -->
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="login-form"
          @submit.prevent="handleLogin"
        >
          <!-- 错误提示 -->
          <div v-if="errorMsg" class="form-error">
            {{ errorMsg }}
          </div>

          <!-- 账号输入 -->
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="管理员账号"
              :prefix-icon="User"
              size="large"
              @input="errorMsg = ''"
            />
          </el-form-item>

          <!-- 密码输入 -->
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="管理员密码"
              :prefix-icon="Lock"
              size="large"
              @input="errorMsg = ''"
            >
              <template #suffix>
                <span class="password-toggle" @click="showPassword = !showPassword">
                  <el-icon v-if="showPassword"><View /></el-icon>
                  <el-icon v-else><Hide /></el-icon>
                </span>
              </template>
            </el-input>
          </el-form-item>

          <!-- 记住密码 -->
          <div class="form-options">
            <el-checkbox v-model="remember">记住登录状态</el-checkbox>
            <el-button type="text" class="forgot-link" disabled>忘记密码？</el-button>
          </div>

          <!-- 登录按钮 -->
          <el-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            native-type="submit"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form>

        <!-- 版权信息 -->
        <div class="form-footer">
          <p>&copy; 2026 悦选商城 版权所有 · 京ICP备2026XXXXXXXX号-1</p>
        </div>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
/**
 * 登录页面组件
 * 左侧深色品牌展示 + 右侧白色登录表单卡片
 * 支持表单验证、密码显示/隐藏切换、记住密码
 */
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, View, Hide } from '@element-plus/icons-vue'
import { useAppStore } from '../../store/app'
import { login as loginApi, getMyPermissions } from '../../api/auth'

const router = useRouter()
const store = useAppStore()
const formRef = ref(null)

/** 登录表单数据 */
const form = reactive({
  username: '',
  password: ''
})

/** 密码显示状态 */
const showPassword = ref(false)

/** 记住密码 */
const remember = ref(false)

/** 加载状态 */
const loading = ref(false)

/** 错误消息 */
const errorMsg = ref('')

/** 表单验证规则 */
const rules = {
  username: [
    { required: true, message: '请输入管理员账号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入管理员密码', trigger: 'blur' }
  ]
}

/**
 * 处理登录提交
 * 调用后端 API -> 存储 token 和用户信息 -> 跳转仪表盘
 */
async function handleLogin() {
  formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    errorMsg.value = ''

    try {
      // 调用后端登录 API
      const res = await loginApi(form.username, form.password)

      // 检查登录结果
      if (!res.token) {
        errorMsg.value = '账号或密码错误，请重试'
        loading.value = false
        return
      }

      // 登录成功，存储信息
      store.login({
        token: res.token,
        name: res.admin?.name || '管理员',
        avatar: res.admin?.avatar || '管',
        role: res.admin?.role || '超级管理员',
        email: res.admin?.email || 'admin@yuexuan.com',
        username: res.admin?.username || form.username
      })

      // 加载当前角色的权限
      try {
        const permRes = await getMyPermissions()
        const perms: Record<string, Record<string, boolean>> = {}
        for (const modKey in permRes.permissions) {
          perms[modKey] = permRes.permissions[modKey]?.values || {}
        }
        store.setPermissions(perms)
      } catch {
        // 权限加载失败不影响登录
      }

      ElMessage.success('登录成功！正在跳转...')
      // 等 Pinia 状态更新完毕再跳转
      await new Promise(r => setTimeout(r, 100))
      // 直接使用 window.location 跳转，避免 router 守卫可能的时序问题
      window.location.href = '/dashboard'
    } catch (e) {
      // 过滤 gRPC 错误前缀，如 "rpc error: code = Unknown desc = 管理员密码错误"
      const rawMsg = e.response?.data?.message || e.response?.data?.reason || ''
      const msg = rawMsg.replace(/rpc error: code = \w+ desc = /g, '')
      if (e.response?.status === 401) {
        errorMsg.value = msg || '账号或密码错误'
      } else if (e.response?.status === 422) {
        errorMsg.value = msg || '账号或密码错误'
      } else if (e.response?.status === 404) {
        errorMsg.value = '服务暂不可用，请稍后再试'
      } else if (e.response?.status === 503) {
        errorMsg.value = '服务暂不可用，请稍后重试'
      } else {
        errorMsg.value = msg || '登录失败，请检查网络连接或联系管理员'
      }
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(135deg, hsl(224, 44%, 10%) 0%, hsl(223, 38%, 16%) 30%, hsl(222, 34%, 22%) 60%, hsl(224, 44%, 10%) 100%);
}

/* ===== 左侧品牌展示区 ===== */
.brand-panel {
  display: none;
  position: relative;
  width: 50%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: linear-gradient(160deg, hsl(224, 48%, 14%) 0%, hsl(223, 42%, 20%) 50%, hsl(222, 48%, 16%) 100%);
}

@media (min-width: 1024px) {
  .brand-panel {
    display: flex;
  }
}

.deco-circle {
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.10);
  pointer-events: none;
}

.deco-circle-1 { width: 480px; height: 480px; top: -120px; right: -80px; opacity: 0.6; }
.deco-circle-2 { width: 320px; height: 320px; bottom: -60px; left: -100px; opacity: 0.4; }
.deco-circle-3 { width: 160px; height: 160px; bottom: 25%; right: 15%; opacity: 0.3; }

.dots-bg {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px);
  background-size: 24px 24px;
}

.brand-content {
  position: relative;
  z-index: 10;
  max-width: 400px;
  padding: 0 48px;
  text-align: center;
}

.brand-logo {
  width: 80px;
  height: 80px;
  margin: 0 auto 32px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--shop-primary), hsl(217, 91%, 38%));
  box-shadow: 0 8px 32px rgba(26, 86, 219, 0.35);
}

.brand-title {
  font-size: 2.25rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #fff;
  margin-bottom: 16px;
}

.brand-divider {
  width: 64px;
  height: 1px;
  margin: 0 auto 20px;
  background: linear-gradient(90deg, transparent, var(--shop-primary-light), transparent);
}

.brand-subtitle {
  font-size: 1.125rem;
  font-weight: 300;
  line-height: 1.6;
  color: hsl(220, 20%, 72%);
}

.feature-list {
  margin-top: 48px;
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.875rem;
  color: hsl(220, 15%, 62%);
}

.feature-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

.version-badge {
  margin-top: 40px;
  display: inline-block;
  padding: 4px 12px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  font-size: 0.75rem;
  letter-spacing: 0.04em;
  color: hsl(220, 12%, 50%);
}

/* ===== 右侧登录表单区 ===== */
.form-panel {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}

@media (min-width: 1024px) {
  .form-panel {
    width: 50%;
  }
}

.form-card {
  width: 100%;
  max-width: 420px;
  padding: 32px 24px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.97);
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.30);
}

@media (min-width: 640px) {
  .form-card {
    padding: 40px;
  }
}

.mobile-brand {
  margin-bottom: 24px;
  text-align: center;
}

@media (min-width: 1024px) {
  .mobile-brand {
    display: none;
  }
}

.mobile-brand-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.mobile-brand-sub {
  margin-top: 4px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.form-header {
  margin-bottom: 32px;
}

.form-title {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--shop-foreground);
}

.form-desc {
  margin-top: 6px;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.form-error {
  margin-bottom: 20px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.875rem;
  background: hsl(0, 84%, 95%);
  color: var(--shop-destructive);
  border: 1px solid hsl(0, 84%, 88%);
}

.login-form :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--shop-input) inset;
  border-radius: var(--shop-radius-md);
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset, 0 0 0 3px hsla(217, 91%, 52%, 0.18);
}

.password-toggle {
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--shop-muted-foreground);
}

.password-toggle:hover {
  color: var(--shop-foreground);
}

.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.forgot-link {
  color: var(--shop-muted-foreground) !important;
  font-size: 0.875rem;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 1rem;
  font-weight: 600;
  border-radius: var(--shop-radius-md);
}

.form-footer {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--shop-border);
  text-align: center;
}

.form-footer p {
  font-size: 0.75rem;
  line-height: 1.6;
  color: var(--shop-muted-foreground);
}
</style>
