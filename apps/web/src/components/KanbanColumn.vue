<template>
  <section class="kanban-col" :class="`col-${column.id}`">
    <header class="col-head">
      <div class="col-name">
        <span class="dot" />
        <h3>{{ column.label }}</h3>
        <span class="count">{{ tasks.length }}</span>
      </div>
      <el-button text circle :icon="Plus" @click="$emit('add')" />
    </header>

    <draggable
      class="task-list"
      :list="tasks"
      item-key="id"
      group="tasks"
      :animation="170"
      :disabled="disabled"
      ghost-class="task-ghost"
      @end="$emit('dragend')"
    >
      <template #item="{ element }">
        <TaskCard
          :task="element"
          @edit="$emit('edit', element)"
          @delete="$emit('delete', element)"
          @pin="$emit('pin', element)"
        />
      </template>
      <template #footer>
        <div v-if="tasks.length === 0" class="drop-empty">
          {{ disabled ? '没有匹配的任务' : '暂无任务，拖拽或点击 + 添加' }}
        </div>
      </template>
    </draggable>

    <button class="add-task" @click="$emit('add')">
      <el-icon><Plus /></el-icon>
      添加任务
    </button>
  </section>
</template>

<script setup lang="ts">
import draggable from 'vuedraggable'
import { Plus } from '@element-plus/icons-vue'
import TaskCard from './TaskCard.vue'
import type { BoardColumn, Task } from '../types'

withDefaults(
  defineProps<{
    column: BoardColumn
    tasks: Task[]
    disabled?: boolean
  }>(),
  { disabled: false },
)

defineEmits<{ add: []; edit: [task: Task]; delete: [task: Task]; pin: [task: Task]; dragend: [] }>()
</script>

<style scoped>
.kanban-col {
  display: flex;
  flex-direction: column;
  min-height: 220px;
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--panel);
}

.col-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.col-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
}

.col-todo .dot {
  background: var(--col-todo);
  box-shadow: 0 0 10px var(--col-todo);
}

.col-doing .dot {
  background: var(--col-doing);
  box-shadow: 0 0 10px var(--col-doing);
}

.col-done .dot {
  background: var(--col-done);
  box-shadow: 0 0 10px var(--col-done);
}

.col-name h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.count {
  padding: 1px 8px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--panel-2);
  color: var(--muted);
  font-size: 11px;
}

.task-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 10px;
  min-height: 96px;
}

.drop-empty {
  padding: 24px 10px;
  border: 1px dashed var(--border-strong);
  border-radius: var(--radius-md);
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.add-task {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  margin-top: 12px;
  padding: 9px;
  border: 1px dashed var(--border-strong);
  border-radius: 10px;
  background: transparent;
  color: var(--muted);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.add-task:hover {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
</style>
