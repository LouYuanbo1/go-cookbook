<template>
  <div class="recipe-method-card">
    <h2>做法步骤</h2>
    <div v-if="recipe" class="recipe-content">{{ recipe }}</div>
    <div v-else class="empty-text">暂无做法描述</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  recipe: string
}>()
</script>

<style scoped>
.recipe-method-card {
  background-color: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
}

.recipe-method-card h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #333;
  line-height: 1.4;
}

.recipe-content {
  font-size: 15px;
  line-height: 1.6;
  color: #555;
  white-space: pre-line;
  word-break: break-word;
}

.empty-text {
  color: #999;
  font-style: italic;
}

@media (max-width: 639px) {
  .recipe-method-card {
    padding: 20px 16px;
  }
}

@media (min-width: 1024px) {
  .recipe-method-card h2 {
    font-size: 20px;
  }
  .recipe-content {
    font-size: 16px;
  }
}
</style>