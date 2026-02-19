import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router.ts' // 导入路由实例
import { createPinia } from 'pinia'
import { useTokenStore } from './stores/token.ts'
import { setupAdminGuard } from './api/guard'



const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 设置路由守卫
setupAdminGuard(router)

// 启动token自动续期检查
const tokenStore = useTokenStore()
setInterval(() => {
  tokenStore.autoRenew()
}, 60000) // 每分钟检查一次

app.mount('#app')
