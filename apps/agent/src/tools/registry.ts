import type { ToolDefinition } from "@earendil-works/pi-coding-agent"
import { dailyStatsTool } from "./stats.js"
import { listNotesTool } from "./notes.js"
import { listTasksTool, getTaskTool } from "./tasks.js"
import { completeTaskTool, createTaskTool, updateTaskTool } from "./taskWrite.js"

export const platformTools = [
  listTasksTool,
  getTaskTool,
  listNotesTool,
  dailyStatsTool,
  createTaskTool,
  updateTaskTool,
  completeTaskTool,
] as ToolDefinition[]

export const platformToolNames = platformTools.map((tool) => tool.name)
