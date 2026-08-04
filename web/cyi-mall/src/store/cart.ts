/**
 * 购物车状态管理
 * 登录态：数据存储在后端（Redis，按用户隔离）
 * 未登录：数据存储在浏览器本地（localStorage），登录后自动合并到服务端
 */
import { defineStore, acceptHMRUpdate } from 'pinia'
import { ref, computed } from 'vue'
import { addToCart, getCartList, updateCartItem, toggleCartItem, selectAllCart, deleteCartItem, clearCart } from '@/api/cart'
import type { CartItem } from '@/api/cart'

/** 购物车展示条目（价格统一转为元） */
export interface CartDisplayItem {
  productId: number
  skuId: number // SKU ID（0 表示商品默认 SKU）
  name: string
  price: number // 元
  quantity: number
  image: string
  stock: number
  selected: boolean
  sku: string
  status: number
}

/** 本地购物车条目（未登录时缓存商品快照，用于合并与展示） */
interface LocalCartItem {
  productId: number
  skuId: number // SKU ID（0 表示商品默认 SKU）
  quantity: number
  name: string
  image: string
  price: number // 元
  sku: string
}

/** 本地购物车存储 key */
const LOCAL_KEY = 'cyi_cart_local'

/** 读取本地购物车 */
function readLocal(): LocalCartItem[] {
  try {
    return JSON.parse(localStorage.getItem(LOCAL_KEY) || '[]') as LocalCartItem[]
  } catch {
    return []
  }
}

/** 写入本地购物车 */
function writeLocal(items: LocalCartItem[]) {
  localStorage.setItem(LOCAL_KEY, JSON.stringify(items))
}

/** 是否已登录（与登录态同源：localStorage token） */
function isLoggedIn(): boolean {
  return !!localStorage.getItem('token')
}

export const useCartStore = defineStore('cart', () => {
  /** 购物车商品列表（价格单位为元） */
  const items = ref<CartDisplayItem[]>([])

  /** 是否加载中 */
  const loading = ref(false)

  /** 选中商品 */
  const selectedItems = computed(() => items.value.filter(i => i.selected))

  /** 选中商品件数 */
  const selectedCount = computed(() => selectedItems.value.length)

  /** 商品总件数 */
  const totalCount = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))

  /** 商品小计（元） */
  const subtotal = computed(() => selectedItems.value.reduce((s, i) => s + i.price * i.quantity, 0))

  /** 是否全部选中 */
  const isAllSelected = computed(() => items.value.length > 0 && items.value.every(i => i.selected))

  /**
   * 加载购物车：登录读后端，未登录读本地
   */
  async function loadCart() {
    loading.value = true
    try {
      if (isLoggedIn()) {
        const res = await getCartList()
        items.value = (res.items || []).map(serverToDisplay)
      } else {
        items.value = readLocal().map(localToDisplay)
      }
    } catch {
      items.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 加入购物车：登录走后端（同商品同规格累加），未登录存本地
   * @param product 商品快照信息
   */
  async function addItem(product: { productId: number; skuId: number; name: string; image: string; price: number; sku?: string; quantity?: number }) {
    const quantity = product.quantity || 1
    if (isLoggedIn()) {
      await addToCart({ productId: product.productId, skuId: product.skuId, quantity })
    } else {
      const local = readLocal()
      const exist = local.find(i => i.productId === product.productId && i.skuId === product.skuId)
      if (exist) {
        exist.quantity += quantity
      } else {
        local.push({
          productId: product.productId,
          skuId: product.skuId,
          quantity,
          name: product.name,
          image: product.image,
          price: product.price,
          sku: product.sku || '默认规格'
        })
      }
      writeLocal(local)
    }
    await loadCart()
  }

  /**
   * 修改数量
   * @param productId 商品ID
   * @param skuId SKU ID
   * @param quantity 目标数量
   */
  async function updateQuantity(productId: number, skuId: number, quantity: number) {
    const next = Math.max(1, quantity)
    if (isLoggedIn()) {
      await updateCartItem({ productId, skuId, quantity: next })
    } else {
      const local = readLocal()
      const item = local.find(i => i.productId === productId && i.skuId === skuId)
      if (item) item.quantity = next
      writeLocal(local)
    }
    await loadCart()
  }

  /**
   * 选中/取消选中单个商品
   */
  async function toggleSelect(productId: number, skuId: number) {
    const item = items.value.find(i => i.productId === productId && i.skuId === skuId)
    if (!item) return
    const selected = !item.selected
    if (isLoggedIn()) {
      await toggleCartItem({ productId, skuId, selected })
    } else {
      const local = readLocal()
      const l = local.find(i => i.productId === productId && i.skuId === skuId)
      if (l) (l as LocalCartItem & { selected: boolean }).selected = selected
      // 本地快照不持久化选中态（刷新默认全选），仅在内存中记录
      item.selected = selected
      return
    }
    item.selected = selected
  }

  /**
   * 全选/取消全选
   */
  async function toggleSelectAll(selected: boolean) {
    if (isLoggedIn()) {
      await selectAllCart(selected)
      await loadCart()
    } else {
      items.value.forEach(i => { i.selected = selected })
    }
  }

  /**
   * 删除商品
   */
  async function removeItem(productId: number, skuId = 0) {
    if (isLoggedIn()) {
      await deleteCartItem(productId, skuId)
    } else {
      writeLocal(readLocal().filter(i => !(i.productId === productId && (skuId === 0 || i.skuId === skuId))))
    }
    await loadCart()
  }

  /**
   * 清空购物车
   */
  async function clear() {
    if (isLoggedIn()) {
      await clearCart()
    } else {
      writeLocal([])
    }
    await loadCart()
  }

  /**
   * 登录成功后合并本地购物车到服务端（仅在登录态下调用）
   */
  async function mergeLocalCart() {
    const local = readLocal()
    if (!local.length) {
      await loadCart()
      return
    }
    for (const item of local) {
      try {
        await addToCart({ productId: item.productId, skuId: item.skuId, quantity: item.quantity })
      } catch {
        // 单个失败不阻塞合并，跳过继续
      }
    }
    writeLocal([])
    await loadCart()
  }

  /** 退出登录时清空内存数据（本地购物车保留，下次登录合并） */
  function reset() {
    items.value = []
  }

  return {
    items, loading, selectedItems, selectedCount, totalCount, subtotal, isAllSelected,
    loadCart, addItem, updateQuantity, toggleSelect, toggleSelectAll, removeItem, clear, mergeLocalCart, reset
  }
})

/** 后端条目转展示条目（分 → 元） */
function serverToDisplay(i: CartItem): CartDisplayItem {
  return {
    productId: i.productId,
    skuId: i.skuId || 0,
    name: i.name,
    price: Number((i.price / 100).toFixed(2)),
    quantity: i.quantity,
    image: i.image,
    stock: i.stock,
    selected: i.selected,
    sku: i.sku,
    status: i.status
  }
}

/** 本地条目转展示条目 */
function localToDisplay(i: LocalCartItem): CartDisplayItem {
  return {
    productId: i.productId,
    skuId: i.skuId || 0,
    name: i.name,
    price: Number(i.price.toFixed(2)),
    quantity: i.quantity,
    image: i.image,
    stock: 99,
    selected: true,
    sku: i.sku,
    status: 1
  }
}

// 启用 store 热更新（修改此文件时无需整页刷新）
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useCartStore, import.meta.hot))
}
