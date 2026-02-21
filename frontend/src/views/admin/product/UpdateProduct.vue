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
    <!-- 提交时的加载状态 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在更新商品信息...</p>
    </div>

    <!-- 主卡片：编辑商品 -->
    <div class="product-main-card">
      <div class="product-header">
        <h1 class="product-title">编辑商品</h1>
        <div class="product-code-container">
          <span class="product-code">支持两种方式快速填充</span>
        </div>
      </div>

      <!-- ======= 方式1：直接商品编码查询 ======= -->
      <div class="search-by-code-section">
        <div class="form-group" style="margin-bottom: 0">
          <label class="form-label">通过商品编码查询</label>
          <div class="code-search-input-group">
            <input
              v-model="searchCode"
              type="text"
              class="form-control"
              placeholder="输入商品编码"
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

      <!-- ======= 方式2：两阶段选择（食材 → 商品） ======= -->
      <div class="selector-section">
        <h3 class="selector-title">从现有食材及商品列表选择</h3>
        <div class="two-stage-picker">
          <!-- 当前已选食材展示及更换按钮 -->
          <div class="selected-ingredient" v-if="selectedIngredientCode">
            <span class="selected-label">当前食材：</span>
            <span class="selected-value">{{ selectedIngredientCode }}</span>

            <!-- :disabled="isEditLocked && !editingIngredient" -->
            <button
              type="button"
              class="btn btn-outline btn-sm"
              @click="openIngredientPicker"
              :disabled="submitting"
            >
              更换食材
            </button>
          </div>
          <div v-else>
            <button
              type="button"
              class="btn btn-outline select-ingredient-btn"
              @click="openIngredientPicker"
            >
              1. 选择食材
            </button>
          </div>

          <!-- 在已选食材基础上选择商品 -->
          <div v-if="selectedIngredientCode" class="product-picker-step">
            <button
              type="button"
              class="btn btn-outline select-product-btn"
              @click="openProductPicker"
              :disabled="!selectedIngredientCode"
            >
              2. 在 "{{ selectedIngredientCode }}" 下选择商品
            </button>
          </div>

          <span v-if="form.productCode" class="selected-hint">
            已选择商品：{{ form.productCode }} - {{ form.name }}
          </span>
        </div>
      </div>

      <!-- ======= 编辑表单：基于查询/选择填充后的数据 ======= -->
      <form @submit.prevent="handleSubmit" class="product-form" v-if="isEditLocked">
        <!-- 商品编码（只读） -->
        <div class="form-group">
          <label class="form-label">
            商品编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.productCode"
            type="text"
            class="form-control"
            disabled
          />
          <div class="hint-text">编码不可修改，如需编辑其他商品请点击「重置」</div>
        </div>

        <!-- 关联食材编码（可更换） -->
        <div class="form-group">
          <label class="form-label">
            关联食材编码 <span class="required-star">*</span>
          </label>
          <div class="ingredient-code-input-group">
            <input
              v-model="form.ingredientCode"
              type="text"
              class="form-control"
              :disabled="!editingIngredient"
              placeholder="不可直接修改，请通过下方按钮更换"
            />
            <button
              type="button"
              class="btn btn-outline btn-sm"
              @click="openIngredientPicker"
            >
              更换食材
            </button>
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
            placeholder="例如: 金沙河挂面"
            required
          />
          <div v-if="formErrors.name" class="error-message">{{ formErrors.name }}</div>
        </div>

        <!-- 价格 & 数量 & 单位（一行两列布局） -->
        <div class="form-row">
          <div class="form-group half">
            <label class="form-label">
              价格 <span class="required-star">*</span>
            </label>
            <input
              v-model.number="form.price"
              type="number"
              step="0.01"
              min="0"
              class="form-control"
              placeholder="¥"
              required
            />
            <div v-if="formErrors.price" class="error-message">{{ formErrors.price }}</div>
          </div>
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
              placeholder="例如: 500"
              required
            />
            <div v-if="formErrors.amount" class="error-message">{{ formErrors.amount }}</div>
          </div>
        </div>

        <!-- 单位 & 过敏原（一行两列） -->
        <div class="form-row">
          <div class="form-group half">
            <label class="form-label">
              单位 <span class="required-star">*</span>
            </label>
            <select v-model="form.unit" class="form-control" required>
              <option v-for="opt in unitOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="form-group half">
            <label class="form-label">
              过敏原 <span class="required-star">*</span>
            </label>
            <select v-model="form.allergenType" class="form-control" required>
              <option v-for="opt in allergenOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>

        <!-- 商品描述 -->
        <div class="form-group">
          <label class="form-label">商品描述</label>
          <textarea
            v-model="form.description"
            class="form-control textarea"
            rows="4"
            placeholder="详细描述商品的规格、特点、产地等"
          ></textarea>
        </div>

        <!-- 图片管理组件 -->
        <ImageManager
          v-model:modelValue="imageList"
          @update:deletedIds="deletedImageIds = $event"
          :max-count="9"
        />

        <!-- 提交按钮区 -->
        <div class="form-actions">
          <router-link to="/products" class="btn btn-secondary">取消</router-link>
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
        请通过上方「编码查询」或「两阶段选择」选定要编辑的商品
      </div>
    </div>

    <!-- 提交成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>商品更新成功！</span>
      </div>
    </Teleport>

    <!-- ========== 食材选择抽屉（复用ScrollPicker） ========== -->
    <Teleport to="body">
      <div
        v-if="showIngredientPicker"
        class="drawer-overlay"
        @click.self="closeIngredientPicker"
      >
        <div
          class="drawer-container"
          :class="{ 'drawer-open': showIngredientPicker }"
        >
          <div class="drawer-header">
            <h3>选择食材</h3>
            <button class="drawer-close" @click="closeIngredientPicker">&times;</button>
          </div>
          <div class="drawer-body">
            <ScrollPicker
              ref="ingredientPickerRef"
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

    <!-- ========== 商品选择抽屉（基于已选食材） ========== -->
    <Teleport to="body">
      <div
        v-if="showProductPicker"
        class="drawer-overlay"
        @click.self="closeProductPicker"
      >
        <div class="drawer-container" :class="{ 'drawer-open': showProductPicker }">
          <div class="drawer-header">
            <h3>选择商品（食材：{{ selectedIngredientCode }}）</h3>
            <button class="drawer-close" @click="closeProductPicker">&times;</button>
          </div>
          <div class="drawer-body">
            <ScrollPicker
              ref="productPickerRef"
              :fetch="fetchProductsByIngredient"
              :limit="15"
              @select="handleProductSelected"
            >
              <template #item="{ item }">
                <div class="product-item">
                  <span class="product-code">{{ item.productCode }}</span>
                  <span class="product-name">{{ item.name }}</span>
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
import ScrollPicker, { type FetchResult } from '../../../components/picker/ScrollPicker.vue';
import ImageManager, { type ImageItem } from '../../../components/image/ImageManager.vue';
import type {NewImageFile, ImageRequest, ImageResponse } from '../../../types/types';
import {v7 as uuidv7} from 'uuid';

// ---------- 枚举选项（需与后端model保持一致）----------
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
const form = reactive({
  productCode: '',
  ingredientCode: '',
  name: '',
  amount: 0,
  unit: 'gram',
  price: 0,
  description: '',
  allergenType: 'none',
});

// 图片列表 & 已删除ID集合
const imageList = ref<ImageItem[]>([]);
const deletedImageIds = ref<Set<number>>(new Set());

const submitting = ref(false);
const showSuccessToast = ref(false);
const isEditLocked = ref(false);
const editingIngredient = ref(false); // 是否允许编辑食材编码（直接输入框不可编辑，通过按钮更换）

const searchCode = ref('');
const searchError = ref('');

// ---------- 两阶段选择状态 ----------
const selectedIngredientCode = ref(''); // 当前选中的食材编码（用于获取商品列表）
const showIngredientPicker = ref(false);
const showProductPicker = ref(false);
const ingredientPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);
const productPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);

//const router = useRouter();

// ---------- 表单校验 ----------
const formErrors = reactive({
  name: '',
  price: '',
  amount: '',
});

watch(
  () => form.name,
  (val) => {
    formErrors.name = val.trim() ? '' : '商品名称不能为空';
  },
  { immediate: true }
);
watch(
  () => form.price,
  (val) => {
    if (val === null || val === undefined ) formErrors.price = '价格不能为空';
    else if (val < 0) formErrors.price = '价格不能为负数';
    else formErrors.price = '';
  },
  { immediate: true }
);
watch(
  () => form.amount,
  (val) => {
    if (val === null || val === undefined ) formErrors.amount = '数量不能为空';
    else if (val < 0) formErrors.amount = '数量不能为负数';
    else formErrors.amount = '';
  },
  { immediate: true }
);

const isFormValid = computed(() => {
  return (
    form.productCode.trim() !== '' &&
    form.ingredientCode.trim() !== '' &&
    form.name.trim() !== '' &&
    form.price > 0 &&
    form.amount > 0 &&
    form.unit &&
    form.allergenType &&
    imageList.value.length <= 9
  );
});

// ---------- API 获取函数 ----------
// 获取食材列表（用于ScrollPicker）
const fetchIngredients = async (cursor: number, limit: number): Promise<FetchResult<any>> => {
  const res = await request({
    url: '/api/ingredients',
    method: 'GET',
    params: { cursor, limit },
  });
  const data = res.data;
  return {
    items: data.ingredients || [],
    cursor: data.cursor || 0,
    hasMore: data.hasMore || false,
  };
};

// 获取指定食材下的商品列表（用于ScrollPicker）
const fetchProductsByIngredient = async (cursor: number, limit: number): Promise<FetchResult<any>> => {
  if (!selectedIngredientCode.value) {
    return { items: [], cursor: 0, hasMore: false };
  }
  /*
  未来可能使用这个路由,现在暂时注释
  const res = await request({
    url: '/api/products',
    method: 'GET',
    params: {
      ingredientCode: selectedIngredientCode.value,
      cursor,
      limit,
    },
  });
  */
  const res = await request({
    url: `/api/ingredients/${selectedIngredientCode.value}/products`,
    method: 'GET',
    params: {
      cursor,
      limit,
    },
  });
  const data = res.data;
  return {
    items: data.products || [],
    cursor: data.cursor || 0,
    hasMore: data.hasMore || false,
  };
};

// ---------- 通用：查询商品详情（通过商品编码）----------
const fetchProductDetail = async (code: string) => {
  try {
    searchError.value = '';
    const res = await request({
      url: `/api/products/${code}`,
      method: 'GET',
    });
    const data = res.data;

    // 填充表单
    form.productCode = data.productCode;
    form.ingredientCode = data.ingredientCode;
    form.name = data.name;
    form.amount = data.amount;
    form.unit = data.unit || 'gram';
    form.price = data.price;
    form.description = data.description || '';
    form.allergenType = data.allergenType || 'none';

    // 同步选中的食材编码（用于后续更换食材等）
    selectedIngredientCode.value = data.ingredientCode;

    // 构建现有图片列表
    const existingImages: ImageItem[] = (data.images || []).map((img: ImageResponse) => ({
      id: img.id,
      url: img.imageURL,
      status: 'existing',
      sortOrder: img.sortOrder,
    }));
    existingImages.sort((a, b) => a.sortOrder - b.sortOrder);
    imageList.value = existingImages;
    deletedImageIds.value.clear();

    isEditLocked.value = true;
    editingIngredient.value = false; // 默认不允许直接修改食材编码，通过按钮更换
    searchCode.value = '';
  } catch (error: any) {
    console.error('查询商品失败:', error);
    if (error.response?.status === 404) {
      searchError.value = '未找到该编码的商品，请确认或前往创建';
    } else {
      searchError.value = '查询失败，请稍后重试';
    }
  }
};

// ---------- 编码查询按钮 ----------
const handleSearchByCode = () => {
  if (!searchCode.value.trim()) return;
  fetchProductDetail(searchCode.value.trim());
};

// ---------- 重置表单 ----------
const resetForm = () => {
  form.productCode = '';
  form.ingredientCode = '';
  form.name = '';
  form.amount = 0;
  form.unit = 'gram';
  form.price = 0;
  form.description = '';
  form.allergenType = 'none';
  imageList.value = [];
  deletedImageIds.value.clear();
  isEditLocked.value = false;
  editingIngredient.value = false;
  searchError.value = '';
  searchCode.value = '';
  selectedIngredientCode.value = '';
};

// ---------- 两阶段选择：抽屉打开/关闭 ----------
const openIngredientPicker = () => {
  showIngredientPicker.value = true;
  // 可刷新列表
  setTimeout(() => {
    ingredientPickerRef.value?.refresh();
  }, 50);
};
const closeIngredientPicker = () => {
  showIngredientPicker.value = false;
};

// 食材选中回调
const handleIngredientSelected = (item: any) => {
  selectedIngredientCode.value = item.ingredientCode;
  closeIngredientPicker();
  // 自动打开商品选择抽屉
  openProductPicker();
};

// 商品选择抽屉
const openProductPicker = () => {
  if (!selectedIngredientCode.value) {
    alert('请先选择食材');
    return;
  }
  showProductPicker.value = true;
  setTimeout(() => {
    productPickerRef.value?.refresh();
  }, 50);
};
const closeProductPicker = () => {
  showProductPicker.value = false;
};

// 商品选中回调 → 填充表单
const handleProductSelected = (item: any) => {
  fetchProductDetail(item.productCode);
  closeProductPicker();
};

// ---------- 提交表单 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;
  if (!form.productCode) {
    alert('商品编码异常，请重新选择商品');
    return;
  }

  submitting.value = true;

  const formData = new FormData();
  formData.append('productCode', form.productCode);
  formData.append('ingredientCode', form.ingredientCode);
  formData.append('name', form.name.trim());
  formData.append('amount', form.amount.toString());
  formData.append('unit', form.unit);
  formData.append('price', form.price.toString());
  formData.append('description', form.description.trim());
  formData.append('allergenType', form.allergenType);

  // 构造图片请求数组
  const imageRequests: ImageRequest[] = [];
  const newImageFiles: NewImageFile[] = [];

  imageList.value.forEach((img, idx) => {
    if (img.status === 'existing' && img.id) {
      imageRequests.push({
        type: 'existing',
        id: img.id,
        sortOrder: idx,
      });
    } else if (img.status === 'new' && img.file) {
      const tempID = "temp"+uuidv7();
      imageRequests.push({
        type: 'new',
        id: 0,
        tempID:tempID,
        sortOrder: idx,
      });
      newImageFiles.push({
        tempID:tempID,
        file: img.file,
      });
    }
  });

  deletedImageIds.value.forEach((id) => {
    imageRequests.push({
      type: 'deleted',
      id,
      sortOrder: 0,
    });
  });

  imageRequests.forEach((req,index) =>{
    formData.append(`images[${index}].type`, req.type);
    formData.append(`images[${index}].id`, req.id.toString());
    formData.append(`images[${index}].tempID`, req.tempID || '');
    formData.append(`images[${index}].sortOrder`, req.sortOrder.toString());
  })

  /*
  newImageFiles.forEach((img) => {
    formData.append('newImageTempIDs', img.tempID);
    formData.append('newImages', img.file);
  });
  */

  
  newImageFiles.forEach((img,index) => {
    formData.append(`newImages[${index}].tempID`, img.tempID);
    formData.append(`newImages[${index}].file`, img.file);
  });
  

  try {
    await request({
      url: `/api/products/${form.productCode}`,
      method: 'PATCH',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
      //router.push('/products'); // 提交后跳转到商品列表
    }, 1500);
  } catch (error) {
    console.error('更新商品失败:', error);
    alert('更新失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
};

// ---------- 监听锁定状态，清空搜索输入 ----------
watch(isEditLocked, (locked) => {
  if (locked) {
    searchCode.value = '';
  }
});
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
/* ========== 完全继承食材编辑页面的精美样式，并针对商品表单微调 ========== */
.product-container {
  font-family: 'PingFang SC', 'Helvetica Neue', Arial, sans-serif;
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
  display: flex;
  flex-direction: column;
  padding: 24px;
}
.product-header {
  margin-bottom: 24px;
  flex-shrink: 0;
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

/* 表单样式 */
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
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23666' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 16px center;
  padding-right: 44px;
}

/* 错误提示 */
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

/* 提示文本 */
.hint-text {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 行内组合（用于价格/数量等） */
.form-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
.form-row .form-group.half {
  flex: 1 1 calc(50% - 8px);
  min-width: 200px;
}
@media (max-width: 639px) {
  .form-row .form-group.half {
    flex: 1 1 100%;
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

/* 两阶段选择区域 */
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
.two-stage-picker {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.selected-ingredient {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  background: #e8f5e9;
  padding: 10px 16px;
  border-radius: 30px;
}
.selected-label {
  font-size: 14px;
  color: #2e7d32;
}
.selected-value {
  font-weight: 600;
  color: #1b5e20;
}
.product-picker-step {
  margin-top: 4px;
}
.select-ingredient-btn,
.select-product-btn {
  min-width: 180px;
}
.selected-hint {
  font-size: 14px;
  color: #4caf50;
  background-color: #e8f5e9;
  padding: 6px 16px;
  border-radius: 30px;
  display: inline-block;
  align-self: flex-start;
}

/* 食材编码输入组（按钮在右侧） */
.ingredient-code-input-group {
  display: flex;
  gap: 12px;
  align-items: center;
}
.ingredient-code-input-group .form-control {
  flex: 1;
  background-color: #f9f9f9;
}
.ingredient-code-input-group .btn {
  flex-shrink: 0;
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

/* 禁用状态输入框 */
.form-control:disabled {
  background-color: #f9f9f9;
  color: #666;
  border-color: #e0e0e0;
  cursor: not-allowed;
}

/* ========== 抽屉样式（完全复用） ========== */
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

/* 列表项样式 */
.ingredient-item,
.product-item {
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  background-color: #fafafa;
  border-radius: 12px;
  transition: background-color 0.2s;
}
.ingredient-item:hover,
.product-item:hover {
  background-color: #f0f9f0;
}
.ingredient-code,
.product-code {
  font-size: 14px;
  font-weight: 600;
  color: #4caf50;
  margin-bottom: 4px;
}
.ingredient-name,
.product-name {
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

/* ========== 响应式微调 ========== */
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
  .ingredient-code-input-group {
    flex-wrap: wrap;
  }
  .ingredient-code-input-group .btn {
    width: 100%;
  }
  .selected-ingredient {
    flex-direction: column;
    align-items: flex-start;
  }
  .drawer-container {
    height: 80vh;
  }
}

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
  }
}

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
  .drawer-overlay {
    display: none;
  }
}
</style>