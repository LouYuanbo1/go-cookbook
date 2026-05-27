<template>
  <div class="excel-import-container">
    <!-- 提交时的加载遮罩 -->
    <div v-if="submitting" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在上传并解析Excel文件...</p>
    </div>

    <!-- 主卡片 -->
    <div class="excel-import-main-card">
      <div class="excel-import-header">
        <h1 class="excel-import-title">Excel导入解析</h1>
        <div class="excel-import-subheader">
          <span class="excel-import-hint">上传Excel表格，后端将自动解析</span>
        </div>
      </div>

      <form @submit.prevent="handleSubmit" class="excel-import-form">
        <!-- 文件选择区域 -->
        <div class="form-group">
          <label class="form-label">
            Excel文件 <span class="required-star">*</span>
          </label>
          <div class="file-select-row">
            <button
              type="button"
              class="btn btn-secondary file-select-btn"
              @click="triggerFileInput"
              :disabled="submitting"
            >
              选择文件
            </button>
            <span class="file-name">{{ fileName || '未选择任何文件' }}</span>
            <!-- 隐藏的原始file input -->
            <input
              ref="fileInput"
              type="file"
              accept=".xlsx, .xls, .csv"
              @change="onFileChange"
              style="display: none"
            />
          </div>
          <div class="file-hint">支持 .xlsx, .xls, .csv 格式，最大20MB</div>
          <div v-if="fileError" class="error-message">{{ fileError }}</div>
        </div>

        <!-- 额外的描述字段 -->
        <!--
        <div class="form-group">
          <label class="form-label">解析备注（可选）</label>
          <textarea
            v-model="remark"
            class="form-control textarea"
            rows="2"
            placeholder="可添加一些说明，例如数据来源、注意事项等"
          ></textarea>
        </div>
        -->

        <!-- 解析结果展示区域（仅在成功返回数据后显示） -->
        <div v-if="result" class="result-panel">
          <h3 class="result-title">解析结果</h3>
          <div class="result-item">
            <span class="result-label">导入总数：</span>
            <span class="result-value">{{ result.importedCount || 0 }}</span>
          </div>
          <div class="result-item">
            <span class="result-label">成功解析：</span>
            <span class="result-value">{{ result.successCount || 0 }}</span>
          </div>
          <div v-if="result.errors && result.errors.length" class="result-errors">
            <div class="result-label">错误详情：</div>
            <ul class="error-list">
              <li v-for="(err, idx) in result.errors.slice(0, 3)" :key="idx" class="error-text">
                {{ err }}
              </li>
              <li v-if="result.errors.length > 3" class="error-text">
                等{{ result.errors.length }}条错误...
              </li>
            </ul>
          </div>
        </div>

        <!-- 表单操作按钮 -->
        <div class="form-actions">
          <!-- 取消按钮：改为普通按钮，发射 close 事件 -->
          <button type="button" class="btn btn-secondary" @click="$emit('close')">取消</button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !isFormValid"
          >
            {{ submitting ? "解析中..." : "开始解析" }}
          </button>
        </div>
      </form>
    </div>

    <!-- 全局提示 Toast (Teleport) -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>文件上传成功，解析已完成！</span>
      </div>
      <div v-if="showErrorToast" class="toast-message error">
        <span>{{ errorToastMessage }}</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import request from '../../api/request';

const props = defineProps<{
  type: string;
  batchSize?: number; // 改为可选，默认200
}>();

// 定义 emits
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'import-success', result: any): void;
  (e: 'import-error', error: any): void;
}>();

// ---------- 类型定义 ----------
interface ParseResult {
  importedCount: number;   // 导入的记录数
  successCount: number;    // 成功解析数
  errors?: string[];       // 可能的错误列表
}

// ---------- 响应式状态 ----------
const fileInput = ref<HTMLInputElement | null>(null);
const selectedFile = ref<File | null>(null);
const fileName = computed(() => selectedFile.value?.name || '');
const fileError = ref('');
//const remark = ref('');
const submitting = ref(false);
const showSuccessToast = ref(false);
const showErrorToast = ref(false);
const errorToastMessage = ref('');
const result = ref<ParseResult | null>(null);

// ---------- 表单有效性验证 ----------
const isFormValid = computed(() => {
  return !!selectedFile.value && !fileError.value;
});

// ---------- 监听文件变化时清除旧错误和结果 ----------
watch(selectedFile, () => {
  fileError.value = '';
  result.value = null;
});

// ---------- 触发隐藏的 file input ----------
const triggerFileInput = () => {
  if (fileInput.value) {
    fileInput.value.click();
  }
};

// ---------- 文件选择处理：校验类型和大小 ----------
const onFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  const files = target.files;
  if (!files || files.length === 0) {
    selectedFile.value = null;
    return;
  }

  const file = files[0];
  if (!file) {
    selectedFile.value = null;
    return;
  }

  const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
  if (!['.xlsx', '.xls', '.csv'].includes(extension)) {
    fileError.value = '文件格式不支持，请上传 .xlsx, .xls 或 .csv 文件';
    selectedFile.value = null;
    if (fileInput.value) fileInput.value.value = '';
    return;
  }

  const maxSize = 20 * 1024 * 1024; // 20MB
  if (file.size > maxSize) {
    fileError.value = '文件大小不能超过20MB';
    selectedFile.value = null;
    if (fileInput.value) fileInput.value.value = '';
    return;
  }

  selectedFile.value = file;
  fileError.value = '';
};

// ---------- 表单提交：上传到 Go 后端解析 ----------
const handleSubmit = async () => {
  if (!isFormValid.value || submitting.value) return;

  const file = selectedFile.value;
  if (!file) return;

  submitting.value = true;
  showErrorToast.value = false;

  const formData = new FormData();
  formData.append('file', file);
  /*
  if (remark.value.trim()) {
    formData.append('remark', remark.value.trim());
  }
  */

  try {
    const response = await request({
      url: `/api/${props.type}/import`,
      method: 'POST',
      data: formData,
      params: {
        batchSize: props.batchSize ?? 200,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      rawResponse: true,
    });
    /*
    const responseData = response.data;
    result.value = {
      importedCount: responseData.data?.importedCount || 0,
      successCount: responseData.data?.successCount || 0,
      errors: responseData.data?.errors || [],
    };
    */
    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
    }, 2000);
    // 发射成功事件和关闭事件（可根据需要决定是否自动关闭）
    emit('import-success', response.msg);
    emit('close'); // 成功自动关闭弹窗
    
  } catch (error: any) {
    console.error('上传解析失败:', error);
    errorToastMessage.value = error.message || '上传失败，请稍后重试';
    showErrorToast.value = true;
    setTimeout(() => {
      showErrorToast.value = false;
    }, 3000);
    
    // 发射错误事件
    emit('import-error', error);
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
/* 样式保持不变，与原来一致 */
.excel-import-container {
  font-family: "PingFang SC", "Helvetica Neue", Arial, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
  color: #333;
  background-color: #f9f9f9;
  /*min-height: 100vh;*/
  box-sizing: border-box;
}

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

.excel-import-main-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.excel-import-header {
  margin-bottom: 24px;
  flex-shrink: 0;
}

.excel-import-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.excel-import-subheader {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.excel-import-hint {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.excel-import-form {
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
  min-height: 80px;
  line-height: 1.6;
}

.error-message {
  font-size: 13px;
  color: #e53935;
  margin-top: 4px;
}

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

.btn-secondary:hover:not(:disabled) {
  background-color: #f5f5f5;
  border-color: #ccc;
  color: #333;
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toast-message {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 30px;
  color: white;
  font-size: 15px;
  font-weight: 500;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  animation: slideDown 0.3s ease;
}

.toast-message.success {
  background-color: #4caf50;
}

.toast-message.error {
  background-color: #e53935;
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

.file-select-row {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.file-select-btn {
  flex-shrink: 0;
  min-width: 100px;
}

.file-name {
  font-size: 14px;
  color: #666;
  word-break: break-all;
  background-color: #f5f5f5;
  padding: 6px 12px;
  border-radius: 20px;
  border: 1px solid #eee;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-hint {
  font-size: 13px;
  color: #999;
  margin-top: 2px;
}

.result-panel {
  background-color: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
  margin-top: 8px;
  border: 1px solid #e9ecef;
}

.result-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #2c3e50;
  padding-bottom: 8px;
  border-bottom: 1px dashed #ddd;
}

.result-item {
  display: flex;
  align-items: baseline;
  margin-bottom: 8px;
  font-size: 15px;
}

.result-label {
  width: 100px;
  color: #666;
  font-weight: 500;
}

.result-value {
  color: #2c3e50;
  font-weight: 600;
}

.result-errors {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #eee;
}

.error-list {
  margin: 8px 0 0 0;
  padding-left: 20px;
  color: #e53935;
  font-size: 13px;
}

.error-text {
  margin-bottom: 4px;
}

@media (max-width: 639px) {
  .file-select-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .file-name {
    max-width: 100%;
    white-space: normal;
    word-break: break-all;
  }
  .result-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  .result-label {
    width: auto;
  }
}

@media (min-width: 768px) {
  .excel-import-main-card {
    padding: 32px;
  }
  .excel-import-title {
    font-size: 28px;
  }
}

@media (min-width: 1024px) {
  .excel-import-container {
    padding: 24px;
  }
  .excel-import-title {
    font-size: 32px;
  }
}

@media print {
  .excel-import-container {
    padding: 0;
    background-color: white;
  }
  .excel-import-main-card {
    box-shadow: none;
    border: 1px solid #ddd;
  }
  .form-actions,
  .file-select-btn,
  .toast-message {
    display: none;
  }
}
</style>