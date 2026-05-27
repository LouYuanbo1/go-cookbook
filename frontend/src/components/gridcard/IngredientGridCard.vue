<template>
  <router-link
    :to="`/ingredients/${ingredient.ingredientCode}`"
    class="ingredient-card-link"
  >
    <div class="ingredient-card">
      <!-- 食材图片 -->
      <div class="ingredient-image-container">
        <img
          v-if="ingredient.image"
          :src="ingredient.image.imageURL"
          :alt="ingredient.name"
          class="ingredient-image"
          @error="imageError"
        />
        <div v-else class="ingredient-image-placeholder">
          <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
        </div>
      </div>

      <!-- 食材信息 -->
      <div class="ingredient-info">
        <div class="ingredient-name">{{ ingredient.name }}</div>
        <div class="ingredient-quantity">{{ ingredient.quantity }}</div>

        <div v-if="ingredient.note" class="ingredient-note">
          {{ ingredient.note }}
        </div>

        <!-- 保留注释的过敏原标签（如有需要可启用） -->
        <!--
        <div
          v-if="ingredient.allergenType"
          class="allergen-tag"
        >
          {{ getAllergenLabel(ingredient.allergenType) }}
        </div>
        -->

        <div class="ingredient-code">
          编码: {{ ingredient.ingredientCode }}
        </div>
      </div>

      <!-- 跳转箭头 -->
      <div class="ingredient-arrow">
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="5" y1="12" x2="19" y2="12"></line>
          <polyline points="12 5 19 12 12 19"></polyline>
        </svg>
      </div>
    </div>
  </router-link>
</template>

<script setup lang="ts">
import type { ViewDishIngredientCard } from "../../types/types"

defineProps<{
  ingredient: ViewDishIngredientCard
}>()

const imageError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjQwMCIgaGVpZ2h0PSIzMDAiIGZpbGw9IiNGM0YzRjMiLz48cGF0aCBkPSJNMjAwIDE1MEMyMjMuODE2IDE1MCAyNDMgMTMwLjgxNiAyNDMgMTA3QzI0MyA4My4xODM4IDIyMy44MTYgNjQgMjAwIDY0QzE3Ni4xODQgNjQgMTU3IDgzLjE4MzggMTU3IDEwN0MxNTcgMTMwLjgxNiAxNzYuMTg0IDE1MCAyMDAgMTUwWiIgZmlsbD0iI0Q4RDhEOCIvPjxwYXRoIGQ9Ik0xMjAgMjI2VjIxNkMxMjAgMjA5LjM4MyAxMjUuMzgzIDIwNCAxMzIgMjA0SDI2OEMyNzQuNjE3IDIwNCAyODAgMjA5LjM4MyAyODAgMjE2VjIyNiIgc3Ryb2tlPSIjRDhEOEQ4IiBzdHJva2Utd2lkdGg9IjEwIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4='
}
</script>

<style scoped>
.ingredient-card-link {
  text-decoration: none;
  color: inherit;
  display: block;
  height: 100%;
  min-width: 0;
}

.ingredient-card {
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

.ingredient-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  border-color: #e0e0e0;
}

.ingredient-image-container {
  height: 140px;
  background-color: #f8f8f8;
  overflow: hidden;
  flex-shrink: 0;
}

.ingredient-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s ease;
  display: block;
}

.ingredient-card:hover .ingredient-image {
  transform: scale(1.05);
}

.ingredient-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f0f0f0;
}

.ingredient-image-placeholder svg {
  color: #aaa;
}

.ingredient-info {
  padding: 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.ingredient-name {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.ingredient-quantity {
  font-size: 14px;
  color: #4caf50;
  font-weight: 600;
  margin-bottom: 8px;
  line-height: 1.4;
}

.ingredient-note {
  font-size: 13px;
  color: #666;
  line-height: 1.4;
  margin-bottom: 12px;
  padding: 8px;
  background-color: #f9f9f9;
  border-radius: 6px;
  word-break: break-word;
  flex-shrink: 0;
}

.ingredient-code {
  font-size: 12px;
  color: #999;
  margin-top: auto;
  line-height: 1.3;
}

.ingredient-arrow {
  position: absolute;
  bottom: 16px;
  right: 16px;
  color: #ccc;
  transition: color 0.3s ease;
  flex-shrink: 0;
}

.ingredient-card:hover .ingredient-arrow {
  color: #4caf50;
}
</style>