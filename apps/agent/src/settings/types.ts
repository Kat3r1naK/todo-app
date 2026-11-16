export interface StoredModelSettings {
  provider: string
  modelId: string
  apiKey: string
  updatedAt: string
}

export interface ModelSettingsInput {
  provider: string
  modelId: string
  apiKey: string
}

export interface ModelSettingsStore {
  readonly kind: "memory" | "redis"
  get(userId: number): Promise<StoredModelSettings | null>
  save(userId: number, input: ModelSettingsInput): Promise<void>
  clear(userId: number): Promise<void>
}
