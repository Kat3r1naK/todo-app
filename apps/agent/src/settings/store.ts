import { Redis } from "ioredis"
import type { AppConfig } from "../config.js"
import { decryptSecret, encryptSecret } from "./crypto.js"
import type { ModelSettingsInput, ModelSettingsStore, StoredModelSettings } from "./types.js"

const KEY_PREFIX = "agent:model-settings:"

interface StoredPayload {
  provider: string
  modelId: string
  apiKeyEnc: string
  updatedAt: string
}

class MemorySettingsStore implements ModelSettingsStore {
  readonly kind = "memory" as const
  private readonly settings = new Map<number, StoredModelSettings>()

  async get(userId: number): Promise<StoredModelSettings | null> {
    return this.settings.get(userId) ?? null
  }

  async save(userId: number, input: ModelSettingsInput): Promise<void> {
    this.settings.set(userId, {
      provider: input.provider,
      modelId: input.modelId,
      apiKey: input.apiKey,
      updatedAt: new Date().toISOString(),
    })
  }

  async clear(userId: number): Promise<void> {
    this.settings.delete(userId)
  }
}

class RedisSettingsStore implements ModelSettingsStore {
  readonly kind = "redis" as const

  constructor(
    private readonly redis: Redis,
    private readonly secret: string,
  ) {}

  async get(userId: number): Promise<StoredModelSettings | null> {
    const raw = await this.redis.get(`${KEY_PREFIX}${userId}`)
    if (!raw) return null
    const payload = JSON.parse(raw) as StoredPayload
    return {
      provider: payload.provider,
      modelId: payload.modelId,
      apiKey: decryptSecret(payload.apiKeyEnc, this.secret),
      updatedAt: payload.updatedAt,
    }
  }

  async save(userId: number, input: ModelSettingsInput): Promise<void> {
    if (!this.secret) {
      throw new Error("MODEL_SETTINGS_SECRET 未配置")
    }
    const payload: StoredPayload = {
      provider: input.provider,
      modelId: input.modelId,
      apiKeyEnc: encryptSecret(input.apiKey, this.secret),
      updatedAt: new Date().toISOString(),
    }
    await this.redis.set(`${KEY_PREFIX}${userId}`, JSON.stringify(payload))
  }

  async clear(userId: number): Promise<void> {
    await this.redis.del(`${KEY_PREFIX}${userId}`)
  }
}

export function createModelSettingsStore(config: AppConfig): ModelSettingsStore {
  if (config.settingsStore === "redis" && config.redisUrl) {
    const redis = new Redis(config.redisUrl, {
      maxRetriesPerRequest: 1,
      enableOfflineQueue: false,
    })
    redis.on("error", (error: Error) => {
      console.warn("[agent] redis error:", error.message)
    })
    if (!config.settingsSecret) {
      console.warn("[agent] MODEL_SETTINGS_SECRET 未配置，保存模型设置会失败")
    }
    return new RedisSettingsStore(redis, config.settingsSecret)
  }
  return new MemorySettingsStore()
}
