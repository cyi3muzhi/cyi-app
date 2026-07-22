// web/admin/src/composables/useProductForm.ts
// 商品添加/编辑表单业务逻辑

import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

interface SpecGroupForm {
  id: number
  name: string
  values: string[]
}

interface SkuForm {
  id: string
  spec: Record<string, string>
  price: number   // 元
  stock: number   // 库存
}

export interface ProductFormState {
  name: string
  category_id: number
  image: string
  _imageKey?: string  // 直传 OSS 后的 objectKey，提交时使用
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
    name: '', category_id: 0, image: '', description: '',
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
      form.category_id = (p as any).categoryId
      form.image = p.image || ''
      form.description = p.description || ''
      // specs
      specGroups.value = (p.specs || []).map(s => ({ id: ++specIdCounter.value, name: s.name, values: [...s.values] }))
      // skus（价格从分转元）
      skus.value = (p.skus || []).map(s => ({
        id: `sku_${s.id}`,
        spec: s.spec,
        price: s.price / 100,
        stock: s.stock || 0,
      }))
    } finally {
      loading.value = false
    }
  }

  /** 校验表单 */
  function validate(): boolean {
    if (!form.name.trim()) { ElMessage.warning('请填写商品名称'); return false }
    if (!form.category_id || form.category_id == 0) { ElMessage.warning('请选择商品分类'); return false }
    return true
  }

  /** 构建提交数据 */
  function buildSubmitData() {
    const specs = specGroups.value.filter(g => g.name)
    const formedSkus = skus.value.map(s => ({
      sku_code: Object.values(s.spec).join('-'),
      price: Math.round(s.price * 100),  // 元 → 分
      stock: s.stock,
      spec: s.spec,
    }))
    return {
      name: form.name,
      category_id: form.category_id,
      image: form._imageKey || form.image || undefined,
      price: formedSkus.length > 0 ? formedSkus[0].price : 0,
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
