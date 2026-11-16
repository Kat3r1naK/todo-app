import { mkdirSync } from "node:fs"
import { join } from "node:path"
import { EventType, type Event } from "@ag-ui/core"
import {
  createAgentSession,
  DefaultResourceLoader,
  ModelRuntime,
  SessionManager,
  SettingsManager,
} from "@earendil-works/pi-coding-agent"
import type { Model } from "@earendil-works/pi-ai"
import type { AppConfig } from "../config.js"
import type { ModelSettingsStore } from "../settings/types.js"
import { ThreadLock } from "../session/lock.js"
import { runContextStorage } from "../tools/context.js"
import { platformTools } from "../tools/registry.js"
import { AguiMapper } from "./aguiMapper.js"
import { buildSystemPrompt } from "./prompt.js"
import type { AgentRuntime, ModelOptions, RunRequest } from "./types.js"

type Session = Awaited<ReturnType<typeof createAgentSession>>["session"]

export class ModelConfigError extends Error {
  constructor(message: string) {
    super(message)
    this.name = "ModelConfigError"
  }
}

export class PiAgentRuntime implements AgentRuntime {
  private readonly sessions = new Map<string, Session>()
  private readonly lock = new ThreadLock()
  private readonly userRuntimes = new Map<number, ModelRuntime>()
  private resourceLoader?: DefaultResourceLoader
  private catalogRuntime?: ModelRuntime
  private catalogOptions?: ModelOptions

  constructor(
    private readonly config: AppConfig,
    private readonly settingsStore: ModelSettingsStore,
  ) {}

  async run(request: RunRequest, emit: (event: Event) => void): Promise<void> {
    const sessionKey = `user:${request.user.id}:${request.threadId}`

    await this.lock.runExclusive(sessionKey, async () => {
      try {
        const { runtime, model } = await this.resolveUserModel(request.user.id)
        const session = await this.getSession(sessionKey, runtime, model)

        const current = session.model
        if (current && (current.provider !== model.provider || current.id !== model.id)) {
          await session.setModel(model)
        }

        const mapper = new AguiMapper(emit)
        const unsubscribe = session.subscribe((event) => {
          switch (event.type) {
            case "message_start":
            case "message_end":
              mapper.handle({ type: event.type, message: { role: event.message.role } })
              break
            case "message_update":
              mapper.handle({
                type: "message_update",
                assistantMessageEvent: {
                  type: event.assistantMessageEvent.type,
                  delta:
                    "delta" in event.assistantMessageEvent &&
                    typeof event.assistantMessageEvent.delta === "string"
                      ? event.assistantMessageEvent.delta
                      : undefined,
                },
              })
              break
            case "tool_execution_start":
              mapper.handle({
                type: "tool_execution_start",
                toolCallId: event.toolCallId,
                toolName: event.toolName,
                args: event.args,
              })
              break
            case "tool_execution_end":
              mapper.handle({
                type: "tool_execution_end",
                toolCallId: event.toolCallId,
                result: event.result,
                isError: event.isError,
                durationMs: event.durationMs,
              })
              break
            default:
              break
          }
        })

        try {
          await runContextStorage.run(
            { runId: request.runId, threadId: request.threadId, token: request.token },
            async () => {
              await session.prompt(request.message)
            },
          )
        } finally {
          unsubscribe()
        }

        const modelError = findModelError(session)
        if (modelError) {
          console.warn(`[agent] run ${request.runId} model error: ${modelError}`)
          emit({
            type: EventType.RUN_ERROR,
            message: sanitizeError(new Error(modelError)),
            code: "MODEL_ERROR",
          })
          return
        }

        emit({
          type: EventType.RUN_FINISHED,
          threadId: request.threadId,
          runId: request.runId,
        })
      } catch (error) {
        console.error(`[agent] run ${request.runId} failed:`, error)
        emit({ type: EventType.RUN_ERROR, message: sanitizeError(error), code: "RUN_ERROR" })
      }
    })
  }

  async listModelOptions(): Promise<ModelOptions> {
    if (this.catalogOptions) return this.catalogOptions
    const runtime = await this.getCatalogRuntime()
    const providers = runtime
      .getProviders()
      .map((provider) => ({
        id: provider.id,
        name: provider.name || provider.id,
        models: runtime
          .getModels(provider.id)
          .map((model) => ({ id: model.id, name: model.name || model.id }))
          .sort((a, b) => a.name.localeCompare(b.name))
          .slice(0, 150),
      }))
      .filter((provider) => provider.models.length > 0)
      .sort((a, b) => a.id.localeCompare(b.id))
    this.catalogOptions = { providers }
    return this.catalogOptions
  }

  async hasModel(provider: string, modelId: string): Promise<boolean> {
    const runtime = await this.getCatalogRuntime()
    return runtime.getModel(provider, modelId) !== undefined
  }

  dispose(): void {
    for (const session of this.sessions.values()) {
      session.dispose()
    }
    this.sessions.clear()
  }

  private async getSession(sessionKey: string, runtime: ModelRuntime, model: Model<any>): Promise<Session> {
    const existing = this.sessions.get(sessionKey)
    if (existing) return existing

    const resourceLoader = await this.ensureResourceLoader()
    const { session } = await createAgentSession({
      cwd: this.config.dataDir,
      agentDir: join(this.config.dataDir, "agent"),
      model,
      modelRuntime: runtime,
      resourceLoader,
      noTools: "builtin",
      customTools: platformTools,
      sessionManager: SessionManager.inMemory(this.config.dataDir),
      settingsManager: SettingsManager.inMemory(),
    })
    this.sessions.set(sessionKey, session)
    return session
  }

  private async ensureResourceLoader(): Promise<DefaultResourceLoader> {
    if (this.resourceLoader) return this.resourceLoader

    mkdirSync(join(this.config.dataDir, "agent"), { recursive: true })
    this.resourceLoader = new DefaultResourceLoader({
      cwd: this.config.dataDir,
      agentDir: join(this.config.dataDir, "agent"),
      systemPromptOverride: () => buildSystemPrompt(),
      appendSystemPromptOverride: () => [],
      skillsOverride: () => ({ skills: [], diagnostics: [] }),
    })
    await this.resourceLoader.reload()
    return this.resourceLoader
  }

  private async getCatalogRuntime(): Promise<ModelRuntime> {
    if (!this.catalogRuntime) {
      this.catalogRuntime = await ModelRuntime.create()
    }
    return this.catalogRuntime
  }

  private async resolveUserModel(userId: number): Promise<{ runtime: ModelRuntime; model: Model<any> }> {
    let stored = null
    try {
      stored = await this.settingsStore.get(userId)
    } catch (error) {
      console.warn("[agent] 读取模型设置失败:", error)
      throw new ModelConfigError("读取模型设置失败，请稍后再试")
    }

    const provider = stored?.provider ?? this.config.modelProvider
    const modelId = stored?.modelId ?? this.config.modelId
    const defaultKey = provider === "deepseek" ? process.env.DEEPSEEK_API_KEY?.trim() : undefined
    const apiKey = stored?.apiKey ?? defaultKey

    if (!apiKey) {
      throw new ModelConfigError("尚未配置模型 API Key，请点击助手右上角的设置按钮填写")
    }

    let runtime = this.userRuntimes.get(userId)
    if (!runtime) {
      runtime = await ModelRuntime.create()
      this.userRuntimes.set(userId, runtime)
    }
    await runtime.setRuntimeApiKey(provider, apiKey)

    const model = runtime.getModel(provider, modelId)
    if (!model) {
      throw new ModelConfigError(`模型不可用：${provider}/${modelId}，请在模型设置中重新选择`)
    }
    console.info(`[agent] user ${userId} → ${provider}/${modelId}${stored ? " (custom)" : " (default)"}`)
    return { runtime, model }
  }
}

function findModelError(session: Session): string | null {
  const messages = session.agent.state.messages
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    const message = messages[index]
    if (message.role !== "assistant") continue
    if (message.stopReason === "error" || message.stopReason === "aborted") {
      return message.errorMessage || "模型调用失败"
    }
    return null
  }
  return null
}

function sanitizeError(error: unknown): string {
  if (error instanceof ModelConfigError) return error.message
  const raw = error instanceof Error ? error.message : String(error)
  if (/401|unauthorized|authentication|invalid.*key|key.*invalid/i.test(raw)) return "模型服务鉴权失败，请检查 API Key"
  if (/no api key/i.test(raw)) return "模型服务尚未配置 API Key，请在模型设置中填写"
  if (/timeout|timed out/i.test(raw)) return "模型响应超时，请稍后再试"
  return "助手暂时不可用，请稍后再试"
}
