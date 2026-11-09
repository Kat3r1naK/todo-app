import { defineStore } from 'pinia'
import { api } from '../api'
import type {
  BoardColumn,
  Note,
  NotePayload,
  NotesResponse,
  NoteResponse,
  OkResponse,
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

interface BoardState {
  columns: Record<TaskStatus, Task[]>
  notes: Note[]
  loading: boolean
  loaded: boolean
}

export const useBoardStore = defineStore('board', {
  state: (): BoardState => ({
    columns: { todo: [], doing: [], done: [] },
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
        for (const task of taskRes.tasks ?? []) {
          const bucket = columns[task.status] ?? columns.todo
          bucket.push(task)
        }
        STATUS_IDS.forEach((id) => sortColumn(columns[id]))
        this.columns = columns
        this.notes = noteRes.notes ?? []
        this.loaded = true
      } finally {
        this.loading = false
      }
    },
    async createTask(payload: TaskPayload): Promise<Task> {
      const { task } = await api<TaskResponse>('/tasks', { method: 'POST', body: payload })
      this.columns[task.status].push(task)
      sortColumn(this.columns[task.status])
      return task
    },
    async updateTask(id: number, patch: TaskPatch): Promise<Task> {
      const { task } = await api<TaskResponse>(`/tasks/${id}`, { method: 'PATCH', body: patch })
      for (const status of STATUS_IDS) {
        const index = this.columns[status].findIndex((item) => item.id === id)
        if (index >= 0) this.columns[status].splice(index, 1)
      }
      this.columns[task.status].push(task)
      sortColumn(this.columns[task.status])
      return task
    },
    async togglePin(task: Task): Promise<Task> {
      return this.updateTask(task.id, { pinned: !task.pinned })
    },
    async deleteTask(id: number): Promise<void> {
      await api<OkResponse>(`/tasks/${id}`, { method: 'DELETE' })
      for (const status of STATUS_IDS) {
        this.columns[status] = this.columns[status].filter((item) => item.id !== id)
      }
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
