# 悦选商城管理后台

基于 Vue 3 + Element Plus 的管理后台前端项目，为管理后台提供统一的用户界面。

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | Vue 3 + Vite |
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
│   │   ├── auth.js        # 认证 API
│   │   ├── index.js       # 全局请求拦截器
│   │   └── mock.js        # Mock 数据
│   ├── assets/            # 静态资源
│   │   ├── styles/        # 全局样式
│   │   │   ├── global.css # 全局样式
│   │   │   └── variables.css # 样式变量
│   │   └── hero.png       # 首页图片
│   ├── components/        # 公共组件
│   │   ├── AppSidebar.vue # 侧边栏导航
│   │   ├── AppTopbar.vue  # 顶部导航栏
│   │   └── HelloWorld.vue # 示例组件
│   ├── layouts/           # 布局组件
│   │   └── AdminLayout.vue # 管理后台布局
│   ├── router/            # 路由配置
│   │   └── index.js       # 路由定义与守卫
│   ├── store/             # 状态管理
│   │   └── app.js         # 全局状态（登录状态、侧边栏等）
│   ├── views/             # 页面视图
│   │   ├── categories/    # 商品分类管理
│   │   ├── coupons/       # 优惠券管理
│   │   ├── dashboard/     # 仪表盘首页
│   │   ├── login/         # 登录页面
│   │   ├── logs/          # 操作日志
│   │   ├── orders/        # 订单管理
│   │   ├── products/      # 商品管理（列表、表单）
│   │   ├── profile/       # 个人信息（头像、基本信息）
│   │   ├── settings/      # 系统设置
│   │   └── users/         # 用户管理
│   ├── App.vue            # 根组件
│   ├── main.js            # 入口文件
│   └── style.css          # 全局样式入口
├── index.html             # HTML 模板
├── package.json           # 依赖配置
├── vite.config.js         # Vite 配置
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
| 个人信息 | `/profile` | 头像、基本信息设置 |
| 系统设置 | `/settings` | 系统配置 |

## API 接口

前端通过 Vite 代理访问 admin BFF 服务：

| API 路径 | 后端服务 | 说明 |
|----------|----------|------|
| `/admin/v1/auth/*` | Admin BFF | 认证接口 |
| `/admin/v1/dashboard/*` | Admin BFF | 仪表盘接口 |
| `/admin/v1/product/*` | Admin BFF | 商品接口 |

### 认证接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/v1/auth/login` | POST | 管理员登录 |
| `/admin/v1/auth/logout` | POST | 管理员登出 |
| `/admin/v1/auth/info` | GET | 获取当前管理员信息 |
| `/admin/v1/auth/avatar` | POST | 更换头像（Base64） |

### 请求拦截器

- **请求拦截器**：自动添加 `Authorization: Bearer {token}` 请求头
- **响应拦截器**：统一错误处理，401 自动跳转登录页

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

Vite 代理配置（`vite.config.js`）：

| 路径 | 目标地址 | 说明 |
|------|----------|------|
| `/admin` | `http://localhost:28000` | admin BFF 服务 |

## 状态管理

使用 Pinia 管理全局状态：

| Store | 职责 |
|-------|------|
| `app` | 登录状态、管理员信息、侧边栏状态 |

### 登录状态

```javascript
const store = useAppStore()

// 登录
store.login({ token, name, avatar, role })

// 登出
store.logout()

// 检查登录状态
store.isLoggedIn
```

## 路由守卫

路由守卫自动检查登录状态：

- 未登录访问受保护页面 → 跳转 `/login`
- 已登录访问 `/login` → 跳转 `/dashboard`

## 响应式设计

支持多种屏幕尺寸：
- 小屏幕（平板）：自适应布局
- 大屏（4K）：内容延展显示

## 开发规范

### 组件命名

- 页面组件：`{Name}View.vue`（如 `DashboardView.vue`）
- 公共组件：`App{Name}.vue`（如 `AppSidebar.vue`）

### API 调用

```javascript
import request from '../api/index'

// GET 请求
const data = await request.get('/path', { params: { id: 1 } })

// POST 请求
const data = await request.post('/path', { name: 'xxx' })
```

### 状态管理

```javascript
import { useAppStore } from '../store/app'

const store = useAppStore()
```
