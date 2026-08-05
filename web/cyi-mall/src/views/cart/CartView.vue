<template>
  <div class="cart-page">
    <!-- 页面标题和步骤指示器（一级页面，无需返回按钮） -->
    <div class="container-shop">
      <h1 class="page-title">
        我的购物车 (<span class="cart-count">{{ cartStore.totalCount }}</span> 件商品)
      </h1>

      <!-- 步骤指示器 -->
      <div class="step-indicator">
        <div class="step-item">
          <div class="step-circle active">1</div>
          <span class="step-label active">我的购物车</span>
        </div>
        <div class="step-connector"></div>
        <div class="step-item">
          <div class="step-circle pending">2</div>
          <span class="step-label pending">确认订单</span>
        </div>
        <div class="step-connector"></div>
        <div class="step-item">
          <div class="step-circle pending">3</div>
          <span class="step-label pending">完成支付</span>
        </div>
      </div>
    </div>

    <!-- 购物车主内容 -->
    <div class="container-shop pb-8">
      <!-- 空购物车状态 -->
      <div v-if="cartStore.items.length === 0" class="empty-cart">
        <div class="empty-cart-icon">
          <svg width="96" height="96" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 100 4 2 2 0 000-4z" />
          </svg>
        </div>
        <p class="empty-title">购物车还是空的</p>
        <p class="empty-desc">快去挑选心仪的商品吧</p>
        <router-link to="/products" class="go-shopping-btn">去逛逛</router-link>
      </div>

      <!-- 购物车内容区（两栏布局） -->
      <div v-else class="cart-grid">
        <!-- 左侧：商品列表 -->
        <div class="cart-items-col">
          <div class="cart-list-card">
            <!-- 全选行 -->
            <div class="select-all-row">
              <label class="checkbox-label">
                <input
                  type="checkbox"
                  class="cart-checkbox"
                  :checked="cartStore.isAllSelected"
                  :indeterminate="isIndeterminate"
                  @change="toggleSelectAll"
                />
                全选
              </label>
              <span class="selected-count-label">
                已选 <strong>{{ cartStore.selectedCount }}</strong> 件
              </span>
              <button class="clear-cart-btn" @click="onClearCart">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                清空
              </button>
            </div>

            <!-- 商品条目列表 -->
            <div
              v-for="item in cartStore.items"
              :key="item.productId + '-' + item.skuId"
              class="cart-item-row"
            >
              <!-- 选中框 -->
              <label class="checkbox-col">
                <input
                  type="checkbox"
                  class="cart-checkbox"
                  :checked="item.selected"
                  @change="cartStore.toggleSelect(item.productId, item.skuId)"
                />
              </label>

              <!-- 缩略图 -->
              <div class="item-image-col">
                <div class="item-image-wrap">
                  <img :src="item.image" :alt="item.name" @error="onImageError" />
                </div>
              </div>

              <!-- 商品信息 -->
              <div class="item-info-col">
                <h3 class="item-name">{{ item.name }}</h3>
                <p class="item-sku">{{ item.sku }}</p>
                <div class="item-actions">
                  <span class="item-price">&yen;{{ item.price.toFixed(2) }}</span>
                  <!-- 数量调整器 -->
                  <div class="qty-adjuster">
                    <button
                      class="qty-btn"
                      :disabled="item.quantity <= 1"
                      @click="cartStore.updateQuantity(item.productId, item.skuId, item.quantity - 1)"
                    >
                      -
                    </button>
                    <span class="qty-value">{{ item.quantity }}</span>
                    <button
                      class="qty-btn"
                      :disabled="item.quantity >= item.stock"
                      @click="cartStore.updateQuantity(item.productId, item.skuId, item.quantity + 1)"
                    >
                      +
                    </button>
                  </div>
                  <span class="item-subtotal">&yen;{{ (item.price * item.quantity).toFixed(2) }}</span>
                  <button class="delete-btn" @click="onRemoveItem(item.productId, item.skuId)" aria-label="删除商品">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：订单摘要 -->
        <div class="cart-summary-col">
          <div class="summary-card">
            <h2 class="summary-title">订单摘要</h2>

            <!-- 已选商品数量 -->
            <div class="summary-row">
              <span class="summary-label">已选商品</span>
              <span class="summary-value">{{ cartStore.selectedCount }} 件</span>
            </div>

            <!-- 商品小计 -->
            <div class="summary-row">
              <span class="summary-label">商品小计</span>
              <span class="summary-value">&yen;{{ cartStore.subtotal.toFixed(2) }}</span>
            </div>

            <!-- 运费 -->
            <div class="summary-row">
              <span class="summary-label">运费</span>
              <span
                class="summary-value"
                :class="shipping === 0 ? 'text-success' : ''"
              >
                {{ shipping === 0 ? '免运费' : '¥' + shipping.toFixed(2) }}
              </span>
            </div>

            <!-- 优惠码 -->
            <div class="promo-section">
              <label class="promo-label">优惠码</label>
              <div class="promo-input-row">
                <input
                  type="text"
                  v-model="promoCode"
                  placeholder="输入优惠码"
                  class="promo-input"
                />
                <button class="promo-apply-btn" @click="applyPromo">使用</button>
              </div>
              <p v-if="promoMessage" class="promo-message" :class="promoMessageType">
                {{ promoMessage }}
              </p>
            </div>

            <!-- 合计 -->
            <div class="total-row">
              <span class="total-label">合计</span>
              <span class="total-price">&yen;{{ totalAmount.toFixed(2) }}</span>
            </div>

            <!-- 去结算按钮 -->
            <button
              class="checkout-btn"
              :disabled="cartStore.selectedCount === 0"
              @click="onCheckout"
            >
              去结算
            </button>

            <!-- 继续购物 -->
            <div class="continue-shopping">
              <router-link to="/products" class="continue-link">
                &larr; 继续购物
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 猜你喜欢推荐区 -->
    <section class="container-shop pb-12" v-if="cartStore.items.length > 0">
      <div class="recommend-header">
        <div class="recommend-bar"></div>
        <h2 class="recommend-title">猜你喜欢</h2>
      </div>
      <div class="recommend-grid">
        <ProductCard v-for="item in recommendProducts" :key="item.id" :product="item" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
/**
 * 购物车视图组件
 * 展示购物车商品列表、选中管理、数量调整、订单结算和推荐商品
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/store/cart'
import ProductCard from '@/components/ProductCard.vue'

const router = useRouter()
const cartStore = useCartStore()

// 页面加载时拉取购物车（登录读后端，未登录读本地）
onMounted(() => {
  cartStore.loadCart()
})

// ---- 优惠码 ----
const promoCode = ref('')
const promoMessage = ref('')
const promoMessageType = ref('')
const promoApplied = ref(false)

/** 应用优惠码 */
function applyPromo() {
  const code = promoCode.value.trim()
  if (!code) {
    promoMessage.value = '请输入优惠码'
    promoMessageType.value = 'error'
    return
  }
  if (code.toUpperCase() === 'WELCOME10') {
    promoMessage.value = '优惠码已生效，已享受 9 折优惠！'
    promoMessageType.value = 'success'
    promoApplied.value = true
  } else {
    promoMessage.value = '无效的优惠码，请重新输入'
    promoMessageType.value = 'error'
    promoApplied.value = false
  }
}

// ---- 全选逻辑 ----
/** 是否处于半选状态 */
const isIndeterminate = computed(() => {
  const checked = cartStore.items.filter(i => i.selected).length
  return checked > 0 && checked < cartStore.items.length
})

/** 切换全选 */
function toggleSelectAll() {
  cartStore.toggleSelectAll(!cartStore.isAllSelected)
}

/**
 * 删除购物车商品（二次确认）
 * @param productId 商品ID
 * @param skuId SKU ID
 */
function onRemoveItem(productId: number, skuId: number) {
  if (!window.confirm('确定将该商品移出购物车吗？')) return
  cartStore.removeItem(productId, skuId)
}

/**
 * 清空购物车（二次确认）
 */
function onClearCart() {
  if (cartStore.items.length === 0) return
  if (!window.confirm('确定清空购物车吗？')) return
  cartStore.clear()
}

/** 商品图片加载失败时的占位处理 */
function onImageError(e: Event) {
  ;(e.target as HTMLImageElement).style.display = 'none'
}

/**
 * 去结算：跳转确认订单页（携带购物车选中的商品，由结算页读取 store）
 */
function onCheckout() {
  if (cartStore.selectedCount === 0) return
  router.push('/checkout')
}

// ---- 运费和合计 ----
/** 运费：满 199 免运费 */
const shipping = computed(() => {
  if (cartStore.selectedCount > 0 && cartStore.subtotal < 199) {
    return 10
  }
  return 0
})

/** 合计金额 */
const totalAmount = computed(() => {
  let total = cartStore.subtotal + shipping.value
  if (promoApplied.value) {
    total *= 0.9
  }
  return total
})

// ---- 推荐商品 ----
const recommendProducts = [
  { id: 4, name: '便携蓝牙音箱 Mini', description: 'IPX7 防水，20 小时续航，TWS 串联', price: 349, originalPrice: 429, image: '/images/product-4.jpg', sales: 3421, rating: 5, tags: ['hot'] },
  { id: 5, name: '机械键盘 K1 青轴', description: '全键热插拔，RGB 背光，PBT 键帽', price: 459, originalPrice: 559, image: '/images/product-5.jpg', sales: 1206, rating: 4.5, tags: [] },
  { id: 6, name: '智能保温杯温显款', description: '304不锈钢，LED 实时显示水温', price: 199, originalPrice: 249, image: '/images/product-6.jpg', sales: 4521, rating: 4.5, tags: ['hot'] },
  { id: 7, name: '便携式移动电源 20000mAh', description: '大容量快充，轻薄便携，多协议兼容', price: 129, image: '/images/product-1.jpg', sales: 5632, rating: 4, tags: [] }
]
</script>

<style scoped>

/* ---- 页面标题 ---- */
.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin-bottom: 16px;
}
.cart-count {
  color: var(--shop-foreground);
}

/* ---- 步骤指示器 ---- */
.step-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 24px;
  overflow-x: auto;
  padding: 8px 0;
}
@media (min-width: 768px) {
  .step-indicator {
    justify-content: flex-start;
  }
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.step-circle {
  width: 32px;
  height: 32px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  font-weight: 600;
  flex-shrink: 0;
}
.step-circle.active {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}
.step-circle.done {
  background: var(--shop-success);
  color: var(--shop-success-foreground);
}
.step-circle.pending {
  background: var(--shop-muted);
  color: var(--shop-muted-foreground);
}

.step-label {
  font-size: 0.875rem;
  white-space: nowrap;
}
.step-label.active {
  color: var(--shop-primary);
  font-weight: 600;
}
.step-label.pending {
  color: var(--shop-muted-foreground);
}
.step-label.done {
  color: var(--shop-foreground);
}

.step-connector {
  width: 48px;
  height: 1px;
  background: var(--shop-border);
  margin: 0 8px;
}

/* ---- 空购物车 ---- */
.empty-cart {
  text-align: center;
  padding: 80px 0;
}
.empty-cart-icon {
  margin-bottom: 24px;
}
.empty-cart-icon svg {
  width: 96px;
  height: 96px;
  color: var(--shop-border);
}
.empty-title {
  font-size: 1.125rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 8px;
}
.empty-desc {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  opacity: 0.7;
  margin-bottom: 24px;
}
.go-shopping-btn {
  display: inline-block;
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  padding: 10px 32px;
  border-radius: var(--shop-radius-lg);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: background var(--shop-transition);
  box-shadow: var(--shop-shadow-sm);
}
.go-shopping-btn:hover {
  background: var(--shop-primary-dark);
}

/* ---- 购物车主布局 ---- */
.cart-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
}
@media (min-width: 1024px) {
  .cart-grid {
    grid-template-columns: 2fr 1fr;
  }
}

/* ---- 购物车列表卡片 ---- */
.cart-list-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
  overflow: hidden;
}

/* ---- 全选行 ---- */
.select-all-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--shop-border);
  background: var(--shop-muted);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
  cursor: pointer;
  user-select: none;
}

.selected-count-label {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-left: auto;
}

.clear-cart-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: none;
  background: none;
  color: var(--shop-muted-foreground);
  font-size: 0.75rem;
  cursor: pointer;
  transition: color var(--shop-transition);
  font-family: var(--shop-font-sans);
}
.clear-cart-btn:hover {
  color: var(--shop-destructive);
}

/* ---- 商品条目 ---- */
.cart-item-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px;
  border-bottom: 1px solid var(--shop-border);
}
.cart-item-row:last-child {
  border-bottom: none;
}

.checkbox-col {
  display: flex;
  align-items: center;
  padding-top: 12px;
  cursor: pointer;
}

.cart-checkbox {
  width: 18px;
  height: 18px;
  accent-color: var(--shop-primary);
  cursor: pointer;
}

.item-image-col {
  flex-shrink: 0;
}

.item-image-wrap {
  width: 80px;
  height: 80px;
  border-radius: var(--shop-radius-lg);
  overflow: hidden;
  background: var(--shop-muted);
}
@media (min-width: 640px) {
  .item-image-wrap {
    width: 96px;
    height: 96px;
  }
}
.item-image-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.item-info-col {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.item-sku {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-top: 4px;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.item-price {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

/* ---- 数量调整器 ---- */
.qty-adjuster {
  display: flex;
  align-items: center;
}

.qty-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--shop-border);
  background: var(--shop-background);
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  transition: background var(--shop-transition), border-color var(--shop-transition);
  user-select: none;
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
}
.qty-btn:first-child {
  border-radius: var(--shop-radius-md) 0 0 var(--shop-radius-md);
}
.qty-btn:last-child {
  border-radius: 0 var(--shop-radius-md) var(--shop-radius-md) 0;
}
.qty-btn:hover {
  background: var(--shop-muted);
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}
.qty-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  border-color: var(--shop-border);
  color: var(--shop-muted-foreground);
  background: var(--shop-muted);
}

.qty-value {
  width: 40px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--shop-border);
  border-bottom: 1px solid var(--shop-border);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

.item-subtotal {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--shop-destructive);
  margin-left: auto;
}

.delete-btn {
  color: var(--shop-muted-foreground);
  transition: color var(--shop-transition);
  cursor: pointer;
  background: none;
  border: none;
  padding: 4px;
  flex-shrink: 0;
}
.delete-btn:hover {
  color: var(--shop-destructive);
}

/* ---- 订单摘要 ---- */
.summary-card {
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-sm);
  padding: 20px;
  position: sticky;
  top: 96px;
}

.summary-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--shop-foreground);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--shop-border);
  margin-bottom: 16px;
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 0.875rem;
}

.summary-label {
  color: var(--shop-muted-foreground);
}

.summary-value {
  font-weight: 500;
  color: var(--shop-foreground);
}

.text-success {
  color: var(--shop-success) !important;
}

/* ---- 优惠码 ---- */
.promo-section {
  padding-top: 12px;
  border-top: 1px solid var(--shop-border);
  margin-bottom: 16px;
}

.promo-label {
  display: block;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  margin-bottom: 6px;
}

.promo-input-row {
  display: flex;
  gap: 8px;
}

.promo-input {
  flex: 1;
  height: 36px;
  padding: 0 12px;
  font-size: 0.875rem;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-input);
  background: var(--shop-background);
  outline: none;
  transition: ring var(--shop-transition);
  font-family: var(--shop-font-sans);
  color: var(--shop-foreground);
}
.promo-input:focus {
  ring: 2px solid var(--shop-ring);
  border-color: var(--shop-ring);
}

.promo-apply-btn {
  height: 36px;
  padding: 0 16px;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  border: none;
  cursor: pointer;
  transition: background var(--shop-transition);
  box-shadow: var(--shop-shadow-sm);
  font-family: var(--shop-font-sans);
}
.promo-apply-btn:hover {
  background: var(--shop-primary-dark);
}

.promo-message {
  font-size: 0.75rem;
  margin-top: 4px;
}
.promo-message.success {
  color: var(--shop-success);
}
.promo-message.error {
  color: var(--shop-destructive);
}

/* ---- 合计 ---- */
.total-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 12px;
  border-top: 1px solid var(--shop-border);
  margin-bottom: 16px;
}

.total-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

.total-price {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-destructive);
}

/* ---- 去结算 ---- */
.checkout-btn {
  width: 100%;
  height: 44px;
  font-size: 0.875rem;
  font-weight: 600;
  border: none;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  cursor: pointer;
  transition: background var(--shop-transition);
  box-shadow: var(--shop-shadow-sm);
  margin-bottom: 12px;
  font-family: var(--shop-font-sans);
}
.checkout-btn:hover {
  background: var(--shop-primary-dark);
}
.checkout-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ---- 继续购物 ---- */
.continue-shopping {
  text-align: center;
}
.continue-link {
  font-size: 0.875rem;
  color: var(--shop-primary);
  text-decoration: none;
  transition: color var(--shop-transition);
}
.continue-link:hover {
  color: var(--shop-primary-dark);
}

/* ---- 猜你喜欢 ---- */
.recommend-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.recommend-bar {
  width: 4px;
  height: 20px;
  background: var(--shop-primary);
  border-radius: 2px;
}

.recommend-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.recommend-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
@media (min-width: 768px) {
  .recommend-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
  }
}
</style>