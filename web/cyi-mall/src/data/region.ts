/**
 * 中国省市区数据（完整版，基于 china-division v2.7.0）
 * 数据来源：国家统计局 2023 年统计用区划代码和城乡划分代码（截止 2023-06-30）
 * 覆盖：全国 31 个大陆省级（含 code）+ 港澳台（无官方 code，用名称作为 code）
 */

import pcaCodeRaw from 'china-division/dist/pca-code.json'
import hkmoTwRaw from 'china-division/dist/HK-MO-TW.json'

/** pca-code.json 节点结构：省 -> 市 -> 区县 */
interface AreaNode {
  code: string
  name: string
  children?: AreaNode[]
}

/** 港澳台结构：{ 省: { 市: [区县] } } */
type HKMoTwData = Record<string, Record<string, string[]>>

/** 省/市/区选项 */
export interface AreaOption {
  code: string // 行政区划代码（港澳台为名称）
  name: string // 名称
}

const pcaCode = pcaCodeRaw as AreaNode[]
const hkmoTw = hkmoTwRaw as HKMoTwData

// 合并大陆 + 港澳台（港澳台无官方 code，用名称作为 code）
const merged: AreaNode[] = [
  ...pcaCode,
  ...Object.entries(hkmoTw).map(([province, cityMap]) => ({
    code: province,
    name: province,
    children: Object.entries(cityMap).map(([city, districts]) => ({
      code: city,
      name: city,
      children: districts.map(d => ({ code: d, name: d }))
    }))
  }))
]

/** 省份列表（全国 34 个省级行政区） */
export const PROVINCES: AreaOption[] = merged.map(p => ({ code: p.code, name: p.name }))

/**
 * 获取指定省份的城市列表
 * @param provinceCode - 省份行政区划代码
 * @returns 城市选项列表
 */
export function getCityOptions(provinceCode: string): AreaOption[] {
  const province = merged.find(p => p.code === provinceCode)
  return (province?.children || []).map(c => ({ code: c.code, name: c.name }))
}

/**
 * 获取指定城市的区县列表
 * @param cityCode - 城市行政区划代码
 * @returns 区县选项列表
 */
export function getDistrictOptions(cityCode: string): AreaOption[] {
  let districts: AreaNode[] = []
  for (const p of merged) {
    const city = p.children?.find(c => c.code === cityCode)
    if (city) {
      districts = city.children || []
      break
    }
  }
  return districts.map(d => ({ code: d.code, name: d.name }))
}

/**
 * 根据省份名称反查行政区划代码（用于表单回显联动）
 * @param name - 省份名称
 * @returns 省份代码，未找到返回空串
 */
export function findProvinceCode(name: string): string {
  return PROVINCES.find(p => p.name === name)?.code || ''
}

/**
 * 根据城市名称反查行政区划代码（用于表单回显联动）
 * @param provinceCode - 省份代码
 * @param name - 城市名称
 * @returns 城市代码，未找到返回空串
 */
export function findCityCode(provinceCode: string, name: string): string {
  return getCityOptions(provinceCode).find(c => c.name === name)?.code || ''
}
