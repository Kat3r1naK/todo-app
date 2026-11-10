export type TaskStatus = 'todo' | 'doing' | 'done'

export type TaskPriority = 'low' | 'medium' | 'high'

export type NoteColor = 'yellow' | 'green' | 'blue' | 'pink' | 'purple'

export interface User {
  id: number
  username: string
  is_admin: boolean
  created_at: string
}

export interface AdminUser extends User {
  autologin_key: string
  task_count: number
}

export interface Subtask {
  id: string
  title: string
  done: boolean
}

export interface Task {
  id: number
  title: string
  description: string
  notes: string
  subtasks: Subtask[]
  status: TaskStatus
  priority: TaskPriority
  pinned: boolean
  position: number
  completed_at: string | null
  archived_at: string | null
  created_at: string
  updated_at: string
}

export interface Note {
  id: number
  content: string
  color: NoteColor
  created_at: string
  updated_at: string
}

export interface BoardColumn {
  id: TaskStatus
  label: string
}

export interface TaskPayload {
  title: string
  description: string
  notes: string
  subtasks: Subtask[]
  status: TaskStatus
  priority: TaskPriority
}

export interface TaskPatch {
  title?: string
  description?: string
  notes?: string
  subtasks?: Subtask[]
  status?: TaskStatus
  priority?: TaskPriority
  pinned?: boolean
  archived?: boolean
}

export interface NotePayload {
  content: string
  color: NoteColor
}

export interface LoginResponse {
  token: string
  user: User
}

export interface MeResponse {
  user: User
}

export interface TasksResponse {
  tasks: Task[]
}

export interface TaskResponse {
  task: Task
}

export interface NotesResponse {
  notes: Note[]
}

export interface NoteResponse {
  note: Note
}

export interface AdminUsersResponse {
  users: AdminUser[]
}

export interface OkResponse {
  ok: boolean
}
