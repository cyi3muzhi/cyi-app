<template>
  <!-- 系统设置页面 - 使用 el-tabs 分标签页 -->
  <div class="settings">
    <el-tabs v-model="activeTab" type="card" class="settings-tabs">
      <!-- 基本信息 -->
      <el-tab-pane label="基本信息" name="basic">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">商城信息</div>
          </template>
          <el-form label-width="100px" label-position="top">
            <el-form-item label="商城名称">
              <el-input v-model="basicForm.shopName" placeholder="请输入商城名称" />
            </el-form-item>
            <el-form-item label="商城简介">
              <el-input
                v-model="basicForm.shopDesc"
                type="textarea"
                :rows="3"
                placeholder="请输入商城简介"
              />
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">联系方式</div>
          </template>
          <el-form label-width="100px" label-position="top">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="客服电话">
                  <el-input v-model="basicForm.phone" placeholder="请输入客服电话" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="客服邮箱">
                  <el-input v-model="basicForm.email" placeholder="请输入客服邮箱" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-card>

        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">品牌信息</div>
          </template>
          <el-form label-width="100px" label-position="top">
            <el-form-item label="ICP 备案号">
              <el-input v-model="basicForm.icp" placeholder="请输入 ICP 备案号" />
            </el-form-item>
          </el-form>
        </el-card>

        <div class="form-actions">
          <el-button type="primary" @click="handleSave('basic')">保存设置</el-button>
          <el-button @click="handleReset">重置</el-button>
        </div>
      </el-tab-pane>

      <!-- 安全设置 -->
      <el-tab-pane label="安全设置" name="security">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">修改密码</div>
          </template>
          <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="120px">
            <el-form-item label="当前密码" prop="currentPassword">
              <el-input
                v-model="passwordForm.currentPassword"
                type="password"
                placeholder="请输入当前密码"
                show-password
                style="max-width: 360px"
              />
            </el-form-item>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="新密码" prop="newPassword">
                  <el-input
                    v-model="passwordForm.newPassword"
                    type="password"
                    placeholder="请输入新密码"
                    show-password
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="确认新密码" prop="confirmPassword">
                  <el-input
                    v-model="passwordForm.confirmPassword"
                    type="password"
                    placeholder="请再次输入新密码"
                    show-password
                  />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item>
              <el-button type="primary" @click="handleUpdatePassword">更新密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">登录安全</div>
          </template>
          <div class="toggle-item">
            <div class="toggle-info">
              <span class="toggle-title">登录尝试次数限制</span>
              <span class="toggle-desc">开启后，同一 IP 连续登录失败超过限制将被临时锁定</span>
            </div>
            <el-switch v-model="securityForm.loginAttempts" />
          </div>
          <div class="toggle-item">
            <div class="toggle-info">
              <span class="toggle-title">登录验证码</span>
              <span class="toggle-desc">开启后，管理员登录需要输入图形验证码</span>
            </div>
            <el-switch v-model="securityForm.loginCaptcha" />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 通知设置 -->
      <el-tab-pane label="通知设置" name="notifications">
        <el-card shadow="never" class="settings-card">
          <template #header>
            <div class="section-title">通知开关</div>
          </template>
          <div class="toggle-item" v-for="item in notificationItems" :key="item.key">
            <div class="toggle-info">
              <span class="toggle-title">{{ item.title }}</span>
              <span class="toggle-desc">{{ item.desc }}</span>
            </div>
            <el-switch v-model="notificationForm[item.key]" />
          </div>
        </el-card>

        <div class="form-actions">
          <el-button type="primary" @click="handleSave('notifications')">保存设置</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
/**
 * 系统设置页面
 * 包含基本信息、安全设置、通知设置三个标签页
 */
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

/** 当前激活的标签页 */
const activeTab = ref('basic')

/** 密码表单引用 */
const passwordFormRef = ref(null)

/** 基本信息表单 */
const basicForm = reactive({
  shopName: '悦选商城',
  shopDesc: '精选好物，品质生活',
  phone: '400-888-8888',
  email: 'service@yuexuan.com',
  icp: '京ICP备2025XXXXXX号'
})

/** 密码修改表单 */
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

/** 密码验证规则 */
const passwordRules = {
  currentPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能小于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

/** 安全设置表单 */
const securityForm = reactive({
  loginAttempts: true,
  loginCaptcha: true
})

/** 通知设置表单 */
const notificationForm = reactive({
  newOrder: true,
  refund: true,
  userRegister: false,
  stockAlert: true
})

/** 通知开关列表配置 */
const notificationItems = [
  { key: 'newOrder', title: '新订单通知', desc: '当用户提交新订单时发送通知' },
  { key: 'refund', title: '退款申请通知', desc: '当用户提交退款申请时发送通知' },
  { key: 'userRegister', title: '用户注册通知', desc: '当新用户注册时发送通知' },
  { key: 'stockAlert', title: '库存预警通知', desc: '当商品库存低于阈值时发送通知' }
]

/**
 * 保存设置（模拟）
 * @param {string} section 设置分区名称
 */
function handleSave(section) {
  const messages = {
    basic: '基本信息设置已保存',
    notifications: '通知设置已保存'
  }
  ElMessage.success(messages[section] || '设置已保存')
}

/**
 * 重置基本信息表单
 */
function handleReset() {
  basicForm.shopName = '悦选商城'
  basicForm.shopDesc = '精选好物，品质生活'
  basicForm.phone = '400-888-8888'
  basicForm.email = 'service@yuexuan.com'
  basicForm.icp = '京ICP备2025XXXXXX号'
  ElMessage.success('表单已重置')
}

/**
 * 更新密码
 */
function handleUpdatePassword() {
  passwordFormRef.value.validate((valid) => {
    if (!valid) return

    ElMessageBox.confirm('确定要更新密码吗？', '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      passwordForm.currentPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
      ElMessage.success('密码更新成功')
    }).catch(() => {})
  })
}
</script>

<style scoped>
.settings {
  width: 100%;
}

.settings-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}

.settings-card {
  margin-bottom: 1.25rem;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.toggle-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--shop-border);
}

.toggle-item:last-of-type {
  border-bottom: none;
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

.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 8px;
}
</style>
