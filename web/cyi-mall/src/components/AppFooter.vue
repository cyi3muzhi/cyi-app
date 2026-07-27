/**
 * 页脚
 */
<template>
  <footer class="site-footer">
    <div class="container-shop">
      <div class="footer-grid">
        <!-- 品牌 -->
        <div class="footer-col">
          <h4 class="footer-title">{{ shopStore.shopName }}</h4>
          <p class="footer-desc">{{ shopStore.shopDescription || '精选好物，悦享生活。我们致力于为每一位用户提供品质优良、价格实惠的购物体验。' }}</p>
        </div>
        <!-- 购物指南 -->
        <div class="footer-col">
          <h4 class="footer-heading">购物指南</h4>
          <ul class="footer-links">
            <li><a href="#">购物流程</a></li>
            <li><a href="#">支付方式</a></li>
            <li><a href="#">配送说明</a></li>
            <li><a href="#">售后服务</a></li>
          </ul>
        </div>
        <!-- 服务保障 -->
        <div class="footer-col">
          <h4 class="footer-heading">服务保障</h4>
          <ul class="footer-links">
            <li><a href="#">退换货政策</a></li>
            <li><a href="#">品质保证</a></li>
            <li><a href="#">联系客服</a></li>
            <li><a href="#">隐私保护</a></li>
          </ul>
        </div>
        <!-- 联系我们 -->
        <div class="footer-col">
          <h4 class="footer-heading">联系我们</h4>
          <ul class="footer-contact">
            <li>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
              <span>北京市朝阳区建国路88号</span>
            </li>
            <li>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.362 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.338 1.85.573 2.81.7A2 2 0 0 1 22 16.92z"/></svg>
              <span>{{ shopStore.contactPhone || '400-888-9999' }}</span>
            </li>
            <li>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              <span>{{ shopStore.contactEmail || 'service@yuexuan.com' }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
    <div class="footer-bottom">
      <div class="container-shop footer-bottom-inner">
        <span>&copy; 2026 {{ shopStore.shopName }}. All rights reserved.</span>
        <div class="footer-bottom-links">
          <a href="#">隐私政策</a>
          <a href="#">服务条款</a>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { useShopStore } from '@/store/shop'

const shopStore = useShopStore()
</script>

<style scoped>
.site-footer {
  background: hsl(220, 30%, 10%);
  color: white;
}

.footer-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 32px;
  padding: 48px 0;
}
@media (min-width: 640px) {
  .footer-grid { grid-template-columns: 1fr 1fr; }
}
@media (min-width: 1024px) {
  .footer-grid { grid-template-columns: 1.5fr 1fr 1fr 1fr; }
}

.footer-title {
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 16px;
  color: white;
}

.footer-desc {
  font-size: 0.875rem;
  line-height: 1.7;
  color: hsl(220, 20%, 70%);
}

.footer-heading {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 16px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: hsl(220, 20%, 70%);
}

.footer-links {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.footer-links a {
  font-size: 0.875rem;
  color: hsl(220, 20%, 75%);
  text-decoration: none;
  transition: color 0.2s;
}
.footer-links a:hover {
  color: white;
}

.footer-contact {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 0.875rem;
  color: hsl(220, 20%, 75%);
}
.footer-contact li {
  display: flex;
  align-items: center;
  gap: 8px;
}
.footer-contact svg {
  flex-shrink: 0;
}

.footer-bottom {
  border-top: 1px solid hsl(220, 25%, 18%);
}

.footer-bottom-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 16px 0;
  font-size: 0.75rem;
  color: hsl(220, 20%, 55%);
}
@media (min-width: 640px) {
  .footer-bottom-inner {
    flex-direction: row;
  }
}

.footer-bottom-links {
  display: flex;
  gap: 16px;
}
.footer-bottom-links a {
  color: inherit;
  text-decoration: none;
}
.footer-bottom-links a:hover {
  text-decoration: underline;
}
</style>
