<template>
  <div class="app-shell">
    <header class="topbar">
      <div class="brand">
        <div class="logo">
          <el-icon :size="20"><Check /></el-icon>
        </div>
        <div class="brand-text">
          <h1>待办面板</h1>
        </div>
      </div>

      <div class="top-actions">
        <el-badge :value="board.notes.length" :hidden="board.notes.length === 0" :offset="[-2, 4]">
          <el-button class="toolbar-button" text :icon="Notebook" @click="notesVisible = true">
            <span class="btn-label">便签</span>
          </el-button>
        </el-badge>
        <el-button
          class="toolbar-button"
          text
          :icon="TrendCharts"
          aria-label="每日统计"
          title="每日统计"
          @click="statsVisible = true"
        />
        <el-badge :value="board.archived.length" :hidden="board.archived.length === 0" :offset="[-2, 4]">
          <el-button
            class="toolbar-button"
            text
            :icon="Box"
            aria-label="归档"
            title="归档"
            @click="archiveVisible = true"
          />
        </el-badge>
        <el-dropdown trigger="click" @command="setTheme">
          <el-button class="toolbar-button" text :icon="Brush" aria-label="切换主题" title="切换主题" />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="theme in THEMES"
                :key="theme.id"
                :command="theme.id"
                :disabled="ui.theme === theme.id"
              >
                <span class="theme-swatch" :class="`theme-${theme.id}`" />
                {{ theme.label }}
                <el-icon v-if="ui.theme === theme.id" class="theme-check"><Check /></el-icon>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" @command="handleUserCommand">
          <button class="user-chip" type="button" aria-label="账号菜单">
            <el-avatar :size="26" shape="square" class="avatar">{{ avatarText }}</el-avatar>
            <span class="username">{{ auth.user?.username }}</span>
            <el-icon :size="12"><ArrowDown /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="auth.isAdmin" command="admin">账号管理</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <main class="board-wrap">
      <div class="board-head">
        <div>
          <p class="board-date">{{ todayText }}</p>
          <h2>我的看板</h2>
          <p class="sub">{{ board.counts.todo }} 个待办，{{ board.counts.doing }} 个进行中</p>
        </div>
        <el-button class="create-task" type="primary" :icon="Plus" @click="openCreate('todo')">新建任务</el-button>
      </div>

      <div class="board-toolbar">
        <div class="board-view"><span>全部任务</span><span class="total">{{ board.total }}</span></div>
        <el-input v-model="search" class="search" placeholder="搜索任务…" aria-label="搜索任务" clearable :prefix-icon="Search" />
      </div>

      <div class="mobile-tabs" role="tablist" aria-label="任务状态切换">
        <button
          v-for="column in COLUMNS"
          :key="column.id"
          type="button"
          role="tab"
          :aria-selected="activeMobileColumn === column.id"
          :class="{ active: activeMobileColumn === column.id }"
          @click="activeMobileColumn = column.id"
        >
          {{ column.label }}<span class="tab-count">{{ board.counts[column.id] }}</span>
        </button>
      </div>

      <div v-loading="board.loading" class="board">
        <KanbanColumn
          v-for="column in COLUMNS"
          :key="column.id"
          :class="{ 'mobile-active': activeMobileColumn === column.id }"
          :column="column"
          :tasks="columnTasks(column.id)"
          :disabled="!!keyword"
          @add="openCreate(column.id)"
          @edit="openEdit"
          @delete="handleDelete"
          @pin="handlePin"
          @toggle="handleToggle"
          @archive="handleArchive"
          @toggle-subtask="handleToggleSubtask"
          @add-subtask="handleAddSubtask"
          @dragend="handleDragEnd"
        />
      </div>

      <footer class="board-footer">
        <span class="drag-hint">{{ keyword ? '搜索时暂不可拖动任务' : '拖动卡片，调整任务的顺序和状态' }}</span>
        <div class="progress-box">
          <span>已完成 {{ board.counts.done }} / {{ board.total }}</span>
          <el-progress :percentage="percent" :stroke-width="3" :show-text="false" />
          <span class="progress-percent">{{ percent }}%</span>
        </div>
      </footer>
    </main>

    <TaskDialog
      v-model="dialogVisible"
      :task="editingTask"
      :default-status="defaultStatus"
      :saving="saving"
      @submit="handleSubmit"
    />
    <NotesDrawer v-model="notesVisible" />
    <StatsDrawer v-model="statsVisible" />
    <ArchiveDrawer v-model="archiveVisible" />
    <AdminDialog v-model="adminVisible" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Box, Brush, Check, Notebook, Plus, Search, TrendCharts } from '@element-plus/icons-vue'
import AdminDialog from '../components/AdminDialog.vue'
import ArchiveDrawer from '../components/ArchiveDrawer.vue'
import KanbanColumn from '../components/KanbanColumn.vue'
import NotesDrawer from '../components/NotesDrawer.vue'
import StatsDrawer from '../components/StatsDrawer.vue'
import TaskDialog from '../components/TaskDialog.vue'
import { COLUMNS, useBoardStore } from '../stores/board'
import { useAuthStore } from '../stores/auth'
import { THEMES, useUiStore } from '../stores/ui'
import type { Task, TaskPayload, TaskStatus } from '../types'

const router = useRouter()
const auth = useAuthStore()
const board = useBoardStore()
const ui = useUiStore()

const search = ref('')
const notesVisible = ref(false)
const statsVisible = ref(false)
const archiveVisible = ref(false)
const adminVisible = ref(false)
const activeMobileColumn = ref<TaskStatus>('todo')
const dialogVisible = ref(false)
const editingTask = ref<Task | null>(null)
const defaultStatus = ref<TaskStatus>('todo')
const saving = ref(false)

const keyword = computed(() => search.value.trim().toLowerCase())
const percent = computed(() => (board.total ? Math.round((board.counts.done / board.total) * 100) : 0))
const avatarText = computed(() => (auth.user?.username || '?').slice(0, 1).toUpperCase())

const now = new Date()
const todayText = `${now.getMonth() + 1} 月 ${now.getDate()} 日 · ${now.toLocaleDateString('zh-CN', { weekday: 'long' })}`

onMounted(async () => {
  if (!auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      return
    }
  }
  try {
    await board.load()
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
})

function columnTasks(id: TaskStatus): Task[] {
  if (!keyword.value) return board.columns[id]
  return board.columns[id].filter(
    (task) =>
      task.title.toLowerCase().includes(keyword.value) ||
      (task.description || '').toLowerCase().includes(keyword.value),
  )
}

function openCreate(status: TaskStatus): void {
  editingTask.value = null
  defaultStatus.value = status
  dialogVisible.value = true
}

function openEdit(task: Task): void {
  editingTask.value = task
  dialogVisible.value = true
}

async function handleSubmit(payload: TaskPayload): Promise<void> {
  saving.value = true
  try {
    if (editingTask.value) {
      await board.updateTask(editingTask.value.id, payload)
      ElMessage.success('任务已更新')
    } else {
      await board.createTask(payload)
      ElMessage.success('任务已创建')
    }
    dialogVisible.value = false
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    saving.value = false
  }
}

async function handleDelete(task: Task): Promise<void> {
  try {
    await ElMessageBox.confirm(`确定删除任务「${task.title}」吗？`, '删除任务', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await board.deleteTask(task.id)
    ElMessage.success('任务已删除')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handlePin(task: Task): Promise<void> {
  try {
    await board.togglePin(task)
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handleToggle(task: Task): Promise<void> {
  try {
    await board.toggleDone(task)
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handleToggleSubtask(task: Task, subtaskId: string): Promise<void> {
  try {
    await board.toggleSubtask(task, subtaskId)
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handleAddSubtask(task: Task, title: string): Promise<void> {
  try {
    await board.addSubtask(task, title)
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handleArchive(task: Task): Promise<void> {
  try {
    await board.archiveTask(task)
    ElMessage.success('已归档，可在顶栏归档中查看')
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function handleDragEnd(): Promise<void> {
  try {
    await board.persistOrder()
  } catch {
    ElMessage.error('排序保存失败，正在刷新')
    try {
      await board.load()
    } catch {
      /* ignore */
    }
  }
}

function setTheme(command: string | number | object): void {
  ui.apply(String(command))
}

function handleUserCommand(command: string | number | object): void {
  const cmd = String(command)
  if (cmd === 'admin') adminVisible.value = true
  if (cmd === 'logout') {
    auth.logout()
    router.replace('/login')
  }
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  background: var(--bg);
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 60px;
  padding: 0 32px;
  border-bottom: 1px solid var(--border);
  background: var(--card);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.logo {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: var(--accent);
  color: var(--card);
}

.brand-text h1 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.top-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-button {
  height: 34px;
  padding: 8px;
  color: var(--muted);
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 220px;
  margin-left: 2px;
  padding: 0 0 0 16px;
  border: 0;
  border-left: 1px solid var(--border);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}

.avatar {
  flex-shrink: 0;
  border-radius: 5px;
  background: var(--accent-soft);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
}

.username {
  overflow: hidden;
  color: var(--text);
  font-size: 12px;
  text-overflow: ellipsis;
}

.board-wrap {
  max-width: 1400px;
  margin: 0 auto;
  padding: 40px 48px 32px;
}

.board-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 28px;
}

.board-date {
  margin: 0 0 6px;
  color: var(--muted);
  font-size: 12px;
}

.board-head h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
}

.sub {
  margin: 6px 0 0;
  color: var(--muted);
  font-size: 13px;
}

.create-task {
  height: 34px;
  padding: 0 14px;
}

.board-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding-bottom: 12px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border);
}

.board-view {
  display: flex;
  align-items: baseline;
  gap: 8px;
  color: var(--text);
  font-size: 13px;
  font-weight: 600;
}

.total {
  color: var(--muted);
  font-size: 12px;
  font-weight: 400;
  font-variant-numeric: tabular-nums;
}

.search {
  width: 240px;
}

.search :deep(.el-input__wrapper) {
  min-height: 32px;
  border-radius: var(--radius-md);
  background: var(--panel);
  box-shadow: none;
}

.search :deep(.el-input__wrapper.is-focus) {
  background: var(--card);
  box-shadow: 0 0 0 1px var(--accent) inset;
}

.board {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: stretch;
  gap: 16px;
}

.mobile-tabs {
  display: none;
}

.board-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-top: 16px;
  color: var(--muted);
  font-size: 12px;
}

.progress-box {
  display: flex;
  align-items: center;
  gap: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.progress-box .el-progress {
  width: 80px;
}

.progress-percent {
  min-width: 28px;
  text-align: right;
}

.theme-swatch {
  width: 12px;
  height: 12px;
  margin-right: 10px;
  border: 1px solid rgb(0 0 0 / 12%);
  border-radius: 3px;
}

.theme-paper { background: #e9ece5; }
.theme-night { background: #343c34; }
.theme-sunrise { background: #e8d7bd; }
.theme-mint { background: #c4dacb; }
.theme-sakura { background: #e6c9d5; }

.theme-check {
  margin-left: 20px;
}

@media (max-width: 1000px) {
  .topbar {
    padding: 0 24px;
  }

  .board-wrap {
    padding: 36px 24px;
  }

  .board {
    gap: 14px;
  }
}

@media (max-width: 760px) {
  .topbar {
    min-height: 60px;
    padding: 0 20px;
  }

  .top-actions {
    gap: 8px;
  }

  .username {
    display: none;
  }

  .user-chip {
    gap: 6px;
    padding-left: 12px;
  }

  .board-wrap {
    padding: 28px 20px;
  }

  .board-head {
    gap: 12px;
    margin-bottom: 24px;
  }

  .board-head h2 {
    font-size: 22px;
  }

  .board-toolbar {
    gap: 12px;
    margin-bottom: 16px;
  }

  .search {
    width: min(190px, 56vw);
  }

  .mobile-tabs {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 3px;
    padding: 3px;
    margin-bottom: 14px;
    border-radius: var(--radius-md);
    background: var(--panel);
  }

  .mobile-tabs button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 9px 6px;
    border: 0;
    border-radius: 4px;
    background: transparent;
    color: var(--muted);
    font-size: 13px;
    cursor: pointer;
  }

  .mobile-tabs button.active {
    background: var(--card);
    color: var(--text);
    box-shadow: 0 1px 2px rgb(20 26 21 / 6%);
  }

  .tab-count {
    color: var(--muted);
    font-size: 11px;
    font-variant-numeric: tabular-nums;
  }

  .board {
    display: block;
  }

  .board :deep(.kanban-col) {
    display: none;
  }

  .board :deep(.kanban-col.mobile-active) {
    display: flex;
  }

  .board-footer {
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 12px;
  }

  .progress-box {
    margin-left: auto;
  }
}

@media (max-width: 480px) {
  .btn-label {
    display: none;
  }

  .top-actions {
    gap: 6px;
  }
}

@media (max-width: 360px) {
  .topbar {
    padding: 0 14px;
  }

  .brand {
    gap: 7px;
  }

  .brand-text h1 {
    font-size: 14px;
  }

  .top-actions {
    gap: 4px;
  }
}
</style>
