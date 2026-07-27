/**
 * 默认布局 - 包含 Header 和 Footer
 */
<template>
  <AppHeader />
  <main class="main-content">
    <router-view />
  </main>
  <AppFooter />
  <BackToTop />
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useShopStore } from '@/store/shop'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import BackToTop from '@/components/BackToTop.vue'

const route = useRoute()
const shopStore = useShopStore()

// 应用启动时加载店铺设置
onMounted(() => {
  shopStore.fetchSettings()
})

// 当店铺名称加载后更新 document title
watch(() => shopStore.shopName, (name) => {
  const title = route.meta.title
  document.title = title ? `${title} - ${name}` : name
})
</script>

<style scoped>
.main-content {
  min-height: calc(100vh - 72px - 380px);
  padding-top: 72px;
}
</style>
