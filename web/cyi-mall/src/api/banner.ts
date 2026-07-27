/**
 * Banner 相关 API
 */
import request from './index'

export interface BannerItem {
  id: number
  title: string
  subtitle: string
  image: string
  link: string
  sort_order: number
  enabled: boolean
  bg_color: string
  btn_text: string
  /** 模板类型：classic | full-image | centered | dark | minimal */
  template?: string
  /** 文字颜色 */
  text_color?: string
}

/** 获取 Banner 列表（只返回启用的） */
export function listBanners() {
  return request.get<any, { banners: BannerItem[] }>('/settings/v1/banners')
}
