<template>
  <div class="cat-picker">
    <!-- 层级面包屑（可点击返回上级） -->
    <div class="cat-picker-crumbs">
      <span
        class="cat-picker-crumb"
        :class="{ active: stack.length === 0 }"
        @click="goRoot"
      >全部</span>
      <template v-for="(c, i) in stack" :key="c.id">
        <span class="cat-picker-sep">/</span>
        <span
          class="cat-picker-crumb"
          :class="{ active: i === stack.length - 1 }"
          @click="goTo(i)"
        >{{ c.name }}</span>
      </template>
    </div>

    <!-- 当前层级分类列表（一行一行，点击进入子级或选中） -->
    <div class="cat-picker-list">
      <div
        v-for="cat in currentViewCats"
        :key="cat.id"
        class="cat-picker-row"
        :class="{ selected: cat.id === selectedId }"
        @click="onRowClick(cat)"
      >
        <span class="cat-picker-icon" v-html="renderIcon(cat.icon)"></span>
        <span class="cat-picker-name">{{ cat.name }}</span>
        <!-- 有子级：右侧箭头，点击进入子级 -->
        <svg v-if="cat.children && cat.children.length" class="cat-picker-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
        <!-- 无子级且选中：对勾 -->
        <svg v-else-if="cat.id === selectedId" class="cat-picker-check" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="20 6 9 17 4 12"/>
        </svg>
      </div>
      <div v-if="currentViewCats.length === 0" class="cat-picker-empty">暂无子分类</div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 商品分类层级钻取选择器
 * 点击一级 → 视图切换为二级列表（一级隐藏），面包屑可逐级返回
 * 有子级的分类点击进入子级，叶子分类点击直接选中筛选
 */
import { ref, computed, watch } from 'vue'
import type { CategoryNode } from '@/types/product'

const props = defineProps<{
  categories: CategoryNode[]
  /** 当前选中的分类 ID（null=全部商品） */
  selectedId: number | null
}>()

const emit = defineEmits<{
  (e: 'select', id: number): void
  (e: 'clear'): void
}>()

/** 当前钻取路径（不含"全部"根节点，不含选中的叶子自身） */
const stack = ref<CategoryNode[]>([])

/** 当前层级展示的分类列表：有钻取路径则显示最后一级的子分类，否则显示顶级 */
const currentViewCats = computed(() => {
  if (stack.value.length === 0) return props.categories
  const last = stack.value[stack.value.length - 1]
  return last.children || []
})

/** 在树中查找祖先链路径 */
function findPath(nodes: CategoryNode[], target: number, trail: number[] = []): number[] | null {
  for (const n of nodes) {
    if (n.id === target) return [...trail, n.id]
    if (n.children && n.children.length > 0) {
      const found = findPath(n.children, target, [...trail, n.id])
      if (found) return found
    }
  }
  return null
}

/** 点击分类行：选中筛选 + 有子级则进入子级视图 */
function onRowClick(cat: CategoryNode) {
  emit('select', cat.id)
  if (cat.children && cat.children.length > 0) {
    stack.value = [...stack.value, cat]
  }
}

/** 点击面包屑回到第 i 层（0 为一级） */
function goTo(i: number) {
  stack.value = stack.value.slice(0, i + 1)
  const target = stack.value[i]
  if (target) emit('select', target.id)
}

/** 返回"全部"：清空钻取路径，清空筛选 */
function goRoot() {
  stack.value = []
  emit('clear')
}

/** 监听选中变化：URL 直达/刷新/前进后退时重建钻取路径 */
watch(
  () => props.selectedId,
  (id) => {
    if (id === null) {
      stack.value = []
      return
    }
    // 已在当前钻取路径内（正常点击钻取），无需重建
    if (stack.value.some(c => c.id === id)) return
    // 根据祖先链重建钻取路径（自身作为当前列表高亮项）
    const path = findPath(props.categories, id)
    if (!path) {
      stack.value = []
      return
    }
    const nodes: CategoryNode[] = []
    let list = props.categories
    for (const pid of path) {
      const n = list.find(x => x.id === pid)
      if (!n) break
      nodes.push(n)
      list = n.children || []
    }
    stack.value = nodes.slice(0, -1)
  }
)

/**
 * 渲染分类图标：SVG 直接输出、图片 URL 用 img、否则按 Emoji 文本展示
 * @param icon 图标内容（SVG / 图片URL / Emoji）
 */
function renderIcon(icon: string): string {
  if (!icon) return '📁'
  if (icon.trim().startsWith('<svg')) return icon
  if (/^(https?:|data:image)/.test(icon.trim())) {
    return `<img src="${icon}" alt="分类图标" style="width:100%;height:100%;object-fit:contain;" />`
  }
  return icon
}
</script>

<style scoped>
.cat-picker-crumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 6px 10px;
  margin-bottom: 6px;
  background: var(--shop-muted);
  border-radius: 8px;
  font-size: 0.75rem;
}
.cat-picker-crumb {
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: color 0.15s;
  white-space: nowrap;
}
.cat-picker-crumb:hover {
  color: var(--shop-primary);
}
.cat-picker-crumb.active {
  color: var(--shop-primary);
  font-weight: 600;
}
.cat-picker-sep {
  color: var(--shop-border);
  font-size: 0.6875rem;
}

.cat-picker-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 380px;
  overflow-y: auto;
  margin: 0 -8px;
  padding: 0 8px;
}
.cat-picker-list::-webkit-scrollbar {
  width: 4px;
}
.cat-picker-list::-webkit-scrollbar-thumb {
  background: var(--shop-border);
  border-radius: 999px;
}

.cat-picker-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border-radius: 8px;
  cursor: pointer;
  user-select: none;
  font-size: 0.8125rem;
  transition: background 0.15s, color 0.15s;
  position: relative;
}
.cat-picker-row:hover {
  background: var(--shop-muted);
}
.cat-picker-row.selected {
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  font-weight: 600;
}
.cat-picker-row.selected::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 14px;
  border-radius: 3px;
  background: var(--shop-primary);
}

.cat-picker-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  line-height: 1;
  color: var(--shop-primary);
}
.cat-picker-icon svg {
  width: 18px;
  height: 18px;
}
.cat-picker-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cat-picker-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--shop-foreground);
}
.cat-picker-row.selected .cat-picker-name {
  color: var(--shop-primary);
}

.cat-picker-arrow {
  flex-shrink: 0;
  color: var(--shop-muted-foreground);
}
.cat-picker-row:hover .cat-picker-arrow {
  color: var(--shop-primary);
}
.cat-picker-check {
  flex-shrink: 0;
  color: var(--shop-primary);
}

.cat-picker-empty {
  padding: 20px 0;
  text-align: center;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}
</style>
