<template>
  <div class="dish-container">
    <!-- 删除时的加载状态 -->
    <div v-if="deleting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在删除菜品...</p>
    </div>

    <!-- 主卡片：删除菜品 -->
    <div class="dish-main-card">
      <div class="dish-header">
        <h1 class="dish-title">删除菜品</h1>
        <div class="dish-code-container">
          <span class="dish-code">支持两种方式快速填充（只读查看）</span>
        </div>
      </div>

      <!-- ======= 方式1：通过菜品编码查询 ======= -->
      <div class="search-by-code-section">
        <div class="form-group" style="margin-bottom: 0">
          <label class="form-label">通过菜品编码查询</label>
          <div class="code-search-input-group">
            <input
              v-model="searchCode"
              type="text"
              class="form-control"
              placeholder="输入菜品编码"
              :disabled="isEditLocked"
            />
            <button
              type="button"
              class="btn btn-primary btn-sm"
              @click="handleSearchByCode"
              :disabled="!searchCode.trim() || isEditLocked"
            >
              查询
            </button>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              @click="resetForm"
              title="清空表单并解锁编码输入"
            >
              重置
            </button>
          </div>
          <div v-if="searchError" class="error-message">{{ searchError }}</div>
        </div>
      </div>

      <!-- ======= 方式2：从菜品列表选择（抽屉触发） ======= -->
      <div class="selector-section">
        <h3 class="selector-title">从现有菜品列表选择</h3>
        <div class="dish-picker-trigger">
          <button
            type="button"
            class="btn btn-outline select-dish-btn"
            @click="openDishPicker"
          >
            选择菜品
          </button>
          <span v-if="form.dishCode" class="selected-hint">
            已选择：{{ form.dishCode }} {{ form.name }}
          </span>
        </div>
      </div>

      <!-- ======= 只读表单：展示选定菜品的数据 ======= -->
      <form @submit.prevent="handleDelete" class="dish-form" v-if="isEditLocked">
        <!-- 菜品编码（只读） -->
        <div class="form-group">
          <label class="form-label">菜品编码</label>
          <input
            v-model="form.dishCode"
            type="text"
            class="form-control"
            disabled
          />
          <div class="hint-text">编码不可修改</div>
        </div>

        <!-- 菜品名称（只读） -->
        <div class="form-group">
          <label class="form-label">菜品名称</label>
          <input
            v-model="form.name"
            type="text"
            class="form-control"
            disabled
          />
        </div>

        <!-- 菜品描述（只读） -->
        <div class="form-group">
          <label class="form-label">菜品描述</label>
          <textarea
            v-model="form.description"
            class="form-control textarea"
            rows="4"
            disabled
          ></textarea>
        </div>

        <!-- 菜品做法（只读） -->
        <div class="form-group">
          <label class="form-label">菜品做法</label>
          <textarea
            v-model="form.recipe"
            class="form-control textarea"
            rows="8"
            disabled
          ></textarea>
        </div>

        <!-- 图片展示（只读） -->
        <div class="form-group">
          <label class="form-label">菜品图片</label>
          <div class="readonly-image-list">
            <div
              v-for="(img, index) in imageList"
              :key="img.id || img.url"
              class="readonly-image-item"
            >
              <img :src="img.url" :alt="`菜品图片${index + 1}`" class="readonly-image" />
            </div>
            <div v-if="!imageList.length" class="no-image-text">暂无图片</div>
          </div>
        </div>

        <!-- ========== 关联食材（只读展示） ========== -->
        <div class="form-group">
          <label class="form-label">关联食材</label>

          <!-- 已选食材列表（只读） -->
          <div v-if="selectedIngredients.length" class="ingredients-list readonly">
            <div
              v-for="(item, index) in selectedIngredients"
              :key="index"
              class="ingredient-item readonly"
            >
              <!-- 无删除按钮 -->
              <div class="ingredient-info">
                <span class="ingredient-name">{{ item.name }}</span>
                <span class="ingredient-code">{{ item.ingredientCode }}</span>
                <span class="ingredient-badge">{{ item.isExisting ? '已存在' : '新增' }}</span>
              </div>
              <div class="ingredient-fields readonly">
                <span class="readonly-field">
                  <span class="field-label">用量：</span>
                  {{ item.quantity || '未填写' }}
                </span>
                <span class="readonly-field">
                  <span class="field-label">备注：</span>
                  {{ item.note || '无' }}
                </span>
              </div>
            </div>
          </div>
          <div v-else class="no-ingredients-text">暂无关联食材</div>
          <!-- 无添加按钮 -->
        </div>

        <!-- 删除按钮区 -->
        <div class="form-actions">
          <router-link to="/dishes" class="btn btn-secondary">取消</router-link>
          <button
            type="submit"
            class="btn btn-danger"
            :disabled="deleting || !isDeleteValid"
          >
            {{ deleting ? "删除中..." : "删除菜品" }}
          </button>
        </div>
      </form>

      <!-- 未锁定编辑时显示提示 -->
      <div v-else class="empty-editor-prompt">
        请通过上方「编码查询」或「菜品列表」选择要删除的菜品
      </div>
    </div>

    <!-- 删除成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>菜品删除成功！</span>
      </div>
    </Teleport>

    <!-- ========== 菜品选择抽屉 ========== -->
    <Teleport to="body">
      <div v-if="showDishPicker" class="drawer-overlay" @click.self="closeDishPicker">
        <div class="drawer-container" :class="{ 'drawer-open': showDishPicker }">
          <div class="drawer-header">
            <h3>选择菜品</h3>
            <button class="drawer-close" @click="closeDishPicker">&times;</button>
          </div>
          <div class="drawer-body">
            <ScrollPicker
              ref="dishScrollPickerRef"
              :fetch="fetchDishes"
              :limit="15"
              @select="handleDishSelected"
            >
              <template #item="{ item }">
                <div class="ingredient-card">
                  <div class="ingredient-card-name">{{ item.name }}</div>
                  <div class="ingredient-card-code">{{ item.dishCode }}</div>
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
import { ref, reactive, computed, watch } from 'vue';
//import { useRouter } from 'vue-router';
import request from '../../../api/request';
import ScrollPicker from '../../../components/picker/ScrollPicker.vue';
import type { FetchResult } from '../../../components/picker/ScrollPicker.vue';
import type { ImageItem } from '../../../components/image/ImageManager.vue'; // 仅类型
import type { ImageResponse, ViewDishIngredientCardListWithCursor, ViewDishCard, ViewDishCardListWithCursor } from '../../../types/types';

// ---------- 类型定义 ----------
interface ViewDishResponse {
  dishCode: string;
  name: string;
  description: string;
  recipe: string;
  images: ImageResponse[];
}

interface SelectedIngredient {
  ingredientCode: string;
  name: string;
  quantity: string;
  note: string;
  isExisting: boolean;
}

// ---------- 表单状态 ----------
const form = reactive({
  dishCode: '',
  name: '',
  description: '',
  recipe: '',
});

const imageList = ref<ImageItem[]>([]);
const selectedIngredients = ref<SelectedIngredient[]>([]);

const deleting = ref(false);
const showSuccessToast = ref(false);
const isEditLocked = ref(false);
const searchCode = ref('');
const searchError = ref('');

const showDishPicker = ref(false);
const dishScrollPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);

//const router = useRouter();

// ---------- 删除按钮有效性：只需菜品编码存在 ----------
const isDeleteValid = computed(() => form.dishCode.trim() !== '');

// ---------- 数据获取函数 ----------
const fetchDishes = async (cursor: number, limit: number): Promise<FetchResult<ViewDishCard>> => {
  const res = await request({
    url: '/api/dishes',
    method: 'GET',
    params: { cursor, limit },
  });
  const data: ViewDishCardListWithCursor = res.data;
  return {
    items: data.dishes || [],
    cursor: data.cursor || 0,
    hasMore: data.hasMore || false,
  };
};

// ---------- 打开/关闭菜品抽屉 ----------
const openDishPicker = () => {
  showDishPicker.value = true;
  setTimeout(() => dishScrollPickerRef.value?.refresh(), 50);
};
const closeDishPicker = () => {
  showDishPicker.value = false;
};

// ---------- 处理菜品选择 ----------
const handleDishSelected = (item: ViewDishCard) => {
  fetchDishDetail(item.dishCode);
  closeDishPicker();
};

// ---------- 通过编码查询 ----------
const handleSearchByCode = () => {
  if (!searchCode.value.trim()) return;
  fetchDishDetail(searchCode.value.trim());
};

// ---------- 获取菜品详情（填充只读数据）----------
const fetchDishDetail = async (code: string) => {
  try {
    searchError.value = '';

    const [dishRes, dishIngredientsRes] = await Promise.all([
      request({ url: `/api/dishes/${code}`, method: 'GET' }),
      request({
        url: `/api/dishes/${code}/ingredients`,
        method: 'GET',
        params: { cursor: 0, limit: 15 },
      }),
    ]);

    const dishData: ViewDishResponse = dishRes.data;

    form.dishCode = dishData.dishCode;
    form.name = dishData.name;
    form.description = dishData.description || '';
    form.recipe = dishData.recipe || '';

    // 填充图片列表（只读）
    const existingImages: ImageItem[] = (dishData.images || []).map((img: any) => ({
      id: img.id,
      url: img.imageURL,
      status: 'existing',
      sortOrder: img.sortOrder,
    }));
    existingImages.sort((a, b) => a.sortOrder - b.sortOrder);
    imageList.value = existingImages;

    const dishIngredientsData: ViewDishIngredientCardListWithCursor = dishIngredientsRes.data;

    // 填充食材列表（只读）
    const ingredients: SelectedIngredient[] = (dishIngredientsData.dishIngredients || []).map((ing: any) => ({
      ingredientCode: ing.ingredientCode,
      name: ing.name,
      quantity: ing.quantity || '',
      note: ing.note || '',
      isExisting: true,
    }));
    selectedIngredients.value = ingredients;

    isEditLocked.value = true;
    searchCode.value = '';
  } catch (error: any) {
    console.error('查询菜品失败:', error);
    if (error.response?.status === 404) {
      searchError.value = '未找到该编码的菜品，请确认';
    } else {
      searchError.value = '查询失败，请稍后重试';
    }
  }
};

// ---------- 重置表单 ----------
const resetForm = () => {
  form.dishCode = '';
  form.name = '';
  form.description = '';
  form.recipe = '';

  imageList.value = [];
  selectedIngredients.value = [];
  isEditLocked.value = false;
  searchError.value = '';
  searchCode.value = '';
};

// ---------- 删除处理 ----------
const handleDelete = async () => {
  if (!isDeleteValid.value) return;
  if (!form.dishCode) {
    alert('菜品编码异常，请重新选择菜品');
    return;
  }

  if (!confirm(`确定要删除菜品“${form.name}（编码：${form.dishCode}）”吗？此操作不可恢复。`)) {
    return;
  }

  deleting.value = true;

  try {
    await request({
      url: `/api/dishes/${form.dishCode}`,
      method: 'DELETE',
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
      //router.push('/dishes');
    }, 1500);
  } catch (error) {
    console.error('删除菜品失败:', error);
    alert('删除失败，请稍后重试');
  } finally {
    deleting.value = false;
  }
};

// 监听锁定状态，清空搜索输入
watch(isEditLocked, (locked) => {
  if (locked) searchCode.value = '';
});
</script>

<style scoped>
/* ========== 完全继承原编辑页面的样式，并添加只读图片列表和只读食材样式 ========== */
.dish-container {
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
  border-top-color: #f44336; /* 红色主题 */
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 主卡片 */
.dish-main-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  display: flex;
  flex-direction: column;
  padding: 24px;
}
.dish-header {
  margin-bottom: 24px;
  flex-shrink: 0;
}
.dish-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}
.dish-code-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.dish-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
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
.hint-text {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
.error-message {
  font-size: 13px;
  color: #e53935;
  margin-top: 4px;
}

/* 选择卡片区 */
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
.dish-picker-trigger {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.select-dish-btn {
  padding: 12px 24px;
  font-size: 15px;
  border-radius: 30px;
  min-width: 140px;
}
.selected-hint {
  font-size: 14px;
  color: #f44336; /* 红色提示 */
  background-color: #ffebee;
  padding: 6px 16px;
  border-radius: 30px;
}

/* 空编辑器提示 */
.empty-editor-prompt {
  padding: 60px 24px;
  text-align: center;
  color: #aaa;
  background: #fafafa;
  border-radius: 16px;
  font-size: 16px;
  border: 1px dashed #ddd;
  margin-top: 16px;
}

/* 表单样式 */
.dish-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
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
  border-color: #f44336;
  box-shadow: 0 0 0 3px rgba(244, 67, 54, 0.1);
}
.form-control.textarea {
  resize: vertical;
  min-height: 100px;
  line-height: 1.6;
}
.form-control:disabled,
textarea:disabled {
  background-color: #f5f5f5;
  color: #666;
  border-color: #e0e0e0;
  cursor: not-allowed;
  opacity: 0.8;
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
.btn-danger {
  background-color: #f44336;
  color: white;
  box-shadow: 0 4px 12px rgba(244, 67, 54, 0.3);
}
.btn-danger:hover:not(:disabled) {
  background-color: #e53935;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(244, 67, 54, 0.4);
}
.btn-danger:disabled {
  background-color: #ef9a9a;
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
.btn-outline {
  background: white;
  border: 1px solid #f44336;
  color: #f44336;
}
.btn-outline:hover {
  background-color: #f44336;
  color: white;
  border-color: #f44336;
}
.btn-sm {
  padding: 8px 20px;
  font-size: 14px;
  border-radius: 30px;
  white-space: nowrap;
  height: 44px;
  line-height: 1;
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

/* 食材列表（只读） */
.ingredients-list.readonly {
  margin-top: 8px;
}
.ingredient-item.readonly {
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 12px;
  background-color: #fafafa;
  margin-bottom: 8px;
}
.ingredient-info {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.ingredient-name {
  font-weight: 600;
  color: #333;
}
.ingredient-code {
  font-size: 12px;
  color: #999;
}
.ingredient-badge {
  font-size: 11px;
  background-color: #e0e0e0;
  color: #666;
  padding: 2px 6px;
  border-radius: 12px;
}
.ingredient-fields.readonly {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  color: #666;
  font-size: 14px;
}
.readonly-field {
  background-color: #f0f0f0;
  padding: 4px 12px;
  border-radius: 16px;
}
.field-label {
  color: #999;
  margin-right: 4px;
}
.no-ingredients-text {
  padding: 16px;
  text-align: center;
  color: #999;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px dashed #ddd;
}

/* 抽屉样式 */
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
  box-sizing: border-box;
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
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.drawer-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}
.drawer-close {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  color: #999;
  padding: 4px 8px;
}
.drawer-close:hover {
  color: #333;
}
.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.ingredient-card {
  padding: 12px 16px;
  border: 1px solid #f0f0f0;
  border-radius: 10px;
  background-color: white;
  transition: all 0.2s ease;
}
.ingredient-card:hover {
  border-color: #f44336;
  box-shadow: 0 2px 8px rgba(244, 67, 54, 0.1);
}
.ingredient-card-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}
.ingredient-card-code {
  font-size: 12px;
  color: #999;
}

/* 提示toast */
.toast-message {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 30px;
  background-color: #f44336;
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

/* 响应式布局 */
@media (max-width: 639px) {
  .dish-container {
    padding: 12px;
  }
  .dish-main-card {
    padding: 20px 16px;
  }
  .dish-title {
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
  .dish-picker-trigger {
    flex-direction: column;
    align-items: stretch;
  }
  .select-dish-btn {
    width: 100%;
  }
  .selected-hint {
    text-align: center;
  }
  .drawer-container {
    height: 80vh;
  }
}
@media (min-width: 768px) {
  .dish-main-card {
    padding: 32px;
  }
  .dish-title {
    font-size: 28px;
  }
  .form-group {
    gap: 12px;
  }
}
@media (min-width: 1024px) {
  .dish-container {
    padding: 24px;
  }
  .dish-title {
    font-size: 32px;
  }
}
@media (min-width: 1400px) {
  .dish-container {
    max-width: 1400px;
    padding: 32px;
  }
  .dish-main-card {
    padding: 40px;
  }
  .dish-title {
    font-size: 36px;
  }
}
@media print {
  .dish-container {
    padding: 0;
    background-color: white;
  }
  .dish-main-card {
    box-shadow: none;
    border: 1px solid #ddd;
  }
  .form-actions,
  .drawer-overlay {
    display: none;
  }
}
</style>