/**
 * 搜索历史管理工具（基于 localStorage）
 * 每用户独立存储，最多 10 条，去重置顶，可清空/单删
 */

/** localStorage key 前缀 */
const STORAGE_KEY_PREFIX = 'cyi_search_history'

/** 最大历史记录条数 */
const MAX_HISTORY = 10

/**
 * 获取当前用户的搜索历史 key
 * 未登录时使用匿名 key，登录后按 userID 区分
 */
function getStorageKey(userId?: number | string | null): string {
  if (userId) {
    return `${STORAGE_KEY_PREFIX}_${userId}`
  }
  return `${STORAGE_KEY_PREFIX}_anonymous`
}

/**
 * 获取搜索历史列表（按时间倒序，最近搜索在前）
 * @param userId 当前用户 ID（未登录传 null/undefined）
 * @returns 搜索关键词数组
 */
export function getSearchHistory(userId?: number | string | null): string[] {
  try {
    const key = getStorageKey(userId)
    const raw = localStorage.getItem(key)
    if (!raw) return []
    const list: string[] = JSON.parse(raw)
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
}

/**
 * 新增一条搜索历史（去重置顶，超出上限截断）
 * @param keyword 搜索关键词
 * @param userId 当前用户 ID（未登录传 null/undefined）
 */
export function addSearchHistory(keyword: string, userId?: number | string | null): void {
  const kw = keyword.trim()
  if (!kw) return
  try {
    const list = getSearchHistory(userId)
    // 去重：移除已存在的相同关键词
    const filtered = list.filter(item => item !== kw)
    // 置顶
    filtered.unshift(kw)
    // 截断到最大条数
    const trimmed = filtered.slice(0, MAX_HISTORY)
    const key = getStorageKey(userId)
    localStorage.setItem(key, JSON.stringify(trimmed))
  } catch (e) {
    console.warn('[searchHistory] add failed:', e)
  }
}

/**
 * 删除单条搜索历史
 * @param keyword 要删除的关键词
 * @param userId 当前用户 ID
 */
export function removeSearchHistory(keyword: string, userId?: number | string | null): void {
  try {
    const list = getSearchHistory(userId)
    const filtered = list.filter(item => item !== keyword)
    const key = getStorageKey(userId)
    localStorage.setItem(key, JSON.stringify(filtered))
  } catch (e) {
    console.warn('[searchHistory] remove failed:', e)
  }
}

/**
 * 清空所有搜索历史
 * @param userId 当前用户 ID
 */
export function clearSearchHistory(userId?: number | string | null): void {
  try {
    const key = getStorageKey(userId)
    localStorage.removeItem(key)
  } catch (e) {
    console.warn('[searchHistory] clear failed:', e)
  }
}
