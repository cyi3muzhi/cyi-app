<template>
  <div class="product-detail-page">
    <!-- 面包屑导航 -->
    <section class="container-shop pt-4 pb-2">
      <div class="breadcrumb text-sm">
        <router-link to="/" class="breadcrumb-link">首页</router-link>
        <span class="separator">/</span>
        <router-link to="/products" class="breadcrumb-link">全部商品</router-link>
        <span class="separator">/</span>
        <span class="current">无线降噪耳机 Pro</span>
      </div>
    </section>

    <!-- 产品概览区域（两栏布局） -->
    <section class="container-shop py-4">
      <div class="product-overview">
        <!-- 左侧：图片画廊 -->
        <div class="gallery-col">
          <!-- 主图 -->
          <div class="main-image-wrap">
            <img :src="mainImage" alt="无线降噪耳机 Pro" />
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
          <h1 class="product-name">无线降噪耳机 Pro</h1>

          <!-- 评分 -->
          <div class="rating-row">
            <div class="rating-stars">
              <span class="stars">{{ ratingStars }}</span>
              <span class="rating-score">4.5</span>
            </div>
            <span class="rating-count">| 已有 2,368 条评价</span>
          </div>

          <!-- 价格区域 -->
          <div class="price-section">
            <div class="price-row">
              <span class="current-price">&yen;1,299.00</span>
              <span class="original-price">&yen;1,599.00</span>
              <span class="save-badge">省 &yen;300</span>
            </div>
            <div class="promo-row">
              <span class="promo-tag">促销</span>
              <span class="promo-text">满199减30 | 满500减100</span>
            </div>
          </div>

          <!-- 颜色选择 -->
          <div class="option-section">
            <div class="option-label-row">
              <span class="option-label">颜色</span>
              <span class="selected-value" id="selected-color-name">{{ selectedColor }}</span>
            </div>
            <div class="color-options">
              <button
                v-for="color in colorOptions"
                :key="color.name"
                class="color-pill"
                :class="{ 'color-pill--active': selectedColor === color.name }"
                @click="selectColor(color.name)"
              >
                <span class="color-dot" :style="{ background: color.hex }"></span>
                {{ color.name }}
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
              <input type="number" class="qty-input" v-model.number="quantity" min="1" max="99" readonly />
              <button class="qty-btn" @click="incrementQty" :disabled="quantity >= 99">+</button>
              <span class="stock-hint">库存 2,586 件</span>
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
        <div class="detail-grid">
          <div>
            <h3 class="section-title">产品亮点</h3>
            <ul class="highlight-list">
              <li v-for="(item, i) in highlights" :key="i" class="highlight-item">
                <span class="highlight-dot"></span>
                <span class="highlight-text">
                  <strong>{{ item.title }}</strong> — {{ item.desc }}
                </span>
              </li>
            </ul>
          </div>
          <div>
            <h3 class="section-title">包装清单</h3>
            <div class="package-table-wrap">
              <table class="package-table">
                <tbody>
                  <tr v-for="(pkg, i) in packageList" :key="i">
                    <td class="pkg-label">{{ pkg.name }}</td>
                    <td class="pkg-value">{{ pkg.qty }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- 标签内容：规格参数 -->
      <div v-show="activeTab === 'specs'" class="tab-panel">
        <div class="specs-table-wrap">
          <table class="specs-table">
            <tbody>
              <tr v-for="(spec, i) in specifications" :key="i">
                <td class="spec-label">{{ spec.label }}</td>
                <td class="spec-value">{{ spec.value }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 标签内容：用户评价 -->
      <div v-show="activeTab === 'reviews'" class="tab-panel">
        <!-- 评价摘要 -->
        <div class="review-summary">
          <div class="review-overall">
            <div class="review-score">4.5</div>
            <div class="stars review-stars">★★★★★</div>
            <div class="review-total">2,368 条评价</div>
          </div>
          <div class="rating-bars">
            <div v-for="(bar, i) in ratingDistribution" :key="i" class="rating-bar-row">
              <span class="rating-bar-label">{{ bar.star }} 星</span>
              <div class="rating-bar-track">
                <div class="rating-bar-fill" :style="{ width: bar.percent + '%' }"></div>
              </div>
              <span class="rating-bar-value">{{ bar.percent }}%</span>
            </div>
          </div>
        </div>

        <!-- 评价卡片列表 -->
        <div class="review-cards">
          <div v-for="(review, i) in reviews" :key="i" class="review-card">
            <div class="review-card-inner">
              <div class="review-avatar" :style="{ background: review.avatarColor }">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
                </svg>
              </div>
              <div class="review-body">
                <div class="review-meta">
                  <span class="reviewer-name">{{ review.name }}</span>
                  <span class="stars">{{ review.stars }}</span>
                  <span class="review-date">{{ review.date }}</span>
                </div>
                <p class="review-content">{{ review.content }}</p>
                <div v-if="review.images?.length" class="review-images">
                  <div v-for="(img, j) in review.images" :key="j" class="review-image-item">
                    <img :src="img" alt="评价配图" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="view-all-reviews">
          <a href="#" class="view-all-link">
            查看全部评价
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
          </a>
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
  </div>
</template>

<script setup lang="ts">
/**
 * 商品详情页视图组件
 * 展示商品完整信息，包括图片画廊、商品信息、规格参数、用户评价和相关推荐
 */
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from '@/store/cart'
import ProductCard from '@/components/ProductCard.vue'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()

// ---- 图片画廊 ----
const galleryImages = [
  '/images/product-1.jpg',
  '/images/product-2.jpg',
  '/images/product-3.jpg',
  '/images/product-4.jpg',
  '/images/product-5.jpg',
  '/images/product-6.jpg'
]

const activeThumb = ref(0)

/** 当前显示的主图 */
const mainImage = computed(() => galleryImages[activeThumb.value])

/** 切换缩略图 */
function switchThumb(index: number) {
  activeThumb.value = index
}

// ---- 评分星星 ----
const ratingStars = '★★★★★'

// ---- 颜色选择 ----
const colorOptions = [
  { name: '曜石黑', hex: '#1a1a1e' },
  { name: '陶瓷白', hex: '#f0f0f0' },
  { name: '星空蓝', hex: '#2d4a7a' }
]

const selectedColor = ref('曜石黑')

/** 选择颜色 */
function selectColor(color: string) {
  selectedColor.value = color
}

// ---- 数量选择 ----
const quantity = ref(1)

/** 增加数量 */
function incrementQty() {
  if (quantity.value < 99) {
    quantity.value++
  }
}

/** 减少数量 */
function decrementQty() {
  if (quantity.value > 1) {
    quantity.value--
  }
}

// ---- 操作按钮 ----
/** 加入购物车 */
function handleAddToCart() {
  cartStore.addItem({
    productId: Number(route.params.id),
    name: '无线降噪耳机 Pro',
    image: '/images/product-1.jpg',
    price: 1299,
    sku: selectedColor.value
  })
}

/** 立即购买 */
function handleBuyNow() {
  router.push('/cart')
}

// ---- 标签页 ----
const tabs = [
  { key: 'detail', label: '商品详情' },
  { key: 'specs', label: '规格参数' },
  { key: 'reviews', label: '用户评价 (2,368)' }
]

const activeTab = ref('detail')

// ---- 商品详情数据 ----
const highlights = [
  { title: '主动降噪', desc: '混合式主动降噪技术，最高 40dB 降噪深度，让您沉浸于纯粹的音乐世界' },
  { title: 'Hi-Res 认证', desc: '支持 LDAC 高清音频编码，无损音质传输，细节丰富饱满' },
  { title: '超长续航', desc: '单次充电续航 40 小时，支持快充 10 分钟播放 3 小时' },
  { title: '舒适佩戴', desc: '人体工学设计，仅重 198g，记忆棉耳罩，长时间佩戴无压力' },
  { title: '多点连接', desc: '支持同时连接两台设备，无缝切换手机与电脑' }
]

const packageList = [
  { name: '耳机主体', qty: 'x 1' },
  { name: 'USB-C 充电线', qty: 'x 1' },
  { name: '3.5mm 音频线', qty: 'x 1' },
  { name: '收纳盒', qty: 'x 1' },
  { name: '说明书 & 保修卡', qty: 'x 1' }
]

// ---- 规格参数 ----
const specifications = [
  { label: '品牌', value: '悦选' },
  { label: '型号', value: 'YX-NC100 Pro' },
  { label: '连接方式', value: '蓝牙 5.3 / 3.5mm 有线' },
  { label: '降噪深度', value: '最高 40dB' },
  { label: '续航时间', value: '约 40 小时（降噪关）/ 30 小时（降噪开）' },
  { label: '充电时间', value: '约 2 小时（USB-C 快充）' },
  { label: '驱动单元', value: '40mm 镀钛振膜' },
  { label: '频率响应', value: '20Hz - 40kHz' },
  { label: '阻抗', value: '32Ω' },
  { label: '蓝牙版本', value: '5.3' },
  { label: '音频编码', value: 'LDAC, AAC, SBC' },
  { label: '重量', value: '约 198g' }
]

// ---- 评分分布 ----
const ratingDistribution = [
  { star: 5, percent: 72 },
  { star: 4, percent: 18 },
  { star: 3, percent: 6 },
  { star: 2, percent: 3 },
  { star: 1, percent: 1 }
]

// ---- 评价数据 ----
const reviews = [
  {
    name: '张***明',
    stars: '★★★★★',
    date: '2026-06-28',
    avatarColor: 'var(--shop-primary)',
    content: '降噪效果非常好，在地铁上完全听不到外界噪音。音质也很棒，低音有力不轰头，高音清晰。佩戴舒适，连续戴了 3 个小时耳朵也不觉得难受。强烈推荐！',
    images: ['/images/product-3.jpg', '/images/product-4.jpg']
  },
  {
    name: '李***婷',
    stars: '★★★★☆',
    date: '2026-06-15',
    avatarColor: 'var(--shop-success)',
    content: '颜值很高，白色款很百搭。降噪效果不错，在办公室使用基本听不到同事的说话声。续航也很满意，一周充一次电就够了。扣一星是因为折叠处稍微有点紧。',
    images: ['/images/product-2.jpg']
  },
  {
    name: '王***飞',
    stars: '★★★★★',
    date: '2026-06-02',
    avatarColor: 'var(--shop-warning)',
    content: '这是买过的第三款降噪耳机了，这款性价比最高。音质在这个价位段无可挑剔，LDAC 连接稳定。快充功能很实用，临时出门充 10 分钟就能用 3 小时。好评！',
    images: []
  }
]

// ---- 相关推荐 ----
const relatedProducts = [
  { id: 2, name: '北欧简约护眼台灯 LED', description: '三档色温调节，无频闪，智能记忆亮度', price: 159, originalPrice: 199, image: '/images/product-2.jpg', sales: 2586, rating: 5, tags: ['hot'] },
  { id: 3, name: '智能运动手表 S3', description: '血氧监测，GPS 定位，7 天续航', price: 899, image: '/images/product-3.jpg', sales: 1852, rating: 4.5, tags: [] },
  { id: 4, name: '便携蓝牙音箱 Mini', description: 'IPX7 防水，20 小时续航，TWS 串联', price: 349, originalPrice: 429, image: '/images/product-4.jpg', sales: 3421, rating: 5, tags: ['hot'] },
  { id: 5, name: '机械键盘 K1 青轴', description: '全键热插拔，RGB 背光，PBT 键帽', price: 459, originalPrice: 559, image: '/images/product-5.jpg', sales: 1206, rating: 4.5, tags: [] }
]
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