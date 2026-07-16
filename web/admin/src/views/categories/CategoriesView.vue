<template>
  <div class="categories-page">
    <div class="categories-layout">
      <div class="tree-panel">
        <div class="tree-panel-header">
          <div class="tree-panel-title">分类目录<span class="tree-panel-count">{{ totalCategoryCount }}</span></div>
          <div class="tree-ctrl-btns">
            <el-tooltip content="展开全部" placement="top"><button class="tree-ctrl-btn" @click="expandAll"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="7 13 12 18 17 13"/><polyline points="7 6 12 11 17 6"/></svg></button></el-tooltip>
            <el-tooltip content="收起全部" placement="top"><button class="tree-ctrl-btn" @click="collapseAll"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="7 11 12 6 17 11"/><polyline points="7 18 12 13 17 18"/></svg></button></el-tooltip>
          </div>
        </div>
        <div class="tree-search-wrap"><div class="tree-search-box"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg><input v-model="treeSearch" type="text" placeholder="搜索分类名称..." /></div></div>
        <div class="tree-panel-body" id="treeBody">
          <div v-for="item in treeNodes" :key="item.id" :class="['tree-item', { active: selectedNodeId === item.id }]" :style="{ paddingLeft: (16 + item._depth * 20) + 'px' }">
            <span v-if="item.hasChildren" :class="['tree-toggle', { expanded: expandedState[item.id] !== false }]" @click="toggleTreeNode(item.id)"><svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="9 18 15 12 9 6"/></svg></span>
            <span v-else class="tree-toggle-placeholder"></span>
            <span class="tree-icon">{{ item.icon || '📁' }}</span>
            <span class="tree-label" @click="selectTreeNode(item.id)">{{ item.name }}</span>
            <span class="tree-count">{{ item.count }}</span>
          </div>
          <div v-if="treeNodes.length === 0" class="tree-empty">无匹配分类</div>
        </div>
        <div class="tree-panel-footer"><span class="total-label">共 <span class="total-num">{{ totalCategoryCount }}</span> 个分类</span><span class="tree-footer-detail">{{ topLevelCount }} 个一级 · {{ subLevelCount }} 个二级</span></div>
      </div>
      <div class="content-panel">
        <div class="breadcrumb-bar">
          <template v-for="(bc, idx) in breadcrumbs" :key="bc.id"><span v-if="idx > 0" class="bc-sep">/</span><span v-if="idx < breadcrumbs.length - 1" class="bc-link" @click="selectTreeNode(bc.id)">{{ bc.name }}</span><span v-else class="bc-current">{{ bc.icon }} {{ bc.name }}</span></template>
          <span v-if="breadcrumbs.length === 0" class="bc-current">全部分类</span>
        </div>
        <div class="table-card">
          <div class="table-toolbar">
            <el-button type="primary" @click="openAddModal"><el-icon><Plus /></el-icon>添加子分类</el-button>
            <el-dropdown trigger="click" @command="handleBatchCommand"><el-button class="btn-secondary btn-sm">批量操作<el-icon><ArrowDown /></el-icon></el-button>
              <template #dropdown><el-dropdown-menu><el-dropdown-item command="enable">批量启用</el-dropdown-item><el-dropdown-item command="disable">批量禁用</el-dropdown-item><el-dropdown-item command="delete" divided style="color: #ef4444;">批量删除</el-dropdown-item></el-dropdown-menu></template>
            </el-dropdown>
            <div class="toolbar-spacer"></div>
            <div class="stats-label">当前分类 <span class="stat-num">{{ currentChildren.length }}</span> 个<span><span class="stat-dot green"></span>已启用 <span class="stat-num">{{ enabledCount }}</span></span><span><span class="stat-dot red"></span>已禁用 <span class="stat-num">{{ disabledCount }}</span></span></div>
          </div>
          <div class="table-wrap">
            <table class="sub-table">
              <thead><tr><th style="width:80px;">排序</th><th>分类名称</th><th>别名/Slug</th><th>状态</th><th style="width:90px;">商品数量</th><th style="width:110px;">展示</th><th style="width:130px;">操作</th></tr></thead>
              <tbody id="tableBody">
                <tr v-if="currentChildren.length === 0"><td colspan="7"><div class="table-empty"><div class="empty-icon">📂</div><div class="empty-text">暂无子分类</div><div class="empty-hint">点击"添加子分类"按钮创建</div></div></td></tr>
                <tr v-for="item in currentChildren" :key="item.id">
                  <td><input type="number" class="sort-input" :value="item.sort" min="0" @change="updateSort(item, ($event.target as HTMLInputElement).value)" /></td>
                  <td><div class="cat-name"><span class="cat-icon">{{ item.icon || '📁' }}</span><span>{{ item.name }}</span><span v-if="item.children && item.children.length > 0" class="child-badge" @click="selectTreeNode(item.id)">{{ item.children.length }} 子级</span></div></td>
                  <td><code class="slug-text">{{ item.slug }}</code></td>
                  <td><label class="toggle-switch" @click="toggleStatus(item)"><span :class="['toggle-track', { active: item.status }]"><span class="toggle-thumb"></span></span><span :class="['toggle-label', { active: item.status }]">{{ item.status ? '启用' : '禁用' }}</span></label></td>
                  <td>{{ item.count || 0 }}</td>
                  <td><button :class="['showcase-star', { active: item.showcase }]" @click="toggleShowcase(item)" :title="item.showcase ? '取消首页展示' : '设为首页展示'">{{ item.showcase ? '⭐' : '☆' }}</button></td>
                  <td><div class="action-links"><button class="action-link edit-link" @click="openEditModal(item)">编辑</button><span class="action-divider"></span><button class="action-link add-link" @click="openAddSubModal(item)">添加子类</button><span class="action-divider"></span><button class="action-link delete-link" @click="confirmDelete(item)">删除</button></div></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="table-footer"><div class="pagination-info">显示 <span class="num">1</span>-<span class="num">{{ currentChildren.length }}</span> 条，共 <span class="num">{{ currentChildren.length }}</span> 个分类</div></div>
        </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEditing ? '编辑分类' : '添加分类'" width="640px" :close-on-click-modal="false" class="category-dialog">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px" size="default">
        <el-form-item label="分类名称" prop="name"><el-input v-model="formData.name" placeholder="请输入分类名称" maxlength="20" show-word-limit /></el-form-item>
        <el-form-item label="别名/Slug" prop="slug"><el-input v-model="formData.slug" placeholder="自动生成，可手动修改" /><span class="form-hint">用于 URL 友好链接，如: iphone-series</span></el-form-item>
        <div class="form-row">
          <el-form-item label="上级分类" label-width="100px" style="flex:1;"><el-select v-model="formData.parentId" placeholder="无（顶级分类）" clearable style="width:100%"><el-option label="无（顶级分类）" :value="null" /><el-option v-for="opt in parentOptions" :key="opt.id" :label="opt.label" :value="opt.id" :disabled="opt.id === editingId" /></el-select></el-form-item>
          <el-form-item label="排序" label-width="60px" style="flex:0 0 160px;"><el-input-number v-model="formData.sort" :min="0" :max="999" controls-position="right" style="width:100%" /></el-form-item>
        </div>
        <el-form-item label="分类图标"><div class="emoji-picker-label"><span class="selected-emoji">{{ formData.icon || '📁' }}</span></div><div class="emoji-grid"><button v-for="emoji in emojiList" :key="emoji" type="button" :class="['emoji-opt', { selected: formData.icon === emoji }]" @click="formData.icon = emoji">{{ emoji }}</button></div></el-form-item>
        <el-form-item label="分类图片"><div class="upload-area" @click="triggerUpload"><img v-if="formData.image" :src="formData.image" class="upload-preview" /><template v-else><div class="upload-icon"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div><div class="upload-text">点击上传分类图片</div><div class="upload-hint">支持 JPG、PNG，建议尺寸 200x200px</div></template></div><input ref="fileInputRef" type="file" accept="image/*" style="display:none" @change="handleUpload" /></el-form-item>
        <el-form-item label="状态"><label class="toggle-switch" @click="formData.status = !formData.status"><span :class="['toggle-track', { active: formData.status }]"><span class="toggle-thumb"></span></span><span :class="['toggle-label', { active: formData.status }]">{{ formData.status ? '启用' : '禁用' }}</span></label></el-form-item>
        <el-form-item label="SEO 设置" style="margin-bottom:0;"><div class="seo-section"><button type="button" class="seo-toggle" @click="seoOpen = !seoOpen">SEO 设置<span :class="['seo-arrow', { open: seoOpen }]">&#9660;</span></button><div v-show="seoOpen" class="seo-body"><el-form-item label="页面标题" label-width="80px"><el-input v-model="formData.seoTitle" placeholder="默认为分类名称" /></el-form-item><el-form-item label="SEO 关键词" label-width="80px"><el-input v-model="formData.seoKeywords" placeholder="多个关键词用逗号分隔" /></el-form-item><el-form-item label="SEO 描述" label-width="80px"><el-input v-model="formData.seoDesc" type="textarea" :rows="3" placeholder="SEO 描述内容" /></el-form-item></div></div></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="saveCategory">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="confirmVisible" title="确认删除" width="400px" :close-on-click-modal="false" class="confirm-dialog">
      <div class="confirm-body"><div class="confirm-icon">⚠️</div><div class="confirm-title">{{ confirmTitle }}</div><div class="confirm-msg">{{ confirmMsg }}</div></div>
      <template #footer><el-button @click="confirmVisible = false">取消</el-button><el-button type="danger" @click="doDelete">确认删除</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, ArrowDown } from '@element-plus/icons-vue'
import { listCategoryTree, createCategory as createCategoryApi, updateCategory as updateCategoryApi, deleteCategory as deleteCategoryApi, updateCategorySort, updateCategoryStatus, toggleCategoryShowcase, batchCategoryOperation } from '../../api/category'

interface CategoryNode { id: number; name: string; slug: string; icon: string; sort: number; status: boolean; showcase: boolean; count: number; image?: string; seoTitle?: string; seoKeywords?: string; seoDesc?: string; parentId?: number; children?: CategoryNode[]; _depth?: number }

const treeData = ref<CategoryNode[]>([])
const catMap: Record<number, CategoryNode> = {}
const treeSearch = ref('')
const selectedNodeId = ref<number | null>(null)
const expandedState = reactive<Record<number, boolean>>({})
const dialogVisible = ref(false); const confirmVisible = ref(false)
const isEditing = ref(false); const editingId = ref<number | null>(null)
const seoOpen = ref(false); const formRef = ref(null); const fileInputRef = ref(null)
const deletingId = ref<number | null>(null); const confirmTitle = ref(''); const confirmMsg = ref('')
const formData = reactive({ id: null as number | null, name: '', slug: '', parentId: null as number | null, sort: 0, icon: '', status: true, showcase: false, image: '', count: 0, seoTitle: '', seoKeywords: '', seoDesc: '' })

const emojiList = ['📱','💻','🎧','⌚','📷','👕','👗','👟','🎒','🍔','🍿','🥤','🥩','🍷','🏠','🍳','🌀','🛋','🛏','📚','📖','✏️','⚽','🎮','🎵','💄','👶','🐾','🌿','🎁','📦','🔧','💊','🚗','✈️','🏕','🔌','🖨','🎤','🎸','🏆','🥇','🧸','🎈','🎀','💎','⌨️','🖥','📡','🕹','📀','💿','📷','🎥','📺','🔋','🛒','🧾','🏷','💳','📊','📋','📌','📍','✂️','🧵','👠','👒','🧤','🧣','🌂','🧴','🪥','🪒','🧹','🪣','🧯','🪜','🛠','🧰','⚙️','🔩','🔗','🧲']

const treeNodes = computed(() => { const f = treeSearch.value.trim().toLowerCase(); const r: (CategoryNode & { _depth: number; hasChildren: boolean })[] = []; flattenNodes(treeData.value, f, 0, r); return r })
const totalCategoryCount = computed(() => countAll(treeData.value))
const topLevelCount = computed(() => treeData.value.length)
const subLevelCount = computed(() => { let c = 0; for (const cat of treeData.value) { if (cat.children) c += cat.children.length } return c })
const breadcrumbs = computed(() => { const r: { id: number; name: string; icon: string }[] = []; let cur = selectedNodeId.value ? getNodeById(selectedNodeId.value) : null; while (cur) { r.unshift({ id: cur.id, name: cur.name, icon: cur.icon || '📁' }); cur = getParentNode(cur.id) } return r })
const currentChildren = computed(() => { if (!selectedNodeId.value) return treeData.value; const n = getNodeById(selectedNodeId.value); return n?.children || [] })
const enabledCount = computed(() => currentChildren.value.filter(c => c.status).length)
const disabledCount = computed(() => currentChildren.value.filter(c => !c.status).length)
const parentOptions = computed(() => { const r: { id: number; label: string }[] = []; flattenOptions(treeData.value, r, '', editingId.value); return r })
const formRules = { name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }, { min: 1, max: 20, message: '长度在 1 到 20 个字符', trigger: 'blur' }], slug: [{ required: true, message: '请输入别名', trigger: 'blur' }] }

async function loadTree() { try { const res = await listCategoryTree(); treeData.value = (res.items || []).map(n => { const node: CategoryNode = { id: n.id, name: n.name || '', slug: n.slug || '', icon: n.icon || '📁', sort: n.sort ?? 0, status: n.status ?? true, showcase: n.showcase ?? false, count: n.count ?? 0, image: n.image || '', seoTitle: n.seo_title || '', seoKeywords: n.seo_keywords || '', seoDesc: n.seo_desc || '', parentId: n.parent_id }; if (n.children && Array.isArray(n.children)) node.children = n.children.map(((child:any) => { const cnode: CategoryNode = { id: child.id, name: child.name || '', slug: child.slug || '', icon: child.icon || '📁', sort: child.sort ?? 0, status: child.status ?? true, showcase: child.showcase ?? false, count: child.count ?? 0, image: child.image || '', seoTitle: child.seo_title || '', seoKeywords: child.seo_keywords || '', seoDesc: child.seo_desc || '', parentId: child.parent_id }; if (child.children && Array.isArray(child.children)) cnode.children = child.children; return cnode })); return node }); buildCatMap(treeData.value); treeData.value.forEach(c => { expandedState[c.id] = true }); if (treeData.value.length > 0 && !selectedNodeId.value) selectedNodeId.value = treeData.value[0].id } catch (e) { console.error('加载分类树失败:', e) } }

function flattenNodes(arr: CategoryNode[], filter: string, depth: number, result: (CategoryNode & { _depth: number; hasChildren: boolean })[]) { for (const node of arr) { const mn = !filter || node.name.toLowerCase().includes(filter); const cm = filter && node.children ? ((n:CategoryNode[]):boolean=>{for(const x of n){if(x.name.toLowerCase().includes(filter))return true;if(x.children&&hasMatchInTree(x.children,filter))return true}return false})(node.children) : false; if (filter && !mn && !cm) continue; const ie = filter ? true : (expandedState[node.id] !== false); result.push({ ...node, _depth: depth, hasChildren: !!(node.children && node.children.length > 0) }); if (node.children && ie) flattenNodes(node.children, filter, depth + 1, result) } }
function hasMatchInTree(arr: CategoryNode[], filter: string): boolean { for (const n of arr) { if (n.name.toLowerCase().includes(filter)) return true; if (n.children && hasMatchInTree(n.children, filter)) return true } return false }
function countAll(arr: CategoryNode[]): number { let c = 0; for (const n of arr) { c++; if (n.children) c += countAll(n.children) } return c }
function buildCatMap(arr: CategoryNode[]) { for (const n of arr) { catMap[n.id] = n; if (n.children) buildCatMap(n.children) } }
function getNodeById(id: number, arr: CategoryNode[] = treeData.value): CategoryNode | null { for (const n of arr) { if (n.id === id) return n; if (n.children) { const f = getNodeById(id, n.children); if (f) return f } } return null }
function getParentNode(id: number, arr: CategoryNode[] = treeData.value, parent: CategoryNode | null = null): CategoryNode | null { for (const n of arr) { if (n.id === id) return parent; if (n.children) { const f = getParentNode(id, n.children, n); if (f) return f } } return null }
function flattenOptions(arr: CategoryNode[], result: { id: number; label: string }[], prefix: string, skipId: number | null) { for (const n of arr) { if (n.id === skipId) continue; result.push({ id: n.id, label: prefix + n.name }); if (n.children) flattenOptions(n.children, result, prefix + '\u00A0\u00A0\u00A0\u00A0', skipId) } }
function toggleTreeNode(id: number) { expandedState[id] = expandedState[id] === false ? true : false }
function expandAll() { for (const key in catMap) { const n = catMap[key]; if (n.children && n.children.length > 0) expandedState[n.id] = true } }
function collapseAll() { for (const key in catMap) expandedState[catMap[key].id] = false }
function selectTreeNode(id: number) { selectedNodeId.value = id }
async function updateSort(item: CategoryNode, val: string) { const s = parseInt(val) || 0; item.sort = s; try { await updateCategorySort(item.id, s); ElMessage.success('排序已更新') } catch (e) { console.error(e) } }
async function toggleStatus(item: CategoryNode) { const ns = !item.status; item.status = ns; try { await updateCategoryStatus(item.id, ns); ElMessage.success(`${item.name} 已${ns ? '启用' : '禁用'}`) } catch (e) { item.status = !ns } }
async function toggleShowcase(item: CategoryNode) { const nv = !item.showcase; item.showcase = nv; try { await toggleCategoryShowcase(item.id, nv); ElMessage.success(`${item.name} 已${nv ? '设为首页展示' : '取消首页展示'}`) } catch (e) { item.showcase = !nv } }
async function handleBatchCommand(cmd: string) { const ids = currentChildren.value.map(c => c.id); if (ids.length === 0) { ElMessage.warning('当前没有可操作的分类'); return }; const m: Record<string, string> = { enable: '启用', disable: '禁用', delete: '删除' }; const action = m[cmd]; if (!action) return; try { await batchCategoryOperation(action, ids); ElMessage.success(`已批量${action} ${ids.length} 个分类`); await loadTree() } catch (e) { console.error(e) } }
function openAddModal() { resetForm(); isEditing.value = false; editingId.value = null; formData.parentId = selectedNodeId.value; dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function openAddSubModal(item: CategoryNode) { resetForm(); isEditing.value = false; editingId.value = null; formData.parentId = item.id; dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function openEditModal(item: CategoryNode) { isEditing.value = true; editingId.value = item.id; Object.assign(formData, { id: item.id, name: item.name, slug: item.slug, parentId: item.parentId ?? getParentNode(item.id)?.id ?? null, sort: item.sort, icon: item.icon, status: item.status, showcase: item.showcase, image: item.image || '', count: item.count, seoTitle: item.seoTitle || '', seoKeywords: item.seoKeywords || '', seoDesc: item.seoDesc || '' }); dialogVisible.value = true; nextTick(() => seoOpen.value = false) }
function resetForm() { Object.assign(formData, { id: null, name: '', slug: '', parentId: null, sort: 0, icon: '📁', status: true, showcase: false, image: '', count: 0, seoTitle: '', seoKeywords: '', seoDesc: '' }) }
async function saveCategory() { if (!formData.name.trim()) { ElMessage.warning('请输入分类名称'); return }; const payload = { name: formData.name.trim(), slug: formData.slug.trim() || formData.name.toLowerCase().replace(/\s+/g, '-'), icon: formData.icon || '📁', sort: formData.sort, status: formData.status, showcase: formData.showcase, image: formData.image, seo_title: formData.seoTitle, seo_keywords: formData.seoKeywords, seo_desc: formData.seoDesc, parent_id: formData.parentId || 0 }; try { if (isEditing.value && editingId.value) { await updateCategoryApi(editingId.value, payload); ElMessage.success(`分类「${formData.name}」已更新`) } else { await createCategoryApi(payload); ElMessage.success(`分类「${formData.name}」已创建`) }; dialogVisible.value = false; await loadTree() } catch (e) { console.error(e) } }
function confirmDelete(item: CategoryNode) { if (item.children && item.children.length > 0) { ElMessage.warning('该分类下有子分类，请先删除子分类'); return }; deletingId.value = item.id; confirmTitle.value = `确认删除「${item.name}」？`; confirmMsg.value = `此操作不可恢复，确定要删除分类「${item.name}」吗？`; confirmVisible.value = true }
async function doDelete() { if (deletingId.value === null) return; try { await deleteCategoryApi(deletingId.value); ElMessage.success('分类已删除'); confirmVisible.value = false; deletingId.value = null; await loadTree() } catch (e) { console.error(e); confirmVisible.value = false } }
function triggerUpload() { fileInputRef.value?.click() }
function handleUpload(event: Event) { const target = event.target as HTMLInputElement; const file = target.files?.[0]; if (!file) return; if (!file.type.startsWith('image/')) { ElMessage.warning('请选择图片文件'); return }; if (file.size > 2 * 1024 * 1024) { ElMessage.warning('图片大小不能超过 2MB'); return }; const reader = new FileReader(); reader.onload = (e) => { formData.image = e.target!.result as string }; reader.readAsDataURL(file); target.value = '' }
onMounted(() => { loadTree() })
</script>

<style scoped>
.categories-page {
  width: 100%;
  height: calc(100vh - 180px);
  display: flex;
  flex-direction: column;
}
.categories-layout {
  display: flex;
  gap: 1.25rem;
  align-items: stretch;
  flex: 1;
  min-height: 0;
}
.tree-panel {
  width: 300px;
  min-width: 300px;
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(12px);
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.7);
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.04);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.tree-panel-header {
  padding: 1.125rem 1.25rem;
  border-bottom: 1px solid rgba(0,0,0,0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.tree-panel-title {
  font-size: 0.9375rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #0f172a;
}
.tree-panel-count {
  font-size: 0.6875rem;
  color: #6366f1;
  background: rgba(99,102,241,0.08);
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  font-weight: 600;
}
.tree-ctrl-btns { display: flex; gap: 0.25rem; }
.tree-ctrl-btn {
  width: 1.5rem;
  height: 1.5rem;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  transition: all 0.15s ease;
}
.tree-ctrl-btn:hover { background: rgba(0,0,0,0.04); color: #475569; }
.tree-search-wrap { padding: 0.75rem 1rem; border-bottom: 1px solid rgba(0,0,0,0.06); }
.tree-search-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #f1f5f9;
  border-radius: 10px;
  padding: 0.5rem 0.75rem;
  transition: all 0.2s ease;
}
.tree-search-box:focus-within { background: #fff; box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
.tree-search-box input { border: none; outline: none; background: transparent; font-size: 0.8125rem; color: #0f172a; width: 100%; }
.tree-search-box input::placeholder { color: #94a3b8; }
.tree-search-box svg { width: 0.9375rem; height: 0.9375rem; color: #94a3b8; flex-shrink: 0; }
.tree-panel-body { flex: 1; overflow-y: auto; padding: 0.375rem 0; }
.tree-panel-body::-webkit-scrollbar { width: 3px; }
.tree-panel-body::-webkit-scrollbar-track { background: transparent; }
.tree-panel-body::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }
.tree-empty { padding: 2rem 1.25rem; text-align: center; color: #94a3b8; font-size: 0.8125rem; }
.tree-panel-footer {
  padding: 0.625rem 1.25rem;
  border-top: 1px solid rgba(0,0,0,0.06);
  font-size: 0.75rem;
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.total-label { font-weight: 500; }
.total-num { font-weight: 700; color: #0f172a; }
.tree-footer-detail { font-size: 0.6875rem; }
.tree-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  transition: all 0.12s ease;
  user-select: none;
  font-size: 0.8125rem;
  position: relative;
  margin: 0 0.375rem;
  border-radius: 8px;
}
.tree-item:hover { background: rgba(0,0,0,0.03); }
.tree-item.active { background: rgba(99,102,241,0.08); color: #4f46e5; font-weight: 600; }
.tree-toggle {
  width: 1rem;
  height: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #94a3b8;
  transition: transform 0.2s ease;
  cursor: pointer;
  border-radius: 4px;
}
.tree-toggle:hover { background: rgba(0,0,0,0.04); }
.tree-toggle.expanded { transform: rotate(90deg); }
.tree-toggle-placeholder { width: 1rem; height: 1rem; flex-shrink: 0; }
.tree-icon { font-size: 1rem; flex-shrink: 0; line-height: 1; }
.tree-label { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tree-count {
  font-size: 0.625rem;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 0.0625rem 0.375rem;
  border-radius: 999px;
  font-weight: 600;
  flex-shrink: 0;
}
.tree-item.active .tree-count { background: rgba(99,102,241,0.12); color: #6366f1; }
.content-panel { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1rem; }
.breadcrumb-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.7);
  border-radius: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.03);
  font-size: 0.8125rem;
  flex-wrap: wrap;
}
.bc-sep { color: #cbd5e1; font-size: 0.75rem; margin: 0 0.125rem; }
.bc-link { color: #64748b; cursor: pointer; transition: color 0.15s ease; font-weight: 500; }
.bc-link:hover { color: #4f46e5; }
.bc-current { color: #0f172a; font-weight: 700; }
.table-card {
  background: rgba(255,255,255,0.88);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.7);
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.04);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}
.table-toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(0,0,0,0.06);
  flex-wrap: wrap;
}
.toolbar-spacer { flex: 1; }
.stats-label {
  font-size: 0.8125rem;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 1rem;
}
.stats-label .stat-num { font-weight: 700; color: #0f172a; }
.stat-dot { display: inline-block; width: 6px; height: 6px; border-radius: 999px; margin-right: 0.25rem; }
.stat-dot.green { background: #22c55e; }
.stat-dot.red { background: #ef4444; }
.btn-secondary.btn-sm {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: #fff;
  color: #475569;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  height: auto;
}
.btn-secondary.btn-sm:hover { border-color: #cbd5e1; background: #f8fafc; }
.table-wrap { overflow: auto; flex: 1; min-height: 0; }
.sub-table { width: 100%; border-collapse: collapse; }
.sub-table thead th {
  text-align: left;
  padding: 0.75rem 1.25rem;
  font-size: 0.6875rem;
  font-weight: 600;
  color: #64748b;
  letter-spacing: 0.04em;
  background: #f8fafc;
  border-bottom: 1px solid #f1f5f9;
  white-space: nowrap;
}
.sub-table tbody tr { transition: background 0.12s ease; }
.sub-table tbody tr:hover { background: #f8fafc; }
.sub-table tbody td {
  padding: 0.75rem 1.25rem;
  font-size: 0.8125rem;
  color: #0f172a;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}
.sub-table tbody tr:last-child td { border-bottom: none; }
.table-empty { text-align: center; padding: 2.5rem 1rem; }
.table-empty .empty-icon { font-size: 2.5rem; margin-bottom: 0.5rem; }
.table-empty .empty-text { font-size: 0.9375rem; font-weight: 600; color: #64748b; margin-bottom: 0.25rem; }
.table-empty .empty-hint { font-size: 0.75rem; color: #94a3b8; }
.cat-name { display: flex; align-items: center; gap: 0.5rem; font-weight: 500; }
.cat-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 8px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  flex-shrink: 0;
}
.child-badge {
  font-size: 0.625rem;
  font-weight: 600;
  color: #6366f1;
  background: rgba(99,102,241,0.08);
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  cursor: pointer;
  white-space: nowrap;
}
.child-badge:hover { background: rgba(99,102,241,0.15); }
.slug-text { font-size: 0.75rem; color: #94a3b8; background: #f8fafc; padding: 0.125rem 0.5rem; border-radius: 4px; }
.sort-input {
  width: 3rem;
  padding: 0.25rem 0.375rem;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  text-align: center;
  background: #fff;
  color: #0f172a;
  transition: all 0.2s ease;
  outline: none;
  -moz-appearance: textfield;
}
.sort-input::-webkit-inner-spin-button, .sort-input::-webkit-outer-spin-button { -webkit-appearance: none; margin: 0; }
.sort-input:hover { border-color: #94a3b8; }
.sort-input:focus { border-color: #6366f1; box-shadow: 0 0 0 2px rgba(99,102,241,0.15); }
.toggle-switch {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}
.toggle-track {
  width: 2.125rem;
  height: 1.25rem;
  border-radius: 999px;
  background: #cbd5e1;
  transition: all 0.25s ease;
  position: relative;
  flex-shrink: 0;
}
.toggle-track.active { background: #6366f1; }
.toggle-thumb {
  position: absolute;
  top: 2px; left: 2px;
  width: 1rem;
  height: 1rem;
  border-radius: 999px;
  background: #fff;
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}
.toggle-track.active .toggle-thumb { transform: translateX(0.875rem); }
.toggle-label { font-size: 0.6875rem; color: #94a3b8; font-weight: 500; }
.toggle-label.active { color: #6366f1; }
.showcase-star {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  border: none;
  background: transparent;
  border-radius: 999px;
  cursor: pointer;
  color: #94a3b8;
  font-size: 1rem;
  transition: all 0.2s ease;
  padding: 0;
  line-height: 1;
}
.showcase-star:hover { background: #f1f5f9; transform: scale(1.15); }
.showcase-star.active { color: #f59e0b; }
.showcase-star.active:hover { background: #fffbeb; }
.action-links { display: flex; align-items: center; gap: 0.125rem; white-space: nowrap; }
.action-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border: none;
  background: transparent;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.action-link.edit-link { color: #6366f1; }
.action-link.edit-link:hover { background: rgba(99,102,241,0.08); }
.action-link.add-link { color: #059669; }
.action-link.add-link:hover { background: rgba(5,150,105,0.08); }
.action-link.delete-link { color: #ef4444; }
.action-link.delete-link:hover { background: #fef2f2; }
.action-divider { width: 1px; height: 1rem; background: #e2e8f0; margin: 0 0.125rem; }
.table-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid rgba(0,0,0,0.06);
  font-size: 0.75rem;
  color: #94a3b8;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.pagination-info { display: flex; align-items: center; gap: 0.5rem; }
.pagination-info .num { font-weight: 700; color: #0f172a; }
.emoji-picker-label { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.5rem; }
.selected-emoji {
  font-size: 1.375rem;
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
}
.emoji-grid { display: grid; grid-template-columns: repeat(8, 1fr); gap: 4px; }
.emoji-opt {
  width: 100%;
  aspect-ratio: 1;
  border: 1px solid #f1f5f9;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  transition: all 0.15s ease;
  padding: 0;
  line-height: 1;
}
.emoji-opt:hover { border-color: #6366f1; background: rgba(99,102,241,0.06); transform: scale(1.15); }
.emoji-opt.selected { border-color: #6366f1; background: rgba(99,102,241,0.08); box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
.upload-area {
  border: 2px dashed #e2e8f0;
  border-radius: 12px;
  padding: 1rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f8fafc;
  max-width: 240px;
}
.upload-area:hover { border-color: #6366f1; background: rgba(99,102,241,0.04); }
.upload-area .upload-icon { font-size: 1.5rem; color: #94a3b8; margin-bottom: 0.375rem; }
.upload-area .upload-text { font-size: 0.75rem; color: #64748b; }
.upload-area .upload-hint { font-size: 0.625rem; color: #94a3b8; margin-top: 0.25rem; }
.upload-preview { max-width: 100%; max-height: 120px; border-radius: 8px; }
.form-row { display: flex; gap: 1rem; }
.form-hint { font-size: 0.6875rem; color: #94a3b8; margin-top: 0.25rem; display: block; }
.seo-section { border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; width: 100%; }
.seo-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: none;
  background: #f8fafc;
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #0f172a;
  transition: background 0.15s ease;
}
.seo-toggle:hover { background: #f1f5f9; }
.seo-arrow { transition: transform 0.2s ease; font-size: 0.75rem; }
.seo-arrow.open { transform: rotate(180deg); }
.seo-body { padding: 1rem; border-top: 1px solid #e2e8f0; }
.confirm-body { text-align: center; padding: 0.5rem 0; }
.confirm-icon { font-size: 2.5rem; margin-bottom: 0.75rem; }
.confirm-title { font-size: 1rem; font-weight: 700; color: #0f172a; margin-bottom: 0.375rem; }
.confirm-msg { font-size: 0.8125rem; color: #64748b; line-height: 1.5; }
.category-dialog :deep(.el-dialog__body) { padding-top: 1.25rem; }
@media (max-width: 1024px) {
  .categories-layout { flex-direction: column; }
  .tree-panel { width: 100%; min-width: 0; max-height: 360px; }
  .content-panel { width: 100%; }
}
@media (max-width: 768px) {
  .form-row { flex-direction: column; gap: 0; }
  .emoji-grid { grid-template-columns: repeat(6, 1fr); }
  .table-toolbar .stats-label { display: none; }
  .sub-table thead th, .sub-table tbody td { padding: 0.5rem 0.75rem; font-size: 0.75rem; }
}
@media (max-width: 480px) {
  .tree-panel-header .tree-ctrl-btns { display: none; }
  .table-toolbar { flex-direction: column; align-items: stretch; }
  .table-toolbar .toolbar-spacer { display: none; }
  .emoji-grid { grid-template-columns: repeat(5, 1fr); }
}
</style>