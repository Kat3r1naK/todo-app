<template>
  <section class="kanban-col" :class="`col-${column.id}`">
    <header class="col-head">
      <div class="col-name">
        <span class="dot" />
        <h3>{{ column.label }}</h3>
        <span class="count">{{ tasks.length }}</span>
      </div>
      <el-button text :icon="Plus" :aria-label="`在${column.label}中新建任务`" @click="$emit('add')" />
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
          @toggle="$emit('toggle', element)"
          @archive="$emit('archive', element)"
          @toggle-subtask="(subtaskId: string) => $emit('toggle-subtask', element, subtaskId)"
          @add-subtask="(title: string) => $emit('add-subtask', element, title)"
        />
      </template>
      <template #footer>
        <div v-if="tasks.length === 0" class="drop-empty">
          <p>{{ disabled ? '没有匹配的任务' : '暂无任务' }}</p>
          <span v-if="!disabled" class="empty-hint">点击 + 新建，或把任务拖到这里</span>
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

defineEmits<{
  add: []
  edit: [task: Task]
  delete: [task: Task]
  pin: [task: Task]
  toggle: [task: Task]
  archive: [task: Task]
  'toggle-subtask': [task: Task, subtaskId: string]
  'add-subtask': [task: Task, title: string]
  dragend: []
}>()
</script>

<style scoped>
.kanban-col {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 360px;
  padding: 12px 10px 10px;
  border-radius: var(--radius-lg);
  background: var(--panel);
  scroll-snap-align: start;
}

.col-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 30px;
  margin: 0 4px 12px;
}

.col-head .el-button {
  width: 28px;
  height: 28px;
  padding: 0;
  color: var(--muted);
}

.col-name {
  display: flex;
  align-items: center;
  gap: 9px;
}

.dot {
  width: 9px;
  height: 9px;
  border: 1.5px solid var(--col-todo);
  border-radius: 50%;
}

.col-doing .dot {
  border-color: var(--col-doing);
  background: linear-gradient(90deg, var(--col-doing) 50%, transparent 50%);
}

.col-done .dot {
  border-color: var(--col-done);
  background: var(--col-done);
}

.col-name h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.count {
  margin-left: 3px;
  color: var(--muted);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.task-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-height: 96px;
}

.drop-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 8px 28px;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.drop-empty p {
  margin: 0 0 4px;
}

.empty-hint {
  font-size: 12px;
  opacity: 0.75;
}

.add-task {
  display: flex;
  align-items: center;
  gap: 7px;
  width: 100%;
  margin-top: 8px;
  padding: 8px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--muted);
  font-size: 12px;
  text-align: left;
  cursor: pointer;
  transition: background-color 0.15s, color 0.15s;
}

.add-task:hover {
  background: var(--panel-2);
  color: var(--text);
}
</style>
