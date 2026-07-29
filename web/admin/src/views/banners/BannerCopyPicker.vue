<template>
  <!-- AI 文案方案选择弹窗 -->
  <el-dialog
    v-model="visible"
    title="选择文案方案"
    width="680px"
    :close-on-click-modal="false"
    destroy-on-close
    class="copy-picker-dialog"
    top="8vh"
    append-to-body
    @close="handleCancel"
  >
    <div class="copy-picker-body">
      <!-- 提示语 -->
      <div class="copy-picker-header">
        <div class="copy-picker-icon">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 20h9" />
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
          </svg>
        </div>
        <div class="copy-picker-info">
          <span class="copy-picker-title">AI 生成结果</span>
          <span class="copy-picker-subtitle">共生成 <strong>{{ options.length }}</strong> 个文案方案，点击卡片选择心仪的方案</span>
        </div>
      </div>

      <!-- 方案卡片列表 -->
      <div class="copy-option-list">
        <div
          v-for="(opt, index) in options"
          :key="index"
          class="copy-option-card"
          :class="{
            'copy-option-card--selected': selectedIndex === index,
            'copy-option-card--first': index === 0,
            'copy-option-card--last': index === options.length - 1
          }"
          :style="{ animationDelay: `${index * 0.08}s` }"
          @click="selectOption(index)"
        >
          <!-- 选中角标 -->
          <div class="selected-badge" v-if="selectedIndex === index">
            <el-icon><Check /></el-icon>
          </div>

          <!-- 方案序号与标签 -->
          <div class="option-card-top">
            <span class="option-badge" :class="'badge--' + getBadgeType(index)">方案 {{ index + 1 }}</span>
            <span class="option-recommend" v-if="index === 0">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
              推荐
            </span>
          </div>

          <!-- 文案内容预览 -->
          <div class="option-card-body">
            <div class="option-preview" :style="{ borderLeftColor: getPreviewColor(index) }">
              <div class="option-title">{{ opt.title || '（无标题）' }}</div>
              <div class="option-subtitle" v-if="opt.subtitle">{{ opt.subtitle }}</div>
              <div class="option-btn-row" v-if="opt.btn_text">
                <span class="option-btn">{{ opt.btn_text }}</span>
              </div>
            </div>
          </div>

          <!-- 选中态高亮条 -->
          <div class="card-glow" v-if="selectedIndex === index" />
        </div>
      </div>
    </div>

    <!-- 底部操作 -->
    <template #footer>
      <div class="copy-picker-footer">
        <div class="footer-preview" v-if="selectedIndex >= 0">
          <span class="footer-preview-label">已选：</span>
          <span class="footer-preview-text">{{ options[selectedIndex].title }}</span>
        </div>
        <div class="footer-actions">
          <el-button class="footer-btn footer-btn--cancel" @click="handleCancel">取消</el-button>
          <el-button
            type="primary"
            class="footer-btn footer-btn--confirm"
            :disabled="selectedIndex === -1"
            @click="handleConfirm"
          >
            <el-icon style="margin-right: 4px;"><Check /></el-icon>
            应用此方案
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
/**
 * Banner AI 文案方案选择弹窗
 * 展示 AI 生成的多组文案方案，支持卡片式点击选择
 */
import { ref, watch } from 'vue'
import { Check } from '@element-plus/icons-vue'
import type { BannerCopyOption } from '../../api/agent'

const props = withDefaults(defineProps<{
  /** 是否显示弹窗 */
  modelValue: boolean
  /** AI 生成的文案方案列表 */
  options: BannerCopyOption[]
}>(), {
  modelValue: false,
  options: () => []
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': [option: BannerCopyOption]
}>()

const visible = ref(props.modelValue)
const selectedIndex = ref(-1)

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    selectedIndex.value = -1
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

/** 预制配色（与 BannersView 同步） */
const previewColors = ['#409eff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1']

/**
 * 获取方案对应的标签类型
 */
function getBadgeType(index: number): string {
  const types = ['primary', 'success', 'warning', 'danger', 'info']
  return types[index % types.length]
}

/**
 * 获取方案预览色
 */
function getPreviewColor(index: number): string {
  return previewColors[index % previewColors.length]
}

/**
 * 选择方案
 */
function selectOption(index: number) {
  selectedIndex.value = index
}

/**
 * 取消
 */
function handleCancel() {
  visible.value = false
}

/**
 * 确认选择
 */
function handleConfirm() {
  if (selectedIndex.value < 0 || selectedIndex.value >= props.options.length) return
  emit('confirm', props.options[selectedIndex.value])
  visible.value = false
}
</script>

<style scoped>
/* ===== 弹窗全局 ===== */
.copy-picker-dialog {
  --dialog-radius: 16px;
}

.copy-picker-dialog :deep(.el-dialog) {
  border-radius: var(--dialog-radius);
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0, 0, 0, 0.18);
  animation: dialogEnter 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes dialogEnter {
  from { opacity: 0; transform: scale(0.92) translateY(16px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.copy-picker-dialog :deep(.el-dialog__header) {
  padding: 18px 24px;
  margin: 0;
  border-bottom: 1px solid #e2e8f0;
  background: linear-gradient(135deg, #f0f7ff 0%, #e8f4fd 100%);
}

.copy-picker-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
  letter-spacing: 0.01em;
}

.copy-picker-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.copy-picker-dialog :deep(.el-dialog__headerbtn) {
  top: 18px;
}

/* ===== 主体 ===== */
.copy-picker-body {
  padding: 20px 24px 12px;
}

/* ===== 头部提示 ===== */
.copy-picker-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 18px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  margin-bottom: 18px;
  border: 1px solid #e2e8f0;
}

.copy-picker-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
}

.copy-picker-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.copy-picker-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
}

.copy-picker-subtitle {
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.5;
}

.copy-picker-subtitle strong {
  color: #409eff;
  font-weight: 700;
}

/* ===== 方案卡片列表 ===== */
.copy-option-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 4px;
}

/* ===== 方案卡片 ===== */
.copy-option-card {
  position: relative;
  border: 2px solid #e2e8f0;
  border-radius: 14px;
  padding: 16px 18px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: #fff;
  overflow: hidden;
  user-select: none;
  animation: cardSlideIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

@keyframes cardSlideIn {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.copy-option-card:hover {
  border-color: #b3d8ff;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}

/* 选中态 */
.copy-option-card--selected {
  border-color: #409eff !important;
  background: linear-gradient(135deg, #f0f7ff 0%, #eaf4fe 100%);
  box-shadow: 0 6px 20px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

/* 选中角标 */
.selected-badge {
  position: absolute;
  top: -1px;
  right: -1px;
  width: 28px;
  height: 28px;
  background: #409eff;
  border-radius: 0 14px 0 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  z-index: 2;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}

/* ===== 卡片顶部：序号与标签 ===== */
.option-card-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.option-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.badge--primary {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.badge--success {
  background: rgba(82, 196, 26, 0.1);
  color: #52c41a;
}

.badge--warning {
  background: rgba(250, 140, 22, 0.1);
  color: #fa8c16;
}

.badge--danger {
  background: rgba(235, 47, 150, 0.1);
  color: #eb2f96;
}

.badge--info {
  background: rgba(114, 46, 209, 0.1);
  color: #722ed1;
}

.copy-option-card--selected .option-badge {
  background: rgba(64, 158, 255, 0.15);
  color: #409eff;
}

.option-recommend {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  background: linear-gradient(135deg, #fdf2e9, #fef7e6);
  color: #d48806;
  border: 1px solid rgba(212, 136, 6, 0.15);
}

.option-recommend svg {
  color: #faad14;
}

/* ===== 文案内容预览区 ===== */
.option-card-body {
  padding-left: 4px;
}

.option-preview {
  padding: 0 0 0 14px;
  border-left: 3px solid #409eff;
  transition: border-left-color 0.3s ease;
}

.option-title {
  font-size: 15px;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.4;
  margin-bottom: 4px;
  transition: color 0.3s ease;
}

.copy-option-card--selected .option-title {
  color: #1e293b;
}

.option-subtitle {
  font-size: 13px;
  color: #64748b;
  line-height: 1.5;
  margin-bottom: 8px;
}

.option-btn-row {
  margin-top: 6px;
}

.option-btn {
  display: inline-block;
  padding: 3px 14px;
  border-radius: 999px;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
  transition: transform 0.2s ease;
}

.copy-option-card:hover .option-btn {
  transform: scale(1.03);
}

/* 选中时按钮色与卡片边框色同步 */
.copy-option-card--selected .option-btn {
  box-shadow: 0 3px 12px rgba(64, 158, 255, 0.3);
}

/* ===== 选中高亮光效 ===== */
.card-glow {
  position: absolute;
  inset: 0;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.03), rgba(64, 158, 255, 0.06));
  pointer-events: none;
  z-index: 0;
}

/* ===== 底部 ===== */
.copy-picker-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0;
}

.footer-preview {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}

.footer-preview-label {
  font-size: 12px;
  color: #94a3b8;
  white-space: nowrap;
}

.footer-preview-text {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 220px;
}

.footer-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.footer-btn {
  border-radius: 10px;
  padding: 8px 20px;
  font-weight: 500;
  transition: all 0.25s ease;
}

.footer-btn--cancel {
  border-color: #e2e8f0;
  color: #64748b;
}

.footer-btn--cancel:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
}

.footer-btn--confirm {
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.25);
}

.footer-btn--confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.35);
}

.copy-picker-dialog :deep(.el-dialog__footer) {
  padding: 12px 24px;
  border-top: 1px solid #e2e8f0;
  background: #fafbfc;
}
</style>
