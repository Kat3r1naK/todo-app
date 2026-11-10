<template>
  <el-dialog
    :model-value="modelValue"
    :title="task ? '编辑任务' : '新建任务'"
    width="560px"
    top="8vh"
    class="task-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" maxlength="200" show-word-limit placeholder="要做什么？" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          maxlength="2000"
          show-word-limit
          placeholder="补充说明（可选）"
        />
      </el-form-item>
      <el-form-item label="子任务">
        <div class="subtask-editor">
          <div
            v-for="(item, index) in form.subtasks"
            :key="item.id"
            class="subtask-row"
            :class="{ done: item.done }"
          >
            <el-checkbox v-model="item.done" size="small" class="subtask-check" />
            <el-input v-model="item.title" class="subtask-input" maxlength="200" />
            <button type="button" class="subtask-remove" aria-label="删除子任务" title="删除子任务" @click="removeSubtask(index)">
              <el-icon :size="13"><Close /></el-icon>
            </button>
          </div>
          <div class="subtask-add">
            <el-input
              v-model="newSubtask"
              maxlength="200"
              placeholder="添加子任务，回车确认"
              @keyup.enter.prevent="addSubtask"
            />
            <el-button :disabled="!newSubtask.trim()" @click="addSubtask">添加</el-button>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="备注">
        <div class="notes-editor">
          <div class="notes-bar">
            <span class="notes-hint">支持 Markdown</span>
            <div class="notes-tabs">
              <button type="button" :class="{ active: notesMode === 'write' }" @click="notesMode = 'write'">编辑</button>
              <button type="button" :class="{ active: notesMode === 'preview' }" @click="notesMode = 'preview'">预览</button>
            </div>
          </div>
          <el-input
            v-if="notesMode === 'write'"
            v-model="form.notes"
            type="textarea"
            :rows="6"
            maxlength="20000"
            placeholder="记录排查过程、解决方案、代码片段…支持 Markdown"
          />
          <div v-else class="notes-preview">
            <MarkdownView v-if="form.notes.trim()" :source="form.notes" />
            <p v-else class="notes-empty">暂无内容</p>
          </div>
        </div>
      </el-form-item>
      <div class="form-row">
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" style="width: 100%">
            <el-option v-for="item in statusOptions" :key="item.id" :label="item.label" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="form.priority" style="width: 100%">
            <el-option label="高" value="high" />
            <el-option label="中" value="medium" />
            <el-option label="低" value="low" />
          </el-select>
        </el-form-item>
      </div>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="submit">
        {{ task ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { Close } from '@element-plus/icons-vue'
import { COLUMNS } from '../stores/board'
import MarkdownView from './MarkdownView.vue'
import type { Subtask, Task, TaskPayload, TaskPriority, TaskStatus } from '../types'

interface Props {
  modelValue: boolean
  task: Task | null
  defaultStatus: TaskStatus
  saving: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [payload: TaskPayload]
}>()

const formRef = ref<FormInstance>()
const statusOptions = COLUMNS
const notesMode = ref<'write' | 'preview'>('write')
const newSubtask = ref('')
const form = reactive<{
  title: string
  description: string
  notes: string
  subtasks: Subtask[]
  status: TaskStatus
  priority: TaskPriority
}>({
  title: '',
  description: '',
  notes: '',
  subtasks: [],
  status: 'todo',
  priority: 'medium',
})
const rules: FormRules = {
  title: [{ required: true, message: '请输入任务标题', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    form.title = props.task?.title || ''
    form.description = props.task?.description || ''
    form.notes = props.task?.notes || ''
    form.subtasks = (props.task?.subtasks ?? []).map((item) => ({ ...item }))
    form.status = props.task?.status || props.defaultStatus
    form.priority = props.task?.priority || 'medium'
    notesMode.value = 'write'
    newSubtask.value = ''
    nextTick(() => formRef.value?.clearValidate())
  },
)

function makeSubtaskID(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

function addSubtask(): void {
  const title = newSubtask.value.trim()
  if (!title) return
  form.subtasks.push({ id: makeSubtaskID(), title, done: false })
  newSubtask.value = ''
}

function removeSubtask(index: number): void {
  form.subtasks.splice(index, 1)
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  emit('submit', {
    title: form.title.trim(),
    description: form.description.trim(),
    notes: form.notes.trim(),
    subtasks: form.subtasks
      .map((item) => ({ ...item, title: item.title.trim() }))
      .filter((item) => item.title),
    status: form.status,
    priority: form.priority,
  })
}
</script>

<style scoped>
.subtask-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.subtask-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.subtask-check {
  flex-shrink: 0;
  margin-right: 0;
}

.subtask-input {
  flex: 1;
}

.subtask-input :deep(.el-input__wrapper) {
  padding: 0 4px;
  background: transparent;
  box-shadow: none;
}

.subtask-input :deep(.el-input__wrapper:hover),
.subtask-input :deep(.el-input__wrapper.is-focus) {
  background: var(--card);
  box-shadow: 0 0 0 1px var(--border) inset;
}

.subtask-row.done .subtask-input :deep(.el-input__inner) {
  color: var(--muted);
  text-decoration: line-through;
}

.subtask-remove {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}

.subtask-remove:hover {
  background: var(--panel-2);
  color: var(--el-color-danger);
}

.subtask-add {
  display: flex;
  gap: 8px;
  margin-top: 2px;
}

.subtask-add .el-input {
  flex: 1;
}

.notes-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.notes-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.notes-hint {
  color: var(--muted);
  font-size: 12px;
}

.notes-tabs {
  display: flex;
  gap: 2px;
  padding: 2px;
  border-radius: 5px;
  background: var(--panel);
}

.notes-tabs button {
  padding: 3px 10px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--muted);
  font-size: 12px;
  cursor: pointer;
}

.notes-tabs button.active {
  background: var(--card);
  color: var(--text);
}

.notes-preview {
  min-height: 120px;
  max-height: 240px;
  padding: 10px 12px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--card);
}

.notes-empty {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

@media (max-width: 480px) {
  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
}
</style>
