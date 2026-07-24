/**
 * 悦选商城管理后台 - 智能体 API 封装
 */
import request from './index'

export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp?: number
}

/**
 * 发送聊天消息（模拟流式响应）
 * @param message 用户输入消息
 * @returns Promise<string> AI 回复内容
 */
export async function sendChatMessage(message: string): Promise<string> {
  try {
    const res = await request.post<any, { message: string }>('/blades/v1/chat', { message })
    return res.message || ''
  } catch (error) {
    console.error('智能体聊天失败:', error)
    return '抱歉，智能体服务暂时不可用，请稍后再试。'
  }
}

/**
 * 发送聊天消息（流式模式）
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
    const response = await fetch('/blades/v1/chat', {
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
            }
          } catch {
            if (data) {
              onChunk(data)
            }
          }
        } else if (trimmed && !trimmed.startsWith('event:')) {
          onChunk(trimmed)
        }
      }
    }

    if (buffer) {
      onChunk(buffer)
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
