/**
 * 购物车状态管理
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  /** 购物车商品列表 */
  const items = ref([
    // 演示数据
    { id: 1, productId: 1, name: '极简商务休闲双肩包', image: '/images/product-1.jpg', price: 299, quantity: 1, selected: true, sku: '深蓝 / 均码', stock: 99 },
    { id: 2, productId: 2, name: '简约无线蓝牙耳机', image: '/images/product-2.jpg', price: 159, quantity: 2, selected: true, sku: '白色 / 标准版', stock: 99 },
    { id: 3, productId: 3, name: '陶瓷杯简约办公水杯', image: '/images/product-3.jpg', price: 89, quantity: 1, selected: true, sku: '米白 / 350ml', stock: 99 }
  ])

  /** 选中商品 */
  const selectedItems = computed(() => items.value.filter(i => i.selected))

  /** 选中数量 */
  const selectedCount = computed(() => selectedItems.value.length)

  /** 商品总件数 */
  const totalCount = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))

  /** 商品小计 */
  const subtotal = computed(() => selectedItems.value.reduce((s, i) => s + i.price * i.quantity, 0))

  /** 选中/取消选中 */
  function toggleSelect(id: number) {
    const item = items.value.find(i => i.id === id)
    if (item) item.selected = !item.selected
  }

  /** 全选/取消全选 */
  function toggleSelectAll(selected: boolean) {
    items.value.forEach(i => { i.selected = selected })
  }

  /** 更新数量 */
  function updateQuantity(id: number, quantity: number) {
    const item = items.value.find(i => i.id === id)
    if (item) item.quantity = Math.max(1, Math.min(quantity, item.stock))
  }

  /** 删除商品 */
  function removeItem(id: number) {
    items.value = items.value.filter(i => i.id !== id)
  }

  /** 添加商品 */
  function addItem(product: { productId: number; name: string; image: string; price: number; sku?: string; quantity?: number }) {
    const exist = items.value.find(i => i.productId === product.productId)
    const qty = product.quantity || 1
    if (exist) {
      exist.quantity += qty
    } else {
      const id = Date.now()
      items.value.push({ ...product, id, quantity: qty, selected: true, stock: 99, sku: product.sku || '' })
    }
  }

  return { items, selectedItems, selectedCount, totalCount, subtotal, toggleSelect, toggleSelectAll, updateQuantity, removeItem, addItem }
})
