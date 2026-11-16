export interface AppConfig {
  port: number
  apiBaseUrl: string
  modelProvider: string
  modelId: string
  dataDir: string
  settingsStore: "memory" | "redis"
  redisUrl: string | null
  settingsSecret: string
}

let cached: AppConfig | null = null

export function getConfig(): AppConfig {
  if (cached) return cached

  const explicitStore = process.env.SETTINGS_STORE?.trim()
  const settingsStore: "memory" | "redis" =
    explicitStore === "memory" || explicitStore === "redis"
      ? explicitStore
      : process.env.REDIS_ADDR
        ? "redis"
        : "memory"

  const addr = process.env.REDIS_ADDR?.trim() || "127.0.0.1:6379"
  const password = process.env.REDIS_PASSWORD?.trim() || ""
  const db = process.env.REDIS_DB?.trim() || "0"
  const redisUrl =
    settingsStore === "redis"
      ? `redis://${password ? `:${encodeURIComponent(password)}@` : ""}${addr}/${db}`
      : null

  cached = {
    port: Number(process.env.PORT ?? 3001),
    apiBaseUrl: (process.env.TODO_API_URL ?? "http://localhost:8080").replace(/\/+$/, ""),
    modelProvider: process.env.MODEL_PROVIDER?.trim() || "deepseek",
    modelId: process.env.MODEL_ID?.trim() || "deepseek-v4-pro",
    dataDir: process.env.AGENT_DATA_DIR?.trim() || ".agent-home",
    settingsStore,
    redisUrl,
    settingsSecret: process.env.MODEL_SETTINGS_SECRET?.trim() || "",
  }
  return cached
}
