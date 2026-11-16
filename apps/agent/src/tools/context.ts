import { AsyncLocalStorage } from "node:async_hooks"

export interface RunContext {
  runId: string
  threadId: string
  token: string
}

export const runContextStorage = new AsyncLocalStorage<RunContext>()

export function getRunContext(): RunContext {
  const ctx = runContextStorage.getStore()
  if (!ctx) throw new Error("缺少运行上下文")
  return ctx
}
