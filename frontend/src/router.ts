// src/router.js
import { createRouter, createWebHistory } from 'vue-router'
// 导入页面组件（从 views/ 目录中）
import AdminDashboard from './views/admin/dashboard/AdminDashboard.vue'

import Dish from './views/user/Dish.vue'
import Home from './views/user/Home.vue'
import Product from './views/user/Product.vue'
import Ingredient from './views/user/Ingredient.vue'
import CreateIngredient from './views/admin/ingredient/CreateIngredient.vue'
import UpdateIngredient from './views/admin/ingredient/UpdateIngredient.vue'
import DeleteIngredient from './views/admin/ingredient/DeleteIngredient.vue'

import CreateProduct from './views/admin/product/CreateProduct.vue'
import UpdateProduct from './views/admin/product/UpdateProduct.vue'
import DeleteProduct from './views/admin/product/DeleteProduct.vue'


import CreateDish from './views/admin/dish/CreateDish.vue'
import UpdateDish from './views/admin/dish/UpdateDish.vue'
import DeleteDish from './views/admin/dish/DeleteDish.vue'



// 定义路由规则：路径 ↔ 页面组件
const routes = [
  // 首页路由
  { path: '/', name: 'Home', component: Home },
  // 菜单路由
  {path: '/admin/dashboard', name: 'AdminDashboard', component: AdminDashboard},

  { path: '/dishes/:dishCode', name: 'Dish', component: Dish },
  { path: '/ingredients/:ingredientCode', name: 'Ingredient', component: Ingredient },
  { path: '/products/:productCode', name: 'Product', component: Product },

  { path: '/ingredients/create', name: 'CreateIngredient', component: CreateIngredient,meta: { requiresAuth: true } },
  { path: '/ingredients/update', name: 'UpdateIngredient', component: UpdateIngredient,meta: { requiresAuth: true } },
  { path: '/ingredients/delete', name: 'DeleteIngredient', component: DeleteIngredient,meta: { requiresAuth: true } },

  { path: '/products/create', name: 'CreateProduct', component: CreateProduct,meta: { requiresAuth: true } },
  { path: '/products/update', name: 'UpdateProduct', component: UpdateProduct,meta: { requiresAuth: true } },
  { path: '/products/delete', name: 'DeleteProduct', component: DeleteProduct,meta: { requiresAuth: true } },

  { path: '/dishes/create', name: 'CreateDish', component: CreateDish,meta: { requiresAuth: true } },
  { path: '/dishes/update', name: 'UpdateDish', component: UpdateDish,meta: { requiresAuth: true } },
  { path: '/dishes/delete', name: 'DeleteDish', component: DeleteDish,meta: { requiresAuth: true } },
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(), // HTML5 历史模式（无 # 号，更美观）
  routes // 注入路由规则
})

export default router