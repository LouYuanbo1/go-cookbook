<template>
  <div class="image-manager">
    <!-- 上传按钮，仅当未达上限可用 -->
    <div
      v-if="innerImageList.length < maxCount"
      class="upload-area"
      @click="triggerFileInput"
    >
      <input
        ref="fileInputRef"
        type="file"
        :accept="accept"
        multiple
        @change="handleFileChange"
        class="hidden-input"
      />
      <div class="upload-placeholder">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="32"
          height="32"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
          <circle cx="8.5" cy="8.5" r="1.5"></circle>
          <polyline points="21 15 16 10 5 21"></polyline>
        </svg>
        <span>{{ uploadText }}</span>
        <span class="upload-hint">{{ uploadHint }}</span>
      </div>
    </div>

    <!-- 可拖拽排序的图片预览网格 -->
    <draggable
      v-if="innerImageList.length"
      v-model="innerImageList"
      group="images"
      :item-key="itemKey"
      class="products-grid preview-grid"
      tag="div"
      :animation="200"
      :handle="handleSelector"
      @end="syncOrder"
    >
      <template #item="{ element, index }">
        <div class="product-card preview-card">
          <div class="product-image-container preview-image">
            <img
              :src="element.url"
              :alt="`图片 ${index + 1}`"
              class="product-image"
              @error="handleImageError($event, element)"
            />
          </div>
          <div class="preview-info">
            <div class="preview-header">
              <span v-if="showDragHandle" class="drag-handle" :title="dragHandleTitle">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <circle cx="9" cy="12" r="1"></circle>
                  <circle cx="9" cy="5" r="1"></circle>
                  <circle cx="9" cy="19" r="1"></circle>
                  <circle cx="15" cy="12" r="1"></circle>
                  <circle cx="15" cy="5" r="1"></circle>
                  <circle cx="15" cy="19" r="1"></circle>
                </svg>
              </span>
              <span class="preview-filename">
                {{ getFileName(element) }}
              </span>
            </div>
            <span class="preview-size">
              {{ getFileSize(element) }}
            </span>
          </div>
          <button
            type="button"
            class="preview-remove"
            @click="removeImage(index)"
            :aria-label="removeButtonAriaLabel"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="3 6 5 6 21 6"></polyline>
              <path
                d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              ></path>
              <line x1="10" y1="11" x2="10" y2="17"></line>
              <line x1="14" y1="11" x2="14" y2="17"></line>
            </svg>
          </button>
        </div>
      </template>
    </draggable>

    <!-- 已达上限提示 -->
    <div v-if="innerImageList.length >= maxCount" class="upload-limit">
      {{ limitMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue';
import draggable from 'vuedraggable';

// ---------- 类型定义 ----------
export interface ImageItem {
  id?: number;        // 现有图片有ID，新图片无
  tempID?: string;     // 新图片上传时使用，用于匹配后端返回的URL
  url: string;        // 预览URL（现有图片为真实URL，新图片为blob）
  file?: File;        // 新图片才有file
  status: 'existing' | 'new';
  sortOrder: number;  // 实时顺序（仅在初始化时使用，拖拽后忽略）
}

// ---------- Props ----------
const props = withDefaults(
  defineProps<{
    /** 双向绑定：图片列表 */
    modelValue: ImageItem[];
    /** 最大图片数量 */
    maxCount?: number;
    /** 允许的文件类型 */
    accept?: string;
    /** 上传区域文案 */
    uploadText?: string;
    /** 上传区域提示 */
    uploadHint?: string;
    /** 拖拽手柄选择器（传递给draggable） */
    handleSelector?: string;
    /** 是否显示拖拽手柄图标 */
    showDragHandle?: boolean;
    /** 拖拽手柄title */
    dragHandleTitle?: string;
    /** 删除按钮aria-label */
    removeButtonAriaLabel?: string;
    /** 已达上限提示 */
    limitMessage?: string;
    /** 用于draggable的item-key，默认'url' */
    itemKey?: string;
    /** 默认占位图（图片加载失败时显示） */
    defaultImage?: string;
  }>(),
  {
    maxCount: 9,
    accept: 'image/*',
    uploadText: '点击添加新图片',
    uploadHint: '支持 jpg、png、webp',
    handleSelector: '.drag-handle',
    showDragHandle: true,
    dragHandleTitle: '拖动调整顺序',
    removeButtonAriaLabel: '删除',
    limitMessage: '已达到最大图片数量',
    itemKey: 'url',
    defaultImage:
      'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjQwMCIgaGVpZ2h0PSIzMDAiIGZpbGw9IiNGM0YzRjMiLz48cGF0aCBkPSJNMjAwIDE1MEMyMjMuODE2IDE1MCAyNDMgMTMwLjgxNiAyNDMgMTA3QzI0MyA4My4xODM4IDIyMy44MTYgNjQgMjAwIDY0QzE3Ni4xODQgNjQgMTU3IDgzLjE4MzggMTU3IDEwN0MxNTcgMTMwLjgxNiAxNzYuMTg0IDE1MCAyMDAgMTUwWiIgZmlsbD0iI0Q4RDhEOCIvPjxwYXRoIGQ9Ik0xMjAgMjI2VjIxNkMxMjAgMjA5LjM4MyAxMjUuMzgzIDIwNCAxMzIgMjA0SDI2OEMyNzQuNjE3IDIwNCAyODAgMjA5LjM4MyAyODAgMjE2VjIyNiIgc3Ryb2tlPSIjRDhEOEQ4IiBzdHJva2Utd2lkdGg9IjEwIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4=',
  }
);

// ---------- Emits ----------
const emit = defineEmits<{
  (e: 'update:modelValue', value: ImageItem[]): void;
  (e: 'update:deletedIds', ids: Set<number>): void; // 被删除的现有图片ID集合
}>();

// ---------- 内部状态 ----------
const innerImageList = ref<ImageItem[]>([...props.modelValue]);
const deletedImageIds = ref<Set<number>>(new Set());
const fileInputRef = ref<HTMLInputElement | null>(null);

// ---------- 单向同步：外部 prop → 内部（不触发 emit）----------
watch(
  () => props.modelValue,
  (val) => {
    innerImageList.value = val.map(item => ({ ...item })); // 浅拷贝，避免引用相同
  },
  { deep: true, immediate: true }
);

// ---------- 监听 deletedImageIds 并通知父组件 ----------
watch(deletedImageIds, (val) => {
  emit('update:deletedIds', new Set(val)); // 传递副本
});

// ---------- 辅助函数：将内部最新状态同步给父组件 ----------
const syncToParent = () => {
  emit('update:modelValue', innerImageList.value);
};

// ---------- 图片上传 ----------
const triggerFileInput = () => {
  if (innerImageList.value.length >= props.maxCount) {
    alert(props.limitMessage);
    return;
  }
  fileInputRef.value?.click();
};

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement;
  const files = Array.from(target.files || []);
  if (files.length === 0) return;

  const total = innerImageList.value.length + files.length;
  if (total > props.maxCount) {
    alert(`最多只能上传 ${props.maxCount} 张图片，当前已选 ${innerImageList.value.length} 张，本次新增 ${files.length} 张将超出限制`);
    if (fileInputRef.value) fileInputRef.value.value = '';
    return;
  }

  files.forEach((file) => {
    const preview = URL.createObjectURL(file);
    innerImageList.value.push({
      file,
      url: preview,
      status: 'new',
      sortOrder: innerImageList.value.length,
    });
  });

  syncToParent(); // 上传后通知父组件更新
  if (fileInputRef.value) fileInputRef.value.value = '';
};

// ---------- 删除图片 ----------
const removeImage = (index: number) => {
  const item = innerImageList.value[index];
  if (!item) return;

  // 记录删除的现有图片ID
  if (item.status === 'existing' && item.id) {
    deletedImageIds.value.add(item.id);
  }

  // 释放新图片的blob URL
  if (item.status === 'new' && item.url.startsWith('blob:')) {
    URL.revokeObjectURL(item.url);
  }

  innerImageList.value.splice(index, 1);
  syncToParent(); // 删除后通知父组件更新
};

// ---------- 拖拽排序结束同步 ----------
const syncOrder = () => {
  // 拖拽后 innerImageList 顺序已由 draggable 自动更新，只需通知父组件
  syncToParent();
};

// ---------- 图片加载失败处理 ----------
const handleImageError = (e: Event, _element: ImageItem) => {
  const target = e.target as HTMLImageElement;
  target.src = props.defaultImage;
};

// ---------- 辅助函数：获取文件名/标识 ----------
const getFileName = (item: ImageItem): string => {
  if (item.file) return item.file.name;
  return `现有图片 #${item.id || ''}`;
};

// ---------- 辅助函数：获取文件大小显示 ----------
const getFileSize = (item: ImageItem): string => {
  if (item.file) {
    const bytes = item.file.size;
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
  return '已存在';
};

// ---------- 组件卸载时清理所有新图片的blob URL ----------
onUnmounted(() => {
  innerImageList.value.forEach((item) => {
    if (item.status === 'new' && item.url.startsWith('blob:')) {
      URL.revokeObjectURL(item.url);
    }
  });
});

// 暴露清理方法（可选）
defineExpose({
  clearBlobs: () => {
    innerImageList.value.forEach((item) => {
      if (item.status === 'new' && item.url.startsWith('blob:')) {
        URL.revokeObjectURL(item.url);
      }
    });
  },
});
</script>

<style scoped>
/* ===== 样式完全继承原创建/编辑表单设计语言 ===== */
.image-manager {
  width: 100%;
}

/* 上传区域 */
.upload-area {
  border: 2px dashed #e0e0e0;
  border-radius: 12px;
  background-color: #fafafa;
  padding: 32px 24px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-area:hover {
  border-color: #4caf50;
  background-color: #f5fff5;
}

.hidden-input {
  display: none;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #999;
}

.upload-placeholder svg {
  color: #bbb;
  margin-bottom: 4px;
}

.upload-placeholder span {
  font-size: 15px;
}

.upload-hint {
  font-size: 12px !important;
  color: #aaa !important;
}

/* 图片预览网格 */
.products-grid.preview-grid {
  display: grid;
  gap: 16px;
  margin-top: 16px;
}

@media (min-width: 640px) {
  .products-grid.preview-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .products-grid.preview-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

.product-card.preview-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background-color: white;
}

.product-card.preview-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08);
}

.product-image-container.preview-image {
  height: 140px;
  background-color: #f8f8f8;
  overflow: hidden;
  flex-shrink: 0;
}

.product-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.preview-info {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: white;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.preview-filename {
  font-size: 13px;
  color: #333;
  word-break: break-all;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
}

.preview-size {
  font-size: 12px;
  color: #999;
}

/* 拖拽手柄 */
.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  color: #999;
  padding: 6px;
  border-radius: 4px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.drag-handle:hover {
  background-color: #f0f0f0;
  color: #666;
}

.drag-handle:active {
  cursor: grabbing;
}

/* 拖拽幽灵效果 */
.sortable-ghost {
  opacity: 0.4;
  background-color: #f0f0f0;
  border: 1px dashed #4caf50;
}

.sortable-drag {
  opacity: 0.8;
  transform: rotate(2deg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

/* 删除按钮 */
.preview-remove {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.6);
  border: none;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  color: white;
  transition: all 0.2s ease;
  z-index: 5;
}

.preview-remove svg {
  height: 16px;
  width: 16px;
  position: absolute;
}

.preview-remove:hover {
  background-color: #e53935;
  transform: scale(1.1);
}

/* 上限提示 */
.upload-limit {
  font-size: 13px;
  color: #f57c00;
  background-color: rgba(245, 124, 0, 0.1);
  padding: 8px 12px;
  border-radius: 6px;
  margin-top: 8px;
}
</style>