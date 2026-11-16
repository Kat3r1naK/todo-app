import type { ToolDefinition } from "@earendil-works/pi-coding-agent"
import type { TSchema } from "typebox"

export type ToolSideEffect = "read" | "write" | "dangerous"

export interface PlatformTool<TParams extends TSchema = TSchema> extends ToolDefinition<TParams> {
  sideEffect: ToolSideEffect
}

export function definePlatformTool<TParams extends TSchema>(tool: PlatformTool<TParams>): PlatformTool<TParams> {
  return tool
}
