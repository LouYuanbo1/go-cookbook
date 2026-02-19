<template>
  <div class="base-detail-card">
    <!-- 图片轮播区域 -->
    <div class="detail-image-carousel">
      <ImageCarousel 
        :images="images" 
        :alt-text="title"
        @image-load="$emit('image-load', $event)"
        @image-error="$emit('image-error', $event)"
      />
    </div>

    <!-- 右侧内容区域 -->
    <div class="detail-content">
      <div class="detail-header">
        <h1 class="detail-title">{{ title }}</h1>
        <div class="detail-codes">
          <slot name="codes"></slot>
        </div>
      </div>
      
      <!-- 描述区域（可选） -->
      <div v-if="$slots.description" class="detail-description">
        <h2>{{ descriptionTitle || '描述' }}</h2>
        <slot name="description"></slot>
      </div>
      
      <!-- 额外详情区域（可选） -->
      <div v-if="$slots.details" class="detail-details">
        <slot name="details"></slot>
      </div>
      
      <!-- 底部区域（可选） -->
      <div v-if="$slots.footer" class="detail-footer">
        <slot name="footer"></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ImageCarousel from '../image/ImageCarousel.vue'

defineProps<{
  images: any[]; // 图片数组
  title: string; // 标题
  descriptionTitle?: string; // 描述区域的标题，默认"描述"
}>()

defineEmits<{
  (e: 'image-load', index: number): void;
  (e: 'image-error', index: number): void;
}>()
</script>

<style scoped>
.base-detail-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  display: flex;
  flex-direction: column;
}

.detail-image-carousel {
  position: relative;
  width: 100%;
  height: 0;
  padding-bottom: 60%; /* 5:3 宽高比 */
  overflow: hidden;
  background-color: #f5f5f5;
  flex-shrink: 0;
}

.detail-content {
  padding: 24px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  box-sizing: border-box;
}

.detail-header {
  margin-bottom: 20px;
  flex-shrink: 0;
}

.detail-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.detail-codes {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-description {
  margin-bottom: 20px;
  flex: 1;
  min-height: 0;
}

.detail-description h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 12px 0;
  color: #333;
  line-height: 1.4;
}

.detail-details {
  margin-bottom: 24px;
}

.detail-footer {
  margin-top: auto;
}

/* 响应式调整 */
@media (min-width: 768px) {
  .base-detail-card {
    flex-direction: row;
    min-height: 400px;
  }
  
  .detail-image-carousel {
    flex: 0 0 50%;
    height: auto;
    padding-bottom: 0;
    max-height: 400px;
  }
  
  .detail-content {
    flex: 0 0 50%;
    padding: 32px;
    max-width: 50%;
    overflow-y: auto;
  }
  
  .detail-title {
    font-size: 28px;
  }
  
  .detail-description {
    margin-bottom: 32px;
  }
}

@media (min-width: 1024px) {
  .detail-title {
    font-size: 32px;
  }
  
  .detail-description h2 {
    font-size: 20px;
  }
}

@media (min-width: 1400px) {
  .base-detail-card {
    min-height: 500px;
  }
  
  .detail-image-carousel {
    max-height: 500px;
  }
  
  .detail-content {
    padding: 40px;
  }
  
  .detail-title {
    font-size: 36px;
  }
}
</style>