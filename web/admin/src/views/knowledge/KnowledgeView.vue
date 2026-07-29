<template>
  <div class="knowledge-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">知识库管理</h2>
        <span class="page-subtitle">上传文档构建 AI 知识库，增强智能体回答准确性</span>
      </div>
      <div class="header-right">
        <el-button type="primary" class="btn-primary" @click="showUploadDialog = true">
          <el-icon><Upload /></el-icon> 上传文档
        </el-button>
        <el-button class="btn-secondary" @click="loadDocs">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-icon stat-icon--blue">
              <el-icon :size="22"><Collection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ docs.length }}</div>
              <div class="stat-label">文档总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-icon stat-icon--green">
              <el-icon :size="22"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ totalChunks }}</div>
              <div class="stat-label">分块总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-icon stat-icon--orange">
              <el-icon :size="22"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ sourceCount }}</div>
              <div class="stat-label">来源分类</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-icon stat-icon--purple">
              <el-icon :size="22"><Cpu /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value stat-value--status" :class="{ 'is-ready': hasIndex }">
                <span class="status-dot"></span>
                {{ hasIndex ? '已就绪' : '未创建' }}
              </div>
              <div class="stat-label">向量索引</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 搜索栏 -->
    <el-card shadow="never" class="search-card">
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索文档标题或标签..."
          clearable
          :prefix-icon="Search"
          style="max-width: 420px"
        />
        <el-select v-model="sourceFilter" placeholder="来源" clearable style="max-width: 160px">
          <el-option label="管理后台" value="admin" />
          <el-option label="用户端" value="user" />
          <el-option label="公开" value="public" />
        </el-select>

        <!-- 活动筛选条件显示 -->
        <div class="active-filters" v-if="sourceFilter || searchKeyword">
          <span v-if="sourceFilter" class="active-filter-tag">
            {{ getSourceLabel(sourceFilter) }}
            <el-icon class="close-icon" @click="sourceFilter = ''"><Close /></el-icon>
          </span>
          <span v-if="searchKeyword" class="active-filter-tag">
            关键词: {{ searchKeyword }}
            <el-icon class="close-icon" @click="searchKeyword = ''"><Close /></el-icon>
          </span>
        </div>
      </div>
    </el-card>

    <!-- 文档列表 -->
    <el-card shadow="never" class="docs-card">
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <span>加载文档列表中...</span>
      </div>

      <div v-else-if="filteredDocs.length === 0" class="empty-state">
        <div class="empty-icon-wrapper">
          <el-icon :size="72" class="empty-icon"><FolderOpened /></el-icon>
        </div>
        <p class="empty-title">还没有知识库文档</p>
        <p class="empty-desc">上传文档后，AI 将自动解析、分块并建立向量索引，为智能体提供精准的知识问答能力</p>

        <div class="format-hints">
          <span class="format-hint-item">TXT</span>
          <span class="format-hint-item">Markdown</span>
          <span class="format-hint-item">PDF</span>
          <span class="format-hint-item">Word</span>
          <span class="format-hint-item">JSON</span>
          <span class="format-hint-item">HTML</span>
        </div>

        <el-button type="primary" size="large" class="empty-cta" @click="showUploadDialog = true">
          <el-icon><Upload /></el-icon> 上传第一个文档
        </el-button>
      </div>

      <el-table v-else :data="filteredDocs" stripe class="docs-table">
        <el-table-column label="文档标题" min-width="260">
          <template #default="{ row }">
            <div class="doc-title">
              <div class="doc-icon-wrapper">
                <el-icon class="doc-icon"><Document /></el-icon>
              </div>
              <div class="doc-title-content">
                <span class="doc-title-text">{{ row.title }}</span>
                <span class="doc-meta" v-if="row.source">来源：{{ getSourceLabel(row.source) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分块数" width="100" align="center">
          <template #default="{ row }">
            <span class="chunk-count">{{ row.chunkCount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="200">
          <template #default="{ row }">
            <div v-if="row.tags && Object.keys(row.tags).length > 0" class="tags-wrap">
              <el-tag
                v-for="(val, key) in row.tags"
                :key="key"
                size="small"
                class="tag-item"
                effect="plain"
              >
                <span class="tag-key">{{ key }}</span>
                <span class="tag-sep">:</span>
                <span class="tag-val">{{ val }}</span>
              </el-tag>
            </div>
            <span v-else class="no-tags">无标签</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" text @click="handleDelete(row)">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 上传文档对话框 -->
    <el-dialog v-model="showUploadDialog" title="上传知识库文档" width="520px" :close-on-click-modal="false" class="upload-dialog">
      <el-form :model="uploadForm" label-position="top" class="upload-form">
        <el-form-item label="文档标题">
          <el-input v-model="uploadForm.title" placeholder="留空则使用文件名" />
        </el-form-item>
        <el-form-item label="文档来源">
          <el-select v-model="uploadForm.source" style="width: 100%">
            <el-option label="管理后台" value="admin" />
            <el-option label="用户端" value="user" />
            <el-option label="公开" value="public" />
          </el-select>
        </el-form-item>
        <el-form-item label="上传文件">
          <el-upload
            drag
            :auto-upload="false"
            :on-change="handleFileChange"
            :show-file-list="false"
            accept=".txt,.md,.markdown,.pdf,.docx,.json,.html"
            class="upload-dragger"
          >
            <el-icon :size="44" class="upload-icon"><UploadFilled /></el-icon>
            <div class="el-upload__text">将文件拖拽到此处，或<em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 TXT、Markdown、PDF、Word、JSON、HTML 格式，最大 50MB</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="uploadForm.file" label="文件信息">
          <div class="file-info">
            <el-icon :size="20"><Document /></el-icon>
            <span>{{ uploadForm.file.name }}</span>
            <el-tag size="small" effect="light">{{ formatFileSize(uploadForm.file.size) }}</el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetUploadForm">取消</el-button>
        <el-button type="primary" class="btn-primary" @click="handleUpload" :loading="uploading" :disabled="!uploadForm.file">
          开始上传
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/**
 * 知识库管理页面
 * 提供文档列表、上传、搜索、删除功能
 */
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Upload,
  Refresh,
  Search,
  Delete,
  Loading,
  FolderOpened,
  Document,
  UploadFilled,
  Collection,
  Connection,
  Cpu,
  Close
} from '@element-plus/icons-vue'
import {
  listKnowledge,
  deleteKnowledge,
  uploadKnowledge,
  type KnowledgeDocInfo
} from '@/api/knowledge'

/** 文档列表 */
const docs = ref<KnowledgeDocInfo[]>([])
const loading = ref(false)

/** 搜索过滤 */
const searchKeyword = ref('')
const sourceFilter = ref('')

/** 上传相关 */
const showUploadDialog = ref(false)
const uploading = ref(false)
const uploadForm = ref({
  file: null as File | null,
  title: '',
  source: 'admin'
})

/** 是否有索引（只要有文档就说明索引存在） */
const hasIndex = computed(() => docs.value.length > 0)

/** 总分块数 */
const totalChunks = computed(() =>
  docs.value.reduce((sum, d) => sum + (d.chunkCount || 0), 0)
)

/** 来源分类数 */
const sourceCount = computed(() => {
  const sources = new Set(docs.value.map(d => d.source))
  return sources.size
})

/** 过滤后的文档列表 */
const filteredDocs = computed(() => {
  let result = docs.value
  if (sourceFilter.value) {
    result = result.filter(d => d.source === sourceFilter.value)
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.trim().toLowerCase()
    result = result.filter(
      d =>
        d.title.toLowerCase().includes(kw) ||
        Object.values(d.tags || {}).some(v => v.toLowerCase().includes(kw))
    )
  }
  return result
})

/** 加载文档列表 */
async function loadDocs() {
  loading.value = true
  try {
    docs.value = await listKnowledge()
  } catch (e: any) {
    ElMessage.error('加载文档列表失败: ' + (e?.message || ''))
  } finally {
    loading.value = false
  }
}

/** 获取来源标签文字 */
function getSourceLabel(source: string): string {
  switch (source) {
    case 'admin': return '管理后台'
    case 'user': return '用户端'
    case 'public': return '公开'
    default: return source
  }
}

/** 文件选择变化 */
function handleFileChange(file: any) {
  if (file && file.raw) {
    uploadForm.value.file = file.raw
    if (!uploadForm.value.title) {
      uploadForm.value.title = file.name
    }
  }
}

/** 格式化文件大小 */
function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(2) + ' MB'
}

/** 重置上传表单 */
function resetUploadForm() {
  uploadForm.value = { file: null, title: '', source: 'admin' }
}

/** 执行上传 */
async function handleUpload() {
  if (!uploadForm.value.file) return
  uploading.value = true
  try {
    const res = await uploadKnowledge(
      uploadForm.value.file,
      uploadForm.value.title || undefined,
      uploadForm.value.source
    )
    ElMessage.success(`上传成功！文档已索引 ${res.chunkCount} 个分块`)
    showUploadDialog.value = false
    resetUploadForm()
    await loadDocs()
  } catch (e: any) {
    ElMessage.error('上传失败: ' + (e?.message || ''))
  } finally {
    uploading.value = false
  }
}

/** 删除文档 */
async function handleDelete(row: KnowledgeDocInfo) {
  try {
    await ElMessageBox.confirm(
      `确定要删除文档「${row.title}」吗？删除后将不可恢复。`,
      '删除确认',
      { type: 'warning' }
    )
    await deleteKnowledge(row.docId, row.source)
    ElMessage.success('删除成功')
    await loadDocs()
  } catch {
    // 取消删除
  }
}

onMounted(() => {
  loadDocs()
})
</script>

<style scoped>
.knowledge-page {
  animation: fadeSlideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  padding-bottom: 24px;
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* ===== 页面头部 ===== */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--shop-border);
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0;
  letter-spacing: -0.01em;
}

.page-subtitle {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
  margin-top: 6px;
  display: block;
  line-height: 1.5;
}

.header-right {
  display: flex;
  gap: 12px;
}

/* ===== 按钮样式 ===== */
.btn-primary {
  border-radius: var(--shop-radius-md);
  font-weight: 600;
  padding: 10px 20px;
  box-shadow: 0 2px 8px hsla(var(--shop-primary-h), var(--shop-primary-s), var(--shop-primary-l), 0.3);
  transition: all var(--shop-transition);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px hsla(var(--shop-primary-h), var(--shop-primary-s), var(--shop-primary-l), 0.4);
}

.btn-secondary {
  border-radius: var(--shop-radius-md);
  border-color: var(--shop-border);
  transition: all var(--shop-transition);
}

.btn-secondary:hover {
  border-color: var(--shop-primary);
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

/* ===== 统计卡片 ===== */
.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: var(--shop-radius-xl);
  border: 1px solid var(--shop-border);
  transition: all var(--shop-transition);
  overflow: hidden;
  position: relative;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--shop-primary), transparent);
  opacity: 0;
  transition: opacity var(--shop-transition);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  border-color: var(--shop-primary-light);
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-card :deep(.el-card__body) {
  padding: 16px 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--shop-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.stat-icon--blue { background: linear-gradient(135deg, #3b82f6, #6366f1); }
.stat-icon--green { background: linear-gradient(135deg, #10b981, #059669); }
.stat-icon--orange { background: linear-gradient(135deg, #f59e0b, #ea580c); }
.stat-icon--purple { background: linear-gradient(135deg, #8b5cf6, #7c3aed); }

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--shop-foreground);
  line-height: 1.2;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--shop-muted-foreground);
  margin-top: 2px;
}

/* 向量索引状态指示器 */
.stat-value--status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.25rem;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--shop-destructive, #ef4444);
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.2);
  flex-shrink: 0;
}

.stat-value--status.is-ready .status-dot {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.2);
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* ===== 搜索栏 ===== */
.search-card {
  margin-bottom: 16px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-border);
  background: var(--shop-card);
}

.search-card :deep(.el-card__body) {
  padding: 12px 16px;
}

.search-bar {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-bar .el-input__wrapper {
  border-radius: var(--shop-radius-md);
  border: 1px solid var(--shop-border);
  transition: all var(--shop-transition);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

.search-bar .el-input__wrapper:hover {
  border-color: var(--shop-primary);
}

.search-bar .el-input__wrapper.is-focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(var(--shop-primary-h), var(--shop-primary-s), var(--shop-primary-l), 0.1);
}

.search-bar .el-select .el-input__wrapper {
  border-radius: var(--shop-radius-md);
}

/* 活动筛选标签 */
.active-filters {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-left: auto;
}

.active-filter-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  background: var(--shop-primary-bg);
  color: var(--shop-primary);
  border-radius: var(--shop-radius-sm);
  font-size: 0.75rem;
  font-weight: 500;
}

.active-filter-tag .close-icon {
  cursor: pointer;
  opacity: 0.7;
  font-size: 0.75rem;
}

.active-filter-tag .close-icon:hover {
  opacity: 1;
}

/* ===== 文档卡片 ===== */
.docs-card {
  border-radius: var(--shop-radius-xl);
  border: 1px solid var(--shop-border);
  overflow: hidden;
}

.docs-card :deep(.el-table) {
  --el-table-border-color: var(--shop-border);
  --el-table-header-bg-color: var(--shop-muted);
  --el-table-row-hover-bg-color: var(--shop-primary-bg);
  font-size: 0.8125rem;
}

.docs-card :deep(.el-table th.el-table__cell) {
  font-weight: 600;
  color: var(--shop-muted-foreground);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.docs-card :deep(.el-table td.el-table__cell) {
  padding: 14px 0;
}

/* ===== 加载状态 ===== */
.loading-state {
  padding: 60px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: var(--shop-muted-foreground);
}

.loading-state .el-icon {
  color: var(--shop-primary);
}

/* ===== 空状态 ===== */
.empty-state {
  padding: 80px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: var(--shop-muted-foreground);
  background: linear-gradient(180deg, var(--shop-muted) 0%, transparent 100%);
  border-radius: var(--shop-radius-xl);
  border: 2px dashed var(--shop-border);
}

.empty-icon-wrapper {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  background: var(--shop-primary-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}

.empty-icon {
  color: var(--shop-primary);
  opacity: 0.7;
}

.empty-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--shop-foreground);
  margin: 0;
}

.empty-desc {
  font-size: 0.875rem;
  margin: 0;
  max-width: 340px;
  text-align: center;
  line-height: 1.6;
}

.format-hints {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  margin-top: 4px;
}

.format-hint-item {
  padding: 4px 10px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-sm);
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
}

.empty-cta {
  margin-top: 8px;
}

/* ===== 表格内容 ===== */
.doc-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.doc-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: var(--shop-radius-md);
  background: var(--shop-primary-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.doc-icon {
  color: var(--shop-primary);
}

.doc-title-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.doc-title-text {
  font-weight: 500;
  color: var(--shop-foreground);
  line-height: 1.4;
}

.doc-meta {
  font-size: 0.6875rem;
  color: var(--shop-muted-foreground);
}

.chunk-count {
  display: inline-block;
  min-width: 32px;
  padding: 3px 10px;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-sm);
  font-weight: 600;
  color: var(--shop-foreground);
  font-variant-numeric: tabular-nums;
}

/* 标签样式优化 */
.tags-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-item {
  margin-right: 0;
  border-radius: var(--shop-radius-sm);
  font-size: 0.6875rem;
}

.tag-key {
  color: var(--shop-muted-foreground);
}

.tag-sep {
  margin: 0 2px;
  opacity: 0.5;
}

.tag-val {
  color: var(--shop-foreground);
  font-weight: 500;
}

.no-tags {
  color: var(--shop-border);
  font-style: italic;
}

/* ===== 上传对话框 ===== */
.upload-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.upload-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.upload-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--shop-foreground);
}

.upload-form :deep(.el-upload-dragger) {
  border-radius: var(--shop-radius-lg);
  border: 2px dashed var(--shop-border);
  background: var(--shop-muted);
  padding: 28px 20px;
  transition: all var(--shop-transition);
}

.upload-form :deep(.el-upload-dragger:hover) {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.upload-form :deep(.el-upload-dragger.is-dragover) {
  border-color: var(--shop-primary);
  background: var(--shop-primary-bg);
  box-shadow: 0 0 0 4px hsla(var(--shop-primary-h), var(--shop-primary-s), var(--shop-primary-l), 0.1);
}

.upload-icon {
  color: var(--shop-primary);
  margin-bottom: 8px;
}

.upload-form :deep(.el-upload__text) {
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
}

.upload-form :deep(.el-upload__text em) {
  color: var(--shop-primary);
  font-style: normal;
  font-weight: 500;
}

.upload-form :deep(.el-upload__tip) {
  color: var(--shop-muted-foreground);
  font-size: 0.75rem;
  margin-top: 12px;
  line-height: 1.5;
}

/* 文件信息展示 */
.file-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: var(--shop-primary-bg);
  border: 1px solid hsla(var(--shop-primary-h), var(--shop-primary-s), var(--shop-primary-l), 0.2);
  border-radius: var(--shop-radius-md);
  width: 100%;
}

.file-info span {
  flex: 1;
  font-weight: 500;
  color: var(--shop-foreground);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ===== 响应式 ===== */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .header-right {
    width: 100%;
  }

  .search-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-bar .el-input,
  .search-bar .el-select {
    max-width: 100% !important;
  }

  .active-filters {
    margin-left: 0;
  }

  .stat-value {
    font-size: 1.5rem;
  }
}

@media (max-width: 480px) {
  .stats-row .el-col {
    max-width: 50%;
    flex: 0 0 50%;
  }
}
</style>
