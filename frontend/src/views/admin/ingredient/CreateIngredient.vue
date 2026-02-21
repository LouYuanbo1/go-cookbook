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
      <p>正在提交食材信息...</p>
    </div>

    <!-- 主表单卡片 -->
    <div class="ingredient-main-card">
      <div class="ingredient-header">
        <h1 class="ingredient-title">创建新食材</h1>
        <div class="ingredient-code-container">
          <span class="ingredient-code">填写食材基础信息</span>
        </div>
      </div>

      <form @submit.prevent="handleSubmit" class="ingredient-form">
        <!-- 食材编码 -->
        <div class="form-group">
          <label class="form-label">
            食材编码 <span class="required-star">*</span>
          </label>
          <input
            v-model="form.ingredientCode"
            type="text"
            class="form-control"
            placeholder="例如: ING001"
            required
          />
          <div v-if="formErrors.ingredientCode" class="error-message">
            {{ formErrors.ingredientCode }}
          </div>
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
          <div v-if="formErrors.name" class="error-message">
            {{ formErrors.name }}
          </div>
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

        <!-- ========== 图片管理组件（完全复用） ========== -->
        <div class="form-group">
          <label class="form-label">
            食材图片
            <span class="image-hint">(最多9张，支持拖拽排序)</span>
          </label>
          <ImageManager
            v-model:modelValue="imageList"
            :max-count="9"
            upload-text="点击添加新图片"
            upload-hint="支持 jpg、png、webp"
          />
        </div>

        <!-- 表单操作按钮 -->
        <div class="form-actions">
          <router-link to="/ingredients" class="btn btn-secondary">
            取消
          </router-link>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !isFormValid"
          >
            {{ submitting ? "提交中..." : "创建食材" }}
          </button>
        </div>
      </form>
    </div>

    <!-- 提交成功提示 (Teleport) -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>食材创建成功！</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
//import { useRouter } from 'vue-router';
import request from '../../../api/request';

// ---------- 导入 ImageManager 组件 ----------
import ImageManager from '../../../components/image/ImageManager.vue';
import type { ImageItem } from '../../../components/image/ImageManager.vue';

// ---------- 类型定义 ----------
interface CreateIngredientForm {
  ingredientCode: string;
  name: string;
  description: string;
}

// ---------- 表单状态 ----------
const form = reactive<CreateIngredientForm>({
  ingredientCode: '',
  name: '',
  description: '',
});

// 图片列表（与 ImageManager 双向绑定）
const imageList = ref<ImageItem[]>([]);

const submitting = ref(false);
const showSuccessToast = ref(false);
//const router = useRouter();

// 表单错误信息
const formErrors = reactive({
  ingredientCode: '',
  name: '',
});

// ---------- 计算属性：表单是否有效 ----------
const isFormValid = computed(() => {
  return (
    form.ingredientCode.trim() !== '' &&
    form.name.trim() !== '' &&
    imageList.value.length <= 9
  );
});

// ---------- 监听表单字段，实时简单校验 ----------
watch(
  () => form.ingredientCode,
  (val) => {
    formErrors.ingredientCode = val.trim() ? '' : '食材编码不能为空';
  },
  { immediate: true }
);

watch(
  () => form.name,
  (val) => {
    formErrors.name = val.trim() ? '' : '食材名称不能为空';
  },
  { immediate: true }
);

// ---------- 表单提交 ----------
const handleSubmit = async () => {
  if (!isFormValid.value) return;

  submitting.value = true;

  const formData = new FormData();
  formData.append('ingredientCode', form.ingredientCode.trim());
  formData.append('name', form.name.trim());
  formData.append('description', form.description.trim());

  // ✅ 从 ImageManager 获取所有新上传的文件（status 为 'new'）
  imageList.value.forEach((item) => {
    if (item.status === 'new' && item.file) {
      formData.append('images', item.file);
    }
  });

  try {
    await request({
      url: '/api/ingredients',
      method: 'POST',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
      //router.push('/ingredients');
    }, 1500);
  } catch (error) {
    console.error('创建食材失败:', error);
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
/* ========== 样式完全继承原有创建表单 ========== */
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
  .upload-area {
    display: none;
  }
}
</style>