/**
 * cyi商城管理后台 - Banner 管理 API
 */
import request from './index'

/** Banner 数据项接口 */
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

/** 获取所有 Banner 列表 */
export function listBanners() {
  return request.get<any, { banners: BannerItem[] }>('/admin/v1/banners')
}

/** 批量保存所有 Banner（全量覆盖） */
export function saveBanners(data: { banners: BannerItem[] }) {
  return request.put<any, { success: boolean }>('/admin/v1/banners', data)
}
