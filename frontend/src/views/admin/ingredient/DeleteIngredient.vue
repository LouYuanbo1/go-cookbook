<template>
  <div class="ingredient-container">
    <!-- 删除时的加载状态 -->
    <div v-if="deleting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在删除食材...</p>
    </div>

    <!-- 主卡片：删除食材 -->
    <div class="ingredient-main-card">
      <div class="ingredient-header">
        <h1 class="ingredient-title">删除食材</h1>
        <div class="ingredient-code-container">
          <span class="ingredient-code">支持两种方式快速定位食材</span>
        </div>
      </div>

      <!-- ======= 方式1：直接查询 ======= -->
      <div class="search-by-code-section">
        <div class="form-group" style="margin-bottom: 0">
          <label class="form-label">通过食材编码查询</label>
          <div class="code-search-input-group">
            <input
              v-model="searchCode"
              type="text"
              class="form-control"
              placeholder="输入食材编码"
              :disabled="isDetailLoaded"
            />
            <button
              type="button"
              class="btn btn-primary btn-sm"
              @click="handleSearchByCode"
              :disabled="!searchCode.trim() || isDetailLoaded"
            >
              查询
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              @click="resetForm"
              title="清空表单并重新查询"
            >
              重置
            </button>
          </div>
          <div v-if="searchError" class="error-message">{{ searchError }}</div>
        </div>
      </div>

      <!-- ======= 方式2：从食材列表选择（抽屉触发） ======= -->
      <div class="selector-section">
        <h3 class="selector-title">从现有食材列表选择</h3>
        <div class="ingredient-picker-trigger">
          <button
            type="button"
            class="btn btn-outline select-ingredient-btn"
            @click="openIngredientPicker"
            :disabled="isDetailLoaded"
          >
            选择食材
          </button>
          <span v-if="form.ingredientCode" class="selected-hint">
            已选择：{{ form.ingredientCode }} {{ form.name }}
          </span>
        </div>
      </div>

      <!-- ======= 只读详情表单：基于查询/选择填充后的数据 ======= -->
      <div v-if="isDetailLoaded" class="ingredient-detail">
        <!-- 食材编码（只读） -->
        <div class="form-group">
          <label class="form-label">食材编码</label>
          <div class="readonly-text">{{ form.ingredientCode }}</div>
        </div>

        <!-- 食材名称 -->
        <div class="form-group">
          <label class="form-label">食材名称</label>
          <div class="readonly-text">{{ form.name }}</div>
        </div>

        <!-- 食材描述 -->
        <div class="form-group">
          <label class="form-label">食材描述</label>
          <div class="readonly-text description">{{ form.description || '无描述' }}</div>
        </div>

        <!-- 图片展示（只读） -->
        <div class="form-group">
          <label class="form-label">食材图片</label>
          <div class="readonly-image-list">
            <div
              v-for="(img, index) in imageList"
              :key="img.id || img.url"
              class="readonly-image-item"
            >
              <img :src="img.url" :alt="`食材图片${index + 1}`" class="readonly-image" />
            </div>
            <div v-if="!imageList.length" class="no-image-text">暂无图片</div>
          </div>
        </div>

        <!-- 删除按钮区 -->
        <div class="form-actions">
          <router-link to="/ingredients" class="btn btn-secondary">取消</router-link>
          <button
            type="button"
            class="btn btn-danger"
            @click="handleDelete"
            :disabled="deleting || !form.ingredientCode"
          >
            {{ deleting ? "删除中..." : "确认删除" }}
          </button>
        </div>
      </div>

      <!-- 未加载详情时显示提示 -->
      <div v-else class="empty-editor-prompt">
        请通过上方「编码查询」或「食材列表」选择要删除的食材
      </div>
    </div>

    <!-- 操作成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>食材删除成功！</span>
      </div>
    </Teleport>

    <!-- ========== 食材选择抽屉（完全复用原样式与逻辑） ========== -->
    <Teleport to="body">
      <div
        v-if="showIngredientPicker"
        class="drawer-overlay"
        @click.self="closeIngredientPicker"
      >
        <div class="drawer-container" :class="{ 'drawer-open': showIngredientPicker }">
          <div class="drawer-header">
            <h3>选择食材</h3>
            <button class="drawer-close" @click="closeIngredientPicker">&times;</button>
          </div>
          <div class="drawer-body">
            <ScrollPicker
              ref="scrollPickerRef"
              :fetch="fetchIngredients"
              :limit="15"
              @select="handleIngredientSelected"
            >
              <template #item="{ item }">
                <div class="ingredient-item">
                  <span class="ingredient-code">{{ item.ingredientCode }}</span>
                  <span class="ingredient-name">{{ item.name }}</span>
                </div>
              </template>
            </ScrollPicker>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
//import { useRouter } from 'vue-router';
import request from '../../../api/request';
import ScrollPicker from '../../../components/picker/ScrollPicker.vue';
import type { FetchResult } from '../../../components/picker/ScrollPicker.vue';
import type { ImageResponse, ViewIngredientCard, ViewIngredientCardListWithCursor } from '../../../types/types';

//const router = useRouter();

// ---------- 表单状态 ----------
const form = reactive({
  ingredientCode: '',
  name: '',
  description: '',
});

// 图片列表（仅用于展示）
const imageList = ref<{ id?: number; url: string }[]>([]);

const deleting = ref(false);
const showSuccessToast = ref(false);
const isDetailLoaded = ref(false); // 控制详情区域显示
const searchCode = ref('');
const searchError = ref('');

// ---------- 抽屉状态 ----------
const showIngredientPicker = ref(false);
// const scrollPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);

// ---------- ScrollPicker 数据获取 ----------
const fetchIngredients = async (cursor: number, limit: number): Promise<FetchResult<ViewIngredientCard>> => {
  const res = await request({
    url: '/api/ingredients',
    method: 'GET',
    params: { cursor, limit },
  });
  const data: ViewIngredientCardListWithCursor = res.data;
  return {
    items: data.ingredients || [],
    cursor: data.cursor || 0,
    hasMore: data.hasMore || false,
  };
};

// ---------- 打开/关闭抽屉 ----------
const openIngredientPicker = () => {
  showIngredientPicker.value = true;
};
const closeIngredientPicker = () => {
  showIngredientPicker.value = false;
};

// ---------- 选中食材（抽屉回调）----------
const handleIngredientSelected = (item: ViewIngredientCard) => {
  fetchIngredientDetail(item.ingredientCode);
  closeIngredientPicker();
};

// ---------- 查询食材详情（复用，同时用于编码查询和抽屉选中）----------
const fetchIngredientDetail = async (code: string) => {
  try {
    searchError.value = '';
    const res = await request({
      url: `/api/ingredients/${code}`,
      method: 'GET',
    });
    const data = res.data;

    // 填充表单
    form.ingredientCode = data.ingredientCode;
    form.name = data.name;
    form.description = data.description || '';

    // 构建图片列表（仅用于展示）
    const images: { id?: number; url: string }[] = (data.images || []).map((img: ImageResponse) => ({
      id: img.id,
      url: img.imageURL,
    }));
    // 按 sortOrder 排序（如果有）
    if (data.images) {
      images.sort((a, b) => (a.id || 0) - (b.id || 0));
    }
    imageList.value = images;

    isDetailLoaded.value = true;
    searchCode.value = ''; // 清空查询输入
  } catch (error: any) {
    console.error('查询食材失败:', error);
    if (error.response?.status === 404) {
      searchError.value = '未找到该编码的食材，请确认';
    } else {
      searchError.value = '查询失败，请稍后重试';
    }
  }
};

// ---------- 编码查询按钮 ----------
const handleSearchByCode = () => {
  if (!searchCode.value.trim()) return;
  fetchIngredientDetail(searchCode.value.trim());
};

// ---------- 重置表单 ----------
const resetForm = () => {
  form.ingredientCode = '';
  form.name = '';
  form.description = '';
  imageList.value = [];
  isDetailLoaded.value = false;
  searchError.value = '';
  searchCode.value = '';
};

// ---------- 删除操作 ----------
const handleDelete = async () => {
  if (!form.ingredientCode) {
    alert('请先选择要删除的食材');
    return;
  }

  const confirmDelete = confirm(`确定要删除食材“${form.name}”吗？此操作不可恢复。`);
  if (!confirmDelete) return;

  deleting.value = true;

  try {
    await request({
      url: `/api/ingredients/${form.ingredientCode}`,
      method: 'DELETE',
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
      //router.push('/ingredients'); // 删除后跳转到列表页
    }, 1500);
  } catch (error) {
    console.error('删除食材失败:', error);
    alert('删除失败，请稍后重试');
  } finally {
    deleting.value = false;
  }
};
</script>

<style scoped>
/* ========== 完全继承原表单的精美样式，并补充只读展示样式 ========== */
.ingredient-container {
  font-family: "PingFang SC", "Helvetica Neue", Arial, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
  color: #333;
  background-color: #f9f9f9;
  min-height: 100vh;
  box-sizing: border-box;
}

/* 加载状态 */
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
  border-top-color: #e53935;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 主卡片 */
.ingredient-main-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.ingredient-header {
  margin-bottom: 24px;
  flex-shrink: 0;
}

.ingredient-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.ingredient-code-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ingredient-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

/* 表单样式（只读区域） */
.ingredient-detail {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-top: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  display: flex;
  align-items: center;
  gap: 4px;
}

.readonly-text {
  padding: 12px 16px;
  font-size: 15px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background-color: #f9f9f9;
  color: #333;
  line-height: 1.5;
  word-break: break-word;
}

.readonly-text.description {
  min-height: 100px;
  white-space: pre-wrap;
}

/* 只读图片列表 */
.readonly-image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 8px;
}

.readonly-image-item {
  width: 100px;
  height: 100px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e0e0e0;
  background-color: #f5f5f5;
}

.readonly-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.no-image-text {
  padding: 20px;
  text-align: center;
  color: #999;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px dashed #ddd;
  width: 100%;
}

/* 按钮样式 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 16px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 24px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 30px;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  text-decoration: none;
  line-height: 1.5;
}

.btn-primary {
  background-color: #4caf50;
  color: white;
  box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
}

.btn-primary:hover:not(:disabled) {
  background-color: #43a047;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(76, 175, 80, 0.4);
}

.btn-primary:disabled {
  background-color: #a5d6a7;
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

.btn-secondary {
  background-color: white;
  color: #666;
  border: 1px solid #ddd;
}

.btn-secondary:hover {
  background-color: #f5f5f5;
  border-color: #ccc;
  color: #333;
}

.btn-danger {
  background-color: #e53935;
  color: white;
  box-shadow: 0 4px 12px rgba(229, 57, 53, 0.3);
}

.btn-danger:hover:not(:disabled) {
  background-color: #d32f2f;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(229, 57, 53, 0.4);
}

.btn-danger:disabled {
  background-color: #ef9a9a;
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

.btn-outline {
  background: white;
  border: 1px solid #4caf50;
  color: #4caf50;
}

.btn-outline:hover:not(:disabled) {
  background-color: #4caf50;
  color: white;
  border-color: #4caf50;
}

.btn-sm {
  padding: 8px 20px;
  font-size: 14px;
  border-radius: 30px;
  white-space: nowrap;
  height: 44px;
  line-height: 1;
}

/* 提示toast */
.toast-message {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 30px;
  background-color: #4caf50;
  color: white;
  font-size: 15px;
  font-weight: 500;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translate(-50%, -20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

/* 编码查询区域 */
.search-by-code-section {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.code-search-input-group {
  display: flex;
  gap: 12px;
  align-items: center;
}

.code-search-input-group .form-control {
  flex: 1;
  margin-bottom: 0;
}

.form-control {
  width: 100%;
  padding: 12px 16px;
  font-size: 15px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.3s ease;
  background-color: #fff;
  box-sizing: border-box;
  font-family: inherit;
}

.form-control:focus {
  outline: none;
  border-color: #4caf50;
  box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}

.form-control:disabled {
  background-color: #f9f9f9;
  color: #666;
  border-color: #e0e0e0;
  cursor: not-allowed;
}

.error-message {
  font-size: 13px;
  color: #e53935;
  margin-top: 4px;
}

/* 提示文本 */
.hint-text {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 食材选择卡片区 */
.selector-section {
  margin-bottom: 28px;
  background: #fafafa;
  border-radius: 16px;
  padding: 20px;
  border: 1px solid #f0f0f0;
}

.selector-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #444;
  display: flex;
  align-items: center;
  gap: 6px;
}

.ingredient-picker-trigger {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.select-ingredient-btn {
  padding: 12px 24px;
  font-size: 15px;
  border-radius: 30px;
  min-width: 140px;
}

.select-ingredient-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.selected-hint {
  font-size: 14px;
  color: #4caf50;
  background-color: #e8f5e9;
  padding: 6px 16px;
  border-radius: 30px;
}

/* 空详情提示 */
.empty-editor-prompt {
  padding: 60px 24px;
  text-align: center;
  color: #aaa;
  background: #fafafa;
  border-radius: 16px;
  font-size: 16px;
  border: 1px dashed #ddd;
  margin-top: 16px;
  transition: all 0.2s;
}

/* 抽屉样式（完全复用原样式） */
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.4);
  z-index: 10000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  transition: background-color 0.3s ease;
}

.drawer-container {
  background-color: white;
  width: 100%;
  max-width: 600px;
  height: 70vh;
  border-top-left-radius: 20px;
  border-top-right-radius: 20px;
  display: flex;
  flex-direction: column;
  transform: translateY(100%);
  transition: transform 0.3s ease-out;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.1);
}

.drawer-container.drawer-open {
  transform: translateY(0);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.drawer-header h3 {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  color: #222;
}

.drawer-close {
  background: none;
  border: none;
  font-size: 28px;
  line-height: 1;
  color: #999;
  cursor: pointer;
  padding: 0 8px;
}

.drawer-close:hover {
  color: #e53935;
}

.drawer-body {
  flex: 1;
  overflow: hidden;
  padding: 16px 16px 0;
}

.ingredient-item {
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  background-color: #fafafa;
  border-radius: 12px;
  transition: background-color 0.2s;
  cursor: pointer;
}

.ingredient-item:hover {
  background-color: #f0f9f0;
}

.ingredient-code {
  font-size: 14px;
  font-weight: 600;
  color: #4caf50;
  margin-bottom: 4px;
}

.ingredient-name {
  font-size: 15px;
  color: #333;
}

/* 响应式布局（与原表单一致） */
@media (max-width: 639px) {
  .ingredient-container {
    padding: 12px;
  }

  .ingredient-main-card {
    padding: 20px 16px;
  }

  .ingredient-title {
    font-size: 22px;
  }

  .form-actions {
    flex-direction: column-reverse;
    gap: 12px;
  }

  .btn {
    width: 100%;
  }

  .code-search-input-group {
    flex-wrap: wrap;
  }

  .code-search-input-group .btn {
    flex: 1;
  }

  .ingredient-picker-trigger {
    flex-direction: column;
    align-items: stretch;
  }

  .select-ingredient-btn {
    width: 100%;
  }

  .selected-hint {
    text-align: center;
  }

  .drawer-container {
    height: 80vh;
    border-radius: 20px 20px 0 0;
  }

  .readonly-image-item {
    width: 80px;
    height: 80px;
  }
}

@media (min-width: 768px) {
  .ingredient-main-card {
    padding: 32px;
  }

  .ingredient-title {
    font-size: 28px;
  }

  .form-group {
    gap: 12px;
  }

  .drawer-container {
    width: 600px;
    border-radius: 20px 20px 0 0;
  }
}

@media (min-width: 1024px) {
  .ingredient-container {
    padding: 24px;
  }

  .ingredient-title {
    font-size: 32px;
  }
}

@media (min-width: 1400px) {
  .ingredient-container {
    max-width: 1400px;
    padding: 32px;
  }

  .ingredient-main-card {
    padding: 40px;
  }

  .ingredient-title {
    font-size: 36px;
  }
}

@media print {
  .ingredient-container {
    padding: 0;
    background-color: white;
  }

  .ingredient-main-card {
    box-shadow: none;
    border: 1px solid #ddd;
  }

  .form-actions,
  .drawer-overlay {
    display: none;
  }
}
</style>