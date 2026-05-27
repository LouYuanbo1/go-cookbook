<template>
  <div class="carousel-container" ref="carouselContainer">
    <div 
      class="carousel-track" 
      :style="{ transform: `translateX(-${currentImageIndex * 100}%)` }"
    >
      <div 
        v-for="(image, index) in images" 
        :key="index"
        class="carousel-slide"
      >
        <img 
          :src="image.imageURL" 
          :alt="altText ? altText + ' - 图片 ' + (index + 1) : `图片 ${index + 1}`"
          class="carousel-image"
          @load="onImageLoad"
          @error="onImageError"
        />
      </div>
    </div>
    
    <!-- 图片指示器 -->
    <div v-if="images.length > 1" class="carousel-indicators">
      <button
        v-for="(_, index) in images"
        :key="index"
        class="indicator"
        :class="{ active: currentImageIndex === index }"
        @click="goToSlide(index)"
        :aria-label="`跳转到图片 ${index + 1}`"
      ></button>
    </div>
    
    <!-- 导航按钮 -->
    <div v-if="images.length > 1" class="carousel-controls">
      <button 
        class="carousel-nav prev" 
        @click="prevSlide"
        :disabled="currentImageIndex === 0"
        aria-label="上一张图片"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6"></polyline>
        </svg>
      </button>
      <button 
        class="carousel-nav next" 
        @click="nextSlide"
        :disabled="currentImageIndex === images.length - 1"
        aria-label="下一张图片"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>
    </div>
    
    <!-- 图片计数 -->
    <div v-if="images.length > 1" class="image-counter">
      {{ currentImageIndex + 1 }} / {{ images.length }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'

interface ImageItem {
  imageURL: string;
  [key: string]: any;
}

const props = defineProps<{
  images: ImageItem[];
  altText?: string;
  autoSlideInterval?: number; // 毫秒，默认5000
}>()

const emit = defineEmits<{
  (e: 'image-load', index: number): void;
  (e: 'image-error', index: number): void;
}>()

const currentImageIndex = ref(0)
let autoSlideTimer: ReturnType<typeof setInterval> | null = null

const nextSlide = () => {
  if (props.images.length === 0) return
  currentImageIndex.value = (currentImageIndex.value + 1) % props.images.length
  resetAutoSlide()
}

const prevSlide = () => {
  if (props.images.length === 0) return
  currentImageIndex.value = (currentImageIndex.value - 1 + props.images.length) % props.images.length
  resetAutoSlide()
}

const goToSlide = (index: number) => {
  currentImageIndex.value = index
  resetAutoSlide()
}

const startAutoSlide = () => {
  if (props.images.length <= 1) return
  if (autoSlideTimer) clearInterval(autoSlideTimer)
  autoSlideTimer = setInterval(() => {
    nextSlide()
  }, props.autoSlideInterval || 5000)
}

const resetAutoSlide = () => {
  if (autoSlideTimer) {
    clearInterval(autoSlideTimer)
    startAutoSlide()
  }
}

const onImageLoad = (/*e: Event*/) => {
  emit('image-load', currentImageIndex.value)
}

const onImageError = (e: Event) => {
  const target = e.target as HTMLImageElement
  target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjQwMCIgaGVpZ2h0PSIzMDAiIGZpbGw9IiNGM0YzRjMiLz48cGF0aCBkPSJNMjAwIDE1MEMyMjMuODE2IDE1MCAyNDMgMTMwLjgxNiAyNDMgMTA3QzI0MyA4My4xODM4IDIyMy44MTYgNjQgMjAwIDY0QzE3Ni4xODQgNjQgMTU3IDgzLjE4MzggMTU3IDEwN0MxNTcgMTMwLjgxNiAxNzYuMTg0IDE1MCAyMDAgMTUwWiIgZmlsbD0iI0Q4RDhEOCIvPjxwYXRoIGQ9Ik0xMjAgMjI2VjIxNkMxMjAgMjA5LjM4MyAxMjUuMzgzIDIwNCAxMzIgMjA0SDI2OEMyNzQuNjE3IDIwNCAyODAgMjA5LjM4MyAyODAgMjE2VjIyNiIgc3Ryb2tlPSIjRDhEOEQ4IiBzdHJva2Utd2lkdGg9IjEwIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4='
  emit('image-error', currentImageIndex.value)
}

onMounted(() => {
  startAutoSlide()
})

onUnmounted(() => {
  if (autoSlideTimer) clearInterval(autoSlideTimer)
})

watch(() => props.images, () => {
  currentImageIndex.value = 0
  resetAutoSlide()
})
</script>

<style scoped>
.carousel-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.carousel-track {
  display: flex;
  height: 100%;
  transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.carousel-slide {
  flex: 0 0 100%;
  height: 100%;
  position: relative;
}

.carousel-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.carousel-indicators {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  gap: 8px;
  z-index: 10;
  padding: 0 20px;
}

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.5);
  border: none;
  cursor: pointer;
  padding: 0;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.indicator.active {
  background-color: white;
  transform: scale(1.2);
}

.indicator:hover {
  background-color: rgba(255, 255, 255, 0.8);
}

.carousel-controls {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 12px;
  pointer-events: none;
}

.carousel-nav {
  pointer-events: auto;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.9);
  border: none;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  z-index: 10;
}

.carousel-nav:hover:not(:disabled) {
  background-color: white;
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.carousel-nav:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.carousel-nav svg {
  color: #333;
  width: 20px;
  height: 20px;
}

.image-counter {
  position: absolute;
  bottom: 20px;
  right: 20px;
  background-color: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  z-index: 10;
  backdrop-filter: blur(4px);
}
</style>