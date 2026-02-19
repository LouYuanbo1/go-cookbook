<template>
  <router-link 
    :to="`/products/${product.productCode}`"
    class="product-card-link"
  >
    <div class="product-card">
      <!-- 商品图片 -->
      <div class="product-image-container">
        <img 
          v-if="product.image"
          :src="product.image.imageURL"
          :alt="product.name"
          class="product-image"
          @error="imageError"
        />
        <div v-else class="product-image-placeholder">
          <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
        </div>
      </div>
      
      <!-- 商品信息 -->
      <div class="product-info">
        <div class="product-name">{{ product.name }}</div>
        <div class="product-code-row">
          <div class="product-code">商品编码: {{ product.productCode }}</div>
        </div>
        
        <div class="product-details">
          <div class="product-detail-row">
            <span class="detail-label">规格:</span>
            <span class="detail-value">{{ product.amount }} {{ getUnitLabel(product.unit) }}</span>
          </div>
          <div class="product-detail-row">
            <span class="detail-label">价格:</span>
            <span class="detail-value price">¥{{ product.price.toFixed(2) }}</span>
          </div>
          <div v-if="product.allergenType && product.allergenType !== 'none'" class="product-detail-row">
            <span class="detail-label">过敏原:</span>
            <span class="detail-value allergen">{{ getAllergenLabel(product.allergenType) }}</span>
          </div>
        </div>
      </div>
      
      <!-- 跳转箭头 -->
      <div class="product-arrow">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="5" y1="12" x2="19" y2="12"></line>
          <polyline points="12 5 19 12 12 19"></polyline>
        </svg>
      </div>
    </div>
  </router-link>
</template>

<script setup lang="ts">
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

interface ViewProductCard {
  productCode: string;
  name: string;
  amount: number;
  unit: UnitType;
  description: string;
  price: number;
  allergenType: AllergenType;
  image?: { imageURL: string };
}

const props = defineProps<{
  product: ViewProductCard;
}>()

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

const imageError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjQwMCIgaGVpZ2h0PSIzMDAiIGZpbGw9IiNGM0YzRjMiLz48cGF0aCBkPSJNMjAwIDE1MEMyMjMuODE2IDE1MCAyNDMgMTMwLjgxNiAyNDMgMTA3QzI0MyA4My4xODM4IDIyMy44MTYgNjQgMjAwIDY0QzE3Ni4xODQgNjQgMTU3IDgzLjE4MzggMTU3IDEwN0MxNTcgMTMwLjgxNiAxNzYuMTg0IDE1MCAyMDAgMTUwWiIgZmlsbD0iI0Q4RDhEOCIvPjxwYXRoIGQ9Ik0xMjAgMjI2VjIxNkMxMjAgMjA5LjM4MyAxMjUuMzgzIDIwNCAxMzIgMjA0SDI2OEMyNzQuNjE3IDIwNCAyODAgMjA5LjM4MyAyODAgMjE2VjIyNiIgc3Ryb2tlPSIjRDhEOEQ4IiBzdHJva2Utd2lkdGg9IjEwIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4='
}
</script>

<style scoped>
.product-card-link {
  text-decoration: none;
  color: inherit;
  display: block;
  height: 100%;
  min-width: 0;
}

.product-card {
  background-color: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
  border: 1px solid #f0f0f0;
  min-width: 0;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  border-color: #e0e0e0;
}

.product-image-container {
  height: 140px;
  background-color: #f8f8f8;
  overflow: hidden;
  flex-shrink: 0;
}

.product-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s ease;
  display: block;
}

.product-card:hover .product-image {
  transform: scale(1.05);
}

.product-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f0f0f0;
}

.product-image-placeholder svg {
  color: #aaa;
}

.product-info {
  padding: 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.product-name {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.product-code-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.product-code,
.product-ingredient-code {
  font-size: 12px;
  color: #999;
  line-height: 1.3;
}

.product-details {
  margin-top: auto;
}

.product-detail-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
}

.product-detail-row:last-child {
  margin-bottom: 0;
}

.product-detail-row .detail-label {
  font-size: 13px;
  color: #666;
  width: 50px;
  flex-shrink: 0;
}

.product-detail-row .detail-value {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  flex: 1;
}

.product-detail-row .detail-value.price {
  color: #e53935;
  font-weight: 600;
  font-size: 14px;
}

.product-detail-row .detail-value.allergen {
  color: #f57c00;
  font-weight: 500;
  background-color: rgba(245, 124, 0, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  display: inline-block;
  font-size: 12px;
}

.product-arrow {
  position: absolute;
  bottom: 16px;
  right: 16px;
  color: #ccc;
  transition: color 0.3s ease;
  flex-shrink: 0;
}

.product-card:hover .product-arrow {
  color: #4CAF50;
}
</style>