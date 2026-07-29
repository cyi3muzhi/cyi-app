<template>
  <!-- Banner 管理页面 -->
  <div class="banners-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">Banner 管理</h2>
        <span class="page-subtitle">管理首页轮播图展示、模板与排序</span>
      </div>
      <div class="header-right">
        <el-button @click="loadBanners" :disabled="loading" class="action-btn">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveAll" class="action-btn primary-btn">
          <el-icon><Check /></el-icon>
          保存全部
        </el-button>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="toolbar-bar">
      <div class="toolbar-left">
        <el-button type="primary" @click="handleAdd" class="add-btn">
          <el-icon><Plus /></el-icon>
          新增 Banner
        </el-button>
      </div>
      <div class="toolbar-right">
        <span class="stats-tag">
          <span class="stats-dot" />
          共 <strong>{{ banners.length }}</strong> 个
        </span>
        <span class="stats-tag">
          <span class="stats-dot stats-dot--green" />
          启用 <strong>{{ enabledCount }}</strong>
        </span>
        <span class="stats-tag">
          <span class="stats-dot stats-dot--red" />
          禁用 <strong>{{ disabledCount }}</strong>
        </span>
      </div>
    </div>

    <!-- ===== 网站预览模拟器 ===== -->
    <div v-if="enabledBanners.length > 0" class="site-preview">
      <div class="preview-header">
        <div class="preview-header-left">
          <el-icon><VideoCameraFilled /></el-icon>
          <span>首页轮播效果预览</span>
        </div>
        <div class="preview-header-right">
          <span class="preview-slide-count">{{ previewSlide + 1 }} / {{ enabledBanners.length }}</span>
          <el-tag size="small" effect="plain" round>{{ getTemplateInfo(enabledBanners[previewSlide]?.template).name }}</el-tag>
        </div>
      </div>
      <div class="preview-browser">
        <!-- 浏览器模拟顶栏 -->
        <div class="browser-bar">
          <div class="browser-dots">
            <span class="browser-dot browser-dot--red" />
            <span class="browser-dot browser-dot--yellow" />
            <span class="browser-dot browser-dot--green" />
          </div>
          <div class="browser-url">
            <el-icon :size="12"><Link /></el-icon>
            <span>yuexuan.shop</span>
          </div>
          <div class="browser-spacer" />
        </div>
        <!-- 浏览器内容区 -->
        <div class="browser-content">
          <div
            class="preview-track"
            :style="{ transform: `translateX(-${previewSlide * 100}%)` }"
          >
            <div
              v-for="(b, i) in enabledBanners"
              :key="i"
              class="preview-slide"
              :class="'ps--' + (b.template || 'classic')"
              :style="getSlideStyle(b)"
            >
              <div class="ps-overlay" v-if="(b.template || 'classic') === 'centered'" />
              <div class="ps-bg-img" v-if="b.image && (b.template || 'classic') === 'full-image'">
                <img :src="b.image" alt="" @error="(e: any) => e.target.style.display='none'" />
              </div>
              <div class="ps-body">
                <div class="ps-text">
                  <h2 class="ps-title" :style="{ color: b.text_color || '#ffffff' }">{{ b.title }}</h2>
                  <p class="ps-subtitle" v-if="b.subtitle" :style="{ color: b.text_color || '#ffffff' }">{{ b.subtitle }}</p>
                  <span class="ps-btn" v-if="b.btn_text && (b.template || 'classic') !== 'minimal'">{{ b.btn_text }}</span>
                </div>
                <div class="ps-image" v-if="b.image && (b.template || 'classic') !== 'full-image' && (b.template || 'classic') !== 'minimal'">
                  <img :src="b.image" alt="" @error="(e: any) => e.target.style.display='none'" />
                </div>
                <div class="ps-image-placeholder" v-else-if="(b.template || 'classic') !== 'full-image' && (b.template || 'classic') !== 'minimal' && !b.image">
                  <el-icon :size="32" color="rgba(255,255,255,0.2)"><Picture /></el-icon>
                </div>
              </div>
              <div class="ps-accent" v-if="(b.template || 'classic') === 'minimal'" />
            </div>
          </div>
          <!-- 指示器圆点 -->
          <div class="preview-dots">
            <button
              v-for="(_, i) in enabledBanners"
              :key="i"
              class="preview-dot"
              :class="{ 'preview-dot--active': previewSlide === i }"
              @click="previewSlide = i"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Banner 列表卡片 -->
    <div class="banner-list" v-loading="loading">
      <!-- 空状态 -->
      <div v-if="banners.length === 0 && !loading" class="empty-state">
        <el-icon :size="64" color="#d0d4dc"><PictureFilled /></el-icon>
        <div class="empty-title">暂无 Banner，点击「新增 Banner」创建</div>
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>新增 Banner
        </el-button>
      </div>

      <!-- Banner 卡片列表 -->
      <transition-group name="list-stagger" tag="div" class="banner-cards">
        <div
          v-for="(banner, index) in banners"
          :key="banner.id"
          class="banner-card"
          :style="{ animationDelay: `${index * 0.05}s` }"
        >
          <!-- 序号 -->
          <div class="card-index">{{ index + 1 }}</div>

          <!-- Banner 预览 -->
          <div class="card-preview">
            <div
              class="card-banner"
              :class="'cb--' + (banner.template || 'classic')"
              :style="{ background: banner.bg_color || '#409eff' }"
            >
              <div class="cb-overlay" v-if="(banner.template || 'classic') === 'centered'" />
              <img
                v-if="banner.image && (banner.template || 'classic') === 'full-image'"
                :src="banner.image" alt=""
                class="cb-bg-img"
                @error="(e: any) => e.target.style.display='none'"
              />
              <div class="cb-body">
                <div class="cb-text">
                  <div class="cb-title" :style="{ color: banner.text_color || '#ffffff' }">{{ banner.title || 'Banner 标题' }}</div>
                  <div class="cb-subtitle" v-if="banner.subtitle" :style="{ color: banner.text_color || '#ffffff' }">{{ banner.subtitle }}</div>
                  <div class="cb-btn" v-if="banner.btn_text && (banner.template || 'classic') !== 'minimal'">{{ banner.btn_text }}</div>
                </div>
                <div class="cb-img" v-if="banner.image && (banner.template || 'classic') !== 'full-image' && (banner.template || 'classic') !== 'minimal'">
                  <img :src="banner.image" alt="" @error="(e: any) => e.target.style.display='none'" />
                </div>
                <div class="cb-placeholder" v-else-if="(banner.template || 'classic') !== 'full-image' && (banner.template || 'classic') !== 'minimal' && !banner.image">
                  <el-icon :size="18" color="rgba(255,255,255,0.25)"><Picture /></el-icon>
                </div>
              </div>
              <div class="cb-accent" v-if="(banner.template || 'classic') === 'minimal'" />
              <div class="cb-tag">{{ getTemplateInfo(banner.template).name }}</div>
            </div>
          </div>

          <!-- 信息区 -->
          <div class="card-info">
            <div class="card-info-row">
              <span class="card-label">标题</span>
              <span class="card-value">{{ banner.title || '-' }}</span>
            </div>
            <div class="card-info-row">
              <span class="card-label">按钮</span>
              <span class="card-value">{{ banner.btn_text || '-' }}</span>
            </div>
            <div class="card-info-row">
              <span class="card-label">链接</span>
              <span class="card-value card-value--link">{{ banner.link || '-' }}</span>
            </div>
          </div>

          <!-- 状态 & 操作 -->
          <div class="card-actions">
            <div class="card-toggle">
              <el-switch
                :model-value="banner.enabled"
                @update:model-value="val => toggleEnabled(index, val)"
                size="small"
                active-text="启用"
                inactive-text="禁用"
              />
            </div>
            <div class="card-buttons">
              <el-tooltip content="上移" placement="top">
                <el-button text size="small" :disabled="index === 0" @click="moveUp(index)" class="card-btn">
                  <el-icon><CaretTop /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="下移" placement="top">
                <el-button text size="small" :disabled="index === banners.length - 1" @click="moveDown(index)" class="card-btn">
                  <el-icon><CaretBottom /></el-icon>
                </el-button>
              </el-tooltip>
              <el-divider direction="vertical" />
              <el-tooltip content="编辑" placement="top">
                <el-button text type="primary" size="small" @click="handleEdit(index)" class="card-btn">
                  <el-icon><EditPen /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button text type="danger" size="small" @click="handleDelete(index)" class="card-btn">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-tooltip>
            </div>
          </div>
        </div>
      </transition-group>
    </div>

    <!-- ===== 编辑弹窗 ===== -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑 Banner' : '新增 Banner'"
      width="820px"
      :close-on-click-modal="false"
      destroy-on-close
      append-to-body
      class="banner-dialog"
      top="5vh"
    >
      <!-- 顶部模板选择器 -->
      <div class="template-picker">
        <div class="template-picker-label">选择模板</div>
        <div class="template-list">
          <div
            v-for="tpl in templateList"
            :key="tpl.key"
            class="template-item"
            :class="{ 'template-item--active': (editForm.template || 'classic') === tpl.key }"
            @click="selectTemplate(tpl.key)"
          >
            <!-- 模板缩略图预览 -->
            <div class="template-thumb" :class="'tthumb--' + tpl.key">
              <div class="tthumb-body">
                <div class="tthumb-title">T</div>
                <div class="tthumb-sub" v-if="tpl.key !== 'minimal'">D</div>
                <div class="tthumb-btn" v-if="tpl.key !== 'minimal'">B</div>
              </div>
              <div class="tthumb-img" v-if="tpl.key === 'classic' || tpl.key === 'centered'" />
              <div class="tthumb-accent" v-if="tpl.key === 'minimal'" />
            </div>
            <div class="template-name">{{ tpl.name }}</div>
          </div>
        </div>
      </div>

      <div class="dialog-body">
        <!-- 顶部：实时预览（全宽） -->
        <div class="dialog-preview-top">
          <div class="preview-top-label">实时预览</div>
          <div
            class="live-preview"
            :class="'template--' + (editForm.template || 'classic')"
            :style="getLivePreviewStyle(editForm)"
          >
            <div class="live-preview-overlay" v-if="(editForm.template || 'classic') === 'centered'" />
            <div class="live-preview-img-bg" v-if="editForm.image && (editForm.template || 'classic') === 'full-image'">
              <img :src="editForm.image" alt="" @error="(e: any) => e.target.style.display='none'" />
            </div>
            <div class="live-preview-body">
              <div class="live-preview-text">
                <div class="live-preview-title" :style="{ color: editForm.text_color || '#ffffff' }">{{ editForm.title || 'Banner 标题' }}</div>
                <div class="live-preview-subtitle" v-if="editForm.subtitle" :style="{ color: editForm.text_color || '#ffffff' }">{{ editForm.subtitle }}</div>
                <div class="live-preview-btn" v-if="editForm.btn_text && (editForm.template || 'classic') !== 'minimal'">
                  {{ editForm.btn_text }}
                </div>
              </div>
              <div class="live-preview-image" v-if="editForm.image && (editForm.template || 'classic') !== 'full-image' && (editForm.template || 'classic') !== 'minimal'">
                <img :src="editForm.image" alt="preview" @error="(e: any) => e.target.style.display='none'" />
              </div>
              <div class="live-preview-img-placeholder" v-else-if="(editForm.template || 'classic') !== 'full-image' && (editForm.template || 'classic') !== 'minimal'">
                <el-icon :size="28" color="rgba(255,255,255,0.25)"><Picture /></el-icon>
              </div>
            </div>
            <div class="live-minimal-accent" v-if="(editForm.template || 'classic') === 'minimal'" />
          </div>
          <div class="preview-template-tag">
            <el-tag size="small" :type="templateTagType" effect="plain">
              {{ getTemplateInfo(editForm.template).name }}
            </el-tag>
          </div>
        </div>

        <!-- 下方：表单 -->
        <div class="dialog-form-area">
          <el-form
            ref="formRef"
            :model="editForm"
            :rules="formRules"
            label-width="80px"
            size="default"
          >
            <!-- 基本信息 -->
            <el-card shadow="never" class="form-card">
              <template #header>
                <span class="section-title">基本信息</span>
              </template>
              <!-- AI 生成文案按钮 -->
              <div class="ai-generate-row">
                <el-button type="primary" size="small" :loading="aiLoading" @click="handleAIGenerate">
                  AI 生成文案
                </el-button>
                <span class="ai-generate-hint">基于当前标题智能生成文案方案</span>
              </div>
              <el-form-item label="标题" prop="title">
                <el-input v-model="editForm.title" placeholder="请输入 Banner 标题" maxlength="50" type="textarea" :rows="2" show-word-limit />
              </el-form-item>
              <el-form-item label="副标题" prop="subtitle">
                <el-input v-model="editForm.subtitle" placeholder="请输入 Banner 副标题（可选）" maxlength="100" type="textarea" :rows="2" />
              </el-form-item>
              <el-form-item label="按钮文字">
                <el-input v-model="editForm.btn_text" placeholder="如：立即查看、了解详情（可选）" maxlength="20" />
              </el-form-item>
              <el-form-item label="文字颜色">
                <div class="color-picker-wrap">
                  <el-color-picker v-model="editForm.text_color" :predefine="['#ffffff', '#000000', '#1e293b', '#f5222d', '#409eff', '#52c41a', '#fa8c16']" />
                  <span class="color-hex">{{ editForm.text_color || '#ffffff' }}</span>
                  <el-input
                    v-model="editForm.text_color"
                    placeholder="#ffffff"
                    maxlength="7"
                    style="width: 100px;"
                    size="small"
                  />
                </div>
              </el-form-item>
            </el-card>

            <!-- 图片与链接 -->
            <el-card shadow="never" class="form-card">
              <template #header>
                <span class="section-title">图片与链接</span>
              </template>
              <el-form-item label="图片" prop="image" v-if="(editForm.template || 'classic') !== 'minimal'">
                <div class="upload-row">
                  <el-input v-model="editForm.image" placeholder="图片 URL 或点击上传">
                    <template #append>
                      <el-button @click="triggerUpload" :loading="uploading" class="upload-btn">
                        <el-icon v-if="!uploading"><Upload /></el-icon>
                        {{ uploading ? '上传中' : '上传' }}
                      </el-button>
                    </template>
                  </el-input>
                </div>
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="image/*"
                  style="display: none"
                  @change="handleFileSelect"
                />
                <div v-if="previewImage" class="image-preview-box" :style="{ backgroundColor: editForm.bg_color || '#f0f2f5' }">
                  <img :src="previewImage" alt="预览" @error="onPreviewImgError" />
                </div>
              </el-form-item>
              <el-form-item label="跳转链接" prop="link">
                <el-input v-model="editForm.link" placeholder="如：/pages/product/detail?id=123">
                  <template #prefix>
                    <el-icon><Link /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-card>

            <!-- 样式设置 -->
            <el-card shadow="never" class="form-card">
              <template #header>
                <span class="section-title">样式设置</span>
              </template>
              <!-- 配色方案 -->
              <el-form-item label="配色方案">
                <div class="color-schemes">
                  <div
                    v-for="cs in colorSchemes"
                    :key="cs.key"
                    class="color-scheme-item"
                    :class="{ 'color-scheme-item--active': editForm.bg_color === cs.bg_color }"
                    @click="editForm.bg_color = cs.bg_color"
                  >
                    <div class="cs-preview" :style="{ background: cs.bg_color }">
                      <el-icon v-if="editForm.bg_color === cs.bg_color" class="cs-check"><Check /></el-icon>
                    </div>
                    <span class="cs-name">{{ cs.name }}</span>
                  </div>
                </div>
              </el-form-item>
              <el-form-item label="自定义颜色">
                <div class="color-picker-wrap">
                  <el-color-picker v-model="editForm.bg_color" :predefine="presetColors" />
                  <span class="color-hex">{{ editForm.bg_color || '未选择' }}</span>
                  <el-input
                    v-model="editForm.bg_color"
                    placeholder="#ffffff"
                    maxlength="7"
                    style="width: 110px; margin-left: 6px;"
                    size="small"
                  />
                </div>
              </el-form-item>
              <el-form-item label="显示状态">
                <el-switch
                  v-model="editForm.enabled"
                  active-text="启用"
                  inactive-text="禁用"
                  inline-prompt
                />
              </el-form-item>
            </el-card>
          </el-form>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="confirmEdit">
            {{ isEditing ? '保存修改' : '添加 Banner' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * Banner 管理页面
 * 展示所有 Banner 列表，支持自定义模板、配色方案、动效预览
 */
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Refresh, Check, EditPen, Delete,
  CaretTop, CaretBottom, Picture, PictureFilled,
  Link, Upload, VideoCameraFilled
} from '@element-plus/icons-vue'
import { listBanners, saveBanners } from '../../api/banner'
import type { BannerItem } from '../../api/banner'
import { uploadToOSS } from '../../utils/upload'
import { generateBannerCopy, type BannerCopyOption } from '../../api/agent'

/* ===== 模板定义 ===== */

/** 模板信息接口 */
interface TemplateInfo {
  key: string
  name: string
  description: string
}

/** 内置模板列表 */
const templateList: TemplateInfo[] = [
  { key: 'classic', name: '经典版', description: '左侧文字 + 右侧图片' },
  { key: 'full-image', name: '全图版', description: '图片全背景，文字浮在图上' },
  { key: 'centered', name: '居中版', description: '文字居中，图片半透明叠加' },
  { key: 'dark', name: '深色版', description: '深色渐变背景，白色文字' },
  { key: 'minimal', name: '极简版', description: '简洁线条 + 标题，无图片' }
]

/** 获取模板信息 */
function getTemplateInfo(templateKey?: string): TemplateInfo {
  return templateList.find(t => t.key === (templateKey || 'classic')) || templateList[0]
}

/** 模板对应的 el-tag 类型 */
const templateTagMap: Record<string, string> = {
  classic: 'primary',
  'full-image': 'success',
  centered: 'warning',
  dark: 'info',
  minimal: ''
}

const templateTagType = computed(() => (templateTagMap[editForm.value.template || 'classic'] as any) || 'primary')

/* ===== 配色方案 ===== */

/** 配色方案接口 */
interface ColorScheme {
  key: string
  name: string
  bg_color: string
}

/** 预制配色方案 */
const colorSchemes: ColorScheme[] = [
  { key: 'blue', name: '活力蓝', bg_color: '#409eff' },
  { key: 'green', name: '森林绿', bg_color: '#52c41a' },
  { key: 'orange', name: '暖阳橙', bg_color: '#fa8c16' },
  { key: 'pink', name: '玫瑰粉', bg_color: '#eb2f96' },
  { key: 'purple', name: '星空紫', bg_color: '#722ed1' },
  { key: 'sky', name: '天空蓝', bg_color: '#13c2c2' },
  { key: 'red', name: '朱砂红', bg_color: '#f5222d' },
  { key: 'slate', name: '芝麻灰', bg_color: '#5a6872' },
  { key: 'dark', name: '暗夜黑', bg_color: '#1a1a2e' },
  { key: 'emerald', name: '薄荷绿', bg_color: '#08979c' },
  { key: 'indigo', name: '靛青蓝', bg_color: '#2f54eb' },
  { key: 'gold', name: '香槟金', bg_color: '#d48806' }
]

/**
 * 规范化 Banner 字段名
 * proto JSON 使用驼峰命名（如 bgColor），但前端接口定义使用蛇形命名（如 bg_color）
 */
function normalizeBanner(b: any): BannerItem {
  return {
    id: b.id ?? 0,
    title: b.title ?? '',
    subtitle: b.subtitle ?? '',
    image: b.image ?? '',
    link: b.link ?? '',
    sort_order: b.sort_order ?? b.sortOrder ?? 0,
    enabled: b.enabled ?? false,
    bg_color: b.bg_color ?? b.bgColor ?? '#409eff',
    btn_text: b.btn_text ?? b.btnText ?? '',
    template: b.template ?? b.template ?? 'classic',
    text_color: b.text_color ?? b.textColor ?? '#ffffff'
  }
}

/* ===== 模板渲染辅助函数 ===== */

/**
 * 获取轮播预览样式
 * 根据模板类型返回不同的背景样式
 */
function getSlideStyle(banner: BannerItem): Record<string, string> {
  const tpl = banner.template || 'classic'
  const bg = banner.bg_color || '#409eff'
  if (tpl === 'dark') {
    return { background: `linear-gradient(135deg, ${bg}, #0f172a)` }
  }
  if (tpl === 'full-image' && banner.image) {
    return { background: bg }
  }
  return { background: bg }
}

/**
 * 获取卡片迷你预览样式
 */
function getMiniPreviewStyle(banner: BannerItem): Record<string, string> {
  const tpl = banner.template || 'classic'
  const bg = banner.bg_color || '#409eff'
  if (tpl === 'dark') {
    return { background: `linear-gradient(135deg, ${bg}, #0f172a)` }
  }
  if (tpl === 'full-image' && banner.image) {
    return { background: bg }
  }
  return { background: bg }
}

/**
 * 获取弹窗实时预览样式
 */
function getLivePreviewStyle(banner: BannerItem): Record<string, string> {
  const tpl = banner.template || 'classic'
  const bg = banner.bg_color || '#409eff'
  if (tpl === 'dark') {
    return { background: `linear-gradient(135deg, ${bg}, #0f172a)` }
  }
  if (tpl === 'full-image' && banner.image) {
    return { background: bg }
  }
  return { background: bg }
}

/* ========== 状态 ========== */

const banners = ref<BannerItem[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingIndex = ref(-1)
const submitting = ref(false)
const formRef = ref<any>(null)
const previewImage = ref('')
const uploading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
/** AI 生成文案 loading */
const aiLoading = ref(false)

/* ========== 计算属性 ========== */

const enabledCount = computed(() => banners.value.filter(b => b.enabled).length)
const disabledCount = computed(() => banners.value.filter(b => !b.enabled).length)

/** 已启用的 Banner（用于轮播预览） */
const enabledBanners = computed(() =>
  banners.value.filter(b => b.enabled).sort((a, b) => a.sort_order - b.sort_order)
)

const previewSlide = ref(0)
let previewTimer: ReturnType<typeof setInterval> | null = null

/** 启动/重置轮播预览 */
function resetPreviewTimer() {
  if (previewTimer) clearInterval(previewTimer)
  if (enabledBanners.value.length <= 1) return
  previewTimer = setInterval(() => {
    previewSlide.value = (previewSlide.value + 1) % enabledBanners.value.length
  }, 4000)
}

watch(() => banners.value.length, () => { previewSlide.value = 0; resetPreviewTimer() })

/* ========== 表单 ========== */

const presetColors = [
  '#409eff', '#52c41a', '#fa8c16', '#eb2f96',
  '#722ed1', '#13c2c2', '#f5222d', '#5a6872',
  '#1a1a2e', '#08979c', '#2f54eb', '#d48806',
  '#ffffff', '#f5f5f5'
]

const editForm = ref<BannerItem>({
  id: 0,
  title: '',
  subtitle: '',
  image: '',
  link: '',
  sort_order: 0,
  enabled: true,
  bg_color: '#409eff',
  btn_text: '',
  template: 'classic',
  text_color: '#ffffff'
})

const formRules = {
  title: [
    { required: true, message: '请输入 Banner 标题', trigger: 'blur' },
    { max: 50, message: '标题不超过 50 个字符', trigger: 'blur' }
  ],
  image: [
    { required: true, message: '请输入图片地址', trigger: 'blur' }
  ]
}

/* ========== 方法 ========== */

/**
 * 选择模板
 * 切换模板时联动更新表单和预览
 */
function selectTemplate(key: string) {
  editForm.value.template = key
  // 极简版不需要图片和按钮文字
  if (key === 'minimal') {
    editForm.value.btn_text = ''
  }
}

/**
 * 加载 Banner 列表
 */
async function loadBanners() {
  loading.value = true
  try {
    const res = await listBanners()
    banners.value = (res.banners || []).map(normalizeBanner).sort((a: any, b: any) => a.sort_order - b.sort_order)
    resetPreviewTimer()
  } catch {
    ElMessage.error('获取 Banner 列表失败')
  } finally {
    loading.value = false
  }
}

function triggerUpload() {
  fileInputRef.value?.click()
}

/**
 * 处理文件选择并上传到 OSS
 */
async function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const result = await uploadToOSS(file, 'banner')
    editForm.value.image = result.publicUrl
    previewImage.value = result.publicUrl
    ElMessage.success('图片上传成功')
  } catch {
    ElMessage.error('图片上传失败，请重试')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function onImgError(event: Event) {
  const target = event.target as HTMLImageElement
  target.style.display = 'none'
}

function onPreviewImgError() {
  ElMessage.warning('图片加载失败，请检查 URL 是否正确')
}

/**
 * AI 生成文案方案
 * 调用 AI 接口生成多组文案，用户选择后填入表单
 */
async function handleAIGenerate() {
  const topic = editForm.value.title
  if (!topic) {
    ElMessage.warning('请先输入 Banner 标题作为生成主题')
    return
  }
  aiLoading.value = true
  try {
    const options = await generateBannerCopy(topic, editForm.value.subtitle)
    if (options.length === 0) {
      ElMessage.warning('AI 生成失败，请重试')
      return
    }
    if (options.length === 1) {
      editForm.value.title = options[0].title
      editForm.value.subtitle = options[0].subtitle
      editForm.value.btn_text = options[0].btn_text
      ElMessage.success('文案已生成')
      return
    }
    // 多组方案时，让用户选择
    const msg = options.map((opt, i) =>
      `方案 ${i + 1}：\n标题：${opt.title}\n副标题：${opt.subtitle}\n按钮：${opt.btn_text}\n`
    ).join('\n')
    const { value } = await ElMessageBox.prompt(msg + '\n请输入方案编号：', '选择文案方案', {
      inputValue: '1',
      inputPlaceholder: '请输入 1~' + options.length,
      confirmButtonText: '确认',
      cancelButtonText: '取消'
    })
    const idx = parseInt(value) - 1
    if (idx >= 0 && idx < options.length) {
      editForm.value.title = options[idx].title
      editForm.value.subtitle = options[idx].subtitle
      editForm.value.btn_text = options[idx].btn_text
      ElMessage.success('文案已应用')
    } else {
      ElMessage.warning('无效的方案编号')
    }
  } catch {
    ElMessage.error('AI 生成文案失败')
  } finally {
    aiLoading.value = false
  }
}

/**
 * 上移 Banner
 */
function moveUp(index: number) {
  if (index <= 0) return
  swapBanners(index, index - 1)
}

/**
 * 下移 Banner
 */
function moveDown(index: number) {
  if (index >= banners.value.length - 1) return
  swapBanners(index, index + 1)
}

/**
 * 交换两个位置的 Banner，并重新生成 sort_order
 */
function swapBanners(from: number, to: number) {
  const arr = banners.value
  ;[arr[from], arr[to]] = [arr[to], arr[from]]
  arr.forEach((b, i) => { b.sort_order = i + 1 })
}

/**
 * 切换 Banner 启用/禁用状态
 */
function toggleEnabled(index: number, val: boolean) {
  banners.value[index].enabled = val
}

/**
 * 打开新增对话框
 */
function handleAdd() {
  isEditing.value = false
  editingIndex.value = -1
  previewImage.value = ''
  editForm.value = {
    id: Date.now(),
    title: '',
    subtitle: '',
    image: '',
    link: '',
    sort_order: banners.value.length + 1,
    enabled: true,
    bg_color: '#409eff',
    btn_text: '',
    template: 'classic'
  }
  dialogVisible.value = true
}

/**
 * 打开编辑对话框
 */
function handleEdit(index: number) {
  isEditing.value = true
  editingIndex.value = index
  const banner = banners.value[index]
  editForm.value = { ...banner }
  previewImage.value = banner.image
  dialogVisible.value = true
}

/**
 * 确认编辑/新增
 */
async function confirmEdit() {
  ;(formRef.value as any)?.validate((valid: boolean) => {
    if (!valid) return

    submitting.value = true
    try {
      const formData = { ...editForm.value }
      // 极简版不需要图片校验
      if (formData.template === 'minimal') {
        formData.image = formData.image || ''
      }
      if (isEditing.value && editingIndex.value >= 0) {
        banners.value[editingIndex.value] = formData
        ElMessage.success('Banner 已更新')
      } else {
        banners.value.push(formData)
        ElMessage.success('Banner 已添加')
      }
      dialogVisible.value = false
    } catch {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

/**
 * 删除 Banner
 */
function handleDelete(index: number) {
  const banner = banners.value[index]
  ElMessageBox.confirm(
    `确定要删除 "${banner.title || '未命名'}" 吗？此操作不可撤销。`,
    '删除确认',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'error' }
  ).then(() => {
    banners.value.splice(index, 1)
    banners.value.forEach((b, i) => { b.sort_order = i + 1 })
    ElMessage.success('Banner 已删除')
  }).catch(() => {})
}

/**
 * 保存全部 Banner 到后端
 */
async function handleSaveAll() {
  for (let i = 0; i < banners.value.length; i++) {
    const b = banners.value[i]
    if (!b.title) {
      ElMessage.warning(`第 ${i + 1} 个 Banner 缺少标题，请补充完整后再保存`)
      return
    }
    if ((b.template || 'classic') !== 'minimal' && !b.image) {
      ElMessage.warning(`第 ${i + 1} 个 Banner 缺少图片地址，请补充完整后再保存`)
      return
    }
  }

  saving.value = true
  try {
    await saveBanners({ banners: banners.value })
    ElMessage.success('全部 Banner 已保存')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

/* ========== 生命周期 ========== */

onMounted(() => { loadBanners() })
onUnmounted(() => { if (previewTimer) clearInterval(previewTimer) })
</script>

<style scoped>
/* ===== 全局动画 ===== */
@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes slideInRight {
  from { opacity: 0; transform: translateX(20px); }
  to { opacity: 1; transform: translateX(0); }
}

.banners-page {
  animation: fadeSlideUp 0.4s ease both;
}

/* ===== 页面标题 ===== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground, #1e293b);
  margin: 0;
}

.page-subtitle {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground, #94a3b8);
}

.header-right {
  display: flex;
  gap: 10px;
}

.action-btn {
  border-radius: 10px;
  padding: 8px 18px;
  font-weight: 500;
  transition: all 0.25s ease;
}

.primary-btn {
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.25);
}

.primary-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.35);
}

/* ===== 操作栏 ===== */
.toolbar-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: #fff;
  border: 1px solid var(--shop-border, #e2e8f0);
  border-radius: 14px;
  margin-bottom: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.add-btn {
  border-radius: 10px;
  padding: 8px 20px;
  font-weight: 500;
  transition: all 0.25s ease;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.add-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stats-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground, #94a3b8);
}

.stats-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--shop-primary, #409eff);
}

.stats-dot--green {
  background: var(--shop-success, #67c23a);
}

.stats-dot--red {
  background: var(--shop-destructive, #f56c6c);
}

/* ===== 网站预览模拟器 ===== */
.site-preview {
  margin-bottom: 24px;
  animation: fadeSlideUp 0.5s ease both;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.preview-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--shop-muted-foreground, #94a3b8);
}

.preview-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.preview-slide-count {
  font-size: 12px;
  color: var(--shop-muted-foreground, #94a3b8);
  font-variant-numeric: tabular-nums;
}

/* 浏览器模拟窗口 */
.preview-browser {
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
  border: 1px solid var(--shop-border, #e2e8f0);
}

.browser-bar {
  display: flex;
  align-items: center;
  height: 38px;
  padding: 0 14px;
  background: #f1f3f5;
  border-bottom: 1px solid #e2e6ea;
  gap: 12px;
}

.browser-dots {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.browser-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}
.browser-dot--red { background: #ff5f57; }
.browser-dot--yellow { background: #febc2e; }
.browser-dot--green { background: #28c840; }

.browser-url {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 3px 12px;
  background: #fff;
  border-radius: 6px;
  font-size: 11px;
  color: #6b7280;
  border: 1px solid #d1d5db;
  max-width: 240px;
}

.browser-spacer {
  width: 52px;
  flex-shrink: 0;
}

/* 浏览器内容区 */
.browser-content {
  position: relative;
  overflow: hidden;
  background: #f8f9fa;
}

.preview-track {
  display: flex;
  transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.preview-slide {
  min-width: 100%;
  min-height: 220px;
  display: flex;
  align-items: center;
  position: relative;
  overflow: hidden;
}

/* 居中版遮罩 */
.ps-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 1;
}

/* 全图版背景 */
.ps-bg-img {
  position: absolute;
  inset: 0;
  z-index: 0;
}
.ps-bg-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ps-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 36px 44px;
  gap: 28px;
  position: relative;
  z-index: 2;
}

.ps--centered .ps-body {
  justify-content: center;
  text-align: center;
}
.ps--centered .ps-text {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.ps-text {
  flex: 1;
  min-width: 0;
}

.ps-title {
  font-size: 26px;
  font-weight: 700;
  color: #fff;
  line-height: 1.3;
  margin: 0 0 6px;
  letter-spacing: -0.02em;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  white-space: pre-line;
  word-break: break-word;
}

.ps-subtitle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
  margin: 0 0 14px;
  white-space: pre-line;
  word-break: break-word;
}

.ps-btn {
  display: inline-block;
  padding: 7px 24px;
  border-radius: 999px;
  background: #fff;
  color: var(--shop-primary, #409eff);
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.ps-image {
  width: 160px;
  height: 120px;
  border-radius: 10px;
  overflow: hidden;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.12);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}
.ps-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ps-image-placeholder {
  width: 160px;
  height: 120px;
  border-radius: 10px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.ps-accent {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 2px;
  background: rgba(255, 255, 255, 0.45);
  border-radius: 2px;
  z-index: 2;
}

/* 深色版文字阴影 */
.ps--dark .ps-title {
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}
.ps--full-image .ps-body {
  background: rgba(0, 0, 0, 0.15);
}

/* 指示器圆点 */
.preview-dots {
  position: absolute;
  bottom: 12px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  z-index: 5;
}

.preview-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.6);
  background: transparent;
  cursor: pointer;
  transition: all 0.3s;
  padding: 0;
}

.preview-dot--active {
  background: #fff;
  border-color: #fff;
  transform: scale(1.2);
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  background: #fafbfc;
  border-radius: 14px;
  border: 1px dashed var(--shop-border, #e2e8f0);
}

.empty-title {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground, #94a3b8);
  margin-top: 16px;
  margin-bottom: 20px;
}

/* ===== Banner 卡片列表容器 ===== */
.banner-list {
  min-height: 200px;
}

.banner-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ===== 列表入场动画（transition-group） ===== */
.list-stagger-enter-active {
  animation: fadeInUp 0.35s ease both;
}

.list-stagger-leave-active {
  animation: fadeInUp 0.25s ease reverse both;
  position: absolute;
}

.list-stagger-move {
  transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

/* ===== Banner 卡片 ===== */
.banner-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 18px;
  background: #fff;
  border: 1px solid var(--shop-border, #e2e8f0);
  border-radius: 14px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: fadeInUp 0.35s ease both;
  position: relative;
}

.banner-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08);
  border-color: var(--shop-primary, #409eff);
}

/* 序号 */
.card-index {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: #f1f5f9;
  font-size: 12px;
  font-weight: 700;
  color: var(--shop-muted-foreground, #94a3b8);
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
}

.banner-card:hover .card-index {
  background: var(--shop-primary, #409eff);
  color: #fff;
}

/* ===== 卡片预览 ===== */
.card-preview {
  flex-shrink: 0;
  width: 240px;
}

.card-banner {
  width: 240px;
  height: 78px;
  border-radius: 10px;
  overflow: hidden;
  position: relative;
  border: 1px solid rgba(255, 255, 255, 0.15);
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.banner-card:hover .card-banner {
  border-color: var(--shop-primary, #409eff);
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.18);
}

.cb-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 1;
}

.cb-bg-img {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cb-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  padding: 0 12px;
  gap: 10px;
  position: relative;
  z-index: 2;
}

.cb--centered .cb-body {
  justify-content: center;
  text-align: center;
}
.cb--centered .cb-text {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.cb-text {
  flex: 1;
  min-width: 0;
}

.cb-title {
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  line-height: 1.3;
  margin-bottom: 2px;
  white-space: pre-line;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.cb-subtitle {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cb-btn {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  background: #fff;
  color: var(--shop-primary, #409eff);
  font-size: 9px;
  font-weight: 600;
  line-height: 1.5;
}

.cb-img {
  width: 48px;
  height: 36px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.12);
}
.cb-img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cb-placeholder {
  width: 48px;
  height: 36px;
  border-radius: 6px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.cb-accent {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  width: 22px;
  height: 2px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 2px;
  z-index: 2;
}

.cb-tag {
  position: absolute;
  top: 4px;
  right: 4px;
  font-size: 8px;
  padding: 1px 6px;
  border-radius: 3px;
  background: rgba(0, 0, 0, 0.3);
  color: rgba(255, 255, 255, 0.85);
  z-index: 3;
  font-weight: 500;
  line-height: 1.5;
}

.cb--dark .cb-title {
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.35);
}
.cb--full-image .cb-body {
  background: rgba(0, 0, 0, 0.2);
}

/* ===== 卡片信息区 ===== */
.card-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.card-info-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-label {
  font-size: 11px;
  color: var(--shop-muted-foreground, #94a3b8);
  background: #f1f5f9;
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
  flex-shrink: 0;
  line-height: 1.6;
}

.card-value {
  font-size: 13px;
  color: var(--shop-foreground, #1e293b);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-value--link {
  color: var(--shop-primary, #409eff);
  font-size: 12px;
  font-family: ui-monospace, monospace;
}

/* ===== 卡片操作区 ===== */
.card-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  flex-shrink: 0;
  padding-left: 14px;
  border-left: 1px solid var(--shop-border, #e2e8f0);
  min-width: 100px;
}

.card-toggle {
  display: flex;
  align-items: center;
}

.card-buttons {
  display: flex;
  align-items: center;
  gap: 2px;
}

.card-btn {
  width: 30px;
  height: 30px;
  border-radius: 8px !important;
  transition: all 0.2s ease;
}

.card-btn:hover {
  background: var(--shop-accent, #f1f5f9) !important;
  transform: scale(1.1);
}

/* ===== 弹窗样式 ===== */
.banner-dialog {
  --dialog-radius: 16px;
}

.banner-dialog :deep(.el-dialog) {
  border-radius: var(--dialog-radius);
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.15);
  animation: dialogEnter 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes dialogEnter {
  from { opacity: 0; transform: scale(0.92) translateY(16px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.banner-dialog :deep(.el-dialog__header) {
  padding: 18px 24px;
  margin: 0;
  border-bottom: 1px solid var(--shop-border, #e2e8f0);
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  position: relative;
}

.banner-dialog :deep(.el-dialog__header::after) {
  content: '';
  position: absolute;
  bottom: 0;
  left: 24px;
  width: 40px;
  height: 2.5px;
  background: var(--shop-primary, #409eff);
  border-radius: 2px;
}

.banner-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 700;
  color: var(--shop-foreground, #1e293b);
  letter-spacing: 0.01em;
}

.banner-dialog :deep(.el-dialog__body) {
  padding: 0;
}

/* ===== 模板选择器 ===== */
.template-picker {
  padding: 18px 24px 14px;
  border-bottom: 1px solid var(--shop-border, #e2e8f0);
  background: #fff;
}

.template-picker-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--shop-foreground, #1e293b);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.template-picker-label::before {
  content: '';
  width: 3px;
  height: 14px;
  background: var(--shop-primary, #409eff);
  border-radius: 2px;
}

.template-list {
  display: flex;
  gap: 10px;
}

.template-item {
  flex: 1;
  cursor: pointer;
  border-radius: 12px;
  padding: 10px 8px 8px;
  border: 2px solid var(--shop-border, #e2e8f0);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  position: relative;
  overflow: hidden;
}

.template-item::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, var(--shop-primary, #409eff), #60a5fa);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.template-item:hover {
  border-color: var(--shop-primary-light, #b3d8ff);
  transform: translateY(-3px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.08);
}

.template-item--active {
  border-color: var(--shop-primary, #409eff) !important;
  background: #f0f7ff;
  box-shadow: 0 6px 20px rgba(64, 158, 255, 0.18);
  transform: translateY(-3px);
}

.template-item--active .template-item {
  background: var(--shop-primary, #409eff);
  color: #fff;
}

/* 模板缩略图 */
.template-thumb {
  width: 100%;
  height: 44px;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  position: relative;
  background: var(--shop-primary, #409eff);
  transition: transform 0.3s ease;
}

.template-item:hover .template-thumb {
  transform: scale(1.04);
}

.template-item--active .template-thumb {
  transform: scale(1.04);
}

.tthumb-body {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 10px;
  width: 100%;
  position: relative;
  z-index: 1;
}

.tthumb-title {
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  line-height: 1;
  letter-spacing: 0.02em;
}

.tthumb-sub {
  font-size: 9px;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1;
}

.tthumb-btn {
  margin-left: auto;
  font-size: 8px;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.9);
  color: var(--shop-primary, #409eff);
  font-weight: 600;
  line-height: 1.5;
}

.tthumb-img {
  width: 20px;
  height: 16px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.2);
  flex-shrink: 0;
}

/* 各模板缩略图风格 */
.tthumb--full-image {
  background: linear-gradient(135deg, #2d3748, #4a5568);
}

.tthumb--centered .tthumb-body {
  justify-content: center;
}

.tthumb--centered {
  background: linear-gradient(135deg, #4a5568, #718096);
}

.tthumb--centered .tthumb-sub {
  display: none;
}

.tthumb--dark {
  background: linear-gradient(135deg, #1a1a2e, #0f172a);
}

.tthumb--minimal {
  background: linear-gradient(135deg, #64748b, #94a3b8);
}

.tthumb--minimal .tthumb-body {
  justify-content: center;
}

.tthumb--minimal .tthumb-sub,
.tthumb--minimal .tthumb-btn {
  display: none;
}

.tthumb-accent {
  position: absolute;
  bottom: 7px;
  left: 50%;
  transform: translateX(-50%);
  width: 18px;
  height: 2px;
  background: rgba(255, 255, 255, 0.45);
  border-radius: 1px;
  z-index: 1;
}

.template-name {
  font-size: 11px;
  color: var(--shop-foreground, #1e293b);
  font-weight: 500;
  text-align: center;
  transition: color 0.3s ease;
}

.template-item--active .template-name {
  color: var(--shop-primary, #409eff);
  font-weight: 600;
}

/* ===== 弹窗主体上下布局 ===== */
.dialog-body {
  padding: 18px 24px 16px;
  max-height: 480px;
  overflow-y: auto;
}

/* 顶部实时预览 */
.dialog-preview-top {
  margin-bottom: 18px;
  animation: fadeSlideUp 0.3s ease both;
}

.preview-top-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--shop-foreground, #1e293b);
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.preview-top-label::before {
  content: '';
  width: 3px;
  height: 14px;
  background: var(--shop-primary, #409eff);
  border-radius: 2px;
}

.preview-template-tag {
  margin-top: 6px;
  text-align: center;
}

/* 实时预览容器 */
.live-preview {
  border-radius: 12px;
  overflow: hidden;
  height: 100px;
  display: flex;
  align-items: center;
  position: relative;
  transition: background 0.5s ease, box-shadow 0.3s ease;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.live-preview:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.live-preview-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 1;
}

.live-preview-img-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.live-preview-img-bg img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.live-preview-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 18px;
  gap: 12px;
  position: relative;
  z-index: 2;
  animation: previewFade 0.35s ease both;
}

@keyframes previewFade {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.template--centered .live-preview-body {
  justify-content: center;
  text-align: center;
}

.template--centered .live-preview-text {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.live-preview-text {
  flex: 1;
  min-width: 0;
}

.live-preview-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  line-height: 1.3;
  margin-bottom: 4px;
  white-space: pre-line;
  word-break: break-word;
  transition: font-size 0.3s ease;
}

.live-preview-subtitle {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.85);
  margin-bottom: 8px;
  white-space: pre-line;
  word-break: break-word;
}

.live-preview-btn {
  display: inline-block;
  padding: 5px 16px;
  border-radius: 999px;
  background: #fff;
  color: var(--shop-primary, #409eff);
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
  transition: transform 0.2s ease;
}

.live-preview-btn:hover {
  transform: scale(1.05);
}

.live-preview-image {
  width: 70px;
  height: 52px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.12);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  animation: previewFade 0.35s ease both;
}

.live-preview-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.live-preview-img-placeholder {
  width: 70px;
  height: 52px;
  border-radius: 6px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: previewFade 0.35s ease both;
}

.live-minimal-accent {
  position: absolute;
  bottom: 14px;
  left: 50%;
  transform: translateX(-50%);
  width: 28px;
  height: 2px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 2px;
  z-index: 2;
}

/* 深色版文字阴影 */
.template--dark .live-preview-title {
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

/* ===== 表单区域 ===== */
.dialog-form-area {
  width: 100%;
}

.dialog-form-area :deep(.el-form-item__label) {
  font-weight: 500;
  color: #475569;
  padding-bottom: 2px;
}

/* 表单卡片 - 无边框设计 */
.form-card {
  margin-bottom: 14px;
  border-radius: 12px !important;
  border: 1px solid var(--shop-border, #e2e8f0) !important;
  transition: box-shadow 0.3s ease, border-color 0.3s ease;
}

.form-card:hover {
  border-color: #cbd5e1 !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.form-card:last-child {
  margin-bottom: 0;
}

.form-card :deep(.el-card__header) {
  padding: 10px 16px;
  border-bottom: 1px solid var(--shop-border, #e2e8f0);
  background: linear-gradient(135deg, #fafbfc 0%, #f8fafc 100%);
  border-radius: 12px 12px 0 0 !important;
}

.form-card :deep(.el-card__body) {
  padding: 14px 16px 10px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--shop-foreground, #1e293b);
  display: flex;
  align-items: center;
  gap: 6px;
}

.section-title::before {
  content: '';
  width: 3px;
  height: 12px;
  background: var(--shop-primary, #409eff);
  border-radius: 2px;
  opacity: 0.6;
}

/* ===== 配色方案 ===== */
.color-schemes {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.color-scheme-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 4px 2px;
  border-radius: 8px;
  transition: all 0.2s ease;
  width: 54px;
}

.color-scheme-item:hover {
  background: #f1f5f9;
  transform: translateY(-2px);
}

.color-scheme-item--active {
  background: #eff6ff;
}

.cs-preview {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.color-scheme-item:hover .cs-preview {
  transform: scale(1.12);
}

.color-scheme-item--active .cs-preview {
  border-color: var(--shop-primary, #409eff);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.3);
  transform: scale(1.12);
}

.cs-check {
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  filter: drop-shadow(0 1px 1px rgba(0, 0, 0, 0.3));
}

.cs-name {
  font-size: 10px;
  color: var(--shop-muted-foreground, #94a3b8);
  text-align: center;
  white-space: nowrap;
}

.color-picker-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
}

.color-hex {
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: var(--shop-muted-foreground, #94a3b8);
  min-width: 40px;
}

/* 图片预览 */
.image-preview-box {
  margin-top: 10px;
  width: 100%;
  max-width: 240px;
  height: 100px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--el-border-color, #dcdfe6);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.image-preview-box:hover {
  border-color: var(--el-color-primary, #409eff);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.image-preview-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* ===== 弹窗底部 ===== */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 24px;
  border-top: 1px solid var(--shop-border, #e2e8f0);
  background: #fafbfc;
}

.dialog-footer .el-button {
  border-radius: 10px;
  padding: 8px 22px;
  font-weight: 500;
  transition: all 0.25s ease;
}

.dialog-footer .el-button--primary {
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.25);
}

.dialog-footer .el-button--primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.35);
}

/* ===== AI 生成文案按钮行 ===== */
.ai-generate-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #f0f7ff 0%, #e8f4fd 100%);
  border-radius: 10px;
  border: 1px solid rgba(64, 158, 255, 0.15);
}
.ai-generate-hint {
  font-size: 0.75rem;
  color: var(--shop-muted-foreground, #94a3b8);
}

/* ===== 上传按钮 ===== */
.upload-btn .el-icon {
  margin-right: 2px;
}

/* ===== 滚动条美化 ===== */
.dialog-body::-webkit-scrollbar {
  width: 4px;
}

.dialog-body::-webkit-scrollbar-track {
  background: transparent;
}

.dialog-body::-webkit-scrollbar-thumb {
  background: #d0d4dc;
  border-radius: 4px;
}

.dialog-body::-webkit-scrollbar-thumb:hover {
  background: #b0b4c0;
}

/* ===== 响应式 ===== */
@media (max-width: 1100px) {
  .dialog-body {
    max-height: none;
    overflow-y: visible;
  }

  .dialog-preview-top {
    margin-bottom: 12px;
  }

  .live-preview {
    height: 80px;
  }
}

@media (max-width: 1024px) {
  .banner-card {
    flex-wrap: wrap;
    gap: 12px;
  }

  .card-preview {
    width: 200px;
  }

  .card-banner {
    width: 200px;
    height: 66px;
  }

  .template-list {
    flex-wrap: wrap;
  }

  .template-item {
    min-width: calc(33% - 8px);
    flex: none;
  }
}

@media (max-width: 768px) {
  .browser-url {
    display: none;
  }
  .browser-spacer {
    display: none;
  }

  .preview-slide {
    min-height: 160px;
  }
  .ps-body {
    padding: 20px 16px;
    gap: 12px;
  }
  .ps-title {
    font-size: 18px;
  }
  .ps-subtitle {
    font-size: 12px;
  }
  .ps-image,
  .ps-image-placeholder {
    width: 100px;
    height: 76px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-right {
    width: 100%;
  }

  .header-right .el-button {
    flex: 1;
  }

  .toolbar-bar {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .banner-card {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .card-preview {
    width: 100%;
  }

  .card-banner {
    width: 100%;
    height: 72px;
  }

  .card-actions {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    border-left: none;
    border-top: 1px solid var(--shop-border, #e2e8f0);
    padding-left: 0;
    padding-top: 8px;
    min-width: 0;
  }

  .template-item {
    min-width: calc(50% - 6px);
    flex: none;
  }

  .dialog-body {
    padding: 12px 16px;
  }
}
</style>
