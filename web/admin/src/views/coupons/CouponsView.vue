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
        <div class="stat-value">{{ stat.count.toLocaleString() }}</div>
        <div class="stat-unit">{{ stat.unit }}</div>
      </div>
    </div>

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

    <el-table
      :data="couponList"
      stripe
      style="width: 100%"
      size="small"
      v-loading="loading"
    >
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="expand-records">
            <div class="expand-header">发放记录（共 {{ row.dist_records?.length || 0 }} 条）</div>
            <el-table :data="row.dist_records || []" size="small" stripe>
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
      <el-table-column prop="min_amount" label="使用门槛" width="120">
        <template #default="{ row }">
          <span v-if="row.min_amount > 0">满 {{ row.min_amount }} 元可用</span>
          <span v-else class="no-threshold">无门槛</span>
        </template>
      </el-table-column>
      <el-table-column label="发放数量/已用" width="140">
        <template #default="{ row }">
          <div class="qty-cell">
            <span>已发 {{ row.issued_count }}</span>
            <el-progress
              :percentage="usagePercent(row)"
              :stroke-width="6"
              size="small"
              :color="usagePercent(row) > 80 ? 'var(--shop-destructive)' : 'var(--shop-primary)'"
            />
            <span class="used-text">已用 {{ row.used_count }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="有效期" width="200">
        <template #default="{ row }">
          <span class="date-text">{{ row.valid_start }} ~ {{ row.valid_end }}</span>
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
            :type="row.status === '已结束' || row.status === '已作废' ? 'primary' : 'warning'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === '已作废' ? '恢复' : row.status === '已结束' ? '恢复' : '作废' }}
          </el-button>
          <el-button text size="small" type="danger" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

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
            <el-radio-button value="立减">立减</el-radio-button>
            <el-radio-button value="折扣">折扣</el-radio-button>
            <el-radio-button value="满减">满减</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="面值" prop="value">
          <el-input-number
            v-model="formData.value"
            :min="1"
            :max="formData.type === '折扣' ? 99 : 99999"
            :precision="0"
            style="width: 200px"
          />
          <span class="form-unit">{{ formData.type === '折扣' ? '%' : '元' }}</span>
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

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import {
  listCoupons, createCoupon, updateCoupon, deleteCoupon,
  toggleCouponStatus, getCouponStats
} from '../../api/coupon'
import type { CouponItem } from '../../types/coupon'

const searchKeyword = ref('')
const activeStatus = ref('')
const dateRange = ref<string[] | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const formRef = ref(null)

const couponList = ref<CouponItem[]>([])
const total = ref(0)
const stats = ref({
  total_count: 0,
  active_count: 0,
  issued_count: 0,
  expired_count: 0
})

const statusButtons = [
  { key: '', label: '全部' },
  { key: '1', label: '进行中' },
  { key: '2', label: '未开始' },
  { key: '3', label: '已结束' },
  { key: '4', label: '已作废' }
]

const statsCards = computed(() => [
  { key: '', label: '全部优惠券', count: stats.value.total_count, unit: '张', color: 'blue' },
  { key: '1', label: '进行中', count: stats.value.active_count, unit: '张', color: 'green' },
  { key: '', label: '已发放', count: stats.value.issued_count, unit: '张', color: 'orange' },
  { key: '', label: '已过期', count: stats.value.expired_count, unit: '张', color: 'red' }
])

const formData = reactive({
  name: '',
  type: '立减',
  value: 10,
  minAmount: 0,
  quantity: 100,
  validDate: null as string[] | null,
  enabled: true
})

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
  try {
    const res = await getCouponStats()
    stats.value = res
  } catch (error) {
    console.error('获取统计数据失败', error)
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

function typeLabel(type: string) {
  const map: Record<string, string> = {
    '立减': '立减',
    '折扣': '折扣',
    '满减': '满减'
  }
  return map[type] || type
}

function formatValue(coupon: CouponItem) {
  if (coupon.type === '折扣') {
    return `${coupon.value}%`
  }
  return `¥${coupon.value}`
}

function usagePercent(coupon: CouponItem) {
  if (coupon.issued_count === 0) return 0
  return Math.round((coupon.used_count / coupon.issued_count) * 100)
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

function handleAdd() {
  isEditing.value = false
  editingId.value = null
  formData.name = ''
  formData.type = '立减'
  formData.value = 10
  formData.minAmount = 0
  formData.quantity = 100
  formData.validDate = null
  formData.enabled = true
  dialogVisible.value = true
}

function handleEdit(row: CouponItem) {
  isEditing.value = true
  editingId.value = row.id
  formData.name = row.name
  formData.type = row.type
  formData.value = row.value
  formData.minAmount = row.min_amount
  formData.quantity = row.quantity
  formData.validDate = [row.valid_start, row.valid_end]
  formData.enabled = row.enabled
  dialogVisible.value = true
}

async function handleSubmit() {
  ;(formRef.value as any)?.validate((valid: boolean) => {
    if (!valid) return
    submitting.value = true
    const [start, end] = formData.validDate || []
    const data = {
      name: formData.name,
      type: formData.type,
      value: formData.value,
      min_amount: formData.minAmount,
      quantity: formData.quantity,
      valid_start: start,
      valid_end: end,
      enabled: formData.enabled
    }
    const promise = isEditing.value
      ? updateCoupon(editingId.value, data)
      : createCoupon(data)

    promise.then(() => {
      ElMessage.success(isEditing.value ? '优惠券已更新' : '优惠券已创建')
      dialogVisible.value = false
      loadCoupons()
      loadStats()
    }).catch(() => {
      ElMessage.error(isEditing.value ? '更新优惠券失败' : '创建优惠券失败')
    }).finally(() => {
      submitting.value = false
    })
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
    toggleCouponStatus(row.id, isVoiding ? 2 : 1).then(() => {
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

watch([searchKeyword, activeStatus, dateRange], () => {
  currentPage.value = 1
  loadCoupons()
})

onMounted(() => {
  loadCoupons()
  loadStats()
})
</script>

<style scoped>
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

.status-dot.进行中 { background: var(--shop-success); }
.status-dot.未开始 { background: var(--shop-warning); }
.status-dot.已结束 { background: #999; }
.status-dot.已作废 { background: var(--shop-destructive); }

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

.form-unit {
  margin-left: 8px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

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