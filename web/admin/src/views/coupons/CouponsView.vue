<template>
  <!-- 优惠券管理页面 - 统计卡片 + 筛选栏 + 优惠券表格（可展开发放记录）+ 新增/编辑弹窗 -->
  <div class="coupon-list">
    <!-- 统计卡片 -->
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
        <div class="stat-value">{{ stat.count.toLocaleString() }}</div>
        <div class="stat-unit">{{ stat.unit }}</div>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索优惠券名称..."
          :prefix-icon="Search"
          size="default"
          style="width: 220px"
          clearable
        />
        <div class="status-btn-group">
          <el-button
            v-for="btn in statusButtons"
            :key="btn.key"
            :type="activeStatus === btn.key ? 'primary' : 'default'"
            size="small"
            @click="activeStatus = btn.key"
          >
            {{ btn.label }}
          </el-button>
        </div>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="default"
          style="width: 240px"
          clearable
        />
      </div>
      <div class="filter-right">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增优惠券
        </el-button>
      </div>
    </div>

    <!-- 优惠券表格 -->
    <el-table
      :data="filteredCoupons"
      stripe
      style="width: 100%"
      size="small"
      v-loading="loading"
    >
      <!-- 可展开的发放记录 -->
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="expand-records">
            <div class="expand-header">发放记录（共 {{ row.distRecords.length }} 条）</div>
            <el-table :data="row.distRecords" size="small" stripe>
              <el-table-column prop="time" label="发放时间" width="160" />
              <el-table-column prop="user" label="领取用户" width="120" />
              <el-table-column prop="source" label="来源" width="100" />
              <el-table-column prop="status" label="使用状态" width="100">
                <template #default="{ row: rec }">
                  <el-tag
                    :type="rec.status === '已使用' ? 'success' : 'info'"
                    size="small"
                    effect="plain"
                  >
                    {{ rec.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="usedTime" label="使用时间" min-width="160">
                <template #default="{ row: rec }">
                  {{ rec.usedTime || '--' }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="name" label="优惠券名称" min-width="160">
        <template #default="{ row }">
          <span class="coupon-name">{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="110">
        <template #default="{ row }">
          <el-tag
            :type="typeTagType(row.type)"
            size="small"
            effect="dark"
          >
            {{ typeLabel(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="面值" width="110">
        <template #default="{ row }">
          <span class="value-text">{{ formatValue(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="minAmount" label="使用门槛" width="120">
        <template #default="{ row }">
          <span v-if="row.minAmount > 0">满 {{ row.minAmount }} 元可用</span>
          <span v-else class="no-threshold">无门槛</span>
        </template>
      </el-table-column>
      <el-table-column label="发放数量/已用" width="140">
        <template #default="{ row }">
          <div class="qty-cell">
            <span>已发 {{ row.totalDist }}</span>
            <el-progress
              :percentage="usagePercent(row)"
              :stroke-width="6"
              size="small"
              :color="usagePercent(row) > 80 ? 'var(--shop-destructive)' : 'var(--shop-primary)'"
            />
            <span class="used-text">已用 {{ row.usedCount }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="有效期" width="200">
        <template #default="{ row }">
          <span class="date-text">{{ row.validStart }} ~ {{ row.validEnd }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <div class="status-cell">
            <span :class="['status-dot', row.status]"></span>
            <span>{{ statusLabel(row.status) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button
            text
            size="small"
            :type="row.status === 'ended' || row.status === 'voided' ? 'primary' : 'warning'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === 'voided' ? '恢复' : row.status === 'ended' ? '恢复' : '作废' }}
          </el-button>
          <el-button text size="small" type="danger" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="filteredCoupons.length"
        layout="total, prev, pager, next, jumper"
        background
        small
      />
    </div>

    <!-- 新增/编辑优惠券弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑优惠券' : '新增优惠券'"
      width="560px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
        label-position="right"
        size="default"
      >
        <el-form-item label="优惠券名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入优惠券名称" maxlength="30" />
        </el-form-item>

        <el-form-item label="优惠类型" prop="type">
          <el-radio-group v-model="formData.type">
            <el-radio-button value="cash">立减</el-radio-button>
            <el-radio-button value="discount">折扣</el-radio-button>
            <el-radio-button value="full-reduce">满减</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="面值" prop="value">
          <el-input-number
            v-model="formData.value"
            :min="1"
            :max="formData.type === 'discount' ? 99 : 99999"
            :precision="formData.type === 'discount' ? 0 : 0"
            style="width: 200px"
          />
          <span class="form-unit">{{ formData.type === 'discount' ? '%' : '元' }}</span>
        </el-form-item>

        <el-form-item label="使用门槛" prop="minAmount">
          <el-input-number
            v-model="formData.minAmount"
            :min="0"
            :max="999999"
            :precision="0"
            style="width: 200px"
          />
          <span class="form-unit">元（0 表示无门槛）</span>
        </el-form-item>

        <el-form-item label="发放数量" prop="quantity">
          <el-input-number
            v-model="formData.quantity"
            :min="1"
            :max="999999"
            :precision="0"
            style="width: 200px"
          />
          <span class="form-unit">张</span>
        </el-form-item>

        <el-form-item label="有效期" prop="validDate">
          <el-date-picker
            v-model="formData.validDate"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 100%"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-switch
            v-model="formData.enabled"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEditing ? '保存修改' : '立即创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
/**
 * 优惠券管理页面
 * 支持统计概览、搜索筛选、新增/编辑弹窗、状态切换、删除确认、可展开发放记录
 */
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

/* ==================== 状态定义 ==================== */

/** 搜索关键词 */
const searchKeyword = ref('')

/** 激活的状态筛选按钮 */
const activeStatus = ref('all')

/** 日期范围 */
const dateRange = ref(null)

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 加载状态 */
const loading = ref(false)

/** 弹窗可见性 */
const dialogVisible = ref(false)

/** 是否编辑模式 */
const isEditing = ref(false)

/** 编辑中的优惠券 ID */
const editingId = ref(null)

/** 提交中 */
const submitting = ref(false)

/** 表单实例引用 */
const formRef = ref(null)

/* ==================== 数据定义 ==================== */

/** 状态筛选按钮配置 */
const statusButtons = [
  { key: 'all', label: '全部' },
  { key: 'active', label: '进行中' },
  { key: 'ended', label: '已结束' },
  { key: 'voided', label: '已作废' }
]

/** 统计卡片配置 */
const statsCards = computed(() => [
  { key: 'all', label: '全部优惠券', count: couponData.length, unit: '张', color: 'blue' },
  { key: 'active', label: '进行中', count: couponData.filter(c => c.status === 'active').length, unit: '张', color: 'green' },
  { key: 'issued', label: '已发放', count: couponData.reduce((sum, c) => sum + c.totalDist, 0), unit: '张', color: 'orange' },
  { key: 'expired', label: '已过期', count: couponData.filter(c => c.status === 'ended' || c.status === 'voided').length, unit: '张', color: 'red' }
])

/** 表单数据 */
const formData = reactive({
  name: '',
  type: 'cash',
  value: 10,
  minAmount: 0,
  quantity: 100,
  validDate: null,
  enabled: true
})

/** 表单校验规则 */
const formRules = {
  name: [
    { required: true, message: '请输入优惠券名称', trigger: 'blur' },
    { max: 30, message: '名称不超过 30 个字符', trigger: 'blur' }
  ],
  type: [{ required: true, message: '请选择优惠类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入面值', trigger: 'blur' }],
  quantity: [{ required: true, message: '请输入发放数量', trigger: 'blur' }],
  validDate: [{ required: true, message: '请选择有效期', trigger: 'change' }]
}

/* ==================== 模拟数据 ==================== */

/** 优惠券 Coupon 类型定义：
 *  @typedef {Object} Coupon
 *  @property {number} id - 优惠券 ID
 *  @property {string} name - 优惠券名称
 *  @property {'cash'|'discount'|'full-reduce'} type - 优惠类型
 *  @property {number} value - 面值（现金/满减为元，折扣为百分比）
 *  @property {number} minAmount - 使用门槛（元）
 *  @property {number} totalDist - 已发放数量
 *  @property {number} usedCount - 已使用数量
 *  @property {string} validStart - 有效期开始
 *  @property {string} validEnd - 有效期结束
 *  @property {'active'|'pending'|'ended'|'voided'} status - 状态
 *  @property {DistRecord[]} distRecords - 发放记录
 */
const couponData = reactive([
  {
    id: 1,
    name: '新人专享立减券',
    type: 'cash',
    value: 20,
    minAmount: 0,
    totalDist: 500,
    usedCount: 320,
    validStart: '2026-06-01',
    validEnd: '2026-08-31',
    status: 'active',
    distRecords: [
      { time: '2026-06-15 10:23', user: '张**', source: '新人礼包', status: '已使用', usedTime: '2026-06-20 14:30' },
      { time: '2026-06-16 09:12', user: '李**', source: '新人礼包', status: '未使用', usedTime: null },
      { time: '2026-06-18 15:45', user: '王**', source: '新人礼包', status: '已使用', usedTime: '2026-06-25 11:20' }
    ]
  },
  {
    id: 2,
    name: '618 满 200 减 50',
    type: 'full-reduce',
    value: 50,
    minAmount: 200,
    totalDist: 1000,
    usedCount: 680,
    validStart: '2026-06-01',
    validEnd: '2026-06-30',
    status: 'ended',
    distRecords: [
      { time: '2026-06-01 00:05', user: '赵**', source: '活动领取', status: '已使用', usedTime: '2026-06-15 20:10' },
      { time: '2026-06-02 08:30', user: '孙**', source: '活动领取', status: '已使用', usedTime: '2026-06-18 09:45' },
      { time: '2026-06-05 14:22', user: '周**', source: '活动领取', status: '未使用', usedTime: null }
    ]
  },
  {
    id: 3,
    name: '会员 8 折券',
    type: 'discount',
    value: 80,
    minAmount: 100,
    totalDist: 300,
    usedCount: 156,
    validStart: '2026-07-01',
    validEnd: '2026-09-30',
    status: 'active',
    distRecords: [
      { time: '2026-07-01 10:00', user: '吴**', source: '会员权益', status: '已使用', usedTime: '2026-07-05 16:30' },
      { time: '2026-07-03 11:15', user: '郑**', source: '会员权益', status: '未使用', usedTime: null }
    ]
  },
  {
    id: 4,
    name: '开学季满 100 减 15',
    type: 'full-reduce',
    value: 15,
    minAmount: 100,
    totalDist: 800,
    usedCount: 210,
    validStart: '2026-08-20',
    validEnd: '2026-09-15',
    status: 'pending',
    distRecords: []
  },
  {
    id: 5,
    name: '全场通用 5 元券',
    type: 'cash',
    value: 5,
    minAmount: 0,
    totalDist: 2000,
    usedCount: 1450,
    validStart: '2026-05-01',
    validEnd: '2026-07-31',
    status: 'ended',
    distRecords: [
      { time: '2026-05-01 00:00', user: '陈**', source: '自动发放', status: '已使用', usedTime: '2026-05-10 12:00' },
      { time: '2026-05-02 00:00', user: '林**', source: '自动发放', status: '已使用', usedTime: '2026-05-15 18:30' }
    ]
  },
  {
    id: 6,
    name: '限时特惠 7 折券',
    type: 'discount',
    value: 70,
    minAmount: 50,
    totalDist: 150,
    usedCount: 45,
    validStart: '2026-07-10',
    validEnd: '2026-08-10',
    status: 'active',
    distRecords: [
      { time: '2026-07-10 10:00', user: '黄**', source: '限时抢购', status: '未使用', usedTime: null }
    ]
  },
  {
    id: 7,
    name: '老用户回馈券',
    type: 'cash',
    value: 30,
    minAmount: 150,
    totalDist: 200,
    usedCount: 30,
    validStart: '2026-06-15',
    validEnd: '2026-07-15',
    status: 'voided',
    distRecords: [
      { time: '2026-06-15 09:00', user: '杨**', source: '短信领取', status: '已使用', usedTime: '2026-06-20 10:00' }
    ]
  },
  {
    id: 8,
    name: '双 11 预售满 500 减 100',
    type: 'full-reduce',
    value: 100,
    minAmount: 500,
    totalDist: 50,
    usedCount: 8,
    validStart: '2026-10-20',
    validEnd: '2026-11-15',
    status: 'pending',
    distRecords: []
  }
])

/* ==================== 计算属性 ==================== */

/** 根据筛选条件过滤优惠券列表 */
const filteredCoupons = computed(() => {
  let list = [...couponData]
  // 状态筛选
  if (activeStatus.value === 'issued') {
    // "已发放" 按发放数量 > 0 筛选
    list = list.filter(c => c.totalDist > 0)
  } else if (activeStatus.value === 'expired') {
    // "已过期" 包含已结束和已作废
    list = list.filter(c => c.status === 'ended' || c.status === 'voided')
  } else if (activeStatus.value !== 'all') {
    list = list.filter(c => c.status === activeStatus.value)
  }
  // 关键词搜索
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(c => c.name.toLowerCase().includes(kw))
  }
  // 日期范围筛选
  if (dateRange.value) {
    const [start, end] = dateRange.value
    list = list.filter(c => {
      return c.validStart >= start && c.validStart <= end
    })
  }
  return list
})

/* ==================== 辅助方法 ==================== */

/**
 * 获取优惠类型对应的 el-tag 类型
 * @param {string} type 优惠类型
 * @returns {string} tag 类型
 */
function typeTagType(type) {
  const map = {
    cash: 'success',
    discount: 'warning',
    'full-reduce': 'primary'
  }
  return map[type] || 'info'
}

/**
 * 获取优惠类型的中文标签
 * @param {string} type 优惠类型
 * @returns {string} 中文标签
 */
function typeLabel(type) {
  const map = {
    cash: '立减',
    discount: '折扣',
    'full-reduce': '满减'
  }
  return map[type] || type
}

/**
 * 格式化面值显示
 * @param {Object} coupon 优惠券数据
 * @returns {string} 格式化后的面值
 */
function formatValue(coupon) {
  if (coupon.type === 'discount') {
    return `${coupon.value}%`
  }
  return `¥${coupon.value}`
}

/**
 * 计算使用率百分比
 * @param {Object} coupon 优惠券数据
 * @returns {number} 使用率
 */
function usagePercent(coupon) {
  if (coupon.totalDist === 0) return 0
  return Math.round((coupon.usedCount / coupon.totalDist) * 100)
}

/**
 * 获取状态对应的中文标签
 * @param {string} status 状态
 * @returns {string} 中文标签
 */
function statusLabel(status) {
  const map = {
    active: '进行中',
    pending: '待生效',
    ended: '已结束',
    voided: '已作废'
  }
  return map[status] || status
}

/* ==================== 事件处理 ==================== */

/** 打开新增优惠券弹窗 */
function handleAdd() {
  isEditing.value = false
  editingId.value = null
  // 重置表单
  formData.name = ''
  formData.type = 'cash'
  formData.value = 10
  formData.minAmount = 0
  formData.quantity = 100
  formData.validDate = null
  formData.enabled = true
  dialogVisible.value = true
}

/**
 * 编辑优惠券
 * @param {Object} row 优惠券行数据
 */
function handleEdit(row) {
  isEditing.value = true
  editingId.value = row.id
  formData.name = row.name
  formData.type = row.type
  formData.value = row.value
  formData.minAmount = row.minAmount
  formData.quantity = row.totalDist
  formData.validDate = [row.validStart, row.validEnd]
  formData.enabled = row.status === 'active' || row.status === 'pending'
  dialogVisible.value = true
}

/** 提交表单（新增/编辑） */
function handleSubmit() {
  formRef.value.validate((valid) => {
    if (!valid) return
    submitting.value = true
    setTimeout(() => {
      const [start, end] = formData.validDate
      const status = formData.enabled ? 'active' : 'pending'
      if (isEditing.value) {
        // 编辑模式：更新现有数据
        const idx = couponData.findIndex(c => c.id === editingId.value)
        if (idx !== -1) {
          couponData[idx].name = formData.name
          couponData[idx].type = formData.type
          couponData[idx].value = formData.value
          couponData[idx].minAmount = formData.minAmount
          couponData[idx].totalDist = formData.quantity
          couponData[idx].validStart = start
          couponData[idx].validEnd = end
          couponData[idx].status = status
        }
        ElMessage.success('优惠券已更新')
      } else {
        // 新增模式：添加到列表
        couponData.unshift({
          id: Date.now(),
          name: formData.name,
          type: formData.type,
          value: formData.value,
          minAmount: formData.minAmount,
          totalDist: formData.quantity,
          usedCount: 0,
          validStart: start,
          validEnd: end,
          status: status,
          distRecords: []
        })
        ElMessage.success('优惠券已创建')
      }
      submitting.value = false
      dialogVisible.value = false
    }, 500)
  })
}

/**
 * 切换优惠券状态（作废/恢复）
 * @param {Object} row 优惠券行数据
 */
function handleToggleStatus(row) {
  const isVoiding = row.status !== 'voided' && row.status !== 'ended'
  const title = isVoiding ? '作废确认' : '恢复确认'
  const message = isVoiding
    ? `确定要将优惠券「${row.name}」作废吗？作废后用户将无法使用。`
    : `确定要恢复优惠券「${row.name}」吗？`

  ElMessageBox.confirm(message, title, {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: isVoiding ? 'warning' : 'info'
  }).then(() => {
    row.status = isVoiding ? 'voided' : 'active'
    ElMessage.success(isVoiding ? '优惠券已作废' : '优惠券已恢复')
  }).catch(() => {})
}

/**
 * 删除优惠券（确认弹窗）
 * @param {Object} row 优惠券行数据
 */
function handleDelete(row) {
  ElMessageBox.confirm(
    `确定要删除优惠券「${row.name}」吗？此操作不可撤销。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
  ).then(() => {
    const idx = couponData.findIndex(c => c.id === row.id)
    if (idx !== -1) {
      couponData.splice(idx, 1)
    }
    ElMessage.success('优惠券已删除')
  }).catch(() => {})
}
</script>

<style scoped>
.coupon-list {
  width: 100%;
}

/* ==================== 统计卡片 ==================== */
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

/* ==================== 筛选栏 ==================== */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px 20px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.filter-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.status-btn-group {
  display: flex;
  gap: 4px;
}

/* ==================== 表格 ==================== */
.coupon-name {
  font-weight: 600;
  color: var(--shop-foreground);
}

.value-text {
  font-weight: 700;
  font-size: 1rem;
  color: var(--shop-destructive);
}

.no-threshold {
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
}

.qty-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 120px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.qty-cell .el-progress {
  margin: 2px 0;
}

.used-text {
  color: var(--shop-muted-foreground);
}

.date-text {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  font-family: var(--shop-font-mono);
}

/* ==================== 状态指示器 ==================== */
.status-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.active { background: var(--shop-success); }
.status-dot.pending { background: var(--shop-warning); }
.status-dot.ended { background: #999; }
.status-dot.voided { background: var(--shop-destructive); }

/* ==================== 展开区域（发放记录） ==================== */
.expand-records {
  padding: 16px 24px;
  background: hsl(220, 14%, 97%);
}

.expand-header {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  margin-bottom: 12px;
}

/* ==================== 弹窗表单 ==================== */
.form-unit {
  margin-left: 8px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ==================== 分页 ==================== */
.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}

/* ==================== 响应式 ==================== */
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
