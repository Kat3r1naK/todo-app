<template>
  <el-drawer
    :model-value="modelValue"
    title="归档"
    size="min(460px, 92vw)"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="archive">
      <div class="archive-top">
        <el-input v-model="keyword" placeholder="搜索标题、描述或备注…" clearable :prefix-icon="Search" />
        <span class="count">{{ filtered.length }} 项</span>
      </div>

      <div v-if="filtered.length" class="list">
        <article v-for="task in filtered" :key="task.id" class="item" :class="{ open: expanded === task.id }">
          <button type="button" class="item-head" @click="toggle(task.id)">
            <div class="item-main">
              <span class="item-title">{{ task.title }}</span>
              <span class="item-meta">
                <span v-if="task.completed_at">完成于 {{ formatDate(task.completed_at) }}</span>
                <span>归档于 {{ formatDate(task.archived_at) }}</span>
                <span v-if="task.notes" class="has-notes">含备注</span>
              </span>
            </div>
            <el-icon class="chevron" :size="14"><ArrowDown /></el-icon>
          </button>
          <div v-if="expanded === task.id" class="item-body">
            <p v-if="task.description" class="desc">{{ task.description }}</p>
            <MarkdownView v-if="task.notes" :source="task.notes" />
            <p v-else class="no-notes">这条任务没有备注</p>
            <div class="item-actions">
              <el-button size="small" @click="restore(task)">恢复到看板</el-button>
              <el-button size="small" type="danger" plain @click="remove(task)">删除</el-button>
            </div>
          </div>
        </article>
      </div>

      <p v-else class="empty">{{ keyword ? '没有匹配的归档任务' : '还没有归档任务' }}</p>
      <p class="tip">只有已完成的任务可以归档，归档不会影响每日统计。</p>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Search } from '@element-plus/icons-vue'
import { useBoardStore } from '../stores/board'
import MarkdownView from './MarkdownView.vue'
import type { Task } from '../types'

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const board = useBoardStore()
const keyword = ref('')
const expanded = ref<number | null>(null)

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return board.archived
  return board.archived.filter(
    (task) =>
      task.title.toLowerCase().includes(kw) ||
      (task.description || '').toLowerCase().includes(kw) ||
      (task.notes || '').toLowerCase().includes(kw),
  )
})

function toggle(id: number): void {
  expanded.value = expanded.value === id ? null : id
}

function formatDate(value: string | null): string {
  if (!value) return ''
  return new Date(value).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

async function restore(task: Task): Promise<void> {
  try {
    await board.restoreTask(task)
    if (expanded.value === task.id) expanded.value = null
    ElMessage.success('已恢复到看板')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function remove(task: Task): Promise<void> {
  try {
    await ElMessageBox.confirm(`确定删除归档任务「${task.title}」吗？删除后不可恢复。`, '删除任务', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await board.deleteTask(task.id)
    if (expanded.value === task.id) expanded.value = null
    ElMessage.success('任务已删除')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}
</script>

<style scoped>
.archive {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.archive-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.archive-top .el-input {
  flex: 1;
}

.count {
  flex-shrink: 0;
  color: var(--muted);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.item {
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--card);
}

.item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 12px 14px;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.item-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.item-title {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 10px;
  color: var(--muted);
  font-size: 11px;
}

.has-notes {
  color: var(--accent);
}

.chevron {
  flex-shrink: 0;
  color: var(--muted);
  transition: transform 0.15s;
}

.item.open .chevron {
  transform: rotate(180deg);
}

.item-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px 14px;
  border-top: 1px dashed var(--border);
}

.desc {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
}

.no-notes {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
}

.item-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 2px;
}

.empty {
  margin: 40px 0;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.tip {
  margin: 2px 0 0;
  color: var(--muted);
  font-size: 11px;
  line-height: 1.7;
}
</style>
