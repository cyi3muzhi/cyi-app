/**
 * cyi商城管理后台 - Mock 模拟数据
 * 后期对接 API 后替换为真实接口调用
 */

/** 仪表盘统计数据 */
export const dashboardStats = {
  todayOrders: { value: 128, unit: '单', growth: 12.5, growthUp: true },
  todayRevenue: { value: 45680, unit: '元', prefix: '¥', growth: 8.3, growthUp: true },
  totalUsers: { value: 23568, unit: '', growth: 5.2, growthUp: true },
  totalProducts: { value: 1286, unit: '', growth: 3.1, growthUp: true }
}

/** 近14天订单趋势数据 */
export const orderTrend = {
  days: ['6/30', '7/1', '7/2', '7/3', '7/4', '7/5', '7/6', '7/7', '7/8', '7/9', '7/10', '7/11', '7/12', '7/13'],
  values: [85, 92, 78, 110, 95, 120, 105, 98, 115, 130, 108, 125, 140, 128]
}

/** 营收分类占比数据 */
export const revenueCategories = [
  { name: '服装鞋帽', percent: 42, color: '#3b82f6' },
  { name: '数码电子', percent: 28, color: '#10b981' },
  { name: '家居生活', percent: 18, color: '#f59e0b' },
  { name: '其他品类', percent: 12, color: '#8b5cf6' }
]

/** 最新订单列表（5条） */
export const latestOrders = [
  { id: 'ORD-20260713-001', user: '王芳', amount: 1280.00, status: 'completed', statusText: '已完成', time: '2026-07-13 14:32' },
  { id: 'ORD-20260713-002', user: '李强', amount: 568.50, status: 'processing', statusText: '处理中', time: '2026-07-13 14:15' },
  { id: 'ORD-20260713-003', user: '赵丽', amount: 2450.00, status: 'processing', statusText: '处理中', time: '2026-07-13 13:58' },
  { id: 'ORD-20260713-004', user: '陈伟', amount: 399.00, status: 'completed', statusText: '已完成', time: '2026-07-13 12:40' },
  { id: 'ORD-20260713-005', user: '刘洋', amount: 1899.00, status: 'cancelled', statusText: '已取消', time: '2026-07-13 11:22' }
]

/** 商品分类列表 */
export const productCategories = ['电子产品', '家居生活', '时尚服饰', '运动户外', '美妆护肤']

/** 商品列表（10条） */
export const products = [
  { id: 1, name: '无线降噪蓝牙耳机 Pro', category: '电子产品', price: 299, stock: 256, sales: 2586, status: '上架中', img: '', sku: 'YX-EP001' },
  { id: 2, name: '北欧简约护眼台灯 LED', category: '家居生活', price: 159, stock: 480, sales: 4213, status: '上架中', img: '', sku: 'YX-LT002' },
  { id: 3, name: '都市休闲双肩背包 大容量', category: '时尚服饰', price: 189, stock: 0, sales: 1847, status: '缺货', img: '', sku: 'YX-BP003' },
  { id: 4, name: '轻量缓震运动跑步鞋', category: '运动户外', price: 259, stock: 320, sales: 3621, status: '上架中', img: '', sku: 'YX-SH004' },
  { id: 5, name: '氨基酸温和洁面乳 120g', category: '美妆护肤', price: 79, stock: 1200, sales: 8942, status: '上架中', img: '', sku: 'YX-CL005' },
  { id: 6, name: '智能运动手表 S3 全天监测', category: '电子产品', price: 459, stock: 15, sales: 5108, status: '上架中', img: '', sku: 'YX-WT006' },
  { id: 7, name: '便携无线榨汁杯 随身榨', category: '家居生活', price: 119, stock: 0, sales: 6374, status: '缺货', img: '', sku: 'YX-JC007' },
  { id: 8, name: '纯棉圆领短袖T恤 三色可选', category: '时尚服饰', price: 89, stock: 2300, sales: 12756, status: '上架中', img: '', sku: 'YX-TS008' },
  { id: 9, name: '运动蓝牙耳机 挂耳式', category: '电子产品', price: 199, stock: 86, sales: 3562, status: '已下架', img: '', sku: 'YX-SE009' },
  { id: 10, name: '玻尿酸保湿精华液 30ml', category: '美妆护肤', price: 169, stock: 450, sales: 2351, status: '上架中', img: '', sku: 'YX-ES010' }
]

/** 订单列表（10条） */
export const orders = [
  { id: 'YX20260701001', time: '2026-07-01 09:23:15', user: '张伟', status: 'completed', statusText: '已完成', amount: 459.00, payment: '微信支付', items: [{ name: '智能运动手表', spec: '深空灰', qty: 1, price: 459.00 }] },
  { id: 'YX20260702002', time: '2026-07-02 14:05:42', user: '李娜', status: 'pending_ship', statusText: '待发货', amount: 598.00, payment: '支付宝', items: [{ name: '无线降噪耳机', spec: '白色', qty: 1, price: 299.00 }, { name: '便携蓝牙音箱', spec: '黑色', qty: 1, price: 299.00 }] },
  { id: 'YX20260703003', time: '2026-07-03 10:11:33', user: '王强', status: 'pending_payment', statusText: '待付款', amount: 189.00, payment: '微信支付', items: [{ name: '简约双肩背包', spec: '灰色', qty: 1, price: 189.00 }] },
  { id: 'YX20260704004', time: '2026-07-04 16:48:21', user: '刘芳', status: 'pending_receive', statusText: '待收货', amount: 798.00, payment: '银行卡', items: [{ name: '轻量跑鞋', spec: '42码/白色', qty: 1, price: 399.00 }, { name: '偏光太阳镜', spec: '黑色框', qty: 1, price: 169.00 }, { name: '智能体脂秤', spec: '白色', qty: 1, price: 129.00 }, { name: '简约双肩背包', spec: '蓝色', qty: 1, price: 101.00 }] },
  { id: 'YX20260705005', time: '2026-07-05 11:30:09', user: '陈静', status: 'pending_ship', statusText: '待发货', amount: 888.00, payment: '微信支付', items: [{ name: '真丝睡衣套装', spec: 'M码/粉色', qty: 1, price: 399.00 }, { name: '智能运动手表', spec: '曜石黑', qty: 1, price: 489.00 }] },
  { id: 'YX20260706006', time: '2026-07-06 08:15:47', user: '赵雷', status: 'cancelled', statusText: '已取消', amount: 259.00, payment: '支付宝', items: [{ name: '便携蓝牙音箱', spec: '红色', qty: 1, price: 259.00 }] },
  { id: 'YX20260707007', time: '2026-07-07 15:42:18', user: '孙悦', status: 'completed', statusText: '已完成', amount: 1268.00, payment: '银行卡', items: [{ name: '无线降噪耳机', spec: '黑色', qty: 1, price: 299.00 }, { name: '偏光太阳镜', spec: '玳瑁框', qty: 1, price: 169.00 }, { name: '轻量跑鞋', spec: '41码/蓝色', qty: 1, price: 399.00 }, { name: '智能体脂秤', spec: '黑色', qty: 1, price: 129.00 }, { name: '简约双肩背包', spec: '蓝色', qty: 1, price: 272.00 }] },
  { id: 'YX20260708008', time: '2026-07-08 20:33:05', user: '周婷', status: 'pending_receive', statusText: '待收货', amount: 398.00, payment: '微信支付', items: [{ name: '简约双肩背包', spec: '黑色', qty: 2, price: 189.00 }, { name: '偏光太阳镜', spec: '银色框', qty: 1, price: 20.00 }] },
  { id: 'YX20260709009', time: '2026-07-09 13:19:22', user: '吴磊', status: 'pending_payment', statusText: '待付款', amount: 399.00, payment: '支付宝', items: [{ name: '轻量跑鞋', spec: '43码/黑色', qty: 1, price: 399.00 }] },
  { id: 'YX20260710010', time: '2026-07-10 09:56:44', user: '郑芳', status: 'completed', statusText: '已完成', amount: 548.00, payment: '微信支付', items: [{ name: '智能运动手表', spec: '玫瑰金', qty: 1, price: 459.00 }, { name: '便携蓝牙音箱', spec: '蓝色', qty: 1, price: 89.00 }] }
]

/** 用户列表（10条） */
export const users = [
  { name: '张伟', phone: '138****1234', level: 'diamond', levelText: '钻石会员', regTime: '2024-01-15', amount: 28960.00, orders: 42, status: 'active', statusText: '正常' },
  { name: '李娜', phone: '159****5678', level: 'gold', levelText: '黄金会员', regTime: '2024-03-22', amount: 12450.00, orders: 28, status: 'active', statusText: '正常' },
  { name: '王强', phone: '186****9012', level: 'regular', levelText: '普通会员', regTime: '2024-06-08', amount: 3280.00, orders: 12, status: 'active', statusText: '正常' },
  { name: '赵敏', phone: '139****3456', level: 'gold', levelText: '黄金会员', regTime: '2024-02-14', amount: 8750.00, orders: 19, status: 'frozen', statusText: '冻结' },
  { name: '刘洋', phone: '177****7890', level: 'diamond', levelText: '钻石会员', regTime: '2023-11-03', amount: 35600.00, orders: 56, status: 'active', statusText: '正常' },
  { name: '陈静', phone: '150****2345', level: 'regular', levelText: '普通会员', regTime: '2024-08-19', amount: 1890.00, orders: 7, status: 'active', statusText: '正常' },
  { name: '杨磊', phone: '188****6789', level: 'gold', levelText: '黄金会员', regTime: '2024-04-25', amount: 6320.00, orders: 15, status: 'active', statusText: '正常' },
  { name: '黄丽', phone: '136****0123', level: 'diamond', levelText: '钻石会员', regTime: '2024-01-08', amount: 52100.00, orders: 78, status: 'active', statusText: '正常' },
  { name: '周杰', phone: '158****4567', level: 'regular', levelText: '普通会员', regTime: '2024-07-12', amount: 2560.00, orders: 9, status: 'frozen', statusText: '冻结' },
  { name: '吴婷', phone: '182****8901', level: 'gold', levelText: '黄金会员', regTime: '2023-12-20', amount: 15800.00, orders: 34, status: 'active', statusText: '正常' }
]

/** 用户统计数据 */
export const userStats = {
  totalUsers: { value: 23568, label: '总用户', icon: 'users' },
  todayNew: { value: 128, label: '今日新增', growth: '+12.3%' },
  activeUsers: { value: 8936, label: '活跃用户', sub: '本月' },
  memberUsers: { value: 5246, label: '会员用户', sub: '占比 22.3%' }
}

/** 快捷操作 */
export const quickActions = [
  { label: '添加商品', desc: '上架新品到商城', icon: 'add', color: 'blue', route: '/products/add' },
  { label: '查看订单', desc: '管理所有订单', icon: 'order', color: 'green', route: '/orders' },
  { label: '用户管理', desc: '查看用户列表', icon: 'user', color: 'orange', route: '/users' },
  { label: '系统公告', desc: '发布平台通知', icon: 'notice', color: 'purple', route: '/settings' }
]
