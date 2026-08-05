<template>
  <!-- 库存预警页：统计卡片 + 预警列表 + 补货操作 -->
  <div class="stock-alert-page">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card stat-all" @click="switchTab('')">
        <div class="stat-icon"><el-icon :size="22"><Box /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.all }}</div>
          <div class="stat-label">全部商品</div>
        </div>
      </div>
      <div class="stat-card stat-empty" @click="switchTab('empty')">
        <div class="stat-icon"><el-icon :size="22"><CircleClose /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.empty }}</div>
          <div class="stat-label">缺货商品</div>
        </div>
      </div>
      <div class="stat-card stat-low" @click="switchTab('low')">
        <div class="stat-icon"><el-icon :size="22"><Warning /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.low }}</div>
          <div class="stat-label">低库存预警（≤ {{ threshold }}）</div>
        </div>
      </div>
      <div class="stat-card stat-normal" @click="switchTab('')">
        <div class="stat-icon"><el-icon :size="22"><CircleCheck /></el-icon></div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.normal }}</div>
          <div class="stat-label">库存充足</div>
        </div>
      </div>
    </div>

    <!-- 阈值设置条 -->
    <div class="threshold-bar">
      <div class="threshold-left">
        <el-icon :size="16" color="#f59e0b"><Bell /></el-icon>
        <span class="threshold-text">库存预警阈值</span>
        <el-input-number
          v-model="threshold"
          :min="0"
          :max="99999"
          size="small"
          :controls="false"
          class="threshold-input"
          @change="onThresholdChange"
        />
        <span class="threshold-hint">当商品可用库存 ≤ 阈值时将被标记为低库存预警</span>
      </div>
      <div class="threshold-right">
        <el-button size="small" text type="primary" @click="saveThreshold">
          {{ thresholdSaved ? '阈值已保存' : '保存阈值' }}
        </el-button>
      </div>
    </div>

    <!-- 筛选工具栏 -->
    <div class="toolbar">
      <div class="search-input-wrap">
        <el-icon class="search-icon"><Search /></el-icon>
        <input v-model="filters.keyword" type="text" placeholder="搜索商品名称..."
          class="search-input" @input="onSearchInput" />
      </div>
      <el-cascader v-model="filters.category_id" :options="categoryTree"
        :props="{ value: 'id', label: 'name', children: 'children', expandTrigger: 'hover', emitPath: false }"
        placeholder="全部分类" clearable class="toolbar-cascader" @change="handleFilter" />
      <el-select v-model="filters.status" placeholder="全部状态" clearable class="toolbar-status" @change="handleFilter">
        <el-option label="上架" value="上架" />
        <el-option label="下架" value="下架" />
      </el-select>
    </div>

    <!-- Tabs -->
    <div class="tab-bar-wrap">
      <div class="tab-bar-inner" ref="barInner">
        <div class="tab-slider" :style="sliderStyle" />
        <div v-for="tab in tabItems" :key="tab.key"
          :ref="el => setTabRef(tab.key, el)"
          :class="['tab-item', { active: currentTab === tab.key }]"
          @click="switchTab(tab.key)">
          <el-icon :size="16">
            <component :is="tab.icon" />
          </el-icon>
          {{ tab.label }}
          <span v-if="tab.count > 0" class="tab-count">{{ tab.count }}</span>
        </div>
      </div>
    </div>

    <!-- 表格 -->
    <el-table :data="products" v-loading="loading" stripe style="width:100%">
      <template #empty>
        <el-empty :description="loading ? '加载中...' : '暂无商品'" />
      </template>
      <el-table-column label="图片" width="80">
        <template #default="{ row }">
          <div class="product-img-cell">
            <img :src="row.image" class="product-img" @click="previewImage(row.image)"
              @error="onImgError($event)" />
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="商品名称" min-width="200" show-overflow-tooltip />
      <el-table-column label="分类" width="130">
        <template #default="{ row }">
          {{ getCategoryName((row as any).categoryId || row.category_id) }}
        </template>
      </el-table-column>
      <el-table-column label="价格" width="110">
        <template #default="{ row }">¥{{ (row.price / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="库存" width="150">
        <template #default="{ row }">
          <div class="stock-cell">
            <el-progress
              :percentage="stockPercent(row.stock)"
              :stroke-width="6"
              :color="stockBarColor(row.stock)"
              class="stock-bar"
              :show-text="false"
            />
            <span :class="stockClass(row.stock)" class="stock-num">{{ row.stock }}</span>
            <el-tag v-if="row.stock === 0" size="small" type="danger" effect="dark" round>缺货</el-tag>
            <el-tag v-else-if="row.stock <= threshold" size="small" type="warning" effect="plain" round>预警</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="销量" width="90">
        <template #default="{ row }">{{ row.sales }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === '上架' ? 'success' : 'info'" size="small" effect="plain" round>
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openRestock(row)">补货</el-button>
          <el-button link @click="$router.push(`/products/edit/${row.id}`)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize"
        :total="totalCount" :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadProducts" @size-change="loadProducts" />
    </div>

    <!-- 补货对话框 -->
    <el-dialog v-model="restockVisible" title="补货 / 调整库存" :width="hasSku ? '580px' : '420px'" :close-on-click-modal="false">
      <div class="restock-body">
        <div class="restock-product">
          <img :src="restockRow?.image" class="restock-img" @error="onImgError($event)" />
          <div class="restock-info">
            <div class="restock-name">{{ restockRow?.name }}</div>
            <div class="restock-current">
              当前总库存：<span :class="stockClass(restockRow?.stock ?? 0)" class="stock-num">{{ restockRow?.stock ?? 0 }}</span>
            </div>
          </div>
        </div>

        <!-- 无规格商品：直接设置总库存 -->
        <div v-if="!hasSku" class="restock-form">
          <div class="restock-form-label">调整后总库存</div>
          <el-input-number v-model="restockTotal" :min="0" :max="999999" :step="10" class="restock-input" />
          <div class="restock-hint">无规格商品，直接设置总库存即可</div>
        </div>

        <!-- 有规格商品：按 SKU 明细逐个设置 -->
        <div v-else class="restock-sku-wrap">
          <div class="restock-form">
            <div class="restock-form-label">快捷设置总库存（均分到各 SKU，可再逐个微调）</div>
            <el-input-number v-model="skuTotal" :min="0" :max="999999" :step="10" class="restock-input" @change="onSkuTotalChange" />
          </div>
          <div class="sku-list">
            <div v-for="r in skuRows" :key="r.sku_id" class="sku-row">
              <div class="sku-info">
                <div class="sku-spec">{{ r.specText || r.sku_code || `SKU #${r.sku_id}` }}</div>
                <div class="sku-current">当前 {{ r.current }}</div>
              </div>
              <el-input-number v-model="r.newStock" :min="0" :max="999999" size="small" class="sku-input" />
            </div>
          </div>
          <div class="restock-hint">SKU 库存合计：{{ skuTotalByRows }} 件</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="restockVisible = false">取消</el-button>
        <el-button type="primary" :loading="restockSaving" @click="confirmRestock">确认补货</el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <teleport to="body">
      <div v-if="previewUrl" class="image-preview-overlay" @click.self="previewUrl = ''">
        <img :src="previewUrl" class="image-preview-full" @click="previewUrl = ''" />
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Box, CircleClose, Warning, CircleCheck, Bell, Search, Grid, Upload, Download
} from '@element-plus/icons-vue'
import { listProducts, getProduct, updateProductStock, updateProductSkuStocks } from '../../api/product'
import { listCategoryTree } from '../../api/category'
import { getSettings, updateSettings } from '../../api/settings'
import type { ProductItem } from '../../types/product'

/** 当前筛选 Tab：''=全部 low=低库存预警 empty=缺货 */
const currentTab = ref('')
const threshold = ref(10)
const thresholdSaved = ref(true)
const products = ref<ProductItem[]>([])
const totalCount = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const categoryTree = ref<any[]>([])
const previewUrl = ref('')
const filters = ref<{ keyword: string; category_id?: number; status: string }>({
  keyword: '', category_id: undefined, status: ''
})

/** 统计卡片数据 */
const stats = ref({ all: 0, empty: 0, low: 0, normal: 0 })

const tabItems = ref([
  { key: '', label: '全部', icon: Grid, count: 0 },
  { key: 'empty', label: '缺货', icon: CircleClose, count: 0 },
  { key: 'low', label: '低库存', icon: Warning, count: 0 },
])

/** 滑动指示器 */
const barInner = ref<HTMLElement | null>(null)
const tabRefs = ref<Record<string, HTMLElement>>({})
const sliderStyle = ref({ width: '0px', left: '0px', opacity: 0 })

function setTabRef(key: string, el: any) {
  if (el) tabRefs.value[key] = el
}

async function updateSlider() {
  await nextTick()
  const activeEl = tabRefs.value[currentTab.value]
  if (activeEl && barInner.value) {
    const containerRect = barInner.value.getBoundingClientRect()
    const elRect = activeEl.getBoundingClientRect()
    sliderStyle.value = {
      width: `${elRect.width}px`,
      left: `${elRect.left - containerRect.left}px`,
      opacity: 1,
    }
  }
}

/** 切换 Tab 并刷新列表 */
function switchTab(tab: string) {
  currentTab.value = tab
  currentPage.value = 1
  updateSlider()
  loadProducts()
}

/** 库存进度条百分比（按阈值的 3 倍为满刻度，避免大库存全部满条） */
function stockPercent(stock: number): number {
  if (stock <= 0) return 0
  const scale = Math.max(threshold.value * 3, 10)
  return Math.min(Math.round((stock / scale) * 100), 100)
}

/** 库存进度条颜色 */
function stockBarColor(stock: number): string {
  if (stock <= 0) return '#ef4444'
  if (stock <= threshold.value) return '#f59e0b'
  return '#10b981'
}

/** 库存数字样式 */
function stockClass(stock: number): string {
  if (stock <= 0) return 'stock-empty'
  if (stock <= threshold.value) return 'stock-low'
  return 'stock-ok'
}

/** 搜索防抖 */
let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { currentPage.value = 1; loadProducts() }, 300)
}

function handleFilter() {
  currentPage.value = 1
  loadProducts()
}

/** 根据当前 Tab 计算列表查询参数 */
function buildListParams() {
  const params: any = {
    page: currentPage.value,
    page_size: pageSize.value,
    keyword: filters.value.keyword || undefined,
    category_id: filters.value.category_id || undefined,
    status: filters.value.status || undefined,
  }
  if (currentTab.value === 'empty') {
    params.stock_eq = 0
  } else if (currentTab.value === 'low') {
    params.stock_lte = threshold.value
  }
  return params
}

/** 加载预警列表 */
async function loadProducts() {
  loading.value = true
  try {
    const res = await listProducts(buildListParams())
    products.value = res.items || []
    totalCount.value = res.total || 0
  } catch {
    products.value = []
    totalCount.value = 0
  } finally {
    loading.value = false
  }
}

/** 加载统计卡片数据（page_size=1 只取 total，减少传输） */
async function loadStats() {
  try {
    const [allRes, emptyRes, lowRes] = await Promise.all([
      listProducts({ page: 1, page_size: 1 }),
      listProducts({ page: 1, page_size: 1, stock_eq: 0 }),
      listProducts({ page: 1, page_size: 1, stock_lte: threshold.value }),
    ])
    stats.value.all = allRes.total || 0
    stats.value.empty = emptyRes.total || 0
    stats.value.low = lowRes.total || 0
    stats.value.normal = Math.max(stats.value.all - stats.value.low, 0)
    // 同步 Tab 徽标数量
    tabItems.value[0].count = stats.value.all
    tabItems.value[1].count = stats.value.empty
    tabItems.value[2].count = stats.value.low
  } catch {
    // 统计加载失败不阻塞页面
  }
}

/** 阈值变更时刷新低库存统计和列表 */
function onThresholdChange() {
  thresholdSaved.value = false
  loadStats()
  if (currentTab.value === 'low') loadProducts()
}

/** 保存预警阈值到系统设置 */
async function saveThreshold() {
  try {
    await updateSettings({ stockAlertThreshold: threshold.value })
    thresholdSaved.value = true
    ElMessage.success('库存预警阈值已保存')
    loadStats()
  } catch {
    // 错误已由拦截器统一提示
  }
}

/** 补货对话框状态 */
const restockVisible = ref(false)
const restockSaving = ref(false)
const restockRow = ref<ProductItem | null>(null)
const restockTotal = ref(0)
/** 是否为有规格商品（按 SKU 明细补货） */
const hasSku = ref(false)
/** SKU 明细行 */
const skuRows = ref<{ sku_id: number; sku_code: string; specText: string; current: number; newStock: number }[]>([])
/** SKU 模式快捷总库存输入 */
const skuTotal = ref(0)

/** SKU 库存合计（各 SKU 输入之和） */
const skuTotalByRows = computed(() => skuRows.value.reduce((sum, r) => sum + r.newStock, 0))

/** 将 SKU 的 spec map 格式化为可读文本，如 "红色 / M" */
function specText(spec: Record<string, string>): string {
  if (!spec || !Object.keys(spec).length) return ''
  return Object.values(spec).join(' / ')
}

/** 打开补货对话框：加载商品详情，有 SKU 则进入明细模式 */
async function openRestock(row: ProductItem) {
  restockRow.value = row
  restockVisible.value = true
  restockSaving.value = false
  try {
    const res = await getProduct(row.id)
    const product = res.product
    if (product.skus && product.skus.length > 0) {
      // 有规格商品：展示每个 SKU 的当前库存，支持逐个设置
      hasSku.value = true
      skuRows.value = product.skus.map(s => ({
        sku_id: s.id,
        sku_code: s.sku_code,
        specText: specText(s.spec),
        current: s.stock,
        newStock: s.stock,
      }))
      skuTotal.value = skuRows.value.reduce((sum, r) => sum + r.newStock, 0)
    } else {
      // 无规格商品：直接设置总库存
      hasSku.value = false
      restockTotal.value = row.stock
    }
  } catch {
    // 加载详情失败时退化为无规格模式，避免阻塞操作
    hasSku.value = false
    restockTotal.value = row.stock
  }
}

/** SKU 快捷总库存变化：均分到各 SKU（余数计入第一条），可再逐个微调 */
function onSkuTotalChange() {
  const n = skuRows.value.length
  if (!n) return
  const each = Math.floor(skuTotal.value / n)
  const remainder = skuTotal.value % n
  skuRows.value.forEach((r, i) => {
    r.newStock = each + (i === 0 ? remainder : 0)
  })
}

/** 确认补货：有 SKU 按明细提交，无 SKU 提交总库存 */
async function confirmRestock() {
  if (!restockRow.value) return
  restockSaving.value = true
  try {
    if (hasSku.value) {
      const items = skuRows.value.map(r => ({ sku_id: r.sku_id, total: r.newStock }))
      await updateProductSkuStocks(restockRow.value.id, items)
    } else {
      await updateProductStock(restockRow.value.id, restockTotal.value)
    }
    ElMessage.success(`「${restockRow.value.name}」库存已更新`)
    restockVisible.value = false
    loadProducts()
    loadStats()
  } catch {
    // 错误已由拦截器统一提示
  } finally {
    restockSaving.value = false
  }
}

/** 图片加载失败时显示 N/A 占位 */
function onImgError(e: Event) {
  const el = e.target as HTMLElement
  el.style.display = 'none'
  const parent = el.parentElement
  if (parent && !parent.querySelector('.img-placeholder-text')) {
    const span = document.createElement('span')
    span.className = 'img-placeholder-text'
    span.textContent = 'N/A'
    parent.appendChild(span)
  }
}

/** 预览商品大图 */
function previewImage(url: string) {
  previewUrl.value = url
}

/** 根据 category_id 在分类树中递归查找分类名称 */
function getCategoryName(categoryId: number): string {
  if (!categoryId || !categoryTree.value.length) return '-'
  function find(nodes: any[]): string | undefined {
    for (const node of nodes) {
      if (node.id === categoryId) return node.name
      if (node.children?.length) {
        const found = find(node.children)
        if (found) return found
      }
    }
    return undefined
  }
  return find(categoryTree.value) || '-'
}

onMounted(async () => {
  try {
    const settings = await getSettings()
    threshold.value = settings.stockAlertThreshold ?? 10
  } catch {
    // 读取失败使用默认阈值 10
  }
  try {
    const res = await listCategoryTree()
    categoryTree.value = [{ id: '0', name: '根目录', children: res.items }]
  } catch {
    categoryTree.value = []
  }
  await loadStats()
  loadProducts()
  nextTick(() => updateSlider())
})
</script>

<style scoped>
.stock-alert-page { padding: 24px; }

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  border-radius: var(--shop-radius-lg, 0.75rem);
  background: #fff;
  border: 1px solid var(--shop-border, #e8e8e8);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.08);
}
.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.stat-all .stat-icon { background: #eff6ff; color: #3b82f6; }
.stat-empty .stat-icon { background: #fef2f2; color: #ef4444; }
.stat-low .stat-icon { background: #fffbeb; color: #f59e0b; }
.stat-normal .stat-icon { background: #ecfdf5; color: #10b981; }
.stat-value { font-size: 26px; font-weight: 700; line-height: 1.2; color: #1e293b; }
.stat-label { font-size: 13px; color: #64748b; margin-top: 2px; }

/* 阈值设置条 */
.threshold-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 16px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  margin-bottom: 16px;
}
.threshold-left { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.threshold-text { font-size: 14px; font-weight: 600; color: #92400e; }
.threshold-input { width: 110px; }
.threshold-hint { font-size: 12px; color: #b45309; }
.threshold-right { flex-shrink: 0; }

/* 工具栏 */
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-input-wrap { position: relative; flex: 1; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #94a3b8; }
.search-input { width: 100%; height: 36px; padding: 0 12px 0 36px; border: 1px solid #d9d9d9; border-radius: 6px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #409eff; }
.toolbar-cascader { width: 200px; }
.toolbar-status { width: 120px; }

/* Tab */
.tab-bar-wrap { margin-bottom: 16px; }
.tab-bar-inner {
  position: relative;
  display: flex;
  gap: 4px;
  padding: 5px;
  background: #f1f5f9;
  border-radius: var(--shop-radius-lg, 0.75rem);
  border: 1px solid var(--shop-border, #e8e8e8);
  width: fit-content;
}
.tab-slider {
  position: absolute;
  top: 5px;
  bottom: 5px;
  border-radius: var(--shop-radius-md, 0.5rem);
  background: var(--shop-primary, #3b82f6);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.35);
  transition: left 0.3s cubic-bezier(0.34, 1.56, 0.64, 1),
              width 0.3s cubic-bezier(0.34, 1.56, 0.64, 1),
              opacity 0.2s ease;
  z-index: 0;
  pointer-events: none;
}
.tab-item {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border-radius: var(--shop-radius-md, 0.5rem);
  cursor: pointer;
  color: var(--shop-muted-foreground, #64748b);
  font-size: 0.85rem;
  font-weight: 500;
  transition: color 0.2s ease;
  user-select: none;
  white-space: nowrap;
  background: transparent;
  border: none;
}
.tab-item:hover { color: var(--shop-primary, #3b82f6); }
.tab-item.active { color: #fff; font-weight: 600; }
.tab-count {
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: rgba(0, 0, 0, 0.25);
  color: #fff;
  font-size: 11px;
  line-height: 18px;
  text-align: center;
}

/* 库存列 */
.stock-cell { display: flex; align-items: center; gap: 8px; }
.stock-bar { flex: 1; max-width: 70px; }
.stock-num { font-weight: 600; min-width: 28px; }
.stock-empty { color: #ef4444; }
.stock-low { color: #f59e0b; }
.stock-ok { color: #10b981; }

/* 分页 */
.pagination-wrap { display: flex; justify-content: flex-end; padding: 16px 0; }

/* 图片 */
.product-img-cell {
  width: 60px; height: 60px; display: flex; align-items: center; justify-content: center;
  overflow: hidden; border-radius: 4px; background: #f5f5f5; position: relative;
}
.product-img { max-width: 100%; max-height: 100%; cursor: pointer; object-fit: cover; }
.img-placeholder-text { color: #ccc; font-size: 12px; }

/* 补货对话框 */
.restock-body { padding: 4px 8px; }
.restock-product {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8fafc;
  border-radius: 8px;
  margin-bottom: 16px;
}
.restock-img {
  width: 56px; height: 56px; border-radius: 6px; object-fit: cover;
  background: #fff; border: 1px solid #e8e8e8;
}
.restock-info { flex: 1; min-width: 0; }
.restock-name { font-size: 14px; font-weight: 600; color: #1e293b; margin-bottom: 4px; }
.restock-current { font-size: 13px; color: #64748b; }
.restock-form-label { font-size: 13px; color: #334155; margin-bottom: 8px; }
.restock-input { width: 100%; }
.restock-hint { font-size: 12px; color: #94a3b8; margin-top: 8px; }

/* SKU 明细补货 */
.restock-sku-wrap { display: flex; flex-direction: column; gap: 12px; }
.sku-list {
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 4px 8px;
  background: #fafbfc;
}
.sku-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 4px;
  border-bottom: 1px dashed #e8e8e8;
}
.sku-row:last-child { border-bottom: none; }
.sku-info { flex: 1; min-width: 0; }
.sku-spec { font-size: 13px; font-weight: 500; color: #1e293b; }
.sku-current { font-size: 12px; color: #94a3b8; margin-top: 2px; }
.sku-input { width: 140px; }
</style>

<!-- 图片预览样式（全局，因为 teleport 到 body） -->
<style>
.image-preview-overlay {
  position: fixed; inset: 0; z-index: 9999;
  background: rgba(0,0,0,0.7);
  display: flex; align-items: center; justify-content: center;
  cursor: zoom-out;
}
.image-preview-full {
  max-width: 80vw; max-height: 80vh;
  border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);
}
</style>
