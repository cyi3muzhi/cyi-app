<template>
  <!-- 商品管理页面 - 毛玻璃卡片风格、Tab切换、添加/编辑弹窗、规格构建器、SKU矩阵 -->
  <div class="product-list">
    <!-- ===== 工具栏 ===== -->
    <div class="toolbar">
      <el-button class="btn-gradient" @click="$router.push('/products/add')">
        <el-icon><Plus /></el-icon>
        添加商品
      </el-button>
      <div class="search-input-wrap">
        <el-icon class="search-icon"><Search /></el-icon>
        <input
          v-model="searchKeyword"
          type="text"
          placeholder="搜索商品名称或品牌..."
          class="search-input"
          @input="handleFilter"
        />
      </div>
      <el-cascader
        v-model="categoryFilter"
        :options="categoryTree"
        :props="{ value: 'name', label: 'name', children: 'children', expandTrigger: 'hover', emitPath: false }"
        placeholder="全部分类"
        clearable
        class="toolbar-cascader"
        @change="handleFilter"
      />
    </div>

    <!-- ===== 批量操作栏（固定占位，防止抖动） ===== -->
    <div class="bulk-bar-placeholder">
      <div v-show="selectedIds.length > 0" class="bulk-bar">
        <span>已选择 <strong class="bulk-count">{{ selectedIds.length }}</strong> 项</span>
        <div class="bulk-actions">
          <el-button size="small" class="bulk-btn" @click="bulkAction('上架')">批量上架</el-button>
          <el-button size="small" class="bulk-btn" @click="bulkAction('下架')">批量下架</el-button>
          <el-button size="small" class="bulk-btn bulk-btn--danger" @click="bulkAction('删除')">批量删除</el-button>
        </div>
      </div>
    </div>

    <!-- ===== Tab 标签切换 ===== -->
    <div class="tab-bar">
      <div
        v-for="tab in tabItems"
        :key="tab.key"
        :class="['tab-item', { active: currentTab === tab.key }]"
        @click="switchTab(tab.key)"
      >
        {{ tab.label }}
        <span class="tab-count">{{ tab.count }}</span>
      </div>
    </div>

    <!-- ===== 商品表格 ===== -->
    <div class="table-wrap">
      <table class="admin-table">
        <thead>
          <tr>
            <th style="width:40px;"><input type="checkbox" class="admin-checkbox" :checked="allChecked" @change="toggleAllChecks" /></th>
            <th style="width:64px;">图片</th>
            <th style="width:220px;">商品 (SPU)</th>
            <th style="width:100px;">分类</th>
            <th style="width:100px;">规格</th>
            <th style="width:130px;">价格区间</th>
            <th style="width:80px;">总库存</th>
            <th style="width:70px;">销量</th>
            <th style="width:100px;">状态</th>
            <th style="width:200px;">操作</th>
          </tr>
        </thead>
        <tbody id="productTableBody">
          <tr v-for="row in pageData" :key="row.id">
            <td><input type="checkbox" class="admin-checkbox" :checked="selectedIds.includes(row.id)" @change="toggleRowCheck(row.id)" /></td>
            <td><el-image class="product-thumb" :src="row.image || ''" :alt="row.name" :preview-src-list="row.image ? [row.image] : []" fit="cover" @error="onImgError" /></td>
            <td>
              <div class="product-name-cell">
                <div class="name-text">{{ row.name }}</div>
                <div class="desc-text">{{ row.description }}</div>
              </div>
            </td>
            <td><span class="cell-muted">{{ row.category }}</span></td>
            <td><span class="spec-badge" @click="openSkuPopup(row)">{{ row.skus.length }} 个SKU</span></td>
            <td><span class="price-range">{{ getPriceRange(row.skus) }}</span></td>
            <td>{{ getTotalStock(row.skus) }}</td>
            <td>{{ getTotalSales(row.skus) }}</td>
            <td><span :class="['status-badge', statusClass(row.status)]">{{ row.status }}</span></td>
            <td>
              <div class="action-links">
                <button class="action-link action-link--edit" @click="openEditModal(row)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  <span>编辑</span>
                </button>
                <span class="action-divider" />
                <button class="action-link action-link--toggle" @click="handleToggleStatus(row)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                  <span>{{ row.status === '上架中' ? '下架' : '上架' }}</span>
                </button>
                <span class="action-divider" />
                <button class="action-link action-link--delete" @click="handleDelete(row)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  <span>删除</span>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="pageData.length === 0">
            <td colspan="10" class="empty-cell">暂无商品数据</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- ===== 分页 ===== -->
    <div class="pagination-wrap">
      <div class="pagination">
        <button
          class="page-btn"
          :class="{ 'page-btn--disabled': currentPage <= 1 }"
          :disabled="currentPage <= 1"
          @click="goPage(currentPage - 1)"
        >上一页</button>
        <template v-for="p in paginationPages" :key="p">
          <span v-if="p === '...'" class="page-btn page-btn--ellipsis">...</span>
          <button
            v-else
            class="page-btn"
            :class="{ 'page-btn--active': p === currentPage }"
            @click="goPage(Number(p))"
          >{{ p }}</button>
        </template>
        <button
          class="page-btn"
          :class="{ 'page-btn--disabled': currentPage >= totalPages }"
          :disabled="currentPage >= totalPages"
          @click="goPage(currentPage + 1)"
        >下一页</button>
      </div>
      <div class="pagination-info">共 {{ filteredProducts.length }} 个商品，第 {{ currentPage }}/{{ totalPages }} 页</div>
    </div>

    <!-- ====== 添加/编辑商品弹窗 ====== -->
    <div v-if="showProductModal" class="modal-overlay" @click.self="closeProductModal">
      <div class="modal modal-wide">
        <div class="modal-header">
          <h2>{{ isEditMode ? '编辑商品' : '添加商品' }}</h2>
          <button class="modal-close" @click="closeProductModal">&times;</button>
        </div>
        <div class="modal-body">
          <!-- 商品名称 + 品牌 -->
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">商品名称 <span class="required">*</span></label>
              <input v-model="form.name" type="text" class="form-input" placeholder="请输入商品名称" />
            </div>
            <div class="form-group">
              <label class="form-label">品牌 <span class="required">*</span></label>
              <input v-model="form.brand" type="text" class="form-input" placeholder="请输入品牌名称" />
            </div>
          </div>
          <!-- 分类（级联选择器） -->
          <div class="form-row">
            <div class="form-group" style="flex:2;">
              <label class="form-label">分类 <span class="required">*</span></label>
              <el-cascader
                v-model="form.category"
                :options="categoryTree"
                :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true }"
                placeholder="请选择分类"
                class="custom-cascader"
                clearable
                filterable
              />
            </div>
            <div class="form-group" style="flex:1;">
              <label class="form-label">状态</label>
              <select v-model="form.status" class="form-select">
                <option value="上架中">上架中</option>
                <option value="已下架">已下架</option>
              </select>
            </div>
          </div>
          <!-- 商品描述 -->
          <div class="form-group">
            <label class="form-label">商品描述</label>
            <textarea v-model="form.description" class="form-textarea" rows="2" placeholder="简短描述商品特点"></textarea>
          </div>
          <!-- 图片上传 -->
          <div class="form-group">
            <label class="form-label">商品图片</label>
            <div
              :class="['upload-area', { 'has-image': form.image }]"
              @click="triggerImageUpload"
            >
              <template v-if="!form.image">
                <div class="upload-placeholder">
                  <div class="upload-icon">+</div>
                  <div class="upload-text">点击上传图片</div>
                  <div class="upload-hint">支持 JPG / PNG</div>
                </div>
              </template>
              <div v-else class="upload-preview">
                <img :src="form.image" alt="preview" />
              </div>
            </div>
            <input ref="imgInputRef" type="file" accept="image/*" style="display:none" @change="handleImageUpload" />
          </div>

          <!-- ===== 规格构建器 ===== -->
          <div class="spec-section-title">商品规格 (SKU 管理)</div>
          <div v-for="(spec, idx) in form.specs" :key="idx" class="spec-row">
            <div class="spec-name-wrap">
              <input
                v-model="spec.name"
                type="text"
                class="form-input"
                placeholder="规格名 (如颜色)"
                @input="regenerateSkuMatrix"
              />
            </div>
            <div class="spec-values-wrap">
              <div class="tag-input-wrap" @click.self="focusTagInput($event, idx)">
                <span v-for="(v, vi) in spec.values" :key="vi" class="tag-badge">
                  {{ v }}
                  <span class="tag-remove" @click="removeSpecValue(idx, vi)">&times;</span>
                </span>
                <input
                  :ref="el => { if (el) tagInputRefs[idx] = el as HTMLInputElement }"
                  type="text"
                  placeholder="输入值后回车"
                  @keydown.enter.prevent="addSpecValue(idx, $event)"
                />
              </div>
            </div>
            <button class="spec-remove-btn" @click="removeSpecGroup(idx)" title="删除规格组">&times;</button>
          </div>
          <button class="add-spec-btn" @click="addSpecGroup">+ 添加规格组</button>

          <!-- ===== SKU 矩阵 ===== -->
          <div v-if="validSpecGroups.length > 0 && form.skus.length > 0" class="sku-matrix-wrap">
            <div class="sku-matrix-title">SKU 列表</div>
            <div class="sku-matrix-scroll">
              <table class="sku-matrix">
                <thead>
                  <tr>
                    <th v-for="s in validSpecGroups" :key="s.name">{{ s.name }}</th>
                    <th>价格 (元)</th>
                    <th>库存</th>
                    <th style="width:32px;"></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="sku in form.skus" :key="sku.id">
                    <td v-for="s in validSpecGroups" :key="s.name">
                      <div class="spec-combo">
                        <span class="combo-tag">{{ sku.combo[s.name] || '' }}</span>
                      </div>
                    </td>
                    <td><input v-model.number="sku.price" type="number" class="sku-price-input" step="0.01" min="0" /></td>
                    <td><input v-model.number="sku.stock" type="number" class="sku-stock-input" min="0" step="1" /></td>
                    <td><button class="sku-del-btn" @click="deleteSkuFromMatrix(sku.id)" title="删除此SKU">&times;</button></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div v-else-if="form.specs.length > 0" class="sku-empty-hint">
            请先添加规格组并填写规格值，系统将自动生成 SKU 矩阵
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="closeProductModal">取消</button>
          <button class="btn-save" @click="saveProduct">保存商品</button>
        </div>
      </div>
    </div>

    <!-- ====== SKU 详情弹窗 ====== -->
    <div v-if="showSkuPopup" class="modal-overlay" @click.self="closeSkuPopup">
      <div class="sku-popup">
        <div class="sku-popup-header">
          <h2>{{ skuPopupProduct?.name }} - SKU 详情</h2>
          <button class="modal-close" @click="closeSkuPopup">&times;</button>
        </div>
        <div class="sku-popup-body">
          <div class="popup-spu-desc" v-if="skuPopupProduct">
            <span><span class="label">品牌：</span>{{ skuPopupProduct.brand }}</span>
            <span><span class="label">分类：</span>{{ skuPopupProduct.category }}</span>
            <span><span class="label">状态：</span>{{ skuPopupProduct.status }}</span>
          </div>
          <table class="sku-detail-table">
            <thead>
              <tr>
                <th>规格组合</th>
                <th>价格 (元)</th>
                <th>库存</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sku in (skuPopupProduct?.skus || [])" :key="sku.id">
                <td>
                  <div class="spec-combo">
                    <span class="combo-tag" v-for="(val, key) in sku.combo" :key="key">{{ val }}</span>
                  </div>
                </td>
                <td class="sku-price">&yen;{{ sku.price.toFixed(2) }}</td>
                <td>{{ sku.stock }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ====== 确认删除弹窗 ====== -->
    <div v-if="showConfirmModal" class="modal-overlay" @click.self="closeConfirmModal">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h2>确认操作</h2>
          <button class="modal-close" @click="closeConfirmModal">&times;</button>
        </div>
        <div class="modal-body">
          <p class="confirm-msg">确定要删除该商品吗？删除后不可恢复。</p>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="closeConfirmModal">取消</button>
          <button class="btn-save btn-danger" @click="confirmDeleteProduct">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 商品管理列表页面
 * 功能：搜索筛选、Tab切换、批量操作、添加/编辑弹窗（含规格构建器+SKU矩阵）、SKU详情弹窗、删除确认
 * 样式：毛玻璃效果、渐变按钮、圆角卡片
 */
import { ref, computed, reactive, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { listCategoryTree } from '../../api/category'
import type { CategoryItem } from '../../api/category'
import { uploadToOSS } from '../../utils/upload'
import {
  listProducts as apiListProducts,
  createProduct as apiCreateProduct,
  updateProduct as apiUpdateProduct,
  deleteProduct as apiDeleteProduct,
  batchProductOperation as apiBatchProductOperation
} from '../../api/product'

const router = useRouter()

// ======================== 类型定义 ========================

/** SKU 接口 */
interface SkuItem {
  id: string
  combo: Record<string, string>
  price: number
  stock: number
  sales?: number
}

/** 规格组接口 */
interface SpecGroup {
  name: string
  values: string[]
}

/** 商品接口（本地富模型，含规格和 SKU） */
interface ProductItem {
  id: number
  name: string
  brand: string
  category: string
  status: string
  description: string
  image: string
  specs: SpecGroup[]
  skus: SkuItem[]
}

/** 加载状态 */
const loading = ref(false)

// ======================== 分类数据（从 API 加载） ========================

/** 分类树（从后端获取） */
const categoryTree = ref<CategoryItem[]>([])

/** 加载分类树 */
async function loadCategoryTree() {
  try {
    const res = await listCategoryTree()
    categoryTree.value = res.items || []
  } catch {
    categoryTree.value = []
  }
}

/** 所有分类名称扁平列表（用于筛选下拉） */
const allCategoryNames = computed(() => {
  const names: string[] = []
  function walk(nodes: CategoryItem[]) {
    for (const n of nodes) {
      names.push(n.name)
      if (n.children) walk(n.children)
    }
  }
  walk(categoryTree.value)
  return names
})

// ======================== 响应式状态 ========================

/** 生成唯一 ID（用于本地 SKU） */
function uid(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
}

/** 商品列表数据（从 API 加载） */
const products = ref<ProductItem[]>([])

/** 总记录数（服务端返回） */
const totalCount = ref(0)

/** 搜索关键词 */
const searchKeyword = ref('')

/** 分类筛选 */
const categoryFilter = ref('')

/** 当前Tab */
const currentTab = ref('全部')

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(6)

/** 选中ID列表 */
const selectedIds = ref<number[]>([])

/** 全选状态 */
const allChecked = computed(() => pageData.value.length > 0 && pageData.value.every(p => selectedIds.value.includes(p.id)))

/** 弹窗显示状态 */
const showProductModal = ref(false)
const showSkuPopup = ref(false)
const showConfirmModal = ref(false)

/** SKU详情弹窗中的商品 */
const skuPopupProduct = ref<ProductItem | null>(null)

/** 是否为编辑模式 */
const isEditMode = ref(false)

/** 待删除的商品ID */
const deleteTargetId = ref<number | null>(null)

/** 图片上传input ref */
const imgInputRef = ref<HTMLInputElement | null>(null)

/** 规格输入框ref集合 */
const tagInputRefs = ref<Record<number, HTMLInputElement>>({})

// ======================== 表单状态 ========================

/** 弹窗表单数据 */
const form = reactive({
  id: 0,
  name: '',
  brand: '',
  category: [] as number[],
  status: '上架中',
  description: '',
  image: '',
  _imageObjectKey: '',
  specs: [] as SpecGroup[],
  skus: [] as SkuItem[]
})

// ======================== 计算属性 ========================

/** Tab项列表（含计数） */
const tabItems = computed(() => {
  const all = products.value.length
  const active = countByStatus('上架中')
  const inactive = countByStatus('已下架')
  const outOfStock = products.value.filter(p => getTotalStock(p.skus) === 0).length
  return [
    { key: '全部', label: '全部', count: all },
    { key: '上架中', label: '上架中', count: active },
    { key: '已下架', label: '已下架', count: inactive },
    { key: '缺货', label: '缺货', count: outOfStock }
  ]
})

/** 过滤后的商品列表 */
const filteredProducts = computed(() => {
  let list = [...products.value]
  const kw = searchKeyword.value.toLowerCase().trim()
  if (kw) {
    list = list.filter(p =>
      p.name.toLowerCase().includes(kw) || p.brand.toLowerCase().includes(kw)
    )
  }
  if (categoryFilter.value) {
    list = list.filter(p => p.category === categoryFilter.value)
  }
  if (currentTab.value === '上架中') {
    list = list.filter(p => p.status === '上架中')
  } else if (currentTab.value === '已下架') {
    list = list.filter(p => p.status === '已下架')
  } else if (currentTab.value === '缺货') {
    list = list.filter(p => getTotalStock(p.skus) === 0)
  }
  return list
})

/** 当前页数据（服务端已分页，无需前端 slice） */
const pageData = computed(() => {
  return products.value
})

/** 总页数（基于服务端返回的 totalCount） */
const totalPages = computed(() => {
  return Math.ceil(totalCount.value / pageSize.value) || 1
})

/** 分页按钮列表 */
const paginationPages = computed(() => {
  const total = totalPages.value
  const cur = currentPage.value
  const pages: (number | string)[] = []
  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= cur - 1 && i <= cur + 1)) {
      pages.push(i)
    } else if (pages[pages.length - 1] !== '...') {
      pages.push('...')
    }
  }
  return pages
})

/** 有效的规格组（有名称且有值） */
const validSpecGroups = computed(() => {
  return form.specs.filter(s => s.name && s.values.length > 0)
})

// ======================== 工具函数 ========================

/**
 * 计算总库存
 * @param skus SKU列表
 */
function getTotalStock(skus: SkuItem[]): number {
  return skus.reduce((sum, s) => sum + s.stock, 0)
}

/**
 * 计算总销量
 * @param skus SKU列表
 */
function getTotalSales(skus: SkuItem[]): number {
  return skus.reduce((sum, s) => sum + (s.sales || 0), 0)
}

/**
 * 获取价格区间文本
 * @param skus SKU列表
 */
function getPriceRange(skus: SkuItem[]): string {
  const prices = skus.map(s => s.price)
  if (prices.length === 0) return '--'
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  if (min === max) return '\u00a5' + min.toFixed(2)
  return '\u00a5' + min.toFixed(2) + ' ~ \u00a5' + max.toFixed(2)
}

/**
 * 获取状态CSS类名
 * @param status 商品状态
 */
function statusClass(status: string): string {
  if (status === '上架中') return 'status-badge--active'
  if (status === '已下架') return 'status-badge--inactive'
  return 'status-badge--outofstock'
}

/**
 * 按状态统计商品数量
 * @param status 状态
 */
function countByStatus(status: string): number {
  return products.value.filter(p => p.status === status).length
}

/**
 * 图片加载失败时显示占位
 */
function onImgError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

// ======================== 搜索/筛选 ========================

/** 搜索/筛选处理 */
function handleFilter() {
  currentPage.value = 1
}

// ======================== Tab切换 ========================

/**
 * 切换Tab
 * @param tab Tab标识
 */
function switchTab(tab: string) {
  currentTab.value = tab
  currentPage.value = 1
  selectedIds.value = []
}

// ======================== 选择/批量操作 ========================

/** 切换行选中 */
function toggleRowCheck(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

/** 全选/取消全选 */
function toggleAllChecks() {
  if (allChecked.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = pageData.value.map(p => p.id)
  }
}

/**
 * 批量操作
 * @param action 操作类型："上架" | "下架" | "删除"
 */
async function bulkAction(action: string) {
  const count = selectedIds.value.length
  if (count === 0) return
  const actionText = action === '上架' || action === '下架' ? action : '删除'
  if (action === '删除') {
    if (!confirm('确定要删除选中的 ' + count + ' 个商品吗？')) return
  }
  try {
    await apiBatchProductOperation(action, [...selectedIds.value])
    ElMessage.success(`批量${actionText}成功`)
    selectedIds.value = []
    await loadProducts()
  } catch {
    // 错误已在拦截器中处理
  }
}

// ======================== 上下架切换 ========================

/**
 * 切换上下架状态
 * @param row 商品
 */
async function handleToggleStatus(row: ProductItem) {
  const newStatus = row.status === '上架中' ? '已下架' : '上架中'
  try {
    await apiUpdateProduct(row.id, { status: newStatus })
    ElMessage.success(`商品已${newStatus}`)
    await loadProducts()
  } catch {
    // 错误已在拦截器中处理
  }
}

// ======================== 删除 ========================

/**
 * 打开删除确认弹窗
 * @param row 商品
 */
function handleDelete(row: ProductItem) {
  deleteTargetId.value = row.id
  showConfirmModal.value = true
}

/** 关闭确认弹窗 */
function closeConfirmModal() {
  showConfirmModal.value = false
  deleteTargetId.value = null
}

/** 确认删除 */
async function confirmDeleteProduct() {
  if (deleteTargetId.value) {
    try {
      await apiDeleteProduct(deleteTargetId.value)
      ElMessage.success('商品已删除')
      await loadProducts()
    } catch {
      // 错误已在拦截器中处理
    }
  }
  closeConfirmModal()
}

// ======================== 添加/编辑弹窗 ========================

/** 重置表单 */
function resetForm() {
  form.id = 0
  form.name = ''
  form.brand = ''
  form.category = []
  form.status = '上架中'
  form.description = ''
  form.image = ''
  form._imageObjectKey = ''
  form.specs = []
  form.skus = []
}

/** 打开添加弹窗 */
function openAddModal() {
  resetForm()
  isEditMode.value = false
  showProductModal.value = true
}

/**
 * 打开编辑弹窗
 * @param row 商品
 */
function openEditModal(row: ProductItem) {
  resetForm()
  isEditMode.value = true
  form.id = row.id
  form.name = row.name
  form.brand = row.brand
  form.status = row.status === '缺货' ? '上架中' : row.status
  form.description = row.description || ''
  form.image = row.image || ''
  form.specs = JSON.parse(JSON.stringify(row.specs))
  form.skus = JSON.parse(JSON.stringify(row.skus))
  setFormCategory(row.category)
  showProductModal.value = true
}

/** 关闭弹窗 */
function closeProductModal() {
  showProductModal.value = false
}

/** 触发图片上传 */
function triggerImageUpload() {
  imgInputRef.value?.click()
}

/**
 * 图片上传处理（压缩 → 直传 OSS）
 * @param e 事件
 */
async function handleImageUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const result = await uploadToOSS(file, 'product')
    form.image = result.publicUrl
    form._imageObjectKey = result.objectKey
  } catch (err) {
    console.error('上传失败:', err)
  }
}

// ======================== 级联分类回填 ========================

/** 根据分类名称查找级联路径（用于编辑时回填） */
function findCategoryPath(name: string): number[] {
  function walk(arr: CategoryItem[], path: number[]): number[] | null {
    for (const item of arr) {
      if (item.name === name) return [...path, item.id]
      if (item.children) {
        const found = walk(item.children, [...path, item.id])
        if (found) return found
      }
    }
    return null
  }
  return walk(categoryTree.value, []) || []
}

/** 在分类树中递归查找指定 ID 的节点名称 */
function findCategoryNameById(id: number): string {
  function walk(arr: CategoryItem[]): string | null {
    for (const item of arr) {
      if (item.id === id) return item.name
      if (item.children) {
        const found = walk(item.children)
        if (found) return found
      }
    }
    return null
  }
  return walk(categoryTree.value) || ''
}

/**
 * 根据分类名称设置级联选择器
 * @param catName 分类名称
 */
function setFormCategory(catName: string) {
  form.category = findCategoryPath(catName)
}

// ======================== 规格构建器 ========================

/** 添加规格组 */
function addSpecGroup() {
  form.specs.push({ name: '', values: [] })
}

/**
 * 删除规格组
 * @param idx 索引
 */
function removeSpecGroup(idx: number) {
  form.specs.splice(idx, 1)
  regenerateSkuMatrix()
}

/**
 * 添加规格值（回车触发）
 * @param specIdx 规格组索引
 * @param e 键盘事件
 */
function addSpecValue(specIdx: number, e: KeyboardEvent) {
  const input = e.target as HTMLInputElement
  const val = input.value.trim()
  if (!val) return
  form.specs[specIdx].values.push(val)
  input.value = ''
  regenerateSkuMatrix()
}

/**
 * 删除规格值
 * @param specIdx 规格组索引
 * @param valIdx 值索引
 */
function removeSpecValue(specIdx: number, valIdx: number) {
  form.specs[specIdx].values.splice(valIdx, 1)
  if (form.specs[specIdx].values.length === 0) {
    form.specs.splice(specIdx, 1)
  }
  regenerateSkuMatrix()
}

/**
 * 聚焦到规格输入框
 * @param e 点击事件
 * @param idx 索引
 */
function focusTagInput(e: MouseEvent, idx: number) {
  const input = tagInputRefs.value[idx]
  if (input) input.focus()
}

/** 重新生成SKU矩阵（笛卡尔积） */
function regenerateSkuMatrix() {
  const validSpecs = form.specs.filter(s => s.name && s.values.length > 0)
  if (validSpecs.length === 0) {
    form.skus = []
    return
  }
  const combos = cartesian(validSpecs.map(s => s.values))
  const newSkus: SkuItem[] = []
  const oldSkus = [...form.skus]
  combos.forEach(combo => {
    const comboMap: Record<string, string> = {}
    const keyParts: string[] = []
    validSpecs.forEach((s, i) => {
      comboMap[s.name] = combo[i]
      keyParts.push(s.name + ':' + combo[i])
    })
    const key = keyParts.join('|')
    const existing = oldSkus.find(sku => {
      const oldKey = Object.keys(sku.combo).map(k => k + ':' + sku.combo[k]).join('|')
      return oldKey === key
    })
    if (existing) {
      newSkus.push(JSON.parse(JSON.stringify(existing)))
    } else {
      newSkus.push({ id: uid(), combo: comboMap, price: 0, stock: 0 })
    }
  })
  form.skus = newSkus
}

/**
 * 从SKU矩阵中删除SKU
 * @param skuId SKU ID
 */
function deleteSkuFromMatrix(skuId: string) {
  form.skus = form.skus.filter(s => s.id !== skuId)
}

/**
 * 笛卡尔积
 * @param arrays 二维数组
 */
function cartesian(arrays: string[][]): string[][] {
  if (arrays.length === 0) return [[]]
  let result: string[][] = [[]]
  for (const sub of arrays) {
    const next: string[][] = []
    for (const r of result) {
      for (const s of sub) {
        next.push([...r, s])
      }
    }
    result = next
  }
  return result
}

// ======================== 保存商品 ========================

/** 保存商品 */
async function saveProduct() {
  if (!form.name.trim() || !form.brand.trim()) {
    ElMessage.warning('请填写商品名称和品牌')
    return
  }
  if (!form.category) {
    ElMessage.warning('请选择商品分类')
    return
  }
  const specs = form.specs.filter(s => s.name && s.values.length > 0)
  const skus = form.skus
  if (specs.length > 0 && skus.length === 0) {
    ElMessage.warning('规格已设置，但未生成SKU，请检查规格配置')
    return
  }

  const categoryName = form.category.length > 0
    ? findCategoryNameById(form.category[form.category.length - 1])
    : ''

  try {
    // specs：直接映射为 API 格式
    const apiSpecs = specs.map(s => ({ name: s.name, values: s.values }))
    // skus：价格从元转为分
    const apiSkus = skus.map(s => ({
      sku_code: s.id,
      price: Math.round(s.price * 100),
      stock: s.stock,
      sales: s.sales || 0,
      spec: s.combo
    }))

    if (isEditMode.value) {
      // 编辑模式：调用更新 API
      await apiUpdateProduct(form.id, {
        name: form.name.trim(),
        category: categoryName,
        price: 0,
        stock: 0,
        description: form.description.trim(),
        image: form._imageObjectKey || form.image || undefined,
        status: form.status,
        specs: apiSpecs,
        skus: apiSkus
      })
      ElMessage.success('商品已更新')
    } else {
      // 添加模式：调用创建 API
      await apiCreateProduct({
        name: form.name.trim(),
        category: categoryName,
        price: 0,
        stock: 0,
        description: form.description.trim(),
        image: form._imageObjectKey || form.image || undefined,
        specs: apiSpecs,
        skus: apiSkus
      })
      ElMessage.success('商品已添加')
    }
    closeProductModal()
    await loadProducts()
  } catch {
    // 错误已在拦截器中处理
  }
}

// ======================== SKU详情弹窗 ========================

/**
 * 打开SKU详情弹窗
 * @param row 商品
 */
function openSkuPopup(row: ProductItem) {
  skuPopupProduct.value = row
  showSkuPopup.value = true
}

/** 关闭SKU详情弹窗 */
function closeSkuPopup() {
  showSkuPopup.value = false
  skuPopupProduct.value = null
}

// ======================== 分页 ========================

/**
 * 跳转到指定页
 * @param page 页码
 */
function goPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadProducts()
}

// ======================== 从 API 加载数据 ========================

/**
 * 从 API 加载商品列表
 * 将 API 返回的数据映射为本地富模型，specs/skus 价格从分转为元
 */
async function loadProducts() {
  loading.value = true
  try {
    const res = await apiListProducts({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value || undefined,
      category: categoryFilter.value || undefined
    })
    totalCount.value = res.total
    products.value = (res.items || []).map(item => ({
      id: item.id,
      name: item.name,
      brand: '',
      category: item.category,
      status: item.status,
      description: '',
      image: item.image,
      specs: (item.specs || []).map(s => ({ name: s.name, values: s.values })),
      skus: (item.skus || []).map(s => ({
        id: String(s.id),
        combo: s.spec || {},
        price: s.price / 100, // 分转元
        stock: s.stock,
        sales: s.sales || 0
      }))
    }))
  } catch {
    products.value = []
  } finally {
    loading.value = false
  }
}

// ======================== 初始化 ========================

onMounted(() => {
  loadCategoryTree()
  loadProducts()
})
</script>

<style scoped>
/* ============================================
   商品管理页面样式
   毛玻璃卡片、渐变按钮、圆角设计
   ============================================ */

.product-list {
  width: 100%;
}

/* ---- 工具栏（毛玻璃） ---- */
.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.7);
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.04);
}

.btn-gradient {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #4f46e5) !important;
  color: #fff !important;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  box-shadow: 0 1px 2px rgba(99,102,241,0.15);
}

.btn-gradient:hover {
  box-shadow: 0 4px 12px rgba(99,102,241,0.3);
  transform: translateY(-1px);
}

.btn-gradient:active {
  transform: scale(0.97);
}

.search-input-wrap {
  position: relative;
  flex: 1;
  min-width: 180px;
  max-width: 280px;
}

.search-input-wrap .search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

.search-input-wrap .search-input {
  width: 100%;
  height: 38px;
  padding: 0 12px 0 36px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  font-family: var(--shop-font-sans);
  background: #f8fafc;
  color: #0f172a;
  outline: none;
  transition: all 0.2s ease;
}

.search-input-wrap .search-input:focus {
  border-color: #6366f1;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.toolbar-select {
  width: 160px;
}

.toolbar-select :deep(.el-select__wrapper) {
  height: 38px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  box-shadow: none;
}

.toolbar-select :deep(.el-select__wrapper:hover),
.toolbar-select :deep(.el-select__wrapper.is-focused) {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.toolbar-cascader {
  width: 160px;
}

.toolbar-cascader :deep(.el-cascader__wrapper) {
  height: 38px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  box-shadow: none;
}

.toolbar-cascader :deep(.el-cascader__wrapper:hover),
.toolbar-cascader :deep(.el-cascader.is-focus .el-cascader__wrapper) {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

/* ---- 批量操作栏占位（固定高度，选不选中都不抖动） ---- */
.bulk-bar-placeholder {
  min-height: 48px;
  margin-bottom: 12px;
}

/* ---- 批量操作栏 ---- */
.bulk-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  background: rgba(99,102,241,0.06);
  border: 1px solid rgba(99,102,241,0.2);
  border-radius: 12px;
  font-size: 0.8125rem;
  color: #0f172a;
}

.bulk-bar .bulk-count {
  font-weight: 600;
  color: #6366f1;
}

.bulk-bar .bulk-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.bulk-btn {
  padding: 5px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  font-family: var(--shop-font-sans);
  cursor: pointer;
  transition: all 0.15s ease;
  background: #fff;
  color: #475569;
}

.bulk-btn:hover {
  border-color: #6366f1;
  color: #6366f1;
}

.bulk-btn--danger {
  color: #ef4444;
  border-color: #ef4444;
}

.bulk-btn--danger:hover {
  background: #ef4444;
  color: #fff;
}

/* ---- Tab 标签切换 ---- */
.tab-bar {
  display: flex;
  gap: 0;
  margin-bottom: 16px;
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.7);
  border-radius: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.03);
  padding: 4px;
  overflow: hidden;
}

.tab-item {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  border-radius: 10px;
  transition: all 0.2s ease;
  user-select: none;
  white-space: nowrap;
}

.tab-item:hover {
  color: #0f172a;
  background: #f8fafc;
}

.tab-item.active {
  color: #fff;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  box-shadow: 0 2px 8px rgba(99,102,241,0.2);
}

.tab-item .tab-count {
  font-size: 0.6875rem;
  font-weight: 600;
  background: rgba(0,0,0,0.06);
  padding: 0.0625rem 0.4375rem;
  border-radius: 999px;
  min-width: 1.25rem;
  text-align: center;
}

.tab-item.active .tab-count {
  background: rgba(255,255,255,0.2);
  color: #fff;
}

/* ---- 表格容器（毛玻璃） ---- */
.table-wrap {
  overflow-x: auto;
  background: rgba(255,255,255,0.88);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.7);
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.04);
}

/* ---- 自定义商品表格 ---- */
.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  table-layout: fixed;
}

.admin-table thead { background: #f8fafc; }

.admin-table thead th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #64748b;
  font-size: 0.6875rem;
  letter-spacing: 0.04em;
  border-bottom: 1px solid #f1f5f9;
  white-space: nowrap;
}

.admin-table tbody tr { transition: background 0.12s ease; }

.admin-table tbody tr:hover { background: rgba(99,102,241,0.03); }

.admin-table tbody td {
  padding: 14px 16px;
  vertical-align: middle;
  border-bottom: 1px solid #f1f5f9;
  color: #0f172a;
}

.admin-table tbody tr:last-child td { border-bottom: none; }

.admin-table .empty-cell {
  text-align: center;
  padding: 40px 16px;
  color: #94a3b8;
}

/* ---- 自定义复选框 ---- */
.admin-checkbox {
  appearance: none;
  width: 16px;
  height: 16px;
  border: 2px solid #cbd5e1;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
  position: relative;
  vertical-align: middle;
}

.admin-checkbox:checked {
  background: #6366f1;
  border-color: #6366f1;
}

.admin-checkbox:checked::after {
  content: '';
  position: absolute;
  left: 3px;
  top: 1px;
  width: 6px;
  height: 9px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.admin-checkbox:focus-visible {
  outline: 2px solid rgba(99,102,241,0.3);
  outline-offset: 2px;
}

/* ---- 表格内元素 ---- */
.product-thumb {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  object-fit: cover;
  background: #f1f5f9;
  display: block;
}

.product-name-cell .name-text {
  font-weight: 600;
  color: #0f172a;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.3;
}

.product-name-cell .desc-text {
  font-size: 0.6875rem;
  color: #94a3b8;
  margin-top: 2px;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cell-muted {
  color: #64748b;
}

/* ---- 状态徽标 ---- */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-badge--active {
  background: #f0fdf4;
  color: #16a34a;
}

.status-badge--active::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #22c55e;
}

.status-badge--inactive {
  background: #f8fafc;
  color: #94a3b8;
}

.status-badge--inactive::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #cbd5e1;
}

.status-badge--outofstock {
  background: #fef2f2;
  color: #dc2626;
}

.status-badge--outofstock::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ef4444;
}

/* ---- 规格徽标 ---- */
.spec-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  background: #eef2ff;
  color: #6366f1;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid rgba(99,102,241,0.18);
}

.spec-badge:hover {
  background: #e0e7ff;
  border-color: rgba(99,102,241,0.35);
  box-shadow: 0 2px 8px rgba(99,102,241,0.12);
}

/* ---- 价格区间 ---- */
.price-range {
  font-weight: 600;
  color: #0f172a;
  white-space: nowrap;
}

/* ---- 操作链接 ---- */
.action-links {
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-link {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  color: #64748b;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.15s ease;
  border: none;
  background: transparent;
  font-family: var(--shop-font-sans);
}

.action-link:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.action-link--edit:hover {
  color: #6366f1;
  background: rgba(99,102,241,0.08);
}

.action-link--toggle:hover {
  color: #f59e0b;
  background: #fffbeb;
}

.action-link--delete:hover {
  color: #ef4444;
  background: #fef2f2;
}

.action-link svg {
  width: 14px;
  height: 14px;
}

.action-divider {
  width: 1px;
  height: 16px;
  background: #e2e8f0;
  flex-shrink: 0;
}

/* ---- 分页 ---- */
.pagination-wrap {
  padding-top: 20px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 34px;
  padding: 0 8px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  font-family: var(--shop-font-sans);
  color: #475569;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s ease;
}

.page-btn:hover:not(.page-btn--active):not(.page-btn--disabled) {
  border-color: #6366f1;
  color: #6366f1;
  background: rgba(99,102,241,0.06);
}

.page-btn--active {
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 2px 8px rgba(99,102,241,0.2);
}

.page-btn--disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-btn--ellipsis {
  border: none;
  cursor: default;
  background: transparent;
  min-width: 24px;
}

.pagination-info {
  text-align: center;
  font-size: 0.75rem;
  color: #94a3b8;
  padding-top: 4px;
}

/* ============================================
   弹窗样式
   ============================================ */

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15,23,42,0.5);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.modal {
  background: #fff;
  border-radius: 20px;
  width: 100%;
  max-width: 580px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 25px 60px rgba(0,0,0,0.15);
  animation: modalIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.modal-wide {
  max-width: 740px;
}

.modal-sm {
  max-width: 400px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
  position: sticky;
  top: 0;
  background: #fff;
  z-index: 1;
}

.modal-header h2 {
  font-size: 1.0625rem;
  font-weight: 700;
  color: #0f172a;
}

.modal-close {
  width: 2rem;
  height: 2rem;
  border: none;
  background: transparent;
  border-radius: 999px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 1.25rem;
  transition: all 0.15s ease;
}

.modal-close:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.modal-body {
  padding: 1.5rem;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #f1f5f9;
  position: sticky;
  bottom: 0;
  background: #fff;
}

/* ---- 表单控件 ---- */
.form-group {
  margin-bottom: 1.125rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 0.375rem;
}

.form-label .required {
  color: #ef4444;
}

.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  color: #0f172a;
  background: #fff;
  font-family: var(--shop-font-sans);
  outline: none;
  transition: all 0.2s ease;
}

.form-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.form-textarea {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  color: #0f172a;
  background: #fff;
  font-family: var(--shop-font-sans);
  outline: none;
  resize: vertical;
  transition: all 0.2s ease;
}

.form-textarea:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.form-select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  color: #0f172a;
  background: #fff;
  font-family: var(--shop-font-sans);
  outline: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.form-select:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.form-row {
  display: flex;
  gap: 1rem;
}

.form-row .form-group {
  flex: 1;
}

.cascading-selects {
  display: flex;
  gap: 8px;
}

.cascading-selects .form-select {
  flex: 1;
}

.btn-save {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1.125rem;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: var(--shop-font-sans);
}

.btn-save:hover {
  box-shadow: 0 4px 12px rgba(99,102,241,0.3);
  transform: translateY(-1px);
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444, #dc2626) !important;
}

.btn-cancel {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1.125rem;
  background: #fff;
  color: #475569;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: var(--shop-font-sans);
}

.btn-cancel:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.confirm-msg {
  font-size: 0.875rem;
  color: #475569;
  line-height: 1.5;
}

/* ---- 图片上传 ---- */
.upload-area {
  border: 2px dashed #e2e8f0;
  border-radius: 12px;
  padding: 1rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f8fafc;
}

.upload-area:hover {
  border-color: #6366f1;
  background: rgba(99,102,241,0.04);
}

.upload-area .upload-icon {
  font-size: 1.5rem;
  color: #94a3b8;
  margin-bottom: 0.25rem;
}

.upload-area .upload-text {
  font-size: 0.75rem;
  color: #64748b;
}

.upload-area .upload-hint {
  font-size: 0.625rem;
  color: #94a3b8;
  margin-top: 0.125rem;
}

.upload-area.has-image {
  border-style: solid;
  border-color: #6366f1;
  padding: 0.5rem;
}

.upload-area.has-image .upload-placeholder {
  display: none;
}

.upload-area .upload-preview {
  max-width: 80px;
  margin: 0 auto;
}

.upload-area .upload-preview img {
  width: 100%;
  border-radius: 6px;
  display: block;
}

/* ---- 规格构建器 ---- */
.spec-section-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #0f172a;
  border-top: 1px solid #f1f5f9;
  padding-top: 1rem;
  margin-top: 0.5rem;
  margin-bottom: 0.75rem;
}

.spec-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px 12px;
  background: #f8fafc;
  border-radius: 10px;
  margin-bottom: 8px;
  border: 1px solid #f1f5f9;
}

.spec-row .spec-name-wrap {
  width: 120px;
  flex-shrink: 0;
}

.spec-row .spec-values-wrap {
  flex: 1;
}

.spec-row .spec-remove-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  border-radius: 6px;
  flex-shrink: 0;
  margin-top: 1.5rem;
  transition: all 0.15s ease;
  font-size: 1.125rem;
}

.spec-row .spec-remove-btn:hover {
  background: #fef2f2;
  color: #ef4444;
}

.tag-input-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
  padding: 4px 6px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
  min-height: 34px;
  cursor: text;
}

.tag-input-wrap:focus-within {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
  background: #eef2ff;
  color: #6366f1;
  border: 1px solid rgba(99,102,241,0.15);
}

.tag-badge .tag-remove {
  cursor: pointer;
  font-size: 0.875rem;
  line-height: 1;
  color: #6366f1;
  opacity: 0.6;
  margin-left: 2px;
}

.tag-badge .tag-remove:hover {
  opacity: 1;
}

.tag-input-wrap input {
  border: none;
  outline: none;
  font-size: 0.8125rem;
  font-family: var(--shop-font-sans);
  color: #0f172a;
  background: transparent;
  flex: 1;
  min-width: 80px;
  padding: 2px 0;
}

.add-spec-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #64748b;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: var(--shop-font-sans);
  margin-top: 4px;
}

.add-spec-btn:hover {
  border-color: #6366f1;
  color: #6366f1;
  background: rgba(99,102,241,0.04);
}

/* ---- SKU 矩阵 ---- */
.sku-matrix-wrap {
  margin-top: 1rem;
}

.sku-matrix-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 6px;
}

.sku-matrix-scroll {
  overflow-x: auto;
}

.sku-matrix {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.75rem;
  border: 1px solid #f1f5f9;
  border-radius: 8px;
  overflow: hidden;
}

.sku-matrix thead {
  background: #f8fafc;
}

.sku-matrix thead th {
  padding: 8px 10px;
  text-align: left;
  font-weight: 600;
  color: #64748b;
  font-size: 0.6875rem;
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.sku-matrix tbody td {
  padding: 7px 10px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.sku-matrix tbody tr:last-child td {
  border-bottom: none;
}

.sku-matrix tbody tr:hover {
  background: rgba(99,102,241,0.03);
}

.sku-price-input {
  width: 80px;
  padding: 4px 6px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.75rem;
  font-family: var(--shop-font-sans);
  color: #0f172a;
  outline: none;
  transition: all 0.15s ease;
}

.sku-price-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.1);
}

.sku-stock-input {
  width: 60px;
  padding: 4px 6px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.75rem;
  font-family: var(--shop-font-sans);
  color: #0f172a;
  outline: none;
  transition: all 0.15s ease;
}

.sku-stock-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.1);
}

.sku-del-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s ease;
  font-size: 1rem;
}

.sku-del-btn:hover {
  background: #fef2f2;
  color: #ef4444;
}

.spec-combo {
  display: flex;
  gap: 3px;
  flex-wrap: wrap;
}

.combo-tag {
  padding: 1px 6px;
  background: #f1f5f9;
  border-radius: 4px;
  font-size: 0.6875rem;
  color: #475569;
}

.sku-empty-hint {
  text-align: center;
  padding: 16px;
  color: #94a3b8;
  font-size: 0.8125rem;
}

/* ---- SKU 详情弹窗 ---- */
.sku-popup {
  background: #fff;
  border-radius: 20px;
  width: 100%;
  max-width: 680px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 25px 60px rgba(0,0,0,0.15);
  animation: modalIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.sku-popup-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
  position: sticky;
  top: 0;
  background: #fff;
  z-index: 1;
}

.sku-popup-header h2 {
  font-size: 1.0625rem;
  font-weight: 700;
  color: #0f172a;
}

.sku-popup-body {
  padding: 1.25rem 1.5rem;
}

.popup-spu-desc {
  font-size: 0.8125rem;
  color: #64748b;
  margin-bottom: 1rem;
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.popup-spu-desc span {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.popup-spu-desc .label {
  color: #94a3b8;
}

.sku-detail-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.75rem;
}

.sku-detail-table thead {
  background: #f8fafc;
}

.sku-detail-table thead th {
  padding: 8px 10px;
  text-align: left;
  font-weight: 600;
  color: #64748b;
  font-size: 0.6875rem;
  border-bottom: 1px solid #e2e8f0;
}

.sku-detail-table tbody td {
  padding: 8px 10px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.sku-detail-table tbody tr:last-child td {
  border-bottom: none;
}

.sku-detail-table tbody tr:hover {
  background: rgba(99,102,241,0.03);
}

.sku-price {
  font-weight: 600;
  color: #0f172a;
}

/* ---- 动画 ---- */
@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .toolbar {
    flex-wrap: wrap;
  }
  .search-input-wrap {
    max-width: 100%;
  }
}

@media (max-width: 767px) {
  .tab-item {
    padding: 8px 10px;
    font-size: 0.75rem;
  }
  .form-row {
    flex-direction: column;
  }
  .cascading-selects {
    flex-direction: column;
  }
  .spec-row {
    flex-direction: column;
  }
  .spec-row .spec-name-wrap {
    width: 100%;
  }
  .modal-wide {
    max-width: 100%;
    margin: 0 0.5rem;
  }
  .sku-popup {
    max-width: 100%;
    margin: 0 0.5rem;
  }
  .product-name-cell .name-text {
    max-width: 120px;
  }
  .action-link span {
    display: none;
  }
}
</style>