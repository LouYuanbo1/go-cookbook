<template>
  <div class="recipe-container">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在加载食谱信息...</p>
    </div>

    <!-- 食谱主卡片 -->
    <BaseDetailCard
      v-if="!loading && dish"
      :images="dish.images || []"
      :title="dish.name"
      description-title="菜品描述"
      @image-load="imageLoaded"
      @image-error="imageError"
    >
      <template #codes>
        <div class="recipe-code">菜品编码: {{ dish.dishCode }}</div>
      </template>

      <template #description>
        <p>{{ dish.description }}</p>
      </template>

      <template #footer>
        <div class="ingredients-header">
          <h2>所需食材</h2>
          <div class="ingredients-count">{{ ingredientList.length }} 种食材</div>
        </div>
      </template>
    </BaseDetailCard>

    <!-- 做法卡片 -->
    <RecipeMethodCard
      v-if="!loading && dish"
      :recipe="dish.recipe"
    />

    <!-- 食材卡片网格 -->
    <div v-if="!loading && dish" class="ingredients-grid">
      <div
        v-for="(ingredient, index) in ingredientList"
        :key="index"
        class="ingredient-card-wrapper"
      >
        <IngredientGridCard :ingredient="ingredient" />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && !dish" class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
      </svg>
      <h2>暂无食谱信息</h2>
      <p>请选择其他食谱或稍后再试</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '../../api/request'
import { useRouter } from 'vue-router'
import type { ImageResponse, ViewDishIngredientCard } from '../../types/types'
import BaseDetailCard from '../../components/detailcard/BaseDetailCard.vue'
import RecipeMethodCard from '../../components/detailcard/RecipeMethodCard.vue'
import IngredientGridCard from '../../components/gridcard/IngredientGridCard.vue'

interface ViewDishResponse {
  dishCode: string
  name: string
  description: string
  recipe: string
  images: ImageResponse[]
}

// 响应式数据
const dish = ref<ViewDishResponse | null>(null)
const ingredientList = ref<ViewDishIngredientCard[]>([])
const loading = ref(true)

// 初始化数据
const initializeDishData = async () => {
  loading.value = true

  const dishCode = useRouter().currentRoute.value.params.dishCode

  try {
    const [dishRes, ingredientsRes] = await Promise.all([
      request({
        url: '/api/dishes/' + dishCode,
        method: 'GET',
      }),
      request({
        url: '/api/dishes/' + dishCode + '/ingredients',
        method: 'GET',
        params: {
          cursor: 0,
          limit: 10,
        },
      }),
    ])

    dish.value = dishRes.data
    ingredientList.value = ingredientsRes.data?.dishIngredients || []
  } catch (error) {
    console.error('Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

// 图片事件处理（可选）
const imageLoaded = (_index: number) => {}
const imageError = (_index: number) => {}

// 生命周期钩子
onMounted(() => {
  initializeDishData()
})
</script>

<style scoped>
.recipe-container {
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

.ingredients-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
}

.ingredients-header h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: #333;
  line-height: 1.4;
}

.ingredients-count {
  font-size: 14px;
  color: #4caf50;
  font-weight: 600;
  background-color: rgba(76, 175, 80, 0.1);
  padding: 4px 10px;
  border-radius: 12px;
  white-space: nowrap;
}

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

.recipe-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

@media (max-width: 639px) {
  .recipe-container {
    padding: 12px;
  }

  .ingredients-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .ingredients-count {
    align-self: flex-start;
  }
}

@media (min-width: 640px) {
  .ingredients-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .ingredients-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

@media (min-width: 1400px) {
  .recipe-container {
    max-width: 1400px;
    padding: 32px;
  }
}
</style>