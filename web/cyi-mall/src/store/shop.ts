/**
 * 店铺设置状态管理
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getPublicSettings } from '@/api/settings'

export const useShopStore = defineStore('shop', () => {
  /** 店铺名称 */
  const shopName = ref('cyi商城')
  /** 店铺描述 */
  const shopDescription = ref('')
  /** 联系电话 */
  const contactPhone = ref('')
  /** 联系邮箱 */
  const contactEmail = ref('')
  /** ICP 备案号 */
  const icpRecord = ref('')
  /** 是否已加载 */
  const loaded = ref(false)

  /** 从后端加载店铺设置 */
  async function fetchSettings() {
    try {
      const data: any = await getPublicSettings()
      console.log('[settings] response:', data)
      // 兼容 camelCase 和 snake_case 两种命名
      shopName.value = data.shopName || data.shop_name || 'cyi商城'
      if (data.shopDescription || data.shop_description) shopDescription.value = data.shopDescription || data.shop_description
      if (data.contactPhone || data.contact_phone) contactPhone.value = data.contactPhone || data.contact_phone
      if (data.contactEmail || data.contact_email) contactEmail.value = data.contactEmail || data.contact_email
      if (data.icpRecord || data.icp_record) icpRecord.value = data.icpRecord || data.icp_record
      loaded.value = true
    } catch (e) {
      console.warn('[settings] fetch failed:', e)
    }
  }

  return { shopName, shopDescription, contactPhone, contactEmail, icpRecord, loaded, fetchSettings }
})
