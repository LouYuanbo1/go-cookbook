<!-- components/AdminLogin.vue -->
<template>
  <div class="login-container">
    <form @submit.prevent="handleLogin" class="login-form">
      <h2>管理员登录</h2>
      
      <div class="form-group">
        <label>密码</label>
        <input 
          v-model="password" 
          type="password" 
          placeholder="请输入密码"
          required
        />
      </div>
      
      <button type="submit" :disabled="loading" class="submit-btn">
        {{ loading ? '登录中...' : '登录' }}
      </button>
      
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTokenStore } from '../../../stores/token'
import request from '../../../api/request'

const router = useRouter()
const tokenStore = useTokenStore()

const password = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await request({
      url: 'api/auth/admin/login',
      method: 'POST',
      data: {
        password: password.value
      }
    })
    
    console.log(response)

    // 保存token
    tokenStore.saveToken(response.data)
    
    // 重定向
    const redirect = router.currentRoute.value.query.redirect as string
    router.push(redirect || '/admin/dashboard')
  } catch (err: any) {
    error.value = err.response?.data?.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>