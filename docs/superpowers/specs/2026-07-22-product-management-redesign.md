# 商品管理页面重设计

## 范围

对 `web/admin/` 的商品管理模块进行重实现，包括商品列表页和商品表单页。视觉风格沿用现有设计，重点提升交互一致性和代码质量。

## 页面结构

### 商品列表页 (`/products`)

```
Toolbar: [➕添加商品] [🔍搜索框] [📂分类下拉]
Bulk Bar: 已选 N 项 [批量上架] [批量下架] [批量删除]
Tabs: [全部] [上架] [下架] [缺货]
Table: 多选 | 图片 | 名称 | 分类 | 价格 | 库存 | 状态 | 操作(编辑/删除)
Pagination: 服务端分页, ← 1 2 3 ... N →
```

- 编辑 → 跳转 `/products/edit/:id`
- 删除 → Popconfirm 确认
- 分页已改为服务端分页（前次修复完成）

### 商品表单页 (`/products/add`, `/products/edit/:id`)

```
Breadcrumb: 商品管理 > 添加商品 / 编辑商品
Section 1 — 基本信息: 名称*, 简介, 分类*, 品牌, 标签, 缩略图, 商品图片
Section 2 — 规格与 SKU: 规格组构建器, SKU 矩阵(自动生成, 编辑价格/库存)
Section 3 — 商品描述: 富文本编辑器
Footer: [取消] [保存]
```

- 添加模式 → 空表单，保存后跳转列表
- 编辑模式 → 加载数据，保存后跳转列表
- 去掉 ProductListView.vue 内嵌的编辑弹窗

## 代码重构

| 文件 | 说明 |
|------|------|
| `src/types/product.ts` | 从 api.ts 拆出所有商品类型接口 |
| `src/composables/useProductList.ts` | 列表页逻辑 (加载/筛选/分页/批量) |
| `src/composables/useProductForm.ts` | 表单逻辑 (加载/校验/保存/规格构建) |
| `src/components/ImageUploader.vue` | 图片上传组件 |

## 接口

沿用现有 API 接口不变：
- `GET /admin/v1/products` — 列表
- `GET /admin/v1/products/:id` — 详情
- `POST /admin/v1/products` — 创建
- `PUT /admin/v1/products/:id` — 更新
- `DELETE /admin/v1/products/:id` — 删除
- `POST /admin/v1/products/batch` — 批量操作
