<template>
  <!-- 商品列表页：面包屑 + 排序筛选 + 商品网格 + 分页 -->
  <div class="product-list-page">
    <!-- ============================================================ -->
    <!-- 面包屑导航 -->
    <!-- ============================================================ -->
    <section class="container-shop breadcrumb-section">
      <div class="breadcrumb">
        <router-link to="/" class="breadcrumb-link">首页</router-link>
        <template v-if="breadcrumbPath.length">
          <span class="breadcrumb-separator">&gt;</span>
          <template v-for="(bc, idx) in breadcrumbPath" :key="bc.id">
            <router-link
              v-if="idx < breadcrumbPath.length - 1"
              :to="{ path: '/products', query: { category: bc.id } }"
              class="breadcrumb-link"
            >{{ bc.name }}</router-link>
            <span v-else class="breadcrumb-current">{{ bc.name }}</span>
            <span v-if="idx < breadcrumbPath.length - 1" class="breadcrumb-separator">&gt;</span>
          </template>
        </template>
        <template v-else>
          <span class="breadcrumb-separator">&gt;</span>
          <span class="breadcrumb-current">{{ activeCategoryName }}</span>
        </template>
      </div>
    </section>

    <!-- ============================================================ -->
    <!-- 页面标题 + 排序 + 视图切换 -->
    <!-- ============================================================ -->
    <section class="container-shop toolbar-section">
      <div class="toolbar">
        <h1 class="toolbar-title">
          <template v-if="activeKeyword">
            搜索：{{ activeKeyword }}
          </template>
          <template v-else>
            {{ activeCategoryName }}
          </template>
          <span class="toolbar-count">（共 {{ totalCount }} 件）</span>
        </h1>
        <div class="toolbar-actions">
          <!-- 移动端筛选按钮 -->
          <button class="filter-toggle-btn" @click="mobileFilterOpen = !mobileFilterOpen">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="4" y1="6" x2="20" y2="6"/><line x1="8" y1="12" x2="20" y2="12"/><line x1="12" y1="18" x2="20" y2="18"/>
            </svg>
            筛选
          </button>
          <!-- 排序下拉 -->
          <select v-model="sortBy" class="sort-select" @change="onSortChange">
            <option value="default">综合排序</option>
            <option value="sales">销量优先</option>
            <option value="price-asc">价格从低到高</option>
            <option value="price-desc">价格从高到低</option>
          </select>
          <!-- 视图切换 -->
          <div class="view-toggle">
            <button
              :class="{ active: viewMode === 'grid' }"
              title="网格视图"
              @click="viewMode = 'grid'"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/>
              </svg>
            </button>
            <button
              :class="{ active: viewMode === 'list' }"
              title="列表视图"
              @click="viewMode = 'list'"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- ============================================================ -->
    <!-- 主体区域：桌面侧边栏 + 商品列表 -->
    <!-- ============================================================ -->
    <main class="container-shop main-section">
      <div class="main-layout">
        <!-- 桌面端侧边栏筛选 -->
        <aside class="sidebar-desktop">
          <div class="sidebar-sticky">
            <!-- 商品分类（层级钻取：点击一级进入二级，一级隐藏，面包屑返回） -->
            <div class="sidebar-block">
              <h4 class="sidebar-title">商品分类</h4>
              <CategoryLevelPicker
                :categories="categories"
                :selected-id="activeCategoryId"
                @select="selectCategory"
                @clear="selectCategory('all')"
              />
            </div>
            <!-- 价格区间 -->
            <div class="sidebar-block">
              <h4 class="sidebar-title">价格区间</h4>
              <div class="price-filter-row">
                <input
                  v-model.number="priceMin"
                  type="number"
                  class="price-input"
                  placeholder="¥最低"
                  min="0"
                />
                <span class="price-separator">—</span>
                <input
                  v-model.number="priceMax"
                  type="number"
                  class="price-input"
                  placeholder="¥最高"
                  min="0"
                />
              </div>
              <button class="price-confirm-btn" @click="onApplyPriceFilter">确定</button>
            </div>
          </div>
        </aside>

        <!-- 商品列表区域 -->
        <section class="product-section">
          <!-- 移动端筛选面板 -->
          <div v-show="mobileFilterOpen" class="filter-mobile-panel">
            <div class="sidebar-block">
              <h4 class="sidebar-title">商品分类</h4>
              <CategoryLevelPicker
                :categories="categories"
                :selected-id="activeCategoryId"
                @select="selectCategory"
                @clear="selectCategory('all')"
              />
            </div>
            <div class="sidebar-block">
              <h4 class="sidebar-title">价格区间</h4>
              <div class="price-filter-row">
                <input v-model.number="priceMin" type="number" class="price-input" placeholder="¥最低" min="0" />
                <span class="price-separator">—</span>
                <input v-model.number="priceMax" type="number" class="price-input" placeholder="¥最高" min="0" />
              </div>
              <button class="price-confirm-btn" @click="onApplyPriceFilter">确定</button>
            </div>
          </div>

          <!-- 加载状态 -->
          <div v-if="loading" class="loading-state">
            <svg class="loading-spinner" width="40" height="40" viewBox="0 0 24 24" fill="none">
              <circle class="spinner-path" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
            </svg>
            <p>加载中...</p>
          </div>

          <!-- 商品网格 -->
          <div v-else :class="['product-grid', { 'product-list-layout': viewMode === 'list' }]">
            <template v-for="item in pagedProducts" :key="item.id">
              <ProductCard :product="item" v-if="viewMode === 'grid'" />
              <!-- 列表视图 -->
              <article v-else class="product-card-list" :style="cardListStyle">
                <router-link :to="`/product/${item.id}`" class="product-list-link">
                  <div class="list-img-wrap">
                    <span v-if="(item as any).tags?.length" class="list-badge" :class="getBadgeClass((item as any).tags[0])">
                      {{ getBadgeText((item as any).tags[0]) }}
                    </span>
                    <img :src="item.image" :alt="item.name" loading="lazy" />
                  </div>
                  <div class="list-info">
                    <h3 class="list-name">{{ item.name }}</h3>
                    <p class="list-desc">{{ stripHtml(item.description) }}</p>
                    <div class="list-price-row">
                      <span class="list-price">¥{{ formatPrice(item.price) }}</span>
                      <span v-if="(item as any).originalPrice && Number((item as any).originalPrice) > Number(item.price)" class="list-original-price">¥{{ formatPrice((item as any).originalPrice) }}</span>
                    </div>
                    <div class="list-meta">
                      <span>已售 {{ formatSales((item as any).sales || 0) }} 件</span>
                      <span>库存 {{ item.stock }} 件</span>
                      <span class="stars">{{ getRatingStars((item as any).rating || 0) }}</span>
                    </div>
                  </div>
                </router-link>
              </article>
            </template>
          </div>

          <!-- 空状态 -->
          <div v-if="!loading && pagedProducts.length === 0" class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--shop-muted-foreground)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
            </svg>
            <!-- 搜索场景空结果 -->
            <p v-if="activeKeyword">没有找到与「{{ activeKeyword }}」相关的商品</p>
            <p v-else>没有找到符合条件的商品</p>
            <!-- 搜索无结果时展示热词推荐 -->
            <div v-if="activeKeyword && hotKeywords.length > 0" class="empty-hot-keywords">
              <p class="empty-hot-title">试试搜索这些：</p>
              <div class="empty-hot-list">
                <span
                  v-for="(kw, idx) in hotKeywords"
                  :key="'empty-hot-' + idx"
                  class="empty-hot-tag"
                  @click="onHotKeywordClick(kw)"
                >{{ kw }}</span>
              </div>
            </div>
          </div>

          <!-- ============================================================ -->
          <!-- 分页器 -->
          <!-- ============================================================ -->
          <div v-if="totalPages > 1" class="pagination">
            <button
              :class="['page-btn', { 'page-btn--disabled': currentPage <= 1 }]"
              :disabled="currentPage <= 1"
              aria-label="上一页"
              @click="onPageChange(currentPage - 1)"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="15 18 9 12 15 6"/>
              </svg>
            </button>
            <button
              v-for="page in visiblePages"
              :key="page"
              :class="['page-btn', { 'page-btn--active': page === currentPage, 'page-btn--ellipsis': page === -1 }]"
              :disabled="page === -1"
              @click="page !== -1 && onPageChange(page)"
            >
              {{ page === -1 ? '...' : page }}
            </button>
            <button
              :class="['page-btn', { 'page-btn--disabled': currentPage >= totalPages }]"
              :disabled="currentPage >= totalPages"
              aria-label="下一页"
              @click="onPageChange(currentPage + 1)"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ProductCard from '@/components/ProductCard.vue'
import CategoryLevelPicker from '@/components/CategoryLevelPicker.vue'
import { getProducts, getCategoryTree, type ProductItem } from '@/api/product'
import { useShopStore } from '@/store/shop'
import type { CategoryNode } from '@/types/product'

const route = useRoute()
const router = useRouter()
const shopStore = useShopStore()

const allProducts = ref<ProductItem[]>([])
const categories = ref<CategoryNode[]>([])

const activeCategory = ref<string>((route.query.category as string) || 'all')
const activeKeyword = ref<string>((route.query.keyword as string) || '')
const sortBy = ref<string>('default')
const viewMode = ref<'grid' | 'list'>('grid')
const mobileFilterOpen = ref(false)

/** 热门搜索词（搜索无结果时推荐） */
const hotKeywords = computed(() => shopStore.hotKeywords || [])

/**
 * 点击热词推荐：跳转到该关键词的搜索结果
 * @param kw 热门搜索词
 */
function onHotKeywordClick(kw: string) {
  router.push({ path: '/products', query: { keyword: kw } })
}

const priceMin = ref<number | undefined>(undefined)
const priceMax = ref<number | undefined>(undefined)

const currentPage = ref(1)
const pageSize = 8
const totalCount = ref(0)
const loading = ref(false)

const cardListStyle = {
  background: 'var(--shop-card)',
  boxShadow: 'var(--shop-shadow-md)'
}

async function loadCategories() {
  try {
    const res = await getCategoryTree()
    categories.value = normalizeTree(res.items || [])
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

/**
 * 归一化分类树：protojson 将 int64 序列化为字符串，统一转为 number，
 * 保证树节点 id 与选中/展开状态（number）严格比较一致
 */
function normalizeTree(nodes: any[]): CategoryNode[] {
  return (nodes || []).map(n => ({
    ...n,
    id: Number(n.id),
    parentId: n.parent_id != null ? Number(n.parent_id) : 0,
    children: normalizeTree(n.children)
  }))
}

/** 当前选中分类的数字 ID（null 表示"全部商品"） */
const activeCategoryId = computed<number | null>(() => {
  return activeCategory.value === 'all' ? null : Number(activeCategory.value)
})

/**
 * 在分类树中查找某个分类的祖先链路径
 * @param nodes 树节点
 * @param target 目标分类 ID
 * @param trail 当前已走过的祖先链
 */
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

/**
 * 选择分类（同步 URL，由 watch 统一触发加载）
 * @param id 分类 ID 或 'all'（全部商品）
 */
function selectCategory(id: number | 'all') {
  if (id === 'all') {
    router.replace({ path: '/products', query: activeKeyword.value ? { keyword: activeKeyword.value } : undefined })
  } else {
    router.replace({ path: '/products', query: { category: String(id), ...(activeKeyword.value ? { keyword: activeKeyword.value } : {}) } })
  }
}

/** 当前选中分类名称（用于标题） */
const activeCategoryName = computed(() => {
  if (activeCategory.value === 'all') return '全部商品'
  const path = findPath(categories.value, Number(activeCategory.value))
  return path ? (categories.value.length ? findName(path) : '') : '全部商品'
})

/** 由祖先链 ID 反查名称（最后一个为当前分类） */
function findName(path: number[]): string {
  let nodes = categories.value
  let name = ''
  for (let i = 0; i < path.length; i++) {
    const n = nodes.find(x => x.id === path[i])
    if (!n) break
    name = n.name
    nodes = n.children || []
  }
  return name
}

/** 面包屑祖先链（首页 > 电子产品 > 手机） */
const breadcrumbPath = computed(() => {
  if (activeCategory.value === 'all') return [] as { id: number; name: string }[]
  const path = findPath(categories.value, Number(activeCategory.value))
  if (!path) return [] as { id: number; name: string }[]
  const result: { id: number; name: string }[] = []
  let nodes = categories.value
  for (const id of path) {
    const n = nodes.find(x => x.id === id)
    if (!n) break
    result.push({ id, name: n.name })
    nodes = n.children || []
  }
  return result
})

async function loadProducts() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      pageSize: pageSize
    }
    if (activeCategory.value !== 'all') {
      params.category = activeCategory.value
    }
    if (activeKeyword.value) {
      params.keyword = activeKeyword.value
    }
    if (sortBy.value !== 'default') {
      params.sort = sortBy.value
    }
    const res = await getProducts(params)
    allProducts.value = res.items || []
    totalCount.value = res.total || 0
  } catch (error) {
    console.error('加载商品列表失败:', error)
    allProducts.value = []
    totalCount.value = 0
  } finally {
    loading.value = false
  }
}

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize)))

const pagedProducts = computed(() => {
  return allProducts.value
})

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
    return pages
  }

  pages.push(1)
  if (currentPage.value > 3) pages.push(-1)

  const start = Math.max(2, currentPage.value - 1)
  const end = Math.min(total - 1, currentPage.value + 1)
  for (let i = start; i <= end; i++) pages.push(i)

  if (currentPage.value < total - 2) pages.push(-1)
  pages.push(total)

  return pages
})

/** 监听 URL 参数变化（分类/关键词），统一触发列表加载 */
watch(
  () => [route.query.category, route.query.keyword],
  () => {
    const nextCategory = (route.query.category as string) || 'all'
    const nextKeyword = (route.query.keyword as string) || ''
    // 仅当条件变化时重置页码并重新加载，避免重复请求
    if (nextCategory !== activeCategory.value || nextKeyword !== activeKeyword.value) {
      activeCategory.value = nextCategory
      activeKeyword.value = nextKeyword
      currentPage.value = 1
      loadProducts()
    }
  }
)

function onSortChange() {
  currentPage.value = 1
  loadProducts()
}

function onApplyPriceFilter() {
  currentPage.value = 1
  loadProducts()
}

function onPageChange(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadProducts()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function getBadgeText(tag: string): string {
  if (tag === 'new') return '新品'
  if (tag === 'hot') return '热卖'
  if (tag === 'sale') return '特惠'
  return ''
}

function getBadgeClass(tag: string): string {
  if (tag === 'new') return 'badge-new'
  if (tag === 'hot') return 'badge-hot'
  if (tag === 'sale') return 'badge-sale'
  return ''
}

function getRatingStars(rating: number): string {
  const r = Math.round(rating)
  return '★'.repeat(r) + '☆'.repeat(5 - r)
}

function formatSales(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return n.toLocaleString()
}

function formatPrice(price: number | string): string {
  const p = Number(price) / 100
  return p.toFixed(2)
}

function stripHtml(html: string): string {
  if (!html) return ''
  const div = document.createElement('div')
  div.innerHTML = html
  return div.textContent || div.innerText || ''
}

onMounted(() => {
  loadCategories()
  loadProducts()
})
</script>

<style scoped>
/* ============================================
   面包屑导航
   ============================================ */
.breadcrumb-section {
  padding-top: 16px;
  padding-bottom: 8px;
}

.breadcrumb {
  font-size: 0.875rem;
}

.breadcrumb-link {
  color: var(--shop-muted-foreground);
  text-decoration: none;
  transition: color 0.2s;
}

.breadcrumb-link:hover {
  color: var(--shop-primary);
}

.breadcrumb-separator {
  margin: 0 6px;
  color: var(--shop-border);
}

.breadcrumb-current {
  color: var(--shop-foreground);
  font-weight: 500;
}

/* ============================================
   工具栏
   ============================================ */
.toolbar-section {
  padding-bottom: 12px;
}

.toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

@media (min-width: 640px) {
  .toolbar {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.toolbar-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.toolbar-count {
  font-size: 1rem;
  font-weight: 400;
  color: var(--shop-muted-foreground);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* ============================================
   排序下拉框
   ============================================ */
.sort-select {
  padding: 6px 32px 6px 14px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-md);
  font-size: 0.875rem;
  font-family: var(--shop-font-sans);
  background: var(--shop-background);
  color: var(--shop-foreground);
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  transition: border-color 0.2s;
}

.sort-select:focus {
  outline: none;
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px var(--shop-primary-bg);
}

/* ============================================
   视图切换
   ============================================ */
.view-toggle {
  display: inline-flex;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-md);
  overflow: hidden;
}

.view-toggle button {
  padding: 6px 10px;
  background: var(--shop-background);
  border: none;
  cursor: pointer;
  color: var(--shop-muted-foreground);
  transition: background 0.2s, color 0.2s;
  line-height: 1;
}

.view-toggle button:hover {
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

.view-toggle button.active {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}

/* 分类筛选由 CategoryLevelPicker 组件承载（层级钻取） */

/* ============================================
   主体布局
   ============================================ */
.main-section {
  padding-top: 16px;
  padding-bottom: 16px;
}

.main-layout {
  display: flex;
  gap: 24px;
}

/* ============================================
   桌面端侧边栏
   ============================================ */
.sidebar-desktop {
  display: none;
  flex-shrink: 0;
  width: 240px;
}

@media (min-width: 1024px) {
  .sidebar-desktop {
    display: block;
  }
}

.sidebar-sticky {
  position: sticky;
  top: 92px;
}

.sidebar-block {
  margin-bottom: 24px;
  padding: 16px;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-muted);
}

.sidebar-title {
  margin-bottom: 12px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.price-filter-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-md);
  font-size: 0.875rem;
  font-family: var(--shop-font-sans);
  transition: border-color 0.2s;
  outline: none;
}

.price-input:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px var(--shop-primary-bg);
}

.price-separator {
  color: var(--shop-muted-foreground);
  flex-shrink: 0;
}

.price-confirm-btn {
  width: 100%;
  margin-top: 12px;
  padding: 8px 16px;
  border: none;
  border-radius: var(--shop-radius-md);
  background: var(--shop-primary);
  color: #fff;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.price-confirm-btn:hover {
  opacity: 0.9;
}

/* ============================================
   移动端筛选面板
   ============================================ */
.filter-mobile-panel {
  margin-bottom: 16px;
  padding: 16px;
  border-radius: var(--shop-radius-lg);
  background: var(--shop-muted);
  border: 1px solid var(--shop-border);
}

@media (min-width: 1024px) {
  .filter-mobile-panel {
    display: none;
  }
}

.filter-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  border: 1px solid var(--shop-border);
  border-radius: 9999px;
  background: var(--shop-background);
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
}

.filter-toggle-btn:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
}

@media (min-width: 1024px) {
  .filter-toggle-btn {
    display: none;
  }
}

/* ============================================
   商品列表区域
   ============================================ */
.product-section {
  flex: 1;
  min-width: 0;
}

/* 网格视图 */
.product-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

@media (min-width: 640px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1280px) {
  .product-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* 列表视图覆盖 */
.product-list-layout {
  grid-template-columns: 1fr !important;
}

/* ============================================
   列表视图卡片
   ============================================ */
.product-card-list {
  border-radius: var(--shop-radius-xl);
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.product-card-list:hover {
  transform: translateY(-4px);
  box-shadow: var(--shop-shadow-xl);
}

.product-list-link {
  display: flex;
  flex-direction: row;
  text-decoration: none;
  color: inherit;
}

.list-img-wrap {
  position: relative;
  width: 192px;
  flex-shrink: 0;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  background: var(--shop-muted);
}

.list-img-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}

.product-card-list:hover .list-img-wrap img {
  transform: scale(1.08);
}

.list-badge {
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

.list-info {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.list-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--shop-card-foreground);
}

.list-desc {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.list-price-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.list-price {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-primary);
}

.list-original-price {
  font-size: 0.8125rem;
  text-decoration: line-through;
  color: var(--shop-muted-foreground);
}

.list-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}

/* ============================================
   加载状态
   ============================================ */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 64px 0;
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

.spinner-path {
  stroke: var(--shop-primary);
  stroke-dasharray: 60;
  stroke-dashoffset: 30;
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes dash {
  0% { stroke-dashoffset: 60; }
  50% { stroke-dashoffset: 0; }
  100% { stroke-dashoffset: -60; }
}

/* ============================================
   空状态
   ============================================ */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 64px 0;
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
}

/* 搜索无结果时的热词推荐 */
.empty-hot-keywords {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.empty-hot-title {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
}
.empty-hot-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  max-width: 400px;
}
.empty-hot-tag {
  padding: 6px 16px;
  border-radius: 999px;
  background: var(--shop-muted);
  font-size: 0.8125rem;
  color: var(--shop-foreground);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.empty-hot-tag:hover {
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
}

/* ============================================
   分页器
   ============================================ */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 40px;
  margin-bottom: 48px;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: var(--shop-radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.2s, color 0.2s, border-color 0.2s;
  cursor: pointer;
  border: 1px solid var(--shop-border);
  background: var(--shop-background);
  color: var(--shop-foreground);
}

.page-btn:hover:not(.page-btn--active):not(.page-btn--disabled) {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.page-btn--active {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  border-color: var(--shop-primary);
}

.page-btn--disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-btn--ellipsis {
  border: none;
  cursor: default;
  width: auto;
  padding: 0 4px;
}
</style>
