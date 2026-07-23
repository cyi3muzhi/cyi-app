<template>
  <div class="order-detail-page">
    <!-- 图片预览弹窗 -->
    <div v-if="previewImage" class="img-preview-modal" @click="closePreview">
      <img :src="previewImage" class="preview-img" />
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="page-loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>加载中...</span>
    </div>

    <!-- 错误状态 -->
    <el-result v-else-if="error" status="error" title="加载失败" :sub-title="error">
      <template #extra>
        <el-button type="primary" @click="loadOrder">重试</el-button>
        <el-button @click="goBack">返回列表</el-button>
      </template>
    </el-result>

    <!-- 订单详情 -->
    <template v-else-if="order">
      <!-- 顶部操作栏 -->
      <div class="detail-header">
        <div class="header-left">
          <el-button text @click="goBack">
            <el-icon><ArrowLeft /></el-icon> 返回
          </el-button>
          <span class="header-title">订单详情</span>
        </div>
        <div class="header-right">
          <el-button
            v-if="order.status === 'paid'"
            type="success"
            @click="openShipDialog"
          >
            立即发货
          </el-button>
          <el-button
            v-if="order.status === 'pending'"
            type="danger"
            plain
            @click="handleCancel"
          >
            取消订单
          </el-button>
        </div>
      </div>

      <!-- 基本信息卡片 -->
      <div class="info-card">
        <div class="card-title">
          <el-icon><InfoFilled /></el-icon> 基本信息
          <el-tag :type="statusTagType(order.status)" size="small" effect="plain" class="status-tag">
            {{ order.statusText }}
          </el-tag>
        </div>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">订单号</span>
            <span class="value mono">{{ order.orderNo }}</span>
          </div>
          <div class="info-item">
            <span class="label">下单时间</span>
            <span class="value">{{ order.createdAt }}</span>
          </div>
          <div class="info-item">
            <span class="label">买家</span>
            <span class="value">{{ order.userName || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">支付方式</span>
            <span class="value">{{ order.payment || '-' }}</span>
          </div>
        </div>
      </div>

      <!-- 商品明细卡片 -->
      <div class="info-card">
        <div class="card-title">
          <el-icon><Goods /></el-icon> 商品明细
          <span class="card-subtitle">共 {{ order.products.length }} 件商品</span>
        </div>
        <el-table :data="order.products" size="small" style="width: 100%" stripe>
          <el-table-column label="商品" min-width="200">
            <template #default="{ row }">
              <div class="product-cell">
                <img v-if="row.image" :src="row.image" class="product-img" @error="onImgError($event)" @click="openPreview(row.image)" />
                <span v-else class="product-avatar" :style="{ background: nameColor(row.name) }">
                  {{ row.name.charAt(0) }}
                </span>
                <span class="product-name">{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="规格" width="120">
            <template #default="{ row }">{{ row.spec || '-' }}</template>
          </el-table-column>
          <el-table-column label="单价" width="100" align="right">
            <template #default="{ row }">&yen;{{ (row.price / 100).toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="数量" width="70" align="center">
            <template #default="{ row }">{{ row.qty }}</template>
          </el-table-column>
          <el-table-column label="小计" width="110" align="right">
            <template #default="{ row }">&yen;{{ ((row.price * row.qty) / 100).toFixed(2) }}</template>
          </el-table-column>
        </el-table>

        <!-- 金额汇总 -->
        <div class="amount-summary">
          <div class="amount-row">
            <span>商品总额</span>
            <span>&yen;{{ (totalProductAmount / 100).toFixed(2) }}</span>
          </div>
          <div class="amount-row amount-total">
            <span>实付金额</span>
            <span class="total-value">&yen;{{ (order.amount / 100).toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- ===== 发货对话框 ===== -->
    <el-dialog
      v-model="shipDialogVisible"
      title="订单发货"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-form ref="shipFormRef" :model="shipForm" :rules="shipRules" label-width="100px">
        <el-form-item label="物流公司" prop="logistics_company">
          <el-select v-model="shipForm.logistics_company" placeholder="请选择物流公司" style="width: 100%">
            <el-option v-for="c in logisticsCompanies" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="物流单号" prop="logistics_no">
          <el-input v-model="shipForm.logistics_no" placeholder="请输入快递单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="confirmShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 订单详情页面
 * 展示订单基本信息、商品明细、金额汇总，支持发货/取消操作
 */
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, InfoFilled, Goods, Loading } from '@element-plus/icons-vue'
import { getOrder, shipOrder, cancelOrder } from '../../api/order'
import type { OrderItem } from '../../types/order'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const error = ref('')
const order = ref<OrderItem | null>(null)
const previewImage = ref('')

function openPreview(src: string) {
  previewImage.value = src
}

function closePreview() {
  previewImage.value = ''
}

/** 商品总额（前端计算） */
const totalProductAmount = computed(() => {
  if (!order.value) return 0
  return order.value.products.reduce((sum, p) => sum + p.price * p.qty, 0)
})

/** 加载订单详情 */
async function loadOrder() {
  loading.value = true
  error.value = ''
  try {
    const id = Number(route.params.id)
    const res = await getOrder(id)
    order.value = res.order
  } catch (e: any) {
    error.value = e?.message || '加载订单详情失败'
  } finally {
    loading.value = false
  }
}

/** 返回列表 */
function goBack() {
  router.push('/orders')
}

/** 状态 → el-tag 类型 */
function statusTagType(status: string): string {
  const map: Record<string, string> = {
    pending: 'danger',
    paid: 'warning',
    shipped: 'primary',
    completed: 'success',
    cancelled: 'info',
    refunding: 'danger',
    refunded: 'info',
  }
  return map[status] || 'info'
}

/** 根据商品名称生成固定颜色 */
function nameColor(name: string): string {
  const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

/** 图片加载失败时隐藏，显示占位色块 */
function onImgError(e: Event) {
  const el = e.target as HTMLElement
  el.style.display = 'none'
}

/* ==================== 发货 ==================== */

const shipDialogVisible = ref(false)
const shipLoading = ref(false)
const shipFormRef = ref<any>(null)

const shipForm = ref({
  logistics_company: '',
  logistics_no: '',
})

const shipRules = {
  logistics_company: [{ required: true, message: '请选择物流公司', trigger: 'change' }],
  logistics_no: [{ required: true, message: '请输入物流单号', trigger: 'blur' }],
}

const logisticsCompanies = ['顺丰快递', '圆通速递', '中通快递', '韵达快递', '申通快递', '极兔速递', '邮政EMS', '德邦快递']

function openShipDialog() {
  shipForm.value = { logistics_company: '', logistics_no: '' }
  shipDialogVisible.value = true
}

async function confirmShip() {
  if (!shipFormRef.value) return
  const valid = await shipFormRef.value.validate().catch(() => false)
  if (!valid || !order.value) return

  shipLoading.value = true
  try {
    await shipOrder(order.value.id, shipForm.value)
    ElMessage.success(`订单 ${order.value.orderNo} 发货成功`)
    shipDialogVisible.value = false
    loadOrder()
  } catch {
    // 错误已由拦截器处理
  } finally {
    shipLoading.value = false
  }
}

/* ==================== 取消订单 ==================== */

async function handleCancel() {
  if (!order.value) return
  try {
    await ElMessageBox.confirm(
      `确定要取消订单 ${order.value.orderNo} 吗？`,
      '取消确认',
      { confirmButtonText: '确定取消', cancelButtonText: '再想想', type: 'warning' }
    )
    await cancelOrder(order.value.id, '管理员取消')
    ElMessage.success(`订单 ${order.value.orderNo} 已取消`)
    loadOrder()
  } catch {
    // 用户取消或接口错误
  }
}

onMounted(loadOrder)
</script>

<style scoped>
.order-detail-page {
  padding: 24px;
}

/* ---- 加载/错误 ---- */
.page-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--shop-muted-foreground, #64748b);
}

/* ---- 顶部操作栏 ---- */
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--shop-foreground, #1e293b);
}

.header-right {
  display: flex;
  gap: 8px;
}

/* ---- 信息卡片 ---- */
.info-card {
  background: #fff;
  border: 1px solid var(--shop-border, #e8e8e8);
  border-radius: var(--shop-radius-lg, 0.75rem);
  padding: 20px 24px;
  margin-bottom: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground, #1e293b);
  margin-bottom: 16px;
}

.card-subtitle {
  font-size: 0.8125rem;
  font-weight: 400;
  color: var(--shop-muted-foreground, #64748b);
}

.status-tag {
  margin-left: auto;
}

/* ---- 基本信息网格 ---- */
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item .label {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground, #64748b);
}

.info-item .value {
  font-size: 0.875rem;
  color: var(--shop-foreground, #1e293b);
}

.info-item .value.mono {
  font-family: var(--shop-font-mono, 'JetBrains Mono', monospace);
  font-size: 0.8125rem;
  word-break: break-all;
}

/* ---- 金额汇总 ---- */
.amount-summary {
  border-top: 1px solid var(--shop-border, #e8e8e8);
  margin-top: 12px;
  padding-top: 12px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.amount-row {
  display: flex;
  justify-content: space-between;
  width: 260px;
  font-size: 0.875rem;
  color: var(--shop-foreground, #1e293b);
}

.amount-total {
  font-weight: 700;
  font-size: 1rem;
  padding-top: 4px;
  border-top: 1px dashed var(--shop-border, #e8e8e8);
}

.total-value {
  color: var(--shop-destructive, #ef4444);
  font-size: 1.125rem;
}

/* ---- 商品图片占位 ---- */
.product-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.product-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--shop-radius-sm, 0.375rem);
  color: #fff;
  font-size: 0.8125rem;
  font-weight: 600;
  flex-shrink: 0;
}

.product-img {
  width: 40px;
  height: 40px;
  border-radius: var(--shop-radius-sm, 0.375rem);
  flex-shrink: 0;
  :deep(img) {
    object-fit: cover;
    width: 100%;
    height: 100%;
    border-radius: inherit;
  }
  /* 预览遮罩层提升层级，避免被表格内容遮挡 */
  :deep(.el-image__preview) {
    z-index: 9999;
  }
}

.product-name {
  font-size: 0.875rem;
  color: var(--shop-foreground, #1e293b);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---- 图片预览弹窗 ---- */
.img-preview-modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: zoom-out;
}

.preview-img {
  max-width: 90%;
  max-height: 90%;
  object-fit: contain;
  border-radius: 8px;
}

/* ---- 响应式 ---- */
@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  .detail-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
