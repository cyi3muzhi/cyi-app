<template>
  <div class="product-detail-page">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <svg class="loading-spinner" width="40" height="40" viewBox="0 0 24 24" fill="none">
        <circle class="spinner-path" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </svg>
      <p>加载中...</p>
    </div>

    <template v-else>
      <!-- 面包屑导航 -->
      <section class="container-shop pt-4 pb-2">
        <div class="breadcrumb text-sm">
          <router-link to="/" class="breadcrumb-link">首页</router-link>
          <span class="separator">/</span>
          <router-link to="/products" class="breadcrumb-link">全部商品</router-link>
          <span class="separator">/</span>
          <span class="current">{{ product?.name }}</span>
        </div>
      </section>

      <!-- 产品概览区域（两栏布局） -->
      <section class="container-shop py-4">
        <div class="product-overview">
          <!-- 左侧：图片画廊 -->
          <div class="gallery-col">
            <!-- 主图 -->
            <div class="main-image-wrap">
              <img :src="mainImage" :alt="product?.name" />
            </div>
            <!-- 缩略图列表 -->
            <div class="thumbnails-row">
              <div
                v-for="(img, index) in galleryImages"
                :key="index"
                class="thumb-item"
                :class="{ 'thumb-item--active': activeThumb === index }"
                @click="switchThumb(index)"
              >
                <img :src="img" :alt="`角度 ${index + 1}`" />
              </div>
            </div>
          </div>

          <!-- 右侧：产品信息 -->
          <div class="info-col">
            <!-- 产品名称 -->
            <h1 class="product-name">{{ product?.name }}</h1>

            <!-- 评分 -->
            <div class="rating-row">
              <div class="rating-stars">
                <span class="stars">{{ ratingStars }}</span>
              </div>
              <span class="rating-count">| 已售 {{ (product?.sales || 0).toLocaleString() }} 件</span>
            </div>

            <!-- 价格区域 -->
            <div class="price-section">
              <div class="price-row">
                <span class="current-price">&yen;{{ formatPrice(currentPrice) }}</span>
              </div>
            </div>

            <!-- 规格选择 -->
            <div v-for="spec in availableSpecs" :key="spec.name" class="option-section">
              <div class="option-label-row">
                <span class="option-label">{{ spec.name }}</span>
              </div>
              <div class="color-options">
                <button
                  v-for="value in spec.values"
                  :key="value"
                  class="color-pill"
                  :class="{ 'color-pill--active': selectedSpecs[spec.name] === value }"
                  @click="selectSpec(spec.name, value)"
                >
                  {{ value }}
                </button>
              </div>
            </div>

            <!-- 数量选择 -->
            <div class="option-section">
              <div class="option-label-row">
                <span class="option-label">数量</span>
              </div>
              <div class="qty-selector">
                <button class="qty-btn" @click="decrementQty" :disabled="quantity <= 1">-</button>
                <input type="number" class="qty-input" v-model.number="quantity" min="1" :max="currentStock" readonly />
                <button class="qty-btn" @click="incrementQty" :disabled="quantity >= currentStock">+</button>
                <span class="stock-hint">库存 {{ currentStock }} 件</span>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="action-buttons">
              <button class="btn btn-primary" @click="handleAddToCart">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
                  <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
                </svg>
                加入购物车
              </button>
              <button class="btn btn-buy" @click="handleBuyNow">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="5 11 9 15 19 5"/>
                </svg>
                立即购买
              </button>
            </div>

            <!-- 服务保障 -->
            <div class="service-guarantees">
              <div class="service-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                </svg>
                正品保障
              </div>
              <div class="service-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
                  <polyline points="21 3 21 9 15 9"/>
                </svg>
                7天退换
              </div>
              <div class="service-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
                </svg>
                免运费
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 产品详情（标签页） -->
      <section class="container-shop py-8">
        <!-- 标签栏 -->
        <div class="tab-bar">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-btn"
            :class="{ 'tab-btn--active': activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- 标签内容：商品详情 -->
        <div v-show="activeTab === 'detail'" class="tab-panel">
          <div v-html="product?.description"></div>
        </div>

        <!-- 标签内容：规格参数 -->
        <div v-show="activeTab === 'specs'" class="tab-panel">
          <div class="specs-table-wrap">
            <table class="specs-table">
              <tbody>
                <tr>
                  <td class="spec-label">商品名称</td>
                  <td class="spec-value">{{ product?.name }}</td>
                </tr>
                <tr>
                  <td class="spec-label">商品分类</td>
                  <td class="spec-value">{{ product?.category || '未分类' }}</td>
                </tr>
                <tr>
                  <td class="spec-label">售价</td>
                  <td class="spec-value">¥{{ formatPrice(product?.price || 0) }}</td>
                </tr>
                <tr>
                  <td class="spec-label">库存</td>
                  <td class="spec-value">{{ product?.stock }} 件</td>
                </tr>
                <tr>
                  <td class="spec-label">销量</td>
                  <td class="spec-value">{{ product?.sales }} 件</td>
                </tr>
                <tr v-for="spec in availableSpecs" :key="spec.name">
                  <td class="spec-label">{{ spec.name }}</td>
                  <td class="spec-value">{{ spec.values.join(' / ') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 标签内容：用户评价 -->
        <div v-show="activeTab === 'reviews'" class="tab-panel">
          <div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            <p>暂无用户评价</p>
          </div>
        </div>
      </section>

      <!-- 相关推荐 -->
      <section class="container-shop py-8 related-section">
        <div class="section-header">
          <div class="section-bar"></div>
          <h2 class="section-heading">相关推荐</h2>
        </div>
        <div class="related-grid">
          <ProductCard v-for="item in relatedProducts" :key="item.id" :product="item" />
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
/**
 * 商品详情页视图组件
 * 展示商品完整信息，包括图片画廊、商品信息、规格参数、用户评价和相关推荐
 */
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from '@/store/cart'
import ProductCard from '@/components/ProductCard.vue'
import { getProductDetail, getProducts, type ProductItem, type SkuInfo } from '@/api/product'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()

const loading = ref(true)
const product = ref<ProductItem | null>(null)

// ---- 图片画廊 ----
const galleryImages = computed(() => {
  if (!product.value) return []
  return product.value.images?.length ? product.value.images : [product.value.image]
})

const activeThumb = ref(0)

const mainImage = computed(() => galleryImages.value[activeThumb.value] || '')

function switchThumb(index: number) {
  activeThumb.value = index
}

// ---- 评分星星 ----
const ratingStars = computed(() => {
  const r = Math.round(product.value?.rating || 0)
  return '★'.repeat(r) + '☆'.repeat(5 - r)
})

// ---- 规格选择 ----
const selectedSpecs = ref<Record<string, string>>({})

const availableSpecs = computed(() => {
  return product.value?.specs || []
})

function selectSpec(specName: string, value: string) {
  selectedSpecs.value[specName] = value
}

// ---- 当前选中的 SKU ----
const selectedSku = computed<SkuInfo | undefined>(() => {
  if (!product.value?.skus?.length) return undefined
  return product.value.skus.find(sku => {
    for (const [key, value] of Object.entries(selectedSpecs.value)) {
      if (sku.spec[key] !== value) return false
    }
    return true
  })
})

const currentPrice = computed(() => {
  const skuPrice = selectedSku.value?.price
  if (skuPrice !== undefined && skuPrice !== null) {
    return skuPrice
  }
  return product.value?.price || 0
})

const currentStock = computed(() => {
  const skuStock = selectedSku.value?.stock
  if (skuStock !== undefined && skuStock !== null) {
    return skuStock
  }
  return product.value?.stock || 0
})

// ---- 数量选择 ----
const quantity = ref(1)

function incrementQty() {
  if (quantity.value < 99 && quantity.value < currentStock.value) {
    quantity.value++
  }
}

function decrementQty() {
  if (quantity.value > 1) {
    quantity.value--
  }
}

// ---- 操作按钮 ----
function handleAddToCart() {
  if (!product.value) return
  const skuInfo = selectedSku.value
  let skuDesc = ''
  if (skuInfo) {
    skuDesc = Object.values(skuInfo.spec).join(' / ')
  } else if (Object.keys(selectedSpecs.value).length) {
    skuDesc = Object.values(selectedSpecs.value).join(' / ')
  }
  cartStore.addItem({
    productId: Number(product.value.id),
    name: product.value.name,
    image: product.value.image,
    price: Number(currentPrice.value) / 100,
    sku: skuDesc || '默认规格',
    quantity: quantity.value
  })
}

function handleBuyNow() {
  handleAddToCart()
  router.push('/cart')
}

// ---- 标签页 ----
const tabs = computed(() => [
  { key: 'detail', label: '商品详情' },
  { key: 'specs', label: '规格参数' },
  { key: 'reviews', label: '用户评价' }
])

const activeTab = ref('detail')

// ---- 格式化价格 ----
function formatPrice(price: number | string): string {
  const p = Number(price) / 100
  return p.toFixed(2)
}

// ---- 剥离 HTML ----
function stripHtml(html: string): string {
  if (!html) return ''
  const div = document.createElement('div')
  div.innerHTML = html
  return div.textContent || div.innerText || ''
}

// ---- 相关推荐 ----
const relatedProducts = ref<ProductItem[]>([])

async function loadProduct() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getProductDetail(id)
    product.value = res.product

    availableSpecs.value.forEach(spec => {
      if (spec.values.length > 0 && !selectedSpecs.value[spec.name]) {
        selectedSpecs.value[spec.name] = spec.values[0]
      }
    })

    const relatedRes = await getProducts({ page: 1, pageSize: 4 })
    relatedProducts.value = relatedRes.items || []
  } catch (error) {
    console.error('加载商品详情失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProduct()
})
</script>

<style scoped>
/* ---- 面包屑 ---- */
.breadcrumb {
  display: flex;
  align-items: center;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}
.breadcrumb-link {
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color var(--shop-transition);
}
.breadcrumb-link:hover {
  color: var(--shop-primary);
}
.separator {
  margin: 0 8px;
  color: var(--shop-border);
}
.current {
  color: var(--shop-foreground);
  font-weight: 500;
}

/* ---- 产品概览两栏布局 ---- */
.product-overview {
  display: flex;
  flex-direction: column;
  gap: 32px;
}
@media (min-width: 1024px) {
  .product-overview {
    flex-direction: row;
  }
  .gallery-col {
    width: 55%;
  }
  .info-col {
    width: 45%;
  }
}
@media (min-width: 1280px) {
  .gallery-col {
    width: 50%;
  }
  .info-col {
    width: 50%;
  }
}

/* ---- 图片画廊 ---- */
.main-image-wrap {
  position: relative;
  overflow: hidden;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-muted);
  aspect-ratio: 1 / 1;
  cursor: crosshair;
  margin-bottom: 12px;
}
.main-image-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}
.main-image-wrap:hover img {
  transform: scale(1.15);
}

.thumbnails-row {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.thumb-item {
  width: 72px;
  height: 72px;
  border-radius: var(--shop-radius-md);
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color var(--shop-transition), opacity var(--shop-transition);
  opacity: 0.6;
  flex-shrink: 0;
}
.thumb-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.thumb-item:hover {
  opacity: 0.85;
}
.thumb-item--active {
  border-color: var(--shop-primary);
  opacity: 1;
}

/* ---- 产品信息 ---- */
.product-name {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin-bottom: 12px;
  line-height: 1.3;
}
@media (min-width: 768px) {
  .product-name {
    font-size: 1.875rem;
  }
}

.rating-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.rating-stars {
  display: flex;
  align-items: center;
  gap: 4px;
}
.stars {
  display: inline-flex;
  gap: 1px;
  color: hsl(38, 92%, 50%);
  font-size: 0.8125rem;
  letter-spacing: 1px;
}
.rating-score {
  font-size: 0.875rem;
  font-weight: 600;
  color: hsl(38, 92%, 50%);
}
.rating-count {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

/* ---- 价格区域 ---- */
.price-section {
  border-radius: var(--shop-radius-xl);
  padding: 20px;
  margin-bottom: 20px;
  background: var(--shop-primary-bg);
}
.price-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 8px;
}
.current-price {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--shop-destructive);
}
.original-price {
  font-size: 1rem;
  text-decoration: line-through;
  color: var(--shop-muted-foreground);
}
.save-badge {
  font-size: 0.8125rem;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 999px;
  background: var(--shop-destructive);
  color: var(--shop-destructive-foreground);
}

.promo-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}
.promo-tag {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--shop-destructive);
  color: var(--shop-destructive-foreground);
}
.promo-text {
  font-size: 0.875rem;
  color: var(--shop-destructive);
}

/* ---- 选项区域 ---- */
.option-section {
  margin-bottom: 20px;
}
.option-label-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.option-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}
.selected-value {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

/* ---- 颜色选择器 ---- */
.color-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.color-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  border-radius: var(--shop-radius-xl);
  border: 2px solid var(--shop-border);
  background: var(--shop-background);
  cursor: pointer;
  transition: all var(--shop-transition);
  font-size: 0.875rem;
  font-weight: 500;
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
}
.color-pill:hover {
  border-color: var(--shop-primary-light);
}
.color-pill--active {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}
.color-dot {
  width: 18px;
  height: 18px;
  border-radius: 999px;
  flex-shrink: 0;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

/* ---- 数量选择器 ---- */
.qty-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}
.qty-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--shop-border);
  background: var(--shop-background);
  cursor: pointer;
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--shop-foreground);
  transition: background var(--shop-transition), border-color var(--shop-transition);
  user-select: none;
  font-family: var(--shop-font-sans);
}
.qty-btn:first-child {
  border-radius: var(--shop-radius-md) 0 0 var(--shop-radius-md);
}
.qty-btn:last-child {
  border-radius: 0 var(--shop-radius-md) var(--shop-radius-md) 0;
}
.qty-btn:hover {
  background: var(--shop-muted);
  border-color: var(--shop-primary-light);
}
.qty-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  border-color: var(--shop-border);
  color: var(--shop-muted-foreground);
  background: var(--shop-muted);
}

.qty-input {
  width: 56px;
  height: 36px;
  text-align: center;
  border-top: 1px solid var(--shop-border);
  border-bottom: 1px solid var(--shop-border);
  border-left: none;
  border-right: none;
  font-size: 0.9375rem;
  font-weight: 600;
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
  outline: none;
  -moz-appearance: textfield;
  background: var(--shop-background);
}
.qty-input::-webkit-inner-spin-button,
.qty-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.stock-hint {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

/* ---- 操作按钮 ---- */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}
@media (min-width: 640px) {
  .action-buttons {
    flex-direction: row;
  }
}

.btn {
  flex: 1;
  height: 48px;
  border-radius: var(--shop-radius-xl);
  font-size: 1rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
  font-family: var(--shop-font-sans);
}
.btn:hover {
  box-shadow: var(--shop-shadow-lg);
}
.btn:active {
  transform: scale(0.98);
}

.btn-primary {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}

.btn-buy {
  background: var(--shop-destructive);
  color: var(--shop-destructive-foreground);
}

/* ---- 服务保障 ---- */
.service-guarantees {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--shop-border);
}

.service-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--shop-muted-foreground);
  font-size: 0.8125rem;
}
.service-item svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--shop-success);
}

/* ---- 标签切换 ---- */
.tab-bar {
  display: flex;
  border-bottom: 1px solid var(--shop-border);
  margin-bottom: 24px;
}

.tab-btn {
  padding: 12px 28px;
  font-size: 0.9375rem;
  font-weight: 500;
  border: none;
  background: none;
  cursor: pointer;
  color: var(--shop-muted-foreground);
  position: relative;
  transition: color var(--shop-transition);
  font-family: var(--shop-font-sans);
  white-space: nowrap;
}
.tab-btn::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%) scaleX(0);
  width: 60%;
  height: 3px;
  background: var(--shop-primary);
  border-radius: 3px 3px 0 0;
  transition: transform var(--shop-transition);
}
.tab-btn:hover {
  color: var(--shop-foreground);
}
.tab-btn--active {
  color: var(--shop-primary);
}
.tab-btn--active::after {
  transform: translateX(-50%) scaleX(1);
}

.tab-panel {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ---- 商品详情 ---- */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 32px;
}
@media (min-width: 768px) {
  .detail-grid {
    grid-template-columns: 1fr 1fr;
  }
}

.section-title {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 16px;
  color: var(--shop-foreground);
}

.highlight-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.highlight-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.highlight-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--shop-primary);
  margin-top: 8px;
  flex-shrink: 0;
}
.highlight-text {
  font-size: 0.875rem;
  color: var(--shop-foreground);
  line-height: 1.6;
}

.package-table-wrap {
  border-radius: var(--shop-radius-xl);
  padding: 20px;
  background: var(--shop-muted);
}

.package-table {
  width: 100%;
  font-size: 0.875rem;
  border-collapse: collapse;
}
.package-table td {
  padding: 8px 0;
}
.pkg-label {
  font-weight: 500;
  color: var(--shop-muted-foreground);
}
.pkg-value {
  text-align: right;
  color: var(--shop-foreground);
}

/* ---- 规格参数 ---- */
.specs-table-wrap {
  max-width: 560px;
}
.specs-table {
  width: 100%;
  font-size: 0.875rem;
  border-collapse: collapse;
}
.specs-table tr {
  border-bottom: 1px solid var(--shop-border);
}
.specs-table td {
  padding: 12px 0;
}
.spec-label {
  width: 160px;
  font-weight: 500;
  color: var(--shop-muted-foreground);
}
.spec-value {
  color: var(--shop-foreground);
}

/* ---- 评价摘要 ---- */
.review-summary {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  border-radius: var(--shop-radius-xl);
  background: var(--shop-muted);
  margin-bottom: 32px;
}
@media (min-width: 768px) {
  .review-summary {
    flex-direction: row;
    gap: 32px;
  }
}

.review-overall {
  display: flex;
  flex-direction: column;
  align-items: center;
}
@media (min-width: 768px) {
  .review-overall {
    align-items: flex-start;
    min-width: 180px;
  }
}

.review-score {
  font-size: 2.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin-bottom: 4px;
}

.review-stars {
  font-size: 1.125rem;
  margin-bottom: 4px;
}

.review-total {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.rating-bars {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rating-bar-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rating-bar-label {
  font-size: 0.875rem;
  width: 40px;
  text-align: right;
  color: var(--shop-muted-foreground);
}

.rating-bar-track {
  flex: 1;
  height: 8px;
  border-radius: 999px;
  background: var(--shop-border);
  overflow: hidden;
}

.rating-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: hsl(38, 92%, 50%);
  transition: width 0.4s ease;
}

.rating-bar-value {
  font-size: 0.875rem;
  width: 40px;
  color: var(--shop-muted-foreground);
}

/* ---- 评价卡片 ---- */
.review-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 32px;
}

.review-card {
  border-radius: var(--shop-radius-xl);
  padding: 20px;
  border: 1px solid var(--shop-border);
  background: var(--shop-card);
}

.review-card-inner {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.review-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.review-body {
  flex: 1;
  min-width: 0;
}

.review-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.reviewer-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.review-date {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.review-content {
  font-size: 0.875rem;
  color: var(--shop-foreground);
  margin-bottom: 12px;
  line-height: 1.6;
}

.review-images {
  display: flex;
  gap: 8px;
}

.review-image-item {
  width: 64px;
  height: 64px;
  border-radius: var(--shop-radius-md);
  overflow: hidden;
  background: var(--shop-muted);
}
.review-image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* ---- 查看全部评价 ---- */
.view-all-reviews {
  text-align: center;
}

.view-all-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-primary);
  text-decoration: none;
  transition: color var(--shop-transition);
}
.view-all-link:hover {
  text-decoration: underline;
}

/* ---- 相关推荐 ---- */
.related-section {
  padding-top: 32px;
  padding-bottom: 48px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.section-bar {
  width: 4px;
  height: 20px;
  background: var(--shop-primary);
  border-radius: 2px;
}

.section-heading {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.related-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
@media (min-width: 768px) {
  .related-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
  }
}
</style>