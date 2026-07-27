/**
 * cyi-mall 路由配置
 */
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('../layouts/DefaultLayout.vue'),
    children: [
      { path: '', name: 'Home', component: () => import('../views/home/HomeView.vue'), meta: { title: '首页' } },
      { path: 'products', name: 'ProductList', component: () => import('../views/products/ProductListView.vue'), meta: { title: '全部商品' } },
      { path: 'product/:id', name: 'ProductDetail', component: () => import('../views/product/ProductDetailView.vue'), meta: { title: '商品详情' } },
      { path: 'cart', name: 'Cart', component: () => import('../views/cart/CartView.vue'), meta: { title: '购物车' } },
      { path: 'user', name: 'User', component: () => import('../views/user/UserCenterView.vue'), meta: { title: '个人中心' } },
      { path: 'user/orders', name: 'OrderList', component: () => import('../views/user/OrderListView.vue'), meta: { title: '我的订单' } }
    ]
  },
  { path: '/login', name: 'Login', component: () => import('../views/login/LoginView.vue'), meta: { title: '登录' } },
  { path: '/register', name: 'Register', component: () => import('../views/register/RegisterView.vue'), meta: { title: '注册' } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 设置标题
router.beforeEach((to, _from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 悦选商城` : '悦选商城'
  next()
})

export default router
