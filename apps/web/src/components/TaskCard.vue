<template>
  <article class="task-card" :class="{ pinned: task.pinned }" @click="$emit('edit')">
    <div class="row">
      <span class="priority" :data-level="task.priority">{{ priorityText }}</span>
      <div class="actions" @click.stop>
        <button
          class="icon-btn"
          :class="{ active: task.pinned }"
          :title="task.pinned ? '取消置顶' : '置顶'"
          @click="$emit('pin')"
        >
          <el-icon :size="14">
            <StarFilled v-if="task.pinned" />
            <Star v-else />
          </el-icon>
        </button>
        <button class="icon-btn danger" title="删除" @click="$emit('delete')">
          <el-icon :size="14"><Delete /></el-icon>
        </button>
      </div>
    </div>

    <h4 class="title">{{ task.title }}</h4>
    <p v-if="task.description" class="desc">{{ task.description }}</p>

    <div class="foot">
      <span>{{ timeText }}</span>
      <span v-if="task.pinned" class="pin-label">
        <el-icon :size="12"><StarFilled /></el-icon>
        已置顶
      </span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Delete, Star, StarFilled } from '@element-plus/icons-vue'
import type { Task } from '../types'

const props = defineProps<{ task: Task }>()

defineEmits<{ edit: []; delete: []; pin: [] }>()

const priorityText = computed(
  () => ({ high: '高优先级', medium: '中优先级', low: '低优先级' })[props.task.priority] || '中优先级',
)

const timeText = computed(() => {
  if (!props.task.created_at) return ''
  return new Date(props.task.created_at).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
  })
})
</script>

<style scoped>
.task-card {
  padding: 12px 12px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--card);
  cursor: grab;
  transition: transform 0.12s ease, box-shadow 0.18s ease, border-color 0.18s ease;
}

.task-card:hover {
  transform: translateY(-1px);
  border-color: var(--border-strong);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.14);
}

.task-card:active {
  cursor: grabbing;
}

.task-card.pinned {
  box-shadow: inset 3px 0 0 var(--pin);
}

.row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.priority {
  padding: 2px 8px;
  border: 1px solid transparent;
  border-radius: 999px;
  font-size: 11px;
  color: var(--el-color-warning);
  background: rgba(217, 165, 46, 0.12);
  border-color: rgba(217, 165, 46, 0.3);
}

.priority[data-level='high'] {
  color: var(--el-color-danger);
  background: rgba(224, 108, 117, 0.12);
  border-color: rgba(224, 108, 117, 0.3);
}

.priority[data-level='low'] {
  color: var(--el-color-success);
  background: rgba(76, 175, 109, 0.12);
  border-color: rgba(76, 175, 109, 0.3);
}

.actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.task-card:hover .actions {
  opacity: 1;
}

.icon-btn.active {
  color: var(--pin);
}

@media (hover: none) {
  .actions {
    opacity: 1;
  }
}

.title {
  margin: 9px 0 0;
  font-size: 14px;
  line-height: 1.45;
  word-break: break-word;
}

.desc {
  display: -webkit-box;
  margin: 6px 0 0;
  overflow: hidden;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.foot {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  color: var(--muted);
  font-size: 11px;
}

.pin-label {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  color: var(--pin);
}
</style>
