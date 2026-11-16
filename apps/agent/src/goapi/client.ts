import { getConfig } from "../config.js"

const DEFAULT_TIMEOUT_MS = 5000

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = "ApiError"
  }
}

interface CallOptions {
  token: string
  method?: string
  body?: unknown
  timeoutMs?: number
}

function extractError(data: unknown, status: number): string {
  if (data && typeof data === "object" && "error" in data) {
    return String((data as { error: unknown }).error)
  }
  return `接口请求失败 (${status})`
}

export async function callApi<T>(path: string, options: CallOptions): Promise<T> {
  const { token, method = "GET", body, timeoutMs = DEFAULT_TIMEOUT_MS } = options
  const url = `${getConfig().apiBaseUrl}/api${path}`

  let res: Response
  try {
    res = await fetch(url, {
      method,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: AbortSignal.timeout(timeoutMs),
    })
  } catch (error) {
    const name = error instanceof Error ? error.name : ""
    if (name === "TimeoutError" || name === "AbortError") {
      throw new ApiError("平台接口请求超时", 0)
    }
    throw new ApiError("无法连接平台接口", 0)
  }

  const data = await res.json().catch(() => null)
  if (!res.ok) {
    throw new ApiError(extractError(data, res.status), res.status)
  }
  return data as T
}
