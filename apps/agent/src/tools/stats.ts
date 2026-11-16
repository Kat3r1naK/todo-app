import { Type } from "typebox"
import { callApi } from "../goapi/client.js"
import { getRunContext } from "./context.js"
import { definePlatformTool } from "./types.js"

interface TaskDto {
  status: string
  completed_at: string | null
  archived_at: string | null
  created_at: string
}

const dayFormatter = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Shanghai" })

function dayKey(iso: string | null): string | null {
  if (!iso) return null
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return null
  return dayFormatter.format(date)
}

const dailyStatsParams = Type.Object({})

export const dailyStatsTool = definePlatformTool<typeof dailyStatsParams>({
  name: "get_daily_stats",
  label: "查询今日统计",
  description: "统计当前用户今天的任务创建数、完成数，以及未归档任务的状态分布。",
  sideEffect: "read",
  parameters: dailyStatsParams,
  async execute() {
    const { token } = getRunContext()
    const data = await callApi<{ tasks: TaskDto[] }>("/tasks", { token })
    const today = dayFormatter.format(new Date())

    const active = data.tasks.filter((task) => !task.archived_at)
    const byStatus = {
      todo: active.filter((task) => task.status === "todo").length,
      doing: active.filter((task) => task.status === "doing").length,
      done: active.filter((task) => task.status === "done").length,
    }
    const stats = {
      date: today,
      createdToday: data.tasks.filter((task) => dayKey(task.created_at) === today).length,
      completedToday: data.tasks.filter((task) => dayKey(task.completed_at) === today).length,
      byStatus,
      activeTotal: active.length,
      archivedTotal: data.tasks.length - active.length,
      note: "平台任务没有截止日期字段，日期统计基于创建时间与完成时间。",
    }

    return {
      content: [{ type: "text" as const, text: JSON.stringify(stats) }],
      details: stats,
    }
  },
})
