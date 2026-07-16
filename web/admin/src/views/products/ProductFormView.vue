<template>
  <!-- 添加/编辑商品表单页面 -->
  <div class="product-form">
    <!-- 面包屑导航 -->
    <div class="form-breadcrumb">
      <a @click.prevent="$router.push('/products')" href="/products">商品管理</a>
      <span class="breadcrumb-sep">&gt;</span>
      <span class="breadcrumb-current">{{ isEdit ? '编辑商品' : '添加商品' }}</span>
    </div>

    <!-- Section 1: 基本信息 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">1</span>基本信息
      </div>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="form-grid"
      >
        <el-form-item prop="name" class="full-width form-item-custom">
          <template #label>
            <label class="form-label">商品名称<span class="required">*</span></label>
          </template>
          <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="120" class="custom-input" />
        </el-form-item>

        <el-form-item class="full-width form-item-custom">
          <template #label>
            <label class="form-label">商品简介</label>
          </template>
          <el-input
            v-model="form.summary"
            type="textarea"
            :rows="3"
            placeholder="简短描述商品特点"
            maxlength="200"
            class="custom-textarea"
          />
          <div class="form-hint">建议 50-100 字，突出商品核心卖点</div>
        </el-form-item>

        <el-form-item prop="category" class="form-item-custom">
          <template #label>
            <label class="form-label">商品分类<span class="required">*</span></label>
          </template>
          <el-cascader
            v-model="form.category"
            :options="cascaderOptions"
            :props="{ value: 'id', label: 'name', children: 'children', checkStrictly: true }"
            placeholder="请选择分类"
            class="custom-cascader"
            clearable
            filterable
          />
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">品牌<span class="optional">（选填）</span></label>
          </template>
          <el-input v-model="form.brand" placeholder="请输入品牌名称" class="custom-input" />
        </el-form-item>

        <el-form-item class="full-width form-item-custom">
          <template #label>
            <label class="form-label">商品标签</label>
          </template>
          <div class="tag-chips">
            <span
              v-for="tag in tagOptions"
              :key="tag.value"
              class="tag-chip"
              :class="{ selected: form.tags.includes(tag.value) }"
              @click="toggleTag(tag.value)"
            >
              {{ tag.label }}
            </span>
          </div>
          <div class="form-hint">点击选择标签，可多选</div>
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 2: 价格库存 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">2</span>价格库存
      </div>
      <el-form
        :model="form"
        :rules="rules"
        label-position="top"
        class="form-grid"
      >
        <el-form-item prop="price" class="form-item-custom">
          <template #label>
            <label class="form-label">销售价格<span class="required">*</span></label>
          </template>
          <div class="form-input-group">
            <span class="input-prefix">&yen;</span>
            <el-input
              v-model="form.price"
              type="number"
              placeholder="0.00"
              class="custom-input with-prefix"
              :min="0"
              step="0.01"
            />
          </div>
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">原价<span class="optional">（选填）</span></label>
          </template>
          <div class="form-input-group">
            <span class="input-prefix">&yen;</span>
            <el-input
              v-model="form.originalPrice"
              type="number"
              placeholder="0.00"
              class="custom-input with-prefix"
              :min="0"
              step="0.01"
            />
          </div>
        </el-form-item>

        <el-form-item prop="stock" class="form-item-custom">
          <template #label>
            <label class="form-label">库存数量<span class="required">*</span></label>
          </template>
          <el-input
            v-model="form.stock"
            type="number"
            placeholder="0"
            class="custom-input"
            :min="0"
            step="1"
          />
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">SKU 编码<span class="optional">（选填）</span></label>
          </template>
          <el-input v-model="form.sku" placeholder="自动生成或手动输入" class="custom-input" />
          <div class="form-hint">留空将自动生成唯一 SKU</div>
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">排序权重</label>
          </template>
          <el-input
            v-model="form.sortWeight"
            type="number"
            placeholder="0"
            class="custom-input"
            :min="0"
            step="1"
          />
          <div class="form-hint">数字越大，排序越靠前</div>
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">计量单位<span class="optional">（选填）</span></label>
          </template>
          <el-input v-model="form.unit" placeholder="例如：件、个、箱" class="custom-input" />
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 3: 商品图片 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">3</span>商品图片
      </div>
      <el-form :model="form" :rules="rules" label-position="top">
        <el-form-item prop="mainImage" class="form-item-custom">
          <template #label>
            <label class="form-label">商品主图<span class="required">*</span></label>
          </template>
          <div
            class="upload-area"
            :class="{ 'has-image': !!form.mainImage }"
            @click="triggerMainUpload"
          >
            <template v-if="!form.mainImage">
              <div class="upload-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
              </div>
              <div class="upload-text">点击上传商品主图</div>
              <div class="upload-hint">支持 JPG、PNG、WebP，建议尺寸 800x800 以上</div>
            </template>
            <template v-else>
              <div class="upload-preview">
                <img :src="form.mainImage" alt="商品主图" />
                <button class="remove-btn" type="button" @click.stop="removeMainImage">&times;</button>
              </div>
            </template>
          </div>
          <input
            ref="mainFileInput"
            type="file"
            accept="image/*"
            style="display:none"
            @change="handleMainImageUpload"
          />
          <div v-if="formErrors.mainImage" class="form-error-text show">请上传商品主图</div>
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">商品附图 <span class="optional">（选填，最多 3 张）</span></label>
          </template>
          <div class="upload-grid">
            <div
              v-for="(img, idx) in subImageSlots"
              :key="idx"
              class="upload-slot"
              :class="{ 'has-image': !!img }"
              @click="triggerSubUpload(idx)"
            >
              <template v-if="!img">
                <span class="slot-icon">+</span>
                <span class="slot-text">添加图片</span>
              </template>
              <img v-else :src="img" :alt="'附图' + (idx + 1)" />
            </div>
          </div>
          <input
            ref="subFileInputs"
            type="file"
            accept="image/*"
            style="display:none"
            @change="handleSubImageUpload"
          />
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 4: 规格管理 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">4</span>规格管理
      </div>
      <el-form :model="form" label-position="top">
        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">规格构建器</label>
          </template>
          <div v-for="(spec, idx) in form.specs" :key="idx" class="spec-row">
            <div class="spec-name-wrap">
              <el-input
                v-model="spec.name"
                placeholder="规格名 (如颜色)"
                class="custom-input"
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
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">SKU 矩阵</label>
          </template>
          <div v-if="validSpecGroups.length > 0 && form.skus.length > 0" class="sku-matrix-wrap">
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
                    <td><el-input v-model.number="sku.price" type="number" class="sku-price-input" step="0.01" min="0" /></td>
                    <td><el-input v-model.number="sku.stock" type="number" class="sku-stock-input" min="0" step="1" /></td>
                    <td><button class="sku-del-btn" @click="deleteSkuFromMatrix(sku.id)" title="删除此SKU">&times;</button></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div v-else-if="form.specs.length > 0" class="sku-empty-hint">
            请先添加规格组并填写规格值，系统将自动生成 SKU 矩阵
          </div>
          <div v-else class="sku-empty-hint">
            如需管理多规格商品，请先添加规格组
          </div>
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 5: 商品详情 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">5</span>商品详情
      </div>
      <el-form :model="form" label-position="top">
        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">富文本内容</label>
          </template>
          <!-- 富文本编辑器容器 -->
          <div class="editor-container">
            <!-- 富文本工具栏（装饰，实际使用 textarea） -->
            <div class="editor-toolbar">
              <button class="toolbar-btn" title="加粗"><strong>B</strong></button>
              <button class="toolbar-btn" title="斜体"><em>I</em></button>
              <button class="toolbar-btn" title="下划线"><u>U</u></button>
              <div class="toolbar-divider"></div>
              <button class="toolbar-btn" title="无序列表">&#8226;</button>
              <button class="toolbar-btn" title="有序列表">1.</button>
              <div class="toolbar-divider"></div>
              <button class="toolbar-btn" title="插入链接">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
              </button>
              <button class="toolbar-btn" title="插入图片">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
              </button>
              <div class="toolbar-divider"></div>
              <button class="toolbar-btn" title="预览">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              </button>
            </div>
            <!-- 文本输入区域 -->
            <div class="editor-textarea-wrapper">
              <textarea
                v-model="form.detail"
                rows="8"
                placeholder="请输入商品详情描述，支持 HTML 格式..."
                class="custom-textarea editor-textarea"
              ></textarea>
            </div>
          </div>
          <div class="form-hint">建议详细描述商品规格、材质、使用说明等信息</div>
        </el-form-item>
      </el-form>
    </div>

    <!-- Section 6: 上下架设置 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">6</span>上下架设置
      </div>
      <div class="form-grid">
        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">上架状态</label>
          </template>
          <div class="toggle-switch" :class="{ active: form.status }" @click="form.status = !form.status">
            <div class="toggle-track"><div class="toggle-thumb"></div></div>
            <span class="toggle-state">{{ form.status ? '上架中' : '已下架' }}</span>
          </div>
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">上架时间</label>
          </template>
          <el-date-picker
            v-model="form.publishDate"
            type="date"
            placeholder="选择上架日期"
            class="custom-date-picker"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>

        <el-form-item class="form-item-custom">
          <template #label>
            <label class="form-label">首页推荐</label>
          </template>
          <div class="toggle-switch" :class="{ active: form.recommend }" @click="form.recommend = !form.recommend">
            <div class="toggle-track"><div class="toggle-thumb"></div></div>
            <span class="toggle-state">{{ form.recommend ? '推荐' : '不推荐' }}</span>
          </div>
          <div class="form-hint">开启后将在首页推荐位展示</div>
        </el-form-item>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="form-card form-card-actions">
      <div class="form-actions">
        <el-button class="btn btn-outline" @click="handleCancel">取消</el-button>
        <el-button class="btn btn-secondary" :loading="submitting" @click="handleSaveDraft">保存草稿</el-button>
        <div class="action-spacer"></div>
        <el-button class="btn btn-primary" :loading="submitting" @click="handleSubmit">提交审核</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 添加/编辑商品表单页面
 * 支持添加和编辑两种模式，根据路由参数判断
 * 包含5个区块：基本信息、价格库存、商品图片、商品详情、上下架设置
 */
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { listCategoryTree } from '../../api/category'
import type { CategoryItem } from '../../api/category'
import { uploadToOSS } from '../../utils/upload'
import {
  getProduct as apiGetProduct,
  createProduct as apiCreateProduct,
  updateProduct as apiUpdateProduct
} from '../../api/product'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)

/** 判断是否为编辑模式 */
const isEdit = computed(() => route.name === 'ProductEdit')

/** 标签选项 */
const tagOptions = [
  { label: '新品', value: 'new' },
  { label: '热卖', value: 'hot' },
  { label: '推荐', value: 'recommend' },
  { label: '特惠', value: 'sale' }
]

/** 分类树数据 */
const categories = ref<CategoryItem[]>([])

/** 级联选择器选项（转换分类树结构） */
const cascaderOptions = computed(() => categories.value)

/**
 * 根据分类名称查找级联路径（用于编辑时回填）
 */
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
  return walk(categories.value, []) || []
}

/**
 * 在分类树中递归查找指定 ID 的节点名称
 */
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
  return walk(categories.value) || ''
}

/** 文件上传相关 ref */
const mainFileInput = ref<HTMLInputElement>()
const subFileInputs = ref<HTMLInputElement>()
/** 当前点击的子图索引 */
const currentSubIndex = ref(-1)

/** 子图插槽（最多3张） */
const subImageSlots = ref<string[]>([ '', '', '' ])

/** 表单错误信息（用于上传校验） */
const formErrors = reactive({
  mainImage: false
})

/**
 * 本地 SKU 接口（前端展示用，price 单位为元）
 */
interface LocalSkuItem {
  id: string
  combo: Record<string, string>
  price: number
  stock: number
  sales?: number
}

/**
 * 本地规格组接口（前端展示用）
 */
interface LocalSpecGroup {
  name: string
  values: string[]
}

/**
 * 表单数据模型
 */
const form = reactive({
  name: '',
  summary: '',
  category: [] as number[],
  brand: '',
  tags: [] as string[],
  price: null as number | null,
  originalPrice: null as number | null,
  stock: null as number | null,
  sku: '',
  sortWeight: 0,
  unit: '',
  mainImage: '',
  _imageObjectKey: '', // OSS object key，提交表单时使用
  detail: '',
  status: true,
  publishDate: '',
  recommend: false,
  specs: [] as LocalSpecGroup[],
  skus: [] as LocalSkuItem[]
})

/**
 * 表单验证规则
 */
const rules: FormRules = {
  name: [
    { required: true, message: '请输入商品名称', trigger: 'blur' },
    { max: 120, message: '商品名称不能超过120个字符', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择商品分类', trigger: 'change' }
  ],
  price: [
    { required: true, message: '请输入销售价格', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '销售价格必须大于0', trigger: 'blur' }
  ],
  stock: [
    { required: true, message: '请输入库存数量', trigger: 'blur' },
    { type: 'number', min: 0, message: '库存数量不能为负数', trigger: 'blur' }
  ],
  mainImage: [
    { required: true, message: '请上传商品主图', trigger: 'change' }
  ]
}

/**
 * 加载分类树数据
 */
async function loadCategories() {
  try {
    const res = await listCategoryTree()
    categories.value = res.items
  } catch {
    // 如果 API 调用失败，使用空列表
    categories.value = []
  }
}

/**
 * 切换标签选中状态
 * @param tag 标签值
 */
function toggleTag(tag: string) {
  const idx = form.tags.indexOf(tag)
  if (idx === -1) {
    form.tags.push(tag)
  } else {
    form.tags.splice(idx, 1)
  }
}

/** 规格输入框ref集合 */
const tagInputRefs = ref<Record<number, HTMLInputElement>>({})

/** 生成唯一 ID（用于本地 SKU） */
function uid(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
}

/** 有效的规格组（有名称且有值） */
const validSpecGroups = computed(() => {
  return form.specs.filter(s => s.name && s.values.length > 0)
})

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
  const newSkus: LocalSkuItem[] = []
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
      newSkus.push({ id: uid(), combo: comboMap, price: form.price || 0, stock: form.stock || 0 })
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

/**
 * 触发主图上传
 */
function triggerMainUpload() {
  if (!form.mainImage) {
    mainFileInput.value?.click()
  }
}

/**
 * 压缩图片到指定最大宽度，返回 Blob
 */
function compressImage(file: File, maxWidth = 800, quality = 0.7): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        let { width, height } = img
        if (width > maxWidth) {
          height = Math.round(height * maxWidth / width)
          width = maxWidth
        }
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')!
        ctx.drawImage(img, 0, 0, width, height)
        canvas.toBlob(blob => {
          if (blob) resolve(blob)
          else reject(new Error('压缩失败'))
        }, 'image/jpeg', quality)
      }
      img.onerror = reject
      img.src = e.target?.result as string
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

/**
 * 处理主图上传（压缩 → 直传 OSS → 返回 object_key）
 */
async function handleMainImageUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  try {
    const blob = await compressImage(file)
    const compressedFile = new File([blob], file.name.replace(/\.[^.]+$/, '.jpg'), { type: 'image/jpeg' })
    // 直传 OSS
    const result = await uploadToOSS(compressedFile, 'product')
    // publicUrl 用于预览，objectKey 用于提交表单
    form.mainImage = result.publicUrl
    form._imageObjectKey = result.objectKey
    formErrors.mainImage = false
  } catch (e) {
    console.error('上传失败:', e)
  }
  target.value = ''
}

/**
 * 移除主图
 */
function removeMainImage() {
  form.mainImage = ''
}

/**
 * 触发子图上传
 * @param index 子图索引
 */
function triggerSubUpload(index: number) {
  currentSubIndex.value = index
  subFileInputs.value?.click()
}

/**
 * 处理子图上传
 */
function handleSubImageUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file || currentSubIndex.value === -1) return

  const reader = new FileReader()
  reader.onload = (e) => {
    subImageSlots.value[currentSubIndex.value] = e.target?.result as string
  }
  reader.readAsDataURL(file)
  target.value = ''
  currentSubIndex.value = -1
}

/**
 * 加载商品数据（编辑模式）
 * 从 API 获取商品详情并回填表单
 */
async function loadProduct() {
  if (!isEdit.value) return
  const id = Number(route.params.id)
  try {
    const res = await apiGetProduct(id)
    const product = res.product
    form.name = product.name
    form.category = findCategoryPath(product.category)
    // API 返回的 price 单位是分，前端显示为元
    form.price = product.price / 100
    form.stock = product.stock === 0 ? null : product.stock
    form.sku = product.sku
    form.mainImage = product.image
    form._imageObjectKey = ''
    form.status = product.status === '上架中'
    // 加载 specs 和 skus（价格分转元）
    form.specs = (product.specs || []).map(s => ({ name: s.name, values: s.values }))
    form.skus = (product.skus || []).map(s => ({
      id: String(s.id),
      combo: s.spec || {},
      price: s.price / 100,
      stock: s.stock,
      sales: s.sales || 0
    }))
  } catch {
    ElMessage.error('加载商品数据失败')
    router.push('/products')
  }
}

/**
 * 表单整体校验
 */
async function validateForm(): Promise<boolean> {
  let valid = true

  // Element Plus 表单校验
  if (formRef.value) {
    try {
      await formRef.value.validate()
    } catch {
      valid = false
    }
  }

  // 主图校验
  if (!form.mainImage) {
    formErrors.mainImage = true
    valid = false
  } else {
    formErrors.mainImage = false
  }

  return valid
}

/**
 * 取消：返回商品列表
 */
function handleCancel() {
  router.push('/products')
}

/**
 * 保存草稿
 * 调用 API 保存商品基本信息
 */
async function handleSaveDraft() {
  submitting.value = true
  try {
    const categoryName = form.category.length > 0
      ? findCategoryNameById(form.category[form.category.length - 1])
      : ''
    // specs：直接映射
    const specs = form.specs.filter(s => s.name && s.values.length > 0)
    // skus：价格从元转为分
    const skus = form.skus.map(s => ({
      sku_code: s.id,
      price: Math.round(s.price * 100),
      stock: s.stock,
      sales: s.sales || 0,
      spec: s.combo
    }))
    const data = {
      name: form.name,
      category: categoryName,
      price: form.price || 0,
      stock: form.stock || 0,
      description: form.detail || undefined,
      image: form._imageObjectKey || form.mainImage || undefined,
      sku: form.sku || undefined,
      specs: specs.length > 0 ? specs : undefined,
      skus: skus.length > 0 ? skus : undefined
    }

    if (isEdit.value) {
      const id = Number(route.params.id)
      await apiUpdateProduct(id, { ...data, status: form.status ? '上架中' : '已下架' })
    } else {
      await apiCreateProduct(data)
    }
    ElMessage.success('商品已保存为草稿')
    router.push('/products')
  } catch {
    ElMessage.error('保存失败，请重试')
  } finally {
    submitting.value = false
  }
}

/**
 * 提交审核
 * 验证表单 -> 调用创建/更新 API -> 返回列表
 */
async function handleSubmit() {
  const valid = await validateForm()
  if (!valid) {
    ElMessage.warning('请检查表单中的必填字段')
    return
  }

  submitting.value = true
  try {
    const categoryName = form.category.length > 0
      ? findCategoryNameById(form.category[form.category.length - 1])
      : ''
    // specs：直接映射
    const specs = form.specs.filter(s => s.name && s.values.length > 0)
    // skus：价格从元转为分
    const skus = form.skus.map(s => ({
      sku_code: s.id,
      price: Math.round(s.price * 100),
      stock: s.stock,
      sales: s.sales || 0,
      spec: s.combo
    }))
    const data = {
      name: form.name,
      category: categoryName,
      price: form.price || 0,
      stock: form.stock || 0,
      description: form.detail || undefined,
      image: form._imageObjectKey || form.mainImage || undefined,
      sku: form.sku || undefined,
      specs: specs.length > 0 ? specs : undefined,
      skus: skus.length > 0 ? skus : undefined
    }

    if (isEdit.value) {
      const id = Number(route.params.id)
      await apiUpdateProduct(id, { ...data, status: form.status ? '上架中' : '已下架' })
      ElMessage.success('商品已更新')
    } else {
      await apiCreateProduct(data)
      ElMessage.success('商品已提交审核')
    }

    router.push('/products')
  } catch {
    ElMessage.error('操作失败，请重试')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadCategories()
  loadProduct()
  // 设置默认上架时间为今天
  const today = new Date()
  form.publishDate = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
})
</script>

<style scoped>
/* ========================================
   商品表单页面样式
   毛玻璃卡片风格，与项目设计 token 一致
   ======================================== */
.product-form {
  max-width: 100%;
}

/* ---- 面包屑导航 ---- */
.form-breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.form-breadcrumb a {
  color: var(--shop-primary);
  text-decoration: none;
  cursor: pointer;
}

.form-breadcrumb a:hover {
  text-decoration: underline;
}

.form-breadcrumb .breadcrumb-sep {
  color: #cbd5e1;
}

.form-breadcrumb .breadcrumb-current {
  color: var(--shop-foreground);
  font-weight: 500;
}

/* 防止表单内文字竖排显示 */
.product-form span,
.product-form label {
  word-break: keep-all;
}

/* ---- 毛玻璃卡片 ---- */
.form-card {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  box-shadow: var(--shop-shadow-sm);
  padding: 1.75rem 2rem;
  margin-bottom: 1.5rem;
  transition: box-shadow 0.2s ease;
}

.form-card:hover {
  box-shadow: var(--shop-shadow-md);
}

.form-card-actions {
  padding: 1rem 2rem;
}

/* ---- 区块标题 ---- */
.form-section-title {
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--shop-foreground);
  margin-bottom: 1.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--shop-border);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-number {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 999px;
  background: var(--shop-primary);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  flex-shrink: 0;
}

/* ---- 双栏网格布局 ---- */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.25rem 1.5rem;
}

.form-grid .full-width {
  grid-column: 1 / -1;
}

/* ---- 自定义表单项样式 ---- */
.form-item-custom {
  margin-bottom: 0.75rem;
}

/* 自定义 label */
.form-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--shop-foreground);
  margin-bottom: 0.375rem;
  line-height: 1.4;
  white-space: nowrap;
}

.form-label .required {
  color: var(--shop-destructive);
  margin-left: 0.125rem;
}

.form-label .optional {
  color: var(--shop-muted-foreground);
  font-weight: 400;
  font-size: 0.75rem;
}

/* ---- 自定义 Input 样式 ---- */
.custom-input {
  --el-input-border-color: var(--shop-border);
  --el-input-hover-border-color: var(--shop-primary);
  --el-input-focus-border-color: var(--shop-primary);
  --el-input-border-radius: var(--shop-radius-md);
  --el-input-font-size: 0.875rem;
  --el-input-height: 2.5rem;
}

.custom-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--shop-border) inset;
  border-radius: var(--shop-radius-md);
  padding: 0 0.875rem;
}

.custom-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset;
}

.custom-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset, 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.custom-input :deep(.el-input__inner) {
  height: 2.5rem;
  color: var(--shop-foreground);
}

.custom-input :deep(.el-input__inner::placeholder) {
  color: #94a3b8;
}

/* 带前缀的输入框 */
.form-input-group {
  position: relative;
}

.input-prefix {
  position: absolute;
  left: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  pointer-events: none;
  z-index: 1;
}

.custom-input.with-prefix :deep(.el-input__wrapper) {
  padding-left: 2rem;
}

/* ---- 自定义 Textarea 样式 ---- */
.custom-textarea {
  --el-input-border-color: var(--shop-border);
  --el-input-hover-border-color: var(--shop-primary);
  --el-input-focus-border-color: var(--shop-primary);
  --el-input-border-radius: var(--shop-radius-md);
  --el-input-font-size: 0.875rem;
}

.custom-textarea :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px var(--shop-border) inset;
  border-radius: var(--shop-radius-md);
  padding: 0.625rem 0.875rem;
  color: var(--shop-foreground);
  font-family: var(--shop-font-sans);
  line-height: 1.6;
  min-height: 80px;
}

.custom-textarea :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset;
}

.custom-textarea :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset, 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.custom-textarea :deep(.el-textarea__inner::placeholder) {
  color: #94a3b8;
}

.editor-textarea :deep(.el-textarea__inner) {
  min-height: 160px !important;
  line-height: 1.6;
}

/* ---- 自定义 Select / Cascader 样式 ---- */
.custom-select, .custom-cascader {
  width: 100%;
}
.custom-select {
  --el-select-border-color-hover: var(--shop-primary);
  --el-select-input-focus-border-color: var(--shop-primary);
}
.custom-cascader {
  --el-cascader-menu-border-color-hover: var(--shop-primary);
  --el-cascader-menu-selected-text-color: var(--shop-primary);
}

.custom-select :deep(.el-select__wrapper) {
  box-shadow: 0 0 0 1px var(--shop-border) inset;
  border-radius: var(--shop-radius-md);
  padding: 0 0.875rem;
  min-height: 2.5rem;
}

.custom-select :deep(.el-select__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset;
}

.custom-select :deep(.el-select__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset, 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.custom-select :deep(.el-select__placeholder) {
  color: #94a3b8;
}

/* ---- 自定义 Date Picker ---- */
.custom-date-picker {
  width: 100%;
  --el-date-editor-date-icon-color: var(--shop-muted-foreground);
}

.custom-date-picker :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--shop-border) inset;
  border-radius: var(--shop-radius-md);
  padding: 0 0.875rem;
  height: 2.5rem;
}

.custom-date-picker :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset;
}

.custom-date-picker :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--shop-primary) inset, 0 0 0 3px rgba(59, 130, 246, 0.12);
}

/* ---- 提示文字 ---- */
.form-hint {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 0.25rem;
  line-height: 1.4;
}

/* ---- 错误信息 ---- */
.form-error-text {
  font-size: 0.75rem;
  color: var(--shop-destructive);
  margin-top: 0.25rem;
  display: none;
}

.form-error-text.show {
  display: block;
}

/* ---- Tag Chips ---- */
.tag-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.875rem;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--shop-border);
  background: #fff;
  color: var(--shop-muted-foreground);
  transition: all 0.2s ease;
  user-select: none;
}

.tag-chip:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.tag-chip.selected {
  background: var(--shop-primary);
  color: #fff;
  border-color: var(--shop-primary);
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.25);
}

/* ---- Toggle Switch ---- */
.toggle-switch {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
}

.toggle-switch .toggle-track {
  position: relative;
  width: 2.75rem;
  height: 1.5rem;
  background: #cbd5e1;
  border-radius: 999px;
  transition: background 0.25s ease;
  flex-shrink: 0;
}

.toggle-switch .toggle-track .toggle-thumb {
  position: absolute;
  top: 0.1875rem;
  left: 0.1875rem;
  width: 1.125rem;
  height: 1.125rem;
  background: #fff;
  border-radius: 999px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  transition: transform 0.25s ease;
}

.toggle-switch.active .toggle-track {
  background: var(--shop-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.toggle-switch.active .toggle-track .toggle-thumb {
  transform: translateX(1.25rem);
}

.toggle-switch .toggle-state {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
  min-width: 2.5rem;
}

.toggle-switch.active .toggle-state {
  color: var(--shop-primary);
}

/* ---- Upload 上传区域 ---- */
.upload-area {
  border: 2px dashed var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.25s ease;
  background: #fafbfc;
}

.upload-area:hover {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.upload-area .upload-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 0.75rem;
  border-radius: 999px;
  background: var(--shop-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  color: var(--shop-muted-foreground);
}

.upload-area:hover .upload-icon {
  background: var(--shop-primary);
  color: #fff;
}

.upload-area .upload-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  font-weight: 500;
}

.upload-area .upload-hint {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 0.25rem;
}

.upload-area.has-image {
  border-style: solid;
  border-color: var(--shop-primary);
  padding: 0.5rem;
}

.upload-area.has-image .upload-preview {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: var(--shop-radius-md);
  background: var(--shop-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
}

.upload-area.has-image .upload-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-area.has-image .upload-preview .remove-btn {
  position: absolute;
  top: 0.375rem;
  right: 0.375rem;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  line-height: 1;
}

.upload-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-top: 0.5rem;
}

.upload-slot {
  aspect-ratio: 16 / 10;
  border: 2px dashed var(--shop-border);
  border-radius: var(--shop-radius-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.25s ease;
  background: #fafbfc;
  gap: 0.5rem;
}

.upload-slot:hover {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.upload-slot .slot-icon {
  font-size: 1.5rem;
  color: var(--shop-muted-foreground);
  line-height: 1;
}

.upload-slot .slot-text {
  font-size: 0.6875rem;
  color: #94a3b8;
}

.upload-slot.has-image {
  border-style: solid;
  border-color: var(--shop-border);
  padding: 0.25rem;
}

.upload-slot.has-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: calc(var(--shop-radius-lg) - 2px);
}

/* ---- Editor 编辑器容器 ---- */
.editor-container {
  display: flex;
  flex-direction: column;
  border-radius: var(--shop-radius-md);
  overflow: hidden;
}

/* ---- Editor 工具栏 ---- */
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--shop-border);
  border-bottom: none;
  background: var(--shop-muted);
  flex-wrap: wrap;
}

.editor-toolbar .toolbar-btn {
  width: 2rem;
  height: 2rem;
  border: none;
  background: transparent;
  border-radius: var(--shop-radius-sm);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
  font-weight: 600;
  font-family: var(--shop-font-sans);
}

.editor-toolbar .toolbar-btn:hover {
  background: #e2e8f0;
  color: var(--shop-foreground);
}

.editor-toolbar .toolbar-divider {
  width: 1px;
  height: 1.25rem;
  background: var(--shop-border);
  margin: 0 0.25rem;
}

/* ---- Editor 文本输入区域 ---- */
.editor-textarea-wrapper {
  border: 1px solid var(--shop-border);
  border-top: none;
}

.editor-textarea-wrapper .editor-textarea {
  width: 100%;
  min-height: 200px;
  padding: 0.625rem 0.875rem;
  border: none;
  outline: none;
  resize: vertical;
  font-size: 0.875rem;
  font-family: var(--shop-font-sans);
  line-height: 1.6;
  color: var(--shop-foreground);
  background: #fff;
}

.editor-textarea-wrapper .editor-textarea::placeholder {
  color: #94a3b8;
}

.editor-textarea-wrapper .editor-textarea:focus {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
}

/* ---- 规格构建器 ---- */
.spec-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px 14px;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-md);
  margin-bottom: 8px;
  border: 1px solid var(--shop-border);
}

.spec-row .spec-name-wrap {
  width: 130px;
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
  color: var(--shop-muted-foreground);
  cursor: pointer;
  border-radius: var(--shop-radius-sm);
  flex-shrink: 0;
  margin-top: 0.5rem;
  transition: all 0.15s ease;
  font-size: 1.125rem;
}

.spec-row .spec-remove-btn:hover {
  background: var(--shop-destructive-bg);
  color: var(--shop-destructive);
}

.tag-input-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
  padding: 4px 6px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-sm);
  background: #fff;
  min-height: 2.5rem;
  cursor: text;
}

.tag-input-wrap:focus-within {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.tag-input-wrap input {
  border: none;
  outline: none;
  font-size: 0.8125rem;
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
  background: transparent;
  flex: 1;
  min-width: 80px;
}

.tag-input-wrap input::placeholder {
  color: #94a3b8;
}

.tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

.tag-badge .tag-remove {
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.15s ease;
}

.tag-badge:hover .tag-remove {
  opacity: 1;
}

.add-spec-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border: 1px dashed var(--shop-border);
  border-radius: var(--shop-radius-md);
  font-size: 0.8125rem;
  font-weight: 500;
  font-family: var(--shop-font-sans);
  color: var(--shop-muted-foreground);
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  margin-top: 8px;
}

.add-spec-btn:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

/* ---- SKU 矩阵 ---- */
.sku-matrix-wrap {
  margin-top: 4px;
}

.sku-matrix-scroll {
  overflow-x: auto;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-md);
}

.sku-matrix {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  min-width: 100%;
}

.sku-matrix thead th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  font-size: 0.75rem;
  background: var(--shop-muted);
  border-bottom: 1px solid var(--shop-border);
  white-space: nowrap;
}

.sku-matrix tbody td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--shop-border);
  vertical-align: middle;
}

.sku-matrix tbody tr:last-child td {
  border-bottom: none;
}

.sku-matrix tbody tr:hover {
  background: var(--shop-primary-bg);
}

.spec-combo {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.combo-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

.sku-price-input,
.sku-stock-input {
  width: 100%;
  max-width: 100px;
  padding: 4px 8px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-sm);
  font-size: 0.8125rem;
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
  outline: none;
  transition: all 0.15s ease;
}

.sku-price-input:focus,
.sku-stock-input:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.sku-del-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  border-radius: var(--shop-radius-sm);
  font-size: 1rem;
  transition: all 0.15s ease;
}

.sku-del-btn:hover {
  background: var(--shop-destructive-bg);
  color: var(--shop-destructive);
}

.sku-empty-hint {
  padding: 1rem;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-md);
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  text-align: center;
}

/* ---- 操作按钮 ---- */
.form-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.action-spacer {
  flex: 1;
}

.btn {
  --el-button-size: 2.5rem;
  font-weight: 600;
  border-radius: var(--shop-radius-md);
  font-size: 0.875rem;
  font-family: var(--shop-font-sans);
  transition: all 0.2s ease;
}

.btn.btn-primary {
  --el-button-bg-color: var(--shop-primary);
  --el-button-border-color: var(--shop-primary);
  --el-button-hover-bg-color: var(--shop-primary-dark);
  --el-button-hover-border-color: var(--shop-primary-dark);
  --el-button-active-bg-color: var(--shop-primary-dark);
  --el-button-active-border-color: var(--shop-primary-dark);
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.25);
}

.btn.btn-primary:hover {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
  transform: translateY(-1px);
}

.btn.btn-secondary {
  --el-button-bg-color: #e2e8f0;
  --el-button-border-color: #e2e8f0;
  --el-button-text-color: var(--shop-foreground);
  --el-button-hover-bg-color: #cbd5e1;
  --el-button-hover-border-color: #cbd5e1;
  --el-button-hover-text-color: var(--shop-foreground);
  --el-button-active-bg-color: #cbd5e1;
  --el-button-active-border-color: #cbd5e1;
}

.btn.btn-outline {
  --el-button-bg-color: transparent;
  --el-button-border-color: var(--shop-border);
  --el-button-text-color: var(--shop-muted-foreground);
  --el-button-hover-bg-color: var(--shop-muted);
  --el-button-hover-border-color: #94a3b8;
  --el-button-hover-text-color: var(--shop-muted-foreground);
  --el-button-active-bg-color: var(--shop-muted);
  --el-button-active-border-color: #94a3b8;
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-grid .full-width {
    grid-column: 1;
  }
}

@media (max-width: 768px) {
  .form-card {
    padding: 1.25rem 1rem;
  }

  .upload-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .upload-grid {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-wrap: wrap;
  }

  .form-actions .action-spacer {
    display: none;
  }

  .form-actions .btn {
    flex: 1;
  }
}
</style>