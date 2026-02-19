// api/guard.ts
import { useTokenStore } from '../stores/token'


export function setupAdminGuard(router: any) {
  router.beforeEach((to: any) => {
    const tokenStore = useTokenStore()
    
    if (to.meta.requiresAuth && !tokenStore.isValid) {
      return { 
        path: '/', 
        //query: { redirect: to.fullPath },
        replace: true // 使用 replace 避免返回按钮回到受保护页面
      }
    }
  })
}