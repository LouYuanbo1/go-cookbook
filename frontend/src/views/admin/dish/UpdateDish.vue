<template>
  <div class="dish-container">
    <!-- 提交时的加载状态 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在更新菜品信息...</p>
    </div>

    <!-- 主卡片：编辑菜品 -->
    <div class="dish-main-card">
      <div class="dish-header">
        <h1 class="dish-title">编辑菜品</h1>
        <div class="dish-code-container">
          <span class="dish-code">支持两种方式快速填充</span>
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

      <!-- ======= 编辑表单：基于查询/选择填充后的数据 ======= -->
      <form @submit.prevent="handleSubmit" class="dish-form" v-if="isEditLocked">
        <!-- 菜品编码（只读） -->
        <div class="form-group">
          <label class="form-label">
            菜品编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.dishCode"
            type="text"
            class="form-control"
            disabled
          />
          <div class="hint-text">编码不可修改，如需编辑其他菜品请点击「重置」</div>
        </div>

        <!-- 菜品名称 -->
        <div class="form-group">
          <label class="form-label">
            菜品名称 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            class="form-control"
            placeholder="例如: 牛肉面"
            required
          />
          <div v-if="formErrors.name" class="error-message">{{ formErrors.name }}</div>
        </div>

        <!-- 菜品描述 -->
        <div class="form-group">
          <label class="form-label">菜品描述</label>
          <textarea
            v-model="form.description"
            class="form-control textarea"
            rows="4"
            placeholder="详细描述菜品的风味、做法、特点等"
          ></textarea>
        </div>

        <!-- 菜品做法 -->
        <div class="form-group">
          <label class="form-label">
            菜品做法
          </label>
          <textarea
            v-model="form.recipe"
            class="form-control textarea"
            rows="8"
            placeholder="详细描述菜品的做法，包括食材准备、烹饪步骤、注意事项等"
          ></textarea>
        </div>

        <!-- ========== 图片管理组件 ========== -->
        <div class="form-group">
          <label class="form-label">
            菜品图片
            <span class="image-hint">(最多9张，支持拖拽排序)</span>
          </label>
          <ImageManager
            v-model:modelValue="imageList"
            @update:deletedIds="deletedImageIds = $event"
            :max-count="9"
            upload-text="点击添加新图片"
            upload-hint="支持 jpg、png、webp"
          />
        </div>

        <!-- ========== 关联食材 ========== -->
        <div class="form-group">
          <label class="form-label">
            关联食材
            <span class="image-hint">(可添加多个)</span>
          </label>

          <!-- 已选食材列表 -->
          <div v-if="selectedIngredients.length" class="ingredients-list">
            <div
              v-for="(item, index) in selectedIngredients"
              :key="index"
              class="ingredient-item"
            >
              <!-- 删除按钮移至右上角 -->
              <button
                type="button"
                class="btn-icon remove-btn"
                @click="removeIngredient(index)"
                aria-label="删除"
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

              <div class="ingredient-info">
                <span class="ingredient-name">{{ item.name }}</span>
                <span class="ingredient-code">{{ item.ingredientCode }}</span>
                <span v-if="item.isExisting" class="ingredient-badge">已存在</span>
                <span v-else class="ingredient-badge new">新增</span>
              </div>
              <div class="ingredient-fields">
                <input
                  v-model="item.quantity"
                  type="text"
                  class="form-control small"
                  placeholder="用量 (如: 200g)"
                />
                <input
                  v-model="item.note"
                  type="text"
                  class="form-control small"
                  placeholder="备注 (可选)"
                />
              </div>
            </div>
          </div>

          <!-- 添加食材按钮 -->
          <button
            type="button"
            class="btn btn-secondary add-ingredient-btn"
            @click="openIngredientPicker = true"
          >
            + 添加食材
          </button>
        </div>

        <!-- 表单操作按钮 -->
        <div class="form-actions">
          <router-link to="/dishes" class="btn btn-secondary">
            取消
          </router-link>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !isFormValid"
          >
            {{ submitting ? "更新中..." : "保存修改" }}
          </button>
        </div>
      </form>

      <!-- 未锁定编辑时显示提示 -->
      <div v-else class="empty-editor-prompt">
        请通过上方「编码查询」或「菜品列表」选择要编辑的菜品
      </div>
    </div>

    <!-- 提交成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>菜品更新成功！</span>
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

    <!-- 食材选择弹窗 -->
    <Teleport to="body">
      <div v-if="openIngredientPicker" class="modal-overlay" @click.self="openIngredientPicker = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>选择食材</h3>
            <button class="close-btn" @click="openIngredientPicker = false">&times;</button>
          </div>
          <div class="modal-body">
            <ScrollPicker
              ref="ingredientScrollPickerRef"
              :fetch="fetchIngredients"
              :limit="15"
              @select="handleIngredientSelect"
            >
              <template #item="{ item }">
                <div class="ingredient-card">
                  <div class="ingredient-card-name">{{ item.name }}</div>
                  <div class="ingredient-card-code">{{ item.ingredientCode }}</div>
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
import ImageManager from '../../../components/image/ImageManager.vue';
import ScrollPicker from '../../../components/picker/ScrollPicker.vue';
import type { ImageItem } from '../../../components/image/ImageManager.vue';
import type { FetchResult } from '../../../components/picker/ScrollPicker.vue';
import type { ImageResponse, ViewIngredientCard, ViewIngredientCardListWithCursor,ViewDishIngredientCard, ViewDishIngredientCardListWithCursor,ViewDishCard, ViewDishCardListWithCursor } from '../../../types/types';
import { v7 as uuidv7 } from 'uuid';

// ---------- 类型定义 ----------
interface ViewDishResponse {
  dishCode: string;
  name: string;
  description: string;
  recipe: string;
  images: ImageResponse[];
}

interface UpdateDishForm {
  dishCode: string;
  name: string;
  description: string;
  recipe: string;
}

interface SelectedIngredient {
  ingredientCode: string;
  name: string;
  quantity: string;
  note: string;
  isExisting: boolean;   // 是否为原有食材（来自后端）
}

// ---------- 表单状态 ----------
const form = reactive<UpdateDishForm>({
  dishCode: '',
  name: '',
  description: '',
  recipe: '',
});

// 图片列表
const imageList = ref<ImageItem[]>([]);
const deletedImageIds = ref<Set<number>>(new Set());

// 选中的食材列表
const selectedIngredients = ref<SelectedIngredient[]>([]);
// 记录被删除的原有食材编码（用于提交 deleted 类型）
const deletedIngredientCodes = ref<Set<string>>(new Set());

const submitting = ref(false);
const showSuccessToast = ref(false);
const isEditLocked = ref(false);
const searchCode = ref('');
const searchError = ref('');
//const router = useRouter();

// 控制抽屉/弹窗
const showDishPicker = ref(false);
const openIngredientPicker = ref(false);

// ---------- 表单校验 ----------
const formErrors = reactive({
  name: '',
});

watch(() => form.name, (val) => {
  formErrors.name = val.trim() ? '' : '菜品名称不能为空';
}, { immediate: true });

const isFormValid = computed(() => {
  return form.dishCode.trim() !== '' && form.name.trim() !== '' && imageList.value.length <= 9;
});

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

// ---------- 打开/关闭菜品抽屉 ----------
const openDishPicker = () => {
  showDishPicker.value = true;
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

// ---------- 获取菜品详情（核心填充逻辑）----------
const fetchDishDetail = async (code: string) => {
  try {
    searchError.value = '';

    const [dishRes, dishIngredientsRes] = await Promise.all([
      request({
        url: '/api/dishes/' + code,
        method: 'GET',
      }),
      request({
        url: '/api/dishes/' + code + '/ingredients',
        method: 'GET',
        params: {
          cursor: 0,
          limit: 15,
        },
      }),
    ]);

    const dishData: ViewDishResponse = dishRes.data;

    // 填充基本信息
    form.dishCode = dishData.dishCode;
    form.name = dishData.name;
    form.description = dishData.description || '';
    form.recipe = dishData.recipe || '';

    // 填充图片列表
    const existingImages: ImageItem[] = (dishData.images || []).map((img: any) => ({
      id: img.id,
      url: img.imageURL,
      status: 'existing',
      sortOrder: img.sortOrder,
    }));
    existingImages.sort((a, b) => a.sortOrder - b.sortOrder);
    imageList.value = existingImages;
    deletedImageIds.value.clear();

    const dishIngredientsData: ViewDishIngredientCardListWithCursor = dishIngredientsRes.data;

    // 填充食材列表
    const ingredients: SelectedIngredient[] = (dishIngredientsData.dishIngredients || []).map((ing: any) => ({
      ingredientCode: ing.ingredientCode,
      name: ing.name,
      quantity: ing.quantity || '',
      note: ing.note || '',
      isExisting: true,
    }));
    selectedIngredients.value = ingredients;
    deletedIngredientCodes.value.clear();

    // 锁定编辑状态
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
  deletedImageIds.value.clear();
  selectedIngredients.value = [];
  deletedIngredientCodes.value.clear();
  isEditLocked.value = false;
  searchError.value = '';
  searchCode.value = '';
};

// ---------- 处理食材选择（从弹窗添加）----------
const handleIngredientSelect = (ingredient: ViewDishIngredientCard) => {
  // 检查是否已经存在（包括已删除记录中？但删除后应移除，所以只需检查当前列表）
  const exists = selectedIngredients.value.some(
    (item) => item.ingredientCode === ingredient.ingredientCode
  );
  if (exists) {
    alert('该食材已添加，请勿重复选择');
    return;
  }

  // 新增食材，isExisting = false
  selectedIngredients.value.push({
    ingredientCode: ingredient.ingredientCode,
    name: ingredient.name,
    quantity: '',
    note: '',
    isExisting: false,
  });

  openIngredientPicker.value = false;
};

// ---------- 删除食材 ----------
const removeIngredient = (index: number) => {
  const item = selectedIngredients.value[index];
  if (!item) return;

  // 如果是原有食材，记录到删除集合
  if (item.isExisting) {
    deletedIngredientCodes.value.add(item.ingredientCode);
  }
  // 从列表中移除
  selectedIngredients.value.splice(index, 1);
};

// ---------- 提交表单 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;
  if (!form.dishCode) {
    alert('菜品编码异常，请重新选择菜品');
    return;
  }

  submitting.value = true;

  const formData = new FormData();
  formData.append('dishCode', form.dishCode);
  formData.append('name', form.name.trim());
  formData.append('description', form.description.trim());
  formData.append('recipe', form.recipe.trim());

  // ===== 构建图片数据 =====
  const newImageFiles: { tempID: string; file: File }[] = [];
  const imageRequests: any[] = [];

  imageList.value.forEach((img, idx) => {
    if (img.status === 'existing' && img.id) {
      imageRequests.push({
        type: 'existing',
        id: img.id,
        tempID: '',
        sortOrder: idx,
      });
    } else if (img.status === 'new' && img.file) {
      const tempID = 'temp' + uuidv7();
      imageRequests.push({
        type: 'new',
        id: 0,
        tempID: tempID,
        sortOrder: idx,
      });
      newImageFiles.push({
        tempID: tempID,
        file: img.file,
      });
    }
  });

  deletedImageIds.value.forEach((id) => {
    imageRequests.push({
      type: 'deleted',
      id,
      tempID: '',
      sortOrder: 0,
    });
  });

  // 添加 images 数组字段
  imageRequests.forEach((req, index) => {
    formData.append(`images[${index}].type`, req.type);
    formData.append(`images[${index}].id`, req.id.toString());
    formData.append(`images[${index}].tempID`, req.tempID || '');
    formData.append(`images[${index}].sortOrder`, req.sortOrder.toString());
  });

  // 添加 newImages 文件
  newImageFiles.forEach((file, index) => {
    formData.append(`newImages[${index}].tempID`, file.tempID);
    formData.append(`newImages[${index}].file`, file.file);
  });

  // ===== 构建食材数据 =====
  // 1. 当前列表中的食材（existing 和 new）
  selectedIngredients.value.forEach((ing, index) => {
    const type = ing.isExisting ? 'existing' : 'new';
    formData.append(`ingredients[${index}].type`, type);
    formData.append(`ingredients[${index}].ingredientCode`, ing.ingredientCode);
    formData.append(`ingredients[${index}].quantity`, ing.quantity.trim());
    formData.append(`ingredients[${index}].note`, ing.note.trim());
  });

  // 2. 被删除的原有食材（deleted）
  let deletedIndex = selectedIngredients.value.length;
  deletedIngredientCodes.value.forEach((code) => {
    formData.append(`ingredients[${deletedIndex}].type`, 'deleted');
    formData.append(`ingredients[${deletedIndex}].ingredientCode`, code);
    formData.append(`ingredients[${deletedIndex}].quantity`, '');
    formData.append(`ingredients[${deletedIndex}].note`, '');
    deletedIndex++;
  });

  try {
    await request({
      url: `/api/dishes/${form.dishCode}`,
      method: 'PATCH',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
     //router.push('/dishes'); // 跳转到菜品列表
    }, 1500);
  } catch (error) {
    console.error('更新菜品失败:', error);
    alert('更新失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
};

// 监听锁定状态，清空搜索输入
watch(isEditLocked, (locked) => {
  if (locked) {
    searchCode.value = '';
  }
});
</script>

<style scoped>
/* ========== 继承创建表单的精美样式 ========== */
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
  border-top-color: #4caf50;
  animation: spin 1s ease-in-out infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
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

/* 提示文本 */
.hint-text {
  font-size: 12px;
  color: #999;
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
  color: #4caf50;
  background-color: #e8f5e9;
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
  transition: all 0.2s;
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

.required-star {
  color: #e53935;
  margin-left: 2px;
}

.image-hint {
  font-size: 13px;
  font-weight: normal;
  color: #999;
  margin-left: 8px;
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

.form-control.textarea {
  resize: vertical;
  min-height: 100px;
  line-height: 1.6;
}

.form-control.small {
  padding: 8px 12px;
  font-size: 14px;
}

.error-message {
  font-size: 13px;
  color: #e53935;
  margin-top: 4px;
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

.btn-outline {
  background: white;
  border: 1px solid #4caf50;
  color: #4caf50;
}

.btn-outline:hover {
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

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background-color: transparent;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-icon:hover {
  background-color: #f0f0f0;
  color: #e53935;
}

.remove-btn {
  flex-shrink: 0;
}

.add-ingredient-btn {
  width: fit-content;
  margin-top: 8px;
}

/* 已选食材列表 */
.ingredients-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 修改 ingredient-item 使其成为相对定位容器，为右上角删除按钮预留空间 */
.ingredient-item {
  position: relative;
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 12px;
  padding-right: 48px; /* 为右侧删除按钮留出足够空间 */
  background-color: #fafafa;
}

/* 删除按钮定位到右上角 */

/* 删除按钮定位到右上角 - 修复图标居中与阴影问题 */
.ingredient-item .remove-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  background-color: rgba(0, 0, 0, 0.6); /* 增加不透明度提升对比度 */
  border: none;
  border-radius: 50%; /* 关键：圆形按钮使阴影自然包裹图标 */
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15); /* 增强阴影深度 */
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  z-index: 10; /* 提升层级避免被内容遮挡 */
  transition: all 0.2s ease;
  padding: 0; /* 消除内边距干扰居中 */
}

.ingredient-item .remove-btn:hover {
  background-color: #e53935; /* 直接使用警示色，避免浅色背景导致图标模糊 */
  color: white;
  box-shadow: 0 4px 10px rgba(229, 57, 53, 0.3);
  transform: scale(1.05);
}

/* 确保 SVG 图标正确继承颜色 */
.ingredient-item .remove-btn svg {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  stroke: currentColor; /* 关键：强制继承按钮 color */
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

.ingredient-badge.new {
  background-color: #e8f5e9;
  color: #4caf50;
}

.ingredient-fields {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.ingredient-fields .form-control.small {
  flex: 1 1 120px;
}

@media (max-width: 480px) {
  .ingredient-fields {
    flex-direction: column;
    align-items: stretch;
  }
  .ingredient-fields .form-control.small {
    width: 100%;
  }
}

/* 弹窗/抽屉样式（复用） */
.modal-overlay,
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  padding: 16px;
  box-sizing: border-box;
}

.drawer-overlay {
  align-items: flex-end;
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

.modal-content {
  background-color: white;
  border-radius: 20px;
  width: 100%;
  max-width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

.drawer-header,
.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.drawer-header h3,
.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.drawer-close,
.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  color: #999;
  padding: 4px 8px;
}

.drawer-close:hover,
.close-btn:hover {
  color: #333;
}

.drawer-body,
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

/* 食材卡片（用于 ScrollPicker 项） */
.ingredient-card {
  padding: 12px 16px;
  border: 1px solid #f0f0f0;
  border-radius: 10px;
  background-color: white;
  transition: all 0.2s ease;
}

.ingredient-card:hover {
  border-color: #4caf50;
  box-shadow: 0 2px 8px rgba(76, 175, 80, 0.1);
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

/* 打印样式 */
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
  .preview-remove,
  .upload-area,
  .drawer-overlay,
  .modal-overlay {
    display: none;
  }
}
</style>
