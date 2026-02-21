<template>
  <div class="top-nav">
    <router-link :to="{ name: 'AdminDashboard' }" class="nav-item">管理员工具栏</router-link>
    <router-link :to="{ name: 'CreateIngredient' }" class="nav-item">创建食材</router-link>
    <router-link :to="{ name: 'UpdateIngredient' }" class="nav-item">更新食材</router-link>
    <router-link :to="{ name: 'DeleteIngredient' }" class="nav-item">删除食材</router-link>
    <router-link :to="{ name: 'CreateProduct' }" class="nav-item">创建产品</router-link>
    <router-link :to="{ name: 'UpdateProduct' }" class="nav-item">更新产品</router-link>
    <router-link :to="{ name: 'DeleteProduct' }" class="nav-item">删除产品</router-link>
    <router-link :to="{ name: 'CreateDish' }" class="nav-item">创建菜品</router-link>
    <router-link :to="{ name: 'UpdateDish' }" class="nav-item">更新菜品</router-link>
    <router-link :to="{ name: 'DeleteDish' }" class="nav-item">删除菜品</router-link>
  </div>
  
  <div class="dish-container">
    <!-- 提交时的加载状态 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在提交菜品信息...</p>
    </div>

    <!-- 主表单卡片 -->
    <div class="dish-main-card">
      <div class="dish-header">
        <h1 class="dish-title">创建新菜品</h1>
        <div class="dish-code-container">
          <span class="dish-code">填写菜品基础信息</span>
        </div>
      </div>

      <form @submit.prevent="handleSubmit" class="dish-form">
        <!-- 菜品编码 -->
        <div class="form-group">
          <label class="form-label">
            菜品编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.dishCode"
            type="text"
            class="form-control"
            placeholder="例如: DISH001"
            required
          />
          <div v-if="formErrors.dishCode" class="error-message">
            {{ formErrors.dishCode }}
          </div>
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
          <div v-if="formErrors.name" class="error-message">
            {{ formErrors.name }}
          </div>
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
              <div class="ingredient-info">
                <span class="ingredient-name">{{ item.name }}</span>
                <span class="ingredient-code">{{ item.ingredientCode }}</span>
              </div>
              <!-- MODIFIED: 将单个用量输入拆分为数量 + 单位下拉框，保留备注输入 -->
              <div class="ingredient-fields">
                <input
                  v-model.number="item.amount"
                  type="number"
                  step="0.01"
                  min="0"
                  class="form-control small"
                  placeholder="用量"
                />
                <select v-model="item.unit" class="form-control small">
                  <option value="" disabled>单位</option>
                  <option v-for="unit in unitOptions" :key="unit.value" :value="unit.value">
                    {{ unit.label }}
                  </option>
                </select>
                <input
                  v-model="item.note"
                  type="text"
                  class="form-control small"
                  placeholder="备注 (可选)"
                />
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
            {{ submitting ? "提交中..." : "创建菜品" }}
          </button>
        </div>
      </form>
    </div>

    <!-- 提交成功提示 (Teleport) -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>菜品创建成功！</span>
      </div>
    </Teleport>

    <!-- 食材选择弹窗 (Teleport) -->
    <Teleport to="body">
      <div v-if="openIngredientPicker" class="modal-overlay" @click.self="openIngredientPicker = false">
        <div class="modal-content">
          <div class="modal-header">
            <h3>选择食材</h3>
            <button class="close-btn" @click="openIngredientPicker = false">&times;</button>
          </div>
          <div class="modal-body">
            <ScrollPicker
              ref="scrollPickerRef"
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
import type { FetchFunction, FetchResult } from '../../../components/picker/ScrollPicker.vue';
import type {ViewIngredientCard,ViewIngredientCardListWithCursor} from '../../../types/types';

// ---------- 枚举选项（从 CreateProduct 复制而来）----------
const unitOptions = [
  { value: 'gram', label: '克' },
  { value: 'kilogram', label: '千克' },
  { value: 'milliliter', label: '毫升' },
  { value: 'liter', label: '升' },
  { value: 'piece', label: '个' },
  { value: 'bowl', label: '碗' },
  { value: 'serving', label: '份' },
];

// ---------- 类型定义 ----------
interface CreateDishForm {
  dishCode: string;
  name: string;
  description: string;
  recipe: string;
}

// MODIFIED: 选中的食材项（用量拆分为 amount + unit）
interface SelectedIngredient {
  ingredientCode: string;
  name: string;
  amount: number | null;      // 数量
  unit: string;               // 单位代码
  note: string;
}

// ---------- 表单状态 ----------
const form = reactive<CreateDishForm>({
  dishCode: '',
  name: '',
  description: '',
  recipe: '',
});

// 图片列表（与 ImageManager 双向绑定）
const imageList = ref<ImageItem[]>([]);

// 选中的食材列表
const selectedIngredients = ref<SelectedIngredient[]>([]);

// 控制食材选择弹窗
const openIngredientPicker = ref(false);

const submitting = ref(false);
const showSuccessToast = ref(false);
//const router = useRouter();

// 表单错误信息
const formErrors = reactive({
  dishCode: '',
  name: '',
});


// ---------- 计算属性：表单是否有效 ----------
const isFormValid = computed(() => {
  return (
    form.dishCode.trim() !== '' &&
    form.name.trim() !== '' &&
    imageList.value.length <= 9
    // 食材可以为空，所以不强制要求
  );
});

// ---------- 监听表单字段，实时简单校验 ----------
watch(
  () => form.dishCode,
  (val) => {
    formErrors.dishCode = val.trim() ? '' : '菜品编码不能为空';
  },
  { immediate: true }
);

watch(
  () => form.name,
  (val) => {
    formErrors.name = val.trim() ? '' : '菜品名称不能为空';
  },
  { immediate: true }
);

// ---------- 获取食材列表的 fetch 函数（供 ScrollPicker 使用）----------
const fetchIngredients: FetchFunction<ViewIngredientCard> = async (cursor, limit): Promise<FetchResult<ViewIngredientCard>> => {
  try {
    const response = await request({
      url: '/api/ingredients',
      method: 'GET',
      params: { cursor, limit },
    });
    const data = response.data as ViewIngredientCardListWithCursor;
    return {
      items: data.ingredients,
      cursor: data.cursor,
      hasMore: data.hasMore,
    };
  } catch (error) {
    console.error('获取食材列表失败:', error);
    return { items: [], cursor: 0, hasMore: false };
  }
};

// ---------- 处理食材选择 ----------
const handleIngredientSelect = (ingredient: ViewIngredientCard) => {
  // 检查是否已经添加
  const exists = selectedIngredients.value.some(
    (item) => item.ingredientCode === ingredient.ingredientCode
  );
  if (exists) {
    alert('该食材已添加，请勿重复选择');
    return;
  }

  // MODIFIED: 初始化 amount 为 null，unit 为空字符串
  selectedIngredients.value.push({
    ingredientCode: ingredient.ingredientCode,
    name: ingredient.name,
    amount: null,
    unit: '',
    note: '',
  });

  // 关闭弹窗
  openIngredientPicker.value = false;
};

// ---------- 删除食材 ----------
const removeIngredient = (index: number) => {
  selectedIngredients.value.splice(index, 1);
};

// ---------- 表单提交 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;

  submitting.value = true;

  const formData = new FormData();
  formData.append('dishCode', form.dishCode.trim());
  formData.append('name', form.name.trim());
  formData.append('description', form.description.trim());
  formData.append('recipe', form.recipe.trim());

  // 添加图片文件
  imageList.value.forEach((item) => {
    if (item.status === 'new' && item.file) {
      formData.append('images', item.file);
    }
  });

  // MODIFIED: 遍历食材，将 amount + unit 组合成 quantity 字符串
  selectedIngredients.value.forEach((item, index) => {
    formData.append(`ingredients[${index}].ingredientCode`, item.ingredientCode);
    // 根据单位代码获取显示标签，组合用量字符串，若未填则留空
    const unitLabel = unitOptions.find(u => u.value === item.unit)?.label || item.unit;
    const quantityStr = item.amount && item.unit ? `${item.amount} ${unitLabel}` : '';
    formData.append(`ingredients[${index}].quantity`, quantityStr);
    formData.append(`ingredients[${index}].note`, item.note.trim());
  });

  try {
    await request({
      url: '/api/dishes',
      method: 'POST',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    
    setTimeout(() => {
      showSuccessToast.value = false;
      //router.push('/dishes');
    }, 1500);

  } catch (error) {
    console.error('创建菜品失败:', error);
    alert('创建失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
.top-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  background-color: #ffffff;
  padding: 12px 24px;
  margin: 20px 24px 16px 24px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  align-items: center;
  justify-content: center;
}

.nav-item {
  padding: 8px 16px;
  color: #2c3e50;
  font-size: 15px;
  font-weight: 500;
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.2s ease;
  position: relative;
}

.nav-item:hover {
  background-color: #f5f7fa;
  color: #409eff;
  transform: translateY(-1px);
}

/* ========== 基础样式继承创建食材表单 ========== */
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

/* MODIFIED: 下拉框样式继承 .form-control，无需额外定义 */
select.form-control.small {
  appearance: none;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23666' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 8px center;
  background-size: 14px;
  padding-right: 28px;
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

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background-color: transparent;
  color: #999;
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

.ingredient-item {
  border: 1px solid #f0f0f0;
  border-radius: 12px;
  padding: 12px;
  background-color: #fafafa;
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

/* MODIFIED: 调整三个输入框的布局，保持在一行或根据屏幕换行 */
.ingredient-fields {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.ingredient-fields .form-control.small {
  flex: 1 1 100px; /* 给每个输入框一个基础宽度，允许收缩 */
}

@media (max-width: 600px) {
  .ingredient-fields {
    flex-direction: column;
    align-items: stretch;
  }
  .ingredient-fields .form-control.small {
    width: 100%;
  }
  .remove-btn {
    align-self: flex-end;
  }
}

/* 弹窗样式 */
.modal-overlay {
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

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  line-height: 1;
  cursor: pointer;
  color: #999;
  padding: 4px 8px;
}

.close-btn:hover {
  color: #333;
}

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
  .modal-overlay {
    display: none;
  }
}
</style>