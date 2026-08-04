<template>
  <!-- ============================================================ -->
  <!-- 1. 页面标题栏                                                  -->
  <!-- ============================================================ -->
  <div class="page-header">
    <div class="container-shop">
      <div class="page-header-inner">
        <router-link to="/user" class="back-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </router-link>
        <h1 class="page-title">收货地址</h1>
        <button class="add-btn" @click="openCreate">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新增地址
        </button>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 2. 地址列表                                                   -->
  <!-- ============================================================ -->
  <div class="address-list">
    <div class="container-shop">
      <!-- 加载中 -->
      <div v-if="loading" class="state-wrap">
        <p class="state-text">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="addresses.length === 0" class="state-wrap">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="var(--shop-border)" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>
        </svg>
        <p class="state-text">还没有收货地址，点击右上角"新增地址"添加</p>
      </div>

      <!-- 地址卡片 -->
      <div v-for="addr in addresses" :key="addr.id" class="address-card" :class="{ 'card-default': addr.isDefault === 1 }">
        <div class="card-main" @click="openEdit(addr)">
          <div class="card-top">
            <div class="receiver-info">
              <span class="receiver-name">{{ addr.receiver }}</span>
              <span class="receiver-phone">{{ addr.phone }}</span>
              <span v-if="addr.isDefault === 1" class="default-badge">默认</span>
            </div>
          </div>
          <p class="address-detail">
            {{ [addr.province, addr.city, addr.district].filter(Boolean).join(' ') }} {{ addr.detail }}
          </p>
        </div>
        <div class="card-actions">
          <button
            v-if="addr.isDefault !== 1"
            class="action-btn"
            @click="onSetDefault(addr)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>
            </svg>
            设为默认
          </button>
          <button class="action-btn" @click="openEdit(addr)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            编辑
          </button>
          <button class="action-btn danger" @click="onDelete(addr)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
            删除
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- ============================================================ -->
  <!-- 3. 新增/编辑地址弹层                                            -->
  <!-- ============================================================ -->
  <div v-if="showForm" class="modal-mask" @click.self="closeForm">
    <div class="modal-card">
      <h2 class="modal-title">{{ form.id ? '编辑地址' : '新增地址' }}</h2>

      <div class="form-group">
        <label class="form-label">收货人</label>
        <input v-model="form.receiver" class="form-input" type="text" placeholder="请输入收货人姓名" />
      </div>

      <div class="form-group">
        <label class="form-label">手机号</label>
        <input v-model="form.phone" class="form-input" type="tel" placeholder="请输入手机号" />
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">省份</label>
          <select v-model="form.province" class="form-select" @change="onProvinceChange">
            <option value="">请选择省份</option>
            <option v-for="p in provinces" :key="p.code" :value="p.name">{{ p.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">城市</label>
          <select v-model="form.city" class="form-select" :disabled="!form.province" @change="onCityChange">
            <option value="">请选择城市</option>
            <option v-for="c in cityOptions" :key="c.code" :value="c.name">{{ c.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">区县</label>
          <select v-model="form.district" class="form-select" :disabled="!form.city">
            <option value="">请选择区县</option>
            <option v-for="d in districtOptions" :key="d.code" :value="d.name">{{ d.name }}</option>
          </select>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">详细地址</label>
        <textarea v-model="form.detail" class="form-textarea" rows="2" placeholder="街道、门牌号等详细信息"></textarea>
      </div>

      <label class="default-check">
        <input v-model="form.isDefault" type="checkbox" :value="1" :checked="form.isDefault === 1" />
        <span>设为默认地址</span>
      </label>

      <div v-if="formError" class="form-error">{{ formError }}</div>

      <div class="modal-actions">
        <button class="modal-btn cancel" @click="closeForm">取消</button>
        <button class="modal-btn submit" :disabled="saving" @click="onSave">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 收货地址管理页面
 * 列表 + 新增/编辑/删除/设为默认，数据来源：/user/v1/address/*
 */
import { ref, computed, onMounted } from 'vue'
import {
  listAddress,
  createAddress,
  updateAddress,
  deleteAddress,
  setDefaultAddress
} from '@/api/user'
import {
  PROVINCES,
  getCityOptions,
  getDistrictOptions,
  findProvinceCode,
  findCityCode,
  type AreaOption
} from '@/data/region'
import type { Address, AddressPayload } from '@/types/user'

const addresses = ref<Address[]>([])
const loading = ref(true)

// ---- 省市区联动选项（完整数据，基于行政区划代码联动） ----
const provinces = PROVINCES
/** 当前选中省份的行政区划代码 */
const provinceCode = computed(() => findProvinceCode(form.value.province))
/** 当前选中省份的城市选项 */
const cityOptions = computed<AreaOption[]>(() => getCityOptions(provinceCode.value))
/** 当前选中城市的行政区划代码 */
const cityCode = computed(() => findCityCode(provinceCode.value, form.value.city))
/** 当前选中城市的区县选项 */
const districtOptions = computed<AreaOption[]>(() => getDistrictOptions(cityCode.value))

// ---- 表单状态 ----
const showForm = ref(false)
const saving = ref(false)
const formError = ref('')
const form = ref<AddressPayload>({
  receiver: '',
  phone: '',
  province: '',
  city: '',
  district: '',
  detail: '',
  isDefault: 0
})

/** 加载地址列表 */
async function fetchAddresses() {
  loading.value = true
  try {
    const res = await listAddress()
    addresses.value = res.items || []
  } catch {
    addresses.value = []
  } finally {
    loading.value = false
  }
}

/** 打开新增表单 */
function openCreate() {
  form.value = { receiver: '', phone: '', province: '', city: '', district: '', detail: '', isDefault: 0 }
  formError.value = ''
  showForm.value = true
}

/** 打开编辑表单 */
function openEdit(addr: Address) {
  form.value = {
    id: addr.id,
    receiver: addr.receiver,
    phone: addr.phone,
    province: addr.province,
    city: addr.city,
    district: addr.district,
    detail: addr.detail,
    isDefault: addr.isDefault
  }
  formError.value = ''
  showForm.value = true
}

/** 选择省份：重置城市和区县 */
function onProvinceChange() {
  form.value.city = ''
  form.value.district = ''
}

/** 选择城市：重置区县 */
function onCityChange() {
  form.value.district = ''
}

/** 关闭表单 */
function closeForm() {
  showForm.value = false
  formError.value = ''
}

/** 表单校验 */
function validate(): boolean {
  if (!form.value.receiver.trim()) {
    formError.value = '请填写收货人姓名'
    return false
  }
  if (!form.value.phone.trim() || !/^1\d{10}$/.test(form.value.phone.trim())) {
    formError.value = '请填写正确的手机号'
    return false
  }
  if (!form.value.detail.trim()) {
    formError.value = '请填写详细地址'
    return false
  }
  return true
}

/** 保存（新增或编辑） */
async function onSave() {
  if (!validate()) return
  saving.value = true
  formError.value = ''
  try {
    if (form.value.id) {
      await updateAddress({ ...form.value, isDefault: form.value.isDefault ? 1 : 0 })
    } else {
      await createAddress({ ...form.value, isDefault: form.value.isDefault ? 1 : 0 })
    }
    closeForm()
    fetchAddresses()
  } catch (e: any) {
    formError.value = e?.response?.data?.message || e?.message || '保存失败，请重试'
  } finally {
    saving.value = false
  }
}

/** 设为默认地址 */
async function onSetDefault(addr: Address) {
  try {
    await setDefaultAddress(addr.id)
    fetchAddresses()
  } catch (e: any) {
    alert(e?.response?.data?.message || e?.message || '设置失败，请重试')
  }
}

/** 删除地址 */
async function onDelete(addr: Address) {
  if (!window.confirm(`确定删除 ${addr.receiver} 的收货地址吗？`)) return
  try {
    await deleteAddress(addr.id)
    fetchAddresses()
  } catch (e: any) {
    alert(e?.response?.data?.message || e?.message || '删除失败，请重试')
  }
}

onMounted(fetchAddresses)
</script>

<style scoped>
/* ============================================
   收货地址页面样式
   ============================================ */

/* ---- 页面标题 ---- */
.page-header {
  padding: 24px 0 0;
}

.page-header-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-link {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--shop-muted);
  color: var(--shop-foreground);
  transition: background 0.2s;
}

.back-link:hover {
  background: var(--shop-border);
}

.page-title {
  flex: 1;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-radius: var(--shop-radius-lg);
  border: none;
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.add-btn:hover {
  opacity: 0.9;
}

/* ---- 地址列表 ---- */
.address-list {
  padding: 24px 0 48px;
}

.state-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 24px;
}

.state-text {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.address-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 24px;
  margin-bottom: 16px;
  background: var(--shop-card);
  border: 1px solid var(--shop-border);
  border-radius: var(--shop-radius-xl);
  transition: box-shadow 0.2s, border-color 0.2s;
}

.address-card:hover {
  box-shadow: var(--shop-shadow-lg);
}

.card-default {
  border-color: var(--shop-primary);
}

.card-main {
  flex: 1;
  cursor: pointer;
  min-width: 0;
}

.card-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.receiver-info {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.receiver-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--shop-foreground);
}

.receiver-phone {
  font-size: 0.875rem;
  color: var(--shop-muted-foreground);
}

.default-badge {
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.address-detail {
  margin-top: 8px;
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--shop-muted-foreground);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  border-radius: var(--shop-radius-md);
  border: none;
  background: transparent;
  font-size: 0.75rem;
  color: var(--shop-muted-foreground);
  cursor: pointer;
  transition: color 0.2s, background 0.2s;
}

.action-btn:hover {
  color: var(--shop-primary);
  background: var(--shop-primary-bg);
}

.action-btn.danger:hover {
  color: var(--shop-destructive);
  background: hsl(0, 84%, 95%);
}

/* ---- 弹层 ---- */
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: rgba(0, 0, 0, 0.45);
}

.modal-card {
  width: 100%;
  max-width: 460px;
  max-height: calc(100vh - 80px);
  overflow-y: auto;
  padding: 24px;
  background: var(--shop-card);
  border-radius: var(--shop-radius-xl);
  box-shadow: var(--shop-shadow-xl);
}

.modal-title {
  margin-bottom: 20px;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--shop-foreground);
}

.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  margin-bottom: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--shop-foreground);
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border-radius: var(--shop-radius-lg);
  border: 1px solid var(--shop-input);
  background: var(--shop-background);
  font-size: 0.875rem;
  color: var(--shop-foreground);
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  font-family: inherit;
}

.form-select {
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
  cursor: pointer;
}

.form-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  border-color: var(--shop-primary);
  box-shadow: 0 0 0 3px hsla(217, 91%, 52%, 0.15);
}

.form-textarea {
  resize: none;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.default-check {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 0.875rem;
  color: var(--shop-foreground);
  cursor: pointer;
}

.form-error {
  margin-bottom: 16px;
  padding: 10px 14px;
  border-radius: var(--shop-radius-lg);
  background: hsl(0, 84%, 95%);
  color: var(--shop-destructive);
  font-size: 0.8125rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.modal-btn {
  padding: 10px 24px;
  border-radius: var(--shop-radius-lg);
  border: none;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.modal-btn.cancel {
  background: var(--shop-muted);
  color: var(--shop-foreground);
}

.modal-btn.submit {
  background: var(--shop-primary);
  color: var(--shop-primary-foreground);
}

.modal-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ---- 响应式 ---- */
@media (max-width: 640px) {
  .address-card {
    flex-direction: column;
    align-items: flex-start;
    padding: 16px;
  }

  .card-actions {
    width: 100%;
    justify-content: flex-end;
    border-top: 1px solid var(--shop-border);
    padding-top: 12px;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
}
</style>
