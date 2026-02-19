// stores/token.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

export const useTokenStore = defineStore('token', () => {
  const router = useRouter()
  
  // 响应式状态
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const tokenExpiry = ref<number>(parseInt(localStorage.getItem('token_expiry') || '0'))
  
  // 计算属性：检查token是否有效
  const isValid = computed(() => {
    return token.value && tokenExpiry.value > Date.now()
  })
  
  // 保存token
  const saveToken = (newToken: string, expiresInHours: number = 24) => {
    token.value = newToken
    tokenExpiry.value = Date.now() + expiresInHours * 60 * 60 * 1000
    
    // 持久化到localStorage
    localStorage.setItem('admin_token', newToken)
    localStorage.setItem('token_expiry', tokenExpiry.value.toString())
  }
  
  // 清除token
  const clearToken = () => {
    token.value = null
    tokenExpiry.value = 0
    localStorage.removeItem('admin_token')
    localStorage.removeItem('token_expiry')
  }
  
  // 检查并处理token过期
  const checkToken = async (): Promise<boolean> => {
    if (!isValid.value) {
      clearToken()
      await router.push('/login')
      return false
    }
    return true
  }
  
  // 自动续期（可选）
  const autoRenew = async () => {
    if (!token.value) return
    
    // 在过期前5分钟尝试续期
    const timeToExpiry = tokenExpiry.value - Date.now()
    if (timeToExpiry < 5 * 60 * 1000 && timeToExpiry > 0) {
      try {
        const response = await axios.post('/api/admin/renew-token', null, {
          headers: { Authorization: `Bearer ${token.value}` }
        })
        
        if (response.data.token) {
          saveToken(response.data.token, response.data.expiresIn)
        }
      } catch (error) {
        console.error('Token renewal failed:', error)
        clearToken()
        await router.push('/login')
      }
    }
  }
  
  return {
    token,
    isValid,
    saveToken,
    clearToken,
    checkToken,
    autoRenew
  }
})