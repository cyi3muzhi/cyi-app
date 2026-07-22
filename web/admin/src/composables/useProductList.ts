// web/admin/src/composables/useProductList.ts
// 商品列表页业务逻辑

import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { listProducts, batchProductOperation, deleteProduct } from '../api/product'

export interface FilterState {
  keyword: string
  category_id: number | undefined
  status: string
  stock_eq: number | undefined  // 0=缺货
}

export function useProductList() {
  const products = ref<any[]>([])
  const totalCount = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(10)
  const loading = ref(false)
  const selectedIds = ref<number[]>([])
  const filters = ref<FilterState>({ keyword: '', category_id: undefined, status: '', stock_eq: undefined })

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
        category_id: filters.value.category_id || undefined,
        status: filters.value.status || undefined,
        stock_eq: filters.value.stock_eq,
      })
      totalCount.value = res.total
      products.value = res.items || []
    } catch {
      ElMessage.error('加载商品列表失败')
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
      window.dispatchEvent(new CustomEvent('product-changed'))
      loadProducts()
    } catch {
      ElMessage.error(`${action}失败`)
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteProduct(id)
      ElMessage.success('删除成功')
      window.dispatchEvent(new CustomEvent('product-changed'))
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
      window.dispatchEvent(new CustomEvent('product-changed'))
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
