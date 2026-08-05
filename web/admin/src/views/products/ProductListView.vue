<template>
  <!-- 商品管理列表页 -->
  <div class="product-list">
    <!-- Toolbar -->
    <div class="toolbar anim-enter" style="animation-delay: 0.02s">
      <el-button type="primary" @click="$router.push('/products/add')">
        <el-icon><Plus /></el-icon> 添加商品
      </el-button>
      <div class="search-input-wrap">
        <el-icon class="search-icon"><Search /></el-icon>
        <input v-model="filters.keyword" type="text" placeholder="搜索商品名称..."
          class="search-input" @input="onSearchInput" />
      </div>
      <el-cascader v-model="filters.category_id" :options="categoryTree"
        :props="{ value: 'id', label: 'name', children: 'children', expandTrigger: 'hover', emitPath: false }"
        placeholder="全部分类" clearable class="toolbar-cascader" @change="handleFilter" />
    </div>

    <!-- Bulk Action Bar -->
    <div class="bulk-bar-placeholder">
      <transition name="bulk-slide">
        <div v-show="selectedIds.length > 0" class="bulk-bar">
          <span>已选择 <strong class="bulk-count">{{ selectedIds.length }}</strong> 项</span>
          <div class="bulk-actions">
            <el-button size="small" @click="bulkAction('上架')">批量上架</el-button>
            <el-button size="small" @click="bulkAction('下架')">批量下架</el-button>
            <el-button size="small" type="danger" plain @click="bulkAction('删除')">批量删除</el-button>
          </div>
        </div>
      </transition>
    </div>

    <!-- Tabs -->
    <div class="tab-bar-wrap anim-enter" style="animation-delay: 0.06s">
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
        </div>
      </div>
    </div>

    <!-- Table -->
    <el-table class="prod-table anim-enter" :data="products" v-loading="loading" @selection-change="onSelectionChange" stripe style="width:100%; animation-delay: 0.1s">
      <template #empty>
        <el-empty :description="loading ? '加载中...' : '暂无商品'" />
      </template>
      <el-table-column type="selection" width="50" />
      <el-table-column label="图片" width="80">
        <template #default="{ row }">
          <div class="product-img-cell">
            <img :src="row.image" class="product-img" @click="previewImage(row.image)"
              @error="onImgError($event)" />
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="商品名称" min-width="200" show-overflow-tooltip />
      <el-table-column label="分类" width="120">
        <template #default="{ row }">
          {{ getCategoryName((row as any).categoryId || row.category_id) }}
        </template>
      </el-table-column>
      <el-table-column label="价格" width="120">
        <template #default="{ row }">¥{{ (row.price / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="库存" width="80">
        <template #default="{ row }">
          <span :class="{ 'stock-warning': row.stock === 0 }">{{ row.stock }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tooltip v-if="row.stock === 0" content="库存为0，建议先补货再上架" placement="top">
            <el-switch :model-value="row.status === '上架'" disabled />
          </el-tooltip>
          <el-switch v-else :model-value="row.status === '上架'"
            @change="handleToggleStatus(row.id, row.status)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="$router.push(`/products/edit/${row.id}`)">编辑</el-button>
          <el-popconfirm title="确认删除该商品？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize"
        :total="totalCount" :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadProducts" @size-change="loadProducts" />
    </div>
  </div>

  <!-- 图片预览 -->
  <teleport to="body">
    <div v-if="previewUrl" class="image-preview-overlay" @click.self="previewUrl = ''">
      <img :src="previewUrl" class="image-preview-full" @click="previewUrl = ''" />
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { Plus, Search, Grid, Upload, Download, WarningFilled } from '@element-plus/icons-vue'
import { listCategoryTree } from '../../api/category'
import { useProductList } from '../../composables/useProductList'

const {
  products, totalCount, currentPage, pageSize, loading,
  selectedIds, filters, loadProducts,
  bulkAction, handleDelete, handleToggleStatus, handleFilter,
} = useProductList()

const categoryTree = ref<any[]>([])
const previewUrl = ref('')

function previewImage(url: string) {
  previewUrl.value = url
}

const tabItems = [
  { key: '', label: '全部', icon: Grid },
  { key: '上架', label: '上架', icon: Upload },
  { key: '下架', label: '下架', icon: Download },
  { key: '缺货', label: '缺货', icon: WarningFilled },
]

const currentTab = ref('')

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

function switchTab(tab: string) {
  currentTab.value = tab
  if (tab === '缺货') {
    filters.value.status = ''
    filters.value.stock_eq = 0  // 后端按库存筛选
  } else {
    filters.value.status = tab
    filters.value.stock_eq = -1  // -1 表示不按库存筛选
  }
  updateSlider()
  handleFilter()
}

/** 搜索防抖 */
let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => handleFilter(), 300)
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

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map(r => r.id)
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
    const res = await listCategoryTree()
    categoryTree.value = [{ id: '0', name: '根目录', children: res.items }]
  } catch {
    categoryTree.value = []
  }
  loadProducts()
  nextTick(() => updateSlider())
})
</script>

<style scoped>
.product-list { padding: 24px; }
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-input-wrap { position: relative; flex: 1; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #94a3b8; }
.search-input { width: 100%; height: 36px; padding: 0 12px 0 36px; border: 1px solid #d9d9d9; border-radius: 6px; font-size: 14px; outline: none; }
.search-input:focus { border-color: #409eff; }
.toolbar-cascader { width: 200px; }
.bulk-bar-placeholder { min-height: 42px; margin-bottom: 8px; }
.bulk-bar { display: flex; align-items: center; gap: 16px; padding: 8px 16px; background: #e6f7ff; border: 1px solid #91d5ff; border-radius: 6px; font-size: 14px; }
.bulk-count { color: #1890ff; }
.bulk-actions { display: flex; gap: 8px; }
.tab-bar-wrap {
  margin-bottom: 16px;
}

.tab-bar-inner {
  position: relative;
  display: flex;
  gap: 4px;
  padding: 5px;
  background: #f1f5f9;
  border-radius: var(--shop-radius-lg, 0.75rem);
  border: 1px solid var(--shop-border, #e8e8e8);
}

/* 滑动指示器 */
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

.tab-item:hover {
  color: var(--shop-primary, #3b82f6);
}

.tab-item.active {
  color: #fff;
  font-weight: 600;
}
.pagination-wrap { display: flex; justify-content: flex-end; padding: 16px 0; }
.img-placeholder { width: 60px; height: 60px; display: flex; align-items: center; justify-content: center; background: #f5f5f5; color: #ccc; font-size: 12px; border-radius: 4px; }
.stock-warning { color: #f56c6c; font-weight: 600; }
.product-img-cell { width:60px; height:60px; display:flex; align-items:center; justify-content:center; overflow:hidden; border-radius:4px; background:#f5f5f5; position:relative; }
.product-img { max-width:100%; max-height:100%; cursor:pointer; object-fit:cover; }
.img-placeholder-text { color:#ccc; font-size:12px; }

/* ============================================
   动效
   ============================================ */

/* 页面区块入场动画 */
@keyframes prodFadeUp {
  from { opacity: 0; transform: translateY(14px); }
  to { opacity: 1; transform: translateY(0); }
}
.anim-enter {
  animation: prodFadeUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

/* 表格行 hover：渐变高亮 + 轻微右移 */
.prod-table :deep(.el-table__row) {
  transition: background-color 0.2s ease;
}
.prod-table :deep(.el-table__row:hover > td) {
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.05), rgba(59, 130, 246, 0.015) 70%, transparent) !important;
}

/* 商品图片 hover：放大 + 发光 */
.product-img-cell {
  transition: box-shadow 0.25s ease, border-radius 0.25s ease;
}
.product-img-cell:hover {
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.22);
}
.product-img-cell:hover .product-img {
  transform: scale(1.1);
}
.product-img {
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* 搜索框 focus 光晕 */
.search-input {
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.search-input:focus {
  border-color: #409eff;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.15);
}

/* Tab 项点击反馈 */
.tab-item:active {
  transform: scale(0.96);
}

/* 批量操作栏滑入滑出 */
.bulk-slide-enter-active,
.bulk-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.bulk-slide-enter-from,
.bulk-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 分页 hover */
.pagination-wrap :deep(.el-pager li),
.pagination-wrap :deep(.btn-prev),
.pagination-wrap :deep(.btn-next) {
  transition: all 0.2s ease;
}
.pagination-wrap :deep(.el-pager li:hover) {
  transform: translateY(-2px);
}
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
