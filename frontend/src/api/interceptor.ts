// api/interceptor.ts
import axios from 'axios'
import { useTokenStore } from '../stores/token'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const tokenStore = useTokenStore()
    
    if (tokenStore.token) {
      config.headers.Authorization = `Bearer ${tokenStore.token}`
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const tokenStore = useTokenStore()
    
    if (error.response?.status === 401) {
      // Token过期或无效
      tokenStore.clearToken()
      window.location.href = '/login'
    }
    
    return Promise.reject(error)
  }
)

export default api