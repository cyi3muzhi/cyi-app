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
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import BackToTop from '@/components/BackToTop.vue'
</script>

<style scoped>
.main-content {
  min-height: calc(100vh - 72px - 380px);
  padding-top: 72px;
}
</style>
