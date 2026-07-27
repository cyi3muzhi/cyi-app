/**
 * 商品卡片组件
 */
<template>
  <article class="product-card" :style="cardStyle">
    <router-link :to="`/product/${product.id}`" class="product-link">
      <div class="img-wrap">
        <span v-if="badge" class="badge" :class="badgeClass">{{ badge }}</span>
        <img :src="product.image" :alt="product.name" loading="lazy" />
      </div>
      <div class="info">
        <h3 class="name desc-line-clamp">{{ product.name }}</h3>
        <p v-if="product.description" class="desc desc-line-clamp">{{ product.description }}</p>
        <div class="price-row">
          <span class="price">¥{{ product.price.toFixed(2) }}</span>
          <span v-if="product.originalPrice > product.price" class="original-price">¥{{ product.originalPrice.toFixed(2) }}</span>
        </div>
        <div class="meta">
          <span>已售 {{ formatSales(product.sales) }} 件</span>
          <span class="stars">{{ ratingStars }}</span>
        </div>
      </div>
    </router-link>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  product: {
    id: number
    name: string
    description?: string
    price: number
    originalPrice?: number
    image: string
    sales: number
    rating: number
    tags?: string[]
  }
}>()

const badge = computed(() => {
  if (!props.product.tags?.length) return ''
  const tag = props.product.tags[0]
  if (tag === 'new') return '新品'
  if (tag === 'hot') return '热卖'
  if (tag === 'sale') return '特惠'
  return ''
})

const badgeClass = computed(() => {
  const tag = props.product.tags?.[0]
  if (tag === 'new') return 'badge-new'
  if (tag === 'hot') return 'badge-hot'
  if (tag === 'sale') return 'badge-sale'
  return ''
})

const cardStyle = {
  background: 'var(--shop-card)',
  boxShadow: 'var(--shop-shadow-md)'
}

const ratingStars = computed(() => {
  const r = Math.round(props.product.rating)
  return '★'.repeat(r) + '☆'.repeat(5 - r)
})

function formatSales(n: number) {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return n.toLocaleString()
}
</script>

<style scoped>
.product-card {
  border-radius: var(--shop-radius-xl);
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}
.product-card:hover {
  transform: translateY(-6px) scale(1.02);
  box-shadow: var(--shop-shadow-xl);
}

.product-link {
  display: block;
  text-decoration: none;
  color: inherit;
}

.img-wrap {
  position: relative;
  overflow: hidden;
  aspect-ratio: 1 / 1;
  background: var(--shop-muted);
}
.img-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}
.product-card:hover .img-wrap img {
  transform: scale(1.08);
}

.badge {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 2px 10px;
  border-radius: var(--shop-radius-sm);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.6;
  z-index: 2;
}
.badge-new { background: var(--shop-success); color: var(--shop-success-foreground); }
.badge-hot { background: var(--shop-warning); color: var(--shop-warning-foreground); }
.badge-sale { background: var(--shop-destructive); color: var(--shop-destructive-foreground); }

.info {
  padding: 16px;
}

.name {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--shop-card-foreground);
}

.desc {
  font-size: 0.75rem;
  margin-bottom: 8px;
  color: var(--shop-muted-foreground);
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
}

.price {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-primary);
}

.original-price {
  font-size: 0.75rem;
  text-decoration: line-through;
  color: var(--shop-muted-foreground);
}

.meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}
</style>
