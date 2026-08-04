<template>
  <div class="records-page">
    <!-- ===== 顶部导航 ===== -->
    <div class="top-bar">
      <el-button text size="small" class="back-btn" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        <span>返回优惠券列表</span>
      </el-button>
    </div>

    <!-- ===== 优惠券信息卡（卡片式） ===== -->
    <div class="coupon-card" v-if="couponInfo" :style="{ '--card-color': getCouponColor(couponInfo.type) }">
      <div class="card-left" :style="{ background: getCouponGradient(couponInfo.type) }">
        <div class="card-notch"></div>
        <div class="card-amount">
          <span class="card-symbol">{{ couponInfo.type === '折扣' ? '' : '¥' }}</span>
          <span class="card-number">{{ displayValue(couponInfo.value, couponInfo.type) }}</span>
          <span class="card-suffix">{{ couponInfo.type === '折扣' ? '%' : '' }}</span>
        </div>
        <div class="card-type">{{ couponInfo.type }}</div>
      </div>
      <div class="card-right">
        <div class="card-header">
          <div class="card-name">{{ couponInfo.name }}</div>
          <div class="card-tags">
            <el-tag :type="typeTagType" size="small" effect="plain">{{ couponInfo.type }}</el-tag>
            <span :class="['status-tag', couponInfo.status]">{{ statusLabel(couponInfo.status) }}</span>
          </div>
        </div>
        <div class="card-info">
          <div class="info-row">
            <div class="info-item">
              <span class="info-label">门槛</span>
              <span class="info-value">{{ couponInfo.minAmount > 0 ? '¥' + formatMoney(couponInfo.minAmount) : '无门槛' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">库存</span>
              <span class="info-value">{{ couponInfo.quantity }} 张</span>
            </div>
            <div class="info-item">
              <span class="info-label">有效期</span>
              <span class="info-value date-value">{{ formatValidity(couponInfo) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 统计卡片 ===== -->
    <div class="stats-grid">
      <div class="stat-card" v-for="s in stats" :key="s.label">
        <div class="stat-label">{{ s.label }}</div>
        <div class="stat-value">{{ s.value }}</div>
        <div class="stat-desc">{{ s.desc }}</div>
      </div>
    </div>

    <!-- ===== 发放记录卡片列表 ===== -->
    <div class="records-list">
      <div class="list-header">
        <span class="list-title">发放明细</span>
        <span class="list-count">共 {{ total }} 条记录</span>
      </div>

      <div class="record-card" v-for="(record, index) in recordsList" :key="record.id">
        <div class="record-index">{{ index + 1 }}</div>
        <div class="record-content">
          <div class="record-top">
            <div class="record-user">
              <el-avatar :size="32" class="user-avatar" :style="{ background: getCouponGradient(couponInfo?.type || '立减') }">
                <img v-if="record.avatar" :src="record.avatar" alt="用户头像" />
                <template v-else>{{ record.userId.toString().slice(-2) }}</template>
              </el-avatar>
              <div class="user-detail">
                <div class="user-id">{{ record.user || ('UID ' + record.userId) }}</div>
                <div class="user-name" v-if="record.user">UID {{ record.userId }}</div>
              </div>
            </div>
            <div class="record-meta">
              <el-tag size="small" effect="plain" class="source-tag">{{ record.source || '手动领取' }}</el-tag>
              <span :class="['record-status', record.status]">
                <el-icon :size="12"><component :is="statusIcon(record.status)" /></el-icon>
                {{ record.status }}
              </span>
            </div>
          </div>
          <div class="record-bottom">
            <div class="record-time">
              <span class="time-label">发放时间</span>
              <span class="time-value">{{ record.time }}</span>
            </div>
            <div class="record-time" v-if="record.usedTime > 0">
              <span class="time-label">使用时间</span>
              <span class="time-value">{{ formatTimestamp(record.usedTime) }}</span>
            </div>
            <div class="record-time" v-else>
              <span class="time-label">使用时间</span>
              <span class="time-value empty">暂未使用</span>
            </div>
            <div class="record-time">
              <span class="time-label">有效期</span>
              <span class="time-value date-value">{{ formatRecordValidity(record) }}</span>
            </div>
            <div class="record-order" v-if="record.orderId && record.orderId > 0">
              <span class="order-label">关联订单</span>
              <a class="order-value link" title="查看订单详情" @click="goOrder(record.orderId)">#{{ record.orderId }}</a>
            </div>
            <div class="record-order" v-else>
              <span class="order-label">关联订单</span>
              <span class="order-value empty">暂无关联订单</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="recordsList.length === 0 && !loading" class="empty-state">
        <el-icon :size="48" color="#d0d4dc"><CircleClose /></el-icon>
        <div class="empty-title">暂无发放记录</div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrap" v-if="total > 0">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[10, 20, 50, 100]"
          background
          small
          @current-change="loadRecords"
          @size-change="loadRecords"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CircleCheck, CircleClose, Warning } from '@element-plus/icons-vue'
import { listCouponRecords, getCoupon } from '../../api/coupon'
import type { DistRecord, CouponItem } from '../../types/coupon'

const route = useRoute()
const router = useRouter()
const couponId = Number(route.params.id)

const couponInfo = ref<CouponItem | null>(null)
const recordsList = ref<DistRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

function goBack() {
  router.push({ name: 'Coupons' })
}

/** 跳转订单详情 */
function goOrder(orderId: number) {
  router.push({ name: 'OrderDetail', params: { id: String(orderId) } })
}

function formatTimestamp(ts: number) {
  if (!ts) return ''
  const date = new Date(ts * 1000)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function formatRecordValidity(record: DistRecord) {
  if (!record.validStart) return '-'
  if (!record.validEnd) return `${record.validStart} ~ 永久有效`
  return `${record.validStart} ~ ${record.validEnd}`
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

/* ==================== 优惠券信息 ==================== */

function getCouponColor(type: string) {
  const map: Record<string, string> = { '立减': '#f56c6c', '折扣': '#e6a23c', '满减': '#409eff' }
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

function statusLabel(status: string) {
  const map: Record<string, string> = { 'pending': '待生效', 'active': '进行中', 'ended': '已结束', 'voided': '已作废' }
  return map[status] || status
}

const typeTagType = computed(() => {
  const map: Record<string, string> = { '立减': 'danger', '折扣': 'warning', '满减': 'primary' }
  return map[couponInfo.value?.type || ''] || ''
})

/* ==================== 统计 ==================== */

const stats = computed(() => {
  if (!couponInfo.value) return []
  const info = couponInfo.value
  const issued = info.issuedCount || 0
  const used = info.usedCount || 0
  const rate = issued > 0 ? Math.round((used / issued) * 100) : 0
  const limitText = info.limitPerUser > 0 ? `每人限 ${info.limitPerUser} 张` : '不限领'
  return [
    { label: '发放总量', value: issued.toLocaleString(), desc: `计划发放 ${info.quantity} 张` },
    { label: '已使用', value: used.toLocaleString(), desc: `使用率 ${rate}%` },
    { label: '剩余库存', value: Math.max(info.quantity - issued, 0).toLocaleString(), desc: limitText },
    { label: '每人限领', value: info.limitPerUser > 0 ? `${info.limitPerUser} 张` : '不限', desc: info.limitPerUser > 0 ? '超出不可领取' : '可多次领取' },
  ]
})

/* ==================== 表格辅助 ==================== */

function statusIcon(status: string) {
  if (status === '已使用') return CircleCheck
  if (status === '已过期') return Warning
  return CircleClose
}

/* ==================== 数据加载 ==================== */

async function loadRecords() {
  loading.value = true
  try {
    const res = await listCouponRecords(couponId, {
      page: page.value,
      page_size: pageSize.value
    })
    recordsList.value = res.items
    total.value = res.total
  } catch {
    ElMessage.error('加载发放记录失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const res = await getCoupon(couponId)
    couponInfo.value = res.coupon
  } catch {
    // fallback: 只显示 ID
  }
  loadRecords()
})
</script>

<style scoped>
.records-page {
  animation: fadeSlideUp 0.3s ease both;
}

/* ===== 顶部导航 ===== */
.top-bar {
  margin-bottom: 16px;
}

.back-btn {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.back-btn:hover { color: var(--shop-primary); }

/* ===== 优惠券信息卡（卡片式） ===== */
.coupon-card {
  display: flex;
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  border: 1px solid var(--shop-border);
  margin-bottom: 20px;
}

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

.card-right {
  flex: 1;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
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
}

.card-tags {
  display: flex;
  align-items: center;
  gap: 8px;
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

.card-info {
  flex: 1;
}

.info-row {
  display: flex;
  gap: 24px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
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

/* ===== 统计卡片 ===== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 16px 18px;
  box-shadow: 0 1px 3px rgba(0,0,0,.06);
  border: 1px solid var(--shop-border);
  transition: transform .15s ease, box-shadow .15s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,.08);
}

.stat-label {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
}

.stat-desc {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
}

/* ===== 发放记录列表 ===== */
.records-list {
  background: #fff;
  border-radius: 14px;
  border: 1px solid var(--shop-border);
  padding: 20px;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.list-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.list-count {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* 记录卡片 */
.record-card {
  display: flex;
  padding: 14px 16px;
  border-bottom: 1px solid #f5f5f5;
  transition: background-color 0.15s ease;
}

.record-card:last-child {
  border-bottom: none;
}

.record-card:hover {
  background-color: #fafafa;
}

.record-index {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 8px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
  margin-right: 14px;
}

.record-content {
  flex: 1;
  min-width: 0;
}

.record-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.record-user {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
}

.user-detail {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.user-id {
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  font-weight: 500;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-name {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
}

.record-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.source-tag {
  font-size: 0.6875rem;
}

.record-status {
  font-size: 0.6875rem;
  padding: 3px 10px;
  border-radius: 10px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.record-status.未使用 { background: #eff6ff; color: #2563eb; }
.record-status.已使用 { background: #f0fdf4; color: #16a34a; }
.record-status.已过期 { background: #fef2f2; color: #dc2626; }

.record-bottom {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.record-time, .record-order {
  display: flex;
  align-items: center;
  gap: 6px;
}

.time-label, .order-label {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  white-space: nowrap;
}

.time-value {
  font-size: 0.75rem;
  color: var(--shop-foreground);
  font-family: var(--shop-font-mono);
}

.time-value.empty {
  color: var(--shop-muted-foreground);
  font-family: inherit;
}

.order-value {
  font-size: 0.75rem;
  color: var(--shop-primary);
  font-family: var(--shop-font-mono);
}

/* 关联订单可点击跳转 */
.order-value.link {
  cursor: pointer;
  text-decoration: none;
  transition: color 0.2s, text-decoration 0.2s;
}
.order-value.link:hover {
  color: var(--shop-primary-dark);
  text-decoration: underline;
}

.order-value.empty {
  color: var(--shop-muted-foreground);
  font-family: inherit;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

.empty-title {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  margin-top: 12px;
}

/* 分页 */
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
