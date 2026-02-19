<template>
  <div class="product-container">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在加载商品信息...</p>
    </div>

    <!-- 商品主卡片 -->
    <BaseDetailCard
      v-if="!loading && product"
      :images="product.images"
      :title="product.name"
      description-title="商品描述"
      @image-load="imageLoaded"
      @image-error="imageError"
    >
      <template #codes>
        <div class="product-code">商品编码: {{ product.productCode }}</div>
        <div class="ingredient-code">食材编码: {{ product.ingredientCode }}</div>
      </template>
      
      <template #description>
        <p>{{ product.description }}</p>
      </template>
      
      <template #details>
        <div class="product-details">
          <div class="detail-row">
            <span class="detail-label">规格:</span>
            <span class="detail-value">{{ product.amount }} {{ getUnitLabel(product.unit) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">价格:</span>
            <span class="detail-value price">¥{{ product.price.toFixed(2) }}</span>
          </div>
          <div v-if="product.allergenType && product.allergenType !== 'none'" class="detail-row">
            <span class="detail-label">过敏原:</span>
            <span class="detail-value allergen">{{ getAllergenLabel(product.allergenType) }}</span>
          </div>
        </div>
      </template>
      
      <template #footer>
        <div class="dishes-header">
          <h2>相关菜品</h2>
          <div class="dishes-count">{{ dishList.length }} 道菜品</div>
        </div>
      </template>
    </BaseDetailCard>

    <!-- 菜品卡片网格 -->
    <div v-if="!loading && product" class="dishes-grid">
      <div 
        v-for="(dish, index) in dishList" 
        :key="index"
        class="dish-card-wrapper"
      >
        <DishGridCard :dish="dish" />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && !product" class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
      </svg>
      <h2>暂无商品信息</h2>
      <p>请选择其他商品或稍后再试</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '../../api/request'
import { useRouter } from 'vue-router'
import BaseDetailCard from '../../components/detailcard/BaseDetailCard.vue'
import DishGridCard from '../../components/gridcard/DishGridCard.vue'

// 类型定义
const UnitType = {
  G: 'g',
  KG: 'kg',
  ML: 'ml',
  L: 'l',
  ITEM: 'item'
} as const;

const AllergenType = {
  NONE: 'none',
  GLUTEN: 'gluten',
  MILK: 'milk',
  EGGS: 'eggs',
  FISH: 'fish',
  SHELLFISH: 'shellfish',
  TREENUTS: 'treenuts',
  PEANUTS: 'peanuts',
  WHEAT: 'wheat',
  SOYBEANS: 'soybeans'
} as const;

type UnitType = typeof UnitType[keyof typeof UnitType];
type AllergenType = typeof AllergenType[keyof typeof AllergenType];

interface ImageResponse {
  id: number;
  sortOrder: number;
  imageURL: string;
}

interface ViewDishCard {
  dishCode: string;
  name: string;
  image: ImageResponse;
}

interface ViewProductResponse {
  productCode: string;
  ingredientCode: string;
  name: string;
  amount: number;
  unit: UnitType;
  description: string;
  price: number;
  allergenType: AllergenType;
  images: ImageResponse[];
}

// 响应式数据
const product = ref<ViewProductResponse | null>(null)
const loading = ref(true)
const dishList = ref<ViewDishCard[]>([])

// 初始化数据
const initializeProductData = async () => {
  loading.value = true
  
  const productCode = useRouter().currentRoute.value.params.productCode
  
  try {
    const [productRes, dishesRes] = await Promise.all([
      request({
        url: '/api/products/' + productCode,
        method: 'GET',
      }),
      request({
        url: '/api/products/' + productCode + '/dishes',
        method: 'GET',
        params: {
          cursor: 0,
          limit: 10
        }
      })
    ])
    
    product.value = productRes.data
    dishList.value = dishesRes.data.dishes || []
    
  } catch (error) {
    console.error('Failed to fetch data:', error)
  } finally {
    loading.value = false
  }
}

// 单位标签映射
const getUnitLabel = (unit: UnitType): string => {
  const unitLabels: Record<UnitType, string> = {
    [UnitType.G]: '克',
    [UnitType.KG]: '千克',
    [UnitType.ML]: '毫升',
    [UnitType.L]: '升',
    [UnitType.ITEM]: '个'
  }
  return unitLabels[unit] || unit
}

// 过敏原标签映射
const getAllergenLabel = (allergen: AllergenType): string => {
  const allergenLabels: Record<AllergenType, string> = {
    [AllergenType.NONE]: '无',
    [AllergenType.GLUTEN]: '麸质',
    [AllergenType.MILK]: '牛奶',
    [AllergenType.EGGS]: '鸡蛋',
    [AllergenType.FISH]: '鱼类',
    [AllergenType.SHELLFISH]: '贝类',
    [AllergenType.TREENUTS]: '坚果',
    [AllergenType.PEANUTS]: '花生',
    [AllergenType.WHEAT]: '小麦',
    [AllergenType.SOYBEANS]: '大豆'
  }
  return allergenLabels[allergen] || allergen
}

// 图片事件处理（可选）
const imageLoaded = (_index: number) => {}
const imageError = (_index: number) => {}

// 生命周期钩子
onMounted(() => {
  initializeProductData()
})
</script>

<style scoped>
.product-container {
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
  border-top-color: #4CAF50;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.product-details {
  margin-bottom: 24px;
  padding: 16px;
  background-color: #f8f9fa;
  border-radius: 12px;
}

.detail-row {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-label {
  font-size: 14px;
  color: #666;
  width: 70px;
  flex-shrink: 0;
}

.detail-value {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  flex: 1;
}

.detail-value.price {
  color: #e53935;
  font-weight: 600;
  font-size: 18px;
}

.detail-value.allergen {
  color: #f57c00;
  font-weight: 500;
  background-color: rgba(245, 124, 0, 0.1);
  padding: 4px 10px;
  border-radius: 6px;
  display: inline-block;
}

.dishes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
}

.dishes-header h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: #333;
  line-height: 1.4;
}

.dishes-count {
  font-size: 14px;
  color: #4CAF50;
  font-weight: 600;
  background-color: rgba(76, 175, 80, 0.1);
  padding: 4px 10px;
  border-radius: 12px;
  white-space: nowrap;
}

.dishes-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.dish-card-wrapper {
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

.product-code, .ingredient-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

@media (min-width: 640px) {
  .dishes-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .dishes-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

@media (min-width: 1400px) {
  .product-container {
    max-width: 1400px;
    padding: 32px;
  }
}
</style>