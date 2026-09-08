<template>
  <div class="home-container">
    <!-- 页面头部 -->
    <header class="home-hero">
      <!-- 管理员登录入口 -->
      <button class="admin-entry-btn" @click="handleAdminEntry" :title="tokenStore.isValid ? '进入管理后台' : '管理员登录'">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
          <circle cx="12" cy="7" r="4"></circle>
        </svg>
        <span class="admin-entry-text">{{ tokenStore.isValid ? '后台' : '登录' }}</span>
      </button>

      <div class="hero-content">
        <div class="hero-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 8h1a4 4 0 0 1 0 8h-1"></path>
            <path d="M2 8h16v9a4 4 0 0 1-4 4H6a4 4 0 0 1-4-4V8z"></path>
            <line x1="6" y1="1" x2="6" y2="4"></line>
            <line x1="10" y1="1" x2="10" y2="4"></line>
            <line x1="14" y1="1" x2="14" y2="4"></line>
          </svg>
        </div>
        <h1 class="hero-title">Go Cookbook</h1>
        <p class="hero-subtitle">探索食材，发现美食的无限可能</p>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-value">{{ totalCount }}</span>
            <span class="stat-label">食材</span>
          </div>
        </div>
      </div>
    </header>

    <!-- 食材网格 -->
    <section class="ingredients-section">
      <div class="section-header">
        <h2 class="section-title">全部食材</h2>
        <span class="section-count">{{ ingredientList.length }} / {{ totalCount }}</span>
      </div>

      <!-- 加载状态 -->
      <div v-if="initialLoading" class="loading-overlay">
        <div class="loading-spinner"></div>
        <p>正在加载食材...</p>
      </div>

      <!-- 食材卡片网格 -->
      <div v-else-if="ingredientList.length > 0" class="ingredients-grid">
        <div
          v-for="ingredient in ingredientList"
          :key="ingredient.id"
          class="ingredient-card-wrapper"
        >
          <IngredientGridCard :ingredient="ingredient" />
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="8" y1="15" x2="16" y2="15"></line>
          <line x1="9" y1="9" x2="9.01" y2="9"></line>
          <line x1="15" y1="9" x2="15.01" y2="9"></line>
        </svg>
        <h2>暂无食材</h2>
        <p>请先添加一些食材再来看看吧</p>
      </div>

      <!-- 加载更多 -->
      <div v-if="hasMore && !initialLoading" class="load-more-container">
        <button
          class="load-more-btn"
          :disabled="loadingMore"
          @click="loadMore"
        >
          <span v-if="loadingMore" class="btn-loading-spinner"></span>
          {{ loadingMore ? '加载中...' : '加载更多食材' }}
        </button>
      </div>

      <!-- 加载更多出错 -->
      <div v-if="loadError" class="load-error">
        <p>{{ loadError }}</p>
        <button class="retry-btn" @click="loadMore">重试</button>
      </div>
    </section>

    <!-- 管理员登录弹窗 -->
    <Teleport to="body">
      <div v-if="showLoginModal" class="modal-overlay" @click.self="closeLoginModal">
        <div class="login-modal">
          <div class="modal-header">
            <h3>管理员登录</h3>
            <button class="modal-close-btn" @click="closeLoginModal">&times;</button>
          </div>
          <form @submit.prevent="handleLogin" class="login-form">
            <div class="form-group">
              <label class="form-label">管理员密码</label>
              <input
                v-model="loginPassword"
                type="password"
                class="form-input"
                placeholder="请输入管理员密码"
                required
                autofocus
              />
            </div>
            <div v-if="loginError" class="login-error">{{ loginError }}</div>
            <button type="submit" class="login-submit-btn" :disabled="loginLoading">
              <span v-if="loginLoading" class="btn-loading-spinner"></span>
              {{ loginLoading ? '登录中...' : '登录' }}
            </button>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '../../api/request'
import { useTokenStore } from '../../stores/token'
import IngredientGridCard from '../../components/gridcard/IngredientGridCard.vue'
import type { IngredientCardResp, IngredientCardCursorResp } from '../../types/types'

const router = useRouter()
const tokenStore = useTokenStore()

const PAGE_SIZE = 12

const ingredientList = ref<IngredientCardResp[]>([])
const initialLoading = ref(true)
const loadingMore = ref(false)
const loadError = ref('')
const cursor = ref(0)
const hasMore = ref(false)
const totalCount = ref(0)

// 登录弹窗相关
const showLoginModal = ref(false)
const loginPassword = ref('')
const loginLoading = ref(false)
const loginError = ref('')

const closeLoginModal = () => {
  showLoginModal.value = false
  loginPassword.value = ''
  loginError.value = ''
}

const handleLogin = async () => {
  loginLoading.value = true
  loginError.value = ''

  try {
    const response = await request({
      url: '/api/auth/admin/login',
      method: 'POST',
      data: {
        password: loginPassword.value,
      },
    })

    tokenStore.saveToken(response.data)
    closeLoginModal()
    router.push('/admin/dashboard')
  } catch (err: any) {
    loginError.value = err.response?.data?.message || '密码错误，请重试'
  } finally {
    loginLoading.value = false
  }
}

const handleAdminEntry = () => {
  if (tokenStore.isValid) {
    router.push('/admin/dashboard')
  } else {
    showLoginModal.value = true
  }
}

const fetchIngredients = async (isLoadMore = false) => {
  if (isLoadMore) {
    loadingMore.value = true
  } else {
    initialLoading.value = true
  }
  loadError.value = ''

  try {
    const response = await request({
      url: '/api/ingredients',
      method: 'GET',
      params: {
        cursor: cursor.value,
        limit: PAGE_SIZE,
      },
    })

    const data = response.data as IngredientCardCursorResp
    const items = data.items || []

    if (isLoadMore) {
      ingredientList.value = [...ingredientList.value, ...items]
    } else {
      ingredientList.value = items
      totalCount.value = items.length
    }

    cursor.value = data.cursor
    hasMore.value = data.has_more
  } catch (error) {
    console.error('获取食材列表失败:', error)
    loadError.value = '加载失败，请检查网络连接后重试'
  } finally {
    initialLoading.value = false
    loadingMore.value = false
  }
}

const loadMore = () => {
  if (!hasMore.value || loadingMore.value) return
  fetchIngredients(true)
}

onMounted(() => {
  fetchIngredients()
})
</script>

<style scoped>
.home-container {
  font-family: 'PingFang SC', 'Helvetica Neue', Arial, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px 32px;
  color: #333;
  background-color: #f9f9f9;
  min-height: 100vh;
  box-sizing: border-box;
}

/* ===== 页面头部 ===== */
.home-hero {
  position: relative;
  background: linear-gradient(135deg, #43a047 0%, #66bb6a 50%, #81c784 100%);
  border-radius: 16px;
  padding: 40px 24px;
  margin-bottom: 32px;
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
  color: white;
  box-shadow: 0 4px 20px rgba(76, 175, 80, 0.25);
}

.hero-content {
  max-width: 600px;
}

.hero-icon {
  margin-bottom: 16px;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  width: 72px;
  height: 72px;
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  backdrop-filter: blur(4px);
}

.hero-icon svg {
  color: white;
}

.hero-title {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 8px 0;
  letter-spacing: 1px;
  line-height: 1.2;
}

.hero-subtitle {
  font-size: 16px;
  margin: 0 0 24px 0;
  opacity: 0.9;
  font-weight: 400;
  line-height: 1.5;
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 32px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  opacity: 0.85;
  font-weight: 500;
}

/* ===== 食材列表区域 ===== */
.ingredients-section {
  padding: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 8px;
}

.section-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: #222;
  line-height: 1.4;
}

.section-count {
  font-size: 14px;
  color: #4caf50;
  font-weight: 600;
  background-color: rgba(76, 175, 80, 0.1);
  padding: 4px 12px;
  border-radius: 12px;
  white-space: nowrap;
}

/* ===== 加载状态 ===== */
.loading-overlay {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  color: #666;
}

.loading-spinner {
  width: 44px;
  height: 44px;
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top-color: #4caf50;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ===== 食材网格 ===== */
.ingredients-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.ingredient-card-wrapper {
  height: 100%;
  min-width: 0;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  text-align: center;
  color: #999;
  padding: 40px 20px;
}

.empty-state svg {
  color: #ddd;
  margin-bottom: 20px;
}

.empty-state h2 {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #555;
  line-height: 1.3;
}

.empty-state p {
  font-size: 15px;
  margin: 0;
  line-height: 1.5;
  color: #999;
}

/* ===== 加载更多 ===== */
.load-more-container {
  display: flex;
  justify-content: center;
  padding: 16px 0 32px;
}

.load-more-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  font-size: 15px;
  font-weight: 600;
  color: #4caf50;
  background-color: white;
  border: 2px solid #4caf50;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
}

.load-more-btn:hover:not(:disabled) {
  background-color: #4caf50;
  color: white;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
}

.load-more-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(76, 175, 80, 0.3);
  border-radius: 50%;
  border-top-color: #4caf50;
  animation: spin 0.8s linear infinite;
}

/* ===== 加载错误 ===== */
.load-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
  color: #e53935;
  text-align: center;
}

.load-error p {
  margin: 0;
  font-size: 14px;
}

.retry-btn {
  padding: 8px 24px;
  font-size: 14px;
  font-weight: 600;
  color: #e53935;
  background-color: white;
  border: 1px solid #e53935;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
}

.retry-btn:hover {
  background-color: #e53935;
  color: white;
}

/* ===== 管理员登录按钮 ===== */
.admin-entry-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  background-color: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  backdrop-filter: blur(4px);
  font-family: inherit;
}

.admin-entry-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
  color: white;
  border-color: rgba(255, 255, 255, 0.5);
  transform: translateY(-1px);
}

.admin-entry-text {
  white-space: nowrap;
}

/* ===== 登录弹窗 ===== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.login-modal {
  background-color: white;
  border-radius: 16px;
  padding: 32px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
  color: #222;
}

.modal-close-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  font-size: 24px;
  color: #999;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s ease;
  line-height: 1;
}

.modal-close-btn:hover {
  background-color: #f0f0f0;
  color: #333;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 14px;
  font-weight: 600;
  color: #555;
}

.form-input {
  padding: 12px 16px;
  font-size: 15px;
  border: 2px solid #e0e0e0;
  border-radius: 10px;
  outline: none;
  transition: border-color 0.3s ease;
  font-family: inherit;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: #4caf50;
  box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}

.login-error {
  font-size: 14px;
  color: #e53935;
  background-color: rgba(229, 57, 53, 0.08);
  padding: 10px 14px;
  border-radius: 8px;
  text-align: center;
}

.login-submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 24px;
  font-size: 16px;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #43a047, #66bb6a);
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
  margin-top: 4px;
}

.login-submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(76, 175, 80, 0.4);
}

.login-submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ===== 响应式 ===== */
@media (min-width: 640px) {
  .home-hero {
    padding: 48px 32px;
  }

  .hero-title {
    font-size: 42px;
  }

  .ingredients-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .home-hero {
    padding: 56px 40px;
  }

  .hero-title {
    font-size: 48px;
  }

  .ingredients-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

@media (min-width: 1400px) {
  .home-container {
    max-width: 1400px;
    padding: 0 32px 40px;
  }

  .ingredients-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
  }
}
</style>