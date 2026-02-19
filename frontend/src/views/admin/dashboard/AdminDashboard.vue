<template>
  <div class="top-nav">
    <router-link :to="{ name: 'Home' }" class="nav-item">首页</router-link>
    <router-link :to="{ name: 'CreateIngredient' }" class="nav-item">创建食材</router-link>
    <router-link :to="{ name: 'UpdateIngredient' }" class="nav-item">更新食材</router-link>
    <router-link :to="{ name: 'DeleteIngredient' }" class="nav-item">删除食材</router-link>
    <router-link :to="{ name: 'CreateProduct' }" class="nav-item">创建产品</router-link>
    <router-link :to="{ name: 'UpdateProduct' }" class="nav-item">更新产品</router-link>
    <router-link :to="{ name: 'DeleteProduct' }" class="nav-item">删除产品</router-link>
    <router-link :to="{ name: 'CreateDish' }" class="nav-item">创建菜单</router-link>
    <router-link :to="{ name: 'UpdateDish' }" class="nav-item">更新菜单</router-link>
    <router-link :to="{ name: 'DeleteDish' }" class="nav-item">删除菜单</router-link>
  </div>
  <div class="util-btn">
    <div class="button-group">
      <button
        class="export-btn"
        :disabled="loading"
        @click="handleExport('ingredients')"
      >
        <span v-if="loading && currentType === 'ingredients'" class="spinner"></span>
        {{ loading && currentType === 'ingredients' ? '导出中...' : '导出食材' }}
      </button>
      <p v-if="error && currentType === 'ingredients'" class="error">{{ error }}</p>
    </div>

    <div class="button-group">
      <button
        class="export-btn"
        :disabled="loading"
        @click="handleExport('products')"
      >
        <span v-if="loading && currentType === 'products'" class="spinner"></span>
        {{ loading && currentType === 'products' ? '导出中...' : '导出产品' }}
      </button>
      <p v-if="error && currentType === 'products'" class="error">{{ error }}</p>
    </div>

    <div class="button-group">
      <button
        class="export-btn"
        :disabled="loading"
        @click="handleExport('dishes')"
      >
        <span v-if="loading && currentType === 'dishes'" class="spinner"></span>
        {{ loading && currentType === 'dishes' ? '导出中...' : '导出菜单' }}
      </button>
      <p v-if="error && currentType === 'dishes'" class="error">{{ error }}</p>
    </div>

    <!-- 新增导入按钮 -->
    <div class="button-group">
      <button class="export-btn" @click="openImportModal('ingredients')">导入食材</button>
    </div>
    <div class="button-group">
      <button class="export-btn" @click="openImportModal('products')">导入商品</button>
    </div>
  </div>

  <!-- 自定义提示 Toast（原有） -->
  <div v-if="toastVisible" class="toast" :class="toastType">
    {{ toastMessage }}
  </div>

  <!-- 导入弹窗 (Teleport) -->
  <Teleport to="body">
    <div v-if="showImportModal" class="modal-overlay" @click.self="closeImportModal">
      <div class="modal-container">
        <ImportExcel
          :type="importType"
          :batchSize="200"
          @close="closeImportModal"
          @import-success="handleImportSuccess"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import request from '../../../api/request';
import { ref } from 'vue';
import { AxiosError } from 'axios';
import ImportExcel from '../../../components/excel/ImportExcel.vue'; // 根据实际路径调整

// 原有状态
const loading = ref(false);
const error = ref<string | null>(null);
const currentType = ref<string>('');
const toastVisible = ref(false);
const toastMessage = ref('');
const toastType = ref<'success' | 'error'>('success');

// 新增弹窗状态
const showImportModal = ref(false);
const importType = ref<'ingredients' | 'products'>('ingredients');

// 原有方法
const showToast = (message: string, type: 'success' | 'error' = 'success') => {
  toastMessage.value = message;
  toastType.value = type;
  toastVisible.value = true;
  setTimeout(() => {
    toastVisible.value = false;
  }, 3000);
};

const handleExport = async (type: string) => {
  if (loading.value) return;
  loading.value = true;
  currentType.value = type;
  error.value = null;

  try {
    const response = await request({
      url: `/api/${type}/export`,
      method: 'GET',
      params: { batchSize: 200 },
      responseType: 'blob',
    });

    const contentType = response.headers['content-type'];
    if (contentType && contentType.includes('application/json')) {
      const text = await response.data.text();
      const errorData = JSON.parse(text);
      throw new Error(errorData.error || errorData.msg || '导出失败');
    }

    const contentDisposition = response.headers['content-disposition'];
    const filename = getFileNameFromContentDisposition(contentDisposition, type);

    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);

    showToast('导出成功', 'success');
  } catch (err: unknown) {
    console.error('导出失败:', err);
    let errorMsg = '未知错误';
    if (err instanceof AxiosError && err.response?.data instanceof Blob) {
      try {
        const text = await err.response.data.text();
        const errorData = JSON.parse(text);
        errorMsg = errorData.error || errorData.msg || '导出失败';
      } catch {
        errorMsg = err.message;
      }
    } else if (err instanceof Error) {
      errorMsg = err.message;
    }
    error.value = errorMsg;
    showToast(errorMsg, 'error');
  } finally {
    loading.value = false;
    currentType.value = '';
  }
};

function getFileNameFromContentDisposition(contentDisposition: string | null, type: string): string {
  const defaultNames: Record<string, string> = {
    ingredients: '食材.xlsx',
    products: '产品.xlsx',
    dishes: '菜单.xlsx',
  };
  if (!contentDisposition) return defaultNames[type] || 'export.xlsx';
  const match = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
  if (match && match[1]) {
    return match[1].replace(/['"]/g, '');
  }
  return defaultNames[type] || 'export.xlsx';
}

// 新增方法
const openImportModal = (type: 'ingredients' | 'products') => {
  importType.value = type;
  showImportModal.value = true;
};

const closeImportModal = () => {
  showImportModal.value = false;
};

const handleImportSuccess = () => {
  // 可以在这里刷新列表数据（如果需要）
  // 例如重新获取食材或商品列表
  showToast('导入成功', 'success');
};
</script>

<style scoped>
/* 全局背景色 */
:global(body) {
  background-color: #f0f2f5;
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* 顶部导航卡片 */
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

.nav-item.router-link-exact-active {
  color: #409eff;
  background-color: #ecf5ff;
}

.nav-item.router-link-exact-active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 16px;
  right: 16px;
  height: 2px;
  background-color: #409eff;
  border-radius: 2px;
}

.util-btn {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  background-color: #ffffff;
  padding: 20px 24px;
  margin: 0 24px 24px 24px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.button-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 140px;
}

.export-btn {
  background-color: #409eff;
  border: none;
  color: white;
  padding: 10px 24px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s, box-shadow 0.2s, transform 0.1s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-width: 120px;
  box-shadow: 0 4px 6px rgba(64, 158, 255, 0.2);
  letter-spacing: 0.5px;
}

.export-btn:hover:not(:disabled) {
  background-color: #66b1ff;
  box-shadow: 0 6px 12px rgba(64, 158, 255, 0.3);
}

.export-btn:active:not(:disabled) {
  transform: scale(0.98);
}

.export-btn:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
  opacity: 0.7;
  box-shadow: none;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 8px;
  margin-bottom: 0;
  background-color: #fef0f0;
  padding: 4px 12px;
  border-radius: 16px;
  border-left: 2px solid #f56c6c;
}

.toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 30px;
  color: white;
  font-size: 14px;
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
  animation: slideDown 0.3s ease;
  text-align: center;
  min-width: 200px;
  backdrop-filter: blur(4px);
}

.toast.success {
  background-color: rgba(103, 194, 58, 0.95);
}

.toast.error {
  background-color: rgba(245, 108, 108, 0.95);
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

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background-color: white;
  border-radius: 16px;
  max-width: 800px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

/* 小屏幕适配 */
@media (max-width: 768px) {
  .top-nav {
    margin: 12px 16px 8px 16px;
    padding: 12px 16px;
  }

  .util-btn {
    margin: 0 16px 16px 16px;
    padding: 16px;
    gap: 16px;
  }

  .button-group {
    min-width: 100%;
  }

  .export-btn {
    width: 100%;
  }
}
</style>