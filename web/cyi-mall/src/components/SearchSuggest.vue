/**
 * 搜索联想下拉面板组件
 * 三段式布局：联想区（keyword≥2 时）+ 历史区 + 热词区
 * 支持键盘导航（↓↑循环、Enter选中、Esc关闭）和点击外部关闭
 */
<template>
  <transition name="suggest-fade">
    <div v-show="visible" class="search-suggest" @click.stop>
      <!-- 联想区：输入 ≥2 字符时展示 -->
      <div v-if="keyword.length >= 2" class="suggest-section">
        <div class="suggest-section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          联想商品
        </div>
        <!-- 联想加载中 -->
        <div v-if="suggestLoading" class="suggest-loading">搜索中...</div>
        <!-- 联想结果列表 -->
        <ul v-else-if="suggestItems.length > 0" class="suggest-list">
          <li
            v-for="(item, idx) in suggestItems"
            :key="'s-' + idx"
            :class="['suggest-item', { active: activeIndex === idx }]"
            @click="onSelect(item)"
            @mouseenter="activeIndex = idx"
          >
            <span class="suggest-item-text" v-html="highlightKeyword(item)"></span>
          </li>
        </ul>
        <!-- 无联想结果 -->
        <div v-else class="suggest-empty">无匹配商品</div>
      </div>

      <!-- 历史区：输入 <2 字符时展示 -->
      <div v-if="keyword.length < 2 && searchHistory.length > 0" class="suggest-section">
        <div class="suggest-section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
          搜索历史
          <span class="suggest-clear" @click="onClearHistory">清空</span>
        </div>
        <ul class="suggest-list">
          <li
            v-for="(item, idx) in searchHistory"
            :key="'h-' + idx"
            :class="['suggest-item', { active: activeIndex === suggestItems.length + idx }]"
            @click="onSelect(item)"
            @mouseenter="activeIndex = suggestItems.length + idx"
          >
            <span class="suggest-item-text">{{ item }}</span>
            <span class="suggest-item-delete" @click.stop="onRemoveHistory(item)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </span>
          </li>
        </ul>
      </div>

      <!-- 热词区：输入 <2 字符时展示 -->
      <div v-if="keyword.length < 2 && hotKeywords.length > 0" class="suggest-section">
        <div class="suggest-section-title">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"/>
          </svg>
          热门搜索
        </div>
        <div class="hot-keywords">
          <span
            v-for="(kw, idx) in hotKeywords"
            :key="'hot-' + idx"
            :class="['hot-keyword', { 'hot-top': idx < 3 }]"
            @click="onSelect(kw)"
          >
            <span class="hot-rank">{{ idx + 1 }}</span>
            {{ kw }}
          </span>
        </div>
      </div>

      <!-- 全空兜底 -->
      <div v-if="keyword.length < 2 && searchHistory.length === 0 && hotKeywords.length === 0" class="suggest-empty-all">
        输入关键词搜索商品
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { suggestProduct } from '@/api/product'
import { useShopStore } from '@/store/shop'
import { useUserStore } from '@/store/user'
import {
  getSearchHistory,
  addSearchHistory,
  removeSearchHistory,
  clearSearchHistory
} from '@/utils/searchHistory'

const props = defineProps<{
  /** 输入框当前值 */
  keyword: string
  /** 面板是否可见 */
  visible: boolean
}>()

const emit = defineEmits<{
  /** 选中某个关键词（联想/历史/热词） */
  (e: 'select', keyword: string): void
  /** 清空历史 */
  (e: 'clear-history'): void
  /** 移除单条历史 */
  (e: 'remove-history', keyword: string): void
}>()

const shopStore = useShopStore()
const userStore = useUserStore()

/** 联想结果列表 */
const suggestItems = ref<string[]>([])
/** 联想加载状态 */
const suggestLoading = ref(false)
/** 搜索历史列表 */
const searchHistory = ref<string[]>([])
/** 热门搜索词 */
const hotKeywords = ref<string[]>([])
/** 当前高亮索引（-1 表示无高亮） */
const activeIndex = ref(-1)

/** 防抖定时器 */
let debounceTimer: ReturnType<typeof setTimeout> | null = null

/** 从 store 读取热词 */
function loadHotKeywords() {
  hotKeywords.value = shopStore.hotKeywords || []
}

/** 从 localStorage 读取搜索历史 */
function loadSearchHistory() {
  searchHistory.value = getSearchHistory(userStore.userId)
}

/** 触发联想请求（防抖 300ms，最小 2 字符） */
function fetchSuggest() {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  const kw = props.keyword.trim()
  if (kw.length < 2) {
    suggestItems.value = []
    activeIndex.value = -1
    return
  }
  debounceTimer = setTimeout(async () => {
    suggestLoading.value = true
    try {
      const res = await suggestProduct({ keyword: kw, limit: 8 })
      suggestItems.value = res.items || []
    } catch {
      suggestItems.value = []
    } finally {
      suggestLoading.value = false
    }
  }, 300)
}

/**
 * 高亮联想词中匹配的关键词部分
 * 用 <mark> 标签包裹匹配字符
 */
function highlightKeyword(text: string): string {
  const kw = props.keyword.trim()
  if (!kw) return text
  // 转义正则特殊字符
  const escaped = kw.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return text.replace(new RegExp(`(${escaped})`, 'gi'), '<mark>$1</mark>')
}

/** 选中关键词 */
function onSelect(keyword: string) {
  if (keyword) {
    addSearchHistory(keyword, userStore.userId)
    emit('select', keyword)
  }
}

/** 清空搜索历史 */
function onClearHistory() {
  clearSearchHistory(userStore.userId)
  searchHistory.value = []
  emit('clear-history')
}

/** 删除单条搜索历史 */
function onRemoveHistory(keyword: string) {
  removeSearchHistory(keyword, userStore.userId)
  loadSearchHistory()
  emit('remove-history', keyword)
}

/**
 * 键盘导航：↓ 下移、↑ 上移、循环高亮
 * @returns 是否消费了该按键（用于阻止默认行为）
 */
function navigateKey(direction: 'down' | 'up') {
  const total = suggestItems.value.length + (props.keyword.length < 2 ? searchHistory.value.length : 0)
  if (total === 0) return
  if (direction === 'down') {
    activeIndex.value = activeIndex.value < total - 1 ? activeIndex.value + 1 : 0
  } else {
    activeIndex.value = activeIndex.value > 0 ? activeIndex.value - 1 : total - 1
  }
}

/**
 * 获取当前高亮项的关键词
 * @returns 高亮项关键词或 null
 */
function getActiveItem(): string | null {
  if (activeIndex.value < 0) return null
  const suggestLen = suggestItems.value.length
  if (activeIndex.value < suggestLen) {
    return suggestItems.value[activeIndex.value]
  }
  const historyIdx = activeIndex.value - suggestLen
  if (historyIdx < searchHistory.value.length) {
    return searchHistory.value[historyIdx]
  }
  return null
}

/** 重置高亮索引 */
function resetActiveIndex() {
  activeIndex.value = -1
}

/** 监听 keyword 变化：触发联想 + 重置高亮 */
watch(
  () => props.keyword,
  () => {
    resetActiveIndex()
    fetchSuggest()
  }
)

/** 监听 visible 变化：显示时刷新历史和热词 */
watch(
  () => props.visible,
  (val) => {
    if (val) {
      loadHotKeywords()
      loadSearchHistory()
    } else {
      resetActiveIndex()
    }
  }
)

/** 暴露键盘导航方法供父组件调用 */
defineExpose({
  navigateKey,
  getActiveItem,
  resetActiveIndex
})

onMounted(() => {
  loadHotKeywords()
  loadSearchHistory()
})

onBeforeUnmount(() => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
})
</script>

<style scoped>
.search-suggest {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 60vh;
  overflow-y: auto;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-lg);
  box-shadow: var(--shop-shadow-xl);
  z-index: 100;
  padding: 8px 0;
}

.suggest-section {
  padding: 4px 0;
}
.suggest-section + .suggest-section {
  border-top: 1px solid var(--shop-border);
}

.suggest-section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px 4px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--shop-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.suggest-clear {
  margin-left: auto;
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  text-transform: none;
  transition: color 0.15s;
}
.suggest-clear:hover {
  color: var(--shop-destructive);
}

.suggest-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.suggest-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 16px;
  font-size: 0.875rem;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: background 0.15s;
  min-height: 36px;
}
.suggest-item:hover,
.suggest-item.active {
  background: var(--shop-primary-bg);
}
.suggest-item.active {
  color: var(--shop-primary);
}

.suggest-item-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.suggest-item-text :deep(mark) {
  background: transparent;
  color: var(--shop-primary);
  font-weight: 600;
}

.suggest-item-delete {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  color: var(--shop-muted-foreground);
  opacity: 0;
  transition: opacity 0.15s, color 0.15s;
  flex-shrink: 0;
}
.suggest-item:hover .suggest-item-delete {
  opacity: 1;
}
.suggest-item-delete:hover {
  color: var(--shop-destructive);
}

.suggest-loading,
.suggest-empty {
  padding: 12px 16px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

.suggest-empty-all {
  padding: 24px 16px;
  text-align: center;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* 热词区 */
.hot-keywords {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 16px 8px;
}

.hot-keyword {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: 999px;
  background: var(--shop-muted);
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.hot-keyword:hover {
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

.hot-rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 3px;
  font-size: 0.6875rem;
  font-weight: 700;
  background: var(--shop-border);
  color: var(--shop-muted-foreground);
}
.hot-top .hot-rank {
  background: var(--shop-destructive);
  color: #fff;
}

/* 过渡动画 */
.suggest-fade-enter-active,
.suggest-fade-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}
.suggest-fade-enter-from,
.suggest-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* 移动端适配 */
@media (max-width: 767px) {
  .search-suggest {
    max-height: 50vh;
  }
  .suggest-item {
    min-height: 40px;
  }
}
</style>
