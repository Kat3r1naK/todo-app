import { serve } from "@hono/node-server"
import { getConfig } from "./config.js"
import { PiAgentRuntime } from "./runtime/piRuntime.js"
import { createServer } from "./server.js"
import { createModelSettingsStore } from "./settings/store.js"

const config = getConfig()
if (!process.env.DEEPSEEK_API_KEY && config.modelProvider === "deepseek") {
  console.warn("[agent] 平台默认模型未设置 DEEPSEEK_API_KEY，用户需自行配置模型")
}

const settingsStore = createModelSettingsStore(config)
console.log(`[agent] settings store: ${settingsStore.kind}`)

const runtime = new PiAgentRuntime(config, settingsStore)
const app = createServer({ runtime, settingsStore, config })

const server = serve({ fetch: app.fetch, port: config.port }, (info) => {
  console.log(`[agent] listening on :${info.port}`)
})

function shutdown(signal: string): void {
  console.log(`[agent] ${signal} received, shutting down`)
  runtime.dispose()
  server.close(() => process.exit(0))
}

process.on("SIGINT", () => shutdown("SIGINT"))
process.on("SIGTERM", () => shutdown("SIGTERM"))
