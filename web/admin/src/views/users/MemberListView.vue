<template>
  <!-- 会员管理占位页面 -->
  <div class="member-list">
    <div class="placeholder-card">
      <div class="placeholder-icon">
        <el-icon :size="64"><Medal /></el-icon>
      </div>
      <h2 class="placeholder-title">会员管理</h2>
      <p class="placeholder-desc">会员管理功能正在开发中，敬请期待...</p>
      <div class="placeholder-features">
        <div class="feature-item">
          <el-icon :size="20"><Cpu /></el-icon>
          <span>会员等级设置</span>
        </div>
        <div class="feature-item">
          <el-icon :size="20"><Present /></el-icon>
          <span>会员权益配置</span>
        </div>
        <div class="feature-item">
          <el-icon :size="20"><TrendCharts /></el-icon>
          <span>成长值体系</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 会员管理 - 占位页面
 * 功能开发中，后续完善
 */
import { Medal, Cpu, Present, TrendCharts } from '@element-plus/icons-vue'
</script>

<style scoped>
.member-list {
  padding: 40px;
}

.placeholder-card {
  background: #fff;
  border-radius: var(--shop-radius-xl);
  padding: 60px 40px;
  text-align: center;
  border: 1px solid var(--shop-border);
  box-shadow: var(--shop-shadow-sm);
}

.placeholder-icon {
  width: 96px;
  height: 96px;
  margin: 0 auto 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.placeholder-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--shop-foreground);
  margin: 0 0 12px;
}

.placeholder-desc {
  font-size: 1rem;
  color: var(--shop-muted-foreground);
  margin: 0 0 32px;
}

.placeholder-features {
  display: flex;
  justify-content: center;
  gap: 32px;
  flex-wrap: wrap;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: var(--shop-muted);
  border-radius: var(--shop-radius-lg);
  color: var(--shop-muted-foreground);
  font-size: 0.875rem;
}

@media (max-width: 768px) {
  .placeholder-features {
    flex-direction: column;
    gap: 12px;
  }
}
</style>
