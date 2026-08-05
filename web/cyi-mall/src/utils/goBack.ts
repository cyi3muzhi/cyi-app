/**
 * 统一返回上一页
 * 优先返回浏览器上一页（从哪个页面进来就回哪个页面）；
 * 无历史记录（如直接刷新打开、外部进入）时跳转到 fallback 兜底页面。
 * @param fallback 无历史记录时的兜底跳转地址
 */
import router from '@/router'

export function goBack(fallback: string) {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push(fallback)
  }
}
