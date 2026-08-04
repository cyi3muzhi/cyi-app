<template>
  <div class="categories-page">
    <!-- 背景装饰 -->
    <div class="bg-deco">
      <div class="bg-blob bg-blob--1" />
      <div class="bg-blob bg-blob--2" />
      <div class="bg-blob bg-blob--3" />
    </div>

    <div class="categories-layout">
      <!-- ===== 左侧：分类树 ===== -->
      <div class="tree-panel">
        <div class="tree-panel-header">
          <div class="tree-panel-title">
            <svg class="tree-title-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
            分类目录
            <span class="tree-panel-badge">{{ totalCategoryCount }}</span>
          </div>
          <div class="tree-ctrl-btns">
            <button class="tree-ctrl-btn" title="展开全部" @click="expandAll">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
            <button class="tree-ctrl-btn" title="收起全部" @click="collapseAll">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
          </div>
        </div>

        <div class="tree-search-wrap">
          <div class="tree-search-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="treeSearch" type="text" placeholder="搜索分类..." />
          </div>
        </div>

        <div class="tree-panel-body" id="treeBody">
          <div v-for="item in treeNodes" :key="item.id" :class="['tree-item', { active: selectedNodeId === item.id }]" :style="{ paddingLeft: (12 + item._depth * 18) + 'px' }" @click="selectTreeNode(item.id)">
            <span v-if="item.hasChildren" :class="['tree-toggle', { expanded: expandedState[item.id] !== false }]" @click.stop="toggleTreeNode(item.id)">
              <svg width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
            </span>
            <span v-else class="tree-toggle-placeholder" />
            <span class="tree-dot" :class="{ active: selectedNodeId === item.id }" />
            <span class="tree-label">{{ item.name }}</span>
            <span class="tree-count">{{ item.count }}</span>
          </div>
          <div v-if="treeNodes.length === 0" class="tree-empty">无匹配分类</div>
        </div>

        <div class="tree-panel-footer">
          <span>共 <strong>{{ totalCategoryCount }}</strong> 个分类</span>
          <span class="tree-footer-detail">{{ topLevelCount }} 个一级 · {{ subLevelCount }} 个二级</span>
        </div>
      </div>

      <!-- ===== 右侧：内容区 ===== -->
      <div class="content-panel">
        <!-- 面包屑 -->
        <div class="breadcrumb-bar">
          <!-- 全部分类入口：点击返回顶级分类列表 -->
          <span
            class="bc-link"
            :class="{ 'bc-current': selectedNodeId === null }"
            @click="selectedNodeId = null"
          >全部分类</span>
          <template v-for="(bc, idx) in breadcrumbs" :key="bc.id">
            <span class="bc-sep">/</span>
            <span v-if="idx < breadcrumbs.length - 1" class="bc-link" @click="selectTreeNode(bc.id)">{{ bc.name }}</span>
            <span v-else class="bc-current">{{ bc.name }}</span>
          </template>
        </div>

        <!-- 工具栏 -->
        <div class="toolbar-bar">
          <div class="toolbar-left">
            <button class="tb-btn tb-btn--primary" @click="openAddModal">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              添加子分类
            </button>
            <div class="tb-dropdown">
              <button class="tb-btn tb-btn--secondary">
                批量操作
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
              </button>
              <div class="tb-dropdown-menu">
                <button @click="handleBatchCommand('enable')">批量启用</button>
                <button @click="handleBatchCommand('disable')">批量禁用</button>
                <div class="tb-dropdown-divider" />
                <button class="danger" @click="handleBatchCommand('delete')">批量删除</button>
              </div>
            </div>
          </div>
          <div class="toolbar-right">
            <div class="stats-tag"><span class="stats-dot" />当前 <strong>{{ currentChildren.length }}</strong></div>
            <div class="stats-tag"><span class="stats-dot stats-dot--green" />启用 <strong>{{ enabledCount }}</strong></div>
            <div class="stats-tag"><span class="stats-dot stats-dot--red" />禁用 <strong>{{ disabledCount }}</strong></div>
          </div>
        </div>

        <!-- 表格 -->
        <div class="table-card">
          <div class="table-header">
            <div class="th-cell th-cell--sort">排序</div>
            <div class="th-cell th-cell--name">分类名称</div>
            <div class="th-cell th-cell--slug">别名</div>
            <div class="th-cell th-cell--status">状态</div>
            <div class="th-cell th-cell--count">商品</div>
            <div class="th-cell th-cell--showcase">展示</div>
            <div class="th-cell th-cell--action">操作</div>
          </div>
          <div class="table-body" id="tableBody">
            <!-- 空状态 -->
            <div v-if="currentChildren.length === 0" class="table-empty">
              <div class="empty-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/></svg>
              </div>
              <div class="empty-title">暂无子分类</div>
              <div class="empty-hint">点击上方「添加子分类」按钮创建</div>
            </div>

            <!-- 表格行 -->
            <div v-for="(item, idx) in currentChildren" :key="item.id" class="table-row" :style="'--idx: ' + idx">
              <div class="td-cell td-cell--sort">
                <input type="number" class="sort-input" :value="item.sort" min="0" @change="updateSort(item, ($event.target as HTMLInputElement).value)" />
              </div>
              <div class="td-cell td-cell--name">
                <div class="cat-avatar" v-html="renderIcon(item.icon)"></div>
                <span class="cat-name-text">{{ item.name }}</span>
                <span v-if="item.children && item.children.length > 0" class="child-chip" @click="selectTreeNode(item.id)">{{ item.children.length }} 子级</span>
              </div>
              <div class="td-cell td-cell--slug"><code class="slug-tag">{{ item.slug }}</code></div>
              <div class="td-cell td-cell--status">
                <button :class="['switcher', { on: item.status }]" @click="toggleStatus(item)">
                  <span class="switcher-knob" />
                </button>
                <span :class="['switcher-label', { on: item.status }]">{{ item.status ? '启用' : '禁用' }}</span>
              </div>
              <div class="td-cell td-cell--count">
                <span class="count-badge">{{ item.count || 0 }}</span>
              </div>
              <div class="td-cell td-cell--showcase">
                <button :class="['star-btn', { active: item.showcase }]" @click="toggleShowcase(item)" :title="item.showcase ? '取消首页展示' : '设为首页展示'">
                  <svg width="18" height="18" viewBox="0 0 24 24" :fill="item.showcase ? '#f59e0b' : 'none'" :stroke="item.showcase ? '#f59e0b' : '#cbd5e1'" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                </button>
              </div>
              <div class="td-cell td-cell--action">
                <button class="act-link act-link--edit" @click="openEditModal(item)">编辑</button>
                <span class="act-divider" />
                <button class="act-link act-link--add" @click="openAddSubModal(item)">添加子类</button>
                <span class="act-divider" />
                <button class="act-link act-link--del" @click="confirmDelete(item)">删除</button>
              </div>
            </div>
          </div>
          <div class="table-footer">
            <span>显示全部 {{ currentChildren.length }} 个分类</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 添加/编辑弹窗 ===== -->
    <Teleport to="body">
      <div v-if="dialogVisible" class="dialog-overlay" @click.self="dialogVisible = false">
        <div class="dialog-panel">
          <div class="dialog-head">
            <button class="dialog-close" @click="dialogVisible = false">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
            <div class="dialog-title">{{ isEditing ? '编辑分类' : '添加分类' }}</div>
            <div class="dialog-sub">{{ isEditing ? '修改分类信息' : '创建一个新的商品分类' }}</div>
          </div>
          <div class="dialog-body">
            <!-- 名称 -->
            <div class="d-field">
              <label class="d-label">分类名称 <span class="d-required">*</span></label>
              <input v-model="formData.name" class="d-input" placeholder="请输入分类名称" maxlength="20" />
              <span class="d-counter">{{ formData.name.length }}/20</span>
            </div>

            <!-- Slug 和排序 -->
            <div class="d-row">
              <div class="d-field" style="flex: 1;">
                <label class="d-label">别名/Slug <span class="d-required">*</span></label>
                <input v-model="formData.slug" class="d-input" placeholder="自动生成，可手动修改" />
                <span class="d-hint">用于 URL 友好链接</span>
              </div>
              <div class="d-field" style="width: 140px;">
                <label class="d-label">排序</label>
                <input v-model.number="formData.sort" type="number" class="d-input" min="0" max="999" style="text-align:center;" />
              </div>
            </div>

            <!-- 上级分类 -->
            <div class="d-field">
              <label class="d-label">上级分类</label>
              <div class="d-select-wrap">
                <select v-model="formData.parentId" class="d-select">
                  <option :value="null">无（顶级分类）</option>
                  <option v-for="opt in parentOptions" :key="opt.id" :value="opt.id" :disabled="opt.id === editingId">{{ opt.label }}</option>
                </select>
                <svg class="d-select-arrow" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
              </div>
            </div>

            <!-- 图标 -->
            <div class="d-field">
              <label class="d-label">分类图标</label>
              <div class="icon-picker">
                <div class="icon-picker-top">
                  <!-- 当前图标预览 -->
                  <div class="icon-current" v-html="renderIcon(formData.icon)"></div>
                  <!-- Tab 切换 -->
                  <div class="icon-tabs">
                    <button type="button" :class="['icon-tab', { active: iconTab === 'svg' }]" @click="iconTab = 'svg'">图标库</button>
                    <button type="button" :class="['icon-tab', { active: iconTab === 'emoji' }]" @click="iconTab = 'emoji'">Emoji</button>
                    <button type="button" :class="['icon-tab', { active: iconTab === 'upload' }]" @click="iconTab = 'upload'">上传</button>
                  </div>
                </div>
                <!-- SVG 图标库 -->
                <div v-if="iconTab === 'svg'" class="icon-panel">
                  <button
                    v-for="ic in svgIconList"
                    :key="ic.name"
                    type="button"
                    :class="['icon-opt', { selected: formData.icon === ic.svg }]"
                    :title="ic.name"
                    @click="formData.icon = ic.svg"
                    v-html="ic.svg"
                  ></button>
                </div>
                <!-- Emoji -->
                <div v-else-if="iconTab === 'emoji'" class="icon-panel">
                  <button
                    v-for="emoji in emojiList"
                    :key="emoji"
                    type="button"
                    :class="['icon-opt', { selected: formData.icon === emoji }]"
                    @click="formData.icon = emoji"
                  >{{ emoji }}</button>
                </div>
                <!-- 上传图片 -->
                <div v-else class="icon-upload-zone" @click="triggerIconUpload">
                  <img v-if="isImageIcon(formData.icon)" :src="formData.icon" class="icon-upload-preview" />
                  <template v-else>
                    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                    <span>点击上传图标图片</span>
                    <small>建议方形小图 · PNG / JPG / SVG</small>
                  </template>
                </div>
                <input ref="iconInputRef" type="file" accept="image/*" style="display:none" @change="handleIconUpload" />
              </div>
            </div>

            <!-- 图片上传 -->
            <div class="d-field">
              <label class="d-label">分类图片</label>
              <div class="upload-zone" @click="triggerUpload">
                <img v-if="formData.image" :src="formData.image" class="upload-preview" />
                <template v-else>
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  <div class="upload-text">点击上传分类图片</div>
                  <div class="upload-hint">建议尺寸 200×200px · PNG / JPG</div>
                </template>
              </div>
              <input ref="fileInputRef" type="file" accept="image/*" style="display:none" @change="handleUpload" />
            </div>

            <!-- 状态 -->
            <div class="d-field">
              <label class="d-label">状态</label>
              <button :class="['switcher', { on: formData.status }]" @click="formData.status = !formData.status">
                <span class="switcher-knob" />
              </button>
              <span :class="['switcher-label', { on: formData.status }]" style="margin-left: 0.5rem;">{{ formData.status ? '启用' : '禁用' }}</span>
            </div>

            <!-- SEO -->
            <div class="d-seo">
              <button class="d-seo-toggle" @click="seoOpen = !seoOpen">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                SEO 设置
                <svg :class="['d-seo-arrow', { open: seoOpen }]" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
              </button>
              <div v-show="seoOpen" class="d-seo-body">
                <!-- AI 生成 SEO 按钮 -->
                <div class="ai-seo-row">
                  <el-button type="primary" size="small" :loading="seoLoading" @click="handleAIGenerateSEO">
                    AI 生成 SEO
                  </el-button>
                  <span class="ai-seo-hint">基于分类名称智能生成 SEO 信息</span>
                </div>
                <div class="d-field">
                  <label class="d-label">页面标题</label>
                  <input v-model="formData.seoTitle" class="d-input" placeholder="默认为分类名称" />
                </div>
                <div class="d-field">
                  <label class="d-label">SEO 关键词</label>
                  <input v-model="formData.seoKeywords" class="d-input" placeholder="多个关键词用逗号分隔" />
                </div>
                <div class="d-field">
                  <label class="d-label">SEO 描述</label>
                  <textarea v-model="formData.seoDesc" class="d-textarea" rows="3" placeholder="SEO 描述内容" />
                </div>
              </div>
            </div>
          </div>
          <div class="dialog-foot">
            <button class="d-btn d-btn--cancel" @click="dialogVisible = false">取消</button>
            <button class="d-btn d-btn--primary" @click="saveCategory">{{ isEditing ? '保存修改' : '确认创建' }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ===== 确认删除弹窗 ===== -->
    <Teleport to="body">
      <div v-if="confirmVisible" class="dialog-overlay" @click.self="confirmVisible = false">
        <div class="dialog-panel dialog-panel--sm">
          <div class="dialog-head">
            <button class="dialog-close" @click="confirmVisible = false">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
            <div class="dialog-title">确认删除</div>
          </div>
          <div class="dialog-body" style="text-align:center;padding:1.5rem 2rem;">
            <div class="del-icon">🗑️</div>
            <div class="del-title">{{ confirmTitle }}</div>
            <div class="del-msg">{{ confirmMsg }}</div>
          </div>
          <div class="dialog-foot">
            <button class="d-btn d-btn--cancel" @click="confirmVisible = false">取消</button>
            <button class="d-btn d-btn--danger" @click="doDelete">确认删除</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { listCategoryTree, createCategory as createCategoryApi, updateCategory as updateCategoryApi, deleteCategory as deleteCategoryApi, updateCategorySort, updateCategoryStatus, toggleCategoryShowcase, batchCategoryOperation } from '../../api/category'
import { generateCategorySEO } from '../../api/agent'

interface CategoryNode { id: number; name: string; slug: string; icon: string; sort: number; status: boolean; showcase: boolean; count: number; image?: string; seoTitle?: string; seoKeywords?: string; seoDesc?: string; parentId?: number; children?: CategoryNode[]; _depth?: number }

const treeData = ref<CategoryNode[]>([])
const catMap: Record<number, CategoryNode> = {}
const treeSearch = ref('')
const selectedNodeId = ref<number | null>(null)
const expandedState = reactive<Record<number, boolean>>({})
const dialogVisible = ref(false); const confirmVisible = ref(false)
const isEditing = ref(false); const editingId = ref<number | null>(null)
const seoOpen = ref(false); const fileInputRef = ref<HTMLInputElement | null>(null)
/** 图标选择器 Tab：svg=图标库 emoji=Emoji upload=上传 */
const iconTab = ref<'svg' | 'emoji' | 'upload'>('svg')
const iconInputRef = ref<HTMLInputElement | null>(null)
/** AI 生成 SEO loading */
const seoLoading = ref(false)
const deletingId = ref<number | null>(null); const confirmTitle = ref(''); const confirmMsg = ref('')
const formData = reactive({ id: null as number | null, name: '', slug: '', parentId: null as number | null, sort: 0, icon: '📁', status: true, showcase: false, image: '', count: 0, seoTitle: '', seoKeywords: '', seoDesc: '' })

/** 精选 SVG 图标库（线性风格，stroke=currentColor） */
interface PresetIcon { name: string; svg: string }
const svgIconList: PresetIcon[] = [
  { name: '手机', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>' },
  { name: '电脑', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>' },
  { name: '耳机', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 18v-6a9 9 0 0 1 18 0v6"/><path d="M21 19a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3zM3 19a2 2 0 0 0 2 2h1a2 2 0 0 0 2-2v-3a2 2 0 0 0-2-2H3z"/></svg>' },
  { name: '手表', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="6"/><polyline points="12 10 12 12 13 13"/><path d="M9 2l1.5 4M15 2l-1.5 4M9 22l1.5-4M15 22l-1.5-4"/></svg>' },
  { name: '相机', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>' },
  { name: '电视', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="7" width="20" height="15" rx="2"/><polyline points="17 2 12 7 7 2"/></svg>' },
  { name: '服装', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.34 0L12 6l-1.5-1.39a5.5 5.5 0 0 0-7.34 0A5.5 5.5 0 0 0 2.07 11L12 21l9.93-10a5.5 5.5 0 0 0-1.09-6.39z"/></svg>' },
  { name: '鞋', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 16v-2.38C4 11.5 2.97 10.5 3 8c.03-2.72 1.49-6 4.5-6C9.37 2 10 3.8 10 5.5c0 3.11-2 5.66-2 8.68V16a2 2 0 1 1-4 0Z"/><path d="M20 20v-2.38c0-2.12 1.03-3.12 1-5.62-.03-2.72-1.49-6-4.5-6C14.63 6 14 7.8 14 9.5c0 3.11 2 5.66 2 8.68V20a2 2 0 1 0 4 0Z"/><path d="M16 17h4"/><path d="M4 13h4"/></svg>' },
  { name: '箱包', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>' },
  { name: '食品', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>' },
  { name: '饮品', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8h1a4 4 0 0 1 0 8h-1"/><path d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z"/><line x1="6" y1="1" x2="6" y2="4"/><line x1="10" y1="1" x2="10" y2="4"/><line x1="14" y1="1" x2="14" y2="4"/></svg>' },
  { name: '家居', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>' },
  { name: '家具', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v3"/><path d="M2 11v5a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-5a2 2 0 0 0-4 0v2H6v-2a2 2 0 0 0-4 0Z"/><path d="M4 18v2"/><path d="M20 18v2"/></svg>' },
  { name: '厨具', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21a1 1 0 0 0 1-1v-5.35c0-.457.316-.844.727-1.041a4 4 0 0 0-2.134-7.589 5 5 0 0 0-9.186 0 4 4 0 0 0-2.134 7.588c.411.198.727.585.727 1.041V20a1 1 0 0 0 1 1Z"/><path d="M6 17h12"/></svg>' },
  { name: '美妆', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9.06 11.9 17.13 3.83a2.83 2.83 0 1 1 4 4L13.1 15.9"/><path d="M7.07 14.94c-1.66 0-3 1.35-3 3.02 0 1.33-2.5 1.52-2 2.02 1.08 1.1 2.49 2.02 4 2.02 2.2 0 4-1.8 4-4.04a3.01 3.01 0 0 0-3-3.02z"/></svg>' },
  { name: '图书', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>' },
  { name: '文具', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>' },
  { name: '运动', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M4.93 4.93l14.14 14.14M9.17 4.93 4.93 9.17M14.83 19.07l4.24-4.24M9.17 19.07l10.6-10.6M4.93 14.83l4.24 4.24"/></svg>' },
  { name: '玩具', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 12 20 22 4 22 4 12"/><rect x="2" y="7" width="20" height="5"/><line x1="12" y1="22" x2="12" y2="7"/><path d="M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7z"/><path d="M12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z"/></svg>' },
  { name: '宠物', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.34 0L12 5.59l-1.5-1.5a5.5 5.5 0 0 0-7.34 7.34L12 20l8.84-8.84a5.5 5.5 0 0 0 0-7.34z"/></svg>' },
  { name: '汽车', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="3" width="15" height="13"/><polygon points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/></svg>' },
  { name: '医药', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="m10.5 20.5 10-10a4.95 4.95 0 1 0-7-7l-10 10a4.95 4.95 0 1 0 7 7Z"/><path d="m8.5 8.5 7 7"/></svg>' },
  { name: '乐器', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>' },
  { name: '办公', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>' },
  { name: '母婴', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>' },
  { name: '出行', svg: '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>' }
]

const emojiList = ['📱','💻','🎧','⌚','📷','👕','👗','👟','🎒','🍔','🍿','🥤','🥩','🍷','🏠','🍳','🌀','🛋','🛏','📚','📖','✏️','⚽','🎮','🎵','💄','👶','🐾','🌿','🎁','📦','🔧','💊','🚗','✈️','🏕','🔌','🖨','🎤','🎸','🏆','🥇','🧸','🎈','🎀','💎','⌨️','🖥','📡','🕹','📀','💿','📷','🎥','📺','🔋','🛒','🧾','🏷','💳','📊','📋','📌','📍','✂️','🧵','👠','👒','🧤','🧣','🌂','🧴','🪥','🪒','🧹','🪣','🧯','🪜','🛠','🧰','⚙️','🔩','🔗','🧲']

const treeNodes = computed(() => { const f = treeSearch.value.trim().toLowerCase(); const r: (CategoryNode & { _depth: number; hasChildren: boolean })[] = []; flattenNodes(treeData.value, f, 0, r); return r })
const totalCategoryCount = computed(() => countAll(treeData.value))
const topLevelCount = computed(() => treeData.value.length)
const subLevelCount = computed(() => { let c = 0; for (const cat of treeData.value) { if (cat.children) c += cat.children.length } return c })
const breadcrumbs = computed(() => { const r: { id: number; name: string }[] = []; let cur = selectedNodeId.value ? getNodeById(selectedNodeId.value) : null; while (cur) { r.unshift({ id: cur.id, name: cur.name }); cur = getParentNode(cur.id) } return r })
const currentChildren = computed(() => { if (!selectedNodeId.value) return treeData.value; const n = getNodeById(selectedNodeId.value); return n?.children || [] })
const enabledCount = computed(() => currentChildren.value.filter(c => c.status).length)
const disabledCount = computed(() => currentChildren.value.filter(c => !c.status).length)
const parentOptions = computed(() => { const r: { id: number; label: string }[] = []; flattenOptions(treeData.value, r, '', editingId.value); return r })

async function loadTree() { try { const res = await listCategoryTree(); treeData.value = (res.items || []).map(n => { const node: CategoryNode = { id: n.id, name: n.name || '', slug: n.slug || '', icon: n.icon || '📁', sort: n.sort ?? 0, status: n.status ?? true, showcase: n.showcase ?? false, count: n.count ?? 0, image: n.image || '', seoTitle: n.seo_title || '', seoKeywords: n.seo_keywords || '', seoDesc: n.seo_desc || '', parentId: n.parent_id }; if (n.children && Array.isArray(n.children)) node.children = n.children.map(((child:any) => { const cnode: CategoryNode = { id: child.id, name: child.name || '', slug: child.slug || '', icon: child.icon || '📁', sort: child.sort ?? 0, status: child.status ?? true, showcase: child.showcase ?? false, count: child.count ?? 0, image: child.image || '', seoTitle: child.seo_title || '', seoKeywords: child.seo_keywords || '', seoDesc: child.seo_desc || '', parentId: child.parent_id }; if (child.children && Array.isArray(child.children)) cnode.children = child.children; return cnode })); return node }); buildCatMap(treeData.value); treeData.value.forEach(c => { expandedState[c.id] = true }); /* 默认展示全部分类（顶级列表），便于直接编辑/删除一级分类 */ } catch (e) { console.error('加载分类树失败:', e) } }

function flattenNodes(arr: CategoryNode[], filter: string, depth: number, result: (CategoryNode & { _depth: number; hasChildren: boolean })[]) { for (const node of arr) { const mn = !filter || node.name.toLowerCase().includes(filter); const cm = filter && node.children ? ((n:CategoryNode[]):boolean=>{for(const x of n){if(x.name.toLowerCase().includes(filter))return true;if(x.children&&hasMatchInTree(x.children,filter))return true}return false})(node.children) : false; if (filter && !mn && !cm) continue; const ie = filter ? true : (expandedState[node.id] !== false); result.push({ ...node, _depth: depth, hasChildren: !!(node.children && node.children.length > 0) }); if (node.children && ie) flattenNodes(node.children, filter, depth + 1, result) } }
function hasMatchInTree(arr: CategoryNode[], filter: string): boolean { for (const n of arr) { if (n.name.toLowerCase().includes(filter)) return true; if (n.children && hasMatchInTree(n.children, filter)) return true } return false }
function countAll(arr: CategoryNode[]): number { let c = 0; for (const n of arr) { c++; if (n.children) c += countAll(n.children) } return c }
function buildCatMap(arr: CategoryNode[]) { for (const n of arr) { catMap[n.id] = n; if (n.children) buildCatMap(n.children) } }
function getNodeById(id: number, arr: CategoryNode[] = treeData.value): CategoryNode | null { for (const n of arr) { if (n.id === id) return n; if (n.children) { const f = getNodeById(id, n.children); if (f) return f } } return null }
function getParentNode(id: number, arr: CategoryNode[] = treeData.value, parent: CategoryNode | null = null): CategoryNode | null { for (const n of arr) { if (n.id === id) return parent; if (n.children) { const f = getParentNode(id, n.children, n); if (f) return f } } return null }
function flattenOptions(arr: CategoryNode[], result: { id: number; label: string }[], prefix: string, skipId: number | null) { for (const n of arr) { if (n.id === skipId) continue; result.push({ id: n.id, label: prefix + n.name }); if (n.children) flattenOptions(n.children, result, prefix + '\u00A0\u00A0\u00A0\u00A0', skipId) } }
function toggleTreeNode(id: number) { expandedState[id] = expandedState[id] === false ? true : false }
function expandAll() { for (const key in catMap) { const n = catMap[key]; if (n.children && n.children.length > 0) expandedState[n.id] = true } }
function collapseAll() { for (const key in catMap) expandedState[catMap[key].id] = false }
function selectTreeNode(id: number | null) { selectedNodeId.value = id }
async function updateSort(item: CategoryNode, val: string) { const s = parseInt(val) || 0; item.sort = s; try { await updateCategorySort(item.id, s); ElMessage.success('排序已更新') } catch (e) { console.error(e) } }
async function toggleStatus(item: CategoryNode) { const ns = !item.status; item.status = ns; try { await updateCategoryStatus(item.id, ns); ElMessage.success(`${item.name} 已${ns ? '启用' : '禁用'}`) } catch (e) { item.status = !ns } }
async function toggleShowcase(item: CategoryNode) { const nv = !item.showcase; item.showcase = nv; try { await toggleCategoryShowcase(item.id, nv); ElMessage.success(`${item.name} 已${nv ? '设为首页展示' : '取消首页展示'}`) } catch (e) { item.showcase = !nv } }
async function handleBatchCommand(cmd: string) { const ids = currentChildren.value.map(c => c.id); if (ids.length === 0) { ElMessage.warning('当前没有可操作的分类'); return }; const m: Record<string, string> = { enable: '启用', disable: '禁用', delete: '删除' }; const action = m[cmd]; if (!action) return; try { await batchCategoryOperation(action, ids); ElMessage.success(`已批量${action} ${ids.length} 个分类`); await loadTree() } catch (e) { console.error(e) } }
function openAddModal() { resetForm(); isEditing.value = false; editingId.value = null; formData.parentId = selectedNodeId.value; dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function openAddSubModal(item: CategoryNode) { resetForm(); isEditing.value = false; editingId.value = null; formData.parentId = item.id; dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function openEditModal(item: CategoryNode) { isEditing.value = true; editingId.value = item.id; Object.assign(formData, { id: item.id, name: item.name, slug: item.slug, parentId: item.parentId ?? getParentNode(item.id)?.id ?? null, sort: item.sort, icon: item.icon, status: item.status, showcase: item.showcase, image: item.image || '', count: item.count, seoTitle: item.seoTitle || '', seoKeywords: item.seoKeywords || '', seoDesc: item.seoDesc || '' }); dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function resetForm() { Object.assign(formData, { id: null, name: '', slug: '', parentId: null, sort: 0, icon: '📁', status: true, showcase: false, image: '', count: 0, seoTitle: '', seoKeywords: '', seoDesc: '' }) }
/**
 * AI 生成 SEO 信息
 * 基于分类名称调用 AI 接口生成 SEO 标题、关键词和描述
 */
async function handleAIGenerateSEO() {
  if (!formData.name.trim()) {
    ElMessage.warning('请先输入分类名称')
    return
  }
  seoLoading.value = true
  try {
    const result = await generateCategorySEO(formData.name.trim())
    if (result.seoTitle || result.seoKeywords || result.seoDescription) {
      if (result.seoTitle) formData.seoTitle = result.seoTitle
      if (result.seoKeywords) formData.seoKeywords = result.seoKeywords
      if (result.seoDescription) formData.seoDesc = result.seoDescription
      ElMessage.success('SEO 信息已生成')
    } else {
      ElMessage.warning('AI 生成 SEO 失败，请重试')
    }
  } catch {
    ElMessage.error('AI 生成 SEO 失败')
  } finally {
    seoLoading.value = false
  }
}
async function saveCategory() { if (!formData.name.trim()) { ElMessage.warning('请输入分类名称'); return }; const payload = { name: formData.name.trim(), slug: formData.slug.trim() || formData.name.toLowerCase().replace(/\s+/g, '-'), icon: formData.icon || '📁', sort: formData.sort, status: formData.status, showcase: formData.showcase, image: formData.image, seo_title: formData.seoTitle, seo_keywords: formData.seoKeywords, seo_desc: formData.seoDesc, parent_id: formData.parentId || 0 }; try { if (isEditing.value && editingId.value) { await updateCategoryApi(editingId.value, payload); ElMessage.success(`分类「${formData.name}」已更新`) } else { await createCategoryApi(payload); ElMessage.success(`分类「${formData.name}」已创建`) }; dialogVisible.value = false; await loadTree() } catch (e) { console.error(e) } }
function confirmDelete(item: CategoryNode) {
  const subCount = countAll(item.children || [])
  deletingId.value = item.id
  confirmTitle.value = `确认删除「${item.name}」？`
  confirmMsg.value = subCount > 0
    ? `该分类下有 ${subCount} 个子分类，将一并删除。若分类下有商品则无法删除。此操作不可恢复。`
    : `此操作不可恢复，确定要删除分类「${item.name}」吗？`
  confirmVisible.value = true
}
async function doDelete() {
  if (deletingId.value === null) return
  try {
    await deleteCategoryApi(deletingId.value)
    ElMessage.success('分类已删除')
    confirmVisible.value = false
    deletingId.value = null
    await loadTree()
  } catch {
    // 错误提示由请求拦截器统一弹出（如：分类下有商品无法删除）
    confirmVisible.value = false
    deletingId.value = null
  }
}
function triggerUpload() { fileInputRef.value?.click() }
function handleUpload(event: Event) { const target = event.target as HTMLInputElement; const file = target.files?.[0]; if (!file) return; if (!file.type.startsWith('image/')) { ElMessage.warning('请选择图片文件'); return }; if (file.size > 2 * 1024 * 1024) { ElMessage.warning('图片大小不能超过 2MB'); return }; const reader = new FileReader(); reader.onload = (e) => { formData.image = e.target!.result as string }; reader.readAsDataURL(file); target.value = '' }

/* ==================== 图标判断与渲染 ==================== */

/** 是否为内联 SVG 图标 */
function isSvgIcon(icon: string): boolean {
  return !!icon && icon.trim().startsWith('<svg')
}
/** 是否为图片 URL（http 或 base64） */
function isImageIcon(icon: string): boolean {
  return !!icon && /^(https?:|data:image)/.test(icon.trim())
}
/**
 * 渲染分类图标：SVG 直接输出，图片用 img 标签，否则按 Emoji 文本展示
 * @param icon 图标内容（SVG/URL/Emoji）
 */
function renderIcon(icon: string): string {
  if (isSvgIcon(icon)) return icon
  if (isImageIcon(icon)) return `<img src="${icon}" alt="图标" style="width:100%;height:100%;object-fit:contain;border-radius:6px;" />`
  return icon || '📁'
}
/** 触发图标图片上传 */
function triggerIconUpload() { iconInputRef.value?.click() }
/** 上传图标图片（存为 base64，icon 字段已扩容为 TEXT） */
function handleIconUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) { ElMessage.warning('请选择图片文件'); return }
  if (file.size > 2 * 1024 * 1024) { ElMessage.warning('图片大小不能超过 2MB'); return }
  const reader = new FileReader()
  reader.onload = (e) => { formData.icon = e.target!.result as string }
  reader.readAsDataURL(file)
  target.value = ''
}
onMounted(() => { loadTree() })
</script>

<style>
/* ===== CSS 变量 ===== */
:root {
  --c-bg: #f0f2f5;
  --c-surface: rgba(255,255,255,0.82);
  --c-border: rgba(0,0,0,0.06);
  --c-text: #0f172a;
  --c-text2: #475569;
  --c-muted: #94a3b8;
  --c-accent: #6366f1;
  --c-accent2: #818cf8;
  --c-green: #10b981;
  --c-red: #ef4444;
  --c-amber: #f59e0b;
  --shadow-sm: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.03);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.03);
  --shadow-lg: 0 8px 30px rgba(0,0,0,0.06), 0 2px 8px rgba(0,0,0,0.04);
  --shadow-xl: 0 20px 60px rgba(0,0,0,0.10), 0 4px 16px rgba(0,0,0,0.06);
  --radius: 14px;
  --radius-sm: 10px;
  --radius-xs: 8px;
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
  --ease-out: cubic-bezier(0.16, 1, 0.3, 1);
}

/* ===== 动效 keyframes ===== */
@keyframes catFadeUp {
  from { opacity: 0; transform: translateY(18px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes catSlideLeft {
  from { opacity: 0; transform: translateX(-24px); }
  to { opacity: 1; transform: translateX(0); }
}
@keyframes catSlideRight {
  from { opacity: 0; transform: translateX(24px); }
  to { opacity: 1; transform: translateX(0); }
}
@keyframes catRowIn {
  from { opacity: 0; transform: translateY(10px) scale(0.97); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes blobFloat {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(20px, -30px) scale(1.05); }
  66% { transform: translate(-15px, 10px) scale(0.95); }
}
@keyframes dialogIn {
  from { opacity: 0; transform: translateY(20px) scale(0.96); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes overlayIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes shimmer {
  0% { background-position: -200% center; }
  100% { background-position: 200% center; }
}
@keyframes starPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}

/* ===== 页面布局 ===== */
.categories-page {
  position: relative;
  width: 100%;
  min-height: calc(100vh - 140px);
  padding: 0.5rem 0;
  animation: catFadeUp 0.5s var(--ease-out) both;
}
.bg-deco {
  position: fixed;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
}
.bg-blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.15;
  animation: blobFloat 20s ease-in-out infinite;
}
.bg-blob--1 { width: 400px; height: 400px; top: -80px; right: -60px; background: radial-gradient(circle, #6366f1, #8b5cf6); }
.bg-blob--2 { width: 300px; height: 300px; bottom: 40px; left: -60px; background: radial-gradient(circle, #06b6d4, #3b82f6); animation-delay: -7s; }
.bg-blob--3 { width: 250px; height: 250px; top: 40%; left: 50%; background: radial-gradient(circle, #f59e0b, #f97316); opacity: 0.08; animation-delay: -14s; }

.categories-layout {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 1.25rem;
  align-items: stretch;
  min-height: calc(100vh - 160px);
}

/* ===== 树形面板 ===== */
.tree-panel {
  width: 280px;
  min-width: 280px;
  background: var(--c-surface);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-radius: var(--radius);
  border: 1px solid var(--c-border);
  box-shadow: var(--shadow-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: catSlideLeft 0.5s var(--ease-out) both;
}
.tree-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.125rem;
  border-bottom: 1px solid var(--c-border);
}
.tree-panel-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--c-text);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.tree-title-icon { color: var(--c-accent); flex-shrink: 0; }
.tree-panel-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  line-height: 1.4;
}
.tree-ctrl-btns { display: flex; gap: 0.125rem; }
.tree-ctrl-btn {
  width: 1.75rem;
  height: 1.75rem;
  border: none;
  background: transparent;
  border-radius: var(--radius-xs);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--c-muted);
  transition: all 0.2s ease;
}
.tree-ctrl-btn:hover { background: rgba(0,0,0,0.05); color: var(--c-text2); }

.tree-search-wrap { padding: 0.625rem 0.875rem; border-bottom: 1px solid var(--c-border); }
.tree-search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #f1f5f9;
  border-radius: var(--radius-xs);
  padding: 0.5rem 0.75rem;
  transition: all 0.2s ease;
}
.tree-search-box:focus-within { background: #fff; box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
.tree-search-box svg { width: 0.875rem; height: 0.875rem; color: var(--c-muted); flex-shrink: 0; }
.tree-search-box input { border: none; outline: none; background: transparent; font-size: 0.8125rem; color: var(--c-text); width: 100%; }
.tree-search-box input::placeholder { color: var(--c-muted); }

.tree-panel-body { flex: 1; overflow-y: auto; padding: 0.25rem 0; }
.tree-panel-body::-webkit-scrollbar { width: 3px; }
.tree-panel-body::-webkit-scrollbar-track { background: transparent; }
.tree-panel-body::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }

.tree-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
  font-size: 0.8125rem;
  margin: 0.125rem 0.5rem;
  border-radius: var(--radius-xs);
  position: relative;
}
.tree-item:hover { background: rgba(0,0,0,0.025); }
.tree-item.active {
  background: linear-gradient(135deg, rgba(99,102,241,0.08), rgba(99,102,241,0.04));
  box-shadow: inset 3px 0 0 var(--c-accent);
  color: var(--c-accent);
  font-weight: 600;
}

.tree-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--c-muted);
  flex-shrink: 0;
  transition: all 0.2s var(--ease-spring);
}
.tree-dot.active { background: var(--c-accent); box-shadow: 0 0 6px rgba(99,102,241,0.4); }

.tree-toggle {
  width: 1rem;
  height: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--c-muted);
  transition: transform 0.2s ease;
  cursor: pointer;
  border-radius: 4px;
}
.tree-toggle:hover { background: rgba(0,0,0,0.04); }
.tree-toggle.expanded { transform: rotate(90deg); }
.tree-toggle-placeholder { width: 1rem; height: 1rem; flex-shrink: 0; }
.tree-label { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tree-count {
  font-size: 0.625rem;
  color: var(--c-muted);
  background: #f1f5f9;
  padding: 0.0625rem 0.375rem;
  border-radius: 999px;
  font-weight: 600;
  flex-shrink: 0;
}
.tree-item.active .tree-count { background: rgba(99,102,241,0.12); color: var(--c-accent); }

.tree-empty { padding: 2rem 1.25rem; text-align: center; color: var(--c-muted); font-size: 0.8125rem; }

.tree-panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.625rem 1.125rem;
  border-top: 1px solid var(--c-border);
  font-size: 0.75rem;
  color: var(--c-muted);
}
.tree-panel-footer strong { color: var(--c-text); font-weight: 700; }
.tree-footer-detail { font-size: 0.6875rem; }

/* ===== 右侧内容面板 ===== */
.content-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  animation: catSlideRight 0.5s var(--ease-out) both;
  animation-delay: 0.08s;
}

/* 面包屑 */
.breadcrumb-bar {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.625rem 1.125rem;
  background: var(--c-surface);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--c-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  font-size: 0.8125rem;
  flex-wrap: wrap;
  animation: catFadeUp 0.4s var(--ease-out) both;
  animation-delay: 0.15s;
}
.bc-sep { color: #cbd5e1; font-size: 0.75rem; }
.bc-link { color: var(--c-text2); cursor: pointer; transition: color 0.15s ease; font-weight: 500; }
.bc-link:hover { color: var(--c-accent); }
.bc-current { color: var(--c-text); font-weight: 700; }

/* 工具栏 */
.toolbar-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  animation: catFadeUp 0.4s var(--ease-out) both;
  animation-delay: 0.2s;
}
.toolbar-left { display: flex; align-items: center; gap: 0.625rem; }
.toolbar-right { display: flex; align-items: center; gap: 0.75rem; }

.tb-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: var(--radius-xs);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s var(--ease-spring);
  white-space: nowrap;
  line-height: 1.4;
}
.tb-btn--primary {
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.tb-btn--primary:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(99,102,241,0.35); }
.tb-btn--primary:active { transform: translateY(0); }
.tb-btn--secondary {
  background: #fff;
  color: var(--c-text2);
  border: 1px solid #e2e8f0;
}
.tb-btn--secondary:hover { border-color: #cbd5e1; background: #f8fafc; }

/* 批量操作下拉 */
.tb-dropdown { position: relative; }
.tb-dropdown:hover .tb-dropdown-menu { opacity: 1; visibility: visible; transform: translateY(0); }
.tb-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  min-width: 140px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-xs);
  box-shadow: var(--shadow-lg);
  padding: 0.375rem;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-4px);
  transition: all 0.2s var(--ease-out);
  z-index: 20;
}
.tb-dropdown-menu button {
  display: block;
  width: 100%;
  text-align: left;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  color: var(--c-text2);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.12s ease;
}
.tb-dropdown-menu button:hover { background: #f1f5f9; color: var(--c-text); }
.tb-dropdown-menu button.danger { color: var(--c-red); }
.tb-dropdown-menu button.danger:hover { background: #fef2f2; }
.tb-dropdown-divider { height: 1px; background: #f1f5f9; margin: 0.25rem 0; }

.stats-tag {
  font-size: 0.75rem;
  color: var(--c-text2);
  display: flex;
  align-items: center;
  gap: 0.3rem;
  background: rgba(255,255,255,0.7);
  backdrop-filter: blur(8px);
  padding: 0.3rem 0.625rem;
  border-radius: 999px;
  border: 1px solid rgba(0,0,0,0.04);
}
.stats-tag strong { font-weight: 700; color: var(--c-text); }
.stats-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--c-muted); flex-shrink: 0; }
.stats-dot--green { background: var(--c-green); }
.stats-dot--red { background: var(--c-red); }

/* ===== 表格卡片 ===== */
.table-card {
  flex: 1;
  min-height: 0;
  background: var(--c-surface);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--c-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: catFadeUp 0.45s var(--ease-out) both;
  animation-delay: 0.25s;
}

/* 表格表头 */
.table-header {
  display: flex;
  align-items: center;
  padding: 0 0.5rem;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  border-bottom: 1px solid #e2e8f0;
}
.th-cell {
  padding: 0.75rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--c-text2);
  letter-spacing: 0.04em;
  white-space: nowrap;
  text-transform: uppercase;
}
.th-cell--sort { width: 72px; text-align: center; }
.th-cell--name { flex: 1; min-width: 0; padding-left: 0.5rem; }
.th-cell--slug { width: 120px; }
.th-cell--status { width: 120px; }
.th-cell--count { width: 72px; text-align: center; }
.th-cell--showcase { width: 72px; text-align: center; }
.th-cell--action { width: 180px; text-align: right; }

/* 表格体 */
.table-body { flex: 1; overflow-y: auto; min-height: 120px; }
.table-body::-webkit-scrollbar { width: 3px; }
.table-body::-webkit-scrollbar-track { background: transparent; }
.table-body::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }

/* 空状态 */
.table-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  gap: 0.5rem;
}
.table-empty .empty-icon { color: #cbd5e1; }
.table-empty .empty-title { font-size: 0.9375rem; font-weight: 600; color: var(--c-text2); }
.table-empty .empty-hint { font-size: 0.75rem; color: var(--c-muted); }

/* 表格行 */
.table-row {
  display: flex;
  align-items: center;
  padding: 0 0.5rem;
  border-bottom: 1px solid #f1f5f9;
  transition: all 0.15s ease;
  animation: catRowIn 0.35s var(--ease-out) both;
  animation-delay: calc(0.15s + var(--idx, 0) * 0.04s);
}
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: rgba(99,102,241,0.03); }

.td-cell {
  padding: 0.625rem 0.75rem;
  font-size: 0.8125rem;
  color: var(--c-text);
  display: flex;
  align-items: center;
}
.td-cell--sort { width: 72px; justify-content: center; }
.td-cell--name { flex: 1; min-width: 0; padding-left: 0.5rem; gap: 0.5rem; }
.td-cell--slug { width: 120px; }
.td-cell--status { width: 120px; gap: 0.5rem; }
.td-cell--count { width: 72px; justify-content: center; }
.td-cell--showcase { width: 72px; justify-content: center; }
.td-cell--action { width: 180px; justify-content: flex-end; gap: 0; }

/* 排序输入框 */
.sort-input {
  width: 2.75rem;
  padding: 0.25rem 0.25rem;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 600;
  text-align: center;
  background: #fff;
  color: var(--c-text);
  transition: all 0.2s ease;
  outline: none;
  -moz-appearance: textfield;
}
.sort-input::-webkit-inner-spin-button,
.sort-input::-webkit-outer-spin-button { -webkit-appearance: none; margin: 0; }
.sort-input:hover { border-color: var(--c-muted); }
.sort-input:focus { border-color: var(--c-accent); box-shadow: 0 0 0 2px rgba(99,102,241,0.15); }

/* 分类名称 */
.cat-avatar {
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-xs);
  background: linear-gradient(135deg, #f1f5f9, #e2e8f0);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  flex-shrink: 0;
}
.cat-name-text { font-weight: 600; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.child-chip {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--c-accent);
  background: rgba(99,102,241,0.08);
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all 0.15s ease;
}
.child-chip:hover { background: rgba(99,102,241,0.15); transform: scale(1.05); }

/* Slug 标签 */
.slug-tag {
  font-size: 0.6875rem;
  color: var(--c-muted);
  background: #f8fafc;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}

/* 状态切换开关 */
.switcher {
  width: 2rem;
  height: 1.125rem;
  border-radius: 999px;
  background: #cbd5e1;
  border: none;
  position: relative;
  cursor: pointer;
  transition: all 0.25s var(--ease-spring);
  padding: 0;
  flex-shrink: 0;
}
.switcher.on { background: var(--c-green); box-shadow: 0 0 8px rgba(16,185,129,0.25); }
.switcher-knob {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 0.875rem;
  height: 0.875rem;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.25s var(--ease-spring);
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}
.switcher.on .switcher-knob { transform: translateX(0.875rem); }
.switcher-label { font-size: 0.6875rem; color: var(--c-muted); font-weight: 500; }
.switcher-label.on { color: var(--c-green); }

/* 商品数量 */
.count-badge {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--c-text2);
  background: #f1f5f9;
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
}

/* 星级展示按钮 */
.star-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border: none;
  background: transparent;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s var(--ease-spring);
  padding: 0;
  line-height: 1;
}
.star-btn:hover { background: #f1f5f9; transform: scale(1.15); }
.star-btn.active:hover { animation: starPulse 0.4s var(--ease-spring); }

/* 操作链接 */
.act-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.act-link--edit { color: var(--c-accent); }
.act-link--edit:hover { background: rgba(99,102,241,0.08); }
.act-link--add { color: #059669; }
.act-link--add:hover { background: rgba(5,150,105,0.08); }
.act-link--del { color: var(--c-red); }
.act-link--del:hover { background: #fef2f2; }
.act-divider { width: 1px; height: 1rem; background: #e2e8f0; margin: 0 0.125rem; }

/* 表格底部 */
.table-footer {
  display: flex;
  align-items: center;
  padding: 0.625rem 1.125rem;
  border-top: 1px solid var(--c-border);
  font-size: 0.75rem;
  color: var(--c-muted);
  gap: 0.5rem;
}

/* ===== 弹窗覆盖层 ===== */
.dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0,0,0,0.35);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: overlayIn 0.2s ease;
}
.dialog-panel {
  background: #fff;
  border-radius: 18px;
  box-shadow: var(--shadow-xl);
  width: 520px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  animation: dialogIn 0.3s var(--ease-out);
  overflow: hidden;
}
.dialog-panel--sm { width: 400px; }

.dialog-head {
  position: relative;
  padding: 1.5rem 1.75rem 1rem;
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
}
.dialog-close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 2rem;
  height: 2rem;
  border: none;
  background: #f1f5f9;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--c-muted);
  transition: all 0.15s ease;
}
.dialog-close:hover { background: #e2e8f0; color: var(--c-text); }
.dialog-title { font-size: 1.0625rem; font-weight: 700; color: var(--c-text); }
.dialog-sub { font-size: 0.75rem; color: var(--c-muted); margin-top: 0.25rem; }

.dialog-body {
  padding: 1.25rem 1.75rem;
  overflow-y: auto;
  flex: 1;
}
.dialog-body::-webkit-scrollbar { width: 3px; }
.dialog-body::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }

.dialog-foot {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.625rem;
  padding: 1rem 1.75rem;
  border-top: 1px solid #f1f5f9;
  flex-shrink: 0;
}

/* 弹窗按钮 */
.d-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1.25rem;
  border: none;
  border-radius: var(--radius-xs);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s var(--ease-spring);
  line-height: 1.4;
}
.d-btn--cancel { background: #f1f5f9; color: var(--c-text2); }
.d-btn--cancel:hover { background: #e2e8f0; }
.d-btn--primary {
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.d-btn--primary:hover { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(99,102,241,0.35); }
.d-btn--danger { background: var(--c-red); color: #fff; }
.d-btn--danger:hover { background: #dc2626; }

/* 弹窗表单字段 */
.d-field { margin-bottom: 1rem; position: relative; }
.d-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--c-text2);
  margin-bottom: 0.375rem;
}
.d-required { color: var(--c-red); }
.d-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-xs);
  font-size: 0.8125rem;
  color: var(--c-text);
  background: #fff;
  outline: none;
  transition: all 0.2s ease;
  box-sizing: border-box;
}
.d-input:hover { border-color: var(--c-muted); }
.d-input:focus { border-color: var(--c-accent); box-shadow: 0 0 0 2px rgba(99,102,241,0.15); }
.d-input::placeholder { color: var(--c-muted); }
.d-counter {
  position: absolute;
  right: 0.625rem;
  bottom: 0.5rem;
  font-size: 0.625rem;
  color: var(--c-muted);
  pointer-events: none;
}
.d-hint { font-size: 0.625rem; color: var(--c-muted); margin-top: 0.25rem; display: block; }

.d-row { display: flex; gap: 0.75rem; }

.d-select-wrap { position: relative; }
.d-select {
  width: 100%;
  padding: 0.5rem 2rem 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-xs);
  font-size: 0.8125rem;
  color: var(--c-text);
  background: #fff;
  outline: none;
  appearance: none;
  -webkit-appearance: none;
  cursor: pointer;
  transition: all 0.2s ease;
  box-sizing: border-box;
}
.d-select:hover { border-color: var(--c-muted); }
.d-select:focus { border-color: var(--c-accent); box-shadow: 0 0 0 2px rgba(99,102,241,0.15); }
.d-select-arrow { position: absolute; right: 0.75rem; top: 50%; transform: translateY(-50%); color: var(--c-muted); pointer-events: none; }

.d-textarea {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-xs);
  font-size: 0.8125rem;
  color: var(--c-text);
  background: #fff;
  outline: none;
  resize: vertical;
  transition: all 0.2s ease;
  box-sizing: border-box;
  font-family: inherit;
}
.d-textarea:hover { border-color: var(--c-muted); }
.d-textarea:focus { border-color: var(--c-accent); box-shadow: 0 0 0 2px rgba(99,102,241,0.15); }
.d-textarea::placeholder { color: var(--c-muted); }

/* 图标选择器 */
.icon-picker-top {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.icon-current {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-xs);
  background: #f8fafc;
  font-size: 1.4rem;
  overflow: hidden;
}
.icon-current svg { width: 24px; height: 24px; color: var(--c-accent); }
.icon-current img { width: 100%; height: 100%; object-fit: contain; }
.icon-tabs {
  display: flex;
  gap: 4px;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 3px;
}
.icon-tab {
  border: none;
  background: transparent;
  padding: 4px 14px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--c-text2);
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}
.icon-tab.active { background: #fff; color: var(--c-accent); box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.icon-panel {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  max-height: 132px;
  overflow-y: auto;
  padding: 4px;
}
.icon-panel::-webkit-scrollbar { width: 2px; }
.icon-panel::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }
.icon-opt {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border: 1px solid #f1f5f9;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  padding: 0;
  line-height: 1;
  transition: all 0.15s var(--ease-spring);
  color: var(--c-text2);
}
.icon-opt svg { width: 20px; height: 20px; }
.icon-opt:hover { border-color: var(--c-accent); color: var(--c-accent); background: rgba(99,102,241,0.06); transform: scale(1.12); z-index: 1; }
.icon-opt.selected { border-color: var(--c-accent); color: var(--c-accent); background: rgba(99,102,241,0.08); box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
.icon-upload-zone {
  border: 2px dashed #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 14px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f8fafc;
  color: var(--c-text2);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
}
.icon-upload-zone:hover { border-color: var(--c-accent); background: rgba(99,102,241,0.04); }
.icon-upload-zone small { font-size: 0.625rem; color: var(--c-muted); }
.icon-upload-preview { max-width: 80px; max-height: 80px; border-radius: var(--radius-xs); object-fit: contain; }

/* 上传区域 */
.upload-zone {
  border: 2px dashed #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 1.25rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f8fafc;
  max-width: 220px;
  color: var(--c-muted);
}
.upload-zone:hover { border-color: var(--c-accent); background: rgba(99,102,241,0.04); }
.upload-zone .upload-text { font-size: 0.75rem; color: var(--c-text2); margin-top: 0.375rem; }
.upload-zone .upload-hint { font-size: 0.625rem; color: var(--c-muted); margin-top: 0.25rem; }
.upload-preview { max-width: 100%; max-height: 120px; border-radius: var(--radius-xs); }

/* SEO 折叠 */
.d-seo { border: 1px solid #e2e8f0; border-radius: var(--radius-xs); overflow: hidden; }
.d-seo-toggle {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: none;
  background: #f8fafc;
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--c-text);
  transition: background 0.15s ease;
}
.d-seo-toggle:hover { background: #f1f5f9; }
.d-seo-arrow { margin-left: auto; transition: transform 0.2s ease; color: var(--c-muted); }
.d-seo-arrow.open { transform: rotate(180deg); }
.d-seo-body { padding: 1rem; border-top: 1px solid #e2e8f0; }
.d-seo-body .d-field:last-child { margin-bottom: 0; }

/* AI 生成 SEO 按钮行 */
.ai-seo-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #f0f7ff 0%, #e8f4fd 100%);
  border-radius: 10px;
  border: 1px solid rgba(99, 102, 241, 0.15);
}
.ai-seo-hint {
  font-size: 0.75rem;
  color: var(--c-muted);
}

/* 删除确认 */
.del-icon { font-size: 2.5rem; margin-bottom: 0.75rem; }
.del-title { font-size: 1rem; font-weight: 700; color: var(--c-text); margin-bottom: 0.375rem; }
.del-msg { font-size: 0.8125rem; color: var(--c-text2); line-height: 1.5; }

/* ===== 响应式 ===== */
@media (max-width: 1024px) {
  .categories-layout { flex-direction: column; }
  .tree-panel { width: 100%; min-width: 0; max-height: 320px; }
  .content-panel { width: 100%; }
}
@media (max-width: 768px) {
  .toolbar-bar { flex-direction: column; align-items: stretch; }
  .toolbar-left { flex-wrap: wrap; }
  .toolbar-right { justify-content: flex-start; }
  .d-row { flex-direction: column; gap: 0; }
  .d-field[style*="width: 140px"] { width: 100% !important; }
  .emoji-scroll { max-height: 80px; }
  .dialog-panel { width: calc(100vw - 2rem); margin: 1rem; max-height: 90vh; }
  .th-cell--slug, .td-cell--slug { display: none; }
  .th-cell--count, .td-cell--count { display: none; }
}
@media (max-width: 480px) {
  .tree-panel-header .tree-ctrl-btns { display: none; }
  .th-cell--showcase, .td-cell--showcase { display: none; }
  .stats-tag { font-size: 0.6875rem; padding: 0.2rem 0.5rem; }
}
</style>
