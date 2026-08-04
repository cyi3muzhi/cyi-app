/**
 * 轻量全局 Toast 提示（无第三方依赖）
 * 用于加入购物车、操作反馈等场景
 */

let container: HTMLDivElement | null = null

/**
 * 弹出 Toast 提示
 * @param message 提示文案
 * @param type 类型 success=成功 error=失败
 */
export function toast(message: string, type: 'success' | 'error' = 'success') {
  // 懒创建容器（挂在 body 上，避免被页面布局影响）
  if (!container) {
    container = document.createElement('div')
    container.style.cssText =
      'position:fixed;top:20px;left:50%;transform:translateX(-50%);z-index:9999;' +
      'display:flex;flex-direction:column;align-items:center;gap:8px;pointer-events:none;'
    document.body.appendChild(container)
  }

  const el = document.createElement('div')
  const bg = type === 'success' ? '#10b981' : '#ef4444'
  el.style.cssText =
    `background:${bg};color:#fff;padding:10px 22px;border-radius:999px;font-size:0.875rem;` +
    'font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.18);white-space:nowrap;' +
    'opacity:0;transform:translateY(-8px);transition:all 0.3s ease;'
  el.textContent = message
  container.appendChild(el)

  // 触发进入动画
  requestAnimationFrame(() => {
    el.style.opacity = '1'
    el.style.transform = 'translateY(0)'
  })

  // 自动消失
  setTimeout(() => {
    el.style.opacity = '0'
    el.style.transform = 'translateY(-8px)'
    setTimeout(() => el.remove(), 300)
  }, 2000)
}
