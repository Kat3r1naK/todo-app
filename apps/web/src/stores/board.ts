import { defineStore } from 'pinia'
import { api } from '../api'
import type {
  BoardColumn,
  Note,
  NotePayload,
  NotesResponse,
  NoteResponse,
  OkResponse,
  Subtask,
  Task,
  TaskPatch,
  TaskPayload,
  TaskResponse,
  TasksResponse,
  TaskStatus,
} from '../types'

export const COLUMNS: BoardColumn[] = [
  { id: 'todo', label: '待办' },
  { id: 'doing', label: '进行中' },
  { id: 'done', label: '已完成' },
]

const STATUS_IDS: TaskStatus[] = COLUMNS.map((column) => column.id)

function sortColumn(list: Task[]): void {
  list.sort((a, b) => a.position - b.position || a.id - b.id)
}

function sortArchived(list: Task[]): void {
  list.sort((a, b) => (b.archived_at || '').localeCompare(a.archived_at || '') || b.id - a.id)
}

function normalizeTask(task: Task): Task {
  task.subtasks = task.subtasks ?? []
  return task
}

function newSubtaskID(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

interface BoardState {
  columns: Record<TaskStatus, Task[]>
  archived: Task[]
  notes: Note[]
  loading: boolean
  loaded: boolean
}

export const useBoardStore = defineStore('board', {
  state: (): BoardState => ({
    columns: { todo: [], doing: [], done: [] },
    archived: [],
    notes: [],
    loading: false,
    loaded: false,
  }),
  getters: {
    counts: (state): Record<TaskStatus, number> => ({
      todo: state.columns.todo.length,
      doing: state.columns.doing.length,
      done: state.columns.done.length,
    }),
    total: (state): number =>
      state.columns.todo.length + state.columns.doing.length + state.columns.done.length,
  },
  actions: {
    async load(): Promise<void> {
      this.loading = true
      try {
        const [taskRes, noteRes] = await Promise.all([
          api<TasksResponse>('/tasks'),
          api<NotesResponse>('/notes'),
        ])
        const columns: Record<TaskStatus, Task[]> = { todo: [], doing: [], done: [] }
        const archived: Task[] = []
        for (const raw of taskRes.tasks ?? []) {
          const task = normalizeTask(raw)
          if (task.archived_at) {
            archived.push(task)
            continue
          }
          const bucket = columns[task.status] ?? columns.todo
          bucket.push(task)
        }
        STATUS_IDS.forEach((id) => sortColumn(columns[id]))
        sortArchived(archived)
        this.columns = columns
        this.archived = archived
        this.notes = noteRes.notes ?? []
        this.loaded = true
      } finally {
        this.loading = false
      }
    },
    async createTask(payload: TaskPayload): Promise<Task> {
      const { task } = await api<TaskResponse>('/tasks', { method: 'POST', body: payload })
      normalizeTask(task)
      this.columns[task.status].push(task)
      sortColumn(this.columns[task.status])
      return task
    },
    async updateTask(id: number, patch: TaskPatch): Promise<Task> {
      const { task } = await api<TaskResponse>(`/tasks/${id}`, { method: 'PATCH', body: patch })
      normalizeTask(task)
      for (const status of STATUS_IDS) {
        this.columns[status] = this.columns[status].filter((item) => item.id !== id)
      }
      this.archived = this.archived.filter((item) => item.id !== id)
      if (task.archived_at) {
        this.archived.push(task)
        sortArchived(this.archived)
      } else {
        this.columns[task.status].push(task)
        sortColumn(this.columns[task.status])
      }
      return task
    },
    async archiveTask(task: Task): Promise<Task> {
      return this.updateTask(task.id, { archived: true })
    },
    async restoreTask(task: Task): Promise<Task> {
      return this.updateTask(task.id, { archived: false })
    },
    async togglePin(task: Task): Promise<Task> {
      return this.updateTask(task.id, { pinned: !task.pinned })
    },
    async toggleDone(task: Task): Promise<Task> {
      return this.updateTask(task.id, { status: task.status === 'done' ? 'todo' : 'done' })
    },
    async toggleSubtask(task: Task, subtaskId: string): Promise<Task> {
      const target = (task.subtasks ?? []).find((item) => item.id === subtaskId)
      if (!target) return task
      target.done = !target.done
      try {
        return await this.updateTask(task.id, { subtasks: task.subtasks })
      } catch (error) {
        target.done = !target.done
        throw error
      }
    },
    async addSubtask(task: Task, title: string): Promise<Task> {
      const trimmed = title.trim()
      if (!trimmed) return task
      const item: Subtask = { id: newSubtaskID(), title: trimmed, done: false }
      task.subtasks = [...(task.subtasks ?? []), item]
      try {
        return await this.updateTask(task.id, { subtasks: task.subtasks })
      } catch (error) {
        task.subtasks = task.subtasks.filter((subtask) => subtask.id !== item.id)
        throw error
      }
    },
    async deleteTask(id: number): Promise<void> {
      await api<OkResponse>(`/tasks/${id}`, { method: 'DELETE' })
      for (const status of STATUS_IDS) {
        this.columns[status] = this.columns[status].filter((item) => item.id !== id)
      }
      this.archived = this.archived.filter((item) => item.id !== id)
    },
    async persistOrder(): Promise<void> {
      const columns: Record<TaskStatus, number[]> = { todo: [], doing: [], done: [] }
      for (const status of STATUS_IDS) {
        columns[status] = this.columns[status].map((task) => task.id)
      }
      await api<OkResponse>('/tasks/reorder', { method: 'POST', body: { columns } })
      for (const status of STATUS_IDS) {
        this.columns[status].forEach((task, index) => {
          task.status = status
          task.position = index
        })
      }
    },
    async createNote(payload: NotePayload): Promise<Note> {
      const { note } = await api<NoteResponse>('/notes', { method: 'POST', body: payload })
      this.notes.unshift(note)
      return note
    },
    async updateNote(id: number, patch: Partial<NotePayload>): Promise<Note> {
      const { note } = await api<NoteResponse>(`/notes/${id}`, { method: 'PATCH', body: patch })
      const index = this.notes.findIndex((item) => item.id === id)
      if (index >= 0) this.notes[index] = note
      return note
    },
    async deleteNote(id: number): Promise<void> {
      await api<OkResponse>(`/notes/${id}`, { method: 'DELETE' })
      this.notes = this.notes.filter((item) => item.id !== id)
    },
  },
})
