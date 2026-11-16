import { randomUUID } from "node:crypto"
import { EventType, PROTOCOL_VERSION, type Event } from "@ag-ui/core"
import { EventSchema } from "@ag-ui/core/schemas"
import { EventEncoder } from "@ag-ui/encoder"
import { Hono } from "hono"
import { streamSSE } from "hono/streaming"
import type { AppConfig } from "./config.js"
import { ApiError } from "./goapi/client.js"
import { verifyUserToken, type AgentUser } from "./goapi/auth.js"
import { AsyncEventQueue } from "./http/events.js"
import type { AgentRuntime } from "./runtime/types.js"
import type { ModelSettingsStore } from "./settings/types.js"

const encoder = new EventEncoder()

interface ChatBody {
  message?: unknown
  threadId?: unknown
}

interface SettingsBody {
  provider?: unknown
  modelId?: unknown
  apiKey?: unknown
  clearApiKey?: unknown
}

type AuthResult = { ok: true; user: AgentUser } | { ok: false; status: 401 | 502; error: string }

function extractToken(header: string | undefined): string {
  let token = (header ?? "").trim()
  if (token.startsWith("Bearer ")) token = token.slice(7).trim()
  return token
}

async function authenticate(token: string): Promise<AuthResult> {
  try {
    return { ok: true, user: await verifyUserToken(token) }
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return { ok: false, status: 401, error: "登录已过期，请重新登录" }
    }
    return { ok: false, status: 502, error: "平台服务暂时不可用" }
  }
}

function maskKey(key: string): string {
  if (!key) return ""
  return key.length <= 4 ? "****" : `****${key.slice(-4)}`
}

export interface ServerDeps {
  runtime: AgentRuntime
  settingsStore: ModelSettingsStore
  config: AppConfig
}

export function createServer({ runtime, settingsStore, config }: ServerDeps): Hono {
  const app = new Hono()

  app.get("/healthz", (c) => c.json({ status: "ok" }))

  app.post("/agent/chat", async (c) => {
    const token = extractToken(c.req.header("Authorization"))
    if (!token) return c.json({ error: "未登录" }, 401)

    const body = await c.req.json<ChatBody>().catch(() => ({}) as ChatBody)
    const message = typeof body.message === "string" ? body.message.trim() : ""
    if (!message) return c.json({ error: "消息不能为空" }, 400)
    if (message.length > 4000) return c.json({ error: "消息过长" }, 400)

    const auth = await authenticate(token)
    if (!auth.ok) return c.json({ error: auth.error }, auth.status)

    const threadId =
      typeof body.threadId === "string" && body.threadId.trim()
        ? body.threadId.trim().slice(0, 100)
        : randomUUID()
    const runId = randomUUID()

    return streamSSE(c, async (stream) => {
      const queue = new AsyncEventQueue<Event>()
      queue.push({
        type: EventType.RUN_STARTED,
        threadId,
        runId,
        protocolVersion: PROTOCOL_VERSION,
      })

      const finished = runtime
        .run({ runId, threadId, token, message, user: auth.user }, (event) => queue.push(event))
        .catch((error) => {
          queue.push({
            type: EventType.RUN_ERROR,
            message: error instanceof Error ? error.message : String(error),
            code: "RUN_ERROR",
          })
        })
        .finally(() => queue.close())

      stream.onAbort(() => queue.close())

      try {
        for await (const event of queue) {
          const parsed = EventSchema.safeParse(event)
          if (!parsed.success) {
            console.error("[agent] invalid AG-UI event", event.type, parsed.error.issues)
          }
          await stream.write(encoder.encodeSSE(event))
        }
      } catch {
        queue.close()
      }
      await finished
    })
  })

  app.get("/agent/model-options", async (c) => {
    const token = extractToken(c.req.header("Authorization"))
    if (!token) return c.json({ error: "未登录" }, 401)
    const auth = await authenticate(token)
    if (!auth.ok) return c.json({ error: auth.error }, auth.status)
    try {
      return c.json(await runtime.listModelOptions())
    } catch {
      return c.json({ error: "模型目录暂时不可用" }, 502)
    }
  })

  app.get("/agent/model-settings", async (c) => {
    const token = extractToken(c.req.header("Authorization"))
    if (!token) return c.json({ error: "未登录" }, 401)
    const auth = await authenticate(token)
    if (!auth.ok) return c.json({ error: auth.error }, auth.status)

    try {
      const stored = await settingsStore.get(auth.user.id)
      if (stored) {
        return c.json({
          provider: stored.provider,
          modelId: stored.modelId,
          hasApiKey: true,
          maskedApiKey: maskKey(stored.apiKey),
          usingDefault: false,
          defaultProvider: config.modelProvider,
          defaultModel: config.modelId,
        })
      }
      return c.json({
        provider: config.modelProvider,
        modelId: config.modelId,
        hasApiKey: false,
        maskedApiKey: "",
        usingDefault: true,
        defaultProvider: config.modelProvider,
        defaultModel: config.modelId,
      })
    } catch {
      return c.json({ error: "读取模型设置失败" }, 503)
    }
  })

  app.put("/agent/model-settings", async (c) => {
    const token = extractToken(c.req.header("Authorization"))
    if (!token) return c.json({ error: "未登录" }, 401)
    const auth = await authenticate(token)
    if (!auth.ok) return c.json({ error: auth.error }, auth.status)

    const body = await c.req.json<SettingsBody>().catch(() => ({}) as SettingsBody)

    if (body.clearApiKey === true) {
      try {
        await settingsStore.clear(auth.user.id)
        return c.json({ ok: true, usingDefault: true })
      } catch {
        return c.json({ error: "清除模型设置失败" }, 503)
      }
    }

    const provider = typeof body.provider === "string" ? body.provider.trim() : ""
    const modelId = typeof body.modelId === "string" ? body.modelId.trim() : ""
    if (!provider || !modelId) return c.json({ error: "请选择服务商与模型" }, 400)
    if (provider.length > 100 || modelId.length > 200) return c.json({ error: "参数过长" }, 400)

    try {
      if (!(await runtime.hasModel(provider, modelId))) {
        return c.json({ error: "模型不存在或不可用" }, 400)
      }
    } catch {
      return c.json({ error: "模型目录暂时不可用" }, 502)
    }

    let existing = null
    try {
      existing = await settingsStore.get(auth.user.id)
    } catch {
      return c.json({ error: "读取模型设置失败" }, 503)
    }

    const apiKey =
      typeof body.apiKey === "string" && body.apiKey.trim() ? body.apiKey.trim() : existing?.apiKey ?? ""
    if (!apiKey) return c.json({ error: "请填写 API Key" }, 400)
    if (apiKey.length > 500) return c.json({ error: "API Key 过长" }, 400)

    if (settingsStore.kind === "redis" && !config.settingsSecret) {
      return c.json({ error: "服务端未配置模型设置加密密钥" }, 503)
    }

    try {
      await settingsStore.save(auth.user.id, { provider, modelId, apiKey })
    } catch {
      return c.json({ error: "保存模型设置失败" }, 503)
    }

    return c.json({
      ok: true,
      provider,
      modelId,
      hasApiKey: true,
      maskedApiKey: maskKey(apiKey),
      usingDefault: false,
    })
  })

  app.notFound((c) => c.json({ error: "not found" }, 404))

  return app
}
