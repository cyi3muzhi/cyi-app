/**
 * cyi商城管理后台 - 知识库 API 封装
 * 接口由 admin BFF (cin/admin) 提供，路径前缀 /admin/v1/agent/knowledge
 */
import request from './index'

/** 知识文档信息 */
export interface KnowledgeDocInfo {
  docId: string
  source: string
  title: string
  chunkCount: number
  tags: Record<string, string>
}

/** 知识搜索结果 */
export interface KnowledgeChunk {
  docId: string
  title: string
  content: string
  score: number
}

/** 上传响应 */
export interface UploadResponse {
  docId: string
  chunkCount: number
  title: string
}

/**
 * 列出知识库文档
 * @param source 来源过滤（可选，如 'admin'）
 */
export async function listKnowledge(source?: string): Promise<KnowledgeDocInfo[]> {
  const params = source ? { source } : {}
  const res = await request.get<any, { docs: KnowledgeDocInfo[] }>('/admin/v1/agent/knowledge/list', { params })
  return res.docs || []
}

/**
 * 删除知识库文档
 * @param docId 文档 ID
 * @param source 来源
 */
export async function deleteKnowledge(docId: string, source: string): Promise<void> {
  await request.post('/admin/v1/agent/knowledge/delete', { doc_id: docId, source })
}

/**
 * 搜索知识库
 * @param query 搜索关键词
 * @param source 来源过滤（可选）
 * @param topK 返回数量上限（默认 5）
 */
export async function searchKnowledge(query: string, source?: string, topK: number = 5): Promise<KnowledgeChunk[]> {
  const res = await request.post<any, { chunks: KnowledgeChunk[] }>('/admin/v1/agent/knowledge/search', {
    query,
    source: source || '',
    top_k: topK
  })
  return res.chunks || []
}

/**
 * 上传知识库文件
 * @param file 文件对象
 * @param title 文档标题（可选，默认使用文件名）
 * @param source 来源（默认 'admin'）
 * @param tags 自定义标签（可选）
 */
export async function uploadKnowledge(
  file: File,
  title?: string,
  source: string = 'admin',
  tags?: Record<string, string>
): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('source', source)
  if (title) {
    formData.append('title', title)
  }
  if (tags) {
    formData.append('tags', JSON.stringify(tags))
  }

  const token = localStorage.getItem('admin_token')
  const baseURL = window.location.origin
  const url = `${baseURL}/admin/v1/agent/knowledge/upload`

  const response = await fetch(url, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: formData
  })

  if (!response.ok) {
    const errText = await response.text()
    let errMsg = '上传失败'
    try {
      const errObj = JSON.parse(errText)
      errMsg = errObj.message || errMsg
    } catch {
      errMsg = errText || errMsg
    }
    throw new Error(errMsg)
  }

  const result = await response.json()
  if (result.code !== 0) {
    throw new Error(result.message || '上传失败')
  }
  return result.data as UploadResponse
}
