import { Type } from "typebox"
import { callApi } from "../goapi/client.js"
import { getRunContext } from "./context.js"
import { definePlatformTool } from "./types.js"

interface TaskDto {
  id: number
  title: string
  description: string
  notes: string
  status: string
  priority: string
  pinned: boolean
  subtasks: Array<{ id: string; title: string; done: boolean }>
  completed_at: string | null
  archived_at: string | null
  created_at: string
  updated_at: string
}

const listTasksParams = Type.Object({
  status: Type.Optional(
    Type.Union([Type.Literal("todo"), Type.Literal("doing"), Type.Literal("done")], {
      description: "按状态过滤",
    }),
  ),
  keyword: Type.Optional(Type.String({ description: "标题关键词，模糊匹配" })),
  includeArchived: Type.Optional(Type.Boolean({ description: "是否包含已归档任务，默认 false" })),
})

export const listTasksTool = definePlatformTool<typeof listTasksParams>({
  name: "list_tasks",
  label: "查询任务",
  description:
    "查询当前用户的任务列表，返回精简字段（不含描述与备注正文）。可按状态、标题关键词过滤，默认不包含已归档任务。",
  sideEffect: "read",
  parameters: listTasksParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const data = await callApi<{ tasks: TaskDto[] }>("/tasks", { token })
    const keyword = params.keyword?.trim().toLowerCase()

    const filtered = data.tasks.filter((task) => {
      if (!params.includeArchived && task.archived_at) return false
      if (params.status && task.status !== params.status) return false
      if (keyword && !task.title.toLowerCase().includes(keyword)) return false
      return true
    })

    const tasks = filtered.slice(0, 50).map((task) => ({
      id: task.id,
      title: task.title,
      status: task.status,
      priority: task.priority,
      pinned: task.pinned,
      subtasks: `${task.subtasks.filter((s) => s.done).length}/${task.subtasks.length}`,
      archived: Boolean(task.archived_at),
      completed_at: task.completed_at,
    }))

    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({ total: filtered.length, returned: tasks.length, tasks }),
        },
      ],
      details: { total: filtered.length },
    }
  },
})

const getTaskParams = Type.Object({
  id: Type.Integer({ minimum: 1, description: "任务 ID" }),
})

export const getTaskTool = definePlatformTool<typeof getTaskParams>({
  name: "get_task",
  label: "查看任务详情",
  description:
    "按任务 ID 获取任务完整信息（描述、备注、子任务）。修改备注或子任务前先用它读取当前内容。",
  sideEffect: "read",
  parameters: getTaskParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const data = await callApi<{ tasks: TaskDto[] }>("/tasks", { token })
    const task = data.tasks.find((item) => item.id === params.id)

    if (!task) {
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({ ok: false, reason: "任务不存在或无权访问" }),
          },
        ],
        details: { ok: false },
      }
    }

    const notes = task.notes ?? ""
    const notesTruncated = notes.length > 8000
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({
            ok: true,
            task: {
              id: task.id,
              title: task.title,
              description: task.description,
              notes: notesTruncated ? `${notes.slice(0, 8000)}…` : notes,
              notesTruncated,
              subtasks: task.subtasks.map((subtask) => ({
                id: subtask.id,
                title: subtask.title,
                done: subtask.done,
              })),
              status: task.status,
              priority: task.priority,
              pinned: task.pinned,
              completed_at: task.completed_at,
              archived_at: task.archived_at,
              created_at: task.created_at,
              updated_at: task.updated_at,
            },
          }),
        },
      ],
      details: { ok: true, taskId: task.id },
    }
  },
})
