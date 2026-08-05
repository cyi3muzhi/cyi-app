# 悦选商城管理后台

基于 Vue 3 + Element Plus 的管理后台前端项目，为管理后台提供统一的用户界面。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Vue 3 + Vite + TypeScript |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| 路由 | Vue Router |
| HTTP 客户端 | Axios |
| 图标 | Element Plus Icons |

## 目录结构

```
web/admin/
├── public/                # 静态资源
│   ├── favicon.svg        # 网站图标
│   └── icons.svg          # 自定义图标
├── src/
│   ├── api/               # API 请求封装
│   │   ├── auth.ts        # 认证 API
│   │   ├── category.ts    # 商品分类 API
│   │   ├── index.ts       # 全局请求拦截器
│   │   ├── logs.ts        # 操作日志 API
│   │   ├── mock.js        # Mock 数据
│   │   ├── role.ts        # 角色权限 API
│   │   └── staff.ts       # 员工管理 API
│   ├── assets/            # 静态资源
│   │   ├── styles/        # 全局样式
│   │   │   ├── global.css # 全局样式
│   │   │   └── variables.css # 样式变量
│   │   ├── hero.png       # 首页图片
│   │   ├── vue.svg        # Vue 图标
│   │   └── vite.svg       # Vite 图标
│   ├── components/        # 公共组件
│   │   ├── AppSidebar.vue # 侧边栏导航
│   │   ├── AppTopbar.vue  # 顶部导航栏
│   │   └── HelloWorld.vue # 示例组件
│   ├── layouts/           # 布局组件
│   │   └── AdminLayout.vue # 管理后台布局
│   ├── router/            # 路由配置
│   │   └── index.ts       # 路由定义与守卫
│   ├── store/             # 状态管理
│   │   └── app.ts         # 全局状态（登录状态、侧边栏、权限等）
│   ├── types/             # TypeScript 类型定义
│   │   └── api.d.ts       # API 接口类型
│   ├── views/             # 页面视图
│   │   ├── categories/    # 商品分类管理
│   │   ├── coupons/       # 优惠券管理
│   │   ├── dashboard/     # 仪表盘首页
│   │   ├── login/         # 登录页面
│   │   ├── logs/          # 操作日志
│   │   ├── orders/        # 订单管理
│   │   ├── products/      # 商品管理（列表、表单）
│   │   ├── profile/       # 个人信息（头像、基本信息）
│   │   ├── roles/         # 角色管理
│   │   ├── settings/      # 系统设置
│   │   ├── staff/         # 员工管理
│   │   └── users/         # 用户管理
│   ├── App.vue            # 根组件
│   ├── main.ts            # 入口文件
│   ├── style.css          # Vite 默认样式
│   └── env.d.ts           # 环境类型声明
├── index.html             # HTML 模板
├── package.json           # 依赖配置
├── tsconfig.json          # TypeScript 配置
├── vite.config.ts         # Vite 配置
└── README.md              # 项目文档
```

## 页面列表

| 页面 | 路径 | 说明 |
|------|------|------|
| 登录页 | `/login` | 管理员登录 |
| 仪表盘 | `/dashboard` | 数据统计概览 |
| 商品管理 | `/products` | 商品列表 |
| 添加商品 | `/products/add` | 创建商品 |
| 编辑商品 | `/products/edit/:id` | 编辑商品 |
| 订单管理 | `/orders` | 订单列表 |
| 用户管理 | `/users` | 用户列表 |
| 商品分类 | `/categories` | 分类管理 |
| 优惠券管理 | `/coupons` | 优惠券管理 |
| 操作日志 | `/logs` | 操作日志查看 |
| 角色管理 | `/roles` | 角色与权限配置 |
| 员工管理 | `/staff` | 后台账号管理 |
| 个人信息 | `/profile` | 头像、基本信息设置 |
| 系统设置 | `/settings` | 系统配置 |

## API 接口

前端通过 Vite 代理访问 admin BFF 服务：

| API 路径 | 后端服务 | 说明 |
|----------|----------|------|
| `/admin/v1/auth/*` | Admin BFF | 认证接口 |
| `/admin/v1/dashboard/*` | Admin BFF | 仪表盘接口 |
| `/admin/v1/product/*` | Admin BFF | 商品接口 |
| `/admin/v1/categories/*` | Admin BFF | 商品分类接口 |
| `/admin/v1/order/*` | Admin BFF | 订单接口 |
| `/admin/v1/user/*` | Admin BFF | 用户接口 |
| `/admin/v1/roles/*` | Admin BFF | 角色管理接口 |
| `/admin/v1/staff/*` | Admin BFF | 员工管理接口 |
| `/admin/v1/permissions/*` | Admin BFF | 权限模块接口 |
| `/admin/v1/settings/*` | Admin BFF | 系统设置接口 |
| `/admin/v1/logs` | Admin BFF | 操作日志接口（列表、统计、操作人） |

### 认证接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/v1/auth/login` | POST | 管理员登录 |
| `/admin/v1/auth/logout` | POST | 管理员登出 |
| `/admin/v1/auth/info` | GET | 获取当前管理员信息 |
| `/admin/v1/auth/info` | PUT | 更新管理员基本信息 |
| `/admin/v1/auth/avatar` | POST | 更换头像（Base64） |
| `/admin/v1/auth/my-permissions` | GET | 获取当前用户权限 |
| `/admin/v1/auth/online-ids` | GET | 获取在线管理员 ID 列表 |

### 商品分类接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/v1/categories/tree` | GET | 获取分类树 |
| `/admin/v1/categories` | GET | 查询子分类列表 |
| `/admin/v1/categories/:id` | GET | 获取单个分类 |
| `/admin/v1/categories` | POST | 创建分类 |
| `/admin/v1/categories/:id` | PUT | 更新分类 |
| `/admin/v1/categories/:id` | DELETE | 删除分类 |
| `/admin/v1/categories/:id/sort` | PUT | 更新分类排序 |
| `/admin/v1/categories/:id/status` | PUT | 更新分类状态 |
| `/admin/v1/categories/:id/showcase` | PUT | 切换分类首页展示 |
| `/admin/v1/categories/batch` | POST | 批量操作分类 |

### 角色管理接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/v1/roles` | GET | 获取角色列表 |
| `/admin/v1/roles` | POST | 创建角色 |
| `/admin/v1/roles/:slug` | PUT | 更新角色 |
| `/admin/v1/roles/:slug` | DELETE | 删除角色 |
| `/admin/v1/roles/:slug/permissions` | GET | 获取角色权限 |
| `/admin/v1/roles/:slug/permissions` | PUT | 更新角色权限 |
| `/admin/v1/permissions/modules` | GET | 获取所有权限模块定义 |
| `/admin/v1/role-name-map` | GET | 获取角色名称映射 |

### 员工管理接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/v1/staff` | GET | 获取员工列表 |
| `/admin/v1/staff` | POST | 创建员工 |
| `/admin/v1/staff/:id` | PUT | 更新员工信息 |
| `/admin/v1/staff/:id/toggle-status` | POST | 切换员工状态 |

### 请求拦截器

- **请求拦截器**：自动添加 `Authorization: Bearer {token}` 请求头
- **响应拦截器**：统一错误处理，401 自动跳转登录页（防重复处理），403 提示无权限，503 提示服务暂不可用

## 运行方式

### 环境要求

- Node.js 18+
- pnpm（推荐）

### 本地开发

```bash
cd web/admin
pnpm install
pnpm run dev
```

开发服务器运行在 `http://localhost:3000`，通过 Vite 代理访问后端服务。

### 构建部署

```bash
cd web/admin
npm run build
```

构建产物输出到 `dist/` 目录。

### 代理配置

Vite 代理配置（`vite.config.ts`）：

| 路径 | 目标地址 | 说明 |
|------|----------|------|
| `/admin/v1` | `http://localhost:28000` | admin BFF 服务 |

## 状态管理

使用 Pinia 管理全局状态：

| Store | 职责 |
|-------|------|
| `app` | 登录状态、管理员信息、侧边栏状态、用户权限 |

### 核心状态

```typescript
const store = useAppStore()

// 登录
store.login({ token, name, avatar, role, email, username })

// 登出
store.logout()

// 检查登录状态
store.isLoggedIn

// 侧边栏折叠
store.sidebarCollapsed
store.toggleSidebar()

// 权限检查
store.setPermissions(perms)
store.hasPerm(module, action)
```

## 路由守卫

路由守卫自动检查登录状态：

- 未登录访问受保护页面 → 跳转 `/login`
- 已登录访问 `/login` → 跳转 `/dashboard`
- 自动设置页面标题（`{title} - cyi商城管理后台`）

## 响应式设计

支持多种屏幕尺寸：
- 小屏幕（平板）：自适应布局，统计卡片 2 列显示
- 大屏（4K）：内容延展显示

## 开发规范

### 组件命名

- 页面组件：`{Name}View.vue`（如 `DashboardView.vue`）
- 公共组件：`App{Name}.vue`（如 `AppSidebar.vue`）

### API 调用

```typescript
import request from '../api/index'

// GET 请求
const data = await request.get('/path', { params: { id: 1 } })

// POST 请求
const data = await request.post('/path', { name: 'xxx' })
```

### 状态管理

```typescript
import { useAppStore } from '../store/app'

const store = useAppStore()
```

### 权限控制

```typescript
import { useAppStore } from '../store/app'

const store = useAppStore()

// 在模板中控制元素显隐
// v-if="store.hasPerm('staff', 'add')"

// 在脚本中检查权限
if (store.hasPerm('product', 'edit')) {
  // 允许编辑
}
```
