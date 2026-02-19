<template>
  <div>
    <h1>This is the homepage, used for debugging web pages.</h1>
    <h1>If you accidentally ended up on this page, please exit, as there is nothing here.</h1>
    <h1>这是首页，用于调试网页。</h1>
    <h1>如果您不小心来到了这个页面，请立即退出，因为这里没有任何内容。</h1>
    <div class="login-container">
    <form @submit.prevent="handleLogin" class="login-form">
      
      <div class="form-group">
        <input 
          v-model="password" 
          type="password" 
          placeholder=""
          required
        />
      </div>
      
      <button type="submit" :disabled="loading" class="submit-btn">
        {{ loading ? 'Logging' : 'Login' }}
      </button>
      
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </form>
  </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTokenStore } from '../../stores/token'
import request from '../../api/request'

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