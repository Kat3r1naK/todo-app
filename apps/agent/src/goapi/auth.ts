import { callApi } from "./client.js"

export interface AgentUser {
  id: number
  username: string
}

export async function verifyUserToken(token: string): Promise<AgentUser> {
  const data = await callApi<{ user: AgentUser }>("/auth/me", { token, timeoutMs: 3000 })
  if (!data.user || typeof data.user.id !== "number" || data.user.id <= 0) {
    throw new Error("用户信息异常")
  }
  return data.user
}
