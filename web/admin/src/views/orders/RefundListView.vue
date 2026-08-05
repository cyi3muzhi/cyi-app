<template>
  <div class="refund-page">
    <!-- ===== 页面头部 ===== -->
    <div class="page-header">
      <div>
        <h2 class="page-title">售后管理</h2>
        <p class="page-subtitle">处理用户退款申请，审核通过后退款并恢复库存，拒绝后订单恢复原状态</p>
      </div>
      <div class="header-right">
        <el-tag v-if="pendingCount > 0" type="warning" effect="light" round>
          待审核 {{ pendingCount }} 单
        </el-tag>
      </div>
    </div>

    <!-- ===== 筛选区 ===== -->
    <div class="filter-bar">
      <el-select v-model="query.status" placeholder="退款状态" clearable class="filter-item status-select" @change="onSearch">
        <el-option label="待审核" value="1" />
        <el-option label="已通过" value="2" />
        <el-option label="已拒绝" value="3" />
      </el-select>
      <el-input
        v-model="query.keyword"
        placeholder="输入订单号搜索"
        clearable
        class="filter-item keyword-input"
        @keyup.enter="onSearch"
        @clear="onSearch"
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="onSearch">查询</el-button>
      <el-button @click="onReset">重置</el-button>
    </div>

    <!-- ===== 退款单表格 ===== -->
    <div class="table-card">
      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column label="退款单号" width="110">
          <template #default="{ row }">RF{{ String(row.refundId).padStart(6, '0') }}</template>
        </el-table-column>
        <el-table-column prop="orderNo" label="订单号" min-width="180" show-overflow-tooltip />
        <el-table-column label="用户" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.userName || ('UID ' + row.userId) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">
            <span class="refund-money">¥{{ (row.refundAmount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="退款原因" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="light" round>{{ row.statusText }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="申请时间" width="160" />
        <el-table-column label="审核人" width="100" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'muted-text': !row.auditor }">{{ row.auditor || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button type="success" size="small" link @click="onApprove(row)">通过</el-button>
              <el-button type="danger" size="small" link @click="onReject(row)">拒绝</el-button>
            </template>
            <el-popover v-else-if="row.status === 3 && row.rejectReason" :width="280" trigger="hover">
              <template #reference>
                <span class="reject-link">拒绝原因</span>
              </template>
              <div class="reject-popover">
                <div class="reject-title">拒绝原因</div>
                <div class="reject-content">{{ row.rejectReason }}</div>
              </div>
            </el-popover>
            <span v-else class="muted-text">—</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!items.length && !loading" class="empty-state">
        <el-icon :size="48" color="#d0d4dc"><CircleClose /></el-icon>
        <div class="empty-title">暂无退款记录</div>
        <div class="empty-desc">用户申请退款后将在这里显示</div>
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
          @current-change="loadList"
          @size-change="onSizeChange"
        />
      </div>
    </div>

    <!-- ===== 拒绝原因弹窗 ===== -->
    <el-dialog v-model="rejectVisible" title="拒绝退款" width="440px" :close-on-click-modal="false">
      <div class="dialog-tip">拒绝后订单将恢复到申请前的状态（已支付/待发货），请填写拒绝原因告知用户。</div>
      <el-input
        v-model="rejectReason"
        type="textarea"
        :rows="3"
        maxlength="200"
        show-word-limit
        placeholder="请输入拒绝原因（必填）"
      />
      <template #footer>
        <el-button @click="rejectVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitting" @click="confirmReject">确认拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 售后管理页面
 * 展示退款/售后单列表，支持状态筛选、订单号搜索、审核通过/拒绝
 * 数据来源：GET /admin/v1/refunds；审核：POST /admin/v1/refunds/{refund_id}/audit
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, CircleClose } from '@element-plus/icons-vue'
import { listRefunds, auditRefund } from '@/api/order'
import type { RefundItem } from '@/types/order'

/** 查询条件 */
const query = reactive({ status: '', keyword: '' })
/** 列表数据 */
const items = ref<RefundItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)

/** 待审核数量（从列表首屏统计，简单展示） */
const pendingCount = computed(() => items.value.filter(i => i.status === 1).length)

/**
 * 查询（重置到第一页）
 */
function onSearch() {
  page.value = 1
  loadList()
}

/**
 * 重置筛选条件
 */
function onReset() {
  query.status = ''
  query.keyword = ''
  onSearch()
}

/**
 * 切换每页条数
 */
function onSizeChange() {
  page.value = 1
  loadList()
}

/**
 * 加载退款单列表
 */
async function loadList() {
  loading.value = true
  try {
    const res = await listRefunds({
      page: page.value,
      page_size: pageSize.value,
      ...(query.status ? { status: query.status } : {}),
      ...(query.keyword ? { keyword: query.keyword } : {})
    })
    items.value = res.items || []
    total.value = res.total || 0
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || e?.message || '加载退款记录失败')
  } finally {
    loading.value = false
  }
}

/**
 * 审核通过（二次确认）
 * @param row - 退款单
 */
function onApprove(row: RefundItem) {
  ElMessageBox.confirm(
    `确定通过订单 ${row.orderNo} 的退款申请吗？通过后将向用户退款 ¥${(row.refundAmount / 100).toFixed(2)} 并恢复库存。`,
    '审核通过',
    { confirmButtonText: '确认通过', cancelButtonText: '取消', type: 'warning' }
  ).then(async () => {
    try {
      await auditRefund(row.refundId, true)
      ElMessage.success('已通过退款，订单退款成功')
      loadList()
    } catch (e: any) {
      ElMessage.error(e?.response?.data?.message || e?.message || '审核失败，请重试')
    }
  }).catch(() => {})
}

// ---- 拒绝审核 ----
const rejectVisible = ref(false)
const rejectReason = ref('')
const submitting = ref(false)
const currentRow = ref<RefundItem | null>(null)

/**
 * 打开拒绝弹窗
 * @param row - 退款单
 */
function onReject(row: RefundItem) {
  currentRow.value = row
  rejectReason.value = ''
  rejectVisible.value = true
}

/**
 * 确认拒绝（校验拒绝原因必填）
 */
async function confirmReject() {
  if (!currentRow.value) return
  if (!rejectReason.value.trim()) {
    ElMessage.warning('请填写拒绝原因')
    return
  }
  submitting.value = true
  try {
    await auditRefund(currentRow.value.refundId, false, rejectReason.value.trim())
    ElMessage.success('已拒绝退款，订单已恢复')
    rejectVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || e?.message || '审核失败，请重试')
  } finally {
    submitting.value = false
  }
}

/**
 * 状态对应的标签类型
 * @param status - 1-待审核 2-已通过 3-已拒绝
 */
function statusTagType(status: number): '' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<number, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    1: 'warning',
    2: 'success',
    3: 'danger'
  }
  return map[status] || 'info'
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
/* ============================================
   售后管理页面样式
   ============================================ */

/* ---- 页面头部 ---- */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}
.page-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #1f2937;
}
.page-subtitle {
  margin-top: 6px;
  font-size: 0.8125rem;
  color: #9ca3af;
}
.header-right {
  padding-top: 4px;
}

/* ---- 筛选区 ---- */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  margin-bottom: 16px;
  background: #fff;
  border: 1px solid #eef0f3;
  border-radius: 12px;
}
.filter-item.status-select {
  width: 140px;
}
.filter-item.keyword-input {
  width: 260px;
}

/* ---- 表格卡片 ---- */
.table-card {
  background: #fff;
  border: 1px solid #eef0f3;
  border-radius: 12px;
  padding: 8px 16px 16px;
}
.refund-money {
  color: #ef4444;
  font-weight: 600;
}
.muted-text {
  color: #9ca3af;
}
.reject-link {
  color: #ef4444;
  font-size: 0.8125rem;
  cursor: pointer;
}
.reject-popover .reject-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 8px;
}
.reject-popover .reject-content {
  font-size: 0.8125rem;
  color: #4b5563;
  line-height: 1.6;
  white-space: pre-wrap;
}

/* ---- 分页 ---- */
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}

/* ---- 空状态 ---- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 56px 0;
}
.empty-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #6b7280;
}
.empty-desc {
  font-size: 0.8125rem;
  color: #9ca3af;
}

/* ---- 拒绝弹窗 ---- */
.dialog-tip {
  font-size: 0.8125rem;
  color: #6b7280;
  line-height: 1.6;
  margin-bottom: 12px;
  padding: 10px 12px;
  background: #fef9f0;
  border: 1px solid #fde8cd;
  border-radius: 8px;
}
</style>
