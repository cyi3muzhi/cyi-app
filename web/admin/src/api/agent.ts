/**
 * cyi商城管理后台 - 智能体 API 封装
 * 接口由 admin BFF (cin/admin) 提供，路径前缀 /admin/v1/agent
 */
import request from './index'

export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp?: number
}

/**
 * 清除当前管理员的聊天历史
 */
export async function clearChatHistory(): Promise<void> {
  await request.post('/admin/v1/agent/chat/clear')
}

/**
 * 获取当前管理员的聊天历史
 */
export async function getChatHistory(): Promise<ChatMessage[]> {
  try {
    const res = await request.get<any, any[]>('/admin/v1/agent/chat/history')
    return res || []
  } catch {
    return []
  }
}

/**
 * 发送聊天消息（非流式）
 * @param message 用户输入消息
 * @returns Promise<string> AI 回复内容
 */
export async function sendChatMessage(message: string): Promise<string> {
  try {
    const res = await request.post<any, { message: string }>('/admin/v1/agent/chat', { message })
    return res.message || ''
  } catch (error) {
    console.error('智能体聊天失败:', error)
    return '抱歉，智能体服务暂时不可用，请稍后再试。'
  }
}

/**
 * AI 生成商品描述 - 调用专用接口
 * @param productName 商品名称
 * @param category 商品分类
 * @param specs 规格信息（可选）
 * @returns Promise<string> 生成的描述文本
 */
export async function generateProductDescription(
  productName: string,
  category?: string,
  specs?: string
): Promise<string> {
  try {
    const res = await request.post<any, { description: string }>('/admin/v1/agent/generate-description', {
      product_name: productName,
      category: category || '',
      specs: specs || ''
    })
    return res.description || ''
  } catch (error) {
    console.error('AI 生成描述失败:', error)
    return ''
  }
}

/**
 * AI 扩写商品描述 - 调用专用接口
 * @param existingDesc 现有描述文本
 * @param productName 商品名称
 * @returns Promise<string> 扩写后的文本
 */
export async function expandProductDescription(
  existingDesc: string,
  productName: string
): Promise<string> {
  try {
    const res = await request.post<any, { description: string }>('/admin/v1/agent/expand-description', {
      product_name: productName,
      content: existingDesc
    })
    return res.description || ''
  } catch (error) {
    console.error('AI 扩写失败:', error)
    return ''
  }
}

/** Banner 文案方案 */
export interface BannerCopyOption {
  title: string
  subtitle: string
  btn_text: string
}

/**
 * AI 生成 Banner 文案
 */
export async function generateBannerCopy(topic: string, subtitle?: string): Promise<BannerCopyOption[]> {
  try {
    const res = await request.post<any, { options: BannerCopyOption[] }>('/admin/v1/agent/generate-banner', {
      topic,
      subtitle: subtitle || ''
    })
    return res.options || []
  } catch (error) {
    console.error('AI 生成 Banner 文案失败:', error)
    return []
  }
}

/**
 * AI 生成分类 SEO 信息
 */
export async function generateCategorySEO(categoryName: string, description?: string): Promise<{
  seoTitle: string
  seoKeywords: string
  seoDescription: string
}> {
  try {
    const res = await request.post<any, { seoTitle: string; seoKeywords: string; seoDescription: string }>('/admin/v1/agent/generate-seo', {
      category_name: categoryName,
      description: description || ''
    })
    return res
  } catch (error) {
    console.error('AI 生成 SEO 失败:', error)
    return { seoTitle: '', seoKeywords: '', seoDescription: '' }
  }
}

/**
 * AI 生成运营摘要
 */
export async function generateSummary(data: {
  today_orders: number
  today_revenue: number
  total_users: number
  total_products: number
  orders_trend?: string
}): Promise<string> {
  try {
    const res = await request.post<any, { summary: string }>('/admin/v1/agent/generate-summary', data)
    return res.summary || ''
  } catch (error) {
    console.error('AI 生成运营摘要失败:', error)
    return ''
  }
}

/** 推荐问题 */
export interface SuggestQuestion {
  label: string
  query: string
}

/**
 * AI 推荐快捷提问
 */
export async function suggestQuestions(): Promise<SuggestQuestion[]> {
  try {
    const res = await request.post<any, { questions: SuggestQuestion[] }>('/admin/v1/agent/suggest-questions')
    return res.questions || []
  } catch (error) {
    console.error('获取推荐问题失败:', error)
    return []
  }
}

/**
 * 发送聊天消息（SSE 流式模式）
 * @param message 用户输入消息
 * @param onChunk 接收流式数据的回调函数
 * @param onError 错误回调函数
 */
export async function sendChatMessageStream(
  message: string,
  onChunk: (chunk: string) => void,
  onError?: (error: Error) => void
): Promise<void> {
  try {
    const token = localStorage.getItem('admin_token')
    const baseURL = window.location.origin
    const url = `${baseURL}/admin/v1/agent/chat/stream`
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      body: JSON.stringify({ message })
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('无法读取响应流')
    }

    const decoder = new TextDecoder('utf-8')
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmed = line.trim()
        if (trimmed.startsWith('data:')) {
          const data = trimmed.slice(5).trim()
          if (data === '[DONE]') {
            return
          }
          try {
            const json = JSON.parse(data)
            if (json.message) {
              onChunk(json.message)
            } else if (json.chunk) {
              onChunk(json.chunk)
            }
          } catch {
            // data 不是 JSON 时，直接输出文本
            if (data) {
              onChunk(data)
            }
          }
        }
        // event: 行、空行等非 data 行直接忽略，不传给用户
      }
    }

    // 剩余 buffer 只输出 data 行内容
    if (buffer) {
      const trimmed = buffer.trim()
      if (trimmed.startsWith('data:')) {
        const data = trimmed.slice(5).trim()
        if (data && data !== '[DONE]') {
          try {
            const json = JSON.parse(data)
            if (json.chunk) onChunk(json.chunk)
            else if (json.message) onChunk(json.message)
          } catch {
            onChunk(data)
          }
        }
      }
    }
  } catch (error) {
    console.error('智能体流式聊天失败:', error)
    if (onError) {
      onError(error as Error)
    } else {
      throw error
    }
  }
}
