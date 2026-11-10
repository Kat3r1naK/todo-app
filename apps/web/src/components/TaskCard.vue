<template>
  <article class="task-card" :class="{ pinned: task.pinned, done: task.status === 'done' }" @click="$emit('edit')">
    <div class="row">
      <button
        type="button"
        class="check"
        :class="{ checked: task.status === 'done' }"
        :title="task.status === 'done' ? '标记为未完成' : '标记为完成'"
        :aria-label="task.status === 'done' ? '标记为未完成' : '标记为完成'"
        :aria-pressed="task.status === 'done'"
        @click.stop="$emit('toggle')"
      >
        <el-icon v-if="task.status === 'done'" :size="11"><Check /></el-icon>
      </button>
      <h4 class="title"><button type="button" class="task-title" @click.stop="$emit('edit')">{{ task.title }}</button></h4>
      <div class="actions" @click.stop>
        <button
          class="icon-btn"
          :class="{ active: task.pinned }"
          :title="task.pinned ? '取消置顶' : '置顶'"
          :aria-label="task.pinned ? '取消置顶' : '置顶任务'"
          :aria-pressed="task.pinned"
          @click="$emit('pin')"
        >
          <el-icon :size="14">
            <StarFilled v-if="task.pinned" />
            <Star v-else />
          </el-icon>
        </button>
        <button
          v-if="task.status === 'done'"
          class="icon-btn"
          title="归档"
          aria-label="归档任务"
          @click="$emit('archive')"
        >
          <el-icon :size="14"><Box /></el-icon>
        </button>
        <button class="icon-btn danger" title="删除" aria-label="删除任务" @click="$emit('delete')">
          <el-icon :size="14"><Delete /></el-icon>
        </button>
      </div>
    </div>

    <p v-if="task.description" class="desc">{{ task.description }}</p>

    <div class="subtask-area" :class="{ nested: subtasks.length }">
      <div v-if="subtasks.length" class="subtask-line">
        <span class="subtask-text">子任务 {{ doneSubtasks }}/{{ subtasks.length }}</span>
        <span class="subtask-track"><i :style="{ width: `${subtaskPercent}%` }" /></span>
        <button
          v-if="!adding"
          type="button"
          class="subtask-add-trigger"
          title="添加子任务"
          aria-label="添加子任务"
          @click.stop="startAdd"
        >
          <el-icon :size="12"><Plus /></el-icon>
        </button>
      </div>
      <ul v-if="subtasks.length" class="subtask-list">
        <li v-for="item in subtasks" :key="item.id" class="subtask-item" :class="{ done: item.done }">
          <button
            type="button"
            class="subtask-check"
            :class="{ checked: item.done }"
            :title="item.done ? '标记为未完成' : '标记为完成'"
            :aria-label="item.done ? '标记子任务为未完成' : '标记子任务为完成'"
            :aria-pressed="item.done"
            @click.stop="$emit('toggle-subtask', item.id)"
          >
            <el-icon v-if="item.done" :size="9"><Check /></el-icon>
          </button>
          <span class="subtask-title">{{ item.title }}</span>
        </li>
      </ul>
      <button v-if="!subtasks.length && !adding" type="button" class="subtask-add-link" @click.stop="startAdd">
        <el-icon :size="11"><Plus /></el-icon>
        <span>添加子任务</span>
      </button>
      <div v-if="adding" class="subtask-add-row" @click.stop>
        <el-input
          ref="addInput"
          v-model="draft"
          size="small"
          maxlength="200"
          placeholder="子任务标题，回车添加"
          @keyup.enter.prevent="commitAdd"
          @keyup.esc="cancelAdd"
          @blur="cancelAdd"
        />
        <el-button size="small" :disabled="!draft.trim()" @mousedown.prevent @click="commitAdd">添加</el-button>
      </div>
    </div>

    <div class="foot">
      <span class="priority" :data-level="task.priority"><i />{{ priorityText }}</span>
      <span class="task-meta">
        <el-icon v-if="task.notes" :size="12" class="note-label" title="含备注"><Document /></el-icon>
        <el-icon v-if="task.pinned" :size="12" class="pin-label" title="已置顶"><StarFilled /></el-icon>
        <time :datetime="task.created_at">{{ timeText }}</time>
      </span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { Box, Check, Delete, Document, Plus, Star, StarFilled } from '@element-plus/icons-vue'
import type { Task } from '../types'

const props = defineProps<{ task: Task }>()

const emit = defineEmits<{
  edit: []
  delete: []
  pin: []
  toggle: []
  archive: []
  'toggle-subtask': [subtaskId: string]
  'add-subtask': [title: string]
}>()

const adding = ref(false)
const draft = ref('')
const addInput = ref<{ focus?: () => void } | null>(null)

function startAdd(): void {
  adding.value = true
  nextTick(() => addInput.value?.focus?.())
}

function cancelAdd(): void {
  adding.value = false
  draft.value = ''
}

function commitAdd(): void {
  const title = draft.value.trim()
  if (!title) return
  emit('add-subtask', title)
  draft.value = ''
  adding.value = false
}

const priorityText = computed(
  () => ({ high: '高优先级', medium: '普通', low: '低优先级' })[props.task.priority] || '普通',
)

const subtasks = computed(() => props.task.subtasks ?? [])
const doneSubtasks = computed(() => subtasks.value.filter((item) => item.done).length)
const subtaskPercent = computed(() =>
  subtasks.value.length ? Math.round((doneSubtasks.value / subtasks.value.length) * 100) : 0,
)

const timeText = computed(() => {
  const value = props.task.status === 'done' && props.task.completed_at ? props.task.completed_at : props.task.created_at
  if (!value) return ''
  return new Date(value).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
  })
})
</script>

<style scoped>
.task-card {
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--card);
  cursor: grab;
  transition: border-color 0.15s;
}

.task-card:hover,
.task-card:focus-within {
  border-color: var(--border-strong);
}

.task-card:active {
  cursor: grabbing;
}

.row {
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 9px;
}

.check {
  position: relative;
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 16px;
  height: 16px;
  margin-top: 4px;
  padding: 0;
  border: 1.5px solid var(--border-strong);
  border-radius: 50%;
  background: transparent;
  color: var(--card);
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.check::after {
  position: absolute;
  inset: -8px;
  border-radius: 50%;
  content: '';
}

.check:hover {
  border-color: var(--accent);
}

.check.checked {
  border-color: var(--accent);
  background: var(--accent);
}

.title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.task-card.done .title {
  color: var(--muted);
  text-decoration: line-through;
  text-decoration-color: var(--border-strong);
}

.task-card.done .desc {
  opacity: 0.7;
}

.task-title {
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.actions {
  display: flex;
  flex-shrink: 0;
  gap: 0;
  margin: -3px -6px 0 0;
  opacity: 0;
  transition: opacity 0.15s;
}

.task-card:hover .actions,
.task-card:focus-within .actions {
  opacity: 1;
}

.icon-btn.active {
  color: var(--pin);
}

.desc {
  display: -webkit-box;
  margin: 6px 0 0;
  overflow: hidden;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.subtask-area {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
}

.subtask-area.nested {
  padding-left: 10px;
}

.subtask-line {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--muted);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.subtask-add-trigger {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 18px;
  height: 18px;
  margin-left: auto;
  padding: 0;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}

.subtask-add-trigger:hover {
  background: var(--panel-2);
  color: var(--text);
}

.subtask-add-link {
  display: inline-flex;
  align-self: flex-start;
  align-items: center;
  gap: 4px;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--muted);
  font-size: 11px;
  cursor: pointer;
}

.subtask-add-link:hover {
  color: var(--text);
}

.subtask-add-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}

.subtask-add-row .el-input {
  flex: 1;
}

.subtask-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.subtask-item {
  display: flex;
  align-items: flex-start;
  gap: 7px;
  font-size: 12px;
  line-height: 1.5;
}

.subtask-check {
  position: relative;
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 14px;
  height: 14px;
  margin-top: 1px;
  padding: 0;
  border: 1.5px solid var(--border-strong);
  border-radius: 50%;
  background: transparent;
  color: var(--card);
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.subtask-check::after {
  position: absolute;
  inset: -7px;
  border-radius: 50%;
  content: '';
}

.subtask-check:hover {
  border-color: var(--accent);
}

.subtask-check.checked {
  border-color: var(--accent);
  background: var(--accent);
}

.subtask-title {
  min-width: 0;
  overflow-wrap: anywhere;
}

.subtask-item.done .subtask-title {
  color: var(--muted);
  text-decoration: line-through;
  text-decoration-color: var(--border-strong);
}

.subtask-text {
  flex-shrink: 0;
}

.subtask-track {
  width: 56px;
  height: 3px;
  overflow: hidden;
  border-radius: 2px;
  background: var(--border);
}

.subtask-track i {
  display: block;
  height: 100%;
  border-radius: 2px;
  background: var(--col-done);
  transition: width 0.15s;
}

.foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 14px;
  color: var(--muted);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.priority {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.priority i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--border-strong);
}

.priority[data-level='high'] i {
  background: #b96954;
}

.priority[data-level='high'] {
  color: #a45643;
}

.priority[data-level='low'] i {
  background: var(--col-done);
}

:global(html[data-theme='night']) .priority[data-level='high'] {
  color: #d69a88;
}

.task-meta {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.pin-label {
  color: var(--pin);
}

.note-label {
  color: var(--muted);
}

@media (hover: none) {
  .actions {
    opacity: 1;
  }
}
</style>
