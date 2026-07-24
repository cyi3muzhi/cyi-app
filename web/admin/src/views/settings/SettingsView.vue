<template>
  <!-- 系统设置页面 -->
  <div class="settings-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">系统设置</h2>
        <span class="page-subtitle">管理商城全局配置与偏好</span>
      </div>
    </div>

    <!-- 设置标签页 -->
    <el-card shadow="never" class="settings-wrapper">
      <el-tabs v-model="activeTab" type="border-card" class="settings-tabs" @tab-click="onTabClick">
        <!-- 🏪 店铺设置 -->
        <el-tab-pane label="🏪 店铺设置" name="shop">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">商城信息</span></template>
              <el-form label-position="top" class="settings-form">
                <el-row :gutter="24">
                  <el-col :xs="24" :md="12">
                    <el-form-item label="商城名称">
                      <el-input v-model="form.shopName" placeholder="请输入商城名称" maxlength="50" show-word-limit />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="12">
                    <el-form-item label="商城 Logo">
                      <div class="logo-upload">
                        <div class="logo-preview" v-if="form.shopLogo">
                          <img :src="form.shopLogo" alt="logo" />
                          <el-button size="small" circle class="logo-remove" @click="form.shopLogo = ''">
                            <el-icon><Close /></el-icon>
                          </el-button>
                        </div>
                        <el-upload
                          v-else
                          action="#"
                          :show-file-list="false"
                          :before-upload="handleLogoUpload"
                          accept="image/*"
                        >
                          <el-button size="small"><el-icon><Plus /></el-icon> 上传 Logo</el-button>
                        </el-upload>
                        <span class="form-hint">建议尺寸 200x200px，支持 PNG/JPG</span>
                      </div>
                    </el-form-item>
                  </el-col>
                </el-row>
                <el-form-item label="商城简介">
                  <el-input v-model="form.shopDescription" type="textarea" :rows="3" placeholder="请输入商城简介" maxlength="200" show-word-limit />
                </el-form-item>
              </el-form>
            </el-card>

            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">联系方式</span></template>
              <el-form label-position="top" class="settings-form">
                <el-row :gutter="24">
                  <el-col :xs="24" :md="12">
                    <el-form-item label="客服电话">
                      <el-input v-model="form.contactPhone" placeholder="请输入客服电话" />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="12">
                    <el-form-item label="客服邮箱">
                      <el-input v-model="form.contactEmail" placeholder="请输入客服邮箱" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-form>
            </el-card>

            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">品牌信息</span></template>
              <el-form label-position="top" class="settings-form">
                <el-form-item label="ICP 备案号">
                  <el-input v-model="form.icpRecord" placeholder="请输入 ICP 备案号" style="max-width: 400px" />
                </el-form-item>
              </el-form>
            </el-card>

            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave">
                <el-icon v-if="!saving"><Check /></el-icon> 保存设置
              </el-button>
              <el-button @click="handleReset('shop')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 📦 订单设置 -->
        <el-tab-pane label="📦 订单设置" name="order">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">订单规则</span></template>
              <el-form label-position="top" class="settings-form">
                <el-row :gutter="24">
                  <el-col :xs="24" :md="8">
                    <el-form-item label="自动确认收货（天）">
                      <el-input-number v-model="form.autoConfirmDays" :min="1" :max="30" />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="8">
                    <el-form-item label="未支付自动关闭（分钟）">
                      <el-input-number v-model="form.autoCloseMinutes" :min="5" :max="1440" :step="5" />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="8">
                    <el-form-item label="退款超时处理（天）">
                      <el-input-number v-model="form.refundTimeoutDays" :min="1" :max="30" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-form>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('order')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 🛒 商品设置 -->
        <el-tab-pane label="🛒 商品设置" name="product">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">商品规则</span></template>
              <el-form label-position="top" class="settings-form">
                <el-row :gutter="24">
                  <el-col :xs="24" :md="8">
                    <el-form-item label="库存预警阈值">
                      <el-input-number v-model="form.stockAlertThreshold" :min="1" :max="999" />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="8">
                    <el-form-item label="新商品默认上架">
                      <el-switch v-model="form.newProductOnline" active-text="是" inactive-text="否" />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="8">
                    <el-form-item label="商品需审核">
                      <el-switch v-model="form.productApproval" active-text="是" inactive-text="否" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-form>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('product')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 🔔 通知设置 -->
        <el-tab-pane label="🔔 通知设置" name="notification">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">通知开关</span></template>
              <div class="toggle-list">
                <div class="toggle-item" v-for="item in notifyItems" :key="item.key">
                  <div class="toggle-info">
                    <span class="toggle-title">{{ item.title }}</span>
                    <span class="toggle-desc">{{ item.desc }}</span>
                  </div>
                  <el-switch
                    :model-value="getNotifyValue(item.key)"
                    @update:model-value="val => setNotifyValue(item.key, val)" />
                </div>
              </div>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('notification')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 🔒 安全策略 -->
        <el-tab-pane label="🔒 安全策略" name="security">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">登录安全</span></template>
              <div class="toggle-list">
                <div class="toggle-item">
                  <div class="toggle-info">
                    <span class="toggle-title">登录尝试次数限制</span>
                    <span class="toggle-desc">开启后，同一 IP 连续登录失败超过限制将被临时锁定</span>
                  </div>
                  <el-switch v-model="form.loginAttemptLimit" />
                </div>
                <div class="toggle-item">
                  <div class="toggle-info">
                    <span class="toggle-title">登录验证码</span>
                    <span class="toggle-desc">开启后，管理员登录需要输入图形验证码</span>
                  </div>
                  <el-switch v-model="form.loginCaptcha" />
                </div>
              </div>
            </el-card>
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">会话管理</span></template>
              <el-form label-position="top" class="settings-form">
                <el-form-item label="会话超时时间（分钟）">
                  <el-input-number v-model="form.sessionTimeoutMinutes" :min="30" :max="1440" :step="30" />
                  <span class="form-hint" style="margin-left:12px">超过此时间无操作将自动退出登录</span>
                </el-form-item>
              </el-form>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('security')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 💳 支付设置 -->
        <el-tab-pane label="💳 支付设置" name="payment">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">支付方式</span></template>
              <el-form label-position="top" class="settings-form">
                <el-form-item label="启用支付方式">
                  <el-checkbox-group v-model="form.enabledPayments">
                    <el-checkbox value="wechat" label="微信支付">
                      <span class="payment-label">微信支付</span>
                    </el-checkbox>
                    <el-checkbox value="alipay" label="支付宝">
                      <span class="payment-label">支付宝</span>
                    </el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                <el-form-item label="默认支付渠道">
                  <el-select v-model="form.defaultPayment" placeholder="请选择默认支付渠道" style="width:240px">
                    <el-option label="微信支付" value="wechat" />
                    <el-option label="支付宝" value="alipay" />
                  </el-select>
                </el-form-item>
              </el-form>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('payment')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 🚚 物流设置 -->
        <el-tab-pane label="🚚 物流设置" name="shipping">
          <div class="tab-body">
            <el-card shadow="never" class="section-card">
              <template #header><span class="section-title">配送规则</span></template>
              <el-form label-position="top" class="settings-form">
                <el-row :gutter="24">
                  <el-col :xs="24" :md="12">
                    <el-form-item label="包邮门槛（元）">
                      <el-input-number v-model="form.freeShippingThreshold" :min="0" :step="10" :precision="2" />
                      <span class="form-hint" style="margin-left:12px">0 表示不启用包邮</span>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :md="12">
                    <el-form-item label="默认物流公司">
                      <el-input v-model="form.defaultShippingCompany" placeholder="请输入默认物流公司" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </el-form>
            </el-card>
            <div class="form-actions">
              <el-button type="primary" :loading="saving" @click="handleSave"><el-icon v-if="!saving"><Check /></el-icon> 保存设置</el-button>
              <el-button @click="handleReset('shipping')">重置</el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/**
 * 系统设置页面
 * 7 个设置分类标签页，对接后端 API 实现数据持久化
 */
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Close, Plus } from '@element-plus/icons-vue'
import { getSettings, updateSettings } from '@/api/settings'
import type { SystemSettings } from '@/api/settings'
import { useAppStore } from '../../store/app'

/** 当前激活的标签页 */
const activeTab = ref('shop')

/** 保存中状态 */
const saving = ref(false)

/** 全局 store */
const store = useAppStore()

/** 原始数据快照，用于重置 */
let originalData: SystemSettings | null = null

/** 设置表单数据 */
const form = reactive<SystemSettings>({
  shopName: '',
  shopLogo: '',
  shopDescription: '',
  contactPhone: '',
  contactEmail: '',
  icpRecord: '',
  autoConfirmDays: 7,
  autoCloseMinutes: 30,
  refundTimeoutDays: 3,
  stockAlertThreshold: 10,
  newProductOnline: true,
  productApproval: false,
  newOrderNotify: true,
  refundNotify: true,
  userRegisterNotify: false,
  stockAlertNotify: true,
  loginAttemptLimit: true,
  loginCaptcha: true,
  sessionTimeoutMinutes: 120,
  enabledPayments: ['wechat', 'alipay'],
  defaultPayment: 'wechat',
  freeShippingThreshold: 99,
  defaultShippingCompany: ''
})

/** 通知开关列表配置 */
const notifyItems = [
  { key: 'newOrderNotify' as keyof SystemSettings, title: '新订单通知', desc: '当用户提交新订单时发送通知' },
  { key: 'refundNotify' as keyof SystemSettings, title: '退款申请通知', desc: '当用户提交退款申请时发送通知' },
  { key: 'userRegisterNotify' as keyof SystemSettings, title: '用户注册通知', desc: '当新用户注册时发送通知' },
  { key: 'stockAlertNotify' as keyof SystemSettings, title: '库存预警通知', desc: '当商品库存低于阈值时发送通知' }
]

/** 页面加载时获取系统设置 */
async function loadSettings() {
  try {
    const data = await getSettings()
    Object.assign(form, data)
    originalData = JSON.parse(JSON.stringify(data))
  } catch {
    ElMessage.error('加载系统设置失败')
  }
}

onMounted(() => { loadSettings() })

/** 获取通知开关值（处理 TypeScript 类型推导） */
function getNotifyValue(key: keyof SystemSettings): boolean {
  return (form as any)[key] as boolean
}

/** 设置通知开关值 */
function setNotifyValue(key: keyof SystemSettings, val: boolean) {
  (form as any)[key] = val
}

/** Tab 点击时重新加载 */
function onTabClick() {
  // 可选：仅当数据未加载时重新加载
}

/** 处理 Logo 上传（示例，需要对接实际的 upload API） */
function handleLogoUpload(file: File): boolean {
  const reader = new FileReader()
  reader.onload = (e) => {
    form.shopLogo = e.target?.result as string || ''
  }
  reader.readAsDataURL(file)
  ElMessage.success('Logo 已上传（预览模式，实际需对接 upload API）')
  return false
}

/** 保存当前 Tab 的设置 */
async function handleSave() {
  saving.value = true
  try {
    await updateSettings(form)
    originalData = JSON.parse(JSON.stringify(form))
    // 保存成功后刷新 store 中的商城名称，侧边栏等组件同步更新
    store.fetchShopName()
    ElMessage.success('设置已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    saving.value = false
  }
}

/** 重置当前 Tab 的表单 */
function handleReset(_tab: string) {
  if (originalData) {
    Object.assign(form, originalData)
    ElMessage.success('表单已重置')
  }
}
</script>

<style scoped>
/* ---- 页面容器 ---- */
.settings-page {
  animation: fadeSlideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  padding-bottom: 24px;
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
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

/* ---- Tab 外层卡片 ---- */
.settings-wrapper {
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
  animation: fadeSlideUp 0.45s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: 0.08s;
}

.settings-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.settings-tabs :deep(.el-tabs__nav-scroll) {
  background: #f8fafc;
}

.settings-tabs :deep(.el-tabs--border-card) {
  border: none;
  box-shadow: none;
}

.tab-body {
  padding: 8px 0;
  animation: tabFadeIn 0.3s ease both;
}

@keyframes tabFadeIn {
  from { opacity: 0; transform: translateX(12px); }
  to { opacity: 1; transform: translateX(0); }
}

/* ---- 分组卡片 ---- */
.section-card {
  margin-bottom: 16px;
  border-radius: var(--shop-radius-lg);
  animation: cardSlideUp 0.4s ease both;
}

.section-card:nth-child(1) { animation-delay: 0.05s; }
.section-card:nth-child(2) { animation-delay: 0.12s; }
.section-card:nth-child(3) { animation-delay: 0.19s; }

@keyframes cardSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.section-card :deep(.el-card__header) {
  padding: 14px 20px;
  border-bottom: 1px solid var(--shop-border);
  background: var(--shop-muted);
}

.section-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.section-card :deep(.el-card__body) {
  padding: 20px;
}

/* ---- 表单 ---- */
.settings-form {
  max-width: 800px;
}

.form-hint {
  display: block;
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
  line-height: 1.4;
}

/* ---- Logo 上传 ---- */
.logo-upload {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-preview {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: var(--shop-radius-md);
  overflow: hidden;
  border: 1px solid var(--shop-border);
}

.logo-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.logo-remove {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 20px;
  height: 20px;
  background: var(--shop-destructive);
  color: #fff;
  border: none;
}

.logo-remove:hover {
  background: #dc2626;
}

/* ---- 通知/安全开关列表 ---- */
.toggle-list {
  display: flex;
  flex-direction: column;
}

.toggle-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid var(--shop-border);
  transition: background 0.15s;
}

.toggle-item:last-of-type {
  border-bottom: none;
}

.toggle-item:hover {
  background: #fafafa;
  margin: 0 -12px;
  padding: 14px 12px;
  border-radius: var(--shop-radius-md);
}

.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toggle-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

.toggle-desc {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 支付方式 ---- */
.payment-label {
  font-size: 0.875rem;
}

/* ---- 操作按钮 ---- */
.form-actions {
  display: flex;
  gap: 12px;
  padding: 16px 0 8px;
  animation: fadeSlideUp 0.3s ease both;
  animation-delay: 0.25s;
}

/* ---- Tab 切换过渡 ---- */
.settings-tabs :deep(.el-tab-pane) {
  animation: tabPaneIn 0.3s ease both;
}

@keyframes tabPaneIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* ---- 响应式 ---- */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .section-card :deep(.el-card__body) {
    padding: 16px;
  }

  .toggle-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>
