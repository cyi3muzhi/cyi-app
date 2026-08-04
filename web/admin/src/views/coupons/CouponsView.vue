<template>
  <div class="coupon-list">
    <div class="stats-grid">
      <div
        class="stat-card"
        v-for="stat in statsCards"
        :key="stat.label"
        @click="activeStatus = stat.key"
      >
        <div class="stat-header">
          <span class="stat-label">{{ stat.label }}</span>
          <div :class="['stat-dot', stat.color]"></div>
        </div>
        <div class="stat-value">
          <span v-if="statsLoading" class="stat-skeleton" />
          <template v-else>{{ stat.count.toLocaleString() }}</template>
        </div>
        <div class="stat-unit">{{ stat.unit }}</div>
      </div>
    </div>

    <div class="filter-bar">
      <div class="filter-left">
        <div class="search-wrapper">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索优惠券名称..."
            :prefix-icon="Search"
            size="default"
            clearable
            @clear="debouncedSearch"
            @input="onSearchInput"
          />
        </div>
        <div class="status-btn-group">
          <button
            v-for="btn in statusButtons"
            :key="btn.key"
            :class="['status-btn', { active: activeStatus === btn.key }, btn.color]"
            @click="onStatusClick(btn.key)"
          >
            <span class="status-btn-inner">{{ btn.label }}</span>
          </button>
        </div>
        <div class="date-wrapper">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            size="default"
            clearable
          />
        </div>
      </div>
      <div class="filter-right">
        <el-button type="primary" @click="handleAdd" class="add-btn">
          <el-icon><Plus /></el-icon>
          新增优惠券
        </el-button>
      </div>
    </div>

    <!-- ====== 卡片列表 ====== -->
    <div class="coupon-card-list">
      <div
        v-for="coupon in couponList"
        :key="coupon.id"
        class="coupon-card"
        :style="{ '--card-color': getCouponColor(coupon.type) }"
      >
        <!-- 卡片左侧：优惠券预览 -->
        <div class="card-left" :style="{ background: getCouponGradient(coupon.type) }">
          <div class="card-notch"></div>
          <div class="card-amount">
            <span class="card-symbol">{{ coupon.type === '折扣' ? '' : '¥' }}</span>
            <span class="card-number">{{ displayValue(coupon.value, coupon.type) }}</span>
            <span class="card-suffix">{{ coupon.type === '折扣' ? '%' : '' }}</span>
          </div>
          <div class="card-type">{{ coupon.type }}</div>
        </div>

        <!-- 卡片右侧：详细信息 -->
        <div class="card-right">
          <div class="card-header">
            <div class="card-name">{{ coupon.name }}</div>
            <div class="card-tags">
              <el-tag :type="typeTagType(coupon.type)" size="small" effect="plain">{{ coupon.type }}</el-tag>
              <span :class="['status-tag', coupon.status]">{{ statusLabel(coupon.status) }}</span>
            </div>
          </div>

          <div class="card-info">
            <div class="info-row">
              <div class="info-item">
                <span class="info-label">门槛</span>
                <span class="info-value">{{ coupon.minAmount > 0 ? '¥' + formatMoney(coupon.minAmount) : '无门槛' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">库存</span>
                <span class="info-value">{{ coupon.quantity }} 张</span>
              </div>
              <div class="info-item">
                <span class="info-label">限领</span>
                <span class="info-value">{{ coupon.limitPerUser > 0 ? coupon.limitPerUser + ' 张' : '不限' }}</span>
              </div>
            </div>
            <div class="info-row">
              <div class="info-item">
                <span class="info-label">有效期</span>
                <span class="info-value date-value">{{ formatValidity(coupon) }}</span>
              </div>
              <div class="info-item flex-1">
                <span class="info-label">发放/使用</span>
                <div class="usage-bar">
                  <span class="usage-text">已发 {{ coupon.issuedCount }} / 已用 {{ coupon.usedCount }}</span>
                  <el-progress
                    :percentage="usagePercent(coupon)"
                    :stroke-width="4"
                    :color="usagePercent(coupon) > 80 ? 'var(--shop-destructive)' : 'var(--card-color)'"
                    class="usage-progress"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="card-actions">
            <el-button text size="small" type="primary" @click="handleEdit(coupon)">
              <el-icon><EditPen /></el-icon>编辑
            </el-button>
            <el-dropdown trigger="click" @command="(cmd:string) => handleAction(coupon, cmd)">
              <el-button text size="small" class="more-btn">
                更多<el-icon><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="records">📋 发放记录</el-dropdown-item>
                  <el-dropdown-item
                    :command="coupon.status === '已作废' || coupon.status === '已结束' ? 'restore' : 'void'"
                  >
                    {{ coupon.status === '已作废' || coupon.status === '已结束' ? '🔄 恢复' : '⛔ 作废' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>🗑️ 删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="couponList.length === 0 && !loading" class="empty-state">
        <el-icon :size="64" color="#d0d4dc"><Ticket /></el-icon>
        <div class="empty-title">{{ emptyDesc }}</div>
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>新增优惠券
        </el-button>
      </div>
    </div>

    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next, jumper"
        background
        small
        @current-change="loadCoupons"
        @size-change="loadCoupons"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      width="660px"
      :close-on-click-modal="false"
      destroy-on-close
      append-to-body
      class="coupon-dialog"
    >
      <!-- ====== 头部：动态类型色图标 ====== -->
      <div class="dialog-header">
        <div class="dialog-header-icon" :style="{ background: typeColor.gradient }">
          <el-icon :size="28"><component :is="typeIcon" /></el-icon>
        </div>
        <div class="dialog-header-text">
          <div class="dialog-header-title">{{ isEditing ? '编辑优惠券' : '新增优惠券' }}</div>
          <div class="dialog-header-desc">
            {{ isEditing ? `ID: ${editingId} · 创建于 ${editCreateTime}` : '配置优惠券的基本信息和优惠规则' }}
          </div>
        </div>
      </div>

      <!-- ====== 预览卡片：动态渐变色 + 剪角装饰 ====== -->
      <div v-if="formData.name" class="preview-card" :style="{ background: typeColor.gradient }">
        <div class="preview-notch"></div>
        <div class="preview-left">
          <div class="preview-amount">
            <span class="preview-symbol">{{ formData.type === '折扣' ? '' : '¥' }}</span>
            <span class="preview-number">{{ formData.value }}</span>
            <span class="preview-suffix">{{ formData.type === '折扣' ? '%' : '' }}</span>
          </div>
          <div class="preview-type">{{ typeLabel }}</div>
        </div>
        <div class="preview-divider"></div>
        <div class="preview-right">
          <div class="preview-name">{{ formData.name }}</div>
          <div class="preview-condition">{{ conditionText }}</div>
        </div>
      </div>
      <!-- 无名称时占位提示 -->
      <div v-else class="preview-placeholder">
        <el-icon :size="40" :color="typeColor.primary"><component :is="typeIcon" /></el-icon>
        <span>填写名称后实时预览</span>
      </div>

      <!-- ====== 表单 ====== -->
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="0"
        size="default"
        class="coupon-form"
      >
        <!-- 基本信息卡片 -->
        <div class="form-card" :style="{ borderLeftColor: typeColor.primary }">
          <div class="form-card-title">
            <el-icon :color="typeColor.primary"><InfoFilled /></el-icon>
            <span>基本信息</span>
          </div>
          <div class="form-card-body">
            <el-form-item prop="name">
              <el-input
                v-model="formData.name"
                placeholder="给优惠券起个响亮的名字"
                maxlength="30"
                clearable
                class="coupon-name-input"
              >
                <template #prefix><el-icon><Ticket /></el-icon></template>
              </el-input>
            </el-form-item>

            <div class="form-row">
              <!-- 类型选择器：带颜色高亮 -->
              <el-form-item prop="type" class="flex-1">
                <div class="field-label">优惠类型</div>
                <el-radio-group v-model="formData.type" @change="onTypeChange" class="type-group">
                  <el-radio-button value="立减">
                    <el-icon><Money /></el-icon>立减
                  </el-radio-button>
                  <el-radio-button value="折扣">
                    <el-icon><Discount /></el-icon>折扣
                  </el-radio-button>
                  <el-radio-button value="满减">
                    <el-icon><Coin /></el-icon>满减
                  </el-radio-button>
                </el-radio-group>
              </el-form-item>
              <!-- 面值输入：与类型联动 -->
              <el-form-item prop="value" class="flex-1">
                <div class="field-label">面值金额</div>
                <div class="value-input-wrap">
                  <el-input-number
                    v-model="formData.value"
                    :min="1"
                    :max="formData.type === '折扣' ? 99 : 99999"
                    :precision="0"
                    class="full-width"
                  />
                  <span class="value-badge" :style="{ background: typeColor.primary }">
                    {{ formData.type === '折扣' ? '%' : '¥' }}
                  </span>
                </div>
              </el-form-item>
            </div>
          </div>
        </div>

        <!-- 使用规则卡片 -->
        <div class="form-card" :style="{ borderLeftColor: typeColor.primary }">
          <div class="form-card-title">
            <el-icon :color="typeColor.primary"><SetUp /></el-icon>
            <span>发放规则</span>
          </div>
          <div class="form-card-body">
            <div class="rule-row-wrapper">
              <div class="rule-item">
                <div class="rule-label">使用门槛</div>
                <div class="rule-input-group">
                  <el-input-number v-model="formData.minAmount" :min="0" :max="999999" :precision="0" controls-position="right" class="rule-input" />
                  <span class="rule-unit">元</span>
                </div>
              </div>
              <div class="rule-item">
                <div class="rule-label">发放库存</div>
                <div class="rule-input-group">
                  <el-input-number v-model="formData.quantity" :min="1" :max="999999" :precision="0" controls-position="right" class="rule-input" />
                  <span class="rule-unit">张</span>
                </div>
              </div>
              <div class="rule-item">
                <div class="rule-label">每人限领</div>
                <div class="rule-input-group">
                  <el-input-number v-model="formData.limitPerUser" :min="0" :max="999999" :precision="0" controls-position="right" class="rule-input" />
                  <span class="rule-unit">张</span>
                </div>
              </div>
            </div>
            <div class="form-hints">
              <span class="hint-item">
                <span class="hint-dot" :style="{ background: typeColor.primary }"></span>
                0 门槛 = 无门槛
              </span>
              <span class="hint-item">
                <span class="hint-dot" :style="{ background: typeColor.primary }"></span>
                0 限领 = 不限制领取
              </span>
            </div>
          </div>
        </div>

        <!-- 有效期卡片 -->
        <div class="form-card" :style="{ borderLeftColor: typeColor.primary }">
          <div class="form-card-title">
            <el-icon :color="typeColor.primary"><Calendar /></el-icon>
            <span>有效期设置</span>
          </div>
          <div class="form-card-body">
            <div class="valid-type-group">
              <el-radio v-model="formData.validType" :label="1">固定日期</el-radio>
              <el-radio v-model="formData.validType" :label="2">领取后N天</el-radio>
              <el-radio v-model="formData.validType" :label="3">永久有效</el-radio>
            </div>
            <el-form-item prop="validDate" v-if="formData.validType === 1">
              <el-date-picker
                v-model="formData.validDate"
                type="daterange"
                range-separator="至"
                start-placeholder="选择开始日期"
                end-placeholder="选择结束日期"
                style="width: 100%"
                value-format="YYYY-MM-DD"
                :popper-class="'date-popper-custom'"
              />
            </el-form-item>
            <el-form-item prop="validDays" v-if="formData.validType === 2" class="valid-days-item">
              <div class="valid-days-input">
                <el-input-number
                  v-model="formData.validDays"
                  :min="1"
                  :max="365"
                  :precision="0"
                  controls-position="right"
                  class="valid-days-number"
                />
                <span class="valid-days-label">天</span>
              </div>
            </el-form-item>
            <div v-if="formData.validType === 3" class="valid-forever">
              <el-icon :color="typeColor.primary"><CircleCheck /></el-icon>
              <span>该优惠券永久有效，无使用期限限制</span>
            </div>
          </div>
        </div>

        <!-- 启用开关卡片 -->
        <div class="form-card switch-card" :style="{ borderLeftColor: typeColor.primary }">
          <div class="form-card-body switch-row">
            <div class="switch-label">
              <el-icon :color="typeColor.primary"><Switch /></el-icon>
              <span>启用状态</span>
              <span class="switch-desc">{{ formData.enabled ? '用户可正常领取该优惠券' : '优惠券暂停发放，已领取的不受影响' }}</span>
            </div>
            <el-form-item prop="enabled">
              <el-switch
                v-model="formData.enabled"
                :active-color="typeColor.primary"
                active-text="启用"
                inactive-text="停用"
              />
            </el-form-item>
          </div>
        </div>
      </el-form>

      <!-- ====== 底部操作 ====== -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false" size="large" class="footer-btn cancel-btn">取消</el-button>
          <el-button
            type="primary"
            :loading="submitting"
            @click="handleSubmit"
            size="large"
            class="footer-btn submit-btn"
            :style="{
              background: submitting ? '' : typeColor.gradient,
              borderColor: submitting ? '' : typeColor.primary
            }"
          >
            <el-icon><Check /></el-icon>
            {{ isEditing ? '保存修改' : '立即创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Search, ArrowDown, Ticket, Money, Discount, Coin,
  InfoFilled, SetUp, Calendar, Switch, Check, EditPen, CircleCheck
} from '@element-plus/icons-vue'
import {
  listCoupons, createCoupon, updateCoupon, deleteCoupon,
  toggleCouponStatus, getCouponStats
} from '../../api/coupon'
import type { CouponItem } from '../../types/coupon'

const router = useRouter()

/* ==================== 状态 ==================== */

const searchKeyword = ref('')
const activeStatus = ref('')
const dateRange = ref<string[] | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const editCreateTime = ref('')
const submitting = ref(false)
const formRef = ref(null)
const statsLoading = ref(false)

const couponList = ref<CouponItem[]>([])
const total = ref(0)
const statsDefault = {
  totalCount: 0,
  activeCount: 0,
  issuedCount: 0,
  expiredCount: 0
}
const stats = ref({ ...statsDefault })

const statusButtons = [
  { key: '', label: '全部', color: 'gray' },
  { key: '1', label: '进行中', color: 'green' },
  { key: '2', label: '未开始', color: 'orange' },
  { key: '3', label: '已结束', color: 'blue' },
  { key: '4', label: '已作废', color: 'red' },
]

const statsCards = computed(() => [
  { key: '', label: '全部优惠券', count: stats.value.totalCount, unit: '张', color: 'blue' },
  { key: '1', label: '进行中', count: stats.value.activeCount, unit: '张', color: 'green' },
  { key: '', label: '已发放', count: stats.value.issuedCount, unit: '张', color: 'orange' },
  { key: '', label: '已过期', count: stats.value.expiredCount, unit: '张', color: 'red' }
])

const statusKeyLabel: Record<string, string> = {
  '': '全部',
  '1': '进行中',
  '2': '未开始',
  '3': '已结束',
  '4': '已作废'
}

const emptyDesc = computed(() => {
  if (loading.value) return ''
  const parts: string[] = []
  if (activeStatus.value) parts.push(statusKeyLabel[activeStatus.value] || '')
  if (searchKeyword.value) parts.push(`"${searchKeyword.value}"`)
  return parts.length > 0 ? `没有找到${parts.join('')}的优惠券` : '暂无优惠券，点击右上角「新增优惠券」创建'
})

/* ==================== 表单 ==================== */

const formData = reactive({
  name: '',
  type: '立减',
  value: 10,
  minAmount: 0,
  quantity: 100,
  limitPerUser: 0,
  validType: 1,
  validDate: null as string[] | null,
  validDays: 7,
  enabled: true
})

const validateValue = (_rule: any, value: number, callback: Function) => {
  if (value == null || value === 0) return callback(new Error('请输入面值'))
  if (formData.type === '折扣') {
    if (value < 1 || value > 99) return callback(new Error('折扣券面值范围为 1~99'))
  } else {
    if (value < 1) return callback(new Error('面值必须大于 0'))
  }
  callback()
}

const validateValidDate = (_rule: any, value: string[] | null, callback: Function) => {
  if (formData.validType === 1 && (!value || value.length < 2)) {
    return callback(new Error('请选择有效期日期'))
  }
  callback()
}

const validateValidDays = (_rule: any, value: number, callback: Function) => {
  if (formData.validType === 2 && (value == null || value <= 0)) {
    return callback(new Error('请输入有效天数'))
  }
  callback()
}

const formRules = {
  name: [
    { required: true, message: '请输入优惠券名称', trigger: 'blur' },
    { max: 30, message: '名称不超过 30 个字符', trigger: 'blur' }
  ],
  type: [{ required: true, message: '请选择优惠类型', trigger: 'change' }],
  value: [{ required: true, validator: validateValue, trigger: 'blur' }],
  quantity: [
    { required: true, message: '请输入发放数量', trigger: 'blur' },
    { type: 'number', min: 1, message: '发放数量必须大于 0', trigger: 'blur' }
  ],
  validDate: [{ validator: validateValidDate, trigger: 'change' }],
  validDays: [{ validator: validateValidDays, trigger: 'blur' }]
}

/* ==================== 对话框辅助 ==================== */

const typeIcon = computed(() => {
  if (formData.type === '折扣') return Discount
  if (formData.type === '满减') return Coin
  return Money
})

/** 类型色彩体系 — 立减红、折扣橙、满减蓝 */
const typeColor = computed(() => {
  const map: Record<string, { primary: string; light: string; lighter: string; gradient: string; glow: string }> = {
    '立减': {
      primary: '#f56c6c',
      light: '#fef0f0',
      lighter: '#fff5f5',
      gradient: 'linear-gradient(135deg, #f56c6c 0%, #e74c3c 100%)',
      glow: 'rgba(245, 108, 108, 0.3)'
    },
    '折扣': {
      primary: '#e6a23c',
      light: '#fdf6ec',
      lighter: '#fffbee',
      gradient: 'linear-gradient(135deg, #e6a23c 0%, #d4890a 100%)',
      glow: 'rgba(230, 162, 60, 0.3)'
    },
    '满减': {
      primary: '#409eff',
      light: '#ecf5ff',
      lighter: '#f0f7ff',
      gradient: 'linear-gradient(135deg, #667eea 0%, #409eff 100%)',
      glow: 'rgba(64, 158, 255, 0.3)'
    }
  }
  return map[formData.type] || map['立减']
})

const typeLabel = computed(() => {
  if (formData.type === '折扣') return `折扣券 · ${formData.value}% OFF`
  if (formData.type === '满减') return `满减券 · 满 ${formData.minAmount || 0} 减 ${formData.value}`
  return `立减券 · 直减 ¥${formData.value}`
})

const conditionText = computed(() => {
  const parts: string[] = []
  if (formData.validType === 1) {
    parts.push(formData.validDate && formData.validDate[0] ? `有效期 ${formData.validDate[0]} ~ ${formData.validDate[1] || ''}` : '')
  } else if (formData.validType === 2) {
    parts.push(`领取后 ${formData.validDays} 天有效`)
  } else {
    parts.push('永久有效')
  }
  parts.push(formData.minAmount > 0 ? `满 ¥${formData.minAmount} 可用` : '无门槛')
  parts.push(formData.limitPerUser > 0 ? `每人限领 ${formData.limitPerUser} 张` : '不限领')
  return parts.filter(Boolean).join(' · ')
})

function onTypeChange() {
  // 切换类型时重置面值为合理值
  if (formData.type === '折扣' && formData.value > 99) formData.value = 50
  if (formData.type !== '折扣' && formData.value <= 99) formData.value = 10
}

/* ==================== 搜索防抖 ==================== */

let searchTimer: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadCoupons()
  }, 400)
}

function debouncedSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  currentPage.value = 1
  loadCoupons()
}

function onStatusClick(key: string) {
  activeStatus.value = activeStatus.value === key ? '' : key
  currentPage.value = 1
  loadCoupons()
}

/* ==================== 数据加载 ==================== */

async function loadCoupons() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (activeStatus.value) {
      params.status = activeStatus.value
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.date_from = dateRange.value[0]
      params.date_to = dateRange.value[1]
    }
    const res = await listCoupons(params)
    couponList.value = res.items
    total.value = res.total
  } catch (error) {
    ElMessage.error('获取优惠券列表失败')
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  statsLoading.value = true
  try {
    const res = await getCouponStats()
    stats.value = { ...statsDefault, ...res }
  } catch (error) {
    console.error('获取统计数据失败', error)
  } finally {
    statsLoading.value = false
  }
}

function typeTagType(type: string) {
  const map: Record<string, string> = {
    '立减': 'success',
    '折扣': 'warning',
    '满减': 'primary'
  }
  return map[type] || 'info'
}

function getCouponColor(type: string) {
  const map: Record<string, string> = {
    '立减': '#f56c6c',
    '折扣': '#e6a23c',
    '满减': '#409eff'
  }
  return map[type] || '#f56c6c'
}

function getCouponGradient(type: string) {
  const map: Record<string, string> = {
    '立减': 'linear-gradient(135deg, #f56c6c 0%, #e74c3c 100%)',
    '折扣': 'linear-gradient(135deg, #e6a23c 0%, #d4890a 100%)',
    '满减': 'linear-gradient(135deg, #667eea 0%, #409eff 100%)'
  }
  return map[type] || map['立减']
}

/**
 * 金额格式化（分 → 元，整数不带小数）
 * @param cents 金额（分）
 */
function formatMoney(cents: number): string {
  const yuan = cents / 100
  return Number.isInteger(yuan) ? String(yuan) : yuan.toFixed(2)
}

/** 面值展示：折扣券显示百分比，其余金额分 → 元 */
function displayValue(value: number, type: string): string {
  return type === '折扣' ? String(value) : formatMoney(value)
}

function formatValue(coupon: CouponItem) {
  if (coupon.type === '折扣') {
    return `${coupon.value}%`
  }
  return `¥${formatMoney(coupon.value)}`
}

function usagePercent(coupon: CouponItem) {
  const issued = coupon.issuedCount || 0
  const used = coupon.usedCount || 0
  if (issued === 0) return 0
  return Math.round((used / issued) * 100)
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    '进行中': '进行中',
    '未开始': '待生效',
    '已结束': '已结束',
    '已作废': '已作废'
  }
  return map[status] || status
}

function formatValidity(coupon: CouponItem) {
  const validType = coupon.validType || 1
  if (validType === 2) {
    return `领取后 ${coupon.validDays || 7} 天`
  }
  if (validType === 3) {
    return '永久有效'
  }
  return `${coupon.validStart} ~ ${coupon.validEnd}`
}

function handleAdd() {
  isEditing.value = false
  editingId.value = null
  formData.name = ''
  formData.type = '立减'
  formData.value = 10
  formData.minAmount = 0
  formData.quantity = 100
  formData.limitPerUser = 0
  formData.validType = 1
  formData.validDate = null
  formData.validDays = 7
  formData.enabled = true
  dialogVisible.value = true
}

function handleEdit(row: CouponItem) {
  isEditing.value = true
  editingId.value = row.id
  editCreateTime.value = row.createTime ? new Date(row.createTime * 1000).toLocaleString() : ''
  formData.name = row.name
  formData.type = row.type
  formData.value = row.type === '折扣' ? row.value : row.value / 100 // 折扣券为百分比，其余为金额（分→元）
  formData.minAmount = row.minAmount / 100 // 分→元
  formData.quantity = row.quantity
  formData.limitPerUser = row.limitPerUser
  formData.validType = row.validType || 1
  formData.validDate = row.validType === 1 ? [row.validStart, row.validEnd] : null
  formData.validDays = row.validDays || 7
  formData.enabled = row.enabled ?? true
  dialogVisible.value = true
}

/** 更多操作下拉菜单 */
function handleAction(row: CouponItem, cmd: string) {
  if (cmd === 'records') {
    router.push({ name: 'CouponRecords', params: { id: row.id } })
  } else if (cmd === 'void') {
    handleToggleStatus(row)
  } else if (cmd === 'restore') {
    handleToggleStatus(row)
  } else if (cmd === 'delete') {
    handleDelete(row)
  }
}

async function handleSubmit() {
  ;(formRef.value as any)?.validate((valid: boolean) => {
    if (!valid) return
    submitting.value = true
    const [start, end] = formData.validDate || []

    const submitData = {
      name: formData.name,
      type: formData.type,
      // 折扣券 value 为百分比（1-99），其余为金额，统一以「分」提交
      value: formData.type === '折扣' ? formData.value : formData.value * 100,
      minAmount: formData.minAmount * 100, // 元 → 分
      quantity: formData.quantity,
      limitPerUser: formData.limitPerUser,
      validType: formData.validType,
      validDays: formData.validDays,
      validStart: formData.validType === 1 ? start : '',
      validEnd: formData.validType === 1 ? end : '',
      enabled: formData.enabled
    }

    if (isEditing.value) {
      updateCoupon(editingId.value!, submitData).then(() => {
        ElMessage.success('优惠券已更新')
        dialogVisible.value = false
        loadCoupons()
        loadStats()
      }).catch(() => {
        ElMessage.error('更新优惠券失败')
      }).finally(() => {
        submitting.value = false
      })
    } else {
      createCoupon(submitData).then(() => {
        ElMessage.success('优惠券已创建')
        dialogVisible.value = false
        loadCoupons()
        loadStats()
      }).catch(() => {
        ElMessage.error('创建优惠券失败')
      }).finally(() => {
        submitting.value = false
      })
    }
  })
}

function handleToggleStatus(row: CouponItem) {
  const isVoiding = row.status !== '已作废' && row.status !== '已结束'
  const title = isVoiding ? '作废确认' : '恢复确认'
  const message = isVoiding
    ? `确定要将优惠券「${row.name}」作废吗？作废后用户将无法使用。`
    : `确定要恢复优惠券「${row.name}」吗？`

  ElMessageBox.confirm(message, title, {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: isVoiding ? 'warning' : 'info'
  }).then(() => {
    toggleCouponStatus(row.id, isVoiding ? 1 : 2).then(() => {
      ElMessage.success(isVoiding ? '优惠券已作废' : '优惠券已恢复')
      loadCoupons()
      loadStats()
    }).catch(() => {
      ElMessage.error('操作失败')
    })
  }).catch(() => {})
}

function handleDelete(row: CouponItem) {
  ElMessageBox.confirm(
    `确定要删除优惠券「${row.name}」吗？此操作不可撤销。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
  ).then(() => {
    deleteCoupon(row.id).then(() => {
      ElMessage.success('优惠券已删除')
      loadCoupons()
      loadStats()
    }).catch(() => {
      ElMessage.error('删除失败')
    })
  }).catch(() => {})
}

watch([activeStatus, dateRange], () => {
  currentPage.value = 1
  loadCoupons()
})

onMounted(() => {
  loadCoupons()
  loadStats()
})
</script>

<style scoped>
/* ---- 动效 ---- */
@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes rowFadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.coupon-list {
  animation: fadeSlideUp 0.35s ease both;
}

.stat-card {
  animation: fadeSlideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}
.stat-card:nth-child(1) { animation-delay: 0s; }
.stat-card:nth-child(2) { animation-delay: 0.08s; }
.stat-card:nth-child(3) { animation-delay: 0.16s; }
.stat-card:nth-child(4) { animation-delay: 0.24s; }

/* 表格行逐行淡入 */
.el-table__body {
  counter-reset: row-counter;
}
.el-table__body tr.row-enter {
  counter-increment: row-counter;
  animation: rowFadeIn 0.35s ease both;
  animation-delay: calc(counter(row-counter) * 25ms);
}

.coupon-list {
  width: 100%;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: #fff;
  border-radius: var(--shop-radius-lg);
  padding: 1.25rem;
  box-shadow: var(--shop-shadow-sm);
  border: 1px solid var(--shop-border);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
  cursor: pointer;
}

.stat-card:hover {
  box-shadow: var(--shop-shadow-md);
  transform: translateY(-1px);
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
}

.stat-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.stat-dot.blue { background: var(--shop-primary); }
.stat-dot.green { background: var(--shop-success); }
.stat-dot.orange { background: var(--shop-warning); }
.stat-dot.red { background: var(--shop-destructive); }

.stat-value {
  font-size: 1.625rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
}

.stat-unit {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 0.125rem;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding: 18px 24px;
  background: #fff;
  border: 1px solid var(--shop-border);
  border-radius: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.filter-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
}

.search-wrapper {
  position: relative;
  width: 260px;
}

.search-wrapper .el-input {
  width: 100%;
}

.search-wrapper .el-input__wrapper {
  border-radius: 10px;
  border: 1px solid var(--shop-border);
  transition: all 0.25s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

.search-wrapper .el-input__wrapper:hover {
  border-color: var(--shop-primary);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}

.search-wrapper .el-input__wrapper.is-focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.search-wrapper .el-input__icon {
  color: var(--shop-muted-foreground);
  transition: color 0.2s ease;
}

.search-wrapper .el-input__wrapper.is-focus .el-input__icon {
  color: var(--shop-primary);
}

.status-btn-group {
  display: flex;
  gap: 8px;
}

.status-btn {
  position: relative;
  padding: 8px 18px;
  border-radius: 20px;
  border: 1.5px solid var(--shop-border);
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.status-btn:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}

.status-btn:active {
  transform: translateY(0);
}

.status-btn.active {
  border-color: transparent;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.status-btn.active::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0;
  transform: scale(0.8);
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.status-btn.active::after {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(
    45deg,
    transparent 30%,
    rgba(255, 255, 255, 0.1) 50%,
    transparent 70%
  );
  transform: rotate(45deg);
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%) rotate(45deg); }
  100% { transform: translateX(100%) rotate(45deg); }
}

.status-btn.active.gray {
  background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
}

.status-btn.active.green {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
}

.status-btn.active.orange {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.status-btn.active.blue {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.status-btn.active.red {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.status-btn-inner {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.status-btn-inner::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.5;
  transition: opacity 0.2s ease;
}

.status-btn.active .status-btn-inner::before {
  opacity: 1;
  background: rgba(255, 255, 255, 0.8);
}

.date-wrapper {
  position: relative;
  width: 260px;
}

.date-wrapper .el-date-picker {
  width: 100%;
}

.date-wrapper .el-input__wrapper {
  border-radius: 10px;
  border: 1px solid var(--shop-border);
  transition: all 0.25s ease;
}

.date-wrapper .el-input__wrapper:hover {
  border-color: var(--shop-primary);
}

.date-wrapper .el-input__wrapper.is-focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.add-btn {
  border-radius: 10px;
  padding: 8px 24px;
  font-weight: 500;
  transition: all 0.25s ease;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.25);
}

.add-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.35);
}

/* ===== 卡片列表 ===== */

.coupon-card-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  margin-bottom: 20px;
}

.coupon-card {
  display: flex;
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  border: 1px solid var(--shop-border);
  transition: all 0.25s ease;
  animation: cardEnter 0.3s ease both;
}

@keyframes cardEnter {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.coupon-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

/* 卡片左侧：优惠券预览区 */
.card-left {
  width: 180px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #fff;
  position: relative;
  padding: 24px 16px;
}

.card-notch {
  position: absolute;
  top: 0;
  right: -20px;
  width: 40px;
  height: 40px;
  background: #fff;
  border-radius: 0 0 0 100%;
  box-shadow: -2px 2px 4px rgba(0, 0, 0, 0.05);
}

.card-amount {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.card-symbol {
  font-size: 1rem;
  opacity: 0.9;
}

.card-number {
  font-size: 2.5rem;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -1px;
}

.card-suffix {
  font-size: 1rem;
  opacity: 0.9;
}

.card-type {
  font-size: 0.75rem;
  opacity: 0.75;
  margin-top: 6px;
}

/* 卡片右侧：信息区 */
.card-right {
  flex: 1;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.card-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 12px;
}

.card-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.status-tag {
  font-size: 0.7rem;
  padding: 3px 10px;
  border-radius: 10px;
  font-weight: 500;
}

.status-tag.进行中 { background: #f0fdf4; color: #16a34a; }
.status-tag.待生效 { background: #fefce8; color: #ca8a04; }
.status-tag.已结束 { background: #f1f5f9; color: #64748b; }
.status-tag.已作废 { background: #fef2f2; color: #dc2626; }

/* 信息行 */
.card-info {
  flex: 1;
}

.info-row {
  display: flex;
  gap: 24px;
  margin-bottom: 10px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-item.flex-1 {
  flex: 1;
}

.info-label {
  font-size: 0.7rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
  padding: 2px 6px;
  background: #f5f5f5;
  border-radius: 4px;
}

.info-value {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.date-value {
  font-family: var(--shop-font-mono);
}

/* 使用进度条 */
.usage-bar {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 160px;
}

.usage-text {
  font-size: 0.7rem;
  color: var(--shop-muted-foreground);
}

.usage-progress {
  margin: 0;
}

/* 操作按钮 */
.card-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.more-btn {
  color: var(--shop-muted-foreground);
}

.more-btn:hover {
  color: var(--shop-primary);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  background: #fafbfc;
  border-radius: 14px;
  border: 1px dashed var(--shop-border);
}

.empty-title {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  margin-top: 16px;
  margin-bottom: 20px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.进行中 { background: var(--shop-success); }
.status-dot.未开始 { background: var(--shop-warning); }
.status-dot.已结束 { background: #999; }
.status-dot.已作废 { background: var(--shop-destructive); }

/* 统计卡片 skeleton 加载 */
.stat-skeleton {
  display: inline-block;
  width: 60px;
  height: 26px;
  border-radius: 4px;
  background: linear-gradient(90deg, #eee 25%, #f5f5f5 50%, #eee 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* ===== 弹窗设计 ===== */

.coupon-dialog {
  border-radius: 16px;
  overflow: hidden;
}

.coupon-dialog :deep(.el-dialog__header) {
  display: none;
}

.coupon-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.coupon-dialog :deep(.el-dialog__footer) {
  padding: 16px 28px 24px;
  border-top: 1px solid var(--shop-border);
  margin: 0;
  background: #fafbfc;
}

.coupon-dialog :deep(.el-overlay) {
  background-color: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
}

/* 弹窗进入动画 */
.coupon-dialog :deep(.el-dialog) {
  animation: dialogEnter 0.3s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

@keyframes dialogEnter {
  from { opacity: 0; transform: scale(0.92) translateY(20px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

/* ===== 头部 ===== */

.dialog-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 28px 28px 0;
}

.dialog-header-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: background 0.4s ease;
}

.dialog-header-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.3;
}

.dialog-header-desc {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 3px;
}

/* ===== 预览卡片 ===== */

.preview-card {
  position: relative;
  margin: 18px 28px 0;
  display: flex;
  align-items: center;
  gap: 20px;
  border-radius: 14px;
  padding: 22px 26px;
  color: #fff;
  overflow: hidden;
  transition: background 0.4s ease;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

/* 剪角装饰 */
.preview-notch {
  position: absolute;
  top: 0;
  right: 0;
  width: 40px;
  height: 40px;
  background: linear-gradient(225deg, rgba(255,255,255,0.2) 50%, transparent 50%);
  pointer-events: none;
}

/* 分隔竖线 */
.preview-divider {
  width: 1px;
  height: 56px;
  background: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.preview-left {
  flex-shrink: 0;
  text-align: center;
  min-width: 110px;
}

.preview-amount {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 2px;
}

.preview-symbol {
  font-size: 0.875rem;
  opacity: 0.9;
}

.preview-number {
  font-size: 2.25rem;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -1px;
}

.preview-suffix {
  font-size: 0.875rem;
  opacity: 0.9;
}

.preview-type {
  font-size: 0.6875rem;
  opacity: 0.75;
  margin-top: 4px;
  letter-spacing: 0.3px;
}

.preview-right {
  flex: 1;
  min-width: 0;
}

.preview-name {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-condition {
  font-size: 0.6875rem;
  opacity: 0.75;
  margin-top: 6px;
  line-height: 1.4;
}

/* 预览占位 */
.preview-placeholder {
  margin: 18px 28px 0;
  border: 2px dashed #e0e3eb;
  border-radius: 14px;
  padding: 28px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #b0b4c0;
  font-size: 0.8125rem;
  background: #f8f9fc;
  transition: border-color 0.3s, background 0.3s;
}

/* ===== 表单 ===== */

.coupon-form {
  padding: 18px 28px 0;
}

/* 表单卡片：白色底 + 左侧彩色装饰条 */
.form-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid var(--shop-border);
  border-left: 4px solid;
  margin-bottom: 14px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.2s ease;
}

.form-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.form-card-title {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-foreground);
  padding: 14px 18px 0;
}

.form-card-body {
  padding: 12px 18px 16px;
}

.form-row {
  display: flex;
  gap: 14px;
}

.form-row .flex-1 {
  flex: 1;
}

.form-row .el-form-item {
  margin-bottom: 0;
}

/* 字段标签 */
.field-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  margin-bottom: 6px;
}

/* 类型选择器：选中态颜色增强 */
.type-group :deep(.el-radio-button__inner) {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.8125rem;
  padding: 8px 16px;
  transition: all 0.25s ease;
}

.type-group :deep(.el-radio-button__orig-radio:checked + .el-radio-button__inner) {
  box-shadow: -1px 0 0 0 transparent;
}

/* 选中的 radio 按钮使用类型色 */
.type-group :deep(.is-active .el-radio-button__inner) {
  font-weight: 600;
}

/* 面值输入框 */
.value-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.value-badge {
  position: absolute;
  right: 2px;
  top: 50%;
  transform: translateY(-50%);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 6px;
  pointer-events: none;
  z-index: 1;
  transition: background 0.4s ease;
}

.value-input-wrap :deep(.el-input-number) {
  width: 100%;
}

.value-input-wrap :deep(.el-input-number .el-input__wrapper) {
  padding-right: 52px;
}

/* 发放规则标签 */
.rule-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  margin-bottom: 6px;
}

.rule-row-wrapper {
  display: flex;
  gap: 16px;
}

.rule-item {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.rule-input-group {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rule-input {
  width: 100% !important;
}

.rule-unit {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

.full-width {
  width: 100% !important;
}

/* 提示文字 */
.form-hints {
  display: flex;
  gap: 20px;
  margin-top: 10px;
}

.hint-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
}

.hint-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
  transition: background 0.4s ease;
}

/* 有效期类型选择器 */
.valid-type-group {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
}

.valid-type-group :deep(.el-radio__label) {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.valid-type-group :deep(.el-radio__input.is-checked .el-radio__inner) {
  border-color: var(--shop-primary);
  background: var(--shop-primary);
}

.valid-days-item {
  margin-bottom: 0;
}

.valid-days-input {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 200px;
}

.valid-days-number {
  width: 120px !important;
}

.valid-days-label {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
}

.valid-forever {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  padding: 8px 0;
}

.valid-forever :deep(.el-icon) {
  font-size: 18px;
}

/* 启用开关卡片 */
.switch-card .form-card-body {
  padding: 6px 18px;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.switch-label {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.switch-desc {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-left: 2px;
  max-width: 260px;
  line-height: 1.3;
}

.switch-row .el-form-item {
  margin-bottom: 0;
}

/* ===== 底部按钮 ===== */

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.footer-btn {
  min-width: 120px;
  border-radius: 10px;
  font-weight: 600;
  transition: all 0.25s ease;
}

.footer-btn.submit-btn {
  border: none;
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: box-shadow 0.25s ease, transform 0.2s ease, background 0.4s ease;
}

.footer-btn.submit-btn:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  transform: translateY(-1px);
}

.footer-btn.submit-btn:active {
  transform: translateY(0);
}

.footer-btn.cancel-btn {
  border: 1px solid var(--shop-border);
  background: #fff;
  color: var(--shop-muted-foreground);
}

.footer-btn.cancel-btn:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  background: #f5f7ff;
}

/* ===== 分页 ===== */

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

@media (max-width: 1024px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .filter-bar { flex-direction: column; align-items: flex-start; }
  .filter-left { width: 100%; }
  .filter-right { width: 100%; }
  .filter-right .el-button { width: 100%; }
}

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr 1fr; gap: 0.75rem; }
  .filter-left { flex-direction: column; }
  .filter-left .el-input,
  .filter-left .el-date-picker { width: 100% !important; }
}

@media (max-width: 480px) {
  .stats-grid { grid-template-columns: 1fr; }
}
</style>