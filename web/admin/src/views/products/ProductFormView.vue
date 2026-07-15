<template>
  <!-- 添加/编辑商品表单页面 -->
  <div class="product-form">
    <!-- 基本信息 -->
    <el-card class="form-card" shadow="never">
      <template #header>
        <div class="form-section-title">
          <span class="section-number">1</span>基本信息
        </div>
      </template>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px" label-position="top">
        <el-form-item label="商品名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入商品名称" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="商品分类" prop="category">
              <el-select v-model="form.category" placeholder="请选择分类" style="width: 100%">
                <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SKU 编码" prop="sku">
              <el-input v-model="form.sku" placeholder="自动生成或手动输入" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="销售价格" prop="price">
              <el-input-number
                v-model="form.price"
                :min="0"
                :precision="2"
                :step="1"
                style="width: 100%"
                placeholder="0.00"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="库存数量" prop="stock">
              <el-input-number
                v-model="form.stock"
                :min="0"
                :step="1"
                style="width: 100%"
                placeholder="0"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="商品描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入商品描述"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 商品图片 -->
    <el-card class="form-card" shadow="never">
      <template #header>
        <div class="form-section-title">
          <span class="section-number">2</span>商品图片
        </div>
      </template>
      <el-upload
        action="#"
        list-type="picture-card"
        :auto-upload="false"
        :limit="1"
        :on-change="handleUploadChange"
      >
        <el-icon><Plus /></el-icon>
        <div class="upload-tip">
          <span>上传商品主图</span>
        </div>
      </el-upload>
    </el-card>

    <!-- 上下架设置 -->
    <el-card class="form-card" shadow="never">
      <template #header>
        <div class="form-section-title">
          <span class="section-number">3</span>上下架设置
        </div>
      </template>
      <el-form-item label="上架状态">
        <el-switch
          v-model="form.status"
          active-value="上架中"
          inactive-value="已下架"
          active-text="上架中"
          inactive-text="已下架"
        />
      </el-form-item>
    </el-card>

    <!-- 提交按钮 -->
    <div class="form-actions">
      <el-button @click="$router.push('/products')">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ isEdit ? '保存修改' : '提交审核' }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 添加/编辑商品表单页面
 * 支持添加和编辑两种模式，根据路由参数判断
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { products, productCategories } from '../../api/mock'

const route = useRoute()
const router = useRouter()
const formRef = ref(null)
const submitting = ref(false)

/** 判断是否为编辑模式 */
const isEdit = computed(() => route.name === 'ProductEdit')

/** 商品分类列表 */
const categories = productCategories

/** 表单数据 */
const form = reactive({
  name: '',
  category: '',
  price: 0,
  stock: 0,
  sku: '',
  description: '',
  status: '上架中'
})

/** 表单验证规则 */
const rules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
  price: [{ required: true, message: '请输入销售价格', trigger: 'blur' }],
  stock: [{ required: true, message: '请输入库存数量', trigger: 'blur' }]
}

/** 上传的图片文件列表 */
const uploadFiles = ref([])

/**
 * 处理上传文件变更
 * @param {Object} file 上传的文件
 */
function handleUploadChange(file) {
  uploadFiles.value.push(file)
}

/**
 * 如果为编辑模式，加载商品数据
 */
function loadProduct() {
  if (!isEdit.value) return
  const id = Number(route.params.id)
  const product = products.find(p => p.id === id)
  if (product) {
    form.name = product.name
    form.category = product.category
    form.price = product.price
    form.stock = product.stock
    form.sku = product.sku
    form.status = product.status
  }
}

onMounted(() => {
  loadProduct()
})

/**
 * 提交表单
 * 验证 -> 保存或更新 mock 数据 -> 返回列表
 */
function handleSubmit() {
  formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      // 模拟接口请求延迟
      await new Promise(resolve => setTimeout(resolve, 500))

      if (isEdit.value) {
        // 编辑模式：更新已有商品
        const id = Number(route.params.id)
        const idx = products.findIndex(p => p.id === id)
        if (idx !== -1) {
          Object.assign(products[idx], {
            name: form.name,
            category: form.category,
            price: form.price,
            stock: form.stock,
            sku: form.sku,
            status: form.status
          })
        }
        ElMessage.success('商品已更新')
      } else {
        // 添加模式：新增商品
        const newId = products.length > 0 ? Math.max(...products.map(p => p.id)) + 1 : 1
        products.push({
          id: newId,
          name: form.name,
          category: form.category,
          price: form.price,
          stock: form.stock,
          sales: 0,
          status: form.status,
          img: '',
          sku: form.sku || `YX-${String(newId).padStart(3, '0')}`
        })
        ElMessage.success('商品添加成功')
      }

      router.push('/products')
    } catch (e) {
      ElMessage.error('操作失败，请重试')
    } finally {
      submitting.value = false
    }
  })
}
</script>

<style scoped>
.product-form {
  width: 100%;
}

.form-card {
  margin-bottom: 1.25rem;
}

.form-section-title {
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--shop-foreground);
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

.upload-tip {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 8px;
}
</style>
