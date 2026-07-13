<template>
  <!-- 商品管理页面 - 工具栏 + 批量操作 + 商品表格 + 分页 -->
  <div class="product-list">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="$router.push('/products/add')">
        <el-icon><Plus /></el-icon>
        添加商品
      </el-button>
      <el-input
        v-model="searchKeyword"
        placeholder="搜索商品名称..."
        :prefix-icon="Search"
        class="search-input"
        clearable
        @input="handleSearch"
      />
      <el-select v-model="categoryFilter" placeholder="全部分类" clearable @change="handleSearch">
        <el-option label="全部分类" value="" />
        <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
      </el-select>
      <el-select v-model="statusFilter" placeholder="全部状态" clearable @change="handleSearch">
        <el-option label="全部状态" value="" />
        <el-option label="上架中" value="上架中" />
        <el-option label="已下架" value="已下架" />
        <el-option label="缺货" value="缺货" />
      </el-select>
    </div>

    <!-- 批量操作栏 -->
    <div v-if="selectedIds.length > 0" class="bulk-bar">
      <span>已选择 <strong>{{ selectedIds.length }}</strong> 项</span>
      <div class="bulk-actions">
        <el-button size="small" @click="bulkAction('上架')">批量上架</el-button>
        <el-button size="small" @click="bulkAction('下架')">批量下架</el-button>
        <el-button size="small" type="danger" @click="bulkAction('删除')">删除所选</el-button>
      </div>
    </div>

    <!-- 商品表格 -->
    <el-table
      :data="filteredProducts"
      stripe
      style="width: 100%"
      @selection-change="handleSelectionChange"
      v-loading="loading"
    >
      <el-table-column type="selection" width="45" />
      <el-table-column label="图片" width="80">
        <template #default>
          <div class="product-thumb">
            <el-icon :size="20"><Picture /></el-icon>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="商品名称" min-width="220">
        <template #default="{ row }">
          <div class="product-name-cell">
            <div>
              <div class="name-text">{{ row.name }}</div>
              <div class="sku-text">SKU: {{ row.sku }}</div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="分类" width="120" />
      <el-table-column label="价格" width="110">
        <template #default="{ row }">
          <span class="price-text">&yen;{{ row.price.toFixed(2) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="库存" width="90">
        <template #default="{ row }">
          <span v-if="row.stock > 0" class="stock-text">{{ row.stock }}</span>
          <span v-else class="stock-empty">缺货</span>
        </template>
      </el-table-column>
      <el-table-column prop="sales" label="销量" width="80">
        <template #default="{ row }">
          {{ row.sales.toLocaleString() }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag
            :type="statusTagType(row.status)"
            size="small"
            effect="plain"
          >
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="handleEdit(row)">
            编辑
          </el-button>
          <el-button
            text
            size="small"
            :type="row.status === '上架中' ? 'warning' : 'success'"
            @click="handleToggleStatus(row)"
          >
            {{ row.status === '上架中' ? '下架' : '上架' }}
          </el-button>
          <el-button text size="small" type="danger" @click="handleDelete(row)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="filteredProducts.length"
        layout="total, prev, pager, next, jumper"
        background
        small
      />
    </div>
  </div>
</template>

<script setup>
/**
 * 商品管理列表页面
 * 支持搜索筛选、批量操作、上下架切换、删除确认
 */
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Picture } from '@element-plus/icons-vue'
import { products, productCategories } from '../../api/mock'

const router = useRouter()

/** 搜索关键词 */
const searchKeyword = ref('')

/** 分类筛选 */
const categoryFilter = ref('')

/** 状态筛选 */
const statusFilter = ref('')

/** 选中行 ID 列表 */
const selectedIds = ref([])

/** 当前页 */
const currentPage = ref(1)

/** 每页条数 */
const pageSize = ref(10)

/** 加载状态 */
const loading = ref(false)

/** 商品分类列表 */
const categories = productCategories

/** 过滤后的商品列表 */
const filteredProducts = computed(() => {
  let list = [...products]
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(p => p.name.toLowerCase().includes(kw))
  }
  if (categoryFilter.value) {
    list = list.filter(p => p.category === categoryFilter.value)
  }
  if (statusFilter.value) {
    list = list.filter(p => p.status === statusFilter.value)
  }
  return list
})

/**
 * 获取状态对应的 el-tag 类型
 * @param {string} status 商品状态
 * @returns {string} tag 类型
 */
function statusTagType(status) {
  if (status === '上架中') return 'success'
  if (status === '已下架') return 'info'
  return 'danger'
}

/**
 * 处理选择框变化
 * @param {Array} selection 选中行数组
 */
function handleSelectionChange(selection) {
  selectedIds.value = selection.map(item => item.id)
}

/**
 * 搜索/筛选输入处理
 */
function handleSearch() {
  currentPage.value = 1
}

/**
 * 编辑商品
 * @param {Object} row 商品行数据
 */
function handleEdit(row) {
  router.push(`/products/edit/${row.id}`)
}

/**
 * 切换上下架状态
 * @param {Object} row 商品行数据
 */
function handleToggleStatus(row) {
  const newStatus = row.status === '上架中' ? '已下架' : '上架中'
  ElMessageBox.confirm(
    `确定将商品「${row.name}」${newStatus}吗？`,
    '提示',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    row.status = newStatus
    ElMessage.success(`商品已${newStatus}`)
  }).catch(() => {})
}

/**
 * 删除商品
 * @param {Object} row 商品行数据
 */
function handleDelete(row) {
  ElMessageBox.confirm(
    `确定要删除商品「${row.name}」吗？此操作不可撤销。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
  ).then(() => {
    const idx = products.findIndex(p => p.id === row.id)
    if (idx !== -1) {
      products.splice(idx, 1)
    }
    ElMessage.success('商品已删除')
  }).catch(() => {})
}

/**
 * 批量操作
 * @param {string} action 操作类型：上架/下架/删除
 */
function bulkAction(action) {
  const count = selectedIds.value.length
  if (count === 0) {
    ElMessage.warning('请先选择商品')
    return
  }

  const actionText = action === '上架' || action === '下架' ? action : '删除'
  ElMessageBox.confirm(
    `确定${actionText}选中的 ${count} 件商品吗？`,
    '批量操作',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(() => {
    if (action === '删除') {
      selectedIds.value.forEach(id => {
        const idx = products.findIndex(p => p.id === id)
        if (idx !== -1) products.splice(idx, 1)
      })
    } else {
      const newStatus = action === '上架' ? '上架中' : '已下架'
      products.forEach(p => {
        if (selectedIds.value.includes(p.id)) {
          p.status = newStatus
        }
      })
    }
    selectedIds.value = []
    ElMessage.success(`批量${actionText}成功`)
  }).catch(() => {})
}
</script>

<style scoped>
.product-list {
  max-width: 1200px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px 20px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
}

.search-input {
  width: 240px;
}

.bulk-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  background: var(--shop-primary-bg);
  border: 1px solid var(--shop-primary);
  border-radius: var(--shop-radius-md);
  margin-bottom: 12px;
  font-size: 0.8125rem;
  color: var(--shop-foreground);
}

.bulk-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.product-thumb {
  width: 44px;
  height: 44px;
  border-radius: var(--shop-radius-sm);
  background: var(--shop-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--shop-muted-foreground);
}

.product-name-cell .name-text {
  font-weight: 500;
  color: var(--shop-foreground);
  max-width: 260px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-name-cell .sku-text {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
  margin-top: 2px;
}

.price-text {
  font-weight: 600;
  color: var(--shop-foreground);
}

.stock-empty {
  color: var(--shop-destructive);
  font-weight: 500;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  padding: 20px 0 8px;
}
</style>
