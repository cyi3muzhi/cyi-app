<template>
  <!-- 首页视图：Hero Banner + 商品分类 + 精选推荐 + 新品上市 -->
  <div class="home-page">
    <!-- ============================================================ -->
    <!-- Hero Banner -->
    <!-- ============================================================ -->
    <section class="hero-banner">
      <div class="container-shop hero-content">
        <div class="hero-text">
          <h1 class="hero-title">夏日焕新 品质生活</h1>
          <p class="hero-subtitle">精选好物 限时特惠</p>
          <router-link to="/products" class="hero-btn">立即抢购</router-link>
        </div>
        <div class="hero-image">
          <div class="hero-placeholder">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary-light)" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            <span>Hero Image</span>
          </div>
        </div>
      </div>
      <!-- 底部波浪装饰 -->
      <div class="hero-wave"></div>
    </section>

    <!-- ============================================================ -->
    <!-- 商品分类 -->
    <!-- ============================================================ -->
    <section class="categories-section">
      <div class="container-shop">
        <h2 class="section-title">商品分类</h2>
        <p class="section-desc">一站式选购，满足你的生活所需</p>
        <div class="categories-grid">
          <router-link
            v-for="cat in categories"
            :key="cat.key"
            :to="{ path: '/products', query: { category: cat.key } }"
            class="category-card"
          >
            <div class="category-icon" v-html="cat.icon"></div>
            <span class="category-name">{{ cat.name }}</span>
          </router-link>
        </div>
      </div>
    </section>

    <!-- ============================================================ -->
    <!-- 精选推荐 -->
    <!-- ============================================================ -->
    <section class="featured-section">
      <div class="container-shop">
        <div class="section-header">
          <h2 class="section-title">精选推荐</h2>
          <router-link to="/products" class="section-more">
            查看更多
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
          </router-link>
        </div>
        <div class="product-grid">
          <ProductCard v-for="item in featuredProducts" :key="item.id" :product="item" />
        </div>
      </div>
    </section>

    <!-- ============================================================ -->
    <!-- 促销横幅 -->
    <!-- ============================================================ -->
    <section class="promo-banner">
      <div class="container-shop promo-content">
        <span class="promo-tag">限时活动</span>
        <h2 class="promo-title">全场满299减50 · 新人专享礼包</h2>
        <p class="promo-desc">活动时间：2026.07.15 - 2026.07.31 | 部分商品除外 | 最终解释权归悦选商城所有</p>
        <router-link to="/products" class="promo-btn">立即选购</router-link>
      </div>
    </section>

    <!-- ============================================================ -->
    <!-- 新品上市 -->
    <!-- ============================================================ -->
    <section class="new-arrivals-section">
      <div class="container-shop">
        <div class="section-header">
          <h2 class="section-title">新品上市</h2>
          <router-link to="/products" class="section-more">
            查看更多
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"/>
            </svg>
          </router-link>
        </div>
        <div class="product-grid">
          <ProductCard v-for="item in newProducts" :key="item.id" :product="item" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import ProductCard from '@/components/ProductCard.vue'

/** 商品分类图标（SVG 片段） */
interface Category {
  key: string
  name: string
  icon: string
}

const categories: Category[] = [
  {
    key: 'electronics',
    name: '电子产品',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <rect x="4" y="2" width="16" height="20" rx="2" ry="2"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="12" y1="18" x2="12" y2="18.01"/>
    </svg>`
  },
  {
    key: 'home',
    name: '家居生活',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
    </svg>`
  },
  {
    key: 'fashion',
    name: '时尚服饰',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M20.84 4.61a5.5 5.5 0 0 0-7.34 0L12 6l-1.5-1.39a5.5 5.5 0 0 0-7.34 0A5.5 5.5 0 0 0 2.07 11L12 21l9.93-10a5.5 5.5 0 0 0-1.09-6.39z"/>
    </svg>`
  },
  {
    key: 'sports',
    name: '运动户外',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="6" cy="6" r="3"/><circle cx="18" cy="18" r="3"/><path d="M18 6l-8 12"/>
    </svg>`
  },
  {
    key: 'beauty',
    name: '美妆护肤',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 3a9 9 0 0 0-9 9c0 4.97 4.03 9 9 9s9-4.03 9-9a9 9 0 0 0-9-9z"/><path d="M8 12a4 4 0 1 0 8 0"/>
    </svg>`
  },
  {
    key: 'books',
    name: '图书文具',
    icon: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--shop-primary)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
    </svg>`
  }
]

/** 商品 Mock 数据类型 */
interface Product {
  id: number
  name: string
  description: string
  price: number
  originalPrice: number
  image: string
  sales: number
  rating: number
  tags?: string[]
  category?: string
}

/** 精选推荐（4个 Mock 商品） */
const featuredProducts: Product[] = [
  {
    id: 1,
    name: '无线降噪耳机',
    description: '旗舰级主动降噪，40小时续航，舒适佩戴',
    price: 299.00,
    originalPrice: 599.00,
    image: 'https://picsum.photos/seed/product1/400/400',
    sales: 2586,
    rating: 4.5,
    tags: ['hot']
  },
  {
    id: 2,
    name: '智能运动手表',
    description: '心率血氧监测，100+运动模式，14天续航',
    price: 459.00,
    originalPrice: 899.00,
    image: 'https://picsum.photos/seed/product2/400/400',
    sales: 5108,
    rating: 5.0,
    tags: ['new']
  },
  {
    id: 3,
    name: '简约双肩背包',
    description: '防水面料，多隔层收纳，电脑独立仓',
    price: 189.00,
    originalPrice: 369.00,
    image: 'https://picsum.photos/seed/product3/400/400',
    sales: 1847,
    rating: 4.8,
    tags: ['hot']
  },
  {
    id: 4,
    name: '轻量跑鞋',
    description: '透气飞织鞋面，缓震大底，轻盈回弹',
    price: 399.00,
    originalPrice: 699.00,
    image: 'https://picsum.photos/seed/product4/400/400',
    sales: 3621,
    rating: 4.8,
    tags: ['sale']
  }
]

/** 新品上市（4个 Mock 商品） */
const newProducts: Product[] = [
  {
    id: 5,
    name: '便携蓝牙音箱',
    description: 'IPX7防水，360°环绕立体声，12小时续航',
    price: 259.00,
    originalPrice: 499.00,
    image: 'https://picsum.photos/seed/product5/400/400',
    sales: 1523,
    rating: 4.6,
    tags: ['new']
  },
  {
    id: 6,
    name: '偏光太阳镜',
    description: 'UV400防护，TR90镜框，轻盈不压鼻',
    price: 169.00,
    originalPrice: 329.00,
    image: 'https://picsum.photos/seed/product6/400/400',
    sales: 2896,
    rating: 4.2,
    tags: ['new']
  },
  {
    id: 7,
    name: '智能体脂秤',
    description: '25项身体数据，APP同步，全家共用',
    price: 129.00,
    originalPrice: 259.00,
    image: 'https://picsum.photos/seed/product7/400/400',
    sales: 4312,
    rating: 4.4,
    tags: ['new']
  },
  {
    id: 8,
    name: '真丝睡衣套装',
    description: '100%桑蚕丝，亲肤顺滑，高级感满',
    price: 399.00,
    originalPrice: 799.00,
    image: 'https://picsum.photos/seed/product8/400/400',
    sales: 987,
    rating: 4.9,
    tags: ['new']
  }
]
</script>

<style scoped>
/* ============================================
   Hero Banner
   ============================================ */
.hero-banner {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, var(--shop-primary), hsl(217, 91%, 72%));
}

.hero-content {
  display: flex;
  flex-direction: column-reverse;
  align-items: center;
  gap: 32px;
  padding-top: 64px;
  padding-bottom: 64px;
}

@media (min-width: 1024px) {
  .hero-content {
    flex-direction: row;
    padding-top: 96px;
    padding-bottom: 96px;
  }
}

.hero-text {
  flex: 1;
  text-align: center;
}

@media (min-width: 1024px) {
  .hero-text {
    text-align: left;
  }
}

.hero-title {
  margin-bottom: 16px;
  font-size: 2.25rem;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.02em;
  color: #fff;
}

@media (min-width: 768px) {
  .hero-title {
    font-size: 3rem;
  }
}

@media (min-width: 1024px) {
  .hero-title {
    font-size: 3.75rem;
  }
}

.hero-subtitle {
  margin-bottom: 32px;
  font-size: 1.125rem;
  color: rgba(255, 255, 255, 0.85);
}

@media (min-width: 768px) {
  .hero-subtitle {
    font-size: 1.25rem;
  }
}

.hero-btn {
  display: inline-block;
  padding: 14px 40px;
  border-radius: 9999px;
  background: var(--shop-background);
  color: var(--shop-primary);
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  box-shadow: var(--shop-shadow-lg);
  transition: transform 0.2s;
}

.hero-btn:hover {
  transform: scale(1.05);
}

.hero-image {
  flex: 1;
  display: flex;
  justify-content: center;
}

@media (min-width: 1024px) {
  .hero-image {
    justify-content: flex-end;
  }
}

.hero-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  max-width: 320px;
  aspect-ratio: 4 / 3;
  border-radius: var(--shop-radius-xl);
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.875rem;
  box-shadow: var(--shop-shadow-xl);
}

@media (min-width: 1024px) {
  .hero-placeholder {
    max-width: 380px;
  }
}

/* 底部波浪装饰 */
.hero-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 48px;
  background: var(--shop-background);
  mask: url("data:image/svg+xml;utf8,<svg viewBox='0 0 1440 48' xmlns='http://www.w3.org/2000/svg'><path d='M0 48h1440V0c-240 32-480 48-720 48S240 32 0 0v48z' fill='white'/></svg>") center / cover no-repeat;
  -webkit-mask: url("data:image/svg+xml;utf8,<svg viewBox='0 0 1440 48' xmlns='http://www.w3.org/2000/svg'><path d='M0 48h1440V0c-240 32-480 48-720 48S240 32 0 0v48z' fill='white'/></svg>") center / cover no-repeat;
}

/* ============================================
   公共标题样式
   ============================================ */
.section-title {
  margin-bottom: 8px;
  font-size: 1.5rem;
  font-weight: 700;
  text-align: center;
  color: var(--shop-foreground);
}

.section-desc {
  margin-bottom: 40px;
  font-size: 0.875rem;
  text-align: center;
  color: var(--shop-muted-foreground);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 32px;
}

.section-header .section-title {
  margin-bottom: 0;
  text-align: left;
}

.section-more {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-primary);
  text-decoration: none;
  transition: opacity 0.2s;
}

.section-more:hover {
  opacity: 0.8;
}

/* ============================================
   商品分类
   ============================================ */
.categories-section {
  padding: 64px 0;
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

@media (min-width: 640px) {
  .categories-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .categories-grid {
    grid-template-columns: repeat(6, 1fr);
  }
}

.category-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 16px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  background: var(--shop-card);
  color: var(--shop-foreground);
  text-decoration: none;
  text-align: center;
  transition: transform 0.2s, box-shadow 0.2s;
}

.category-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shop-shadow-md);
}

.category-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--shop-primary-bg);
}

.category-icon :deep(svg) {
  display: block;
}

.category-name {
  font-size: 0.875rem;
  font-weight: 500;
}

/* ============================================
   精选推荐
   ============================================ */
.featured-section {
  padding: 64px 0;
  background: var(--shop-muted);
}

/* ============================================
   商品网格
   ============================================ */
.product-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
}

@media (min-width: 640px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .product-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* ============================================
   促销横幅
   ============================================ */
.promo-banner {
  position: relative;
  overflow: hidden;
  padding: 80px 0;
  background: linear-gradient(135deg, hsl(217, 91%, 96%), hsl(217, 91%, 88%));
}

.promo-content {
  text-align: center;
}

.promo-tag {
  display: inline-block;
  margin-bottom: 16px;
  padding: 4px 16px;
  border-radius: 9999px;
  background: var(--shop-primary);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.promo-title {
  margin-bottom: 16px;
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

@media (min-width: 768px) {
  .promo-title {
    font-size: 2.25rem;
  }
}

.promo-desc {
  margin-bottom: 32px;
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.promo-btn {
  display: inline-block;
  padding: 12px 40px;
  border-radius: 9999px;
  background: var(--shop-primary);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  box-shadow: var(--shop-shadow-lg);
  transition: transform 0.2s;
}

.promo-btn:hover {
  transform: scale(1.05);
}

/* ============================================
   新品上市
   ============================================ */
.new-arrivals-section {
  padding: 64px 0;
}
</style>
