/**
 * 回到顶部按钮
 */
<template>
  <button
    v-show="visible"
    class="back-to-top"
    @click="scrollToTop"
    aria-label="回到顶部"
  >
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="18 15 12 9 6 15"/>
    </svg>
  </button>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const visible = ref(false)
let scrollHandler: (() => void) | null = null

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  scrollHandler = () => {
    visible.value = window.scrollY > 400
  }
  window.addEventListener('scroll', scrollHandler)
})

onUnmounted(() => {
  if (scrollHandler) {
    window.removeEventListener('scroll', scrollHandler)
  }
})
</script>

<style scoped>
.back-to-top {
  position: fixed;
  bottom: 32px;
  right: 32px;
  z-index: 40;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--shop-primary);
  color: white;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s, opacity 0.3s;
  cursor: pointer;
}
.back-to-top:hover {
  transform: scale(1.1);
}
</style>
