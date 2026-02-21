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

  <div class="ingredient-container">
    <!-- 提交时的加载状态 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在更新食材信息...</p>
    </div>

    <!-- 主卡片：编辑食材 -->
    <div class="ingredient-main-card">
      <div class="ingredient-header">
        <h1 class="ingredient-title">编辑食材</h1>
        <div class="ingredient-code-container">
          <span class="ingredient-code">支持两种方式快速填充</span>
        </div>
      </div>

      <!-- ======= 方式1：直接查询 ======= -->
      <div class="search-by-code-section">
        <div class="form-group" style="margin-bottom: 0;">
          <label class="form-label">通过食材编码查询</label>
          <div class="code-search-input-group">
            <input
              v-model="searchCode"
              type="text"
              class="form-control"
              placeholder="输入食材编码"
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

      <!-- ======= 方式2：从食材列表选择（抽屉触发） ======= -->
      <div class="selector-section">
        <h3 class="selector-title">从现有食材列表选择</h3>
        <!-- 按钮：点击打开抽屉 -->
        <div class="ingredient-picker-trigger">
          <button
            type="button"
            class="btn btn-outline select-ingredient-btn"
            @click="openIngredientPicker"
          >
            选择食材
          </button>
          <span v-if="form.ingredientCode" class="selected-hint">
            已选择：{{ form.ingredientCode }} {{ form.name }}
          </span>
        </div>
      </div>

      <!-- ======= 编辑表单：基于查询/选择填充后的数据 ======= -->
      <form @submit.prevent="handleSubmit" class="ingredient-form" v-if="isEditLocked">
        <!-- 食材编码（只读） -->
        <div class="form-group">
          <label class="form-label">
            食材编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.ingredientCode"
            type="text"
            class="form-control"
            disabled
          />
          <div class="hint-text">编码不可修改，如需编辑其他食材请点击「重置」</div>
        </div>

        <!-- 食材名称 -->
        <div class="form-group">
          <label class="form-label">
            食材名称 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.name"
            type="text"
            class="form-control"
            placeholder="例如: 面条"
            required
          />
          <div v-if="formErrors.name" class="error-message">{{ formErrors.name }}</div>
        </div>

        <!-- 食材描述 -->
        <div class="form-group">
          <label class="form-label">食材描述</label>
          <textarea
            v-model="form.description"
            class="form-control textarea"
            rows="4"
            placeholder="详细描述食材的特点、产地、风味等"
          ></textarea>
        </div>

        <!-- 图片管理组件，双向绑定 -->
        <ImageManager
          v-model:modelValue="imageList"
          @update:deletedIds="deletedImageIds = $event"
          :max-count="9"
        />

        <!-- 提交按钮区 -->
        <div class="form-actions">
          <router-link to="/ingredients" class="btn btn-secondary">取消</router-link>
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
        请通过上方「编码查询」或「食材列表」选择要编辑的食材
      </div>
    </div>

    <!-- 提交成功提示 -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>食材更新成功！</span>
      </div>
    </Teleport>

    <!-- ========== 食材选择抽屉（完全复用上方商品发布页面的样式与逻辑） ========== -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import request from '../../../api/request';
import ScrollPicker from '../../../components/picker/ScrollPicker.vue';
import type { FetchResult } from '../../../components/picker/ScrollPicker.vue';
import ImageManager from '../../../components/image/ImageManager.vue';
import type { ImageItem } from '../../../components/image/ImageManager.vue';
import type { NewImageFile,ImageRequest,ImageResponse, ViewIngredientCard, ViewIngredientCardListWithCursor } from '../../../types/types';
import { v7 as uuidv7 } from 'uuid';

// ---------- 表单状态 ----------
const form = reactive({
  ingredientCode: '',
  name: '',
  description: '',
});

// 图片列表 & 已删除ID集合（与 ImageManager 双向绑定）
const imageList = ref<ImageItem[]>([]);
const deletedImageIds = ref<Set<number>>(new Set());

const submitting = ref(false);
const showSuccessToast = ref(false);
const isEditLocked = ref(false);
const searchCode = ref('');
const searchError = ref('');
//const defaultImage = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjQwMCIgaGVpZ2h0PSIzMDAiIGZpbGw9IiNGM0YzRjMiLz48cGF0aCBkPSJNMjAwIDE1MEMyMjMuODE2IDE1MCAyNDMgMTMwLjgxNiAyNDMgMTA3QzI0MyA4My4xODM4IDIyMy44MTYgNjQgMjAwIDY0QzE3Ni4xODQgNjQgMTU3IDgzLjE4MzggMTU3IDEwN0MxNTcgMTMwLjgxNiAxNzYuMTg0IDE1MCAyMDAgMTUwWiIgZmlsbD0iI0Q4RDhEOCIvPjxwYXRoIGQ9Ik0xMjAgMjI2VjIxNkMxMjAgMjA5LjM4MyAxMjUuMzgzIDIwNCAxMzIgMjA0SDI2OEMyNzQuNjE3IDIwNCAyODAgMjA5LjM4MyAyODAgMjE2VjIyNiIgc3Ryb2tlPSIjRDhEOEQ4IiBzdHJva2Utd2lkdGg9IjEwIiBzdHJva2UtbGluZWNhcD0icm91bmQiLz48L3N2Zz4=';

// ---------- 抽屉状态 ----------
const showIngredientPicker = ref(false);
//const scrollPickerRef = ref<InstanceType<typeof ScrollPicker> | null>(null);

// ---------- 表单校验 ----------
const formErrors = reactive({
  name: '',
});

watch(() => form.name, (val) => {
  formErrors.name = val.trim() ? '' : '食材名称不能为空';
}, { immediate: true });

const isFormValid = computed(() => {
  return form.ingredientCode.trim() !== '' && form.name.trim() !== '' && imageList.value.length <= 9;
});

// ---------- 图片错误处理（仅用于 ScrollPicker）----------
/*
const handleImageLoadError = (event: Event) => {
  const target = event.target as HTMLImageElement;
  target.src = defaultImage;
};
*/

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
  /*
  // 可选：刷新列表
  setTimeout(() => {
    scrollPickerRef.value?.refresh();
  }, 50);
  */
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

    console.log("查询到的食材详情:", data);

    // 填充表单
    form.ingredientCode = data.ingredientCode;
    form.name = data.name;
    form.description = data.description || '';

    // 构建现有图片列表（直接赋值给 imageList，ImageManager 会自动接管）
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
    searchCode.value = '';
  } catch (error: any) {
    console.error('查询食材失败:', error);
    if (error.response?.status === 404) {
      searchError.value = '未找到该编码的食材，请确认或前往创建';
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
  deletedImageIds.value.clear();
  isEditLocked.value = false;
  searchError.value = '';
  searchCode.value = '';
};

// ---------- 提交表单 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;
  if (!form.ingredientCode) {
    alert('食材编码异常，请重新选择食材');
    return;
  }

  submitting.value = true;

  const formData = new FormData();
  formData.append('ingredientCode', form.ingredientCode);
  formData.append('name', form.name.trim());
  formData.append('description', form.description.trim());

  const newImageFiles: NewImageFile[] = [];

  
  const imageRequests: ImageRequest[] = [];

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
        tempID: tempID,
        sortOrder: idx,
      });
      newImageFiles.push({
        tempID: tempID,
        file: img.file,
      });
    }
  });


  console.log("删除图片ID列表:", deletedImageIds.value);


  deletedImageIds.value.forEach(id => {
    console.log("删除图片ID:", id);
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
  
  
  newImageFiles.forEach((img,index) => {
    formData.append(`newImages[${index}].tempID`, img.tempID);
    formData.append(`newImages[${index}].file`, img.file);
  });
  

  try {
    await request({
      url: `/api/ingredients/${form.ingredientCode}`,
      method: 'PATCH',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
      // router.push('/ingredients'); // 按需开启
      // 更新成功后重置表单
      resetForm();
    }, 1500);
  } catch (error) {
    console.error('更新食材失败:', error);
    alert('更新失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
};

// ---------- 监听锁定状态 ----------
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
/* ========== 完全继承创建表单的精美样式 ========== */
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

/* 表单样式 */
.ingredient-form {
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

/* ========== 更新表单专属样式 ========== */
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

/* 食材选择卡片区（现为按钮触发抽屉） */
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

/* 禁用状态编码输入框 */
.form-control:disabled {
  background-color: #f9f9f9;
  color: #666;
  border-color: #e0e0e0;
  cursor: not-allowed;
}

/* ========== 食材选择抽屉（完全复制上方商品发布页面样式） ========== */
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

/* 食材列表项样式（抽屉内） */
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

/* ========== 响应式布局 ========== */
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

/* 打印样式 */
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
  .preview-remove,
  .upload-area,
  .drawer-overlay {
    display: none;
  }
}
</style>