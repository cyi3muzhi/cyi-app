<template>
  <!-- 商品分类管理页面 - 统计卡片 + 分类表格 + 新增/编辑表单 -->
  <div class="categories-page">
    <!-- 页面标题与操作按钮 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">商品分类</h2>
        <span class="page-subtitle">管理商品分类层级与展示排序</span>
      </div>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加分类
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in categoryStats" :key="stat.label">
        <div class="stat-header">
          <div>
            <div class="stat-label">{{ stat.label }}</div>
            <div class="stat-value">{{ stat.value }}</div>
          </div>
          <div class="stat-icon" :class="stat.color">
            <el-icon :size="20"><component :is="stat.icon" /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/编辑分类表单卡片（可折叠） -->
    <div class="form-card" v-if="formVisible">
      <div class="form-card-header">
        <div class="form-card-title">
          <el-icon><component :is="isEditing ? 'Edit' : 'Plus'" /></el-icon>
          {{ isEditing ? '编辑分类' : '新增分类' }}
        </div>
        <div class="form-card-actions">
          <el-button text size="small" @click="handleResetForm">取消</el-button>
          <el-button type="primary" size="small" @click="handleSaveCategory">
            {{ isEditing ? '保存修改' : '确认添加' }}
          </el-button>
        </div>
      </div>
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
        size="small"
        class="category-form"
      >
        <div class="form-grid">
          <div class="form-main">
            <!-- 分类名称 -->
            <el-form-item label="分类名称" prop="name">
              <el-input v-model="formData.name" placeholder="请输入分类名称" maxlength="20" show-word-limit />
            </el-form-item>

            <!-- 父级分类 -->
            <el-form-item label="父级分类">
              <el-select v-model="formData.parentId" placeholder="顶级分类（不选）" clearable style="width: 100%">
                <el-option label="顶级分类（不选）" :value="null" />
                <el-option
                  v-for="cat in topLevelCategories"
                  :key="cat.id"
                  :label="cat.icon + ' ' + cat.name"
                  :value="cat.id"
                  :disabled="cat.id === editingId"
                />
              </el-select>
            </el-form-item>

            <!-- 排序 -->
            <el-form-item label="排序">
              <el-input-number
                v-model="formData.sort"
                :min="0"
                :max="999"
                controls-position="right"
                style="width: 160px"
              />
              <span class="form-help-text">数字越小越靠前</span>
            </el-form-item>

            <!-- 描述 -->
            <el-form-item label="描述">
              <el-input
                v-model="formData.description"
                type="textarea"
                :rows="3"
                placeholder="请输入分类描述"
                maxlength="200"
                show-word-limit
              />
            </el-form-item>

            <!-- 状态 -->
            <el-form-item label="状态">
              <el-switch
                v-model="formData.status"
                active-value="active"
                inactive-value="inactive"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>
          </div>

          <div class="form-side">
            <!-- 图标选择 -->
            <el-form-item label="分类图标" prop="icon">
              <div class="emoji-picker">
                <div
                  v-for="emoji in emojiList"
                  :key="emoji"
                  class="emoji-item"
                  :class="{ active: formData.icon === emoji }"
                  @click="formData.icon = emoji"
                >
                  {{ emoji }}
                </div>
              </div>
            </el-form-item>

            <!-- 图片上传 -->
            <el-form-item label="分类图片">
              <div class="upload-area" @click="triggerUpload">
                <img v-if="formData.image" :src="formData.image" class="upload-preview" />
                <div v-else class="upload-placeholder">
                  <el-icon :size="24"><Picture /></el-icon>
                  <span>点击上传图片</span>
                </div>
              </div>
              <input
                ref="fileInputRef"
                type="file"
                accept="image/*"
                style="display: none"
                @change="handleFileChange"
              />
              <span class="form-help-text">建议尺寸 200x200px</span>
            </el-form-item>
          </div>
        </div>
      </el-form>
    </div>

    <!-- 分类表格 -->
    <el-card shadow="never" class="table-card">
      <el-table :data="categoryList" stripe style="width: 100%" v-loading="loading" row-key="id">
        <el-table-column label="分类名称" min-width="220">
          <template #default="{ row }">
            <div class="category-name-cell">
              <span class="category-icon">{{ row.icon }}</span>
              <div class="category-info">
                <span class="category-name">{{ row.name }}</span>
                <span v-if="row.parentId" class="category-badge">子分类</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="商品数量" width="100" align="center">
          <template #default="{ row }">
            <span class="product-count">{{ row.productCount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="排序" width="120" align="center">
          <template #default="{ row }">
            <div class="sort-control">
              <el-input-number
                v-model="row.sort"
                :min="0"
                :max="999"
                size="small"
                controls-position="right"
                style="width: 110px"
                @change="handleSortChange(row)"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              active-value="active"
              inactive-value="inactive"
              inline-prompt
              active-icon="Check"
              inactive-icon="Close"
              @change="(val) => handleStatusChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button text size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/**
 * 商品分类管理页面
 * 支持分类的增删改查、排序调整、状态切换、图片与图标设置
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Folder,
  Collection,
  Grid,
  Picture,
  Check,
  Close
} from '@element-plus/icons-vue'

// ============================================================
// Mock 数据定义
// ============================================================

/** 预置 emoji 图标列表 */
const emojiList = [
  '📱', '💻', '⌚', '📷', '🎧', '🖥️', '⌨️', '🖱️',
  '👗', '👔', '👟', '👠', '👒', '🧢', '🧥', '👖',
  '🏠', '🛋️', '🛏️', '🚿', '🧹', '💡', '🔧', '🪴',
  '🍎', '🍔', '🥤', '🧃', '🍵', '☕', '🍰', '🍪',
  '⚽', '🏀', '🏃', '🚴', '🏊', '🎾', '⛷️', '🧘',
  '💄', '💅', '🧴', '💊', '🪥', '🪒', '💎', '🎁'
]

/** 模拟分类数据 */
const categoriesMock = reactive([
  { id: 1, name: '电子产品', icon: '📱', description: '手机、电脑、数码配件等电子产品', sort: 1, status: 'active', parentId: null, image: '', productCount: 128, createdAt: '2026-01-15' },
  { id: 2, name: '智能手机', icon: '📱', description: '各品牌智能手机及配件', sort: 1, status: 'active', parentId: 1, image: '', productCount: 56, createdAt: '2026-01-15' },
  { id: 3, name: '笔记本电脑', icon: '💻', description: '轻薄本、游戏本、办公本', sort: 2, status: 'active', parentId: 1, image: '', productCount: 38, createdAt: '2026-02-10' },
  { id: 4, name: '智能穿戴', icon: '⌚', description: '手表、手环、智能眼镜', sort: 3, status: 'inactive', parentId: 1, image: '', productCount: 12, createdAt: '2026-03-05' },
  { id: 5, name: '时尚服饰', icon: '👗', description: '男女装、鞋帽箱包', sort: 2, status: 'active', parentId: null, image: '', productCount: 256, createdAt: '2026-01-20' },
  { id: 6, name: '女装', icon: '👗', description: '连衣裙、上衣、裤装', sort: 1, status: 'active', parentId: 5, image: '', productCount: 98, createdAt: '2026-01-20' },
  { id: 7, name: '男装', icon: '👔', description: '衬衫、T恤、外套', sort: 2, status: 'active', parentId: 5, image: '', productCount: 76, createdAt: '2026-02-01' },
  { id: 8, name: '鞋靴', icon: '👟', description: '运动鞋、休闲鞋、靴子', sort: 3, status: 'active', parentId: 5, image: '', productCount: 52, createdAt: '2026-02-15' },
  { id: 9, name: '家居生活', icon: '🏠', description: '家具、家纺、日用百货', sort: 3, status: 'active', parentId: null, image: '', productCount: 186, createdAt: '2026-01-25' },
  { id: 10, name: '厨房用品', icon: '🍳', description: '锅具、餐具、厨房小电器', sort: 1, status: 'active', parentId: 9, image: '', productCount: 45, createdAt: '2026-01-25' },
  { id: 11, name: '家纺布艺', icon: '🛏️', description: '床品、窗帘、毛巾', sort: 2, status: 'active', parentId: 9, image: '', productCount: 33, createdAt: '2026-03-10' },
  { id: 12, name: '运动户外', icon: '⚽', description: '运动器材、户外装备', sort: 4, status: 'active', parentId: null, image: '', productCount: 94, createdAt: '2026-02-20' },
  { id: 13, name: '美妆护肤', icon: '💄', description: '彩妆、护肤品、个人护理', sort: 5, status: 'active', parentId: null, image: '', productCount: 142, createdAt: '2026-01-30' },
  { id: 14, name: '面部护理', icon: '🧴', description: '面霜、精华、面膜', sort: 1, status: 'active', parentId: 13, image: '', productCount: 68, createdAt: '2026-01-30' },
  { id: 15, name: '彩妆', icon: '💄', description: '口红、眼影、粉底', sort: 2, status: 'inactive', parentId: 13, image: '', productCount: 41, createdAt: '2026-04-05' },
  { id: 16, name: '食品饮料', icon: '🍎', description: '休闲零食、饮料冲调', sort: 6, status: 'active', parentId: null, image: '', productCount: 210, createdAt: '2026-03-15' }
])

// ============================================================
// 响应式状态
// ============================================================

/** 表单引用 */
const formRef = ref(null)

/** 文件输入引用 */
const fileInputRef = ref(null)

/** 加载状态 */
const loading = ref(false)

/** 表单是否可见 */
const formVisible = ref(false)

/** 是否处于编辑模式 */
const isEditing = ref(false)

/** 正在编辑的分类 ID */
const editingId = ref(null)

/** 表单数据 */
const formData = reactive({
  id: null,
  name: '',
  parentId: null,
  sort: 0,
  icon: '',
  description: '',
  status: 'active',
  image: '',
  productCount: 0
})

// ============================================================
// 计算属性
// ============================================================

/** 分类统计数据 */
const categoryStats = computed(() => [
  {
    label: '全部分类',
    value: categoriesMock.length,
    icon: Collection,
    color: 'blue'
  },
  {
    label: '顶级分类',
    value: categoriesMock.filter(c => !c.parentId).length,
    icon: Folder,
    color: 'purple'
  },
  {
    label: '子分类',
    value: categoriesMock.filter(c => c.parentId).length,
    icon: Folder,
    color: 'green'
  },
  {
    label: '已启用',
    value: categoriesMock.filter(c => c.status === 'active').length,
    icon: Folder,
    color: 'orange'
  }
])

/** 顶级分类列表（供父级选择器使用） */
const topLevelCategories = computed(() =>
  categoriesMock.filter(c => !c.parentId)
)

/** 展示分类列表（顶级分类带子分类缩进） */
const categoryList = computed(() => {
  const tops = categoriesMock.filter(c => !c.parentId).sort((a, b) => a.sort - b.sort)
  const result = []
  tops.forEach(top => {
    result.push(top)
    const children = categoriesMock
      .filter(c => c.parentId === top.id)
      .sort((a, b) => a.sort - b.sort)
    children.forEach(child => {
      result.push(child)
    })
  })
  return result
})

// ============================================================
// 表单校验规则
// ============================================================

const formRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 1, max: 20, message: '长度在 1 到 20 个字符', trigger: 'blur' }
  ],
  icon: [
    { required: true, message: '请选择一个图标', trigger: 'change' }
  ]
}

// ============================================================
// 操作方法
// ============================================================

/**
 * 打开新增分类表单
 */
function handleAdd() {
  isEditing.value = false
  editingId.value = null
  resetFormData()
  formVisible.value = true
}

/**
 * 编辑分类 - 填充表单
 * @param {Object} row 分类行数据
 */
function handleEdit(row) {
  isEditing.value = true
  editingId.value = row.id
  Object.assign(formData, {
    id: row.id,
    name: row.name,
    parentId: row.parentId,
    sort: row.sort,
    icon: row.icon,
    description: row.description,
    status: row.status,
    image: row.image,
    productCount: row.productCount
  })
  formVisible.value = true
  // 滚动到表单位置
  setTimeout(() => {
    const el = document.querySelector('.form-card')
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, 100)
}

/**
 * 保存分类（新增或更新）
 */
async function handleSaveCategory() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    ElMessage.warning('请完善表单信息')
    return
  }

  if (isEditing.value) {
    // 更新已有分类
    const idx = categoriesMock.findIndex(c => c.id === editingId.value)
    if (idx !== -1) {
      Object.assign(categoriesMock[idx], { ...formData })
    }
    ElMessage.success('分类已更新')
  } else {
    // 新增分类
    const newId = Math.max(...categoriesMock.map(c => c.id), 0) + 1
    categoriesMock.push({
      id: newId,
      name: formData.name,
      icon: formData.icon,
      description: formData.description,
      sort: formData.sort,
      status: formData.status,
      parentId: formData.parentId,
      image: formData.image,
      productCount: 0,
      createdAt: new Date().toISOString().slice(0, 10)
    })
    ElMessage.success('分类已添加')
  }

  handleResetForm()
}

/**
 * 重置表单并关闭表单卡片
 */
function handleResetForm() {
  formVisible.value = false
  isEditing.value = false
  editingId.value = null
  resetFormData()
  formRef.value?.resetFields()
}

/**
 * 重置表单数据到默认值
 */
function resetFormData() {
  formData.id = null
  formData.name = ''
  formData.parentId = null
  formData.sort = 0
  formData.icon = ''
  formData.description = ''
  formData.status = 'active'
  formData.image = ''
  formData.productCount = 0
}

/**
 * 删除分类 - 确认后删除
 * @param {Object} row 分类行数据
 */
function handleDelete(row) {
  // 检查是否有子分类
  const hasChildren = categoriesMock.some(c => c.parentId === row.id)
  if (hasChildren) {
    ElMessage.warning(`分类「${row.name}」下有子分类，请先删除子分类`)
    return
  }

  ElMessageBox.confirm(
    `确定要删除分类「${row.icon} ${row.name}」吗？此操作不可撤销。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error', draggable: true }
  ).then(() => {
    const idx = categoriesMock.findIndex(c => c.id === row.id)
    if (idx !== -1) {
      categoriesMock.splice(idx, 1)
    }
    ElMessage.success(`分类「${row.name}」已删除`)
  }).catch(() => {
    // 用户取消操作
  })
}

/**
 * 切换分类启用/禁用状态
 * @param {Object} row 分类行数据
 * @param {string} val 新的状态值
 */
function handleStatusChange(row, val) {
  const statusText = val === 'active' ? '启用' : '禁用'
  ElMessage.success(`分类「${row.name}」已${statusText}`)
}

/**
 * 排序值变更处理
 * @param {Object} row 分类行数据
 */
function handleSortChange(row) {
  ElMessage.success(`分类「${row.name}」排序已更新`)
}

/**
 * 触发文件选择
 */
function triggerUpload() {
  fileInputRef.value?.click()
}

/**
 * 处理文件选择 - 生成预览 URL
 * @param {Event} event 文件选择事件
 */
function handleFileChange(event) {
  const file = event.target.files?.[0]
  if (!file) return

  // 校验文件类型
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }

  // 校验文件大小（最大 2MB）
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('图片大小不能超过 2MB')
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    formData.image = e.target!.result as string
  }
  reader.readAsDataURL(file)

  // 重置 input 以便重复选择同一文件
  event.target.value = ''
}

// ============================================================
// 生命周期
// ============================================================

onMounted(() => {
  loading.value = true
  // 模拟数据加载延迟
  setTimeout(() => {
    loading.value = false
  }, 300)
})
</script>

<style scoped>
.categories-page {
  width: 100%;
}

/* ---- 页面标题 ---- */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0;
  line-height: 1.4;
}

.page-subtitle {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ---- 统计卡片 ---- */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 1.25rem;
}

.stat-card {
  background: var(--shop-card);
  border-radius: var(--shop-radius-lg);
  padding: 1.25rem;
  box-shadow: var(--shop-shadow-sm);
  border: 1px solid var(--shop-border);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

.stat-card:hover {
  box-shadow: var(--shop-shadow-md);
  transform: translateY(-2px);
}

.stat-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.stat-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
}

.stat-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.blue { background: var(--shop-primary-bg); color: var(--shop-primary); }
.stat-icon.green { background: #f0fdf4; color: var(--shop-success); }
.stat-icon.orange { background: #fef3c7; color: #d97706; }
.stat-icon.purple { background: #e0e7ff; color: #4338ca; }

/* ---- 表单卡片 ---- */
.form-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  padding: 1.25rem;
  margin-bottom: 1.25rem;
  box-shadow: var(--shop-shadow-sm);
  animation: slideDown 0.25s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.form-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--shop-border);
}

.form-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.form-card-actions {
  display: flex;
  gap: 8px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
}

.form-help-text {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-left: 8px;
}

/* ---- Emoji 选择器 ---- */
.emoji-picker {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  padding: 8px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-md);
  background: var(--shop-muted);
  max-height: 180px;
  overflow-y: auto;
}

.emoji-item {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--shop-radius-sm);
  cursor: pointer;
  font-size: 1.125rem;
  transition: background 0.15s ease, transform 0.15s ease;
  user-select: none;
}

.emoji-item:hover {
  background: #e5e7eb;
  transform: scale(1.15);
}

.emoji-item.active {
  background: var(--shop-primary-bg);
  box-shadow: 0 0 0 2px var(--shop-primary);
  transform: scale(1.15);
}

/* ---- 图片上传区域 ---- */
.upload-area {
  width: 100%;
  aspect-ratio: 1 / 1;
  max-width: 200px;
  border: 2px dashed var(--shop-border);
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.upload-area:hover {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.upload-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
}

/* ---- 表格卡片 ---- */
.table-card {
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
}

.table-card :deep(.el-card__body) {
  padding: 0;
}

/* ---- 表格单元格样式 ---- */
.category-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.category-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--shop-radius-md);
  background: var(--shop-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  flex-shrink: 0;
}

.category-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.category-name {
  font-weight: 500;
  color: var(--shop-foreground);
}

.category-badge {
  font-size: 0.6875rem;
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
  padding: 0 6px;
  border-radius: 4px;
  display: inline-block;
  width: fit-content;
}

.product-count {
  font-weight: 600;
  color: var(--shop-foreground);
}

.sort-control {
  display: flex;
  justify-content: center;
}

/* ---- 响应式 ---- */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .emoji-picker {
    grid-template-columns: repeat(6, 1fr);
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

</style>
