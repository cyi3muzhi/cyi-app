<template>
  <!-- 统一返回按钮：返回上一页，无历史时兜底跳转；light 模式用于彩色卡片上 -->
  <button class="back-link" :class="{ light }" title="返回上一页" @click="goBack(fallback)">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="15 18 9 12 15 6"/>
    </svg>
  </button>
</template>

<script setup lang="ts">
import { goBack } from '@/utils/goBack'

/** 无历史记录时的兜底跳转地址；light 表示浅色按钮（用于渐变卡片上） */
const props = withDefaults(defineProps<{ fallback?: string; light?: boolean }>(), {
  fallback: '/',
  light: false
})
</script>

<style scoped>
.back-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-foreground);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.2s, transform 0.2s;
}
.back-link:hover {
  background: var(--shop-border);
  transform: translateX(-2px);
}
/* 浅色模式：半透明白底，用于渐变/深色卡片上的返回按钮 */
.back-link.light {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.35);
  backdrop-filter: blur(4px);
}
.back-link.light:hover {
  background: rgba(255, 255, 255, 0.4);
}
</style>
