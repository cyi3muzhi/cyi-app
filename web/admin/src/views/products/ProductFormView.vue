<template>
  <!-- 添加/编辑商品表单 -->
  <div class="product-form">
    <!-- Breadcrumb -->
    <div class="form-breadcrumb">
      <router-link to="/products">商品管理</router-link>
      <span class="sep"> &gt; </span>
      <span>{{ isEdit ? '编辑商品' : '添加商品' }}</span>
    </div>

    <!-- Section 1: 基本信息 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">1</span>基本信息
      </div>
      <el-form label-position="top" class="form-grid">
        <el-form-item label="商品名称" required class="full-width">
          <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="120" />
        </el-form-item>
        <el-form-item label="商品分类" required>
          <el-cascader v-if="cascaderReady" v-model="form.category_id" :options="cascaderOptions"
            :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true, emitPath: false }"
            placeholder="请选择分类" clearable filterable />
        </el-form-item>
        <el-form-item label="商品主图" class="full-width">
          <div class="upload-area">
            <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleFileChange">
              <el-button type="primary">选择图片</el-button>
            </el-upload>
          </div>
          <div v-if="form.image" class="main-image-preview">
            <img :src="form.image" class="preview-img" />
          </div>
          <div v-else class="image-placeholder">请选择商品主图</div>
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 2: 规格与 SKU -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">2</span>规格与 SKU
      </div>
      <div class="spec-builder">
        <div v-for="(group, i) in specGroups" :key="group.id" class="spec-group-row">
          <el-input v-model="group.name" placeholder="规格名称（如：颜色）" style="width:140px" />
          <el-select v-model="group.values" multiple filterable allow-create
            default-first-option placeholder="输入规格值 回车确认" style="flex:1"
            @change="(vals: any) => setSpecValues(i, vals)" />
          <el-button :icon="Delete" circle size="small" type="danger" @click="removeSpecGroup(i)" />
        </div>
        <el-button type="primary" plain @click="addSpecGroup">
          <el-icon><Plus /></el-icon> 添加规格
        </el-button>
      </div>

      <!-- SKU 矩阵 -->
      <div v-if="skus.length > 0" class="sku-matrix">
        <h4 class="sku-title">SKU 列表（共 {{ skus.length }} 项）</h4>

        <!-- 批量操作栏 -->
        <div class="sku-bulk-bar">
          <div class="sku-bulk-row">
            <span class="bulk-label">价格</span>
            <el-input-number v-model="bulkPrice" :min="0" :precision="2" size="small" style="width:110px" />
            <el-button size="small" @click="applyBulkPrice">全部</el-button>
            <el-button size="small" :disabled="skuSelection.length === 0" @click="applyBulkPriceToSelected">选中</el-button>
            <el-button size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkPriceToFiltered">筛选({{ filteredSkuIds.length }})</el-button>
          </div>
          <div class="sku-bulk-row">
            <span class="bulk-label">库存</span>
            <el-input-number v-model="bulkStock" :min="0" size="small" style="width:110px" />
            <el-button size="small" @click="applyBulkStock">全部</el-button>
            <el-button size="small" :disabled="skuSelection.length === 0" @click="applyBulkStockToSelected">选中</el-button>
            <el-button size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkStockToFiltered">筛选({{ filteredSkuIds.length }})</el-button>
          </div>
          <div class="sku-bulk-row">
            <span class="bulk-label">图片</span>
            <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleBulkSkuImageUpload">
              <el-button v-if="!bulkSkuImage" size="small">选择</el-button>
            </el-upload>
            <img v-if="bulkSkuImage" :src="bulkSkuImage" style="width:28px;height:28px;border-radius:4px;object-fit:cover;" />
            <el-button v-if="bulkSkuImage" size="small" @click="applyBulkSkuImage(skus, '全部', skus.length)">全部</el-button>
            <el-button v-if="bulkSkuImage" size="small" :disabled="skuSelection.length === 0" @click="applyBulkSkuImage(skus.filter(s => skuSelection.includes(s.id)), '选中', skuSelection.length)">选中</el-button>
            <el-button v-if="bulkSkuImage" size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkSkuImage(skus.filter(s => filteredSkuIds.includes(s.id)), '筛选', filteredSkuIds.length)">筛选</el-button>
            <el-button v-if="bulkSkuImage" size="small" text type="info" @click="bulkSkuImage = ''">取消</el-button>
          </div>
        </div>

        <!-- 规格筛选栏 -->
        <div class="sku-filter-bar">
          <template v-if="activeSpecs.length > 0">
            <span class="filter-bar-label">规格筛选：</span>
            <template v-for="group in activeSpecs" :key="group.name">
              <span class="filter-group-label">{{ group.name }}</span>
              <button
                v-for="val in group.values" :key="val"
                :class="['filter-chip', { active: isSpecValueSelected(group.name, val) }]"
                @click="toggleSpecFilter(group.name, val)">
                {{ val }}
              </button>
            </template>
            <span v-if="filteredSkuIds.length > 0" class="filter-count">匹配 {{ filteredSkuIds.length }} 项</span>
            <el-button v-if="hasActiveFilter" size="small" text type="info" @click="clearSpecFilter">清除筛选</el-button>
            <el-divider direction="vertical" />
          </template>
          <span style="color:#999;font-size:12px;">已选 {{ skuSelection.length }} 项</span>
          <el-button size="small" text :disabled="skuSelection.length === 0" @click="skuSelection = []">清除选中</el-button>
        </div>

        <table class="sku-table">
          <thead>
            <tr>
              <th style="width:36px">
                <el-checkbox :indeterminate="isIndeterminate" :model-value="isAllSelected" @change="toggleAllSelection" />
              </th>
              <th v-for="key of Object.keys(skus[0].spec)" :key="key">{{ key }}</th>
              <th>价格 (元)</th>
              <th>库存</th>
              <th>图片</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(sku, i) in skus" :key="sku.id" :class="{ 'sku-row-selected': skuSelection.includes(sku.id), 'sku-row-dimmed': hasActiveFilter && !filteredSkuIds.includes(sku.id) }">
              <td>
                <el-checkbox :model-value="skuSelection.includes(sku.id)" @change="(val: boolean) => toggleSkuSelection(sku.id, val)" />
              </td>
              <td v-for="v of Object.values(sku.spec)" :key="v">{{ v }}</td>
              <td>
                <el-input-number v-model="sku.price" :min="0" :precision="2" size="small" controls-position="right" style="width:130px" />
              </td>
              <td>
                <el-input-number v-model="sku.stock" :min="0" size="small" controls-position="right" style="width:130px" />
              </td>
              <td>
                <div style="display:flex;align-items:center;gap:4px;">
                  <img v-if="sku.image" :src="sku.image" style="width:36px;height:36px;border-radius:4px;object-fit:cover;cursor:pointer;" @click="previewCarousel(sku.image)" />
                  <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="(f: any) => uploadSkuImage(f, i)">
                    <el-button size="small">选择</el-button>
                  </el-upload>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Section 3: 商品轮播图 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">3</span>商品轮播图
      </div>
      <div class="gallery-grid">
        <div v-for="(img, i) in form.images" :key="i" class="gallery-item">
          <img :src="img" class="gallery-thumb" @click="previewCarousel(img)" />
          <el-button size="small" circle type="danger" class="gallery-remove" @click="removeCarouselImage(i)">x</el-button>
        </div>
        <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleCarouselUpload">
          <div class="gallery-add">
            <el-icon><Plus /></el-icon>
            <span>添加图片</span>
          </div>
        </el-upload>
      </div>
    </div>

    <!-- Section 4: 商品描述 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">4</span>商品描述
      </div>
      <Toolbar v-if="editorReady" style="border-bottom:1px solid #ccc" :editor="editorRef" :defaultConfig="toolbarConfig" />
      <Editor v-if="editorReady" style="min-height:300px;overflow-y:hidden;" v-model="form.description" :defaultConfig="editorConfig" @onCreated="handleCreated" />
    </div>

    <!-- Footer -->
    <div class="form-footer">
      <el-button @click="$router.push('/products')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">
        {{ isEdit ? '保存修改' : '创建商品' }}
      </el-button>
    </div>

    <!-- 图片预览弹窗 -->
    <el-dialog v-model="carouselPreviewVisible" title="图片预览" width="600px" top="5vh" destroy-on-close>
      <img v-if="carouselPreviewUrl" :src="carouselPreviewUrl" style="width:100%;border-radius:8px;" />
    </el-dialog>

    <!-- Loading -->
    <div v-if="loading" v-loading="loading" class="loading-overlay" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listCategoryTree } from '../../api/category'
import { useProductForm } from '../../composables/useProductForm'
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import type { IDomEditor } from '@wangeditor/editor'

/** 富文本编辑器 */
const editorRef = shallowRef<IDomEditor>()
const toolbarConfig = {}
const editorConfig = { placeholder: '请输入商品描述...' }
const handleCreated = (editor: IDomEditor) => { editorRef.value = editor }
const editorReady = ref(false) // 数据加载完成后才渲染编辑器，确保内容正确回显

/** 批量操作：SKU 矩阵的批量价格/库存 */
const bulkPrice = ref(0)
const bulkStock = ref(0)

/** SKU 行选中 */
const skuSelection = ref<string[]>([])
const isIndeterminate = computed(() => skuSelection.value.length > 0 && skuSelection.value.length < skus.value.length)
const isAllSelected = computed(() => skus.value.length > 0 && skuSelection.value.length === skus.value.length)

function toggleAllSelection(checked: boolean) {
  skuSelection.value = checked ? skus.value.map(s => s.id) : []
}

function toggleSkuSelection(id: string, checked: boolean) {
  if (checked) {
    skuSelection.value.push(id)
  } else {
    skuSelection.value = skuSelection.value.filter(sid => sid !== id)
  }
}

/** 图片预览 */
const carouselPreviewVisible = ref(false)
const carouselPreviewUrl = ref('')

function previewCarousel(url: string) {
  carouselPreviewUrl.value = url
  carouselPreviewVisible.value = true
}

/** 批量操作确认弹窗 */
async function confirmBulk(label: string, count: number): Promise<boolean> {
  if (count === 0) return false
  try {
    await ElMessageBox.confirm(
      `确定将【${label}】应用到选中的 ${count} 项 SKU 吗？`,
      '确认批量操作',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    return true
  } catch {
    return false
  }
}

async function applyBulkPrice() {
  if (!await confirmBulk('价格', skus.value.length)) return
  skus.value.forEach((s: { price: number }) => { s.price = bulkPrice.value })
}

async function applyBulkStock() {
  if (!await confirmBulk('库存', skus.value.length)) return
  skus.value.forEach((s: { stock: number }) => { s.stock = bulkStock.value })
}

async function applyBulkPriceToSelected() {
  if (!await confirmBulk('价格', skuSelection.value.length)) return
  skus.value.forEach((s: { id: string; price: number }) => {
    if (skuSelection.value.includes(s.id)) s.price = bulkPrice.value
  })
}

async function applyBulkStockToSelected() {
  if (!await confirmBulk('库存', skuSelection.value.length)) return
  skus.value.forEach((s: { id: string; stock: number }) => {
    if (skuSelection.value.includes(s.id)) s.stock = bulkStock.value
  })
}

async function applyBulkPriceToFiltered() {
  if (!await confirmBulk('价格', filteredSkuIds.value.length)) return
  const ids = filteredSkuIds.value
  skus.value.forEach((s: { id: string; price: number }) => {
    if (ids.includes(s.id)) s.price = bulkPrice.value
  })
}

async function applyBulkStockToFiltered() {
  if (!await confirmBulk('库存', filteredSkuIds.value.length)) return
  const ids = filteredSkuIds.value
  skus.value.forEach((s: { id: string; stock: number }) => {
    if (ids.includes(s.id)) s.stock = bulkStock.value
  })
}

/** 批量图片设置 */
const bulkSkuImage = ref('')

/** 规格筛选状态 */
const specFilter = ref<Record<string, string[]>>({})

/** 有名称和值的规格组（用于筛选栏渲染） */
const activeSpecs = computed(() => specGroups.value.filter(g => g.name && g.values.length > 0))

/** 当前是否有活跃的筛选条件 */
const hasActiveFilter = computed(() => Object.values(specFilter.value).some(v => v.length > 0))

/** 筛选匹配的 SKU ID 列表（同规格 AND，规格值 OR） */
const filteredSkuIds = computed(() => {
  const entries = Object.entries(specFilter.value).filter(([, vals]) => vals.length > 0)
  if (entries.length === 0) return []
  return skus.value
    .filter(sku => entries.every(([specName, selectedVals]) => selectedVals.includes(sku.spec[specName])))
    .map(s => s.id)
})

/** 筛选变化时自动更新复选框，取消筛选值后对应行的勾选也一并清除 */
watch(filteredSkuIds, (ids) => {
  skuSelection.value = [...ids]
})

function isSpecValueSelected(specName: string, value: string): boolean {
  return specFilter.value[specName]?.includes(value) ?? false
}

function toggleSpecFilter(specName: string, value: string) {
  if (!specFilter.value[specName]) specFilter.value[specName] = []
  const idx = specFilter.value[specName].indexOf(value)
  if (idx >= 0) {
    specFilter.value[specName].splice(idx, 1)
  } else {
    specFilter.value[specName].push(value)
  }
}

function clearSpecFilter() {
  specFilter.value = {}
}

async function handleBulkSkuImageUpload(uploadFile: { raw?: File }) {
  if (!uploadFile.raw) return
  try {
    const { getUploadToken } = await import('../../api/upload')
    const { uploadUrl, publicUrl } = await getUploadToken(uploadFile.raw.name, 'product')
    await fetch(uploadUrl, { method: 'PUT', body: uploadFile.raw })
    bulkSkuImage.value = publicUrl
  } catch (e) {
    console.error('批量上传 SKU 图片失败', e)
  }
}

async function applyBulkSkuImage(targetSkus: typeof skus.value, label: string, count: number) {
  if (!await confirmBulk('图片', count)) return
  targetSkus.forEach(s => { s.image = bulkSkuImage.value })
  ElMessage.success(`已设置 ${targetSkus.length} 项 SKU 图片`)
  bulkSkuImage.value = '' // 清空，允许继续上传下一张
}

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)
const cascaderOptions = ref<any[]>([])
const cascaderReady = ref(false)

const {
  form, specGroups, skus, loading, saving,
  addSpecGroup, removeSpecGroup, setSpecValues,
  loadProduct, save,
} = useProductForm()

async function handleSave() {
  const ok = await save(isEdit.value ? Number(route.params.id) : undefined)
  if (ok) router.push('/products')
}

/** 图片文件选择处理 - 直传 OSS */
async function handleFileChange(uploadFile: { raw?: File }) {
  if (!uploadFile.raw) return
  const file = uploadFile.raw
  try {
    const { getUploadToken } = await import('../../api/upload')
    const { uploadUrl, objectKey, publicUrl } = await getUploadToken(file.name, 'product')
    // 直传到 OSS
    await fetch(uploadUrl, { method: 'PUT', body: file })
    // 预览用 publicUrl（完整 URL），提交用 objectKey（后端补全 bucket 前缀）
    form.image = publicUrl
    form._imageKey = objectKey
  } catch (e) {
    console.error('上传图片失败', e)
  }
}

/** 轮播图上传 - 直传 OSS */
async function handleCarouselUpload(uploadFile: { raw?: File }) {
  if (!uploadFile.raw) return
  const file = uploadFile.raw
  try {
    const { getUploadToken } = await import('../../api/upload')
    const { uploadUrl, publicUrl } = await getUploadToken(file.name, 'product')
    await fetch(uploadUrl, { method: 'PUT', body: file })
    form.images = [...(form.images || []), publicUrl]
  } catch (e) {
    console.error('上传轮播图失败', e)
  }
}

/** 删除轮播图 */
function removeCarouselImage(index: number) {
  form.images = form.images.filter((_: string, i: number) => i !== index)
}

/** SKU 图片上传 - 直传 OSS */
async function uploadSkuImage(uploadFile: { raw?: File }, index: number) {
  if (!uploadFile.raw) return
  const file = uploadFile.raw
  try {
    const { getUploadToken } = await import('../../api/upload')
    const { uploadUrl, publicUrl } = await getUploadToken(file.name, 'product')
    await fetch(uploadUrl, { method: 'PUT', body: file })
    skus.value[index].image = publicUrl
  } catch (e) {
    console.error('上传 SKU 图片失败', e)
  }
}

onMounted(async () => {
  // 加载分类树
  try {
    const res = await listCategoryTree()
    cascaderOptions.value = [{ id: '0', name: '根目录', children: res.items }]
  } catch {
    cascaderOptions.value = []
  }
  // 编辑模式加载数据
  if (isEdit.value) {
    const productId = Number(route.params.id)
    if (!Number.isFinite(productId)) {
      ElMessage.error('无效的商品ID')
      router.push('/products')
      return
    }
    await loadProduct(productId)
  }
  // 等选项和值都就绪后再渲染 cascader 和编辑器，确保数据回显正确
  cascaderReady.value = true
  editorReady.value = true
})
</script>

<style scoped>
.product-form { padding: 24px; position: relative; }
.form-breadcrumb { margin-bottom: 24px; color: #666; font-size: 14px; }
.form-breadcrumb a { color: #409eff; text-decoration: none; }
.sep { margin: 0 8px; color: #ccc; }
.form-card { background: #fff; border-radius: 8px; padding: 24px; margin-bottom: 16px; box-shadow: 0 1px 4px rgba(0,0,0,0.06); }
.form-section-title { font-size: 16px; font-weight: 600; margin-bottom: 20px; display: flex; align-items: center; gap: 8px; }
.section-number { display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; border-radius: 50%; background: #409eff; color: #fff; font-size: 13px; font-weight: 600; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.full-width { grid-column: 1 / -1; }
.spec-builder { display: flex; flex-direction: column; gap: 12px; }
.spec-group-row { display: flex; align-items: center; gap: 8px; }
.sku-matrix { margin-top: 16px; overflow-x: auto; }
.sku-title { font-size: 14px; font-weight: 600; margin-bottom: 8px; color: #333; }
.sku-bulk-bar { display: flex; flex-direction: column; gap: 6px; margin-bottom: 8px; padding: 8px 12px; background: #fafafa; border-radius: 6px; font-size: 13px; color: #666; }
.sku-bulk-row { display: flex; align-items: center; gap: 4px; flex-wrap: wrap; }
.bulk-label { font-weight: 600; color: #333; white-space: nowrap; }
.sku-filter-bar { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; padding: 6px 12px; background: #fafafa; border-radius: 6px; font-size: 13px; flex-wrap: wrap; }
.filter-bar-label { color: #666; white-space: nowrap; }
.filter-group-label { margin-left: 4px; color: #999; font-size: 12px; white-space: nowrap; }
.filter-group-label::after { content: '：'; }
.filter-chip { display: inline-flex; align-items: center; padding: 2px 10px; border: 1px solid #d9d9d9; border-radius: 12px; background: #fff; color: #666; font-size: 12px; cursor: pointer; transition: all .2s; outline: none; white-space: nowrap; }
.filter-chip:hover { border-color: #409eff; color: #409eff; }
.filter-chip.active { border-color: #409eff; background: #ecf5ff; color: #409eff; }
.filter-count { color: #409eff; font-size: 12px; margin-left: 4px; white-space: nowrap; }
.sku-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.sku-table th, .sku-table td { border: 1px solid #e8e8e8; padding: 8px 12px; text-align: center; }
.sku-table th { background: #fafafa; font-weight: 600; white-space: nowrap; }
.sku-table .sku-row-selected { background: #ecf5ff; }
.sku-table .sku-row-dimmed { opacity: 0.35; }
.desc-editor { width: 100%; min-height: 300px; border: 1px solid #d9d9d9; border-radius: 6px; padding: 12px; font-size: 14px; line-height: 1.6; resize: vertical; outline: none; }
.desc-editor:focus { border-color: #409eff; }
.form-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 0; }
.upload-area { display: flex; gap: 8px; align-items: center; width: 100%; }
.main-image-preview { margin-top: 12px; }
.preview-img { max-width: 300px; max-height: 300px; border-radius: 8px; border: 1px solid #e8e8e8; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.image-placeholder { width: 300px; height: 200px; display: flex; align-items: center; justify-content: center; background: #fafafa; border: 2px dashed #d9d9d9; border-radius: 8px; color: #bbb; font-size: 14px; margin-top: 12px; }
.loading-overlay { position: absolute; inset: 0; background: rgba(255,255,255,0.6); z-index: 10; border-radius: 8px; }

/* 轮播图画廊 */
.gallery-grid { display: flex; flex-wrap: wrap; gap: 12px; }
.gallery-item { position: relative; width: 100px; height: 100px; border-radius: 8px; overflow: hidden; border: 1px solid #e8e8e8; cursor: pointer; }
.gallery-item:hover { border-color: #409eff; }
.gallery-thumb { width: 100%; height: 100%; object-fit: cover; }
.gallery-remove { position: absolute; top: 2px; right: 2px; width: 20px; height: 20px; padding: 0; min-width: 20px; font-size: 12px; }
.gallery-add { width: 100px; height: 100px; display: flex; flex-direction: column; align-items: center; justify-content: center; border: 2px dashed #d9d9d9; border-radius: 8px; color: #999; font-size: 12px; cursor: pointer; transition: all 0.2s; }
.gallery-add:hover { border-color: #409eff; color: #409eff; }
.gallery-add .el-icon { font-size: 24px; margin-bottom: 4px; }
</style>
