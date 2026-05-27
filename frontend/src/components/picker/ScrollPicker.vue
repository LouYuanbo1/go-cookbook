<template>
  <div class="scroll-picker">
    <div class="scroll-wrapper" ref="scrollContainerRef" @scroll="handleScroll">
      <!-- 初始加载 -->
      <div v-if="loading && !items.length" class="status-indicator">
        {{ loadingText }}
      </div>

      <!-- 空状态 -->
      <div v-else-if="!items.length && !hasMore" class="status-indicator">
        {{ emptyText }}
      </div>

      <!-- 列表项（使用插槽自定义渲染） -->
      <div v-else class="items-grid">
        <div
          v-for="(item, index) in items"
          :key="getItemKey(item, index)"
          class="item-card"
          @click="$emit('select', item)"
        >
          <slot name="item" :item="item" :index="index">
            <!-- 默认插槽：显示简单文本 -->
            <div class="default-item">{{ JSON.stringify(item) }}</div>
          </slot>
        </div>

        <!-- 加载更多指示器 -->
        <div v-if="loadingMore" class="status-indicator more">
          {{ loadingMoreText }}
        </div>
        <div v-else-if="!hasMore && items.length > 0" class="status-indicator end">
          {{ noMoreText }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted/*, onUnmounted, nextTick*/ } from 'vue';

// ---------- 类型定义 ----------
// 分页数据结果接口
export interface FetchResult<T = any> {
  items: T[];
  cursor: number;      // 下一页游标（0 表示没有下一页）
  hasMore: boolean;
}

// 分页数据获取函数类型
export type FetchFunction<T = any> = (
  cursor: number,
  limit: number
) => Promise<FetchResult<T>>;

// ---------- Props ----------
/* 组件 props 定义
以“:”传入，props代表父组件传递给子组件的属性

e.g.
<!-- 父组件 -->
<template>
  <ScrollPicker 
    :fetch="loadData"          <!-- Props: 传入加载函数 -->
    :limit="15"             <!-- Props: 传入每页数量 -->
    @select="onItemSelected"   <!-- Emit: 监听选择事件 -->
    v-model:items="allItems"   <!-- 双向绑定（语法糖） -->
  />
</template>


*/
const props = withDefaults(
  defineProps<{
    //这里类似于回调函数，子组件通过调用fetch函数来获取数据
    fetch: FetchFunction;               // 必传：分页数据获取函数
    itemKey?: string | ((item: any, index: number) => string); // 唯一键生成
    limit?: number;                    // 每页数量，默认 15
    loadingText?: string;             // 初始加载文案
    loadingMoreText?: string;         // 加载更多文案
    emptyText?: string;              // 空数据文案
    noMoreText?: string;            // 无更多文案
    scrollThreshold?: number;       // 触发加载更多的滚动距离阈值，默认 50px
  }>(),
  {
    itemKey: 'id',
    limit: 15,
    loadingText: '加载中...',
    loadingMoreText: '加载更多...',
    emptyText: '暂无数据',
    noMoreText: '没有更多了',
    scrollThreshold: 50,
  }
);

// ---------- 事件 ----------
/* 组件事件定义
以“@”传入，emit代表子组件触发的事件
后面的参数为事件触发时传递的参数
*/

/*
(e: 'select', item: any): void 这行声明意味着：
子组件可以通过 emit('select', item) 将内部的 item 数据作为参数传递给父组件，父组件通过监听 @select 事件接收这个数据。
*/

const emit = defineEmits<{
  (e: 'select', item: any): void;     // 点击卡片时触发
  (e: 'update:items', items: any[]): void; // 可选：同步当前列表数据
}>();

// ---------- 状态 ----------
const items = ref<any[]>([]);
const cursor = ref<number>(0);
const hasMore = ref<boolean>(true);
const loading = ref<boolean>(false);
const loadingMore = ref<boolean>(false);
const scrollContainerRef = ref<HTMLElement | null>(null);

// 生成唯一 key
const getItemKey = (item: any, index: number): string => {
  if (typeof props.itemKey === 'function') {
    return props.itemKey(item, index);
  }
  return item[props.itemKey] ?? index.toString();
};

// ---------- 加载初始数据 ----------
const loadInitial = async () => {
  loading.value = true;
  try {
    const result = await props.fetch(0, props.limit);
    items.value = result.items || [];
    cursor.value = result.cursor || 0;
    hasMore.value = result.hasMore || false;
    //触发更新父组件的items
    emit('update:items', items.value);
  } catch (error) {
    console.error('ScrollPicker 初始加载失败:', error);
    items.value = [];
    hasMore.value = false;
  } finally {
    loading.value = false;
  }
};

// ---------- 加载更多数据 ----------
const loadMore = async () => {
  if (!hasMore.value || loadingMore.value || loading.value) return;
  loadingMore.value = true;
  try {
    const result = await props.fetch(cursor.value, props.limit);
    const newItems = result.items || [];
    items.value.push(...newItems);
    cursor.value = result.cursor || 0;
    hasMore.value = result.hasMore || false;
    //触发更新父组件的items
    emit('update:items', items.value);
  } catch (error) {
    console.error('ScrollPicker 加载更多失败:', error);
  } finally {
    loadingMore.value = false;
  }
};

// ---------- 滚动监听 ----------
const handleScroll = () => {
  const container = scrollContainerRef.value;
  if (!container) return;
  const scrollBottom = container.scrollHeight - container.scrollTop - container.clientHeight;
  if (scrollBottom < props.scrollThreshold && hasMore.value && !loadingMore.value && !loading.value) {
    loadMore();
  }
};

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadInitial,
  loadMore,
  items,
});

// 初始化
onMounted(() => {
  loadInitial();
});

// 清理（如果未来需要监听某些变化）
// onUnmounted(() => {});
</script>

<style scoped>
.scroll-picker {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: inherit;
}

.scroll-wrapper {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  padding: 8px 4px;
  box-sizing: border-box;
}

.items-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

/* 响应式网格（调用方可覆盖） */
@media (min-width: 640px) {
  .items-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (min-width: 1024px) {
  .items-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.item-card {
  cursor: pointer;
  transition: all 0.2s ease;
}

.status-indicator {
  padding: 24px 16px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

.status-indicator.more,
.status-indicator.end {
  padding: 16px;
  grid-column: 1 / -1;
}

.default-item {
  padding: 16px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  background: white;
  word-break: break-all;
}
</style>