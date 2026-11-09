<template>
  <div class="app-shell">
    <header class="topbar">
      <div class="brand">
        <div class="logo">
          <el-icon :size="22"><Checked /></el-icon>
        </div>
        <div class="brand-text">
          <h1>待办面板</h1>
          <p>VUE · GIN · MYSQL</p>
        </div>
      </div>

      <div class="top-actions">
        <el-input v-model="search" class="search" placeholder="搜索任务" clearable :prefix-icon="Search" />
        <el-badge :value="board.notes.length" :hidden="board.notes.length === 0" :offset="[-2, 4]">
          <el-button circle :icon="Notebook" @click="notesVisible = true" />
        </el-badge>
        <el-dropdown trigger="click" @command="setTheme">
          <el-button circle :icon="Brush" />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="theme in THEMES"
                :key="theme.id"
                :command="theme.id"
                :disabled="ui.theme === theme.id"
              >
                {{ theme.label }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" @command="handleUserCommand">
          <div class="user-chip">
            <el-avatar :size="28" class="avatar">{{ avatarText }}</el-avatar>
            <span class="username">{{ auth.user?.username }}</span>
            <el-icon :size="12"><ArrowDown /></el-icon>
          </div>
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
          <h2>{{ greeting }}，{{ auth.user?.username || '朋友' }}</h2>
          <p class="sub">{{ todayText }} · 共 {{ board.total }} 个任务，已完成 {{ board.counts.done }} 个</p>
        </div>
        <div class="progress-box">
          <div class="progress-label">
            <span>完成进度</span>
            <span>{{ percent }}%</span>
          </div>
          <el-progress :percentage="percent" :stroke-width="8" :show-text="false" />
        </div>
      </div>

      <div v-loading="board.loading" class="board">
        <KanbanColumn
          v-for="column in COLUMNS"
          :key="column.id"
          :column="column"
          :tasks="columnTasks(column.id)"
          :disabled="!!keyword"
          @add="openCreate(column.id)"
          @edit="openEdit"
          @delete="handleDelete"
          @pin="handlePin"
          @dragend="handleDragEnd"
        />
      </div>
    </main>

    <TaskDialog
      v-model="dialogVisible"
      :task="editingTask"
      :default-status="defaultStatus"
      :saving="saving"
      @submit="handleSubmit"
    />
    <NotesDrawer v-model="notesVisible" />
    <AdminDialog v-model="adminVisible" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Brush, Checked, Notebook, Search } from '@element-plus/icons-vue'
import AdminDialog from '../components/AdminDialog.vue'
import KanbanColumn from '../components/KanbanColumn.vue'
import NotesDrawer from '../components/NotesDrawer.vue'
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
const adminVisible = ref(false)
const dialogVisible = ref(false)
const editingTask = ref<Task | null>(null)
const defaultStatus = ref<TaskStatus>('todo')
const saving = ref(false)

const keyword = computed(() => search.value.trim().toLowerCase())
const percent = computed(() => (board.total ? Math.round((board.counts.done / board.total) * 100) : 0))
const avatarText = computed(() => (auth.user?.username || '?').slice(0, 1).toUpperCase())

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 5) return '夜深了'
  if (hour < 11) return '早上好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  if (hour < 22) return '晚上好'
  return '夜深了'
})

const todayText: string = new Date().toLocaleDateString('zh-CN', {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  weekday: 'long',
})

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
  background: var(--bg-glow), var(--bg);
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 28px;
  border-bottom: 1px solid var(--border);
  background: var(--topbar-bg);
  backdrop-filter: blur(16px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, var(--el-color-primary), var(--col-doing));
  box-shadow: 0 6px 16px rgba(74, 158, 255, 0.3);
}

.brand-text h1 {
  margin: 0;
  font-size: 18px;
  letter-spacing: 0.5px;
}

.brand-text p {
  margin: 3px 0 0;
  font-size: 9px;
  letter-spacing: 2px;
  color: var(--muted);
}

.top-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search {
  width: 210px;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--panel);
  cursor: pointer;
  outline: none;
}

.avatar {
  background: var(--el-color-primary);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
}

.username {
  font-size: 13px;
}

.board-wrap {
  max-width: 1440px;
  margin: 0 auto;
  padding: 26px 28px 52px;
}

.board-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.board-head h2 {
  margin: 0;
  font-size: 22px;
}

.board-head .sub {
  margin: 7px 0 0;
  font-size: 13px;
  color: var(--muted);
}

.progress-box {
  width: 230px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--muted);
}

.board {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  align-items: start;
}

@media (max-width: 1000px) {
  .topbar {
    padding: 12px 16px;
  }

  .search {
    width: 150px;
  }

  .board-wrap {
    padding: 18px 16px 40px;
  }

  .board {
    grid-auto-flow: column;
    grid-template-columns: none;
    grid-auto-columns: minmax(280px, 86vw);
    overflow-x: auto;
    padding-bottom: 12px;
  }
}

@media (max-width: 640px) {
  .username,
  .search {
    display: none;
  }

  .brand-text p {
    display: none;
  }

  .board-head h2 {
    font-size: 19px;
  }

  .progress-box {
    width: 100%;
  }
}
</style>
