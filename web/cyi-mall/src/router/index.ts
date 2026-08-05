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
      { path: 'checkout', name: 'Checkout', component: () => import('../views/checkout/CheckoutView.vue'), meta: { title: '确认订单', requiresAuth: true } },
      { path: 'user', name: 'User', component: () => import('../views/user/UserCenterView.vue'), meta: { title: '个人中心', requiresAuth: true } },
      { path: 'user/edit', name: 'ProfileEdit', component: () => import('../views/user/ProfileEditView.vue'), meta: { title: '编辑资料', requiresAuth: true } },
      { path: 'user/orders', name: 'OrderList', component: () => import('../views/user/OrderListView.vue'), meta: { title: '我的订单', requiresAuth: true } },
      { path: 'user/orders/:id', name: 'OrderDetail', component: () => import('../views/user/OrderDetailView.vue'), meta: { title: '订单详情', requiresAuth: true } },
      { path: 'user/address', name: 'AddressList', component: () => import('../views/user/AddressView.vue'), meta: { title: '收货地址', requiresAuth: true } },
      { path: 'user/coupons', name: 'MyCoupons', component: () => import('../views/coupons/MyCouponView.vue'), meta: { title: '我的优惠券', requiresAuth: true } },
      { path: 'coupons', name: 'CouponCenter', component: () => import('../views/coupons/CouponCenterView.vue'), meta: { title: '领券中心', requiresAuth: true } }
    ]
  },
  { path: '/login', name: 'Login', component: () => import('../views/login/LoginView.vue'), meta: { title: '登录' } },
  { path: '/register', name: 'Register', component: () => import('../views/register/RegisterView.vue'), meta: { title: '注册' } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 未登录访问需登录页面时跳转登录页
router.beforeEach((to, _from, next) => {
  console.log('[route]', to.fullPath, 'requiresAuth:', !!to.meta.requiresAuth, 'token:', !!localStorage.getItem('token'))

  document.title = to.meta.title ? `${to.meta.title} - cyi商城` : 'cyi商城'

  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }
  }

  next()
})

export default router
