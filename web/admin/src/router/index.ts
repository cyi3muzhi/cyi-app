/**
 * 悦选商城管理后台 - 路由配置
 */
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAppStore } from '../store/app'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/LoginView.vue'),
    meta: { title: '管理员登录' }
  },
  {
    path: '/',
    component: () => import('../layouts/AdminLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/dashboard/DashboardView.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('../views/products/ProductListView.vue'),
        meta: { title: '商品管理' }
      },
      {
        path: 'products/add',
        name: 'ProductAdd',
        component: () => import('../views/products/ProductFormView.vue'),
        meta: { title: '添加商品' }
      },
      {
        path: 'products/edit/:id',
        name: 'ProductEdit',
        component: () => import('../views/products/ProductFormView.vue'),
        meta: { title: '编辑商品' }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('../views/orders/OrderListView.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('../views/users/UserListView.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('../views/categories/CategoriesView.vue'),
        meta: { title: '商品分类' }
      },
      {
        path: 'coupons',
        name: 'Coupons',
        component: () => import('../views/coupons/CouponsView.vue'),
        meta: { title: '优惠券管理' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../views/logs/LogsView.vue'),
        meta: { title: '操作日志' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('../views/roles/RolesView.vue'),
        meta: { title: '角色管理' }
      },
      {
        path: 'staff',
        name: 'Staff',
        component: () => import('../views/staff/StaffView.vue'),
        meta: { title: '员工管理' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/profile/ProfileView.vue'),
        meta: { title: '个人信息' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/settings/SettingsView.vue'),
        meta: { title: '系统设置' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 未登录跳转登录页
router.beforeEach((to, _from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 悦选商城管理后台` : '悦选商城管理后台'

  const store = useAppStore()
  if (to.name !== 'Login' && !store.isLoggedIn) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && store.isLoggedIn) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
