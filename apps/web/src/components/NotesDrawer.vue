<template>
  <el-drawer
    :model-value="modelValue"
    title="便签"
    size="min(420px, 92vw)"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="notes">
      <div class="compose">
        <el-input v-model="draft" type="textarea" :rows="3" maxlength="2000" placeholder="记点什么…" />
        <div class="compose-bar">
          <div class="swatches">
            <button
              v-for="item in COLORS"
              :key="item"
              class="swatch"
              :class="[item, { active: color === item }]"
              :aria-label="`${COLOR_LABELS[item]}便签`"
              :aria-pressed="color === item"
              @click="color = item"
            />
          </div>
          <el-button type="primary" :loading="adding" :disabled="!draft.trim()" @click="addNote">添加</el-button>
        </div>
      </div>

      <div class="note-list">
        <div v-for="note in store.notes" :key="note.id" class="note" :class="`note-${note.color}`">
          <template v-if="editingId === note.id">
            <el-input v-model="editDraft" type="textarea" :rows="3" maxlength="2000" />
            <div class="note-actions">
              <el-button size="small" @click="cancelEdit">取消</el-button>
              <el-button size="small" type="primary" :loading="savingEdit" @click="saveEdit(note)">保存</el-button>
            </div>
          </template>
          <template v-else>
            <p class="note-content">{{ note.content }}</p>
            <div class="note-foot">
              <span>{{ formatTime(note.updated_at) }}</span>
              <span class="note-actions">
                <button class="icon-btn" aria-label="编辑便签" title="编辑便签" @click="startEdit(note)">
                  <el-icon :size="13"><EditPen /></el-icon>
                </button>
                <button class="icon-btn danger" aria-label="删除便签" title="删除便签" @click="removeNote(note)">
                  <el-icon :size="13"><Delete /></el-icon>
                </button>
              </span>
            </div>
          </template>
        </div>
        <p v-if="store.notes.length === 0" class="notes-empty">还没有便签，随手记下想到的事。</p>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, EditPen } from '@element-plus/icons-vue'
import { useBoardStore } from '../stores/board'
import type { Note, NoteColor } from '../types'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const store = useBoardStore()
const COLORS: NoteColor[] = ['yellow', 'green', 'blue', 'pink', 'purple']
const COLOR_LABELS: Record<NoteColor, string> = { yellow: '黄色', green: '绿色', blue: '蓝色', pink: '粉色', purple: '紫色' }

const draft = ref('')
const color = ref<NoteColor>('yellow')
const adding = ref(false)
const editingId = ref(0)
const editDraft = ref('')
const savingEdit = ref(false)

async function addNote(): Promise<void> {
  const content = draft.value.trim()
  if (!content) return
  adding.value = true
  try {
    await store.createNote({ content, color: color.value })
    draft.value = ''
    ElMessage.success('便签已添加')
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    adding.value = false
  }
}

function startEdit(note: Note): void {
  editingId.value = note.id
  editDraft.value = note.content
}

function cancelEdit(): void {
  editingId.value = 0
  editDraft.value = ''
}

async function saveEdit(note: Note): Promise<void> {
  const content = editDraft.value.trim()
  if (!content) {
    ElMessage.warning('便签内容不能为空')
    return
  }
  savingEdit.value = true
  try {
    await store.updateNote(note.id, { content })
    cancelEdit()
    ElMessage.success('便签已更新')
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    savingEdit.value = false
  }
}

async function removeNote(note: Note): Promise<void> {
  try {
    await store.deleteNote(note.id)
    ElMessage.success('便签已删除')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

function formatTime(value: string): string {
  if (!value) return ''
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.notes {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.compose {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.compose-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.swatches {
  display: flex;
  gap: 8px;
}

.swatch {
  width: 22px;
  height: 22px;
  border: 2px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  background-clip: padding-box;
}

.swatch:hover {
  outline: 1px solid var(--border-strong);
  outline-offset: 2px;
}

.swatch.active {
  outline: 1.5px solid var(--accent);
  outline-offset: 2px;
}

.swatch.yellow {
  background: #f2d478;
}

.swatch.green {
  background: #a8d8a0;
}

.swatch.blue {
  background: #a9cdec;
}

.swatch.pink {
  background: #f0b6c6;
}

.swatch.purple {
  background: #cbb7ea;
}

.note-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 24px;
}

.note {
  padding: 16px;
  border: 1px solid rgb(60 50 38 / 8%);
  border-radius: var(--radius-md);
  color: #3c3226;
}

.note-yellow {
  background: #f4edcf;
}

.note-green {
  background: #e5eddf;
}

.note-blue {
  background: #e1eaf1;
}

.note-pink {
  background: #f2e3e6;
}

.note-purple {
  background: #ebe5f1;
}

.note-content {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.note-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  color: rgba(60, 50, 38, 0.55);
  font-size: 11px;
}

.note-foot .icon-btn {
  color: rgba(60, 50, 38, 0.55);
}

.note-foot .icon-btn:hover {
  background: rgba(60, 50, 38, 0.12);
  color: #3c3226;
}

.note-actions {
  display: flex;
  gap: 4px;
}

.notes-empty {
  margin: 28px 0;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.note :deep(.el-textarea__inner) {
  color: #3c3226;
  background: rgba(255, 255, 255, 0.65);
}
</style>
