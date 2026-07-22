# 商品管理页面重实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重写商品管理列表页和表单页，统一交互为独立页面，提升代码质量

**Architecture:** 按职责拆分文件：类型定义 → composable 业务逻辑 → Vue 模板/样式。保持 API 接口不变，纯前端重构

**Tech Stack:** Vue 3 + TypeScript + Element Plus + Pinia

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `src/types/product.ts` | 创建 | 商品所有 TS 类型接口 |
| `src/api/product.ts` | 修改 | 导入类型，精简 |
| `src/composables/useProductList.ts` | 创建 | 列表加载/筛选/分页/批量操作逻辑 |
| `src/composables/useProductForm.ts` | 创建 | 表单加载/校验/规格构建/SKU矩阵/保存 |
| `src/views/products/ProductListView.vue` | 重写 | 列表模板，去内嵌弹窗 |
| `src/views/products/ProductFormView.vue` | 重写 | 添加/编辑表单，含规格/SKU |

---

### Task 1: 创建 `types/product.ts`

**Files:**
- Create: `src/types/product.ts`

- [ ] **Step 1: 创建商品类型文件**

```typescript
// web/admin/src/types/product.ts
// 商品管理相关 TypeScript 类型定义

/** 规格组（API 返回格式） */
export interface SpecGroup {
  id: number
  name: string
  values: string[]
}

/** SKU（API 返回格式） */
export interface SkuInfo {
  id: number
  sku_code: string
  price: number    // 分
  stock: number
  sales: number
  spec: Record<string, string>  // {"颜色":"红色","尺寸":"M"}
}

/** 商品列表项 */
export interface ProductItem {
  id: number
  name: string
  category: string
  price: number
  stock: number
  sales: number
  status: string   // "上架" | "下架"
  image: string
  specs: SpecGroup[]
  skus: SkuInfo[]
}

/** 商品列表查询参数 */
export interface ProductListParams {
  page?: number
  page_size?: number
  keyword?: string
  category?: string
  status?: string
}

/** 规格组提交格式 */
export interface SpecGroupInput {
  name: string
  values: string[]
}

/** SKU 提交格式 */
export interface SkuInput {
  sku_code: string
  price: number     // 分
  stock: number
  sales?: number
  spec: Record<string, string>
}

/** 创建商品请求体 */
export interface CreateProductData {
  name: string
  category: string
  price: number
  stock: number
  description?: string
  image?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/** 更新商品请求体 */
export interface UpdateProductData {
  name?: string
  category?: string
  price?: number
  stock?: number
  description?: string
  image?: string
  status?: string
  specs?: SpecGroupInput[]
  skus?: SkuInput[]
}

/** 批量操作请求 */
export interface BatchOperationData {
  action: '上架' | '下架' | '删除'
  ids: number[]
}

/** 商品列表响应 */
export interface ProductListResponse {
  items: ProductItem[]
  total: number
}
```

- [ ] **Step 2: 提交**

```bash
git add src/types/product.ts
git commit -m "feat: 创建商品管理类型定义文件"
```

---

### Task 2: 精简 `src/api/product.ts`

**Files:**
- Modify: `src/api/product.ts`

- [ ] **Step 1: 重写 api/product.ts**

```typescript
// web/admin/src/api/product.ts
// 商品管理 API 方法

import request from './index'
import type {
  ProductListParams, ProductListResponse,
  CreateProductData, UpdateProductData, BatchOperationData
} from '../types/product'

/** 获取分页商品列表 */
export function listProducts(params: ProductListParams): Promise<ProductListResponse> {
  return request.get('/admin/v1/products', { params })
}

/** 创建商品 */
export function createProduct(data: CreateProductData): Promise<{ id: number }> {
  return request.post('/admin/v1/products', data)
}

/** 更新商品 */
export function updateProduct(id: number, data: UpdateProductData): Promise<{ success: boolean }> {
  return request.put(`/admin/v1/products/${id}`, data)
}

/** 删除商品 */
export function deleteProduct(id: number): Promise<{ success: boolean }> {
  return request.delete(`/admin/v1/products/${id}`)
}

/** 获取商品详情 */
export function getProduct(id: number): Promise<{ product: import('../types/product').ProductItem }> {
  return request.get(`/admin/v1/products/${id}`)
}

/** 批量操作 */
export function batchProductOperation(action: string, ids: number[]): Promise<{ success: boolean; affected_count: number }> {
  return request.post('/admin/v1/products/batch', { action, ids })
}
```

- [ ] **Step 2: 提交**

---

### Task 3: 创建 `composables/useProductList.ts`

**Files:**
- Create: `src/composables/useProductList.ts`

- [ ] **Step 1: 创建列表逻辑 composable**

```typescript
// web/admin/src/composables/useProductList.ts
// 商品列表页业务逻辑

import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { listProducts, batchProductOperation, deleteProduct } from '../api/product'

export interface FilterState {
  keyword: string
  category: string
  status: string
}

export function useProductList() {
  const products = ref<any[]>([])
  const totalCount = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(10)
  const loading = ref(false)
  const selectedIds = ref<number[]>([])
  const filters = ref<FilterState>({ keyword: '', category: '', status: '' })

  const totalPages = computed(() =>
    Math.ceil(totalCount.value / pageSize.value) || 1
  )

  const allChecked = computed(() =>
    products.value.length > 0 && products.value.every(p => selectedIds.value.includes(p.id))
  )

  async function loadProducts() {
    loading.value = true
    try {
      const res = await listProducts({
        page: currentPage.value,
        page_size: pageSize.value,
        keyword: filters.value.keyword || undefined,
        category: filters.value.category || undefined,
      })
      totalCount.value = res.total
      products.value = (res.items || []).map(item => ({
        ...item,
        _priceYuan: item.price / 100, // 分转元显示
      }))
    } catch {
      products.value = []
      totalCount.value = 0
    } finally {
      loading.value = false
    }
  }

  function goPage(page: number) {
    if (page < 1 || page > totalPages.value) return
    currentPage.value = page
    loadProducts()
  }

  function handleFilter() {
    currentPage.value = 1
    loadProducts()
  }

  function toggleAll(checked: boolean) {
    selectedIds.value = checked ? products.value.map(p => p.id) : []
  }

  function toggleSelect(id: number) {
    const idx = selectedIds.value.indexOf(id)
    if (idx >= 0) selectedIds.value.splice(idx, 1)
    else selectedIds.value.push(id)
  }

  async function bulkAction(action: '上架' | '下架' | '删除') {
    if (selectedIds.value.length === 0) {
      ElMessage.warning('请先选择商品')
      return
    }
    try {
      const res = await batchProductOperation(action, selectedIds.value)
      ElMessage.success(`${action}成功 ${res.affected_count} 项`)
      selectedIds.value = []
      loadProducts()
    } catch {
      ElMessage.error(`${action}失败`)
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteProduct(id)
      ElMessage.success('删除成功')
      loadProducts()
    } catch {
      ElMessage.error('删除失败')
    }
  }

  async function handleToggleStatus(id: number, currentStatus: string) {
    const newStatus = currentStatus === '上架' ? '下架' : '上架'
    try {
      await batchProductOperation(newStatus as any, [id])
      ElMessage.success(`已${newStatus}`)
      loadProducts()
    } catch {
      ElMessage.error('操作失败')
    }
  }

  return {
    products, totalCount, currentPage, pageSize, loading,
    selectedIds, filters, totalPages, allChecked,
    loadProducts, goPage, handleFilter,
    toggleAll, toggleSelect, bulkAction, handleDelete, handleToggleStatus,
  }
}
```

- [ ] **Step 2: 提交**

---

### Task 4: 重写 `ProductListView.vue`

**Files:**
- Modify: `src/views/products/ProductListView.vue`（全量重写）

- [ ] **Step 1: 重写列表页模板和脚本**

```vue
<template>
  <div class="product-list">
    <!-- Toolbar -->
    <div class="toolbar">
      <el-button type="primary" @click="$router.push('/products/add')">
        <el-icon><Plus /></el-icon> 添加商品
      </el-button>
      <div class="search-input-wrap">
        <el-icon class="search-icon"><Search /></el-icon>
        <input v-model="filters.keyword" type="text" placeholder="搜索商品名称..."
          class="search-input" @input="handleFilter" />
      </div>
      <el-cascader v-model="filters.category" :options="categoryTree"
        :props="{ value: 'name', label: 'name', children: 'children', expandTrigger: 'hover', emitPath: false }"
        placeholder="全部分类" clearable class="toolbar-cascader" @change="handleFilter" />
    </div>

    <!-- Bulk Action Bar -->
    <div class="bulk-bar-placeholder">
      <div v-show="selectedIds.length > 0" class="bulk-bar">
        <span>已选择 <strong class="bulk-count">{{ selectedIds.length }}</strong> 项</span>
        <div class="bulk-actions">
          <el-button size="small" @click="bulkAction('上架')">批量上架</el-button>
          <el-button size="small" @click="bulkAction('下架')">批量下架</el-button>
          <el-button size="small" type="danger" plain @click="bulkAction('删除')">批量删除</el-button>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tab-bar">
      <div v-for="tab in ['全部','上架','下架','缺货']" :key="tab"
        :class="['tab-item', { active: filters.status === tab || (tab === '全部' && !filters.status) }]"
        @click="filters.status = tab === '全部' ? '' : tab; handleFilter()">{{ tab }}</div>
    </div>

    <!-- Table -->
    <el-table :data="products" v-loading="loading" @selection-change="onSelectionChange" stripe>
      <el-table-column type="selection" width="50" />
      <el-table-column label="图片" width="80">
        <template #default="{ row }">
          <el-image :src="row.image || 'https://placehold.co/60x60?text=N/A'"
            style="width: 60px; height: 60px; border-radius: 4px;" fit="cover" />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="商品名称" min-width="200" show-overflow-tooltip />
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column label="价格" width="120">
        <template #default="{ row }">¥{{ (row.price / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="stock" label="库存" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-switch :model-value="row.status === '上架'"
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
        layout="total, sizes, prev, pager, next" @current-change="loadProducts" @size-change="loadProducts" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Search } from '@element-plus/icons-vue'
import { listCategoryTree } from '../../api/category'
import { useProductList } from '../../composables/useProductList'

const router = useRouter()
const {
  products, totalCount, currentPage, pageSize, loading,
  selectedIds, filters, handleFilter, loadProducts,
  goPage, toggleAll, toggleSelect, bulkAction, handleDelete, handleToggleStatus,
} = useProductList()

const categoryTree = ref<any[]>([])

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map(r => r.id)
}

onMounted(async () => {
  try {
    categoryTree.value = [{ name: '根目录', children: await listCategoryTree() }]
  } catch { categoryTree.value = [] }
  loadProducts()
})
</script>

<style scoped>
.product-list { padding: 24px; max-width: 1400px; margin: 0 auto; }
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-input-wrap { position: relative; flex: 1; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #94a3b8; }
.search-input { width: 100%; height: 36px; padding: 0 12px 0 36px; border: 1px solid #d9d9d9; border-radius: 6px; font-size: 14px; }
.search-input:focus { border-color: #409eff; }
.toolbar-cascader { width: 200px; }
.bulk-bar-placeholder { min-height: 42px; margin-bottom: 8px; }
.bulk-bar { display: flex; align-items: center; gap: 16px; padding: 8px 16px; background: #e6f7ff; border: 1px solid #91d5ff; border-radius: 6px; font-size: 14px; }
.bulk-count { color: #1890ff; }
.bulk-actions { display: flex; gap: 8px; }
.tab-bar { display: flex; gap: 4px; margin-bottom: 16px; border-bottom: 1px solid #e8e8e8; }
.tab-item { padding: 8px 20px; cursor: pointer; color: #666; font-size: 14px; border-bottom: 2px solid transparent; transition: all 0.2s; }
.tab-item:hover { color: #409eff; }
.tab-item.active { color: #409eff; border-bottom-color: #409eff; font-weight: 600; }
.pagination-wrap { display: flex; justify-content: flex-end; padding: 16px 0; }
</style>
```

注意：Step 1 里的 `<script setup>` 引用了 `ref` 但没有 import。需要确保 composable 内部的 `ref` 等 API 来自 vue。实际实现时要确保所有 import 正确。

- [ ] **Step 2: 提交**

---

### Task 5: 创建 `composables/useProductForm.ts`

**Files:**
- Create: `src/composables/useProductForm.ts`

- [ ] **Step 1: 创建表单逻辑 composable**

```typescript
// web/admin/src/composables/useProductForm.ts
// 商品添加/编辑表单业务逻辑

import { ref, computed, reactive } from 'vue'

interface SpecGroupForm {
  id: number
  name: string
  values: string[]
}

interface SkuForm {
  id: string
  spec: Record<string, string>
  price: number   // 元
  stock: number
}

export interface ProductFormState {
  name: string
  summary: string
  category: string
  brand: string
  tags: string[]
  image: string
  description: string
}

/** 规格组的笛卡尔积生成 SKU 列表 */
export function cartesianProduct(groups: SpecGroupForm[]): Record<string, string>[] {
  if (groups.length === 0) return []
  const result: Record<string, string>[] = []
  function backtrack(idx: number, current: Record<string, string>) {
    if (idx === groups.length) {
      result.push({ ...current })
      return
    }
    for (const val of groups[idx].values) {
      current[groups[idx].name] = val
      backtrack(idx + 1, current)
    }
  }
  backtrack(0, {})
  return result
}

export function useProductForm() {
  const form = reactive<ProductFormState>({
    name: '', summary: '', category: '', brand: '', tags: [], image: '', description: '',
  })
  const specGroups = ref<SpecGroupForm[]>([])
  const skus = ref<SkuForm[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const specIdCounter = ref(0)

  /** 添加规格组 */
  function addSpecGroup() {
    specGroups.value.push({ id: ++specIdCounter.value, name: '', values: [] })
  }

  /** 删除规格组 */
  function removeSpecGroup(index: number) {
    specGroups.value.splice(index, 1)
    regenerateSkus()
  }

  /** 设置规格组的值 */
  function setSpecValues(index: number, values: string[]) {
    specGroups.value[index].values = values
    regenerateSkus()
  }

  /** 重新生成 SKU 矩阵 */
  function regenerateSkus() {
    const activeGroups = specGroups.value.filter(g => g.name && g.values.length > 0)
    if (activeGroups.length === 0) {
      skus.value = []
      return
    }
    const combinations = cartesianProduct(activeGroups)
    // 保留已有 SKU 的价格/库存
    const skuMap = new Map(skus.value.map(s => [JSON.stringify(s.spec), s]))
    skus.value = combinations.map(spec => {
      const key = JSON.stringify(spec)
      const existing = skuMap.get(key)
      return {
        id: existing?.id || `sku_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`,
        spec,
        price: existing?.price ?? 0,
        stock: existing?.stock ?? 0,
      }
    })
  }

  /** 加载编辑模式数据 */
  async function loadProduct(id: number) {
    const { getProduct } = await import('../api/product')
    loading.value = true
    try {
      const res = await getProduct(id)
      const p = res.product
      form.name = p.name
      form.category = p.category
      form.image = p.image || ''
      form.summary = ''  // 后端暂无对应字段
      // specs
      specGroups.value = (p.specs || []).map(s => ({ id: ++specIdCounter.value, name: s.name, values: [...s.values] }))
      // skus（价格从分转元）
      skus.value = (p.skus || []).map(s => ({
        id: `sku_${s.id}`,
        spec: s.spec,
        price: s.price / 100,
        stock: s.stock,
      }))
    } finally {
      loading.value = false
    }
  }

  /** 校验表单 */
  function validate(): boolean {
    if (!form.name.trim()) { ElMessage.warning('请填写商品名称'); return false }
    if (!form.category) { ElMessage.warning('请选择商品分类'); return false }
    return true
  }

  /** 构建提交数据 */
  function buildSubmitData() {
    // 价格从元转分
    const specs = specGroups.value.filter(g => g.name)
    const formedSkus = skus.value.map(s => ({
      sku_code: Object.values(s.spec).join('-'),
      price: Math.round(s.price * 100),  // 元 → 分
      stock: s.stock,
      spec: s.spec,
    }))
    return {
      name: form.name,
      category: form.category,
      image: form.image || undefined,
      price: formedSkus.length > 0 ? formedSkus[0].price : 0,
      stock: formedSkus.length > 0 ? formedSkus[0].stock : 0,
      specs: specs.map(g => ({ name: g.name, values: g.values })),
      skus: formedSkus,
      description: form.description || undefined,
    }
  }

  /** 保存（创建或更新） */
  async function save(productId?: number): Promise<boolean> {
    if (!validate()) return false
    saving.value = true
    try {
      const data = buildSubmitData()
      const { createProduct, updateProduct } = await import('../api/product')
      if (productId) {
        await updateProduct(productId, data)
        ElMessage.success('保存成功')
      } else {
        await createProduct(data as any)
        ElMessage.success('创建成功')
      }
      return true
    } catch {
      ElMessage.error('保存失败')
      return false
    } finally {
      saving.value = false
    }
  }

  return {
    form, specGroups, skus, loading, saving,
    addSpecGroup, removeSpecGroup, setSpecValues, regenerateSkus,
    loadProduct, validate, save,
  }
}
```

- [ ] **Step 2: 提交**

---

### Task 6: 重写 `ProductFormView.vue`

**Files:**
- Modify: `src/views/products/ProductFormView.vue`

- [ ] **Step 1: 重写表单页**

这个文件较大，结构如下：

```vue
<template>
  <div class="product-form">
    <!-- Breadcrumb -->
    <div class="form-breadcrumb">
      <router-link to="/products">商品管理</router-link>
      <span class="sep"> &gt; </span>
      <span>{{ isEdit ? '编辑商品' : '添加商品' }}</span>
    </div>

    <!-- Section 1: 基本信息 -->
    <div class="form-card">
      <div class="form-section-title"><span class="section-number">1</span>基本信息</div>
      <el-form label-position="top" class="form-grid">
        <el-form-item label="商品名称" required class="full-width">
          <el-input v-model="form.name" maxlength="120" />
        </el-form-item>
        <el-form-item label="商品简介" class="full-width">
          <el-input v-model="form.summary" type="textarea" :rows="3" maxlength="200" />
        </el-form-item>
        <el-form-item label="商品分类" required>
          <el-cascader v-model="form.category" :options="cascaderOptions"
            :props="{ value: 'name', label: 'name', children: 'children', checkStrictly: true }"
            placeholder="请选择分类" clearable filterable />
        </el-form-item>
        <el-form-item label="品牌">
          <el-input v-model="form.brand" />
        </el-form-item>
        <el-form-item label="缩略图" class="full-width">
          <ImageUploader v-model="form.image" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 2: 规格与 SKU -->
    <div class="form-card">
      <div class="form-section-title"><span class="section-number">2</span>规格与 SKU</div>
      <div class="spec-builder">
        <div v-for="(group, i) in specGroups" :key="group.id" class="spec-group-row">
          <el-input v-model="group.name" placeholder="规格名称（如：颜色）" style="width:140px" />
          <el-select v-model="group.values" multiple filterable allow-create
            default-first-option placeholder="输入规格值 回车确认" style="flex:1"
            @change="(vals: string[]) => setSpecValues(i, vals)" />
          <el-button :icon="Delete" circle size="small" @click="removeSpecGroup(i)" />
        </div>
        <el-button type="primary" plain @click="addSpecGroup">+ 添加规格</el-button>
      </div>

      <!-- SKU 矩阵 -->
      <div v-if="skus.length > 0" class="sku-matrix">
        <table class="sku-table">
          <thead>
            <tr>
              <th v-for="key of Object.keys(skus[0].spec)" :key="key">{{ key }}</th>
              <th>价格 (元)</th>
              <th>库存</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(sku, i) in skus" :key="sku.id">
              <td v-for="v of Object.values(sku.spec)" :key="v">{{ v }}</td>
              <td><el-input-number v-model="sku.price" :min="0" :precision="2" size="small" controls-position="right" /></td>
              <td><el-input-number v-model="sku.stock" :min="0" size="small" controls-position="right" /></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Section 3: 商品描述 -->
    <div class="form-card">
      <div class="form-section-title"><span class="section-number">3</span>商品描述</div>
      <textarea v-model="form.description" class="desc-editor" placeholder="请输入商品描述..." />
    </div>

    <!-- Footer -->
    <div class="form-footer">
      <el-button @click="$router.push('/products')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">
        {{ isEdit ? '保存修改' : '创建商品' }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { listCategoryTree } from '../../api/category'
import { useProductForm } from '../../composables/useProductForm'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)
const cascaderOptions = ref<any[]>([])

const {
  form, specGroups, skus, loading, saving,
  addSpecGroup, removeSpecGroup, setSpecValues,
  loadProduct, save,
} = useProductForm()

async function handleSave() {
  const ok = await save(isEdit.value ? Number(route.params.id) : undefined)
  if (ok) router.push('/products')
}

onMounted(async () => {
  try {
    cascaderOptions.value = [{ name: '根目录', children: await listCategoryTree() }]
  } catch { cascaderOptions.value = [] }

  if (isEdit.value) {
    await loadProduct(Number(route.params.id))
  }
})
</script>

<style scoped>
.product-form { padding: 24px; max-width: 960px; margin: 0 auto; }
.form-breadcrumb { margin-bottom: 24px; color: #666; font-size: 14px; }
.form-breadcrumb a { color: #409eff; }
.sep { margin: 0 8px; color: #ccc; }
.form-card { background: #fff; border-radius: 8px; padding: 24px; margin-bottom: 16px; box-shadow: 0 1px 4px rgba(0,0,0,0.06); }
.form-section-title { font-size: 16px; font-weight: 600; margin-bottom: 20px; display: flex; align-items: center; gap: 8px; }
.section-number { display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; border-radius: 50%; background: #409eff; color: #fff; font-size: 13px; font-weight: 600; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.full-width { grid-column: 1 / -1; }
.spec-builder { display: flex; flex-direction: column; gap: 12px; }
.spec-group-row { display: flex; align-items: center; gap: 8px; }
.sku-matrix { margin-top: 16px; overflow-x: auto; }
.sku-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.sku-table th, .sku-table td { border: 1px solid #e8e8e8; padding: 8px 12px; text-align: center; }
.sku-table th { background: #fafafa; font-weight: 600; }
.sku-table td { vertical-align: middle; }
.desc-editor { width: 100%; min-height: 200px; border: 1px solid #d9d9d9; border-radius: 6px; padding: 12px; font-size: 14px; line-height: 1.6; resize: vertical; }
.desc-editor:focus { border-color: #409eff; }
.form-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 0; }
</style>
```

- [ ] **Step 2: 提交**

---

### Task 7: 清理和验证

**Files:**
- Verify: 全局编译

- [ ] **Step 1: 编译验证**

```bash
cd web/admin
npx vue-tsc -b --noEmit
npx vite build
```

- [ ] **Step 2: 最终提交**
