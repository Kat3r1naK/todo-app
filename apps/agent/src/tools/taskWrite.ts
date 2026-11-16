import { Type } from "typebox"
import { callApi } from "../goapi/client.js"
import { getRunContext } from "./context.js"
import { definePlatformTool } from "./types.js"

interface TaskDto {
  id: number
  title: string
  description: string
  notes: string
  subtasks: Array<{ id: string; title: string; done: boolean }>
  status: string
  priority: string
  pinned: boolean
  completed_at: string | null
  archived_at: string | null
  created_at: string
  updated_at: string
}

function summarizeTask(task: TaskDto) {
  return {
    id: task.id,
    title: task.title,
    status: task.status,
    priority: task.priority,
    subtasks: `${task.subtasks.filter((s) => s.done).length}/${task.subtasks.length}`,
    completed_at: task.completed_at,
  }
}

const createTaskParams = Type.Object({
  title: Type.String({ minLength: 1, maxLength: 200, description: "任务标题，简洁明确" }),
  description: Type.Optional(Type.String({ description: "任务的补充说明" })),
  notes: Type.Optional(Type.String({ description: "任务的 Markdown 备注" })),
  priority: Type.Optional(
    Type.Union([Type.Literal("low"), Type.Literal("medium"), Type.Literal("high")], {
      description: "优先级，默认 medium",
    }),
  ),
  status: Type.Optional(
    Type.Union([Type.Literal("todo"), Type.Literal("doing"), Type.Literal("done")], {
      description: "初始状态，默认 todo",
    }),
  ),
  subtasks: Type.Optional(
    Type.Array(
      Type.Object({
        title: Type.String({ minLength: 1, maxLength: 200 }),
        done: Type.Optional(Type.Boolean()),
      }),
      { maxItems: 20, description: "子任务列表" },
    ),
  ),
})

export const createTaskTool = definePlatformTool<typeof createTaskParams>({
  name: "create_task",
  label: "创建任务",
  description:
    "在当前用户的看板中创建任务。平台没有截止日期字段，时间信息请写入标题或描述。创建后返回任务 ID 与状态。",
  sideEffect: "write",
  parameters: createTaskParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const body: Record<string, unknown> = { title: params.title.trim() }
    if (params.description !== undefined) body.description = params.description
    if (params.notes !== undefined) body.notes = params.notes
    if (params.priority) body.priority = params.priority
    if (params.status) body.status = params.status
    if (params.subtasks?.length) {
      body.subtasks = params.subtasks.map((subtask) => ({
        title: subtask.title,
        done: subtask.done ?? false,
      }))
    }

    const data = await callApi<{ task: TaskDto }>("/tasks", { token, method: "POST", body })
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({ created: summarizeTask(data.task) }),
        },
      ],
      details: { taskId: data.task.id },
    }
  },
})

const updateTaskParams = Type.Object({
  id: Type.Integer({ minimum: 1, description: "任务 ID（先用 list_tasks 或 get_task 获取）" }),
  title: Type.Optional(Type.String({ minLength: 1, maxLength: 200 })),
  description: Type.Optional(Type.String()),
  notes: Type.Optional(Type.String()),
  status: Type.Optional(
    Type.Union([Type.Literal("todo"), Type.Literal("doing"), Type.Literal("done")], {
      description: "新的状态；完成传 done，重新打开传 todo",
    }),
  ),
  priority: Type.Optional(
    Type.Union([Type.Literal("low"), Type.Literal("medium"), Type.Literal("high")]),
  ),
  pinned: Type.Optional(Type.Boolean()),
  subtasks: Type.Optional(
    Type.Array(
      Type.Object({
        id: Type.Optional(Type.String({ description: "已有子任务保留原 ID" })),
        title: Type.String({ minLength: 1, maxLength: 200 }),
        done: Type.Optional(Type.Boolean()),
      }),
      {
        maxItems: 100,
        description: "整体替换子任务列表；修改前先用 get_task 取得当前列表并带上未修改项",
      },
    ),
  ),
})

export const updateTaskTool = definePlatformTool<typeof updateTaskParams>({
  name: "update_task",
  label: "修改任务",
  description:
    "按任务 ID 修改标题、描述、备注、状态、优先级、置顶或子任务。subtasks 为整体替换，必须包含未修改的子任务。",
  sideEffect: "write",
  parameters: updateTaskParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const body: Record<string, unknown> = {}
    if (params.title !== undefined) body.title = params.title
    if (params.description !== undefined) body.description = params.description
    if (params.notes !== undefined) body.notes = params.notes
    if (params.status !== undefined) body.status = params.status
    if (params.priority !== undefined) body.priority = params.priority
    if (params.pinned !== undefined) body.pinned = params.pinned
    if (params.subtasks !== undefined) {
      body.subtasks = params.subtasks.map((subtask) => ({
        id: subtask.id,
        title: subtask.title,
        done: subtask.done ?? false,
      }))
    }

    const data = await callApi<{ task: TaskDto }>(`/tasks/${params.id}`, {
      token,
      method: "PATCH",
      body,
    })
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({ updated: summarizeTask(data.task) }),
        },
      ],
      details: { taskId: data.task.id },
    }
  },
})

const completeTaskParams = Type.Object({
  keyword: Type.String({ minLength: 1, description: "任务标题关键词，用于模糊匹配" }),
})

export const completeTaskTool = definePlatformTool<typeof completeTaskParams>({
  name: "complete_task",
  label: "完成任务",
  description:
    "按标题关键词把任务标记为已完成。匹配到多个任务时不会修改，返回候选列表让用户确认。",
  sideEffect: "write",
  parameters: completeTaskParams,
  async execute(_toolCallId, params) {
    const { token } = getRunContext()
    const data = await callApi<{ tasks: TaskDto[] }>("/tasks", { token })
    const keyword = params.keyword.trim().toLowerCase()
    const active = data.tasks.filter((task) => !task.archived_at)
    const exact = active.filter((task) => task.title.toLowerCase() === keyword)
    const matches = exact.length > 0 ? exact : active.filter((task) => task.title.toLowerCase().includes(keyword))

    if (matches.length === 0) {
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({ ok: false, reason: "没有找到匹配的任务，可以先用 list_tasks 查询" }),
          },
        ],
        details: { ok: false },
      }
    }

    if (matches.length > 1) {
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({
              ok: false,
              reason: "匹配到多个任务，需要用户确认具体是哪一条",
              candidates: matches.slice(0, 10).map(summarizeTask),
            }),
          },
        ],
        details: { ok: false, candidates: matches.length },
      }
    }

    const target = matches[0]
    if (target.status === "done") {
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({ ok: true, alreadyDone: true, task: summarizeTask(target) }),
          },
        ],
        details: { ok: true },
      }
    }

    const updated = await callApi<{ task: TaskDto }>(`/tasks/${target.id}`, {
      token,
      method: "PATCH",
      body: { status: "done" },
    })
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify({ ok: true, completed: summarizeTask(updated.task) }),
        },
      ],
      details: { ok: true, taskId: updated.task.id },
    }
  },
})
