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

  <div class="product-container">
    <!-- 提交加载状态 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在发布商品榜单...</p>
    </div>

    <!-- 主表单卡片 -->
    <div class="product-main-card">
      <div class="product-header">
        <h1 class="product-title">发布新商品</h1>
        <div class="product-code-container">
          <span class="product-code">完善商品信息，绑定食材</span>
        </div>
      </div>

      <form @submit.prevent="handleSubmit" class="product-form">
        <!-- 商品编码 -->
        <div class="form-group">
          <label class="form-label">
            商品编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.productCode"
            type="text"
            class="form-control"
            placeholder="例如: PROD001"
            required
          />
          <div v-if="formErrors.productCode" class="error-message">
            {{ formErrors.productCode }}
          </div>
        </div>

        <!-- 商品名称 -->
        <div class="form-group">
          <label class="form-label">
            商品名称 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            class="form-control"
            placeholder="例如: 招牌牛肉面"
            required
          />
          <div v-if="formErrors.name" class="error-message">
            {{ formErrors.name }}
          </div>
        </div>

        <!-- 食材编码（支持输入或选择） -->
        <div class="form-group">
          <label class="form-label">
            绑定食材 <span class="required-star">*</span>
          </label>
          <div class="ingredient-code-wrapper">
            <input
              v-model="form.ingredientCode"
              type="text"
              class="form-control ingredient-code-input"
              placeholder="输入食材编码，或点击右侧选择"
              required
            />
            <button
              type="button"
              class="btn btn-outline select-ingredient-btn"
              @click="openIngredientPicker"
            >
              选择食材
            </button>
          </div>
          <div v-if="formErrors.ingredientCode" class="error-message">
            {{ formErrors.ingredientCode }}
          </div>
        </div>

        <!-- 份量 & 单位（并列） -->
        <div class="form-row">
          <div class="form-group half">
            <label class="form-label">
              量值 <span class="required-star">*</span>
            </label>
            <input
              v-model.number="form.amount"
              type="number"
              step="0.01"
              min="0"
              class="form-control"
              placeholder="例如: 300"
              required
            />
            <div v-if="formErrors.amount" class="error-message">
              {{ formErrors.amount }}
            </div>
          </div>
          <div class="form-group half">
            <label class="form-label">
              单位 <span class="required-star">*</span>
            </label>
            <select v-model="form.unit" class="form-control" required>
              <option value="" disabled>请选择单位</option>
              <option v-for="unit in unitOptions" :key="unit.value" :value="unit.value">
                {{ unit.label }}
              </option>
            </select>
            <div v-if="formErrors.unit" class="error-message">
              {{ formErrors.unit }}
            </div>
          </div>
        </div>

        <!-- 价格 -->
        <div class="form-group">
          <label class="form-label">
            价格（元） <span class="required-star">*</span>
          </label>
          <input
            v-model.number="form.price"
            type="number"
            step="0.01"
            min="0"
            class="form-control"
            placeholder="例如: 28.00"
            required
          />
          <div v-if="formErrors.price" class="error-message">
            {{ formErrors.price }}
          </div>
        </div>

        <!-- 过敏原类型 -->
        <div class="form-group">
          <label class="form-label">
            过敏原信息 <span class="required-star">*</span>
          </label>
          <select v-model="form.allergenType" class="form-control" required>
            <option value="" disabled>请选择过敏原类型</option>
            <option v-for="type in allergenOptions" :key="type.value" :value="type.value">
              {{ type.label }}
            </option>
          </select>
          <div v-if="formErrors.allergenType" class="error-message">
            {{ formErrors.allergenType }}
          </div>
        </div>

        <!-- 商品描述 -->
        <div class="form-group">
          <label class="form-label">商品描述</label>
          <textarea
            v-model="form.description"
            class="form-control textarea"
            rows="4"
            placeholder="详细介绍商品的口味、原料、制作工艺等"
          ></textarea>
        </div>

        <!-- ========== 图片管理组件（完全复用） ========== -->
        <div class="form-group">
          <label class="form-label">
            商品图片
            <span class="image-hint">(最多9张，支持拖拽排序)</span>
          </label>
          <ImageManager
            v-model:modelValue="imageList"
            :max-count="9"
            upload-text="点击添加商品图片"
            upload-hint="支持 jpg、png、webp"
          />
        </div>

        <!-- 表单操作按钮 -->
        <div class="form-actions">
          <router-link to="/products" class="btn btn-secondary">
            取消
          </router-link>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !isFormValid"
          >
            {{ submitting ? "发布中..." : "发布商品" }}
          </button>
        </div>
      </form>
    </div>

    <!-- 食材选择抽屉（移动端适配） -->
    <Teleport to="body">
      <div v-if="showIngredientPicker" class="drawer-overlay" @click.self="closeIngredientPicker">
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

    <!-- 提交成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>商品榜单发布成功！</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
//import { useRouter } from 'vue-router';
import request from '../../../api/request';

// ---------- 导入组件 ----------
import ImageManager from '../../../components/image/ImageManager.vue';
import ScrollPicker, { type FetchFunction, type FetchResult } from '../../../components/picker/ScrollPicker.vue';
import type { ImageItem } from '../../../components/image/ImageManager.vue';
import type { ViewIngredientCard ,ViewIngredientCardListWithCursor} from '../../../types/types';


interface CreateProductForm {
  productCode: string;
  ingredientCode: string;
  name: string;
  amount: number | null;
  unit: string;
  description: string;
  price: number | null;
  allergenType: string;
}

// ---------- 枚举选项（需根据后端 model.UnitType 和 model.AllergenType 调整） ----------
const unitOptions = [
  { value: 'gram', label: '克' },
  { value: 'kilogram', label: '千克' },
  { value: 'milliliter', label: '毫升' },
  { value: 'liter', label: '升' },
  { value: 'piece', label: '个' },
  { value: 'bowl', label: '碗' },
  { value: 'serving', label: '份' },
];

const allergenOptions = [
  { value: 'none', label: '无过敏原' },
  { value: 'gluten', label: '麸质' },
  { value: 'crustacean', label: '甲壳类' },
  { value: 'egg', label: '蛋类' },
  { value: 'fish', label: '鱼类' },
  { value: 'peanut', label: '花生' },
  { value: 'soy', label: '大豆' },
  { value: 'milk', label: '乳制品' },
  { value: 'nut', label: '坚果' },
  { value: 'celery', label: '芹菜' },
  { value: 'mustard', label: '芥末' },
  { value: 'sesame', label: '芝麻' },
  { value: 'sulphite', label: '亚硫酸盐' },
  { value: 'lupin', label: '羽扇豆' },
  { value: 'mollusc', label: '软体动物' },
];

// ---------- 表单状态 ----------
const form = reactive<CreateProductForm>({
  productCode: '',
  ingredientCode: '',
  name: '',
  amount: null,
  unit: '',
  description: '',
  price: null,
  allergenType: '',
});

// 图片列表（与 ImageManager 双向绑定）
const imageList = ref<ImageItem[]>([]);

const submitting = ref(false);
const showSuccessToast = ref(false);
//const router = useRouter();

// 食材选择器控制
const showIngredientPicker = ref(false);
//const scrollPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);

// 表单错误信息
const formErrors = reactive({
  productCode: '',
  ingredientCode: '',
  name: '',
  amount: '',
  unit: '',
  price: '',
  allergenType: '',
});

// ---------- 计算属性：表单是否有效 ----------
const isFormValid = computed(() => {
  return (
    form.productCode.trim() !== '' &&
    form.ingredientCode.trim() !== '' &&
    form.name.trim() !== '' &&
    form.amount !== null &&
    !isNaN(form.amount) &&
    form.amount > 0 &&
    form.unit !== '' &&
    form.price !== null &&
    !isNaN(form.price) &&
    form.price >= 0 &&
    form.allergenType !== '' &&
    imageList.value.length <= 9
  );
});

// ---------- 简单实时校验 ----------
watch(
  () => form.productCode,
  (val) => {
    formErrors.productCode = val.trim() ? '' : '商品编码不能为空';
  },
  { immediate: true }
);
watch(
  () => form.ingredientCode,
  (val) => {
    formErrors.ingredientCode = val.trim() ? '' : '请填写或选择食材编码';
  },
  { immediate: true }
);
watch(
  () => form.name,
  (val) => {
    formErrors.name = val.trim() ? '' : '商品名称不能为空';
  },
  { immediate: true }
);
watch(
  () => form.amount,
  (val) => {
    formErrors.amount = val && val > 0 ? '' : '请输入有效的份量';
  },
  { immediate: true }
);
watch(
  () => form.unit,
  (val) => {
    formErrors.unit = val ? '' : '请选择单位';
  },
  { immediate: true }
);
watch(
  () => form.price,
  (val) => {
    formErrors.price = val !== null && val >= 0 ? '' : '请输入有效的价格';
  },
  { immediate: true }
);
watch(
  () => form.allergenType,
  (val) => {
    formErrors.allergenType = val ? '' : '请选择过敏原信息';
  },
  { immediate: true }
);

// ---------- 获取食材列表（用于 ScrollPicker）----------
const fetchIngredients: FetchFunction = async (cursor: number, limit: number): Promise<FetchResult<ViewIngredientCard>> => {
  try {
    // 假设后端接口：GET /api/ingredients?cursor=0&limit=20
    const response = await request({
      url: '/api/ingredients',
      method: 'GET',
      params: { cursor, limit },
    });
    // 根据实际返回结构调整，这里假设返回 { items, cursor, hasMore }
    const data: ViewIngredientCardListWithCursor = response.data;
    return {
      items: data.ingredients || [],
      cursor: data.cursor || 0,
      hasMore: data.hasMore || false,
    };
  } catch (error) {
    console.error('获取食材列表失败:', error);
    return {
      items: [],
      cursor: 0,
      hasMore: false,
    };
  }
};

// ---------- 打开食材选择器 ----------
const openIngredientPicker = () => {
  showIngredientPicker.value = true;
  /*
  // 可选：刷新列表
  setTimeout(() => {
    scrollPickerRef.value?.refresh();
  }, 50);
  */
};

// ---------- 关闭选择器 ----------
const closeIngredientPicker = () => {
  showIngredientPicker.value = false;
};

// ---------- 选中食材后的回调 ----------
const handleIngredientSelected = (item: any) => {
  form.ingredientCode = item.ingredientCode; // 假设食材对象包含 ingredientCode
  closeIngredientPicker();
};

// ---------- 表单提交 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;

  submitting.value = true;

  const formData = new FormData();
  formData.append('productCode', form.productCode.trim());
  formData.append('ingredientCode', form.ingredientCode.trim());
  formData.append('name', form.name.trim());
  formData.append('amount', String(form.amount));
  formData.append('unit', form.unit);
  formData.append('price', String(form.price));
  formData.append('allergenType', form.allergenType);
  formData.append('description', form.description.trim());

  // 添加新上传的图片文件
  imageList.value.forEach((item) => {
    if (item.status === 'new' && item.file) {
      formData.append('images', item.file);
    }
  });

  try {
    await request({
      url: '/api/products', // 假设创建商品的接口
      method: 'POST',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    /*
    setTimeout(() => {
      showSuccessToast.value = false;
      router.push('/products'); // 跳转到商品榜单列表页
    }, 1500);
    */
  } catch (error) {
    console.error('发布商品失败:', error);
    alert('发布失败，请稍后重试');
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
/* ========== 全局容器样式，与食材创建表单保持设计语言 ========== */
.product-container {
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
.product-main-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  padding: 24px;
}

.product-header {
  margin-bottom: 24px;
}

.product-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.product-code-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.product-code {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

/* 表单通用 */
.product-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.form-group.half {
  flex: 1 1 calc(50% - 8px);
  min-width: 180px;
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

select.form-control {
  appearance: none;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23666' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 12px center;
  background-size: 16px;
  padding-right: 40px;
}

.error-message {
  font-size: 13px;
  color: #e53935;
  margin-top: 4px;
}

/* 食材编码输入与选择按钮组合 */
.ingredient-code-wrapper {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.ingredient-code-input {
  flex: 1 1 200px;
}

.select-ingredient-btn {
  flex: 0 0 auto;
  white-space: nowrap;
  padding: 12px 20px;
  font-size: 15px;
  border-radius: 30px;
  background-color: white;
  border: 1px solid #4caf50;
  color: #4caf50;
  transition: all 0.3s ease;
  cursor: pointer;
}

.select-ingredient-btn:hover {
  background-color: #4caf50;
  color: white;
  border-color: #4caf50;
}

.btn-outline {
  background: white;
  border: 1px solid #ddd;
}

/* 按钮组 */
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

/* ===== 食材选择抽屉 (移动优先) ===== */
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

/* 食材列表项样式 */
.ingredient-item {
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  background-color: #fafafa;
  border-radius: 12px;
  transition: background-color 0.2s;
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

/* ===== 响应式设计：移动端优先增强 ===== */
@media (max-width: 639px) {
  .product-container {
    padding: 12px;
  }

  .product-main-card {
    padding: 20px 16px;
  }

  .product-title {
    font-size: 22px;
  }

  .form-row {
    flex-direction: column;
    gap: 16px;
  }

  .form-group.half {
    flex: 1 1 100%;
    min-width: auto;
  }

  .ingredient-code-wrapper {
    flex-direction: column;
    align-items: stretch;
  }

  .select-ingredient-btn {
    width: 100%;
  }

  .form-actions {
    flex-direction: column-reverse;
    gap: 12px;
  }

  .btn {
    width: 100%;
  }

  .drawer-container {
    height: 80vh;
    border-radius: 20px 20px 0 0;
  }
}

/* 中等屏幕优化 */
@media (min-width: 768px) {
  .product-main-card {
    padding: 32px;
  }

  .product-title {
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

/* 大屏幕 */
@media (min-width: 1024px) {
  .product-container {
    padding: 24px;
  }

  .product-title {
    font-size: 32px;
  }
}

@media (min-width: 1400px) {
  .product-container {
    max-width: 1400px;
    padding: 32px;
  }

  .product-main-card {
    padding: 40px;
  }

  .product-title {
    font-size: 36px;
  }
}

/* 打印样式 */
@media print {
  .product-container {
    padding: 0;
    background-color: white;
  }
  .product-main-card {
    box-shadow: none;
    border: 1px solid #ddd;
  }
  .form-actions,
  .drawer-overlay,
  .preview-remove,
  .upload-area {
    display: none;
  }
}
</style>