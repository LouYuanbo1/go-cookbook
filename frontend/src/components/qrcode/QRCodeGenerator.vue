<template>
  <div class="qrcode-container">
    <!-- 生成时的加载遮罩 -->
    <div v-if="generating" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>正在生成二维码...</p>
    </div>

    <!-- 主卡片 -->
    <div class="qrcode-main-card">
      <div class="qrcode-header">
        <h1 class="qrcode-title">二维码生成器</h1>
        <div class="qrcode-subheader">
          <span class="qrcode-hint">输入文本或链接，生成二维码</span>
        </div>
      </div>

      <form @submit.prevent="handleGenerate" class="qrcode-form">
        <!-- 输入区域 -->
        <div class="form-group">
          <label class="form-label">
            二维码内容 <span class="required-star">*</span>
          </label>
          <input
            v-model="url"
            type="text"
            class="form-control"
            placeholder="请输入文本或URL"
            :disabled="generating"
            @input="clearUrlError"
          />
          <div class="file-hint">支持任意文本，长度建议不超过500字符</div>
          <div v-if="urlError" class="error-message">{{ urlError }}</div>
        </div>

        <!-- 二维码展示区域（仅生成成功后显示） -->
        <div v-if="qrcodeImageUrl" class="result-panel">
          <h3 class="result-title">生成结果</h3>
          <div class="qrcode-image-wrapper">
            <img :src="qrcodeImageUrl" alt="二维码" class="qrcode-image" />
          </div>
          <div class="result-item">
            <span class="result-label">内容：</span>
            <span class="result-value qrcode-content-preview">{{ urlPreview }}</span>
          </div>
          <div class="result-actions">
            <button type="button" class="btn btn-secondary btn-small" @click="downloadQRCode">
              下载二维码
            </button>
          </div>
        </div>

        <!-- 表单操作按钮（增加取消按钮） -->
        <div class="form-actions">
          <!-- 新增取消按钮，点击关闭弹窗 -->
          <button type="button" class="btn btn-secondary" @click="$emit('close')">取消</button>
          <button type="button" class="btn btn-secondary" @click="resetForm" :disabled="generating">
            重置
          </button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="generating || !isFormValid"
          >
            {{ generating ? "生成中..." : "生成二维码" }}
          </button>
        </div>
      </form>
    </div>

    <!-- 全局提示 Toast (Teleport) -->
    <Teleport to="body">
      <div v-if="showSuccessToast" class="toast-message success">
        <span>二维码生成成功！</span>
      </div>
      <div v-if="showErrorToast" class="toast-message error">
        <span>{{ errorToastMessage }}</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue';
import request from '../../api/request';

// ---------- 新增：定义 emits ----------
const emit = defineEmits<{
  (e: 'close'): void;
}>();

// ---------- 响应式状态 ----------
const url = ref('');
const urlError = ref('');
const generating = ref(false);
const showSuccessToast = ref(false);
const showErrorToast = ref(false);
const errorToastMessage = ref('');
const qrcodeImageUrl = ref<string | null>(null);

// ---------- 计算属性 ----------
const isFormValid = computed(() => {
  return url.value.trim() !== '' && !urlError.value;
});

// 预览内容（过长截断）
const urlPreview = computed(() => {
  const val = url.value.trim();
  return val.length > 40 ? val.slice(0, 40) + '…' : val;
});

// ---------- 清理 ----------
const clearUrlError = () => {
  urlError.value = '';
};

// 重置表单
const resetForm = () => {
  url.value = '';
  urlError.value = '';
  if (qrcodeImageUrl.value) {
    URL.revokeObjectURL(qrcodeImageUrl.value);
    qrcodeImageUrl.value = null;
  }
};

// 监听内容变化，清除之前生成的二维码（可选）
watch(url, () => {
  if (qrcodeImageUrl.value) {
    URL.revokeObjectURL(qrcodeImageUrl.value);
    qrcodeImageUrl.value = null;
  }
});

// ---------- 二维码生成 ----------
const handleGenerate = async () => {
  if (!isFormValid.value || generating.value) return;

  // 基本内容校验
  const content = url.value.trim();
  if (content === '') {
    urlError.value = '内容不能为空';
    return;
  }
  // 简单长度限制
  if (content.length > 1000) {
    urlError.value = '内容过长，请控制在1000字符以内';
    return;
  }

  generating.value = true;
  showErrorToast.value = false;

  // 释放之前生成的图片URL
  if (qrcodeImageUrl.value) {
    URL.revokeObjectURL(qrcodeImageUrl.value);
    qrcodeImageUrl.value = null;
  }

  try {
    // 调用后端接口，使用 POST 方法，参数通过 query 传递
    const response = await request({
      url: '/api/qrcode',
      method: 'POST',
      params: { url: content },  // 后端通过 Query 获取
      responseType: 'blob',       // 期望返回图片二进制
      rawResponse: true,           // 返回完整响应，用于判断 content-type
    });

    // 检查响应类型
    const contentType = response.headers['content-type'];
    if (contentType && contentType.includes('application/json')) {
      // 错误情况：后端返回了 JSON 错误
      const blob = response.data as Blob;
      const errorText = await blob.text();
      let errorMsg = '生成失败';
      try {
        const errorData = JSON.parse(errorText);
        errorMsg = errorData.error || errorData.msg || errorMsg;
      } catch (e) {
        errorMsg = errorText || errorMsg;
      }
      throw new Error(errorMsg);
    }

    // 成功，创建对象 URL
    const blob = response.data as Blob;
    if (blob.size === 0) {
      throw new Error('生成的图片为空');
    }
    qrcodeImageUrl.value = URL.createObjectURL(blob);

    // 显示成功提示
    showSuccessToast.value = true;
    setTimeout(() => {
      showSuccessToast.value = false;
    }, 2000);
  } catch (error: any) {
    console.error('二维码生成失败:', error);
    errorToastMessage.value = error.message || '生成失败，请稍后重试';
    showErrorToast.value = true;
    setTimeout(() => {
      showErrorToast.value = false;
    }, 3000);
  } finally {
    generating.value = false;
  }
};

// ---------- 下载二维码 ----------
const downloadQRCode = () => {
  if (!qrcodeImageUrl.value) return;
  const link = document.createElement('a');
  link.href = qrcodeImageUrl.value;
  link.download = `qrcode_${Date.now()}.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// 组件卸载前释放 blob URL
onBeforeUnmount(() => {
  if (qrcodeImageUrl.value) {
    URL.revokeObjectURL(qrcodeImageUrl.value);
  }
});
</script>

<style scoped>

.qrcode-container {
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

.qrcode-main-card {
  background-color: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  display: flex;
  flex-direction: column;
  padding: 24px;
}

.qrcode-header {
  margin-bottom: 24px;
  flex-shrink: 0;
}

.qrcode-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #222;
  line-height: 1.3;
  word-break: break-word;
}

.qrcode-subheader {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.qrcode-hint {
  font-size: 14px;
  color: #666;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.qrcode-form {
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

.btn-small {
  padding: 6px 16px;
  font-size: 13px;
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

/* 二维码专属样式 */
.qrcode-image-wrapper {
  display: flex;
  justify-content: center;
  margin: 20px 0;
  padding: 16px;
  background-color: white;
  border-radius: 8px;
  box-shadow: inset 0 2px 6px rgba(0, 0, 0, 0.03);
}

.qrcode-image {
  max-width: 100%;
  width: 200px;
  height: auto;
  image-rendering: crisp-edges; /* 保持二维码清晰 */
  border: 1px solid #eee;
  border-radius: 8px;
}

.qrcode-content-preview {
  word-break: break-all;
  flex: 1;
  line-height: 1.5;
}

.result-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  gap: 12px;
}

/* 响应式调整 */
@media (max-width: 639px) {
  .qrcode-main-card {
    padding: 16px;
  }
  .qrcode-title {
    font-size: 20px;
  }
  .result-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  .result-label {
    width: auto;
  }
  .qrcode-image {
    width: 160px;
  }
}

@media (min-width: 768px) {
  .qrcode-main-card {
    padding: 32px;
  }
  .qrcode-title {
    font-size: 28px;
  }
}

@media (min-width: 1024px) {
  .qrcode-container {
    padding: 24px;
  }
  .qrcode-title {
    font-size: 32px;
  }
}

@media print {
  .qrcode-container {
    padding: 0;
    background-color: white;
  }
  .qrcode-main-card {
    box-shadow: none;
    border: 1px solid #ddd;
  }
  .form-actions,
  .btn,
  .toast-message {
    display: none;
  }
}
</style>