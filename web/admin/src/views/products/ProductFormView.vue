<template>
  <!-- 添加/编辑商品表单 - 重设计 -->
  <div class="pf">
    <!-- 全屏图片预览 -->
    <div v-if="previewImage" class="pf-preview" @click="previewImage = ''">
      <div class="pf-preview-bg" />
      <img :src="previewImage" class="pf-preview-img" />
      <button class="pf-preview-close" @click="previewImage = ''">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
    </div>

    <div class="pf-layout">
      <!-- ===== 左侧步骤导航 ===== -->
      <aside class="pf-nav">
        <div class="pf-nav-inner">
          <div class="pf-nav-back">
            <router-link to="/products" class="pf-nav-back-link">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
              返回列表
            </router-link>
          </div>
          <div class="pf-nav-steps">
            <div v-for="(s, i) in steps" :key="i" class="pf-step" :class="{ 'pf-step--done': scrolledSection > i, 'pf-step--active': scrolledSection === i }" @click="scrollToSection(i)">
              <div class="pf-step-dot">
                <svg v-if="scrolledSection > i" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
                <span v-else>{{ s.num }}</span>
              </div>
              <div class="pf-step-label">{{ s.label }}</div>
            </div>
          </div>
        </div>
      </aside>

      <!-- ===== 右侧内容区 ===== -->
      <main class="pf-content">
        <!-- Section 1: 基本信息 -->
        <section :ref="setSectionRef(0)" class="pf-card" data-section="0">
          <div class="pf-card-accent" />
          <div class="pf-card-head">
            <div class="pf-card-badge">01</div>
            <div>
              <div class="pf-card-title">基本信息</div>
              <div class="pf-card-desc">商品名称、分类与主图</div>
            </div>
          </div>
          <el-form label-position="top" class="pf-grid">
            <el-form-item label="商品名称" required class="pf-full">
              <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="120" />
            </el-form-item>
            <el-form-item label="商品分类" required>
              <el-cascader v-if="cascaderReady" v-model="form.category_id" :options="cascaderOptions"
                :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true, emitPath: false }"
                placeholder="请选择分类" clearable filterable />
            </el-form-item>
            <el-form-item label="商品主图" class="pf-full">
              <div class="pf-imgbox">
                <div v-if="form.image" class="pf-imgbox-preview">
                  <img :src="form.image" @click="previewImage = form.image" />
                  <button class="pf-imgbox-remove" @click="form.image = ''">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
                <el-upload v-else :show-file-list="false" action="" :auto-upload="false" :on-change="handleFileChange" class="pf-imgbox-upload">
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                  <span>点击上传主图</span>
                </el-upload>
              </div>
            </el-form-item>
          </el-form>
        </section>

        <!-- Section 2: 规格与 SKU -->
        <section :ref="setSectionRef(1)" class="pf-card" data-section="1">
          <div class="pf-card-accent pf-accent-2" />
          <div class="pf-card-head">
            <div class="pf-card-badge pf-badge-2">02</div>
            <div>
              <div class="pf-card-title">规格与 SKU</div>
              <div class="pf-card-desc">设置商品规格、库存与价格</div>
            </div>
          </div>
          <div class="pf-specs">
            <div v-for="(group, i) in specGroups" :key="group.id" class="pf-spec-row">
              <el-input v-model="group.name" placeholder="规格名称（如：颜色）" style="width:140px" />
              <el-select v-model="group.values" multiple filterable allow-create
                default-first-option placeholder="输入规格值 回车确认" style="flex:1"
                @change="(vals: any) => setSpecValues(i, vals)" />
              <el-button :icon="Delete" circle size="small" type="danger" @click="removeSpecGroup(i)" />
            </div>
            <el-button type="primary" plain @click="addSpecGroup" class="pf-spec-add">
              <el-icon><Plus /></el-icon> 添加规格
            </el-button>
          </div>

          <div v-if="skus.length > 0" class="pf-sku">
            <div class="pf-sku-head">
              <span class="pf-sku-title">SKU 明细 <span class="pf-sku-count">{{ skus.length }} 项</span></span>
            </div>

            <!-- 批量工具栏 -->
            <div class="pf-sku-tools">
              <div class="pf-sku-tool-row">
                <span class="pf-tool-label">价格</span>
                <el-input-number v-model="bulkPrice" :min="0" :precision="2" size="small" style="width:110px" />
                <el-button size="small" @click="applyBulkPrice">全部</el-button>
                <el-button size="small" :disabled="skuSelection.length === 0" @click="applyBulkPriceToSelected">选中</el-button>
                <el-button size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkPriceToFiltered">筛选({{ filteredSkuIds.length }})</el-button>
              </div>
              <div class="pf-sku-tool-row">
                <span class="pf-tool-label">库存</span>
                <el-input-number v-model="bulkStock" :min="0" size="small" style="width:110px" />
                <el-button size="small" @click="applyBulkStock">全部</el-button>
                <el-button size="small" :disabled="skuSelection.length === 0" @click="applyBulkStockToSelected">选中</el-button>
                <el-button size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkStockToFiltered">筛选({{ filteredSkuIds.length }})</el-button>
              </div>
              <div class="pf-sku-tool-row">
                <span class="pf-tool-label">图片</span>
                <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleBulkSkuImageUpload">
                  <el-button v-if="!bulkSkuImage" size="small">选择</el-button>
                </el-upload>
                <img v-if="bulkSkuImage" :src="bulkSkuImage" class="pf-bulk-img" />
                <el-button v-if="bulkSkuImage" size="small" @click="applyBulkSkuImage(skus, '全部', skus.length)">全部</el-button>
                <el-button v-if="bulkSkuImage" size="small" :disabled="skuSelection.length === 0" @click="applyBulkSkuImage(skus.filter(s => skuSelection.includes(s.id)), '选中', skuSelection.length)">选中</el-button>
                <el-button v-if="bulkSkuImage" size="small" type="primary" :disabled="filteredSkuIds.length === 0" @click="applyBulkSkuImage(skus.filter(s => filteredSkuIds.includes(s.id)), '筛选', filteredSkuIds.length)">筛选</el-button>
                <el-button v-if="bulkSkuImage" size="small" text type="info" @click="bulkSkuImage = ''">取消</el-button>
              </div>
            </div>

            <!-- 规格筛选 -->
            <div v-if="activeSpecs.length > 0" class="pf-sku-filter">
              <span class="pf-filter-label">筛选</span>
              <template v-for="group in activeSpecs" :key="group.name">
                <span class="pf-filter-group">{{ group.name }}</span>
                <button v-for="val in group.values" :key="val"
                  :class="['pf-filter-chip', { active: isSpecValueSelected(group.name, val) }]"
                  @click="toggleSpecFilter(group.name, val)">{{ val }}</button>
              </template>
              <span v-if="filteredSkuIds.length > 0" class="pf-filter-count">匹配 {{ filteredSkuIds.length }} 项</span>
              <el-button v-if="hasActiveFilter" size="small" text @click="clearSpecFilter">清除</el-button>
              <span class="pf-filter-sep" />
              <span class="pf-filter-sel">已选 {{ skuSelection.length }} 项</span>
              <el-button size="small" text :disabled="skuSelection.length === 0" @click="skuSelection = []">清除</el-button>
            </div>

            <!-- SKU 表格 -->
            <div class="pf-sku-table-wrap">
              <table class="pf-sku-table">
                <thead>
                  <tr>
                    <th style="width:40px">
                      <el-checkbox :indeterminate="isIndeterminate" :model-value="isAllSelected" @change="toggleAllSelection" />
                    </th>
                    <th v-for="key of Object.keys(skus[0].spec)" :key="key">{{ key }}</th>
                    <th>价格</th>
                    <th>库存</th>
                    <th>图片</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(sku, i) in skus" :key="sku.id"
                    :class="{ 'pf-sku-row-sel': skuSelection.includes(sku.id), 'pf-sku-row-dim': hasActiveFilter && !filteredSkuIds.includes(sku.id) }">
                    <td><el-checkbox :model-value="skuSelection.includes(sku.id)" @change="(val: boolean) => toggleSkuSelection(sku.id, val)" /></td>
                    <td v-for="v of Object.values(sku.spec)" :key="v" class="pf-sku-spec-val">{{ v }}</td>
                    <td><el-input-number v-model="sku.price" :min="0" :precision="2" size="small" controls-position="right" style="width:120px" /></td>
                    <td><el-input-number v-model="sku.stock" :min="0" size="small" controls-position="right" style="width:120px" /></td>
                    <td>
                      <div class="pf-sku-img-cell">
                        <img v-if="sku.image" :src="sku.image" class="pf-sku-img" @click="previewImage = sku.image" />
                        <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="(f: any) => uploadSkuImage(f, i)">
                          <el-button size="small">{{ sku.image ? '更换' : '上传' }}</el-button>
                        </el-upload>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <!-- Section 3: 轮播图 -->
        <section :ref="setSectionRef(2)" class="pf-card" data-section="2">
          <div class="pf-card-accent pf-accent-3" />
          <div class="pf-card-head">
            <div class="pf-card-badge pf-badge-3">03</div>
            <div>
              <div class="pf-card-title">商品轮播图</div>
              <div class="pf-card-desc">展示商品多角度图片</div>
            </div>
          </div>
          <div class="pf-gallery">
            <div v-for="(img, i) in form.images" :key="i" class="pf-gallery-item">
              <img :src="img" @click="previewImage = img" />
              <button class="pf-gallery-remove" @click="removeCarouselImage(i)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleCarouselUpload">
              <div class="pf-gallery-add">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                <span>添加图片</span>
              </div>
            </el-upload>
          </div>
        </section>

        <!-- Section 4: 商品描述 -->
        <section :ref="setSectionRef(3)" class="pf-card" data-section="3">
          <div class="pf-card-accent pf-accent-4" />
          <div class="pf-card-head">
            <div class="pf-card-badge pf-badge-4">04</div>
            <div>
              <div class="pf-card-title">商品描述</div>
              <div class="pf-card-desc">详细描述商品信息与卖点</div>
            </div>
          </div>
          <Toolbar v-if="editorReady" class="pf-editor-tb" :editor="editorRef" :defaultConfig="toolbarConfig" />
          <Editor v-if="editorReady" class="pf-editor-body" v-model="form.description" :defaultConfig="editorConfig" @onCreated="handleCreated" />
        </section>

        <!-- 底部操作栏 -->
        <div class="pf-actions">
          <div class="pf-actions-inner">
            <el-button @click="$router.push('/products')" class="pf-btn pf-btn--outline">取消</el-button>
            <el-button type="primary" :loading="saving" class="pf-btn pf-btn--primary" @click="handleSave">
              {{ isEdit ? '保存修改' : '创建商品' }}
            </el-button>
          </div>
        </div>
      </main>
    </div>

    <div v-if="loading" v-loading="loading" class="pf-loading" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, shallowRef, onBeforeUnmount } from 'vue'
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

/** 步骤导航定义 */
const steps = [
  { num: '1', label: '基本信息' },
  { num: '2', label: '规格与 SKU' },
  { num: '3', label: '轮播图' },
  { num: '4', label: '描述' }
]

/** 每个 section 的 ref */
const sectionRefs = ref<(HTMLElement | null)[]>([null, null, null, null])

/** 设置 section 的 template ref */
function setSectionRef(i: number) {
  return (el: any) => { sectionRefs.value[i] = el as HTMLElement | null }
}

/** 当前滚动到的 section 索引 */
const scrolledSection = ref(0)

/** 滚动监听：更新当前 section */
let scrollTimer: ReturnType<typeof setTimeout>
function onScroll() {
  clearTimeout(scrollTimer)
  scrollTimer = setTimeout(() => {
    const items = sectionRefs.value.filter(Boolean) as HTMLElement[]
    if (items.length === 0) return
    // 顶部偏移（顶部栏高度 + 间距）
    const offset = 100
    const scrollTop = window.scrollY + offset
    let idx = 0
    for (let i = 0; i < items.length; i++) {
      if (items[i].offsetTop <= scrollTop) idx = i
    }
    scrolledSection.value = idx
  }, 60)
}

function scrollToSection(index: number) {
  const el = sectionRefs.value[index]
  if (el) {
    const top = el.getBoundingClientRect().top + window.scrollY - 80
    window.scrollTo({ top, behavior: 'smooth' })
  }
}

onMounted(() => {
  // 延迟初始化，等 DOM 渲染完成
  setTimeout(() => onScroll(), 100)
  window.addEventListener('scroll', onScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})

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

/** 图片预览弹窗 */
const previewImage = ref('')

function closePreview() {
  previewImage.value = ''
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
/* ============================================
   商品表单 - 现代编辑页设计
   ============================================ */

/* ---------- 基础 ---------- */
.pf {
  width: 100%;
  position: relative;
  --pf-primary: var(--shop-primary, #409eff);
  --pf-bg: #f5f6fa;
  --pf-card: #ffffff;
  --pf-border: #e8ecf1;
  --pf-text: #1a2332;
  --pf-muted: #8892a8;
  --pf-accent-1: #409eff;
  --pf-accent-2: #f59e0b;
  --pf-accent-3: #10b981;
  --pf-accent-4: #8b5cf6;
}

/* ---------- 布局 ---------- */
.pf-layout {
  display: flex;
  gap: 0;
  min-height: calc(100vh - 64px);
}

/* ---------- 左侧导航 ---------- */
.pf-nav {
  width: 200px;
  flex-shrink: 0;
  position: relative;
}

.pf-nav-inner {
  position: sticky;
  top: 80px;
  padding: 24px 16px;
}

.pf-nav-back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
  color: var(--pf-muted);
  text-decoration: none;
  transition: color 0.15s;
}
.pf-nav-back-link:hover {
  color: var(--pf-primary);
}

.pf-nav-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--pf-text);
  margin: 20px 0 24px;
  line-height: 1.3;
}

/* ---------- 步骤 ---------- */
.pf-nav-steps {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pf-step {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}
.pf-step:hover {
  background: rgba(0,0,0,0.03);
}
.pf-step--active {
  background: rgba(64,158,255,0.08);
}

.pf-step-dot {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  flex-shrink: 0;
  background: var(--pf-border);
  color: var(--pf-muted);
  transition: all 0.3s;
}
.pf-step--active .pf-step-dot {
  background: var(--pf-primary);
  color: #fff;
  box-shadow: 0 2px 8px rgba(64,158,255,0.3);
}
.pf-step--done .pf-step-dot {
  background: #10b981;
  color: #fff;
}

.pf-step-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--pf-muted);
  transition: color 0.2s;
}
.pf-step--active .pf-step-label {
  color: var(--pf-text);
  font-weight: 600;
}
.pf-step--done .pf-step-label {
  color: #10b981;
}

/* ---------- 主内容区 ---------- */
.pf-content {
  flex: 1;
  min-width: 0;
  padding: 24px 40px 120px;
}

/* ---------- 卡片 ---------- */
.pf-card {
  background: var(--pf-card);
  border: 1px solid var(--pf-border);
  border-radius: 14px;
  padding: 28px 28px 32px;
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
  animation: pfFadeUp 0.5s ease both;
  transition: box-shadow 0.25s;
}
.pf-card:nth-child(1) { animation-delay: 0.05s; }
.pf-card:nth-child(2) { animation-delay: 0.12s; }
.pf-card:nth-child(3) { animation-delay: 0.19s; }
.pf-card:nth-child(4) { animation-delay: 0.26s; }
.pf-card:hover {
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
}

@keyframes pfFadeUp {
  from { opacity: 0; transform: translateY(24px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 左侧色带 */
.pf-card-accent {
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: var(--pf-accent-1);
  border-radius: 0 2px 2px 0;
}
.pf-accent-2 { background: var(--pf-accent-2); }
.pf-accent-3 { background: var(--pf-accent-3); }
.pf-accent-4 { background: var(--pf-accent-4); }

/* 卡片头部 */
.pf-card-head {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
  padding-left: 4px;
}

.pf-card-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  font-weight: 700;
  background: rgba(64,158,255,0.1);
  color: var(--pf-accent-1);
  flex-shrink: 0;
}
.pf-badge-2 { background: rgba(245,158,11,0.1); color: var(--pf-accent-2); }
.pf-badge-3 { background: rgba(16,185,129,0.1); color: var(--pf-accent-3); }
.pf-badge-4 { background: rgba(139,92,246,0.1); color: var(--pf-accent-4); }

.pf-card-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--pf-text);
}

.pf-card-desc {
  font-size: 0.75rem;
  color: var(--pf-muted);
  margin-top: 2px;
}

/* ---------- 表单网格 ---------- */
.pf-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px 32px;
  padding-left: 4px;
  max-width: 900px;
}
.pf-full { grid-column: 1 / -1; }

/* ---------- 图片上传 ---------- */
.pf-imgbox {
  display: flex;
  gap: 16px;
}

.pf-imgbox-upload {
  width: 180px;
  height: 180px;
  border: 2px dashed var(--pf-border);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--pf-muted);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.25s;
}
.pf-imgbox-upload:hover {
  border-color: var(--pf-primary);
  color: var(--pf-primary);
  background: rgba(64,158,255,0.04);
}

.pf-imgbox-preview {
  position: relative;
  width: 180px;
  height: 180px;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--pf-border);
  cursor: pointer;
}
.pf-imgbox-preview:hover {
  border-color: var(--pf-primary);
}
.pf-imgbox-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.pf-imgbox-remove {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: none;
  background: rgba(0,0,0,0.5);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.pf-imgbox-preview:hover .pf-imgbox-remove {
  opacity: 1;
}

/* ---------- 规格 ---------- */
.pf-specs {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-left: 4px;
}
.pf-spec-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pf-spec-add {
  align-self: flex-start;
}

/* ---------- SKU ---------- */
.pf-sku {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--pf-border);
  padding-left: 4px;
}

.pf-sku-head {
  margin-bottom: 14px;
}
.pf-sku-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--pf-text);
}
.pf-sku-count {
  font-weight: 400;
  color: var(--pf-muted);
  font-size: 0.75rem;
  margin-left: 6px;
}

/* 批量工具栏 */
.pf-sku-tools {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
  padding: 12px 16px;
  background: #f8f9fc;
  border-radius: 10px;
  font-size: 0.8125rem;
}
.pf-sku-tool-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.pf-tool-label {
  font-weight: 600;
  color: var(--pf-text);
  white-space: nowrap;
  min-width: 32px;
}
.pf-bulk-img {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  object-fit: cover;
}

/* 规格筛选 */
.pf-sku-filter {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
  padding: 8px 12px;
  background: #f8f9fc;
  border-radius: 10px;
  font-size: 0.8125rem;
  flex-wrap: wrap;
}
.pf-filter-label {
  font-weight: 600;
  color: var(--pf-text);
  white-space: nowrap;
}
.pf-filter-group {
  color: var(--pf-muted);
  font-size: 0.75rem;
}
.pf-filter-group::after {
  content: '：';
}
.pf-filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border: 1px solid var(--pf-border);
  border-radius: 999px;
  background: #fff;
  color: var(--pf-muted);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
  outline: none;
}
.pf-filter-chip:hover {
  border-color: var(--pf-primary);
  color: var(--pf-primary);
}
.pf-filter-chip.active {
  border-color: var(--pf-primary);
  background: rgba(64,158,255,0.08);
  color: var(--pf-primary);
  font-weight: 600;
}
.pf-filter-count {
  color: var(--pf-primary);
  font-size: 0.75rem;
}
.pf-filter-sep {
  width: 1px;
  height: 14px;
  background: var(--pf-border);
  margin: 0 4px;
}
.pf-filter-sel {
  color: var(--pf-muted);
  font-size: 0.75rem;
}

/* SKU 表格 */
.pf-sku-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--pf-border);
  border-radius: 10px;
}
.pf-sku-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}
.pf-sku-table th,
.pf-sku-table td {
  padding: 10px 12px;
  text-align: center;
  border-bottom: 1px solid var(--pf-border);
}
.pf-sku-table thead th {
  background: #f8f9fc;
  font-weight: 600;
  color: var(--pf-text);
  white-space: nowrap;
  position: sticky;
  top: 0;
  z-index: 2;
}
.pf-sku-table tbody tr:last-child td {
  border-bottom: none;
}
.pf-sku-table tbody tr:hover {
  background: rgba(64,158,255,0.03);
}
.pf-sku-row-sel {
  background: rgba(64,158,255,0.06) !important;
}
.pf-sku-row-dim {
  opacity: 0.3;
}
.pf-sku-spec-val {
  font-weight: 500;
  color: var(--pf-text);
}

.pf-sku-img-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  justify-content: center;
}
.pf-sku-img {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  object-fit: cover;
  cursor: pointer;
  transition: transform 0.15s;
}
.pf-sku-img:hover {
  transform: scale(1.2);
}

/* ---------- 轮播图 ---------- */
.pf-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding-left: 4px;
}
.pf-gallery-item {
  position: relative;
  width: 100px;
  height: 100px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--pf-border);
  cursor: pointer;
  transition: border-color 0.2s;
}
.pf-gallery-item:hover {
  border-color: var(--pf-primary);
}
.pf-gallery-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pf-gallery-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  border: none;
  background: rgba(0,0,0,0.45);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.pf-gallery-item:hover .pf-gallery-remove {
  opacity: 1;
}
.pf-gallery-add {
  width: 100px;
  height: 100px;
  border-radius: 10px;
  border: 2px dashed var(--pf-border);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: var(--pf-muted);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}
.pf-gallery-add:hover {
  border-color: var(--pf-primary);
  color: var(--pf-primary);
  background: rgba(64,158,255,0.04);
}

/* ---------- 编辑器 ---------- */
.pf-editor-tb {
  border: 1px solid var(--pf-border);
  border-bottom: none;
  border-radius: 10px 10px 0 0;
  padding-left: 4px;
}
.pf-editor-body {
  min-height: 320px;
  border: 1px solid var(--pf-border);
  border-radius: 0 0 10px 10px;
  padding-left: 4px;
}

/* ---------- 底部操作栏 ---------- */
.pf-actions {
  position: fixed;
  bottom: 0;
  left: 200px;
  right: 0;
  background: rgba(255,255,255,0.92);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-top: 1px solid var(--pf-border);
  z-index: 50;
  padding: 0 32px;
  animation: pfFadeUp 0.4s ease both;
  animation-delay: 0.35s;
}
.pf-actions-inner {
  max-width: 100%;
  padding: 14px 40px;
  margin: 0 auto;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.pf-btn {
  min-width: 100px;
  border-radius: 10px !important;
  font-weight: 600;
  height: 40px;
}
.pf-btn--outline {
  border: 1px solid var(--pf-border);
}
.pf-btn--primary {
  box-shadow: 0 4px 12px rgba(64,158,255,0.25);
}

/* ---------- 图片预览 ---------- */
.pf-preview {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pfOverlayIn 0.2s ease;
}
.pf-preview-bg {
  position: absolute;
  inset: 0;
  background: rgba(0,0,0,0.75);
  backdrop-filter: blur(4px);
}
.pf-preview-img {
  position: relative;
  max-width: 80vw;
  max-height: 80vh;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.4);
  animation: pfScaleIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.pf-preview-close {
  position: absolute;
  top: 24px;
  right: 24px;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: none;
  background: rgba(255,255,255,0.15);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
}
.pf-preview-close:hover {
  background: rgba(255,255,255,0.25);
}

@keyframes pfOverlayIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes pfScaleIn {
  from { opacity: 0; transform: scale(0.92); }
  to { opacity: 1; transform: scale(1); }
}

/* ---------- Loading ---------- */
.pf-loading {
  position: absolute;
  inset: 0;
  background: rgba(255,255,255,0.6);
  z-index: 10;
  border-radius: 14px;
}

/* ---------- 响应式 ---------- */
@media (max-width: 900px) {
  .pf-nav {
    display: none;
  }
  .pf-content {
    padding: 16px 16px 100px;
  }
  .pf-actions {
    left: 0;
    padding: 0 16px;
  }
  .pf-actions-inner {
    max-width: 100%;
  }
}
</style>
