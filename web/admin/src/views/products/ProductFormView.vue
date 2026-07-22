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
        <el-form-item label="缩略图" class="full-width">
          <div class="upload-area">
            <el-input v-model="form.image" placeholder="输入图片URL或点击上传" style="flex:1" />
            <el-upload :show-file-list="false" action="" :auto-upload="false" :on-change="handleFileChange">
              <el-button>选择图片</el-button>
            </el-upload>
          </div>
          <img v-if="form.image" :src="form.image" class="preview-img" />
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
              <td>
                <el-input-number v-model="sku.price" :min="0" :precision="2" size="small" controls-position="right" style="width:130px" />
              </td>
              <td>
                <el-input-number v-model="sku.stock" :min="0" size="small" controls-position="right" style="width:130px" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Section 3: 商品描述 -->
    <div class="form-card">
      <div class="form-section-title">
        <span class="section-number">3</span>商品描述
      </div>
      <textarea v-model="form.description" class="desc-editor" placeholder="请输入商品描述..." />
    </div>

    <!-- Footer -->
    <div class="form-footer">
      <el-button @click="$router.push('/products')">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">
        {{ isEdit ? '保存修改' : '创建商品' }}
      </el-button>
    </div>

    <!-- Loading -->
    <div v-if="loading" v-loading="loading" class="loading-overlay" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Plus, Delete } from '@element-plus/icons-vue'
import { listCategoryTree } from '../../api/category'
import { useProductForm } from '../../composables/useProductForm'

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
    await loadProduct(Number(route.params.id))
  }
  // 等选项和值都就绪后再渲染 cascader，确保匹配生效
  cascaderReady.value = true
})
</script>

<style scoped>
.product-form { padding: 24px; max-width: 960px; margin: 0 auto; position: relative; }
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
.sku-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.sku-table th, .sku-table td { border: 1px solid #e8e8e8; padding: 8px 12px; text-align: center; }
.sku-table th { background: #fafafa; font-weight: 600; white-space: nowrap; }
.desc-editor { width: 100%; min-height: 200px; border: 1px solid #d9d9d9; border-radius: 6px; padding: 12px; font-size: 14px; line-height: 1.6; resize: vertical; outline: none; }
.desc-editor:focus { border-color: #409eff; }
.form-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 0; }
.upload-area { display: flex; gap: 8px; align-items: center; width: 100%; }
.preview-img { max-width: 200px; max-height: 200px; margin-top: 8px; border-radius: 4px; border: 1px solid #e8e8e8; }
.loading-overlay { position: absolute; inset: 0; background: rgba(255,255,255,0.6); z-index: 10; border-radius: 8px; }
</style>
