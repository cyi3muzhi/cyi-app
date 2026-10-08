/**
 * 默认布局 - 包含 Header 和 Footer
 */
<template>
  <div class="layout-root">
    <AppHeader />
    <main class="main-content">
      <router-view />
    </main>
    <AppFooter />
    <BackToTop />
    <CustomerServiceFloat />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useShopStore } from '@/store/shop'
import { useUserStore } from '@/store/user'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import BackToTop from '@/components/BackToTop.vue'
import CustomerServiceFloat from '@/components/CustomerServiceFloat.vue'

const route = useRoute()
const shopStore = useShopStore()
const userStore = useUserStore()

// 应用启动时加载店铺设置；若有 token 则拉取用户信息（恢复昵称/头像）
onMounted(() => {
  shopStore.fetchSettings()
  if (userStore.isLoggedIn) {
    userStore.fetchUserInfo()
  }
})

// 当店铺名称加载后更新 document title
watch(() => shopStore.shopName, (name) => {
  const title = route.meta.title
  document.title = title ? `${title} - ${name}` : name
})
</script>

<style scoped>
/* 根容器：flex 列布局，min-height 撑满视口。
   内容不足一屏时 footer 吸附底部，超出时 footer 自然跟随内容下移 */
.layout-root {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  min-height: 100dvh; /* 移动端地址栏动态高度更准确 */
}

.main-content {
  flex: 1; /* 占满剩余空间，将 footer 推到底部 */
  padding-top: 72px; /* header 高度（fixed 定位，占位由 padding 补偿） */
}
</style>
