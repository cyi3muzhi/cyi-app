/**
 * 顶部导航栏
 */
<template>
  <header class="site-header">
    <div class="container-shop header-inner">
      <!-- Logo -->
      <router-link to="/" class="logo no-underline">
        <span class="logo-text">{{ shopStore.shopName }}</span>
      </router-link>

      <!-- 搜索栏（桌面） -->
      <div class="search-desktop">
        <div class="search-wrap">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索商品..."
            class="search-input"
            @keyup.enter="onSearch"
          />
        </div>
      </div>

      <!-- 导航链接 -->
      <nav class="nav-links">
        <router-link to="/" class="nav-item" exact-active-class="nav-active">首页</router-link>
        <router-link to="/products" class="nav-item" active-class="nav-active">全部商品</router-link>
        <router-link to="/cart" class="nav-item cart-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
            <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
          </svg>
          <span>购物车</span>
          <span v-if="cartCount > 0" class="cart-badge">{{ cartCount }}</span>
        </router-link>
        <router-link v-if="!isLoggedIn" to="/login" class="nav-item nav-btn">登录</router-link>
        <router-link v-else to="/user" class="nav-item user-avatar" active-class="nav-active" :title="userStore.nickname || '个人中心'">
          <img
            v-if="userStore.avatar && !avatarFailed"
            :src="userStore.avatar"
            alt="用户头像"
            class="avatar-img"
            @error="avatarFailed = true"
          />
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
          </svg>
        </router-link>

        <!-- 移动端搜索触发 -->
        <button class="mobile-search-btn" @click="showMobileSearch = !showMobileSearch" aria-label="搜索">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
        </button>
      </nav>
    </div>

    <!-- 移动端搜索栏 -->
    <div v-show="showMobileSearch" class="mobile-search-panel">
      <div class="container-shop">
        <div class="search-wrap">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索商品..."
            class="search-input"
            @keyup.enter="onSearch"
          />
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/store/cart'
import { useUserStore } from '@/store/user'
import { useShopStore } from '@/store/shop'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()
const shopStore = useShopStore()

const keyword = ref('')
const showMobileSearch = ref(false)

/** 头像加载失败标记（回退为默认人像图标） */
const avatarFailed = ref(false)

const cartCount = computed(() => cartStore.totalCount)
const isLoggedIn = computed(() => userStore.isLoggedIn)

// 挂载时加载购物车数量（登录读后端，未登录读本地）
onMounted(() => {
  cartStore.loadCart()
})

function onSearch() {
  if (keyword.value.trim()) {
    router.push({ path: '/products', query: { keyword: keyword.value.trim() } })
    keyword.value = ''
    showMobileSearch.value = false
  }
}
</script>

<style scoped>
.site-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  height: 72px;
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.logo-text {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-primary);
  letter-spacing: -0.02em;
}

.search-desktop {
  display: none;
  flex: 1;
  max-width: 420px;
  margin: 0 24px;
}
@media (min-width: 768px) {
  .search-desktop { display: flex; }
  .mobile-search-btn { display: none; }
}

.search-wrap {
  position: relative;
  width: 100%;
}
.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--shop-muted-foreground);
  pointer-events: none;
}
.search-input {
  width: 100%;
  height: 40px;
  padding: 0 16px 0 40px;
  border-radius: 999px;
  border: 1px solid var(--shop-border);
  background: var(--shop-muted);
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.search-input:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px var(--shop-primary-bg);
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 16px;
}
@media (max-width: 767px) {
  .nav-links {
    gap: 8px;
  }
}

.nav-item {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
  white-space: nowrap;
}
.nav-item:hover,
.nav-active {
  color: var(--shop-primary);
}

.cart-link {
  position: relative;
  display: flex;
  align-items: center;
  gap: 4px;
}
.cart-badge {
  position: absolute;
  top: -8px;
  right: -12px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--shop-destructive);
  color: #fff;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
}

.nav-btn {
  padding: 6px 18px;
  border-radius: var(--shop-radius-md);
  background: var(--shop-primary);
  color: var(--shop-primary-foreground) !important;
  transition: opacity 0.2s;
}
.nav-btn:hover {
  opacity: 0.9;
}

.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 2px solid var(--shop-border);
  transition: border-color 0.2s, color 0.2s;
  overflow: hidden;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.user-avatar:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

.mobile-search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  background: none;
  border: none;
  color: var(--shop-muted-foreground);
}
@media (min-width: 768px) {
  .mobile-search-btn {
    display: none;
  }
}

.mobile-search-panel {
  padding: 0 0 12px;
  background: var(--shop-background);
  border-top: 1px solid var(--shop-border);
}
</style>
