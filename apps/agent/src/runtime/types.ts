import type { Event } from "@ag-ui/core"

export interface RunUser {
  id: number
  username: string
}

export interface RunRequest {
  runId: string
  threadId: string
  token: string
  message: string
  user: RunUser
}

export interface ModelOption {
  id: string
  name: string
}

export interface ProviderOption {
  id: string
  name: string
  models: ModelOption[]
}

export interface ModelOptions {
  providers: ProviderOption[]
}

export interface AgentRuntime {
  run(request: RunRequest, emit: (event: Event) => void): Promise<void>
  listModelOptions(): Promise<ModelOptions>
  hasModel(provider: string, modelId: string): Promise<boolean>
  dispose(): void
}
