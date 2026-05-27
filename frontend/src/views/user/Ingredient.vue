<template>
  <div class="ingredient-container">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在加载食材信息...</p>
    </div>

    <!-- 食材主卡片 -->
    <BaseDetailCard
      v-if="!loading && ingredient"
      :images="ingredient.images"
      :title="ingredient.name"
      description-title="食材描述"
      @image-load="imageLoaded"
      @image-error="imageError"
    >
      <template #codes>
        <div class="ingredient-code">食材编码: {{ ingredient.ingredientCode }}</div>
      </template>

      <template #description>
        <p>{{ ingredient.description }}</p>
      </template>

      <template #footer>
        <div class="products-header">
          <h2>相关商品</h2>
          <div class="products-count">{{ productList.length }} 个商品</div>
        </div>
      </template>
    </BaseDetailCard>

    <!-- 商品卡片网格 -->
    <div v-if="!loading && ingredient" class="products-grid">
      <div
        v-for="(product, index) in productList"
        :key="index"
        class="product-card-wrapper"
      >
        <ProductGridCard :product="product" />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && !ingredient" class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
      </svg>
      <h2>暂无食材信息</h2>
      <p>请选择其他食材或稍后再试</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '../../api/request'
import { useRouter } from 'vue-router'
import BaseDetailCard from '../../components/detailcard/BaseDetailCard.vue'
import ProductGridCard from '../../components/gridcard/ProductGridCard.vue'

// 类型定义（可根据实际情况抽取到公共文件）
interface ImageResponse {
  id: number
  sortOrder: number
  imageURL: string
}

interface ViewProductCard {
  productCode: string
  name: string
  amount: number
  unit: 'g' | 'kg' | 'ml' | 'l' | 'item'
  description: string
  price: number
  allergenType: 'none' | 'gluten' | 'milk' | 'eggs' | 'fish' | 'shellfish' | 'treenuts' | 'peanuts' | 'wheat' | 'soybeans'
  image: ImageResponse
}

interface ViewIngredientResponse {
  ingredientCode: string
  name: string
  description: string
  images: ImageResponse[]
}

// 响应式数据
const ingredient = ref<ViewIngredientResponse | null>(null)
const loading = ref(true)
const productList = ref<ViewProductCard[]>([])

// 初始化数据
const initializeIngredientData = async () => {
  loading.value = true

  const ingredientCode = useRouter().currentRoute.value.params.ingredientCode

  try {
    const [ingredientRes, productsRes] = await Promise.all([
      request({
        url: '/api/ingredients/' + ingredientCode,
        method: 'GET',
      }),
      request({
        url: '/api/ingredients/' + ingredientCode + '/products',
        method: 'GET',
        params: {
          cursor: 0,
          limit: 10,
        },
      }),
    ])

    ingredient.value = ingredientRes.data
    productList.value = productsRes.data.products || []
  } catch (error) {
    console.error('Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

// 图片事件处理（可选）
const imageLoaded = (_index: number) => {
  // 可以在这里添加图片加载完成的处理逻辑
}
const imageError = (_index: number) => {
  // 可以在这里添加图片加载错误的处理逻辑
}

// 生命周期钩子
onMounted(() => {
  initializeIngredientData()
})
</script>

<style scoped>
.ingredient-container {
  font-family: 'PingFang SC', 'Helvetica Neue', Arial, sans-serif;  
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
  color: #333;
  background-color: #f9f9f9;
  min-height: 100vh;
  box-sizing: border-box;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 80vh;
  color: #666;
}

.loading-spinner {
  width: 50px;
  height: 50px;
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

.products-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
}

.products-header h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: #333;
  line-height: 1.4;
}

.products-count {
  font-size: 14px;
  color: #4caf50;
  font-weight: 600;
  background-color: rgba(76, 175, 80, 0.1);
  padding: 4px 10px;
  border-radius: 12px;
  white-space: nowrap;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.product-card-wrapper {
  height: 100%;
  min-width: 0;
}

.empty-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
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
  line-height: 1.3;
}

.empty-state p {
  font-size: 15px;
  margin: 0;
  line-height: 1.5;
}

.ingredient-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

@media (min-width: 640px) {
  .products-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .products-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

@media (min-width: 1400px) {
  .ingredient-container {
    max-width: 1400px;
    padding: 32px;
  }
}
</style>