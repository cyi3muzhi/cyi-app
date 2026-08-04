<template>
  <div class="order-detail-page">
    <!-- 全屏图片预览 -->
    <div v-if="previewImage" class="od-preview" @click="closePreview">
      <div class="od-preview-bg" />
      <img :src="previewImage" class="od-preview-img" />
      <button class="od-preview-close" @click="closePreview">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
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
      <!-- ===== 顶部状态横幅 ===== -->
      <div class="status-banner" :style="bannerStyle">
        <div class="banner-top">
          <el-button text class="back-btn" @click="goBack">
            <el-icon><ArrowLeft /></el-icon> 返回列表
          </el-button>
          <div class="banner-actions">
            <el-button
              v-if="order.status === 'paid'"
              type="primary"
              class="action-btn"
              @click="openShipDialog"
            >
              立即发货
            </el-button>
            <el-button
              v-if="order.status === 'pending'"
              plain
              class="action-btn action-btn--danger"
              @click="handleCancel"
            >
              取消订单
            </el-button>
          </div>
        </div>
        <div class="banner-body">
          <div class="banner-status-row">
            <span class="status-dot" />
            <span class="banner-status-text">{{ order.statusText }}</span>
          </div>
          <div class="banner-order-no">{{ order.orderNo }}</div>
          <div class="banner-time">
            <el-icon :size="14"><Clock /></el-icon>
            <span>{{ order.createdAt }}</span>
          </div>
        </div>
      </div>

      <!-- ===== 信息卡片双栏布局 ===== -->
      <div class="info-grid-row">
        <div class="info-card" style="--i: 0">
          <div class="info-card-accent" />
          <div class="card-header">
            <span class="card-header-icon"><InfoFilled /></span>
            <span>订单信息</span>
          </div>
          <div class="card-body">
            <div class="field-row">
              <div class="field">
                <span class="field-label">订单号</span>
                <span class="field-value mono">{{ order.orderNo }}</span>
              </div>
              <div class="field">
                <span class="field-label">下单时间</span>
                <span class="field-value">{{ order.createdAt }}</span>
              </div>
            </div>
            <div class="field-row">
              <div class="field">
                <span class="field-label">支付方式</span>
                <span class="field-value">{{ order.payment || '-' }}</span>
              </div>
              <div class="field">
                <span class="field-label">订单状态</span>
                <span class="field-value">
                  <el-tag :type="statusTagType(order.status)" size="small" effect="light">
                    {{ order.statusText }}
                  </el-tag>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="info-card" style="--i: 1">
          <div class="info-card-accent info-card-accent--2" />
          <div class="card-header">
            <span class="card-header-icon"><UserFilled /></span>
            <span>买家信息</span>
          </div>
          <div class="card-body">
            <div class="field-row">
              <div class="field">
                <span class="field-label">买家姓名</span>
                <span class="field-value">{{ order.userName || '-' }}</span>
              </div>
              <div class="field">
                <span class="field-label">买家 ID</span>
                <span class="field-value mono">{{ order.id }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== 收货与物流信息卡片 ===== -->
      <div class="info-card logistics-card" style="--i: 1.5">
        <div class="info-card-accent info-card-accent--3" />
        <div class="card-header">
          <span class="card-header-icon"><Van /></span>
          <span>收货与物流信息</span>
          <span v-if="order.logisticsCompany" class="card-subtitle">已发货</span>
        </div>
        <div class="card-body">
          <div class="field-row">
            <div class="field">
              <span class="field-label">收货人</span>
              <span class="field-value">{{ order.receiver || '-' }}</span>
            </div>
            <div class="field">
              <span class="field-label">联系电话</span>
              <span class="field-value">{{ order.phone || '-' }}</span>
            </div>
          </div>
          <div class="field-row">
            <div class="field field--wide">
              <span class="field-label">收货地址</span>
              <span class="field-value">{{ order.address || '-' }}</span>
            </div>
          </div>
          <div class="field-row">
            <div class="field">
              <span class="field-label">物流公司</span>
              <span class="field-value">{{ order.logisticsCompany || '未发货' }}</span>
            </div>
            <div class="field">
              <span class="field-label">快递单号</span>
              <span class="field-value mono">{{ order.logisticsNo || '-' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== 商品明细卡片 ===== -->
      <div class="product-card" style="--i: 2">
        <div class="card-header">
          <span class="card-header-icon"><Goods /></span>
          <span>商品明细</span>
          <span class="card-subtitle">共 {{ order.products.length }} 件商品</span>
        </div>

        <el-table
          :data="order.products"
          size="small"
          style="width: 100%"
          stripe
          :show-header="true"
        >
          <el-table-column label="商品" min-width="220">
            <template #default="{ row, $index }">
              <div class="product-cell">
                <img
                  v-if="row.image"
                  :src="row.image"
                  class="product-img"
                  @error="onImgError($event)"
                  @click="openPreview(row.image)"
                  style="cursor: zoom-in"
                />
                <span v-else class="product-avatar" :style="{ background: nameColor(row.name) }">
                  {{ row.name.charAt(0) }}
                </span>
                <div class="product-info">
                  <span class="product-name">{{ row.name }}</span>
                  <span class="product-spec">{{ row.spec || '默认规格' }}</span>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="单价" width="100" align="right">
            <template #default="{ row }">
              <span class="price-text">&yen;{{ (row.price / 100).toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="数量" width="70" align="center">
            <template #default="{ row }">
              <span class="qty-badge">×{{ row.qty }}</span>
            </template>
          </el-table-column>
          <el-table-column label="小计" width="110" align="right">
            <template #default="{ row }">
              <span class="subtotal-text">&yen;{{ ((row.price * row.qty) / 100).toFixed(2) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- 金额汇总卡 -->
        <div class="od-summary">
          <div class="od-summary-body">
            <div class="od-summary-row">
              <span class="od-summary-label">商品总额</span>
              <span class="od-summary-value">&yen;{{ (totalProductAmount / 100).toFixed(2) }}</span>
            </div>
            <div class="od-summary-divider" />
            <div class="od-summary-row od-summary-total">
              <span class="od-summary-label">实付金额</span>
              <span class="od-summary-price">&yen;{{ (order.amount / 100).toFixed(2) }}</span>
            </div>
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
import { ArrowLeft, InfoFilled, Goods, Loading, Clock, UserFilled, Van } from '@element-plus/icons-vue'
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

/** 根据状态返回横幅渐变背景 */
const bannerStyle = computed(() => {
  const gradients: Record<string, string> = {
    pending:   'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
    paid:      'linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%)',
    shipped:   'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
    completed: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    cancelled: 'linear-gradient(135deg, #94a3b8 0%, #64748b 100%)',
    refunding: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
    refunded:  'linear-gradient(135deg, #94a3b8 0%, #64748b 100%)',
  }
  return { background: gradients[order.value?.status || ''] || gradients.cancelled }
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
/* ========================================
   订单详情页
   ========================================= */
@keyframes odFadeUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes odBannerIn {
  from { opacity: 0; transform: translateY(-16px) scale(0.97); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes odScaleIn {
  from { opacity: 0; transform: scale(0.92); }
  to { opacity: 1; transform: scale(1); }
}

.order-detail-page {
  padding: 24px;
  width: 100%;
  animation: odFadeUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.page-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--shop-muted-foreground, #64748b);
}

/* ===== 顶部状态横幅 ===== */
.status-banner {
  position: relative;
  border-radius: 14px;
  padding: 24px 28px;
  margin-bottom: 20px;
  color: #fff;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  animation: odBannerIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.banner-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}
.back-btn { color: rgba(255, 255, 255, 0.85) !important; font-size: 0.8125rem; }
.back-btn:hover { color: #fff !important; background: rgba(255, 255, 255, 0.15) !important; }
.banner-actions { display: flex; gap: 8px; }
.action-btn {
  border: 1px solid rgba(255, 255, 255, 0.4) !important;
  color: #fff !important;
  background: rgba(255, 255, 255, 0.15) !important;
  backdrop-filter: blur(4px);
  font-weight: 500;
}
.action-btn:hover { background: rgba(255, 255, 255, 0.28) !important; border-color: rgba(255, 255, 255, 0.6) !important; }
.action-btn--danger { background: rgba(239, 68, 68, 0.35) !important; border-color: rgba(239, 68, 68, 0.5) !important; }
.action-btn--danger:hover { background: rgba(239, 68, 68, 0.55) !important; }

.banner-body { position: relative; z-index: 1; }
.banner-status-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.status-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: #fff; box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.3); }
.banner-status-text { font-size: 0.9375rem; font-weight: 600; letter-spacing: 0.02em; }
.banner-order-no { font-size: 1.375rem; font-weight: 700; font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); margin-bottom: 6px; letter-spacing: 0.01em; }
.banner-time { display: flex; align-items: center; gap: 6px; font-size: 0.8125rem; opacity: 0.85; }

/* ===== 信息卡片 ===== */
.info-grid-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }

.info-card {
  background: #fff;
  border: 1px solid var(--shop-border, #e8e8e8);
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  transition: box-shadow 0.25s ease, transform 0.25s ease;
  animation: odFadeUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: calc(var(--i, 0) * 0.1s);
}
.info-card:hover { box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); transform: translateY(-1px); }

.info-card-accent {
  position: absolute; top: 0; left: 0; width: 4px; height: 100%;
  background: var(--shop-primary, #3b82f6);
  border-radius: 0 2px 2px 0;
}
.info-card-accent--2 { background: #10b981; }
.info-card-accent--3 { background: #f59e0b; }
/* 单张卡片（收货与物流信息）独立成行 */
.logistics-card { margin-bottom: 16px; }

.card-header {
  display: flex; align-items: center; gap: 8px;
  padding: 14px 20px 14px 24px;
  font-size: 0.875rem; font-weight: 600;
  color: var(--shop-foreground, #1e293b);
  border-bottom: 1px solid var(--shop-border, #e8e8e8);
  background: #fafbfc;
}
.card-header-icon { display: flex; align-items: center; color: var(--shop-primary, #3b82f6); font-size: 1rem; }
.card-subtitle { margin-left: auto; font-size: 0.75rem; font-weight: 400; color: var(--shop-muted-foreground, #94a3b8); }
.card-body { padding: 16px 20px 16px 24px; }

.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 12px; }
.field-row:last-child { margin-bottom: 0; }
.field { display: flex; flex-direction: column; gap: 3px; }
.field--wide { grid-column: 1 / -1; }
.field-label { font-size: 0.75rem; color: var(--shop-muted-foreground, #94a3b8); font-weight: 500; text-transform: uppercase; letter-spacing: 0.03em; }
.field-value { font-size: 0.875rem; color: var(--shop-foreground, #1e293b); font-weight: 500; }
.field-value.mono { font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); font-size: 0.8125rem; word-break: break-all; }

/* ===== 商品明细 ===== */
.product-card {
  background: #fff;
  border: 1px solid var(--shop-border, #e8e8e8);
  border-radius: 12px;
  overflow: hidden;
  transition: box-shadow 0.25s ease;
  animation: odFadeUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: 0.2s;
}
.product-card:hover { box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06); }

.product-cell { display: flex; align-items: center; gap: 12px; }
.product-img {
  width: 44px; height: 44px; border-radius: 8px; object-fit: cover;
  flex-shrink: 0; border: 1px solid var(--shop-border, #e8e8e8);
  cursor: zoom-in; transition: transform 0.15s;
}
.product-img:hover { transform: scale(1.1); }
.product-avatar { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; border-radius: 8px; color: #fff; font-size: 0.8125rem; font-weight: 600; flex-shrink: 0; }
.product-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.product-name { font-size: 0.875rem; font-weight: 500; color: var(--shop-foreground, #1e293b); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.product-spec { font-size: 0.75rem; color: var(--shop-muted-foreground, #94a3b8); }
.price-text { font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); font-size: 0.8125rem; color: var(--shop-foreground, #1e293b); }
.qty-badge { display: inline-block; padding: 1px 8px; border-radius: 4px; background: var(--shop-muted, #f1f5f9); color: var(--shop-muted-foreground, #64748b); font-size: 0.8125rem; font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); }
.subtotal-text { font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); font-size: 0.875rem; font-weight: 600; color: var(--shop-foreground, #1e293b); }

/* ===== 金额汇总 ===== */
.od-summary {
  margin: 16px 20px 20px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border: 1px solid var(--shop-border, #e8e8e8);
  border-radius: 10px;
  overflow: hidden;
  animation: odFadeUp 0.35s ease both;
  animation-delay: 0.3s;
}
.od-summary-body { padding: 16px 24px; display: flex; flex-direction: column; align-items: flex-end; gap: 6px; }
.od-summary-row { display: flex; justify-content: space-between; width: 280px; align-items: center; }
.od-summary-label { font-size: 0.8125rem; color: var(--shop-muted-foreground, #64748b); }
.od-summary-value { font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); font-size: 0.9375rem; font-weight: 600; color: var(--shop-foreground, #1e293b); }
.od-summary-divider { width: 280px; height: 1px; background: var(--shop-border, #e8e8e8); }
.od-summary-price { font-family: var(--shop-font-mono, 'JetBrains Mono', monospace); font-size: 1.375rem; font-weight: 700; color: var(--shop-destructive, #ef4444); letter-spacing: -0.01em; }

/* ===== 图片预览 ===== */
.od-preview { position: fixed; inset: 0; z-index: 9999; display: flex; align-items: center; justify-content: center; }
.od-preview-bg { position: absolute; inset: 0; background: rgba(0, 0, 0, 0.75); backdrop-filter: blur(6px); }
.od-preview-img { position: relative; max-width: 80vw; max-height: 80vh; border-radius: 12px; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4); animation: odScaleIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1); }
.od-preview-close { position: absolute; top: 24px; right: 24px; width: 40px; height: 40px; border-radius: 10px; border: none; background: rgba(255, 255, 255, 0.15); color: #fff; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: background 0.15s; }
.od-preview-close:hover { background: rgba(255, 255, 255, 0.25); }

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .info-grid-row { grid-template-columns: 1fr; }
  .banner-order-no { font-size: 1.125rem; }
  .od-summary-row, .od-summary-divider { width: 100%; }
  .od-summary-price { font-size: 1.125rem; }
}
</style>
